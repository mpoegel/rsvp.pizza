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
	year := flag.Int("year", time.Now().Year(), "year for report")
	flag.Parse()

	calendarID := os.Getenv("CALENDAR_ID")

	googleCal, err := pizza.NewGoogleCalendar(*credentialsFile, *tokenFile, calendarID, context.Background())
	if err != nil {
		slog.Error("could not create google calendar client", "err", err)
		os.Exit(1)
	}

	start := time.Time{}
	start = start.AddDate(*year, 1, 1)
	end := time.Time{}
	end = end.AddDate(*year, 12, 31)
	events, err := googleCal.ListEventsBetween(start, end, 100)
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
			if _, ok := data.Friends[attendee.Email]; !ok {
				data.Friends[attendee.Email] = 1
			} else {
				data.Friends[attendee.Email]++
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
