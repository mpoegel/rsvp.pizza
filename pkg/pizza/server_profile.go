package pizza

import (
	"log/slog"
	"net/http"
	"path"
	"text/template"

	types "github.com/mpoegel/rsvp.pizza/pkg/types"
)

type Preference struct {
	Name       string
	IsSelected bool
}

type ProfilePageData struct {
	LoggedIn   bool
	Name       string
	Toppings   []Preference
	Cheese     []Preference
	Sauce      []Preference
	Doneness   []Preference
	PixelPizza PixelPizzaPageData
}

func (s *Server) HandleGetProfile(w http.ResponseWriter, r *http.Request) {
	plate, err := template.ParseFiles(path.Join(s.config.StaticDir, "html/profile.html"))
	if err != nil {
		slog.Error("template index failure", "error", err)
		s.Handle500(w, r)
		return
	}
	if _, err = plate.ParseGlob(path.Join(s.config.StaticDir, "html/snippets/*.html")); err != nil {
		slog.Error("template snippets parse failure", "error", err)
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
			slog.Error("failed to get preferences", "error", err, "email", claims.Email)
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

		data.PixelPizza.Pizza = NewPixelPizzaFromPreferences(prefs).Render("darkblue")
		data.PixelPizza.Size = "33px"
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
		{Name: types.Sausage.String(), IsSelected: toppings[types.Sausage]},
		{Name: types.Ham.String(), IsSelected: toppings[types.Ham]},
		{Name: types.Pineapple.String(), IsSelected: toppings[types.Pineapple]},
		{Name: types.Green_Pepper.String(), IsSelected: toppings[types.Green_Pepper]},
		{Name: types.Mushroom.String(), IsSelected: toppings[types.Mushroom]},
	}
	data.Cheese = []Preference{
		{Name: types.Shredded_Mozzarella.String(), IsSelected: cheeses[types.Shredded_Mozzarella]},
		{Name: types.Whole_Mozzarella.String(), IsSelected: cheeses[types.Whole_Mozzarella]},
		{Name: types.Cheddar.String(), IsSelected: cheeses[types.Cheddar]},
		{Name: types.Ricotta.String(), IsSelected: cheeses[types.Ricotta]},
		{Name: types.Parmesan.String(), IsSelected: cheeses[types.Parmesan]},
	}
	data.Sauce = []Preference{
		{Name: types.Raw_Tomatoes.String(), IsSelected: sauces[types.Raw_Tomatoes]},
		{Name: types.Cooked_Tomatoes.String(), IsSelected: sauces[types.Cooked_Tomatoes]},
		{Name: types.Basil_Pesto.String(), IsSelected: sauces[types.Basil_Pesto]},
		{Name: types.Vodka.String(), IsSelected: sauces[types.Vodka]},
		{Name: types.Alfredo.String(), IsSelected: sauces[types.Alfredo]},
	}
	data.Doneness = []Preference{
		{Name: types.Well_Done.String(), IsSelected: doneness == types.Well_Done},
		{Name: types.Medium_Well.String(), IsSelected: doneness == types.Medium_Well},
		{Name: types.Medium.String(), IsSelected: doneness == types.Medium},
		{Name: types.Medium_Rare.String(), IsSelected: doneness == types.Medium_Rare},
		{Name: types.Rare.String(), IsSelected: doneness == types.Rare},
	}

	if err = plate.ExecuteTemplate(w, "Profile", data); err != nil {
		slog.Error("template execution failure", "error", err)
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
		slog.Error("form parse failure on profile edit", "error", err)
		w.Write(getToast("bad request"))
		return
	}

	prefs := Preferences{
		Toppings: types.ParseToppings(r.Form["toppings"]),
		Cheese:   types.ParseCheeses(r.Form["cheese"]),
		Sauce:    types.ParseSauces(r.Form["sauce"]),
		Doneness: types.ParseDoneness(r.Form["doneness"][0]),
	}

	slog.Info("got profile update", "preferences", prefs)

	if err := s.store.SetPreferences(claims.Email, prefs); err != nil {
		slog.Error("failed to set preferences", "error", err, "email", claims.Email)
		w.Write(getToast("failed to set preferences"))
		return
	}

	w.Write(getToast("preferences updated"))
}
