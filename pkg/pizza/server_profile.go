package pizza

import (
	"html/template"
	"net/http"
	"path"

	"github.com/mpoegel/rsvp.pizza/pkg/types"
	zap "go.uber.org/zap"
)

type Preference struct {
	Name       string
	IsSelected bool
}

type ProfilePageData struct {
	LoggedIn bool
	Name     string
	Toppings []Preference
	Cheese   []Preference
	Sauce    []Preference
	Doneness []Preference
}

func (s *Server) HandleGetProfile(w http.ResponseWriter, r *http.Request) {
	plate, err := template.ParseFiles(path.Join(s.config.StaticDir, "html/profile.html"))
	if err != nil {
		Log.Error("template index failure", zap.Error(err))
		s.Handle500(w, r)
		return
	}

	toppings := make(map[types.Topping]bool)
	cheeses := make(map[types.Cheese]bool)
	sauces := make(map[types.Sauce]bool)
	var doneness types.Doneness

	data := ProfilePageData{
		LoggedIn: false,
	}

	claims, ok := s.authenticateRequest(r)
	if ok {
		data.LoggedIn = true
		data.Name = claims.GivenName

		prefs, err := s.store.GetPreferences(claims.Email)
		if err != nil {
			Log.Error("failed to get preferences", zap.Error(err), zap.String("email", claims.Email))
		}
		for _, t := range prefs.Toppings {
			toppings[t] = true
		}
		for _, c := range prefs.Cheese {
			cheeses[c] = true
		}
		for _, s := range prefs.Sauce {
			sauces[s] = true
		}
		doneness = prefs.Doneness
	}

	data.Toppings = []Preference{
		{Name: types.Banana_Peppers.String(), IsSelected: toppings[types.Banana_Peppers]},
		{Name: types.Basil.String(), IsSelected: toppings[types.Basil]},
		{Name: types.Barbecue_Chicken.String(), IsSelected: toppings[types.Barbecue_Chicken]},
		{Name: types.Buffalo_Chicken.String(), IsSelected: toppings[types.Buffalo_Chicken]},
		{Name: types.Jalapeno.String(), IsSelected: toppings[types.Jalapeno]},
		{Name: types.Pepperoni.String(), IsSelected: toppings[types.Pepperoni]},
		{Name: types.Prosciutto.String(), IsSelected: toppings[types.Prosciutto]},
		{Name: types.Soppressata.String(), IsSelected: toppings[types.Soppressata]},
	}
	data.Cheese = []Preference{
		{Name: types.Shredded_Mozzarella.String(), IsSelected: cheeses[types.Shredded_Mozzarella]},
		{Name: types.Whole_Mozzarella.String(), IsSelected: cheeses[types.Whole_Mozzarella]},
		{Name: types.Cheddar.String(), IsSelected: cheeses[types.Cheddar]},
		{Name: types.Ricotta.String(), IsSelected: cheeses[types.Ricotta]},
	}
	data.Sauce = []Preference{
		{Name: types.Raw_Tomatoes.String(), IsSelected: sauces[types.Raw_Tomatoes]},
		{Name: types.Cooked_Tomatoes.String(), IsSelected: sauces[types.Cooked_Tomatoes]},
		{Name: types.Basil_Pesto.String(), IsSelected: sauces[types.Basil_Pesto]},
	}
	data.Doneness = []Preference{
		{Name: types.Well_Done.String(), IsSelected: doneness == types.Well_Done},
		{Name: types.Medium_Well.String(), IsSelected: doneness == types.Medium_Well},
		{Name: types.Medium.String(), IsSelected: doneness == types.Medium},
		{Name: types.Medium_Rare.String(), IsSelected: doneness == types.Medium_Rare},
		{Name: types.Rare.String(), IsSelected: doneness == types.Rare},
	}

	if err = plate.ExecuteTemplate(w, "Profile", data); err != nil {
		Log.Error("template execution failure", zap.Error(err))
		s.Handle500(w, r)
		return
	}
}

func (s *Server) HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticateRequest(r)
	if !ok {
		w.Write(getToast("not logged in"))
		return
	}

	if err := r.ParseForm(); err != nil {
		Log.Error("form parse failure on profile edit", zap.Error(err))
		w.Write(getToast("bad request"))
		return
	}

	prefs := Preferences{
		Toppings: types.ParseToppings(r.Form["toppings"]),
		Cheese:   types.ParseCheeses(r.Form["cheese"]),
		Sauce:    types.ParseSauces(r.Form["sauce"]),
		Doneness: types.ParseDoneness(r.Form["doneness"][0]),
	}

	Log.Info("got profile update", zap.Any("preferences", prefs))

	if err := s.store.SetPreferences(claims.Email, prefs); err != nil {
		Log.Error("failed to set preferences", zap.Error(err), zap.String("email", claims.Email))
		w.Write(getToast("failed to set preferences"))
		return
	}

	w.Write(getToast("preferences updated"))
}
