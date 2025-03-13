package api

import (
	"fmt"
	"io"
	"reflect"
	"time"

	jsonapi "github.com/hashicorp/jsonapi"
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
	ID   string `jsonapi:"primary,guest"`
	Name string `jsonapi:"attr,name,omitempty"`
}

func (g *Guest) JSONAPILinks() *jsonapi.Links {
	return &jsonapi.Links{
		"self":    fmt.Sprintf("/api/guest/%s", g.ID),
		"profile": fmt.Sprintf("/api/guest/%s/profile", g.ID),
	}
}

type GuestProfile struct {
	ID       string   `jsonapi:"primary,guest"`
	Email    string   `jsonapi:"attr,email"`
	Toppings []string `jsonapi:"attr,toppings"`
	Cheese   []string `jsonapi:"attr,cheese"`
	Sauce    []string `jsonapi:"attr,sauce"`
	Doneness string   `jsonapi:"attr,doneness"`
}

func (p *GuestProfile) JSONAPILinks() *jsonapi.Links {
	return &jsonapi.Links{
		"self": fmt.Sprintf("/api/guest/%s/profile", p.ID),
	}
}

func UnmarshalFriday(r io.Reader) (*Friday, error) {
	friday := &Friday{}
	if err := jsonapi.UnmarshalPayload(r, friday); err != nil {
		return nil, err
	}
	return friday, nil
}

func UnmarshalFridays(r io.Reader) ([]*Friday, error) {
	payload, err := jsonapi.UnmarshalManyPayload(r, reflect.TypeOf(new(Friday)))
	if err != nil {
		return nil, err
	}
	fridays := make([]*Friday, len(payload))
	for i, f := range payload {
		fridays[i] = f.(*Friday)
	}
	return fridays, nil
}
