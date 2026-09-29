package route

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	stdjson "encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"opennamu/route/tool"
)

// A site backup is a zip holding a consistent SQLite snapshot (data.db) and
// every uploaded file (images/*), so a redeploy can be restored in one step.

const clan_backup_version = 1
const clan_backup_db_max_size = 512 * 1024 * 1024
const clan_backup_image_max_size = 64 * 1024 * 1024

var clan_backup_required_tables = []string{"data", "history", "other", "user_set"}

type clan_backup_manifest struct {
	Version int    `json:"version"`
	Created string `json:"created"`
	Site    string `json:"site"`
}

type Clan_backup_result struct {
	Tables    int
	Documents int
	Images    int
	Skipped   []string
	Snapshot  string
}

func clan_backup_images_dir() string {
	return filepath.Join("data", "images")
}

// Clan_backup_snapshot writes a consistent copy of the live database to path.
func Clan_backup_snapshot(db *sql.DB, path string) error {
	if tool.DB_type() != "sqlite" {
		return errors.New("clan_backup_sqlite_only")
	}
	_ = os.Remove(path)
	// VACUUM cannot run inside a transaction, so bypass Exec_DB.
	_, err := db.Exec("VACUUM INTO ?", path)
	return err
}

func Api_clan_backup_export(config tool.Config) ([]byte, string, error) {
	db := tool.DB_connect()
	defer tool.DB_close(db)
	if !Clan_admin(db, config) {
		return nil, "", errors.New("auth")
	}

	temp_dir, err := os.MkdirTemp("", "slbm-backup-")
	if err != nil {
		return nil, "", err
	}
	defer os.RemoveAll(temp_dir)

	snapshot := filepath.Join(temp_dir, "data.db")
	if err := Clan_backup_snapshot(db, snapshot); err != nil {
		return nil, "", err
	}

	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	add_file := func(name string, source string) error {
		file, err := os.Open(source)
		if err != nil {
			return err
		}
		defer file.Close()
		writer, err := archive.Create(name)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, file)
		return err
	}

	now := time.Now()
	manifest, _ := stdjson.Marshal(clan_backup_manifest{
		Version: clan_backup_version,
		Created: now.Format("2006-01-02 15:04:05"),
		Site:    tool.Get_setting_value(db, "name", "", ""),
	})
	writer, err := archive.Create("manifest.json")
	if err != nil {
		return nil, "", err
	}
	if _, err := writer.Write(manifest); err != nil {
		return nil, "", err
	}
	if err := add_file("data.db", snapshot); err != nil {
		return nil, "", err
	}

	entries, err := os.ReadDir(clan_backup_images_dir())
	if err != nil && !os.IsNotExist(err) {
		return nil, "", err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if err := add_file("images/"+entry.Name(), filepath.Join(clan_backup_images_dir(), entry.Name())); err != nil {
			return nil, "", err
		}
	}
	if err := archive.Close(); err != nil {
		return nil, "", err
	}

	tool.Do_insert_auth_history(db, config.IP, "clan_backup_export")
	return buffer.Bytes(), "slbm-backup-" + now.Format("20060102-150405") + ".zip", nil
}

func clan_backup_valid_image_name(name string) bool {
	return name != "" && name == filepath.Base(name) && !strings.HasPrefix(name, ".") &&
		!strings.ContainsAny(name, `/\:`)
}

