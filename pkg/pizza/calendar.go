package pizza

import "time"

type Calendar interface {
	CreateEvent(CalendarEvent) error
	GetEvent(eventID string) (CalendarEvent, error)
	InviteToEvent(eventID, email, name string) error
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
