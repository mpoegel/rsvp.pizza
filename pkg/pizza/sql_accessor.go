package pizza

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLAccessor struct {
	db *sql.DB
}

func NewSQLAccessor(dbfile string, skipPatch bool) (*SQLAccessor, error) {
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		return nil, err
	}
	a := &SQLAccessor{db}
	if skipPatch {
		return a, nil
	}
	return a, a.PatchTables()
}

func (a *SQLAccessor) Close() {
	a.db.Close()
}

func (a *SQLAccessor) CreateTables() error {
	stmt := `CREATE TABLE friends (
		id          integer PRIMARY KEY AUTOINCREMENT,
		email       text NOT NULL UNIQUE,
		name        text,
		preferences text default "{}"
	)`
	if _, err := a.db.Exec(stmt); err != nil {
		return err
	}
	stmt = `CREATE TABLE fridays (
		start_time    datetime NOT NULL PRIMARY KEY,
		invited_group text,
		details       text,
		invited       text default "[]",
		max_guests    int default 10,
		enabled       bool default true
	)`
	if _, err := a.db.Exec(stmt); err != nil {
		return err
	}
	stmt = `CREATE TABLE versions (
		name    text NOT NULL PRIMARY KEY,
		version int NOT NULL
	)`
	if _, err := a.db.Exec(stmt); err != nil {
		return err
	}
	stmt = `CREATE TABLE app_versions (
		name    text NOT NULL PRIMARY KEY,
		version int NOT NULL
	)`
	if _, err := a.db.Exec(stmt); err != nil {
		return err
	}
	_, err := a.db.Exec(`INSERT INTO app_versions (name, version) VALUES ('schema', 6)`)
	return err
}

func (a *SQLAccessor) PatchTables() error {
	var schemaVersion int
	if err := a.db.QueryRow("SELECT version FROM app_versions WHERE name='schema'").Scan(&schemaVersion); err != nil {
		// minimum patch 2 is required
		return err
	}
	// apply missing patches
	for p := schemaVersion + 1; p < len(AllPatches); p++ {
		if err := AllPatches[p](a); err != nil {
			return err
		}
	}
	// save new schema version
	stmt, err := a.db.Prepare("UPDATE app_versions SET version=? WHERE name='schema'")
	if err != nil {
		return nil
	}
	_, err = stmt.Exec(len(AllPatches) - 1)
	return err
}

func (a *SQLAccessor) GetFriendByID(ID string) (Friend, error) {
	friend := Friend{}
	stmt, err := a.db.Prepare("select email, name from friends where id = ?")
	if err != nil {
		return friend, err
	}
	err = stmt.QueryRow(ID).Scan(&friend.Email, &friend.Name)
	friend.ID = ID
	return friend, err
}

func (a *SQLAccessor) GetFriendByEmail(email string) (Friend, error) {
	friend := Friend{}
	stmt, err := a.db.Prepare("select id, name from friends where email = ?")
	if err != nil {
		return friend, err
	}
	var id int64
	err = stmt.QueryRow(email).Scan(&id, &friend.Name)
	friend.ID = strconv.FormatInt(id, 10)
	friend.Email = email
	return friend, err
}

func (a *SQLAccessor) GetUpcomingFridays(daysAhead int) ([]Friday, error) {
	return a.GetUpcomingFridaysAfter(time.Now(), daysAhead)
}

func (a *SQLAccessor) GetUpcomingFridaysAfter(after time.Time, daysAhead int) ([]Friday, error) {
	before := after.AddDate(0, 0, daysAhead)
	stmt, err := a.db.Prepare(`SELECT start_time, invited_group, details, invited, max_guests, enabled FROM fridays 
		WHERE start_time <= ? AND start_time >= ?`)
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(before, after)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Friday, 0)
	for rows.Next() {
		var friday Friday
		var rawInvited string
		err = rows.Scan(&friday.Date, &friday.Group, &friday.Details, &rawInvited, &friday.MaxGuests, &friday.Enabled)
		if err != nil {
			return nil, err
		}
		if err = json.Unmarshal([]byte(rawInvited), &friday.Guests); err != nil {
			return nil, err
		}
		result = append(result, friday)
	}
	return result, nil
}

