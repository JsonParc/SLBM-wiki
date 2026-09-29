package tool

import (
	"database/sql"
	_ "embed"
)

//go:embed slbm-frontpage.namumark
var slbm_frontpage []byte

func slbm_seed_setting(db *sql.DB, name string, value string) {
	var existing string
	if QueryRow_DB(db, "select data from other where name = ? and coverage = '' limit 1", []any{&existing}, name) {
		return
	}
	Exec_DB(db, "insert into other(name,data,coverage) values (?,?,'')", name, value)
}

func Ensure_slbm_defaults(db *sql.DB) {
	var frontpage string
	if !QueryRow_DB(db, "select data from data where title = ? limit 1", []any{&frontpage}, "SLBM: 대문") {
		Exec_DB(db, "insert into data(title,data) values (?,?)", "SLBM: 대문", string(slbm_frontpage))
	}
	Exec_DB(db, "update other set data = ? where name = 'frontpage' and coverage = '' and (data = '' or data = 'FrontPage')", "SLBM: 대문")
	Exec_DB(db, "update other set data = ? where name = 'name' and coverage = '' and (data = '' or data = 'Wiki')", "SLBM 클랜 위키")

	slbm_seed_setting(db, "frontpage_type", "document")
	slbm_seed_setting(db, "frontpage", "SLBM: 대문")
	slbm_seed_setting(db, "name", "SLBM 클랜 위키")
	slbm_seed_setting(db, "requires_approval", "on")
	slbm_seed_setting(db, "approval_question", "클랜 닉네임과 가입 사유를 입력해 주세요.")
	slbm_seed_setting(db, "application_expiration_date", "0")
	slbm_seed_setting(db, "application_expiration_action", "")
	slbm_seed_setting(db, "slbm_members_only", "on")
	slbm_seed_setting(db, "load_ip_select", "REMOTE_ADDR")

	var captcha string
	if !QueryRow_DB(db, "select acl from alist where name = ? and acl = ? limit 1", []any{&captcha}, "ip", "captcha_pass") {
		Exec_DB(db, "insert into alist(name,acl) values (?,?)", "ip", "captcha_pass")
	}
}
