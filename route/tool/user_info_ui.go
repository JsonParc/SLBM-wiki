package tool

import (
	"database/sql"
	"strconv"
	"strings"
)

func Get_user_info_ui(db *sql.DB, config Config, user_name string) string {
	data := Get_user_info_ui_internal(db, config, user_name)
	return data + `<div>` + Get_language(db, "point", false) + `: ` + strconv.Itoa(Get_user_point(db, user_name)) + `</div>`
}

func Get_user_info_ui_internal(db *sql.DB, config Config, user_name string) string {
	auth_name := Get_user_auth(db, user_name)
	auth_date := Get_auth_date(db, user_name)
	if auth_date != "0" {
		auth_name += " (~" + auth_date + ")"
	}

	ban_state := Get_language(db, "normal", false)
	if Auth_group_name_ban(auth_name) {
		ban_state = "<a href=\"/auth/give_list\">" + Get_language(db, "ban", false) + "</a>"
	}

	online_icon := ""
	if !IP_or_user(user_name) && Get_user_set_exists(db, user_name, "pw") && Get_user_set_data(db, user_name, "online_status") == "on" {
		online_icon = "🔴 "
		if Check_online_user(user_name) {
			online_icon = "🟢 "
		}
	}

	// Clan staff roles are stored apart from the wiki permission group, so show them on their own line.
	clan_role_line := ""
	if !IP_or_user(user_name) && Get_user_set_exists(db, user_name, "pw") {
		role_key := "clan_role_none"
		switch Get_clan_role(db, user_name) {
		case Clan_role_admin:
			role_key = "clan_role_admin"
		case Clan_role_sus:
			role_key = "clan_role_sus"
		case Clan_role_manager:
			role_key = "clan_role_manager"
		case Clan_role_guard:
			role_key = "clan_role_guard"
		}
		clan_role_line = `<div class="slbm-clan-role">` + Get_language(db, "clan_role", true) + `: ` + Get_language(db, role_key, true) + `</div>`
	}

	level_data := Get_level(db, user_name)
	return `<div class="user_info_table">` + Get_language(db, "user_name", false) + `: ` + Get_user_profile_image_ui(db, user_name) + online_icon + IP_parser(db, user_name, config.IP) + `</div>` +
		`<div>` + Get_language(db, "authority", false) + `: ` + HTML_escape(auth_name) + `</div>` +
		clan_role_line +
		`<div>` + Get_language(db, "state", false) + `: ` + ban_state + `</div>` +
		`<div>` + Get_language(db, "level", false) + `: ` + HTML_escape(level_data[0]) + ` (` + HTML_escape(level_data[1]) + ` / ` + HTML_escape(level_data[2]) + `)</div>`
}

func Replace_user_info_ui(db *sql.DB, config Config, data string) string {
	const start_tag = `<div id="opennamu_get_user_info">`
	const end_tag = `</div>`

	for {
		start_index := strings.Index(data, start_tag)
		if start_index < 0 {
			return data
		}
		content_start := start_index + len(start_tag)
		end_offset := strings.Index(data[content_start:], end_tag)
		if end_offset < 0 {
			return data
		}
		end_index := content_start + end_offset
		user_name := HTML_unescape(data[content_start:end_index])
		data = data[:start_index] + Get_user_info_ui(db, config, user_name) + data[end_index+len(end_tag):]
	}
}
