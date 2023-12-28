package main

import (
	"fmt"
	"os"

	"github.com/mpoegel/rsvp.pizza/pkg/pizza"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("command required: [run, edit, patch]")
		os.Exit(1)
	}
	switch args[1] {
	case "run":
		pizza.Run(os.Args[2:])
	case "edit":
		pizza.Edit(os.Args[2:])
	case "patch":
		pizza.Patch(os.Args[2:])
	default:
		fmt.Println("command must be one of [run, edit, patch]")
		os.Exit(1)
	}
}
