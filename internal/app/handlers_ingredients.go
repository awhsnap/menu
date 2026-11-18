package app

import (
	"net/http"
	"strconv"
	"strings"

	"menuapp/internal/store"
)

type ingredientsPageData struct {
	navData
	Ingredients []store.Ingredient
	Error       string
}

func (a *App) handleIngredients(w http.ResponseWriter, r *http.Request) {
	u, rest, ok := a.requireRestaurant(w, r)
	if !ok {
		return
	}
	ingredients, err := a.Store.IngredientsByRestaurant(rest.ID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	nav := a.nav(r, u, rest, flashFromQuery(r))
	a.render(w, "ingredients.html", ingredientsPageData{navData: nav, Ingredients: ingredients})
}

func (a *App) handleIngredientCreate(w http.ResponseWriter, r *http.Request) {
	_, rest, ok := a.requireRestaurant(w, r)
	if !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	unit := strings.TrimSpace(r.FormValue("unit"))
	supplier := strings.TrimSpace(r.FormValue("supplier"))
	cost := parsePriceToCentsFloat(r.FormValue("unit_cost"))
	yieldPct, err := strconv.ParseFloat(r.FormValue("yield_pct"), 64)
	if err != nil || yieldPct <= 0 {
		yieldPct = 100
	}
	if name != "" && unit != "" {
		a.Store.CreateIngredient(rest.ID, name, unit, cost, supplier, yieldPct)
	}
	http.Redirect(w, r, "/r/"+itoa(rest.ID)+"/ingredients", http.StatusSeeOther)
}

func (a *App) handleIngredientUpdate(w http.ResponseWriter, r *http.Request) {
	_, rest, ok := a.requireRestaurant(w, r)
	if !ok {
		return
	}
	id, ok := pathID(r, "id")
	if !ok {
		http.NotFound(w, r)
		return
	}
	ing, err := a.Store.IngredientByID(id)
	if err != nil || ing.RestaurantID != rest.ID {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	unit := strings.TrimSpace(r.FormValue("unit"))
	supplier := strings.TrimSpace(r.FormValue("supplier"))
	yieldPct, err := strconv.ParseFloat(r.FormValue("yield_pct"), 64)
	if err != nil || yieldPct <= 0 {
		yieldPct = 100
	}
	newCost := parsePriceToCentsFloat(r.FormValue("unit_cost"))

	if name != "" && unit != "" {
		a.Store.UpdateIngredient(ing.ID, name, unit, supplier, yieldPct)
	}
	changed, _ := a.Store.UpdateIngredientCost(ing.ID, newCost)

	flash := "Ingredient updated."
	if changed {
		flash = "Ingredient updated. Price change recorded — check items using it for margin impact."
	}
	http.Redirect(w, r, redirectFlash("/r/"+itoa(rest.ID)+"/ingredients", flash), http.StatusSeeOther)
}

func (a *App) handleIngredientDelete(w http.ResponseWriter, r *http.Request) {
	_, rest, ok := a.requireRestaurant(w, r)
	if !ok {
		return
	}
	id, ok := pathID(r, "id")
	if !ok {
		http.NotFound(w, r)
		return
	}
	ing, err := a.Store.IngredientByID(id)
	if err != nil || ing.RestaurantID != rest.ID {
		http.NotFound(w, r)
		return
	}
	a.Store.DeleteIngredient(id)
	http.Redirect(w, r, "/r/"+itoa(rest.ID)+"/ingredients", http.StatusSeeOther)
}
