package app

import (
	"net/http"

	"menuapp/internal/store"
)

type marginAlert struct {
	Item      store.Item
	CostCents int64
	MarginPct float64
}

type dashboardPageData struct {
	navData
	MenuCount       int
	ItemCount       int
	ActiveItemCount int
	IngredientCount int
	HasSales        bool
	MarginAlerts    []marginAlert
}

func (a *App) handleDashboard(w http.ResponseWriter, r *http.Request) {
	u, rest, ok := a.requireRestaurant(w, r)
	if !ok {
		return
	}

	menus, err := a.Store.MenusByRestaurant(rest.ID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	items, err := a.Store.ItemsByRestaurant(rest.ID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	ingredients, err := a.Store.IngredientsByRestaurant(rest.ID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	hasSales, _ := a.Store.HasAnySales(rest.ID)

	activeCount := 0
	var alerts []marginAlert
	for _, it := range items {
		if it.Active {
			activeCount++
		}
		if it.PriceCents == 0 {
			continue
		}
		cost, _ := a.Store.PlateCostCents(it.ID)
		if cost == 0 {
			continue
		}
		costPct := float64(cost) / float64(it.PriceCents) * 100
		if costPct > rest.TargetFoodCostPct {
			alerts = append(alerts, marginAlert{Item: it, CostCents: cost, MarginPct: 100 - costPct})
		}
	}

	nav := a.nav(r, u, rest, flashFromQuery(r))
	a.render(w, "dashboard.html", dashboardPageData{
		navData:         nav,
		MenuCount:       len(menus),
		ItemCount:       len(items),
		ActiveItemCount: activeCount,
		IngredientCount: len(ingredients),
		HasSales:        hasSales,
		MarginAlerts:    alerts,
	})
}
