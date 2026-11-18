package app

import (
	"net/http"
	"strconv"
	"strings"
)

type settingsPageData struct {
	navData
	Error string
}

func (a *App) handleSettings(w http.ResponseWriter, r *http.Request) {
	u, rest, ok := a.requireRestaurant(w, r)
	if !ok {
		return
	}
	nav := a.nav(r, u, rest, flashFromQuery(r))
	a.render(w, "settings.html", settingsPageData{navData: nav})
}

func (a *App) handleSettingsSave(w http.ResponseWriter, r *http.Request) {
	u, rest, ok := a.requireRestaurant(w, r)
	if !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	target, err := strconv.ParseFloat(r.FormValue("target_food_cost_pct"), 64)
	if name == "" || err != nil {
		nav := a.nav(r, u, rest, "")
		a.render(w, "settings.html", settingsPageData{navData: nav, Error: "Enter a valid name and target food cost %."})
		return
	}
	if err := a.Store.UpdateRestaurant(rest.ID, name, target); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, redirectFlash("/r/"+itoa(rest.ID)+"/settings", "Settings saved."), http.StatusSeeOther)
}
