package route

import (
	"crypto/subtle"
	"database/sql"
	"errors"
	"net"
	"net/url"
	"opennamu/route/tool"
)

type Clan_user struct {
	ID        string
	Date      string
	Group     string
	IPs       []Clan_ip
	Protected bool
}
type Clan_ip struct {
	Address string
	Group   string
}

func Clan_admin(db *sql.DB, config tool.Config) bool {
	return !tool.IP_or_user(config.IP) && tool.Check_permission(db, "admin", config.IP)
}

func Api_clan_users(config tool.Config) ([]Clan_user, error) {
	db := tool.DB_connect()
	defer tool.DB_close(db)
	if !Clan_admin(db, config) {
		return nil, errors.New("require auth")
	}
	rows := tool.Query_DB(db, "select distinct id from user_set where name = 'pw' order by id")
	users := []Clan_user{}
	for rows.Next() {
		var user Clan_user
		if err := rows.Scan(&user.ID); err != nil {
			rows.Close()
			return nil, err
		}
		users = append(users, user)
	}
	rows.Close()
	for index := range users {
		user := &users[index]
		user.Date = User_value(db, user.ID, "date")
		user.Group = tool.Get_user_auth(db, user.ID)
		user.Protected = tool.Check_permission(db, "admin", user.ID) || user.ID == config.IP
		ip_rows := tool.Query_DB(db, "select distinct ip from ua_d where name = ? order by ip", user.ID)
		for ip_rows.Next() {
			var address string
			if ip_rows.Scan(&address) == nil && net.ParseIP(address) != nil {
				user.IPs = append(user.IPs, Clan_ip{Address: address})
			}
		}
		ip_rows.Close()
		for i := range user.IPs {
			user.IPs[i].Group = tool.Get_user_auth(db, user.IPs[i].Address)
		}
	}
	return users, nil
}

func Api_clan_user_post(config tool.Config, values url.Values) error {
	db := tool.DB_connect()
	defer tool.DB_close(db)
	if !Clan_admin(db, config) {
		return errors.New("require auth")
	}
	token, _ := config.Session.Get("clan_csrf").(string)
	if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(values.Get("csrf"))) != 1 {
		return errors.New("require auth")
	}
	user_id, action, target, address := values.Get("id"), values.Get("action"), values.Get("target"), values.Get("ip")
	if User_value(db, user_id, "pw") == "" {
		return errors.New("not exist")
	}
	if user_id == config.IP || tool.Check_permission(db, "admin", user_id) {
		return errors.New("require auth")
	}
	if action != "ban" && action != "unban" && action != "delete" && action != "reset" {
		return errors.New("error")
	}
	targets := []string{}
	if action == "ban" || action == "unban" {
		if target != "account" && target != "ip" && target != "both" {
			return errors.New("error")
		}
		if target == "account" || target == "both" {
			targets = append(targets, user_id)
		}
		if target == "ip" || target == "both" {
			var known string
			if net.ParseIP(address) == nil || !tool.QueryRow_DB(db, "select ip from ua_d where name = ? and ip = ? limit 1", []any{&known}, user_id, address) {
				return errors.New("error")
			}
			targets = append(targets, address)
		}
	}
	password_hash := ""
	if action == "reset" {
		password := values.Get("password")
		minimum := max(12, tool.Str_to_int(User_other(db, "password_min_length")))
		if tool.Get_len(password) < minimum || len(password) > 1024 || password == user_id {
			return errors.New("password too short")
		}
		password_hash = tool.Password_encode(db, password, tool.Get_user_encode(db, user_id))
	}
	if action == "delete" && values.Get("confirm") != user_id {
		return errors.New("error")
	}
	return tool.DB_transaction(db, func(tx *sql.Tx) error {
		switch action {
		case "ban", "unban":
			if action == "ban" && (target == "account" || target == "both") {
				if _, err := tx.Exec(tool.DB_change("delete from user_set where id = ? and name = 'slbm_session_epoch'"), user_id); err != nil {
					return err
				}
				if _, err := tx.Exec(tool.DB_change("insert into user_set (id, name, data) values (?, 'slbm_session_epoch', ?)"), user_id, tool.Get_random_key(32)); err != nil {
					return err
				}
			}
			for _, subject := range targets {
				if action == "unban" {
					var group string
					tool.QueryRow_DB(tx, "select data from user_set where id = ? and name = 'acl'", []any{&group}, subject)
					if !tool.Auth_group_name_ban(group) {
						continue
					}
				}
				tool.Do_auth_insert(tx, subject, "", "SLBM user management", "ban_without_site", config.IP, "", action == "unban")
			}
		case "delete":
			if _, err := tx.Exec(tool.DB_change("delete from user_set where id = ?"), user_id); err != nil {
				return err
			}
			// Reserve the ID so deleted identities cannot be impersonated by registering again.
			if _, err := tx.Exec(tool.DB_change("insert into user_set (id, name, data) values (?, 'slbm_deleted', '1')"), user_id); err != nil {
				return err
			}
		case "reset":
			if _, err := tx.Exec(tool.DB_change("update user_set set data = ? where id = ? and name = 'pw'"), password_hash, user_id); err != nil {
				return err
			}
		}
		tool.Do_insert_auth_history(tx, config.IP, "clan_"+action+" ("+user_id+") "+target+" "+address)
		return nil
	})
}
