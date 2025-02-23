package main

//go:generate mockery

import (
	"errors"
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
	var err error
	switch args[1] {
	case "run":
		pizza.Run(os.Args[2:])
	case "patch":
		pizza.Patch(os.Args[2:])
	default:
		err = errors.New("command must be one of [run, edit, patch]")
	}
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
