package app

import (
	"net/http"
	"strconv"
	"strings"

	"menuapp/internal/auth"
)

type restaurantsPageData struct {
	navData
	Error string
}

func (a *App) handleRestaurantsList(w http.ResponseWriter, r *http.Request) {
	u, ok := a.requireUser(w, r)
	if !ok {
		return
	}
	nav := a.nav(r, u, nil, flashFromQuery(r))
	a.render(w, "restaurants.html", restaurantsPageData{navData: nav})
}

func (a *App) handleRestaurantCreate(w http.ResponseWriter, r *http.Request) {
	u, ok := a.requireUser(w, r)
	if !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	targetStr := strings.TrimSpace(r.FormValue("target_food_cost_pct"))
	target := 30.0
	if targetStr != "" {
		if v, err := strconv.ParseFloat(targetStr, 64); err == nil {
			target = v
		}
	}
	if name == "" {
		nav := a.nav(r, u, nil, "")
		a.render(w, "restaurants.html", restaurantsPageData{navData: nav, Error: "Restaurant name is required."})
		return
	}
	rest, err := a.Store.CreateRestaurant(u.ID, name, target)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	// seed a starter menu so the owner lands somewhere useful
	menuID, _ := a.Store.CreateMenu(rest.ID, "Main Menu")
	_, _ = a.Store.CreateSection(menuID, "Starters")

	http.SetCookie(w, auth.NewActiveRestaurantCookie(rest.ID))
	http.Redirect(w, r, "/r/"+itoa(rest.ID)+"/dashboard", http.StatusSeeOther)
}

func (a *App) handleRestaurantSelect(w http.ResponseWriter, r *http.Request) {
	u, ok := a.requireUser(w, r)
	if !ok {
		return
	}
	rid, ok := pathID(r, "rid")
	if !ok {
		http.NotFound(w, r)
		return
	}
	rest, err := a.Store.RestaurantByID(rid)
	if err != nil || rest.OwnerUserID != u.ID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	http.SetCookie(w, auth.NewActiveRestaurantCookie(rest.ID))
	http.Redirect(w, r, "/r/"+itoa(rest.ID)+"/dashboard", http.StatusSeeOther)
}
