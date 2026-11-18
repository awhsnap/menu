package app

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"menuapp/internal/forecast"
	"menuapp/internal/store"
)

type itemDetailPageData struct {
	navData
	Item          store.Item
	RecipeLines   []store.RecipeLine
	Ingredients   []store.Ingredient
	CostCents     int64
	MarginCents   int64
	MarginPct     float64
	TargetPct     float64
	BelowTarget   bool
	Error         string
}

func (a *App) handleItemDetail(w http.ResponseWriter, r *http.Request) {
	u, rest, ok := a.requireRestaurant(w, r)
	if !ok {
		return
	}
	it, ok := a.loadOwnedItem(w, r, rest.ID)
	if !ok {
		return
	}
	lines, err := a.Store.RecipeLinesByItem(it.ID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	ingredients, err := a.Store.IngredientsByRestaurant(rest.ID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	cost, _ := a.Store.PlateCostCents(it.ID)
	margin := it.PriceCents - cost
	marginPct := 0.0
	if it.PriceCents > 0 {
		marginPct = float64(margin) / float64(it.PriceCents) * 100
	}

	nav := a.nav(r, u, rest, flashFromQuery(r))
	a.render(w, "item_detail.html", itemDetailPageData{
		navData:     nav,
		Item:        *it,
		RecipeLines: lines,
		Ingredients: ingredients,
		CostCents:   cost,
		MarginCents: margin,
		MarginPct:   marginPct,
		TargetPct:   rest.TargetFoodCostPct,
		BelowTarget: it.PriceCents > 0 && cost > 0 && (float64(cost)/float64(it.PriceCents)*100) > rest.TargetFoodCostPct,
	})
}

// loadOwnedItem loads the {id} path item and verifies it belongs to restaurantID.
func (a *App) loadOwnedItem(w http.ResponseWriter, r *http.Request, restaurantID int64) (*store.Item, bool) {
	id, ok := pathID(r, "id")
	if !ok {
		http.NotFound(w, r)
		return nil, false
	}
	it, err := a.Store.ItemByID(id)
	if err != nil || it.RestaurantID != restaurantID {
		http.NotFound(w, r)
		return nil, false
	}
	return it, true
}

func (a *App) handleItemUpdate(w http.ResponseWriter, r *http.Request) {
	_, rest, ok := a.requireRestaurant(w, r)
	if !ok {
		return
	}
	it, ok := a.loadOwnedItem(w, r, rest.ID)
	if !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	desc := strings.TrimSpace(r.FormValue("description"))
	price := parsePriceToCents(r.FormValue("price"))
	if name != "" {
		a.Store.UpdateItem(it.ID, name, desc, price)
	}
	http.Redirect(w, r, "/r/"+itoa(rest.ID)+"/items/"+itoa(it.ID), http.StatusSeeOther)
}

func (a *App) handleItemToggle(w http.ResponseWriter, r *http.Request) {
	_, rest, ok := a.requireRestaurant(w, r)
	if !ok {
		return
	}
	it, ok := a.loadOwnedItem(w, r, rest.ID)
	if !ok {
		return
	}
	a.Store.ToggleItemActive(it.ID)
	back := r.FormValue("back")
	if back == "" {
		back = "/r/" + itoa(rest.ID) + "/items/" + itoa(it.ID)
	}
	http.Redirect(w, r, back, http.StatusSeeOther)
}

func (a *App) handleItemDelete(w http.ResponseWriter, r *http.Request) {
	_, rest, ok := a.requireRestaurant(w, r)
	if !ok {
		return
	}
	it, ok := a.loadOwnedItem(w, r, rest.ID)
	if !ok {
		return
	}
	sectionID := it.SectionID
	sec, _ := a.Store.SectionByID(sectionID)
	a.Store.DeleteItem(it.ID)
	dest := "/r/" + itoa(rest.ID) + "/menus"
	if sec != nil {
		dest = "/r/" + itoa(rest.ID) + "/menus/" + itoa(sec.MenuID)
	}
	http.Redirect(w, r, dest, http.StatusSeeOther)
}

func (a *App) handleRecipeLineAdd(w http.ResponseWriter, r *http.Request) {
	_, rest, ok := a.requireRestaurant(w, r)
	if !ok {
		return
	}
	it, ok := a.loadOwnedItem(w, r, rest.ID)
	if !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	ingID, err1 := strconv.ParseInt(r.FormValue("ingredient_id"), 10, 64)
	qty, err2 := strconv.ParseFloat(r.FormValue("quantity"), 64)
	if err1 == nil && err2 == nil && qty > 0 {
		ing, err := a.Store.IngredientByID(ingID)
		if err == nil && ing.RestaurantID == rest.ID {
			a.Store.AddRecipeLine(it.ID, ingID, qty)
		}
	}
	http.Redirect(w, r, "/r/"+itoa(rest.ID)+"/items/"+itoa(it.ID), http.StatusSeeOther)
}

func (a *App) handleRecipeLineDelete(w http.ResponseWriter, r *http.Request) {
	_, rest, ok := a.requireRestaurant(w, r)
	if !ok {
		return
	}
	lid, ok := pathID(r, "lid")
	if !ok {
		http.NotFound(w, r)
		return
	}
	a.Store.DeleteRecipeLine(lid)
	back := r.FormValue("back")
	if back == "" {
		back = "/r/" + itoa(rest.ID) + "/dashboard"
	}
	http.Redirect(w, r, back, http.StatusSeeOther)
}

// --- Forecast ---

type itemForecastPageData struct {
	navData
	Item   store.Item
	Result forecast.Result
	ChartSVG template.HTML
}

func (a *App) handleItemForecast(w http.ResponseWriter, r *http.Request) {
	u, rest, ok := a.requireRestaurant(w, r)
	if !ok {
		return
	}
	it, ok := a.loadOwnedItem(w, r, rest.ID)
	if !ok {
		return
	}
	sales, err := a.Store.SalesByItem(it.ID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	res := forecast.Forecast(sales, 7)

	nav := a.nav(r, u, rest, "")
	a.render(w, "item_forecast.html", itemForecastPageData{
		navData:  nav,
		Item:     *it,
		Result:   res,
		ChartSVG: forecastSVG(res),
	})
}
