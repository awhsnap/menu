package forecast

import (
	"testing"

	"menuapp/internal/store"
)

func TestForecast_EmptyInput(t *testing.T) {
	res := Forecast(nil, 7)
	if res.HasEnoughData {
		t.Fatalf("expected HasEnoughData=false for empty input")
	}
	if len(res.Forecast) != 0 {
		t.Fatalf("expected no forecast points, got %d", len(res.Forecast))
	}
}

func TestForecast_FlatSeries(t *testing.T) {
	sales := []store.ItemSales{
		{SaleDate: "2026-01-01", Quantity: 10},
		{SaleDate: "2026-01-02", Quantity: 10},
		{SaleDate: "2026-01-03", Quantity: 10},
		{SaleDate: "2026-01-04", Quantity: 10},
	}
	res := Forecast(sales, 3)
	if !res.HasEnoughData {
		t.Fatalf("expected enough data")
	}
	if len(res.Forecast) != 3 {
		t.Fatalf("expected 3 forecast points, got %d", len(res.Forecast))
	}
	for _, p := range res.Forecast {
		if p.Quantity < 9.5 || p.Quantity > 10.5 {
			t.Errorf("flat series forecast should stay near 10, got %v on %s", p.Quantity, p.Date)
		}
	}
}

func TestForecast_FillsGaps(t *testing.T) {
	// Only two sparse sales dates three days apart; the series between them
	// should be filled with zeros, not skipped.
	sales := []store.ItemSales{
		{SaleDate: "2026-01-01", Quantity: 5},
		{SaleDate: "2026-01-04", Quantity: 5},
	}
	res := Forecast(sales, 1)
	if len(res.History) != 4 {
		t.Fatalf("expected 4 days of history (gaps filled), got %d", len(res.History))
	}
	if res.History[1].Quantity != 0 || res.History[2].Quantity != 0 {
		t.Fatalf("expected gap days to be zero-filled, got %+v", res.History)
	}
}

func TestForecast_UpwardTrendProjectsHigher(t *testing.T) {
	sales := []store.ItemSales{
		{SaleDate: "2026-01-01", Quantity: 5},
		{SaleDate: "2026-01-02", Quantity: 10},
		{SaleDate: "2026-01-03", Quantity: 15},
		{SaleDate: "2026-01-04", Quantity: 20},
	}
	res := Forecast(sales, 2)
	if res.TrendPerDay <= 0 {
		t.Fatalf("expected positive trend, got %v", res.TrendPerDay)
	}
	if res.Forecast[1].Quantity <= res.Forecast[0].Quantity {
		t.Fatalf("expected forecast to keep climbing: %+v", res.Forecast)
	}
}

func TestForecast_NeverNegative(t *testing.T) {
	sales := []store.ItemSales{
		{SaleDate: "2026-01-01", Quantity: 20},
		{SaleDate: "2026-01-02", Quantity: 10},
		{SaleDate: "2026-01-03", Quantity: 0},
		{SaleDate: "2026-01-04", Quantity: 0},
	}
	res := Forecast(sales, 14)
	for _, p := range res.Forecast {
		if p.Quantity < 0 {
			t.Fatalf("forecast quantity went negative: %+v", p)
		}
	}
}
