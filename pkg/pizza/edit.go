package pizza

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

func Edit(args []string) {
	fs := flag.NewFlagSet("edit", flag.ExitOnError)
	isInteractive := fs.Bool("i", false, "interactive mode")
	add := fs.String("a", "", `add a new friend or friday, formatted as
		friend://<email>/<first name>/<last name> - to add a new friend
		friday://YYYY/MM/DD - to add a new friday`)
	remove := fs.String("d", "", `remove a new friend or friday, formatted as
	friend://<email> - to remove a friend
	friday://YYYY/MM/DD - to remove a friday`)
	list := fs.String("list", "", "list [friends, fridays]")
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

	if len(*add) > 0 {
		addParts := strings.SplitN(*add, "://", 2)
		if len(addParts) != 2 {
			fmt.Println("invalid add argument")
			os.Exit(1)
		}
		switch addParts[0] {
		case "friend":
			addNewFriend(accessor, addParts[1])
		case "friday":
			addNewFriday(accessor, addParts[1])
		default:
			fmt.Println("invalid add target")
			os.Exit(1)
		}
	}

	if len(*remove) > 0 {
		removeParts := strings.SplitN(*remove, "://", 2)
		if len(removeParts) != 2 {
			fmt.Println("invalid remove argument")
			os.Exit(1)
		}
		switch removeParts[0] {
		case "friend":
			removeFriend(accessor, removeParts[1])
		case "friday":
			removeFriday(accessor, removeParts[1])
		default:
			fmt.Println("invalid remove target")
			os.Exit(1)
		}
	}

	if len(*list) > 0 {
		switch *list {
		case "friends":
			listFriends(accessor)
		case "fridays":
			listFridays(accessor)
		default:
			fmt.Println("invalid list target")
			os.Exit(1)
		}
	}

	if *isInteractive {
		interactiveEdit(accessor)
	}
}

func addNewFriend(accessor Accessor, newFriend string) {
	newFriendParts := strings.SplitN(newFriend, "/", 2)
	if len(newFriendParts) != 2 {
		fmt.Printf("invalid friend format: %s\n", newFriend)
		os.Exit(1)
	}
	email := newFriendParts[0]
	name := newFriendParts[1]
	name = strings.ReplaceAll(name, "/", " ")
	if err := accessor.AddFriend(email, name); err != nil {
		fmt.Printf("failed to add friend: %v\n", err)
	} else {
		fmt.Printf("added new friend: %s\n", email)
	}
}

func addNewFriday(accessor Accessor, newFriday string) {
	newFridayParts := strings.SplitN(newFriday, "/", 3)
	if len(newFridayParts) != 3 {
		fmt.Printf("invalid friday format: %s\n", newFriday)
		os.Exit(1)
	}
	year, err1 := strconv.Atoi(newFridayParts[0])
	month, err2 := strconv.Atoi(newFridayParts[1])
	day, err3 := strconv.Atoi(newFridayParts[2])
	if err1 != nil && err2 != nil && err3 != nil {
		fmt.Printf("invalid friday: %s\n", newFriday)
		os.Exit(1)
	}
	loc, _ := time.LoadLocation("America/New_York")
	f := time.Date(year, time.Month(month), day, 17, 30, 0, 0, loc)
	if err := accessor.AddFriday(f); err != nil {
		fmt.Printf("accessor failure: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("added friday: %s\n", formatEditDate(f))
}

func removeFriend(accessor Accessor, friend string) {
	if err := accessor.RemoveFriend(friend); err != nil {
		fmt.Printf("accessor failure: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("removed friend: %s\n", friend)
}

func removeFriday(accessor Accessor, fridayStr string) {
	newFridayParts := strings.SplitN(fridayStr, "/", 3)
	if len(newFridayParts) != 3 {
		fmt.Printf("invalid friday format: %s\n", fridayStr)
		os.Exit(1)
	}
	year, err1 := strconv.Atoi(newFridayParts[0])
	month, err2 := strconv.Atoi(newFridayParts[1])
	day, err3 := strconv.Atoi(newFridayParts[2])
	if err1 != nil && err2 != nil && err3 != nil {
		fmt.Printf("invalid friday: %s\n", fridayStr)
		os.Exit(1)
	}
	loc, _ := time.LoadLocation("America/New_York")
	f := time.Date(year, time.Month(month), day, 17, 30, 0, 0, loc)
	if err := accessor.RemoveFriday(f); err != nil {
		fmt.Printf("accessor failure: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("removed friday: %s\n", formatEditDate(f))
}

func listFriends(accessor Accessor) {
	friends, err := accessor.ListFriends()
	if err != nil {
		fmt.Printf("accessor failure: %v\n", err)
		os.Exit(1)
	}
	for _, f := range friends {
		fmt.Printf("%s - %s\n", f.Email, f.Name)
	}
}

func listFridays(accessor Accessor) {
	fridays, err := accessor.ListFridays()
	if err != nil {
		fmt.Printf("accessor failure: %v\n", err)
		os.Exit(1)
	}
	for _, f := range fridays {
		fmt.Printf("%s\n", f.Date.Format(time.RFC822))
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
			if err := accessor.RemoveFriday(nextFourFridays[index]); err != nil {
				fmt.Printf("accessor failure: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("removed %s\n", formatEditDate(nextFourFridays[index]))
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
