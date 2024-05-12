package api

import (
	"fmt"
	"time"

	jsonapi "github.com/google/jsonapi"
)

type Friday struct {
	ID        string    `jsonapi:"primary,friday"`
	StartTime time.Time `jsonapi:"attr,start_time"`
	Details   string    `jsonapi:"attr,details"`
	Guests    []*Guest  `jsonapi:"relation,guests"`
}

func (f *Friday) JSONAPILinks() *jsonapi.Links {
	return &jsonapi.Links{
		"self": fmt.Sprintf("/api/friday/%s", f.ID),
	}
}

type Guest struct {
	ID    string `jsonapi:"primary,guest"`
	Email string `jsonapi:"attr,email"`
	Name  string `jsonapi:"attr,name,omitempty"`
}
