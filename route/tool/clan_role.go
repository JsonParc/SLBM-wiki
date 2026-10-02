package tool

import "database/sql"

// Clan staff roles sit below site admins (acl admin/owner) and never grant
// wiki admin permissions. They are stored as user_set name 'slbm_clan_role'.
const Clan_role_admin = "admin"
const Clan_role_sus = "sus"
const Clan_role_manager = "manager"
const Clan_role_guard = "guard"

// Get_stored_clan_role returns the saved staff role, ignoring site admin status.
func Get_stored_clan_role(db DB_runner, user_id string) string {
	role := ""
	QueryRow_DB(db, "select data from user_set where id = ? and name = 'slbm_clan_role' limit 1", []any{&role}, user_id)
	if role == Clan_role_sus || role == Clan_role_manager || role == Clan_role_guard {
		return role
	}
	return ""
}

// Get_clan_role returns "admin" for site admins, the staff role for active
// members, and "" for guests, banned accounts, and regular members.
func Get_clan_role(db *sql.DB, user_id string) string {
	if IP_or_user(user_id) {
		return ""
	}
	auth_info := Get_auth_info(db, user_id)
	if auth_info["admin"] || auth_info["owner"] {
		return Clan_role_admin
	}
	if Auth_group_name_ban(Get_user_auth(db, user_id)) || !Get_user_set_exists(db, user_id, "pw") {
		return ""
	}
	return Get_stored_clan_role(db, user_id)
}
