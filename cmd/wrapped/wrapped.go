package main

import (
	"context"
	"encoding/json"
	"flag"
	"log/slog"
	"os"
	"time"

	pizza "github.com/mpoegel/rsvp.pizza/pkg/pizza"
)

type WrappedData struct {
	Friends      map[string]int `json:"friends"`
	TotalFridays int            `json:"totalFridays"`
}

func main() {
	credentialsFile := flag.String("cred", "credentials.json", "google calendar credential file")
	tokenFile := flag.String("token", "token.json", "google calendar token file")
	out := flag.String("out", "wrapped.json", "output file")
	flag.Parse()

	calendarID := os.Getenv("CALENDAR_ID")

	googleCal, err := pizza.NewGoogleCalendar(*credentialsFile, *tokenFile, calendarID, context.Background())
	if err != nil {
		slog.Error("could not create google calendar client", "err", err)
		os.Exit(1)
	}

	cal := pizza.NewCalendar(googleCal)
	start := time.Time{}
	start = start.AddDate(2023, 1, 1)
	end := time.Time{}
	end = end.AddDate(2023, 12, 31)
	events, err := cal.ListEventsBetween(start, end, 100)
	if err != nil {
		slog.Error("could not get events", "err", err)
		os.Exit(1)
	}

	data := WrappedData{
		Friends:      map[string]int{},
		TotalFridays: 0,
	}
	for _, event := range events {
		slog.Info("processing event", "summary", event.Summary, "start", event.StartTime)
		for _, attendee := range event.Attendees {
			if _, ok := data.Friends[attendee]; !ok {
				data.Friends[attendee] = 1
			} else {
				data.Friends[attendee]++
			}
		}
		data.TotalFridays++
	}

	fp, err := os.OpenFile(*out, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		slog.Error("could not open output file", "err", err)
		os.Exit(1)
	}
	encoder := json.NewEncoder(fp)
	if err = encoder.Encode(data); err != nil {
		slog.Error("could not encode results", "err", err)
		os.Exit(1)
	}

	slog.Info("done")
}
