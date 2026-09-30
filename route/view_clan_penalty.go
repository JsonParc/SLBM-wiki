package route

import (
	"net/url"
	"strconv"
	"strings"

	"opennamu/route/tool"
)

var clan_penalty_sorts = []string{"recent", "points", "name"}

func View_clan_penalty_restore(config tool.Config, csrf string, raw []byte) string {
	err := Api_clan_penalty_import(config, csrf, raw)
	if err == nil {
		return View_clan_penalty(config, "recent", nil, "clan_penalty_restored")
	}
	if err.Error() == "clan_penalty_invalid" {
		return View_clan_penalty(config, "recent", nil, "clan_penalty_invalid")
	}
	db := tool.DB_connect()
	defer tool.DB_close(db)
	return tool.Get_error_page(db, config, err.Error())
}

// View_clan_penalty renders the penalty list. values holds an add/delete form
// post; notice is a language key to show above the page.
func View_clan_penalty(config tool.Config, sort_by string, values url.Values, notice string) string {
	db := tool.DB_connect()
	defer tool.DB_close(db)
	lang := func(key string) string { return tool.Get_language(db, key, true) }
	escape := tool.HTML_escape

	if values != nil {
		if err := Api_clan_penalty_post(config, values); err != nil {
			if err.Error() != "clan_penalty_input" {
				return tool.Get_error_page(db, config, err.Error())
			}
			notice = "clan_penalty_input"
		} else {
			return tool.Get_redirect("/clan/penalty")
		}
	}
	if !tool.Arr_in_str(clan_penalty_sorts, sort_by) {
		sort_by = "recent"
	}
	page, err := Api_clan_penalty_list(config, sort_by)
	if err != nil {
		if err.Error() == "require auth" {
			return tool.Get_error_page(db, config, "auth")
		}
		return tool.Get_error_page(db, config, "error")
	}
	token, ok := Clan_csrf_token(config)
	if !ok {
		return tool.Get_error_page(db, config, "error")
	}
	csrf := `<input type="hidden" name="csrf" value="` + escape(token) + `">`

	var body strings.Builder
	if notice != "" {
		body.WriteString(`<p class="slbm-penalty-notice">` + lang(notice) + `</p>`)
	}
	body.WriteString(`<h2>` + lang("clan_penalty_add") + `</h2><form method="post" action="/clan/penalty">` + csrf +
		`<input type="hidden" name="action" value="add">` +
		`<p><label>` + lang("clan_penalty_name") + ` <input name="name" required maxlength="100" autocomplete="off"></label></p>` +
		`<p><label>` + lang("clan_penalty_points") + ` <input type="number" name="points" required min="1" max="1000" value="1"></label></p>` +
		`<p><label>` + lang("clan_penalty_reason") + `<textarea name="reason" required maxlength="1000" rows="3"></textarea></label></p>` +
		`<p><button type="submit">` + lang("clan_penalty_add") + `</button></p></form><hr class="main_hr">`)

	body.WriteString(`<h2>` + lang("clan_penalty_list") + `</h2><p>` + lang("clan_penalty_sort") + `: `)
	for index, name := range clan_penalty_sorts {
		if index > 0 {
			body.WriteString(` · `)
		}
		label := lang("clan_penalty_sort_" + name)
		if name == sort_by {
			body.WriteString(`<strong>` + label + `</strong>`)
		} else {
			body.WriteString(`<a href="/clan/penalty/sort/` + name + `">` + label + `</a>`)
		}
	}
	body.WriteString(`</p>`)
	if len(page.Rows) == 0 {
		body.WriteString(`<p>` + lang("clan_penalty_empty") + `</p>`)
	} else {
		body.WriteString(`<div class="table_safe"><table class="slbm-penalty-table"><thead><tr><th>` + lang("clan_penalty_name") + `</th><th>` +
			lang("clan_penalty_points") + `</th><th>` + lang("clan_penalty_total") + `</th><th>` + lang("clan_penalty_reason") + `</th><th>` +
			lang("clan_penalty_date") + `</th><th>` + lang("clan_penalty_by") + `</th><th></th></tr></thead><tbody>`)
		for _, row := range page.Rows {
			body.WriteString(`<tr><td>` + escape(row.Name) + `</td><td>` + strconv.Itoa(row.Points) + `</td><td>` + strconv.Itoa(row.Total) +
				`</td><td>` + escape(row.Reason) + `</td><td>` + escape(row.Date) + `</td><td>` + escape(row.By) + `</td><td>` +
				`<form method="post" action="/clan/penalty" onsubmit="return confirm(this.dataset.confirm)" data-confirm="` + escape(lang("clan_penalty_delete_confirm")) + `">` + csrf +
				`<input type="hidden" name="id" value="` + escape(row.ID) + `"><button name="action" value="delete">` + lang("delete") + `</button></form></td></tr>`)
		}
		body.WriteString(`</tbody></table></div>`)
	}
	body.WriteString(`<hr class="main_hr">`)

	body.WriteString(`<h2>` + lang("clan_penalty_file") + `</h2><p><a href="/clan/penalty/download">` + lang("clan_penalty_download") + `</a></p>`)
	if page.Can_restore {
		body.WriteString(`<form method="post" action="/clan/penalty/restore" enctype="multipart/form-data" onsubmit="return confirm(this.dataset.confirm)" data-confirm="` +
			escape(lang("clan_penalty_restore_confirm")) + `">` + csrf +
			`<p>` + lang("clan_penalty_restore_help") + `</p><p><input type="file" name="penalty" accept=".json,application/json" required></p>` +
			`<p><button type="submit">` + lang("clan_penalty_restore") + `</button></p></form>`)
	}
	body.WriteString(`<hr class="main_hr">`)

	body.WriteString(`<h2>` + lang("clan_penalty_log") + `</h2>`)
	if len(page.Log) == 0 {
		body.WriteString(`<p>` + lang("clan_penalty_log_empty") + `</p>`)
	} else {
		body.WriteString(`<ul class="slbm-penalty-log">`)
		for _, item := range page.Log {
			action := escape(item.Action)
			if tool.Arr_in_str([]string{"add", "delete", "restore"}, item.Action) {
				action = lang("clan_penalty_action_" + item.Action)
			}
			line := escape(item.Time) + ` · ` + escape(item.By) + ` · ` + action
			if item.Action == "restore" {
				line += ` (` + strconv.Itoa(item.Points) + `)`
			} else {
				line += ` · ` + escape(item.Name) + ` ` + strconv.Itoa(item.Points) + ` · ` + escape(item.Reason)
			}
			body.WriteString(`<li>` + line + `</li>`)
		}
		body.WriteString(`</ul>`)
	}
	return tool.Get_template(db, config, lang("clan_penalty"), body.String(), []any{}, [][]any{}, map[string]string{})
}
