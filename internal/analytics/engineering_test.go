package analytics

import "testing"

func TestClassify_NoSales(t *testing.T) {
	items := []ItemFigures{
		{ItemID: 1, Name: "A", PriceCents: 1000, MarginCents: 500},
		{ItemID: 2, Name: "B", PriceCents: 1000, MarginCents: 200},
	}
	out := Classify(items)
	for _, it := range out {
		if it.Quadrant != Unranked {
			t.Errorf("expected Unranked with no sales, got %s for %s", it.Quadrant, it.Name)
		}
	}
}

func TestClassify_FourQuadrants(t *testing.T) {
	// Two items high popularity (above avg qty), two low; two high margin, two low.
	items := []ItemFigures{
		{ItemID: 1, Name: "Star", PriceCents: 1500, MarginCents: 1000, QuantitySold: 100}, // high pop, high margin
		{ItemID: 2, Name: "Plowhorse", PriceCents: 1000, MarginCents: 100, QuantitySold: 100}, // high pop, low margin
		{ItemID: 3, Name: "Puzzle", PriceCents: 1500, MarginCents: 1000, QuantitySold: 5},  // low pop, high margin
		{ItemID: 4, Name: "Dog", PriceCents: 1000, MarginCents: 100, QuantitySold: 5},      // low pop, low margin
	}
	out := Classify(items)
	byName := map[string]ItemFigures{}
	for _, it := range out {
		byName[it.Name] = it
	}
	if byName["Star"].Quadrant != Star {
		t.Errorf("expected Star, got %s", byName["Star"].Quadrant)
	}
	if byName["Plowhorse"].Quadrant != Plowhorse {
		t.Errorf("expected Plowhorse, got %s", byName["Plowhorse"].Quadrant)
	}
	if byName["Puzzle"].Quadrant != Puzzle {
		t.Errorf("expected Puzzle, got %s", byName["Puzzle"].Quadrant)
	}
	if byName["Dog"].Quadrant != Dog {
		t.Errorf("expected Dog, got %s", byName["Dog"].Quadrant)
	}
}

func TestClassify_ZeroSalesItemUnrankedNotDog(t *testing.T) {
	items := []ItemFigures{
		{ItemID: 1, Name: "Sold", PriceCents: 1000, MarginCents: 500, QuantitySold: 10},
		{ItemID: 2, Name: "NoSalesYet", PriceCents: 1000, MarginCents: 500, QuantitySold: 0},
	}
	out := Classify(items)
	for _, it := range out {
		if it.Name == "NoSalesYet" && it.Quadrant != Unranked {
			t.Errorf("item with zero sales should be Unranked, got %s", it.Quadrant)
		}
	}
}
