package route

import (
	"net/url"
	"opennamu/route/tool"
)

func View_clan_document(config tool.Config, values url.Values) string {
	db := tool.DB_connect()
	defer tool.DB_close(db)
	if tool.IP_or_user(config.IP) || !tool.Get_user_set_exists(db, config.IP, "pw") {
		return tool.Get_redirect("/login")
	}
	if values != nil {
		target, err := Api_clan_document_target(config, values.Get("title"))
		if err != nil {
			return tool.Get_error_page(db, config, err.Error())
		}
		return tool.Get_redirect(target)
	}
	lang := func(key string) string { return tool.Get_language(db, key, true) }
	body := `<p>` + lang("clan_new_document_help") + `</p><form method="post"><label for="new_document_title">` + lang("document_name") + `</label><input id="new_document_title" name="title" required autofocus class="__ON_INPUT__"><hr class="main_hr"><button type="submit">` + lang("clan_new_document") + `</button></form>`
	return tool.Get_template(db, config, lang("clan_new_document"), body, []any{}, [][]any{{"user", lang("return")}}, map[string]string{})
}
