package pizza

import (
	"time"

	f "github.com/fauna/faunadb-go/v4/faunadb"
	"go.uber.org/zap"
)

var faunaClient *f.FaunaClient

func newFaunaClient(secret string) {
	faunaClient = f.NewFaunaClient(secret)
}

func IsFriendAllowed(friendEmail string) (bool, error) {
	qRes, err := faunaClient.Query(
		f.Exists(f.MatchTerm(f.Index("all_emails"), friendEmail)),
	)
	if err != nil {
		Log.Error("fauna error", zap.Error(err))
		return false, err
	}
	var exists bool
	if err := qRes.Get(&exists); err != nil {
		Log.Error("fauna parse error", zap.Error(err))
		return false, err
	}
	return exists, nil
}

func GetFriendName(friendEmail string) (string, error) {
	/*
		Get(Select(
			"ref",
			Get(Match(Index("all_emails"), "test@email.com"))
		))
	*/
	var name string
	qRes, err := faunaClient.Query(f.Get(f.MatchTerm(f.Index("all_emails"), friendEmail)))
	if err != nil {
		Log.Error("fauna error", zap.Error(err))
		return name, err
	}
	if err = qRes.At(f.ObjKey("data", "name")).Get(&name); err != nil {
		Log.Error("fauna decode error", zap.Error(err))
		return name, err
	}
	return name, nil
}

func GetAllFridays() ([]time.Time, error) {
	qRes, err := faunaClient.Query(f.Paginate(f.Match(f.Index("all_fridays"))))
	if err != nil {
		Log.Error("fauna error", zap.Error(err))
		return nil, err
	}
	var arr []time.Time
	if err = qRes.At(f.ObjKey("data")).Get(&arr); err != nil {
		Log.Error("fauna decode error", zap.Error(err))
		return nil, err
	}
	Log.Debug("got all fridays", zap.Times("fridays", arr))
	return arr, nil
}

func GetUpcomingFridays(daysAhead int) ([]time.Time, error) {
	/*
		Map(
			Paginate(
				Range(
					Match(Index("all_fridays_range")),
					Now(),
					TimeAdd(Now(), 30, "days")
				)
			),
			Lambda('x', Select(0, Var('x')))
		)
	*/
	qRes, err := faunaClient.Query(f.Map(f.Paginate(f.Range(
		f.Match(f.Index("all_fridays_range")),
		f.Now(),
		f.TimeAdd(f.Now(), daysAhead, "days"),
	)), f.Lambda("x", f.Select(0, f.Var("x")))))
	if err != nil {
		Log.Error("fauna error", zap.Error(err))
		return nil, err
	}
	var times []time.Time
	if err = qRes.At(f.ObjKey("data")).Get(&times); err != nil {
		Log.Error("fauna decode error", zap.Error(err))
		return nil, err
	}

	Log.Debug("got all fridays", zap.Times("fridays", times))

	return times, nil
}

func CreateRSVP(friendEmail, code string, pendingDates []time.Time) error {
	qRes, err := faunaClient.Query(
		f.Update(
			f.Select(
				"ref",
				f.Get(f.MatchTerm(f.Index("all_emails"), friendEmail)),
			),
			f.Obj{"data": f.Obj{
				"pending_rsvps": pendingDates,
				"rsvp_code":     code,
			}},
		),
	)
	if err != nil {
		Log.Error("fauna error", zap.Error(err))
		return err
	}
	Log.Debug("rsvp created: %+v", zap.Any("result", qRes))
	return nil
}

func ConfirmRSVP(friendEmail, code string) error {
	qRes, err := faunaClient.Query(
		f.Let().Bind(
			"pending", f.Select([]string{"data", "pending_rsvps"},
				f.Get(f.MatchTerm(f.Index("rsvp_codes"), []string{friendEmail, code}))),
		).Bind(
			"ref", f.Select("ref",
				f.Get(f.MatchTerm(f.Index("rsvp_codes"), []string{friendEmail, code}))),
		).In(
			f.Update(f.Var("ref"), f.Obj{
				"data": f.Obj{
					"confirmed_rsvps": f.Var("pending"),
				},
			}),
		),
	)
	if err != nil {
		Log.Error("fauna error", zap.Error(err))
		return err
	}
	Log.Debug("rsvp confirmed", zap.Any("result", qRes))
	return nil
}
