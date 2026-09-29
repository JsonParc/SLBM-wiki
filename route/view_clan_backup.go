package route

import (
	"crypto/subtle"
	"strconv"
	"strings"

	"opennamu/route/tool"
)

// View_clan_backup renders the backup page. With restore set, raw holds the
// uploaded zip and csrf the submitted token.
func View_clan_backup(config tool.Config, restore bool, csrf string, raw []byte, upload_error string) string {
	db := tool.DB_connect()
	defer tool.DB_close(db)
	if !Clan_admin(db, config) {
		return tool.Get_error_page(db, config, "auth")
	}
	lang := func(key string) string { return tool.Get_language(db, key, true) }
	escape := tool.HTML_escape

	token, _ := config.Session.Get("clan_csrf").(string)
	if token == "" {
		token = tool.Get_random_key(48)
		config.Session.Set("clan_csrf", token)
		if config.Session.Save() != nil {
			return tool.Get_error_page(db, config, "error")
		}
	}

	var notice strings.Builder
	if upload_error != "" {
		notice.WriteString(`<p class="slbm-backup-error">` + lang(upload_error) + `</p>`)
	} else if restore {
		if subtle.ConstantTimeCompare([]byte(token), []byte(csrf)) != 1 {
			return tool.Get_error_page(db, config, "auth")
		}
		result, err := Api_clan_backup_restore(config, raw)
		if err != nil {
			message := err.Error()
			if message != "auth" && message != "clan_backup_invalid" && message != "clan_backup_sqlite_only" {
				message = "clan_backup_failed"
			}
			notice.WriteString(`<p class="slbm-backup-error">` + lang(message) + `</p>`)
		} else {
			notice.WriteString(`<p class="slbm-backup-done">` + lang("clan_backup_restored") + `</p><ul>` +
				`<li>` + lang("clan_backup_documents") + `: ` + strconv.Itoa(result.Documents) + `</li>` +
				`<li>` + lang("clan_backup_images") + `: ` + strconv.Itoa(result.Images) + `</li>` +
				`<li>` + lang("clan_backup_snapshot") + `: ` + escape(result.Snapshot) + `</li>`)
			if len(result.Skipped) > 0 {
				notice.WriteString(`<li>` + lang("clan_backup_skipped") + `: ` + escape(strings.Join(result.Skipped, ", ")) + `</li>`)
			}
			notice.WriteString(`</ul>`)
		}
	}

	body := notice.String() +
		`<h2>` + lang("clan_backup_download") + `</h2>` +
		`<p>` + lang("clan_backup_download_help") + `</p>` +
		`<p><a class="slbm-backup-button" href="/clan/backup/download">` + lang("clan_backup_download") + `</a></p>` +
		`<hr class="main_hr">` +
		`<h2>` + lang("clan_backup_restore") + `</h2>`
	if tool.Check_permission(db, "owner", config.IP) {
		body += `<p>` + lang("clan_backup_restore_help") + `</p>` +
			`<form method="post" enctype="multipart/form-data" onsubmit="return confirm(this.dataset.confirm)" data-confirm="` + escape(lang("clan_backup_restore_confirm")) + `">` +
			`<input type="hidden" name="csrf" value="` + escape(token) + `">` +
			`<p><input type="file" name="backup" accept=".zip,application/zip" required></p>` +
			`<p><label><input type="checkbox" name="agree" value="1" required> ` + lang("clan_backup_restore_agree") + `</label></p>` +
			`<p><button type="submit">` + lang("clan_backup_restore") + `</button></p></form>`
	} else {
		body += `<p>` + lang("clan_backup_owner_only") + `</p>`
	}

	return tool.Get_template(db, config, lang("clan_backup"), body, []any{}, [][]any{{"manager", lang("return")}}, map[string]string{})
}
