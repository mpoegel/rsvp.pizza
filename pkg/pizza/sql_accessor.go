package pizza

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLAccessor struct {
	db *sql.DB
}

func NewSQLAccessor(dbfile string) (*SQLAccessor, error) {
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		return nil, err
	}
	a := &SQLAccessor{db}
	return a, a.PatchTables()
}

func (a *SQLAccessor) Close() {
	a.db.Close()
}

func (a *SQLAccessor) CreateTables() error {
	stmt := "CREATE TABLE friends (email text NOT NULL PRIMARY KEY, name text)"
	if _, err := a.db.Exec(stmt); err != nil {
		return err
	}
	stmt = "CREATE TABLE fridays (start_time datetime NOT NULL PRIMARY KEY)"
	if _, err := a.db.Exec(stmt); err != nil {
		return err
	}
	stmt = "CREATE TABLE versions (name text NOT NULL PRIMARY KEY, version int NOT NULL)"
	if _, err := a.db.Exec(stmt); err != nil {
		return err
	}
	_, err := a.db.Exec(`INSERT INTO app_versions (name, version) VALUES ('schema', 2)`)
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

func (a *SQLAccessor) IsFriendAllowed(email string) (bool, error) {
	stmt, err := a.db.Prepare("select count(email) from friends where email = ?")
	if err != nil {
		return false, err
	}
	count := 0
	err = stmt.QueryRow(email).Scan(&count)
	if err != nil {
		return false, nil
	}
	return count == 1, nil
}

func (a *SQLAccessor) GetFriendName(email string) (string, error) {
	stmt, err := a.db.Prepare("select name from friends where email = ?")
	if err != nil {
		return "", err
	}
	var name string
	err = stmt.QueryRow(email).Scan(&name)
	return name, err
}

func (a *SQLAccessor) GetUpcomingFridays(daysAhead int) ([]time.Time, error) {
	return a.GetUpcomingFridaysAfter(time.Now(), daysAhead)
}

func (a *SQLAccessor) GetUpcomingFridaysAfter(after time.Time, daysAhead int) ([]time.Time, error) {
	before := after.AddDate(0, 0, daysAhead)
	stmt, err := a.db.Prepare("select start_time from fridays where start_time <= ? and start_time >= ?")
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(before, after)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]time.Time, 0)
	for rows.Next() {
		var friday time.Time
		err = rows.Scan(&friday)
		if err != nil {
			return nil, err
		}
		result = append(result, friday)
	}
	return result, nil
}

func (a *SQLAccessor) AddFriend(email, name string) error {
	stmt, err := a.db.Prepare("insert into friends (email, name) values (?, ?)")
	if err != nil {
		return nil
	}
	_, err = stmt.Exec(email, name)
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

func (a *SQLAccessor) ListFriends() ([]Friend, error) {
	stmt := "select email, name from friends"
	rows, err := a.db.Query(stmt)
	if err != nil {
		return nil, err
	}
	res := make([]Friend, 0)
	for rows.Next() {
		f := Friend{}
		err = rows.Scan(&f.Email, &f.Name)
		if err != nil {
			return nil, err
		}
		res = append(res, f)
	}
	return res, nil
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

func (a *SQLAccessor) RemoveFriend(email string) error {
	stmt, err := a.db.Prepare("delete from friends where email = ?")
	if err != nil {
		return nil
	}
	_, err = stmt.Exec(email)
	return err
}

func (a *SQLAccessor) RemoveFriday(date time.Time) error {
	stmt, err := a.db.Prepare("delete from fridays where start_time = ?")
	if err != nil {
		return nil
	}
	_, err = stmt.Exec(date)
	return err
}
