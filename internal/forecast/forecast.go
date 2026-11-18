// Package forecast implements a small, dependency-free demand forecast:
// a trailing moving average blended with a linear trend fit over the sales
// history, projected forward. It is intentionally simple (no external ML
// libraries) so it runs instantly in-process for the MVP; see the PRD's v2
// AI-assisted forecasting for a more sophisticated model later.
package forecast

import (
	"math"
	"sort"
	"time"

	"menuapp/internal/store"
)

const dateLayout = "2006-01-02"
const movingAverageWindow = 7

type Point struct {
	Date     string
	Quantity float64
}

type Result struct {
	History        []Point // the daily series actually observed (gaps filled with 0)
	Forecast       []Point // projected daily demand for the horizon
	DailyAverage   float64 // mean daily quantity over history
	TrendPerDay    float64 // slope of the linear fit (units/day)
	HasEnoughData  bool    // true once there are at least 3 distinct sale dates
	TotalForecast  float64 // sum of the forecast horizon
}

// Forecast builds a daily series from sparse (date, quantity) sales rows,
// fills gaps with zero, fits a least-squares line plus a trailing moving
// average, and projects `horizonDays` forward.
func Forecast(sales []store.ItemSales, horizonDays int) Result {
	if horizonDays <= 0 {
		horizonDays = 7
	}
	if len(sales) == 0 {
		return Result{}
	}

	// Parse + sort by date.
	type parsed struct {
		t   time.Time
		qty float64
	}
	var rows []parsed
	for _, s := range sales {
		t, err := time.Parse(dateLayout, s.SaleDate)
		if err != nil {
			continue
		}
		rows = append(rows, parsed{t: t, qty: s.Quantity})
	}
	if len(rows) == 0 {
		return Result{}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].t.Before(rows[j].t) })

	start, end := rows[0].t, rows[len(rows)-1].t
	byDate := map[string]float64{}
	for _, r := range rows {
		byDate[r.t.Format(dateLayout)] += r.qty
	}

	var history []Point
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		key := d.Format(dateLayout)
		history = append(history, Point{Date: key, Quantity: byDate[key]})
	}

	n := len(history)
	distinctDates := len(byDate)

	// Linear regression (least squares) over day index -> quantity.
	var sumX, sumY, sumXY, sumXX float64
	for i, p := range history {
		x := float64(i)
		sumX += x
		sumY += p.Quantity
		sumXY += x * p.Quantity
		sumXX += x * x
	}
	var slope float64
	denom := float64(n)*sumXX - sumX*sumX
	if n > 1 && denom != 0 {
		slope = (float64(n)*sumXY - sumX*sumY) / denom
	}

	// Trailing moving average as the forecast's baseline level (more robust
	// to a single noisy recent day than using the regression line alone).
	window := movingAverageWindow
	if window > n {
		window = n
	}
	var maSum float64
	for i := n - window; i < n; i++ {
		maSum += history[i].Quantity
	}
	baseline := maSum / float64(window)

	dailyAvg := sumY / float64(n)

	forecastPts := make([]Point, 0, horizonDays)
	var total float64
	for i := 1; i <= horizonDays; i++ {
		d := end.AddDate(0, 0, i)
		val := baseline + slope*float64(i)
		if val < 0 {
			val = 0
		}
		total += val
		forecastPts = append(forecastPts, Point{Date: d.Format(dateLayout), Quantity: math.Round(val*100) / 100})
	}

	return Result{
		History:       history,
		Forecast:      forecastPts,
		DailyAverage:  dailyAvg,
		TrendPerDay:   slope,
		HasEnoughData: distinctDates >= 3,
		TotalForecast: total,
	}
}
