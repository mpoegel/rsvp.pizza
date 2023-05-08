package pizza

import "time"

type CalendarSource interface {
	CreateEvent(CalendarEvent) error
	GetEvent(eventID string) (CalendarEvent, error)
	InviteToEvent(eventID, email, name string) error
	ListEvents(numEvents int) ([]CalendarEvent, error)
	CancelEvent(eventID string) error
	ActivateEvent(eventID string) error
}

type Calendar struct {
	source CalendarSource
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

func NewCalendar(source CalendarSource) *Calendar {
	return &Calendar{
		source: source,
	}
}

func (c *Calendar) CreateEvent(newEvent CalendarEvent) error {
	return c.source.CreateEvent(newEvent)
}

func (c *Calendar) GetEvent(eventID string) (CalendarEvent, error) {
	return c.source.GetEvent(eventID)
}

func (c *Calendar) InviteToEvent(eventID, email, name string) error {
	return c.source.InviteToEvent(eventID, email, name)
}

func (c *Calendar) ListEvents(numEvents int) ([]CalendarEvent, error) {
	return c.source.ListEvents(numEvents)
}

func (c *Calendar) CancelEvent(eventID string) error {
	return c.source.CancelEvent(eventID)
}

func (c *Calendar) ActivateEvent(eventID string) error {
	return c.source.ActivateEvent(eventID)
}
