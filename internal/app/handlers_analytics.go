package app

import (
	"html/template"
	"net/http"
	"sort"

	"menuapp/internal/analytics"
)

type analyticsPageData struct {
	navData
	Items     []analytics.ItemFigures
	ChartSVG  template.HTML
	HasSales  bool
}

func (a *App) handleAnalytics(w http.ResponseWriter, r *http.Request) {
	u, rest, ok := a.requireRestaurant(w, r)
	if !ok {
		return
	}
	items, err := a.Store.ItemsByRestaurant(rest.ID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	totals, err := a.Store.SalesTotalsByRestaurant(rest.ID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	hasSales, _ := a.Store.HasAnySales(rest.ID)

	var figures []analytics.ItemFigures
	for _, it := range items {
		cost, _ := a.Store.PlateCostCents(it.ID)
		margin := it.PriceCents - cost
		marginPct := 0.0
		if it.PriceCents > 0 {
			marginPct = float64(margin) / float64(it.PriceCents) * 100
		}
		figures = append(figures, analytics.ItemFigures{
			ItemID:       it.ID,
			Name:         it.Name,
			PriceCents:   it.PriceCents,
			CostCents:    cost,
			QuantitySold: totals[it.ID],
			MarginCents:  margin,
			MarginPct:    marginPct,
		})
	}
	classified := analytics.Classify(figures)
	sort.Slice(classified, func(i, j int) bool { return classified[i].QuantitySold > classified[j].QuantitySold })

	nav := a.nav(r, u, rest, "")
	a.render(w, "analytics.html", analyticsPageData{
		navData:  nav,
		Items:    classified,
		ChartSVG: quadrantSVG(classified),
		HasSales: hasSales,
	})
}
