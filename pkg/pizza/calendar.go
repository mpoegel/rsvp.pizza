package pizza

import (
	"errors"
	"time"
)

type Calendar interface {
	CreateEvent(CalendarEvent) error
	GetEvent(eventID string) (CalendarEvent, error)
	InviteToEvent(eventID, email, name string) error
	DeclineEvent(eventID, email string) error
	ListEvents(numEvents int) ([]CalendarEvent, error)
	ListEventsBetween(start, end time.Time, numEvents int) ([]CalendarEvent, error)
	CancelEvent(eventID string) error
	ActivateEvent(eventID string) error
}

type CalendarEvent struct {
	AnyoneCanAddSelf      bool
	Attendees             []string
	Description           string
	EndTime               time.Time
	GuestsCanInviteOthers bool
	GuestsCanModify       bool
	Id                    string
	Locked                bool
	StartTime             time.Time
	Status                string
	Summary               string
	Visibility            string
}

var (
	ErrNotInvited = errors.New("not invited")
)
