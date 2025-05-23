package pizza

import (
	"flag"
	"log/slog"
	"os"
)

var AllPatches []func(*SQLAccessor) error

func Patch(args []string) {
	fs := flag.NewFlagSet("patch", flag.ExitOnError)
	isInit := fs.Bool("init", false, "initialize all tables")
	_ = fs.Bool("drop", false, "drop all tables")
	n := fs.Int("n", 0, "patch number")
	fs.Parse(args)

	config := LoadConfigEnv()
	var accessor *SQLAccessor
	var err error
	accessor, err = NewSQLAccessor(config.DBFile, true)
	if err != nil {
		slog.Error("sql accessor init failure", "error", err)
		os.Exit(1)
	}

	if *isInit {
		accessor.CreateTables()
	}

	switch *n {
	case 1:
		err = Patch001(accessor)
	case 2:
		err = Patch002(accessor)
	case 3:
		err = Patch003(accessor)
	case 4:
		err = Patch004(accessor)
	case 5:
		err = Patch005(accessor)
	case 6:
		err = Patch006(accessor)
	}

	if err != nil {
		slog.Error("patch failed", "n", *n, "error", err)
	}
}

func Patch001(a *SQLAccessor) error {
	// fridays should be unique
	stmt := `CREATE TABLE IF NOT EXISTS fridays_new (start_time datetime NOT NULL PRIMARY KEY);
			INSERT INTO fridays_new SELECT start_time FROM fridays;
			DROP TABLE fridays;
			ALTER TABLE fridays_new RENAME TO fridays`
	_, err := a.db.Exec(stmt)
	return err
}

func Patch002(a *SQLAccessor) error {
	// create the versions table
	stmt := `CREATE TABLE IF NOT EXISTS app_versions (name text NOT NULL PRIMARY KEY, version int NOT NULL)`
	_, err := a.db.Exec(stmt)
	if err != nil {
		return err
	}
	stmt = `INSERT INTO app_versions (name, version) VALUES ('schema', 2)`
	_, err = a.db.Exec(stmt)
	return err
}

func Patch003(a *SQLAccessor) error {
	stmt := `ALTER TABLE fridays ADD COLUMN invited_group text;
			ALTER TABLE fridays ADD COLUMN details text;`
	_, err := a.db.Exec(stmt)
	return err
}

func Patch004(a *SQLAccessor) error {
	stmt := `ALTER TABLE friends ADD COLUMN preferences text default "{}";`
	_, err := a.db.Exec(stmt)
	return err
}

func Patch005(a *SQLAccessor) error {
	stmt := `ALTER TABLE fridays ADD COLUMN invited text default "[]";
			ALTER TABLE fridays ADD COLUMN max_guests int default 10;
			ALTER TABLE fridays ADD COLUMN enabled bool default true;`
	_, err := a.db.Exec(stmt)
	return err
}

func Patch006(a *SQLAccessor) error {
	stmt := `CREATE TABLE IF NOT EXISTS friends_new 
				(id integer PRIMARY KEY AUTOINCREMENT,
				 email text NOT NULL UNIQUE,
				 name text,
				 preferences text default "{}");
			INSERT INTO friends_new (email, name, preferences) SELECT * FROM friends;
			DROP TABLE friends;
			ALTER TABLE friends_new RENAME TO friends;`
	_, err := a.db.Exec(stmt)
	return err
}
