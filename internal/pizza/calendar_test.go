package pizza_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/mpoegel/rsvp.pizza/internal/pizza"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCalendarInvite(t *testing.T) {
	require.Nil(t, pizza.InitCalendarClient("../../credentials.json", "../../token.json", os.Getenv("CALENDAR_ID"), context.Background()))

	eventID := "349446585587859531"
	est, err := time.LoadLocation("America/New_York")
	require.Nil(t, err)
	start := time.Date(2023, 4, 5, 17, 30, 0, 0, est)
	end := time.Date(2023, 4, 5, 22, 00, 0, 0, est)

	_, err = pizza.InviteToCalendarEvent(eventID, start, end, "Test User", os.Getenv("TEST_EMAIL"))
	require.Nil(t, err)
	pizza.Log.Debug("invite sent", zap.String("eventID", eventID), zap.Time("start", start), zap.Time("end", end))
}
