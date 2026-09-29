package route

import (
	"net/url"
	"opennamu/route/tool"
	"strconv"
	"strings"
)

func View_clan_sidebar(config tool.Config) string {
	db := tool.DB_connect()
	defer tool.DB_close(db)
	lang := func(key string) string { return tool.Get_language(db, key, true) }
	winners, err := Api_clan_winners(config, false)
	if err != nil {
		return ""
	}
	var body strings.Builder
	body.WriteString(`<div class="slbm-hall slbm-side-card"><h2><button type="button" class="slbm-hall-toggle" aria-expanded="false" aria-controls="slbm-hall-list">` + lang("clan_hall") + ` <span aria-hidden="true">▾</span></button></h2>`)
	body.WriteString(`<div class="slbm-hall-stage" aria-hidden="true"></div><ol id="slbm-hall-list" class="slbm-hall-list">`)
	for i, entry := range winners {
		body.WriteString(`<li><span class="slbm-rank">` + strconv.Itoa(i+1) + `</span><span class="slbm-winner-name">` + tool.HTML_escape(entry.Name) + `</span><span class="slbm-wins">` + strconv.Itoa(entry.Wins) + lang("clan_wins_suffix") + `</span></li>`)
	}
	body.WriteString(`</ol>`)
	if len(winners) == 0 {
		body.WriteString(`<p>` + lang("clan_hall_empty") + `</p>`)
	}
	if Clan_admin(db, config) {
		body.WriteString(`<a class="slbm-hall-manage" href="/clan/hall">` + lang("clan_hall_manage") + `</a>`)
	}
	body.WriteString(`</div><div class="slbm-side-card slbm-recent"><h2><a href="/recent_changes">` + lang("clan_recent") + ` <span aria-hidden="true">›</span></a></h2><ul>`)
	changes := Api_clan_recent(config)
	for _, change := range changes {
		body.WriteString(`<li><a href="/w/` + tool.Url_parser(change.Title) + `" title="` + tool.HTML_escape(change.Title) + `">` + tool.HTML_escape(change.Title) + `</a><time datetime="` + tool.HTML_escape(strings.Replace(change.Date, " ", "T", 1)) + `+09:00" title="` + tool.HTML_escape(change.Date) + `">` + tool.HTML_escape(change.Date) + `</time></li>`)
	}
	body.WriteString(`</ul>`)
	if len(changes) == 0 {
		body.WriteString(`<p>` + lang("clan_recent_empty") + `</p>`)
	}
	body.WriteString(`</div>`)
	return body.String()
}

func View_clan_hall(config tool.Config, values url.Values) string {
	db := tool.DB_connect()
	defer tool.DB_close(db)
	if values != nil {
		if err := Api_clan_winners_post(config, values); err != nil {
			return tool.Get_error_page(db, config, "error")
		}
		return tool.Get_redirect("/clan/hall")
	}
	entries, err := Api_clan_winners(config, true)
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
	body := `<p>` + lang("clan_hall_help") + `</p><form method="post"><input type="hidden" name="csrf" value="` + tool.HTML_escape(token) + `">`
	for i := 0; i < 5; i++ {
		entry := Clan_winner{}
		if i < len(entries) {
			entry = entries[i]
		}
		suffix := strconv.Itoa(i)
		body += `<fieldset class="slbm-hall-field"><legend>` + strconv.Itoa(i+1) + `</legend><p><label>` + lang("clan_winner_name") + ` <input name="name_` + suffix + `" maxlength="128" value="` + tool.HTML_escape(entry.Name) + `"></label></p><p><label>` + lang("clan_wins") + ` <input type="number" name="wins_` + suffix + `" min="0" max="1000000" step="1" value="` + strconv.Itoa(entry.Wins) + `" required></label></p></fieldset>`
	}
	body += `<button type="submit">` + lang("save") + `</button></form>`
	return tool.Get_template(db, config, lang("clan_hall_manage"), body, []any{}, [][]any{{"user", lang("return")}}, map[string]string{})
}
