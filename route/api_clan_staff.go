package route

import (
	"crypto/subtle"
	"database/sql"
	"errors"
	"net/url"
	"time"

	"opennamu/route/tool"
)

// Sus members may ban at most this many accounts per calendar month.
const Clan_sus_monthly_ban_limit = 3

type Clan_staff_member struct {
	ID       string
	Role     string
	Banned   bool
	Can_role bool
	Can_ban  bool
}

type Clan_staff_page struct {
	Caller_role  string
	Bans_left    int
	Members      []Clan_staff_member
	Role_choices []string
	Ban_limit    int
}

func clan_staff_month() string {
	return time.Now().Format("2006-01")
}

func clan_sus_bans_used(db tool.DB_runner, user_id string) int {
	count := 0
	tool.QueryRow_DB(db, "select count(*) from user_set where id = ? and name = 'slbm_sus_ban' and data like ?", []any{&count}, user_id, clan_staff_month()+"|%")
	return count
}

// clan_staff_can_manage reports whether the caller may change the target's
// role or ban state. Site admins are never managed from here, and sus
// members only manage regular members and managers.
func clan_staff_can_manage(caller string, caller_role string, target string, target_role string) bool {
	if target == caller || target_role == tool.Clan_role_admin {
		return false
	}
	switch caller_role {
	case tool.Clan_role_admin:
		return true
	case tool.Clan_role_sus:
		return target_role != tool.Clan_role_sus
	case tool.Clan_role_manager:
		return target_role == tool.Clan_role_guard
	}
	return false
}

func clan_staff_role_choices(caller_role string) []string {
	if caller_role == tool.Clan_role_admin {
		return []string{"", tool.Clan_role_manager, tool.Clan_role_guard, tool.Clan_role_sus}
	}
	if caller_role == tool.Clan_role_sus || caller_role == tool.Clan_role_manager {
		return []string{"", tool.Clan_role_guard}
	}
	return []string{}
}

func clan_staff_target_role(db *sql.DB, user_id string) string {
	if tool.Check_permission(db, "admin", user_id) {
		return tool.Clan_role_admin
	}
	return tool.Get_stored_clan_role(db, user_id)
}

func Api_clan_staff(config tool.Config) (Clan_staff_page, error) {
	page := Clan_staff_page{Ban_limit: Clan_sus_monthly_ban_limit}
	db := tool.DB_connect()
	defer tool.DB_close(db)
	page.Caller_role = tool.Get_clan_role(db, config.IP)
	if page.Caller_role != tool.Clan_role_admin && page.Caller_role != tool.Clan_role_sus && page.Caller_role != tool.Clan_role_manager {
		return page, errors.New("require auth")
	}
	page.Role_choices = clan_staff_role_choices(page.Caller_role)
	page.Bans_left = -1
	if page.Caller_role == tool.Clan_role_sus {
		page.Bans_left = max(0, Clan_sus_monthly_ban_limit-clan_sus_bans_used(db, config.IP))
	}

	rows := tool.Query_DB(db, "select distinct id from user_set where name = 'pw' order by id")
	for rows.Next() {
		member := Clan_staff_member{}
		if err := rows.Scan(&member.ID); err != nil {
			rows.Close()
			return page, err
		}
		page.Members = append(page.Members, member)
	}
	rows.Close()
	for index := range page.Members {
		member := &page.Members[index]
		member.Role = clan_staff_target_role(db, member.ID)
		member.Banned = tool.Auth_group_name_ban(tool.Get_user_auth(db, member.ID))
		member.Can_ban = clan_staff_can_manage(config.IP, page.Caller_role, member.ID, member.Role)
		member.Can_role = member.Can_ban && !member.Banned
	}
	return page, nil
}

// Api_clan_staff_post changes a member's staff role (action "role") or bans
// ("ban") / unbans ("unban") the member's account.
func Api_clan_staff_post(config tool.Config, values url.Values) error {
	db := tool.DB_connect()
	defer tool.DB_close(db)
	caller_role := tool.Get_clan_role(db, config.IP)
	if caller_role != tool.Clan_role_admin && caller_role != tool.Clan_role_sus && caller_role != tool.Clan_role_manager {
		return errors.New("require auth")
	}
	token, _ := config.Session.Get("clan_csrf").(string)
	if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(values.Get("csrf"))) != 1 {
		return errors.New("require auth")
	}

	user_id, action := values.Get("id"), values.Get("action")
	if User_value(db, user_id, "pw") == "" {
		return errors.New("not exist")
	}
	target_role := clan_staff_target_role(db, user_id)
	if !clan_staff_can_manage(config.IP, caller_role, user_id, target_role) {
		return errors.New("require auth")
	}

	switch action {
	case "role":
		role := values.Get("role")
		if !tool.Arr_in_str(clan_staff_role_choices(caller_role), role) {
			return errors.New("require auth")
		}
		if tool.Auth_group_name_ban(tool.Get_user_auth(db, user_id)) {
			return errors.New("error")
		}
		return tool.DB_transaction(db, func(tx *sql.Tx) error {
			if _, err := tx.Exec(tool.DB_change("delete from user_set where id = ? and name = 'slbm_clan_role'"), user_id); err != nil {
				return err
			}
			if role != "" {
				if _, err := tx.Exec(tool.DB_change("insert into user_set (id, name, data) values (?, 'slbm_clan_role', ?)"), user_id, role); err != nil {
					return err
				}
			}
			tool.Do_insert_auth_history(tx, config.IP, "clan_role ("+user_id+") "+target_role+" -> "+role)
			return nil
		})
	case "ban", "unban":
		// Do not spend a sus member's monthly ban on an account that is already banned.
		if tool.Auth_group_name_ban(tool.Get_user_auth(db, user_id)) == (action == "ban") {
			return errors.New("error")
		}
		return tool.DB_transaction(db, func(tx *sql.Tx) error {
			if action == "ban" && caller_role == tool.Clan_role_sus {
				if clan_sus_bans_used(tx, config.IP) >= Clan_sus_monthly_ban_limit {
					return errors.New("clan_ban_limit")
				}
				if _, err := tx.Exec(tool.DB_change("insert into user_set (id, name, data) values (?, 'slbm_sus_ban', ?)"), config.IP, clan_staff_month()+"|"+user_id+"|"+tool.Get_time()); err != nil {
					return err
				}
			}
			if err := Clan_ban_tx(tx, config, user_id, []string{user_id}, true, action == "unban"); err != nil {
				return err
			}
			tool.Do_insert_auth_history(tx, config.IP, "clan_staff_"+action+" ("+user_id+")")
			return nil
		})
	}
	return errors.New("error")
}
