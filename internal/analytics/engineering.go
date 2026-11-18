// Package analytics computes the classic "menu engineering" matrix
// (stars / plowhorses / puzzles / dogs) and margin figures used by the
// dashboard. See PRD section 5.5 (COST-5).
package analytics

type Quadrant string

const (
	Star      Quadrant = "Star"       // high popularity, high profitability
	Plowhorse Quadrant = "Plowhorse"  // high popularity, low profitability
	Puzzle    Quadrant = "Puzzle"     // low popularity, high profitability
	Dog       Quadrant = "Dog"        // low popularity, low profitability
	Unranked  Quadrant = "Unranked"   // not enough sales data yet
)

type ItemFigures struct {
	ItemID          int64
	Name            string
	PriceCents      int64
	CostCents       int64
	QuantitySold    float64
	MarginCents     int64   // price - cost
	MarginPct       float64 // margin / price * 100
	PopularityIndex float64 // this item's share of total units sold, relative to an even split
	Quadrant        Quadrant
}

// Classify runs the menu-engineering matrix over a restaurant's items.
// Items with zero recorded sales are still returned (so the dashboard can
// show them) but are marked Unranked rather than forced into "Dog", since
// zero sales usually means "no sales data imported yet", not "a bad dish".
func Classify(items []ItemFigures) []ItemFigures {
	n := len(items)
	if n == 0 {
		return items
	}

	var totalQty float64
	var weightedMarginSum float64 // sum(margin_i * qty_i)
	var totalQtySoldItems int
	for _, it := range items {
		totalQty += it.QuantitySold
		weightedMarginSum += float64(it.MarginCents) * it.QuantitySold
		if it.QuantitySold > 0 {
			totalQtySoldItems++
		}
	}

	if totalQty == 0 {
		// No sales data at all: can't rank anything.
		out := make([]ItemFigures, n)
		for i, it := range items {
			it.Quadrant = Unranked
			out[i] = it
		}
		return out
	}

	avgPopularity := totalQty / float64(n) // "even share" baseline used by the classic 70% rule's simpler cousin
	avgMargin := weightedMarginSum / totalQty

	out := make([]ItemFigures, n)
	for i, it := range items {
		it.PopularityIndex = it.QuantitySold / avgPopularity
		if it.QuantitySold == 0 {
			it.Quadrant = Unranked
			out[i] = it
			continue
		}
		highPopularity := it.QuantitySold >= avgPopularity
		highProfitability := float64(it.MarginCents) >= avgMargin
		switch {
		case highPopularity && highProfitability:
			it.Quadrant = Star
		case highPopularity && !highProfitability:
			it.Quadrant = Plowhorse
		case !highPopularity && highProfitability:
			it.Quadrant = Puzzle
		default:
			it.Quadrant = Dog
		}
		out[i] = it
	}
	return out
}
