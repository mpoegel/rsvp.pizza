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
	return &SQLAccessor{db}, nil
}

func (a *SQLAccessor) Close() {
	a.db.Close()
}

func (a *SQLAccessor) CreateTables() error {
	stmt := "create table friends (email text not null primary key, name text)"
	if _, err := a.db.Exec(stmt); err != nil {
		return err
	}
	stmt = "create table fridays (start_time datetime)"
	if _, err := a.db.Exec(stmt); err != nil {
		return err
	}
	return nil
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

func (a *SQLAccessor) AddFriday(date time.Time) error {
	stmt, err := a.db.Prepare("insert into fridays (start_time) values (?)")
	if err != nil {
		return nil
	}
	_, err = stmt.Exec(date)
	return err
}
