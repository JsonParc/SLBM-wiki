package route

import (
	"errors"
	"opennamu/route/tool"
	"strings"
)

func Api_clan_document_target(config tool.Config, title string) (string, error) {
	db := tool.DB_connect()
	defer tool.DB_close(db)
	if tool.IP_or_user(config.IP) || !tool.Get_user_set_exists(db, config.IP, "pw") {
		return "", errors.New("auth")
	}
	if strings.TrimSpace(title) == "" {
		return "", errors.New("empty title")
	}
	if !tool.Do_title_length_check(db, title, "document") {
		return "", errors.New("title length")
	}
	if !tool.Check_acl(db, title, "", "document_make_acl", config.IP) || !tool.Check_acl(db, title, "", "document_edit", config.IP) {
		return "", errors.New("auth")
	}
	var existing string
	if tool.QueryRow_DB(db, "select title from data where title = ?", []any{&existing}, title) {
		return "/w/" + tool.Url_parser(title), nil
	}
	return "/edit/" + tool.Url_parser(title), nil
}
