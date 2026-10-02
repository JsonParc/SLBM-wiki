package tool

import "database/sql"

const secret_document_setting = "secret_document"

func Is_secret_document(db DB_runner, doc_name string) bool {
	value := ""
	return QueryRow_DB(db, "select set_data from data_set where doc_name = ? and doc_rev = '' and set_name = ? limit 1", []any{&value}, doc_name, secret_document_setting) && value == "1"
}

func Can_view_secret_document(db *sql.DB, user_id string) bool {
	role := Get_clan_role(db, user_id)
	return role == Clan_role_admin || role == Clan_role_sus || role == Clan_role_manager || role == Clan_role_guard
}

func Can_manage_secret_document(db *sql.DB, user_id string) bool {
	role := Get_clan_role(db, user_id)
	return role == Clan_role_admin || role == Clan_role_sus || role == Clan_role_manager
}

func Filter_secret_document_titles(db *sql.DB, user_id string, titles []string) []string {
	if Can_view_secret_document(db, user_id) {
		return titles
	}
	filtered := make([]string, 0, len(titles))
	for _, title := range titles {
		if !Is_secret_document(db, title) {
			filtered = append(filtered, title)
		}
	}
	return filtered
}

func Can_access_document(db *sql.DB, user_id string, doc_name string) bool {
	return !Is_secret_document(db, doc_name) || Can_view_secret_document(db, user_id)
}