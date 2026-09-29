package tool

import "database/sql"

func Clan_credential(db *sql.DB, user_id string) string {
	var password, epoch string
	if !QueryRow_DB(db, "select data from user_set where id = ? and name = 'pw'", []any{&password}, user_id) {
		return ""
	}
	QueryRow_DB(db, "select data from user_set where id = ? and name = 'slbm_session_epoch'", []any{&epoch}, user_id)
	return password + ":" + epoch
}
