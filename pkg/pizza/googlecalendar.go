package pizza

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

const (
	DefaultGoogleCalendarTimeout  = 5 * time.Second
	DefaultGoogleCalendarTimezone = "America/New_York"
)

var (
	ErrEventNotFound = errors.New("event not found")
)

type GoogleCalendar struct {
	srv      *calendar.Service
	id       string
	Timezone string
	Timeout  time.Duration
}

func NewGoogleCalendar(credentialFile, tokenFile, id string, ctx context.Context) (*GoogleCalendar, error) {
	b, err := os.ReadFile(credentialFile)
	if err != nil {
		return nil, err
	}
	config, err := google.ConfigFromJSON(b, calendar.CalendarEventsScope)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(tokenFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	if err = json.NewDecoder(f).Decode(tok); err != nil {
		return nil, err
	}
	client := config.Client(context.Background(), tok)
	if srv, err := calendar.NewService(ctx, option.WithHTTPClient(client)); err != nil {
		return nil, err
	} else {
		cal := &GoogleCalendar{srv, id, DefaultGoogleCalendarTimezone, DefaultGoogleCalendarTimeout}
		return cal, nil
	}
}

func (c *GoogleCalendar) CreateEvent(newEvent CalendarEvent) error {
	event := calendar.Event{
		AnyoneCanAddSelf: newEvent.AnyoneCanAddSelf,
		Description:      newEvent.Description,
		End: &calendar.EventDateTime{
			DateTime: newEvent.EndTime.Format(time.RFC3339),
			TimeZone: c.Timezone,
		},
		GuestsCanInviteOthers: &newEvent.GuestsCanInviteOthers,
		GuestsCanModify:       newEvent.GuestsCanModify,
		Id:                    newEvent.Id,
		Locked:                newEvent.Locked,
		Reminders:             nil,
		Start: &calendar.EventDateTime{
			DateTime: newEvent.StartTime.Format(time.RFC3339),
			TimeZone: c.Timezone,
		},
		Status:     newEvent.Status,
		Summary:    newEvent.Summary,
		Visibility: newEvent.Visibility,
	}
	// TODO add timeout
	_, err := c.srv.Events.Insert(c.id, &event).Context(context.Background()).Do()
	return err
}

func (c *GoogleCalendar) getCalendarEvent(eventID string) (*calendar.Event, error) {
	// TODO add timeout
	if event, err := c.srv.Events.Get(c.id, eventID).Do(); err == nil {
		return event, nil
	} else if err.Error() == "googleapi: Error 404: Not Found, notFound" {
		return nil, ErrEventNotFound
	} else {
		return nil, err
	}
}

func (c *GoogleCalendar) googleEventToEvent(event *calendar.Event) CalendarEvent {
	var startTime time.Time
	var endTime time.Time
	var err error
	if len(event.Start.DateTime) > 0 {
		startTime, err = time.Parse(time.RFC3339, event.Start.DateTime)
		if err != nil {
			slog.Error("could not parse event start time", "eventID", event.Id, "time", event.Start.DateTime)
		}
	}
	if len(event.End.DateTime) > 0 {
		endTime, err = time.Parse(time.RFC3339, event.End.DateTime)
		if err != nil {
			slog.Error("could not parse event end time", "eventID", event.Id, "time", event.End.DateTime)
		}
	}
	attendees := make([]string, len(event.Attendees))
	for i, attendee := range event.Attendees {
		attendees[i] = attendee.Email
	}
	evt := CalendarEvent{
		AnyoneCanAddSelf: event.AnyoneCanAddSelf,
		Attendees:        attendees,
		Description:      event.Description,
		EndTime:          endTime,
		GuestsCanModify:  event.GuestsCanModify,
		Id:               event.Id,
		Locked:           event.Locked,
		StartTime:        startTime,
		Status:           event.Status,
		Summary:          event.Summary,
		Visibility:       event.Visibility,
	}
	if event.GuestsCanInviteOthers != nil {
		evt.GuestsCanInviteOthers = *event.GuestsCanInviteOthers
	}
	return evt
}

func (c *GoogleCalendar) GetEvent(eventID string) (CalendarEvent, error) {
	// TODO add timeout
	if event, err := c.getCalendarEvent(eventID); err != nil {
		return CalendarEvent{}, err
	} else {
		return c.googleEventToEvent(event), nil
	}
}

func (c *GoogleCalendar) InviteToEvent(eventID, email, name string) error {
	// TODO add locks
	event, err := c.getCalendarEvent(eventID)
	if err != nil {
		return err
	}

	for _, attendee := range event.Attendees {
		if attendee.Email == email && attendee.ResponseStatus != "declined" {
			slog.Info("already invited", "email", email, "eventID", eventID)
			return nil
		}
	}

	event.Attendees = append(event.Attendees, &calendar.EventAttendee{
		Email:          email,
		DisplayName:    name,
		ResponseStatus: "needsAction",
	})

	// TODO add timeout
	_, err = c.srv.Events.Update(c.id, eventID, event).Do()
	return err
}

func (c *GoogleCalendar) DeclineEvent(eventID, email string) error {
	// TODO add locks
	event, err := c.getCalendarEvent(eventID)
	if err != nil {
		return err
	}

	found := false
	for _, attendee := range event.Attendees {
		if attendee.Email == email {
			attendee.ResponseStatus = "declined"
			found = true
			break
		}
	}
	if !found {
		return ErrNotInvited
	}

	// TODO add timeout
	_, err = c.srv.Events.Update(c.id, eventID, event).Do()
	return err
}

func (c *GoogleCalendar) ListEvents(numEvents int) ([]CalendarEvent, error) {
	t := time.Now().Format(time.RFC3339)
	// TODO add timeout
	events, err := c.srv.Events.List(c.id).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(t).
		MaxResults(int64(numEvents)).
		OrderBy("startTime").
		Do()
	if err != nil {
		return nil, err
	}
	result := make([]CalendarEvent, len(events.Items))
	for i := range result {
		result[i] = c.googleEventToEvent(events.Items[i])
	}
	return result, err
}

func (c *GoogleCalendar) ListEventsBetween(start, end time.Time, numEvents int) ([]CalendarEvent, error) {
	// TODO add timeout
	// TODO use start time
	events, err := c.srv.Events.List(c.id).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMax(end.Format(time.RFC3339)).
		MaxResults(int64(numEvents)).
		OrderBy("startTime").
		Do()
	if err != nil {
		return nil, err
	}
	result := make([]CalendarEvent, len(events.Items))
	for i := range result {
		result[i] = c.googleEventToEvent(events.Items[i])
	}
	return result, err
}

func (c *GoogleCalendar) CancelEvent(eventID string) error {
	return c.srv.Events.Delete(c.id, eventID).Do()
}

func (c *GoogleCalendar) ActivateEvent(eventID string) error {
	event, err := c.getCalendarEvent(eventID)
	if err != nil {
		return err
	}
	event.Status = "confirmed"
	_, err = c.srv.Events.Update(c.id, eventID, event).Do()
	return err
}
