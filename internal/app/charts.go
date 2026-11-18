package app

import (
	"fmt"
	"html/template"
	"strings"

	"menuapp/internal/analytics"
	"menuapp/internal/forecast"
)

// forecastSVG renders a small server-generated bar chart: recent sales
// history in one color, the projected horizon in another. No JS, no CDN —
// just inline SVG so the MVP has zero client-side dependencies.
func forecastSVG(res forecast.Result) template.HTML {
	const width, height, padding = 720.0, 220.0, 28.0
	const barGap = 4.0

	hist := res.History
	if len(hist) > 21 {
		hist = hist[len(hist)-21:] // cap history shown so the chart stays readable
	}
	total := len(hist) + len(res.Forecast)
	if total == 0 {
		return template.HTML(`<p class="muted">Not enough sales data yet to chart.</p>`)
	}

	maxVal := 0.0
	for _, p := range hist {
		if p.Quantity > maxVal {
			maxVal = p.Quantity
		}
	}
	for _, p := range res.Forecast {
		if p.Quantity > maxVal {
			maxVal = p.Quantity
		}
	}
	if maxVal == 0 {
		maxVal = 1
	}

	plotW := width - padding*2
	plotH := height - padding*2
	barW := (plotW - float64(total-1)*barGap) / float64(total)
	if barW < 2 {
		barW = 2
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<svg viewBox="0 0 %.0f %.0f" xmlns="http://www.w3.org/2000/svg" class="chart">`, width, height)
	// baseline
	fmt.Fprintf(&b, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#d8dce3" stroke-width="1"/>`,
		padding, height-padding, width-padding, height-padding)

	x := padding
	for _, p := range hist {
		h := (p.Quantity / maxVal) * plotH
		y := height - padding - h
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="#6b7cff" rx="1.5"><title>%s: %.1f</title></rect>`,
			x, y, barW, h, p.Date, p.Quantity)
		x += barW + barGap
	}
	for _, p := range res.Forecast {
		h := (p.Quantity / maxVal) * plotH
		y := height - padding - h
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="#ffb74d" rx="1.5"><title>%s (forecast): %.1f</title></rect>`,
			x, y, barW, h, p.Date, p.Quantity)
		x += barW + barGap
	}
	b.WriteString(`</svg>`)
	return template.HTML(b.String())
}

// quadrantSVG renders the classic 2x2 menu-engineering scatter: popularity
// on the x-axis, margin (profitability) on the y-axis, one dot per item.
func quadrantSVG(items []analytics.ItemFigures) template.HTML {
	const size, padding = 480.0, 40.0

	var ranked []analytics.ItemFigures
	for _, it := range items {
		if it.Quadrant != analytics.Unranked {
			ranked = append(ranked, it)
		}
	}
	if len(ranked) == 0 {
		return template.HTML(`<p class="muted">Import some sales data to see the menu-engineering chart.</p>`)
	}

	maxPop := 0.0
	maxMargin, minMargin := 0.0, 0.0
	for i, it := range ranked {
		if it.PopularityIndex > maxPop {
			maxPop = it.PopularityIndex
		}
		m := float64(it.MarginCents)
		if i == 0 || m > maxMargin {
			maxMargin = m
		}
		if i == 0 || m < minMargin {
			minMargin = m
		}
	}
	if maxPop == 0 {
		maxPop = 1
	}
	marginRange := maxMargin - minMargin
	if marginRange == 0 {
		marginRange = 1
	}

	plot := size - padding*2
	colorFor := func(q analytics.Quadrant) string {
		switch q {
		case analytics.Star:
			return "#2e7d32"
		case analytics.Plowhorse:
			return "#1976d2"
		case analytics.Puzzle:
			return "#f57c00"
		default:
			return "#c62828"
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<svg viewBox="0 0 %.0f %.0f" xmlns="http://www.w3.org/2000/svg" class="chart">`, size, size)
	// quadrant divider lines at the midpoint (popularity = avg, margin = avg)
	midX := padding + plot/2
	midY := padding + plot/2
	fmt.Fprintf(&b, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#e3e6ec" stroke-width="1"/>`, midX, padding, midX, size-padding)
	fmt.Fprintf(&b, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#e3e6ec" stroke-width="1"/>`, padding, midY, size-padding, midY)
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="none" stroke="#d8dce3"/>`, padding, padding, plot, plot)

	for _, it := range ranked {
		px := padding + (it.PopularityIndex/ (maxPop*1.15))*plot
		py := size - padding - ((float64(it.MarginCents)-minMargin)/(marginRange*1.15+0.0001))*plot
		if px > size-padding {
			px = size - padding
		}
		if py < padding {
			py = padding
		}
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="6" fill="%s" fill-opacity="0.85"><title>%s — %s (qty %.0f, margin %s)</title></circle>`,
			px, py, colorFor(it.Quadrant), escapeAttr(it.Name), it.Quadrant, it.QuantitySold, moneyStr(it.MarginCents))
	}
	b.WriteString(`</svg>`)
	return template.HTML(b.String())
}

func moneyStr(cents int64) string {
	neg := ""
	if cents < 0 {
		neg = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s$%d.%02d", neg, cents/100, cents%100)
}

func escapeAttr(s string) string {
	r := strings.NewReplacer(`&`, "&amp;", `"`, "&quot;", `<`, "&lt;", `>`, "&gt;")
	return r.Replace(s)
}
