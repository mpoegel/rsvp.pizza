package pizza

import (
	"flag"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
)

func Edit(args []string) {
	nextFriday := time.Now()

	fs := flag.NewFlagSet("edit", flag.ExitOnError)
	isInteractive := fs.Bool("i", false, "interactive mode")
	_ = fs.String("addFriday", formatEditDate(nextFriday), "add friday")
	_ = fs.String("removeFriday", formatEditDate(nextFriday), "remove friday")
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

	if *isInteractive {
		interactiveEdit(accessor)
	}
}

func interactiveEdit(accessor Accessor) {
	nextFourFridays := make([]time.Time, 4)
	start := time.Now()
	loc, _ := time.LoadLocation("America/New_York")
	const helpStr = "a # - add friday; d # - remove friday; n - next; b - back; s - stop"
	action := "n"
	index := 0
	for action != "s" {
		switch action {
		case "b":
			start = start.AddDate(0, -1, -14)
			fallthrough
		case "n":
			nextFourFridays[0] = time.Date(start.Year(), start.Month(), start.Day(), 17, 30, 0, 0, loc)
			for nextFourFridays[0].Weekday() != time.Friday {
				nextFourFridays[0] = nextFourFridays[0].AddDate(0, 0, 1)
			}
			for i := 1; i < len(nextFourFridays); i++ {
				nextFourFridays[i] = nextFourFridays[i-1].AddDate(0, 0, 7)
			}
			scheduledFridays, err := accessor.GetUpcomingFridaysAfter(nextFourFridays[0].AddDate(0, 0, -2), 36)
			if err != nil {
				fmt.Printf("accessor failure: %v\n", err)
				os.Exit(1)
			}
			fi := 0
			for i, t := range nextFourFridays {
				symbol := " "
				if fi < len(scheduledFridays) && t.Equal(scheduledFridays[fi]) {
					symbol = "✔"
					fi++
				}
				fmt.Printf("%d) %s %s\n", i, symbol, formatEditDate(t))
			}
			start = nextFourFridays[3]
			fmt.Printf("\n%s\n", helpStr)
		case "a":
			err := accessor.AddFriday(nextFourFridays[index])
			if err != nil {
				fmt.Printf("accessor faliure: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("added %s\n", formatEditDate(nextFourFridays[index]))
		case "d":
			// TODO
			fmt.Printf("remove %d\n", index)
		default:
			fmt.Println(helpStr)
		}
		fmt.Print("> ")
		fmt.Scanf("%s %d", &action, &index)
	}
}

func formatEditDate(dt time.Time) string {
	return dt.Format(time.DateOnly)
}
