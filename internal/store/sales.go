package store

import "strings"

func (s *Store) CreateSalesImport(restaurantID int64, filename string) (int64, error) {
	res, err := s.DB.Exec(`INSERT INTO sales_imports (restaurant_id, filename) VALUES (?, ?)`, restaurantID, filename)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) FinalizeSalesImport(importID int64, rowCount, matchedCount int) error {
	_, err := s.DB.Exec(`UPDATE sales_imports SET row_count = ?, matched_count = ? WHERE id = ?`, rowCount, matchedCount, importID)
	return err
}

// InsertSalesLine looks up the item by (case-insensitive) name within the
// restaurant so the forecast/analytics engine can attribute sales to a menu
// item; unmatched rows are still stored (item_id NULL) so nothing is silently
// dropped, but they won't feed per-item forecasting.
func (s *Store) InsertSalesLine(restaurantID, importID int64, itemName string, quantity float64, saleDate string) (matched bool, err error) {
	var itemID *int64
	row := s.DB.QueryRow(`SELECT id FROM items WHERE restaurant_id = ? AND lower(name) = lower(?)`, restaurantID, strings.TrimSpace(itemName))
	var id int64
	if scanErr := row.Scan(&id); scanErr == nil {
		itemID = &id
		matched = true
	}
	_, err = s.DB.Exec(`INSERT INTO sales_lines (restaurant_id, import_id, item_id, item_name, quantity, sale_date) VALUES (?, ?, ?, ?, ?, ?)`,
		restaurantID, importID, itemID, itemName, quantity, saleDate)
	return matched, err
}

type ItemSales struct {
	SaleDate string
	Quantity float64
}

// SalesByItem returns all (date, quantity) rows for one item, ordered by date.
func (s *Store) SalesByItem(itemID int64) ([]ItemSales, error) {
	rows, err := s.DB.Query(`SELECT sale_date, SUM(quantity) FROM sales_lines WHERE item_id = ? GROUP BY sale_date ORDER BY sale_date`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ItemSales
	for rows.Next() {
		var is ItemSales
		if err := rows.Scan(&is.SaleDate, &is.Quantity); err != nil {
			return nil, err
		}
		out = append(out, is)
	}
	return out, rows.Err()
}

type ItemSalesTotal struct {
	ItemID   int64
	Quantity float64
}

// SalesTotalsByRestaurant sums quantity sold per item across all imported sales.
func (s *Store) SalesTotalsByRestaurant(restaurantID int64) (map[int64]float64, error) {
	rows, err := s.DB.Query(`SELECT item_id, SUM(quantity) FROM sales_lines WHERE restaurant_id = ? AND item_id IS NOT NULL GROUP BY item_id`, restaurantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64]float64{}
	for rows.Next() {
		var id int64
		var qty float64
		if err := rows.Scan(&id, &qty); err != nil {
			return nil, err
		}
		out[id] = qty
	}
	return out, rows.Err()
}

func (s *Store) HasAnySales(restaurantID int64) (bool, error) {
	var n int
	row := s.DB.QueryRow(`SELECT COUNT(1) FROM sales_lines WHERE restaurant_id = ?`, restaurantID)
	if err := row.Scan(&n); err != nil {
		return false, err
	}
	return n > 0, nil
}
