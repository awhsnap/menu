package app

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type salesImportSummary struct {
	Filename     string
	RowCount     int
	MatchedCount int
}

type salesPageData struct {
	navData
	Summary *salesImportSummary
	Error   string
}

func (a *App) handleSales(w http.ResponseWriter, r *http.Request) {
	u, rest, ok := a.requireRestaurant(w, r)
	if !ok {
		return
	}
	nav := a.nav(r, u, rest, flashFromQuery(r))
	a.render(w, "sales.html", salesPageData{navData: nav})
}

// handleSalesImport accepts a CSV with columns item/item_name, quantity/qty,
// date/sale_date (YYYY-MM-DD). Rows whose item name matches an existing menu
// item (case-insensitive) are linked for forecasting & menu engineering;
// unmatched rows are still stored so nothing from the file is silently lost.
func (a *App) handleSalesImport(w http.ResponseWriter, r *http.Request) {
	u, rest, ok := a.requireRestaurant(w, r)
	if !ok {
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		nav := a.nav(r, u, rest, "")
		a.render(w, "sales.html", salesPageData{navData: nav, Error: "Could not read the uploaded file."})
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		nav := a.nav(r, u, rest, "")
		a.render(w, "sales.html", salesPageData{navData: nav, Error: "Choose a CSV file to upload."})
		return
	}
	defer file.Close()

	rows, parseErr := parseSalesCSV(file)
	if parseErr != nil {
		nav := a.nav(r, u, rest, "")
		a.render(w, "sales.html", salesPageData{navData: nav, Error: "Could not parse that CSV: " + parseErr.Error()})
		return
	}
	if len(rows) == 0 {
		nav := a.nav(r, u, rest, "")
		a.render(w, "sales.html", salesPageData{navData: nav, Error: "The file had no data rows. Expected columns: item, quantity, date."})
		return
	}

	importID, err := a.Store.CreateSalesImport(rest.ID, header.Filename)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	matched := 0
	for _, row := range rows {
		ok, err := a.Store.InsertSalesLine(rest.ID, importID, row.ItemName, row.Quantity, row.Date)
		if err != nil {
			continue
		}
		if ok {
			matched++
		}
	}
	a.Store.FinalizeSalesImport(importID, len(rows), matched)

	nav := a.nav(r, u, rest, "")
	a.render(w, "sales.html", salesPageData{
		navData: nav,
		Summary: &salesImportSummary{Filename: header.Filename, RowCount: len(rows), MatchedCount: matched},
	})
}

type salesCSVRow struct {
	ItemName string
	Quantity float64
	Date     string // YYYY-MM-DD
}

func parseSalesCSV(r io.Reader) ([]salesCSVRow, error) {
	cr := csv.NewReader(r)
	cr.TrimLeadingSpace = true
	records, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) < 2 {
		return nil, nil
	}
	header := records[0]
	col := map[string]int{}
	for i, h := range header {
		col[strings.ToLower(strings.TrimSpace(h))] = i
	}
	itemCol, ok1 := firstCol(col, "item", "item_name", "name")
	qtyCol, ok2 := firstCol(col, "quantity", "qty")
	dateCol, ok3 := firstCol(col, "date", "sale_date", "period")
	if !ok1 || !ok2 || !ok3 {
		return nil, fmt.Errorf("expected columns item, quantity, date (found: %s)", strings.Join(header, ", "))
	}

	var out []salesCSVRow
	for _, rec := range records[1:] {
		if itemCol >= len(rec) || qtyCol >= len(rec) || dateCol >= len(rec) {
			continue
		}
		name := strings.TrimSpace(rec[itemCol])
		if name == "" {
			continue
		}
		qty, err := strconv.ParseFloat(strings.TrimSpace(rec[qtyCol]), 64)
		if err != nil {
			continue
		}
		date, err := normalizeDate(strings.TrimSpace(rec[dateCol]))
		if err != nil {
			continue
		}
		out = append(out, salesCSVRow{ItemName: name, Quantity: qty, Date: date})
	}
	return out, nil
}

func firstCol(col map[string]int, names ...string) (int, bool) {
	for _, n := range names {
		if i, ok := col[n]; ok {
			return i, true
		}
	}
	return 0, false
}

var dateLayouts = []string{"2006-01-02", "01/02/2006", "1/2/2006", "2006/01/02"}

func normalizeDate(s string) (string, error) {
	for _, layout := range dateLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Format("2006-01-02"), nil
		}
	}
	return "", fmt.Errorf("unrecognized date %q", s)
}