func clan_backup_table_columns(conn *sql.Conn, schema string, table string) ([]string, error) {
	rows, err := conn.QueryContext(context.Background(), "select name from pragma_table_info(?, ?)", table, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns := []string{}
	for rows.Next() {
		name := ""
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		columns = append(columns, name)
	}
	return columns, rows.Err()
}

func clan_backup_titles(db *sql.DB) map[string]bool {
	titles := map[string]bool{}
	rows := tool.Query_DB(db, "select title from data")
	defer rows.Close()
	for rows.Next() {
		title := ""
		if rows.Scan(&title) == nil {
			titles[title] = true
		}
	}
	return titles
}

// clan_backup_check_db makes sure the uploaded file is an intact wiki database.
func clan_backup_check_db(path string) error {
	check, err := sql.Open("sqlite", path)
	if err != nil {
		return err
	}
	defer check.Close()
	result := ""
	if err := check.QueryRow("pragma quick_check").Scan(&result); err != nil || result != "ok" {
		return errors.New("clan_backup_invalid")
	}
	for _, table := range clan_backup_required_tables {
		name := ""
		if check.QueryRow("select name from sqlite_master where type = 'table' and name = ?", table).Scan(&name) != nil {
			return errors.New("clan_backup_invalid")
		}
	}
	return nil
}

func Api_clan_backup_restore(config tool.Config, raw []byte) (Clan_backup_result, error) {
	result := Clan_backup_result{}
	db := tool.DB_connect()
	defer tool.DB_close(db)
	if !Clan_admin(db, config) || !tool.Check_permission(db, "owner", config.IP) {
		return result, errors.New("auth")
	}
	if tool.DB_type() != "sqlite" {
		return result, errors.New("clan_backup_sqlite_only")
	}

	archive, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return result, errors.New("clan_backup_invalid")
	}

	temp_dir, err := os.MkdirTemp("", "slbm-restore-")
	if err != nil {
		return result, err
	}
	defer os.RemoveAll(temp_dir)

	backup_db := ""
	images := []*zip.File{}
	for _, file := range archive.File {
		switch {
		case file.Name == "data.db":
			if file.UncompressedSize64 > clan_backup_db_max_size {
				return result, errors.New("clan_backup_invalid")
			}
			backup_db = filepath.Join(temp_dir, "data.db")
			if err := clan_backup_extract(file, backup_db, clan_backup_db_max_size); err != nil {
				return result, errors.New("clan_backup_invalid")
			}
		case strings.HasPrefix(file.Name, "images/") && !file.FileInfo().IsDir():
			if !clan_backup_valid_image_name(strings.TrimPrefix(file.Name, "images/")) || file.UncompressedSize64 > clan_backup_image_max_size {
				return result, errors.New("clan_backup_invalid")
			}
			images = append(images, file)
		}
	}
	if backup_db == "" {
		return result, errors.New("clan_backup_invalid")
	}
	if err := clan_backup_check_db(backup_db); err != nil {
		return result, err
	}

	// Keep the current state on disk so a wrong upload can still be undone.
	if err := os.MkdirAll(filepath.Join("data", "backups"), 0o755); err != nil {
		return result, err
	}
	result.Snapshot = filepath.Join("data", "backups", "before-restore-"+time.Now().Format("20060102-150405")+".db")
	if err := Clan_backup_snapshot(db, result.Snapshot); err != nil {
		return result, err
	}

	old_titles := clan_backup_titles(db)

	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		return result, err
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, "attach database ? as backup", backup_db); err != nil {
		return result, err
	}
	defer conn.ExecContext(ctx, "detach database backup")

	table_names := []string{}
	for table := range tool.DB_table_list() {
		table_names = append(table_names, table)
	}
	sort.Strings(table_names)

	// Resolve columns before the transaction; the attached schema is read-only here.
	type table_copy struct {
		name    string
		columns []string
	}
	copies := []table_copy{}
	for _, table := range table_names {
		backup_columns, err := clan_backup_table_columns(conn, "backup", table)
		if err != nil {
			return result, err
		}
		if len(backup_columns) == 0 {
			result.Skipped = append(result.Skipped, table)
			continue
		}
		main_columns, err := clan_backup_table_columns(conn, "main", table)
		if err != nil {
			return result, err
		}
		in_backup := map[string]bool{}
		for _, column := range backup_columns {
			in_backup[column] = true
		}
		shared := []string{}
		for _, column := range main_columns {
			if in_backup[column] {
				shared = append(shared, `"`+column+`"`)
			}
		}
		copies = append(copies, table_copy{table, shared})
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return result, err
	}
	for _, copy := range copies {
		if _, err := tx.Exec(`delete from main."` + copy.name + `"`); err != nil {
			tx.Rollback()
			return result, err
		}
		if len(copy.columns) == 0 {
			continue
		}
		column_list := strings.Join(copy.columns, ",")
		if _, err := tx.Exec(`insert into main."` + copy.name + `" (` + column_list + `) select ` + column_list + ` from backup."` + copy.name + `"`); err != nil {
			tx.Rollback()
			return result, err
		}
		result.Tables++
	}
	if err := tx.Commit(); err != nil {
		return result, err
	}

	if err := os.MkdirAll(clan_backup_images_dir(), 0o755); err != nil {
		return result, err
	}
	for _, file := range images {
		target := filepath.Join(clan_backup_images_dir(), strings.TrimPrefix(file.Name, "images/"))
		if err := clan_backup_extract(file, target, clan_backup_image_max_size); err != nil {
			return result, err
		}
		result.Images++
	}

	// Bring the live search index in line with the restored documents.
	rows := tool.Query_DB(db, "select title, coalesce(data, '') from data")
	for rows.Next() {
		title, data := "", ""
		if rows.Scan(&title, &data) != nil {
			continue
		}
		delete(old_titles, title)
		tool.Search_index_update(title, data)
		result.Documents++
	}
	rows.Close()
	for title := range old_titles {
		tool.Search_index_delete(title)
	}
	_ = tool.Search_bbs_index_mark_rebuild()

	tool.Do_insert_auth_history(db, config.IP, "clan_backup_restore")
	return result, nil
}

func clan_backup_extract(file *zip.File, target string, limit int64) error {
	source, err := file.Open()
	if err != nil {
		return err
	}
	defer source.Close()
	temp := target + ".restore"
	output, err := os.Create(temp)
	if err != nil {
		return err
	}
	written, err := io.Copy(output, io.LimitReader(source, limit+1))
	close_err := output.Close()
	if err == nil {
		err = close_err
	}
	if err == nil && written > limit {
		err = errors.New("clan_backup_invalid")
	}
	if err != nil {
		os.Remove(temp)
		return err
	}
	return os.Rename(temp, target)
}
