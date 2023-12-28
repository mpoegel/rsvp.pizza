package pizza

import (
	"flag"

	"go.uber.org/zap"
)

func Patch(args []string) {
	fs := flag.NewFlagSet("patch", flag.ExitOnError)
	isInit := fs.Bool("init", false, "initialize all tables")
	_ = fs.Bool("drop", false, "drop all tables")
	_ = fs.Int("n", 0, "patch number")
	fs.Parse(args)

	config := LoadConfigEnv()
	var accessor Accessor
	var err error
	if config.UseSQLite {
		accessor, err = NewSQLAccessor(config.DBFile)
		if err != nil {
			Log.Fatal("sql accessor init failure", zap.Error(err))
		}
	}

	if *isInit {
		accessor.CreateTables()
	}
}
