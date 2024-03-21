package pizza

import (
	"flag"

	"go.uber.org/zap"
)

func Patch(args []string) {
	fs := flag.NewFlagSet("patch", flag.ExitOnError)
	isInit := fs.Bool("init", false, "initialize all tables")
	_ = fs.Bool("drop", false, "drop all tables")
	n := fs.Int("n", 0, "patch number")
	fs.Parse(args)

	config := LoadConfigEnv()
	var accessor *SQLAccessor
	var err error
	if config.UseSQLite {
		accessor, err = NewSQLAccessor(config.DBFile)
		if err != nil {
			Log.Fatal("sql accessor init failure", zap.Error(err))
		}
	} else {
		Log.Fatal("must use sql accessor")
	}

	if *isInit {
		accessor.CreateTables()
	}

	switch *n {
	case 1:
		err = Patch001(accessor)
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
