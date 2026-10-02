package route

import (
	"net/url"
	"strconv"
	"strings"

	"opennamu/route/tool"
)

// Clan_csrf_token returns the session's clan form token, creating it if needed.
func Clan_csrf_token(config tool.Config) (string, bool) {
	token, _ := config.Session.Get("clan_csrf").(string)
	if token != "" {
		return token, true
	}
	token = tool.Get_random_key(48)
	config.Session.Set("clan_csrf", token)
	return token, config.Session.Save() == nil
}

func Clan_role_label(lang func(string) string, role string) string {
	switch role {
	case tool.Clan_role_admin:
		return lang("clan_role_admin")
	case tool.Clan_role_sus:
		return lang("clan_role_sus")
	case tool.Clan_role_manager:
		return lang("clan_role_manager")
	case tool.Clan_role_guard:
		return lang("clan_role_guard")
	}
	return lang("clan_role_none")
}

func View_clan_staff(config tool.Config, values url.Values) string {
	db := tool.DB_connect()
	defer tool.DB_close(db)
	lang := func(key string) string { return tool.Get_language(db, key, true) }
	if values != nil {
		if err := Api_clan_staff_post(config, values); err != nil {
			if err.Error() == "clan_ban_limit" {
				return tool.Get_template(db, config, lang("clan_staff"), `<p>`+lang("clan_ban_limit")+`</p><p><a href="/clan/staff">`+lang("return")+`</a></p>`, []any{}, [][]any{{"clan/staff", lang("return")}}, map[string]string{})
			}
			return tool.Get_error_page(db, config, err.Error())
		}
		return tool.Get_redirect("/clan/staff")
	}
	page, err := Api_clan_staff(config)
	if err != nil {
		return tool.Get_error_page(db, config, "auth")
	}
	token, ok := Clan_csrf_token(config)
	if !ok {
		return tool.Get_error_page(db, config, "error")
	}
	escape := tool.HTML_escape

	var body strings.Builder
	body.WriteString(`<p>` + lang("clan_staff_help") + `</p>`)
	if page.Caller_role == tool.Clan_role_sus {
		body.WriteString(`<p>` + lang("clan_ban_left") + `: ` + strconv.Itoa(page.Bans_left) + ` / ` + strconv.Itoa(page.Ban_limit) + `</p>`)
	}
	body.WriteString(`<hr class="main_hr">`)
	for _, member := range page.Members {
		status := Clan_role_label(lang, member.Role)
		if member.Banned {
			status += ` · ` + lang("clan_banned")
		}
		body.WriteString(`<section class="clan-staff-member"><h2>` + escape(member.ID) + `</h2><p>` + status + `</p>`)
		hidden := `<input type="hidden" name="csrf" value="` + escape(token) + `"><input type="hidden" name="id" value="` + escape(member.ID) + `">`
		if member.Can_role {
			body.WriteString(`<form method="post">` + hidden + `<select name="role" aria-label="` + escape(lang("clan_role")) + `">`)
			for _, role := range page.Role_choices {
				selected := ""
				if role == member.Role {
					selected = " selected"
				}
				body.WriteString(`<option value="` + role + `"` + selected + `>` + Clan_role_label(lang, role) + `</option>`)
			}
			body.WriteString(`</select> <button name="action" value="role">` + lang("clan_role_save") + `</button></form>`)
		}
		if member.Can_ban {
			action, label := "ban", "clan_ban"
			if member.Banned {
				action, label = "unban", "clan_unban"
			}
			body.WriteString(`<form method="post">` + hidden + `<button name="action" value="` + action + `">` + lang(label) + `</button></form>`)
		}
		body.WriteString(`</section><hr class="main_hr">`)
	}
	return tool.Get_template(db, config, lang("clan_staff"), body.String(), []any{}, [][]any{}, map[string]string{})
}