func (a *SQLAccessor) AddFriend(email, name string) error {
	stmt, err := a.db.Prepare("INSERT INTO friends (email, name) VALUES (?, ?) ON CONFLICT (email) DO UPDATE SET name=?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(email, name, name)
	return err
}

func (a *SQLAccessor) DoesFridayExist(date time.Time) (bool, error) {
	stmt, err := a.db.Prepare("SELECT COUNT(*) FROM fridays WHERE start_time = ?")
	if err != nil {
		return false, err
	}
	rows, err := stmt.Query(date)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	var count int
	if !rows.Next() {
		return false, nil
	}
	if err = rows.Scan(&count); err != nil {
		return false, err
	}
	return count == 1, nil
}

func (a *SQLAccessor) AddFriday(date time.Time) error {
	exists, err := a.DoesFridayExist(date)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	stmt, err := a.db.Prepare("insert into fridays (start_time) values (?)")
	if err != nil {
		return nil
	}
	_, err = stmt.Exec(date)
	return err
}

func (a *SQLAccessor) ListFridays() ([]Friday, error) {
	stmt := "select start_time from fridays"
	rows, err := a.db.Query(stmt)
	if err != nil {
		return nil, err
	}
	loc, _ := time.LoadLocation("America/New_York")
	res := make([]Friday, 0)
	for rows.Next() {
		f := Friday{}
		err = rows.Scan(&f.Date)
		f.Date = f.Date.In(loc)
		if err != nil {
			return nil, err
		}
		res = append(res, f)
	}
	return res, nil
}

func (a *SQLAccessor) GetFriday(date time.Time) (Friday, error) {
	stmt, err := a.db.Prepare("select start_time, invited_group, details, invited, max_guests, enabled from fridays where start_time = ?")
	if err != nil {
		return Friday{}, err
	}
	var friday Friday
	var rawInvited string
	err = stmt.QueryRow(date).Scan(&friday.Date, &friday.Group, &friday.Details, &rawInvited, &friday.MaxGuests, &friday.Enabled)
	if err != nil {
		return friday, err
	}
	err = json.Unmarshal([]byte(rawInvited), &friday.Guests)
	return friday, err
}

func (a *SQLAccessor) RemoveFriday(date time.Time) error {
	stmt, err := a.db.Prepare("delete from fridays where start_time = ?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(date)
	return err
}

func (a *SQLAccessor) UpdateFriday(friday Friday) error {
	stmt, err := a.db.Prepare("UPDATE fridays SET invited_group=?, details=?, max_guests=?, enabled=? WHERE start_time=?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(friday.Group, friday.Details, friday.MaxGuests, friday.Enabled, friday.Date)
	return err
}

func (a *SQLAccessor) AddFriendToFriday(email string, friday Friday) error {
	var invited []string
	stmt, err := a.db.Prepare("SELECT invited FROM fridays WHERE start_time = ?")
	if err != nil {
		return err
	}
	var rawInvited string
	err = stmt.QueryRow(friday.Date).Scan(&rawInvited)
	if err != nil {
		return err
	}
	if err = json.Unmarshal([]byte(rawInvited), &invited); err != nil {
		return err
	}
	// ensure uniqueness
	for _, guest := range invited {
		if guest == email {
			// already invited
			return nil
		}
	}
	// not invited yet
	stmt, err = a.db.Prepare("UPDATE fridays SET invited = json_insert(invited, '$[#]', ?) WHERE start_time = ?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(email, friday.Date)
	return err
}

func (a *SQLAccessor) GetPreferences(email string) (Preferences, error) {
	var prefs Preferences
	stmt, err := a.db.Prepare("SELECT preferences FROM friends WHERE email=?")
	if err != nil {
		return prefs, err
	}
	var rawPreferences string
	err = stmt.QueryRow(email).Scan(&rawPreferences)
	if err != nil {
		return prefs, err
	}
	err = json.Unmarshal([]byte(rawPreferences), &prefs)
	return prefs, err
}

func (a *SQLAccessor) SetPreferences(email string, prefs Preferences) error {
	rawPrefs, err := json.Marshal(prefs)
	if err != nil {
		return err
	}
	stmt, err := a.db.Prepare("UPDATE friends SET preferences=? WHERE email=?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(rawPrefs, email)
	return err
}
