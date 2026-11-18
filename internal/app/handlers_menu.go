package app

import (
	"net/http"
	"strconv"
	"strings"

	"menuapp/internal/store"
)

type menusPageData struct {
	navData
	Menus []store.Menu
	Error string
}

func (a *App) handleMenus(w http.ResponseWriter, r *http.Request) {
	u, rest, ok := a.requireRestaurant(w, r)
	if !ok {
		return
	}
	menus, err := a.Store.MenusByRestaurant(rest.ID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	nav := a.nav(r, u, rest, flashFromQuery(r))
	a.render(w, "menus.html", menusPageData{navData: nav, Menus: menus})
}

func (a *App) handleMenuCreate(w http.ResponseWriter, r *http.Request) {
	_, rest, ok := a.requireRestaurant(w, r)
	if !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		http.Redirect(w, r, "/r/"+itoa(rest.ID)+"/menus", http.StatusSeeOther)
		return
	}
	id, err := a.Store.CreateMenu(rest.ID, name)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/r/"+itoa(rest.ID)+"/menus/"+itoa(id), http.StatusSeeOther)
}

type sectionWithItems struct {
	Section store.Section
	Items   []itemWithCost
}

type itemWithCost struct {
	Item        store.Item
	CostCents   int64
	MarginCents int64
	MarginPct   float64
}

type menuDetailPageData struct {
	navData
	Menu     store.Menu
	Sections []sectionWithItems
}

func (a *App) handleMenuDetail(w http.ResponseWriter, r *http.Request) {
	u, rest, ok := a.requireRestaurant(w, r)
	if !ok {
		return
	}
	mid, ok := pathID(r, "mid")
	if !ok {
		http.NotFound(w, r)
		return
	}
	menu, err := a.Store.MenuByID(mid)
	if err != nil || menu.RestaurantID != rest.ID {
		http.NotFound(w, r)
		return
	}
	sections, err := a.Store.SectionsByMenu(menu.ID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	var out []sectionWithItems
	for _, sec := range sections {
		items, err := a.Store.ItemsBySection(sec.ID)
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		var withCost []itemWithCost
		for _, it := range items {
			cost, _ := a.Store.PlateCostCents(it.ID)
			margin := it.PriceCents - cost
			marginPct := 0.0
			if it.PriceCents > 0 {
				marginPct = float64(margin) / float64(it.PriceCents) * 100
			}
			withCost = append(withCost, itemWithCost{Item: it, CostCents: cost, MarginCents: margin, MarginPct: marginPct})
		}
		out = append(out, sectionWithItems{Section: sec, Items: withCost})
	}

	nav := a.nav(r, u, rest, flashFromQuery(r))
	a.render(w, "menu_detail.html", menuDetailPageData{navData: nav, Menu: *menu, Sections: out})
}

func (a *App) handleSectionCreate(w http.ResponseWriter, r *http.Request) {
	_, rest, ok := a.requireRestaurant(w, r)
	if !ok {
		return
	}
	mid, ok := pathID(r, "mid")
	if !ok {
		http.NotFound(w, r)
		return
	}
	menu, err := a.Store.MenuByID(mid)
	if err != nil || menu.RestaurantID != rest.ID {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name != "" {
		a.Store.CreateSection(menu.ID, name)
	}
	http.Redirect(w, r, "/r/"+itoa(rest.ID)+"/menus/"+itoa(menu.ID), http.StatusSeeOther)
}

func (a *App) handleItemCreate(w http.ResponseWriter, r *http.Request) {
	_, rest, ok := a.requireRestaurant(w, r)
	if !ok {
		return
	}
	sid, ok := pathID(r, "sid")
	if !ok {
		http.NotFound(w, r)
		return
	}
	sec, err := a.Store.SectionByID(sid)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	menu, err := a.Store.MenuByID(sec.MenuID)
	if err != nil || menu.RestaurantID != rest.ID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	desc := strings.TrimSpace(r.FormValue("description"))
	priceCents := parsePriceToCents(r.FormValue("price"))
	if name != "" {
		a.Store.CreateItem(rest.ID, sec.ID, name, desc, priceCents)
	}
	http.Redirect(w, r, "/r/"+itoa(rest.ID)+"/menus/"+itoa(menu.ID), http.StatusSeeOther)
}

// parsePriceToCents parses a user-entered dollar string like "12.50" into cents.
func parsePriceToCents(s string) int64 {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "$")
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return int64(f*100 + 0.5)
}

// parsePriceToCentsFloat parses a user-entered dollar string into cents
// *without* rounding to a whole cent — used for ingredient unit costs, which
// are routinely sub-cent (e.g. $0.012 per gram).
func parsePriceToCentsFloat(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "$")
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f * 100
}
