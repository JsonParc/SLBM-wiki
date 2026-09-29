package route

import (
	"net/url"
	"opennamu/route/tool"
	"strings"
)

func View_clan_users(config tool.Config, values url.Values) string {
	db := tool.DB_connect()
	defer tool.DB_close(db)
	if values != nil {
		if err := Api_clan_user_post(config, values); err != nil {
			return tool.Get_error_page(db, config, err.Error())
		}
		return tool.Get_redirect("/clan/users")
	}
	users, err := Api_clan_users(config)
	if err != nil {
		return tool.Get_error_page(db, config, "auth")
	}
	token, _ := config.Session.Get("clan_csrf").(string)
	if token == "" {
		token = tool.Get_random_key(48)
		config.Session.Set("clan_csrf", token)
		if config.Session.Save() != nil {
			return tool.Get_error_page(db, config, "error")
		}
	}
	lang := func(key string) string { return tool.Get_language(db, key, true) }
	escape := tool.HTML_escape
	var body strings.Builder
	body.WriteString(`<p><a href="/app_submit">` + lang("application_list") + `</a></p><p>` + lang("clan_password_note") + `</p><p>` + lang("clan_ip_note") + `</p>`)
	for _, user := range users {
		body.WriteString(`<section class="clan-user"><h2>` + escape(user.ID) + `</h2><p>` + lang("clan_joined") + `: ` + escape(user.Date) + ` · ` + lang("clan_status") + `: ` + escape(user.Group) + `</p><ul>`)
		for _, ip := range user.IPs {
			body.WriteString(`<li>` + escape(ip.Address) + ` (` + escape(ip.Group) + `)</li>`)
		}
		body.WriteString(`</ul>`)
		if user.Protected {
			body.WriteString(`<p>` + lang("clan_protected") + `</p></section>`)
			continue
		}
		hidden := `<input type="hidden" name="csrf" value="` + escape(token) + `"><input type="hidden" name="id" value="` + escape(user.ID) + `">`
		body.WriteString(`<form method="post">` + hidden + `<label>` + lang("clan_target") + ` <select name="target"><option value="account">` + lang("clan_account") + `</option><option value="ip">IP</option><option value="both">` + lang("clan_both") + `</option></select></label><p><label>IP <select name="ip"><option value="">` + lang("clan_select_ip") + `</option>`)
		for _, ip := range user.IPs {
			body.WriteString(`<option value="` + escape(ip.Address) + `">` + escape(ip.Address) + `</option>`)
		}
		body.WriteString(`</select></label></p><p><button name="action" value="ban">` + lang("clan_ban") + `</button> <button name="action" value="unban">` + lang("clan_unban") + `</button></p></form><hr class="main_hr">`)
		body.WriteString(`<form method="post">` + hidden + `<label>` + lang("clan_new_password") + ` <input type="password" name="password" minlength="12" maxlength="1024" required autocomplete="new-password"></label><p><button name="action" value="reset">` + lang("clan_reset") + `</button></p></form><hr class="main_hr">`)
		body.WriteString(`<form method="post">` + hidden + `<label>` + lang("clan_delete_note") + ` <input name="confirm" required autocomplete="off"></label><p><button name="action" value="delete">` + lang("clan_delete") + `</button></p></form></section><hr class="main_hr">`)
	}
	return tool.Get_template(db, config, lang("clan_users"), body.String(), []any{}, [][]any{{"manager", lang("return")}}, map[string]string{})
}
