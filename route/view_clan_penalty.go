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
			if err.Error() != "clan_penalty_input" && err.Error() != "clan_penalty_zero" {
				return tool.Get_error_page(db, config, err.Error())
			}
			notice = err.Error()
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
	if len(page.Summaries) == 0 {
		body.WriteString(`<p>` + lang("clan_penalty_empty") + `</p>`)
	} else {
		body.WriteString(`<div class="table_safe"><table class="slbm-penalty-table"><thead><tr><th>` + lang("clan_penalty_name") + `</th><th>` +
			lang("clan_penalty_total") + `</th><th>` + lang("clan_penalty_count") + `</th><th>` + lang("clan_penalty_last_date") + `</th><th>` +
			lang("clan_penalty_last_reason") + `</th><th></th></tr></thead><tbody>`)
		for _, row := range page.Summaries {
			body.WriteString(`<tr><td>` + escape(row.Name) + `</td><td>` + strconv.Itoa(row.Total) + `</td><td>` + strconv.Itoa(row.Count) +
				`</td><td>` + escape(row.Last_date) + `</td><td>` + escape(row.Last_reason) + `</td><td>` +
				`<details class="slbm-penalty-adjust"><summary>` + lang("clan_penalty_adjust") + `</summary>` +
				`<form method="post" action="/clan/penalty">` + csrf + `<input type="hidden" name="name" value="` + escape(row.Name) + `">` +
				`<p><label>` + lang("clan_penalty_reason") + ` <input name="reason" required maxlength="1000" autocomplete="off"></label></p>` +
				`<p><button name="action" value="increase">` + lang("clan_penalty_increase") + `</button> ` +
				`<button name="action" value="decrease">` + lang("clan_penalty_decrease") + `</button></p></form></details></td></tr>`)
		}
		body.WriteString(`</tbody></table></div>`)
	}
	body.WriteString(`<hr class="main_hr">`)

	// Individual entries are a read-only record, folded by default.
	body.WriteString(`<details class="slbm-penalty-entries"><summary><h2><span aria-hidden="true">▸ </span>` + lang("clan_penalty_entries") + ` (` + strconv.Itoa(len(page.Entries)) + `)</h2></summary>`)
	if len(page.Entries) == 0 {
		body.WriteString(`<p>` + lang("clan_penalty_empty") + `</p>`)
	} else {
		body.WriteString(`<div class="table_safe"><table class="slbm-penalty-entry-table"><thead><tr><th>` + lang("clan_penalty_name") + `</th><th>` +
			lang("clan_penalty_points") + `</th><th>` + lang("clan_penalty_reason") + `</th><th>` + lang("clan_penalty_date") + `</th><th>` +
			lang("clan_penalty_by") + `</th></tr></thead><tbody>`)
		for _, entry := range page.Entries {
			body.WriteString(`<tr><td>` + escape(entry.Name) + `</td><td>` + strconv.Itoa(entry.Points) + `</td><td>` + escape(entry.Reason) +
				`</td><td>` + escape(entry.Date) + `</td><td>` + escape(entry.By) + `</td></tr>`)
		}
		body.WriteString(`</tbody></table></div>`)
	}
	body.WriteString(`</details><hr class="main_hr">`)

	body.WriteString(`<h2>` + lang("clan_penalty_file") + `</h2><p><a href="/clan/penalty/download">` + lang("clan_penalty_download") + `</a></p>`)
	if page.Can_restore {
		body.WriteString(`<form method="post" action="/clan/penalty/restore" enctype="multipart/form-data" onsubmit="return confirm(this.dataset.confirm)" data-confirm="` +
			escape(lang("clan_penalty_restore_confirm")) + `">` + csrf +
			`<p>` + lang("clan_penalty_restore_help") + `</p><p><input type="file" name="penalty" accept=".json,application/json" required></p>` +
			`<p><button type="submit">` + lang("clan_penalty_restore") + `</button></p></form>`)
	}
	body.WriteString(`<hr class="main_hr">`)

	body.WriteString(`<details class="slbm-penalty-log-section"><summary><h2><span aria-hidden="true">▸ </span>` + lang("clan_penalty_log") + ` (` + strconv.Itoa(len(page.Log)) + `)</h2></summary>`)
	if len(page.Log) == 0 {
		body.WriteString(`<p>` + lang("clan_penalty_log_empty") + `</p>`)
	} else {
		body.WriteString(`<ul class="slbm-penalty-log">`)
		for _, item := range page.Log {
			action := escape(item.Action)
			if tool.Arr_in_str(Clan_penalty_actions, item.Action) {
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
	body.WriteString(`</details>`)
	return tool.Get_template(db, config, lang("clan_penalty"), body.String(), []any{}, [][]any{}, map[string]string{})
}
