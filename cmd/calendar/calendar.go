package main

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/mpoegel/rsvp.pizza/pkg/pizza"
	"github.com/stretchr/testify/require"
)

func TestCalendarInvite(t *testing.T) {
	gCal, err := pizza.NewGoogleCalendar("../../credentials.json", "../../token.json", os.Getenv("CALENDAR_ID"), context.Background())
	require.Nil(t, err)

	eventID := "349446585587859531"
	require.Nil(t, err)
	start := time.Now()
	end := time.Now().Add(1 * time.Hour)

	newEvent := pizza.CalendarEvent{
		AnyoneCanAddSelf:      false,
		Description:           "Test pizza event",
		EndTime:               start,
		GuestsCanInviteOthers: false,
		GuestsCanModify:       false,
		Id:                    "",
		Locked:                true,
		StartTime:             end,
		Status:                "confirmed",
		Summary:               "Test Pizza Friday",
		Visibility:            "private",
	}

	err = gCal.CreateEvent(newEvent)
	require.Nil(t, err)

	err = gCal.InviteToEvent(eventID, os.Getenv("TEST_EMAIL"), "Test User")
	require.Nil(t, err)

	event, err := gCal.GetEvent(eventID)
	require.Nil(t, err)

	slog.Debug("got event", "event", event)
}

func main() {

}
