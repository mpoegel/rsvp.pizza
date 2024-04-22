package pizza

import (
	"flag"

	"go.uber.org/zap"
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
	accessor, err = NewSQLAccessor(config.DBFile)
	if err != nil {
		Log.Fatal("sql accessor init failure", zap.Error(err))
	}

	if *isInit {
		accessor.CreateTables()
	}

	switch *n {
	case 1:
		err = Patch001(accessor)
	case 2:
		err = Patch002(accessor)
	}

	if err != nil {
		Log.Error("patch failed", zap.Int("n", *n), zap.Error(err))
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
