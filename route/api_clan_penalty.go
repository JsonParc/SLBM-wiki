package route

import (
	"crypto/subtle"
	stdjson "encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"opennamu/route/tool"
)

// Clan penalty records live in a JSON file so they can be downloaded and
// restored on their own. Every change is appended to the file's log.

const clan_penalty_version = 1
const clan_penalty_max_points = 1000
const clan_penalty_max_name = 100
const clan_penalty_max_reason = 1000
const Clan_penalty_max_upload = 16 * 1024 * 1024

var clan_penalty_lock sync.Mutex

type Clan_penalty_entry struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Points int    `json:"points"`
	Reason string `json:"reason"`
	Date   string `json:"date"`
	By     string `json:"by"`
}

type Clan_penalty_log struct {
	ID       string `json:"id"`
	Time     string `json:"time"`
	By       string `json:"by"`
	Action   string `json:"action"`
	Entry_id string `json:"entry_id,omitempty"`
	Name     string `json:"name,omitempty"`
	Points   int    `json:"points,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

// Clan_penalty_actions lists the log actions this version writes.
var Clan_penalty_actions = []string{"add", "increase", "decrease", "delete", "restore"}

type Clan_penalty_file struct {
	Version int                  `json:"version"`
	Entries []Clan_penalty_entry `json:"entries"`
	Log     []Clan_penalty_log   `json:"log"`
}

type Clan_penalty_row struct {
	Clan_penalty_entry
	Total int
}

type Clan_penalty_page struct {
	Rows        []Clan_penalty_row
	Log         []Clan_penalty_log
	Can_restore bool
}

func Clan_penalty_path() string {
	return filepath.Join("data", "clan_penalty.json")
}

func Clan_penalty_staff(role string) bool {
	return role == tool.Clan_role_admin || role == tool.Clan_role_sus || role == tool.Clan_role_manager
}

func Clan_penalty_can_restore(role string) bool {
	return role == tool.Clan_role_admin || role == tool.Clan_role_sus
}

// Clan_penalty_parse validates a penalty file and fills missing slices.
func Clan_penalty_parse(raw []byte) (Clan_penalty_file, error) {
	data := Clan_penalty_file{}
	if err := stdjson.Unmarshal(raw, &data); err != nil || data.Version != clan_penalty_version {
		return data, errors.New("clan_penalty_invalid")
	}
	if data.Entries == nil {
		data.Entries = []Clan_penalty_entry{}
	}
	if data.Log == nil {
		data.Log = []Clan_penalty_log{}
	}
	seen := map[string]bool{}
	for _, entry := range data.Entries {
		// Stored entries may be negative: a decrease is recorded as its own entry.
		if entry.ID == "" || seen[entry.ID] || entry.Points == 0 || clan_penalty_check(entry.Name, max(entry.Points, -entry.Points), entry.Reason) != nil {
			return data, errors.New("clan_penalty_invalid")
		}
		seen[entry.ID] = true
	}
	return data, nil
}

func clan_penalty_check(name string, points int, reason string) error {
	if name == "" || utf8.RuneCountInString(name) > clan_penalty_max_name ||
		points < 1 || points > clan_penalty_max_points ||
		reason == "" || utf8.RuneCountInString(reason) > clan_penalty_max_reason {
		return errors.New("clan_penalty_input")
	}
	return nil
}

// clan_penalty_read must be called with clan_penalty_lock held.
func clan_penalty_read() (Clan_penalty_file, error) {
	raw, err := os.ReadFile(Clan_penalty_path())
	if os.IsNotExist(err) {
		return Clan_penalty_file{Version: clan_penalty_version, Entries: []Clan_penalty_entry{}, Log: []Clan_penalty_log{}}, nil
	}
	if err != nil {
		return Clan_penalty_file{}, err
	}
	return Clan_penalty_parse(raw)
}

// Clan_penalty_write atomically replaces the penalty file. Callers must hold
// clan_penalty_lock.
func Clan_penalty_write(raw []byte) error {
	if err := os.MkdirAll(filepath.Dir(Clan_penalty_path()), 0o755); err != nil {
		return err
	}
	temp := Clan_penalty_path() + ".tmp"
	if err := os.WriteFile(temp, raw, 0o644); err != nil {
		return err
	}
	return os.Rename(temp, Clan_penalty_path())
}

func clan_penalty_save(data Clan_penalty_file) error {
	raw, err := stdjson.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return Clan_penalty_write(raw)
}

func clan_penalty_role(config tool.Config) string {
	db := tool.DB_connect()
	defer tool.DB_close(db)
	return tool.Get_clan_role(db, config.IP)
}

func clan_penalty_log(config tool.Config, action string, entry Clan_penalty_entry) Clan_penalty_log {
	return Clan_penalty_log{
		ID: tool.Get_random_key(16), Time: tool.Get_time(), By: config.IP, Action: action,
		Entry_id: entry.ID, Name: entry.Name, Points: entry.Points, Reason: entry.Reason,
	}
}

// Api_clan_penalty_list returns entries sorted by "name", "points" (total per
// name), or "recent", plus the change log newest first.
func Api_clan_penalty_list(config tool.Config, sort_by string) (Clan_penalty_page, error) {
	page := Clan_penalty_page{}
	role := clan_penalty_role(config)
	if !Clan_penalty_staff(role) {
		return page, errors.New("require auth")
	}
	page.Can_restore = Clan_penalty_can_restore(role)

	clan_penalty_lock.Lock()
	data, err := clan_penalty_read()
	clan_penalty_lock.Unlock()
	if err != nil {
		return page, err
	}

	totals := map[string]int{}
	for _, entry := range data.Entries {
		totals[entry.Name] += entry.Points
	}
	// Newest first, so entries added within the same second keep a stable order.
	for index := len(data.Entries) - 1; index >= 0; index-- {
		entry := data.Entries[index]
		page.Rows = append(page.Rows, Clan_penalty_row{entry, totals[entry.Name]})
	}
	sort.SliceStable(page.Rows, func(i, j int) bool {
		left, right := page.Rows[i], page.Rows[j]
		switch sort_by {
		case "name":
			if left.Name != right.Name {
				return left.Name < right.Name
			}
		case "points":
			if left.Total != right.Total {
				return left.Total > right.Total
			}
			if left.Name != right.Name {
				return left.Name < right.Name
			}
		}
		return left.Date > right.Date
	})

	page.Log = make([]Clan_penalty_log, len(data.Log))
	for index, item := range data.Log {
		page.Log[len(data.Log)-1-index] = item
	}
	return page, nil
}

// Api_clan_penalty_post adds (action "add") or deletes (action "delete") an entry.
func Api_clan_penalty_post(config tool.Config, values url.Values) error {
	if !Clan_penalty_staff(clan_penalty_role(config)) {
		return errors.New("require auth")
	}
	token, _ := config.Session.Get("clan_csrf").(string)
	if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(values.Get("csrf"))) != 1 {
		return errors.New("require auth")
	}

	clan_penalty_lock.Lock()
	defer clan_penalty_lock.Unlock()
	data, err := clan_penalty_read()
	if err != nil {
		return err
	}

	switch values.Get("action") {
	case "add":
		points, err := strconv.Atoi(strings.TrimSpace(values.Get("points")))
		if err != nil {
			return errors.New("clan_penalty_input")
		}
		entry := Clan_penalty_entry{
			ID: tool.Get_random_key(16), Name: strings.TrimSpace(values.Get("name")), Points: points,
			Reason: strings.TrimSpace(values.Get("reason")), Date: tool.Get_time(), By: config.IP,
		}
		if err := clan_penalty_check(entry.Name, entry.Points, entry.Reason); err != nil {
			return err
		}
		data.Entries = append(data.Entries, entry)
		data.Log = append(data.Log, clan_penalty_log(config, "add", entry))
	case "increase", "decrease":
		// Adjust an existing name by one point; the change is its own entry.
		name := values.Get("name")
		total, exists := 0, false
		for _, entry := range data.Entries {
			if entry.Name == name {
				total += entry.Points
				exists = true
			}
		}
		if !exists {
			return errors.New("not exist")
		}
		points := 1
		if values.Get("action") == "decrease" {
			if total <= 0 {
				return errors.New("clan_penalty_zero")
			}
			points = -1
		}
		entry := Clan_penalty_entry{
			ID: tool.Get_random_key(16), Name: name, Points: points,
			Reason: strings.TrimSpace(values.Get("reason")), Date: tool.Get_time(), By: config.IP,
		}
		if err := clan_penalty_check(entry.Name, 1, entry.Reason); err != nil {
			return err
		}
		data.Entries = append(data.Entries, entry)
		data.Log = append(data.Log, clan_penalty_log(config, values.Get("action"), entry))
	case "delete":
		id := values.Get("id")
		index := -1
		for i, entry := range data.Entries {
			if entry.ID == id {
				index = i
			}
		}
		if index < 0 {
			return errors.New("not exist")
		}
		data.Log = append(data.Log, clan_penalty_log(config, "delete", data.Entries[index]))
		data.Entries = append(data.Entries[:index], data.Entries[index+1:]...)
	default:
		return errors.New("error")
	}
	return clan_penalty_save(data)
}

func Api_clan_penalty_export(config tool.Config) ([]byte, error) {
	if !Clan_penalty_staff(clan_penalty_role(config)) {
		return nil, errors.New("require auth")
	}
	clan_penalty_lock.Lock()
	data, err := clan_penalty_read()
	clan_penalty_lock.Unlock()
	if err != nil {
		return nil, err
	}
	return stdjson.MarshalIndent(data, "", "  ")
}

// Api_clan_penalty_import replaces the entries with an uploaded file. The log
// keeps both histories and records the restore itself.
func Api_clan_penalty_import(config tool.Config, csrf string, raw []byte) error {
	if !Clan_penalty_can_restore(clan_penalty_role(config)) {
		return errors.New("require auth")
	}
	token, _ := config.Session.Get("clan_csrf").(string)
	if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(csrf)) != 1 {
		return errors.New("require auth")
	}
	uploaded, err := Clan_penalty_parse(raw)
	if err != nil {
		return err
	}

	clan_penalty_lock.Lock()
	defer clan_penalty_lock.Unlock()
	current, err := clan_penalty_read()
	if err != nil {
		return err
	}
	log := uploaded.Log
	known := map[string]bool{}
	for _, item := range log {
		known[item.ID] = true
	}
	for _, item := range current.Log {
		if !known[item.ID] {
			log = append(log, item)
		}
	}
	sort.SliceStable(log, func(i, j int) bool { return log[i].Time < log[j].Time })
	log = append(log, Clan_penalty_log{
		ID: tool.Get_random_key(16), Time: tool.Get_time(), By: config.IP, Action: "restore",
		Points: len(uploaded.Entries),
	})
	return clan_penalty_save(Clan_penalty_file{Version: clan_penalty_version, Entries: uploaded.Entries, Log: log})
}
