package app

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"

	"menuapp/internal/store"
)

var pageFiles = []string{
	"login.html",
	"register.html",
	"restaurants.html",
	"dashboard.html",
	"menus.html",
	"menu_detail.html",
	"item_detail.html",
	"ingredients.html",
	"sales.html",
	"analytics.html",
	"item_forecast.html",
	"settings.html",
}

var funcMap = template.FuncMap{
	"money": func(cents int64) string {
		neg := ""
		if cents < 0 {
			neg = "-"
			cents = -cents
		}
		return fmt.Sprintf("%s$%d.%02d", neg, cents/100, cents%100)
	},
	"pct": func(v float64) string {
		return fmt.Sprintf("%.1f%%", v)
	},
	"num": func(v float64) string {
		return fmt.Sprintf("%.1f", v)
	},
	"add": func(a, b int) int { return a + b },
	"div2": func(cents int64) float64 { return float64(cents) / 100.0 },
	"dollarsF": func(cents float64) float64 { return cents / 100.0 },
	"unitcost": func(cents float64) string {
		return fmt.Sprintf("$%.4f", cents/100.0)
	},
	"lineCost": func(rl store.RecipeLine) int64 {
		yield := rl.YieldPct
		if yield <= 0 {
			yield = 100
		}
		return int64(float64(rl.UnitCostCents)*rl.Quantity*(100.0/yield) + 0.5)
	},
}

func (a *App) loadTemplates(webFS fs.FS) error {
	a.templates = map[string]*template.Template{}
	for _, page := range pageFiles {
		t, err := template.New("layout.html").Funcs(funcMap).ParseFS(webFS, "templates/layout.html", "templates/"+page)
		if err != nil {
			return fmt.Errorf("parse %s: %w", page, err)
		}
		a.templates[page] = t
	}
	return nil
}

// navData is embedded in every page's data struct so the layout can render
// the top nav (signed-in user, restaurant switcher).
type navData struct {
	User        *store.User
	Restaurants []store.Restaurant
	Active      *store.Restaurant
	Flash       string
}

func (a *App) nav(r *http.Request, u *store.User, active *store.Restaurant, flash string) navData {
	var rests []store.Restaurant
	if u != nil {
		rests, _ = a.Store.RestaurantsByOwner(u.ID)
	}
	return navData{User: u, Restaurants: rests, Active: active, Flash: flash}
}

func (a *App) render(w http.ResponseWriter, page string, data any) {
	t, ok := a.templates[page]
	if !ok {
		http.Error(w, "template not found: "+page, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func flashFromQuery(r *http.Request) string {
	return r.URL.Query().Get("flash")
}
