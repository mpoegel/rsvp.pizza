package pizza

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Calendar struct {
	srv        *calendar.Service
	id         string
	eventCache map[string]*calendar.Event
}

var cal *Calendar

func InitCalendarClient(credentialFile, tokenFile, id string, ctx context.Context) error {
	b, err := os.ReadFile(credentialFile)
	if err != nil {
		return err
	}
	config, err := google.ConfigFromJSON(b, calendar.CalendarEventsScope)
	if err != nil {
		return err
	}

	f, err := os.Open(tokenFile)
	if err != nil {
		return err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	if err = json.NewDecoder(f).Decode(tok); err != nil {
		return err
	}
	client := config.Client(context.Background(), tok)
	if srv, err := calendar.NewService(ctx, option.WithHTTPClient(client)); err != nil {
		return err
	} else {
		cal = &Calendar{srv, id, make(map[string]*calendar.Event)}
		return nil
	}
}

func CreateCalendarEvent(eventID string, start, end time.Time) (*calendar.Event, error) {
	description := "Welcome to Pizza Friday!"
	timezone := "America/New_York"
	guestsCanInviteOthers := false
	event := calendar.Event{
		AnyoneCanAddSelf: false,
		Description:      description,
		End: &calendar.EventDateTime{
			DateTime: end.Format(time.RFC3339),
			TimeZone: timezone,
		},
		GuestsCanInviteOthers: &guestsCanInviteOthers,
		GuestsCanModify:       false,
		Id:                    eventID,
		Locked:                true,
		Reminders:             nil,
		Start: &calendar.EventDateTime{
			DateTime: start.Format(time.RFC3339),
			TimeZone: timezone,
		},
		Status:     "confirmed",
		Summary:    "Pizza Friday",
		Visibility: "private",
	}
	// TODO add timeout
	return cal.srv.Events.Insert(cal.id, &event).Context(context.Background()).Do()
}

func GetCalendarEvent(eventID string) (*calendar.Event, error) {
	// TODO add timeout
	if event, ok := cal.eventCache[eventID]; ok {
		return event, nil
	}
	if event, err := cal.srv.Events.Get(cal.id, eventID).Do(); err == nil {
		cal.eventCache[eventID] = event
		return event, nil
	} else if err != nil && err.Error() == "googleapi: Error 404: Not Found, notFound" {
		cal.eventCache[eventID] = nil
		return nil, nil
	} else {
		return nil, err
	}
}

func InviteToCalendarEvent(eventID string, start, end time.Time, name, email string) (*calendar.Event, error) {
	// TODO add locks
	event, err := GetCalendarEvent(eventID)
	if err != nil {
		Log.Info("event does not exist, creating new", zap.String("eventID", eventID))
		event, err = CreateCalendarEvent(eventID, start, end)
		if err != nil {
			Log.Error("failed to create event", zap.String("eventID", eventID), zap.Error(err))
			return nil, err
		}
		Log.Info("event created", zap.String("eventID", event.Id))
	}
	event.Attendees = append(event.Attendees, &calendar.EventAttendee{
		DisplayName: name,
		Email:       email,
	})
	// TODO add timeout
	event, err = cal.srv.Events.Update(cal.id, eventID, event).Do()
	if err != nil {
		cal.eventCache[eventID] = event
	}
	return event, err
}

func ListEvents(numEvents int64) (*calendar.Events, error) {
	t := time.Now().Format(time.RFC3339)
	// TODO add timeout
	events, err := cal.srv.Events.List(cal.id).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(t).
		MaxResults(numEvents).
		OrderBy("startTime").
		Do()
	return events, err
}
