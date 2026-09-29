package route

import (
	"crypto/subtle"
	"database/sql"
	stdjson "encoding/json"
	"errors"
	"net/url"
	"opennamu/route/tool"
	"sort"
	"strconv"
	"strings"
)

type Clan_winner struct {
	Name string `json:"name"`
	Wins int    `json:"wins"`
}
type Clan_change struct {
	Title string
	Date  string
}

func Clan_winners_read(db *sql.DB) ([]Clan_winner, error) {
	raw := tool.Get_setting_value(db, "slbm_hall_of_fame", "", "")
	if raw == "" {
		return []Clan_winner{{"뉴비에요", 2}, {"KOREAGAMER17ˢᴸᴮᴹ", 2}, {"Shiroko SusツWïSHˢᴸᴮᴹ", 1}, {"김유정2호", 1}}, nil
	}
	entries := []Clan_winner{}
	if err := stdjson.Unmarshal([]byte(raw), &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func Api_clan_winners(config tool.Config, editing bool) ([]Clan_winner, error) {
	db := tool.DB_connect()
	defer tool.DB_close(db)
	if !tool.Check_permission(db, "site_view", config.IP) || (editing && !Clan_admin(db, config)) {
		return nil, errors.New("auth")
	}
	entries, err := Clan_winners_read(db)
	if err != nil {
		return nil, err
	}
	if editing {
		return entries, nil
	}
	visible := []Clan_winner{}
	for _, entry := range entries {
		if entry.Name != "" && entry.Wins > 0 {
			visible = append(visible, entry)
		}
	}
	sort.SliceStable(visible, func(i, j int) bool { return visible[i].Wins > visible[j].Wins })
	if len(visible) > 5 {
		visible = visible[:5]
	}
	return visible, nil
}

func Api_clan_winners_post(config tool.Config, values url.Values) error {
	db := tool.DB_connect()
	defer tool.DB_close(db)
	if !Clan_admin(db, config) {
		return errors.New("auth")
	}
	token, _ := config.Session.Get("clan_csrf").(string)
	if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(values.Get("csrf"))) != 1 {
		return errors.New("auth")
	}
	entries := []Clan_winner{}
	seen := map[string]bool{}
	for i := 0; i < 5; i++ {
		suffix := strconv.Itoa(i)
		name := strings.TrimSpace(values.Get("name_" + suffix))
		raw_wins := values.Get("wins_" + suffix)
		if raw_wins == "" {
			raw_wins = "0"
		}
		wins, err := strconv.Atoi(raw_wins)
		if err != nil || wins < 0 || wins > 1000000 || tool.Get_len(name) > 128 || (name != "" && seen[name]) {
			return errors.New("error")
		}
		if name != "" {
			entries = append(entries, Clan_winner{name, wins})
			seen[name] = true
		}
	}
	raw, err := stdjson.Marshal(entries)
	if err != nil {
		return err
	}
	return tool.DB_transaction(db, func(tx *sql.Tx) error {
		if _, err := tx.Exec(tool.DB_change("delete from other where name='slbm_hall_of_fame' and coverage=''")); err != nil {
			return err
		}
		if _, err := tx.Exec(tool.DB_change("insert into other(name,data,coverage) values ('slbm_hall_of_fame',?,'')"), string(raw)); err != nil {
			return err
		}
		tool.Do_insert_auth_history(tx, config.IP, "clan_hall_of_fame_update")
		return nil
	})
}

func Api_clan_recent(config tool.Config) []Clan_change {
	db := tool.DB_connect()
	defer tool.DB_close(db)
	result := []Clan_change{}
	if !tool.Check_permission(db, "site_view", config.IP) || !tool.Check_permission(db, "history_view", config.IP) {
		return result
	}
	// Fetch batches before checking page ACLs; never expose a restricted title.
	for offset := 0; offset < 500 && len(result) < 10; offset += 50 {
		rows := tool.Query_DB(db, "select title,date from history where hide='' order by date desc, title asc, cast(id as signed) desc limit 50 offset ?", offset)
		batch := []Clan_change{}
		for rows.Next() {
			var change Clan_change
			if rows.Scan(&change.Title, &change.Date) == nil {
				batch = append(batch, change)
			}
		}
		rows.Close()
		for _, change := range batch {
			if tool.Check_acl(db, change.Title, "", "render", config.IP) {
				result = append(result, change)
				if len(result) == 10 {
					return result
				}
			}
		}
		if len(batch) < 50 {
			break
		}
	}
	return result
}
