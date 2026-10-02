package route

import (
	"database/sql"
	"errors"

	"opennamu/route/tool"
)

func Secret_document_toggle(config tool.Config, action string, doc_name string) error {
	db := tool.DB_connect()
	defer tool.DB_close(db)
	if !tool.Can_manage_secret_document(db, config.IP) {
		return errors.New("auth")
	}
	if action != "set" && action != "unset" {
		return errors.New("not found")
	}
	var existing string
	if !tool.QueryRow_DB(db, "select title from data where title = ?", []any{&existing}, doc_name) {
		return errors.New("not found")
	}
	return tool.DB_transaction(db, func(tx *sql.Tx) error {
		if _, err := tx.Exec(tool.DB_change("delete from data_set where doc_name = ? and doc_rev = '' and set_name = 'secret_document'"), doc_name); err != nil {
			return err
		}
		if action == "set" {
			_, err := tx.Exec(tool.DB_change("insert into data_set (doc_name, doc_rev, set_name, set_data) values (?, '', 'secret_document', '1')"), doc_name)
			return err
		}
		return nil
	})
}

func Secret_document_redirect(config tool.Config, action string, doc_name string) string {
	if err := Secret_document_toggle(config, action, doc_name); err != nil {
		return tool.Get_redirect("/w/" + tool.Url_parser(doc_name))
	}
	return tool.Get_redirect("/w/" + tool.Url_parser(doc_name))
}