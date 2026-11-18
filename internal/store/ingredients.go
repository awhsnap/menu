package store

import (
	"database/sql"
	"errors"
)

func (s *Store) CreateIngredient(restaurantID int64, name, unit string, unitCostCents float64, supplier string, yieldPct float64) (int64, error) {
	res, err := s.DB.Exec(`INSERT INTO ingredients (restaurant_id, name, unit, unit_cost_cents, supplier, yield_pct) VALUES (?, ?, ?, ?, ?, ?)`,
		restaurantID, name, unit, unitCostCents, supplier, yieldPct)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	s.recordPriceHistory(id, unitCostCents)
	return id, nil
}

func (s *Store) recordPriceHistory(ingredientID int64, unitCostCents float64) {
	s.DB.Exec(`INSERT INTO ingredient_price_history (ingredient_id, unit_cost_cents) VALUES (?, ?)`, ingredientID, unitCostCents)
}

// UpdateIngredientCost updates the unit cost and logs price history when it changed.
// Returns true if the cost actually changed (used to trigger margin-alert checks).
func (s *Store) UpdateIngredientCost(ingredientID int64, newUnitCostCents float64) (bool, error) {
	ing, err := s.IngredientByID(ingredientID)
	if err != nil {
		return false, err
	}
	if ing.UnitCostCents == newUnitCostCents {
		return false, nil
	}
	if _, err := s.DB.Exec(`UPDATE ingredients SET unit_cost_cents = ?, updated_at = datetime('now') WHERE id = ?`, newUnitCostCents, ingredientID); err != nil {
		return false, err
	}
	s.recordPriceHistory(ingredientID, newUnitCostCents)
	return true, nil
}

func (s *Store) UpdateIngredient(id int64, name, unit string, supplier string, yieldPct float64) error {
	_, err := s.DB.Exec(`UPDATE ingredients SET name=?, unit=?, supplier=?, yield_pct=?, updated_at=datetime('now') WHERE id=?`,
		name, unit, supplier, yieldPct, id)
	return err
}

func (s *Store) IngredientsByRestaurant(restaurantID int64) ([]Ingredient, error) {
	rows, err := s.DB.Query(`SELECT id, restaurant_id, name, unit, unit_cost_cents, supplier, yield_pct, updated_at
		FROM ingredients WHERE restaurant_id = ? ORDER BY name`, restaurantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Ingredient
	for rows.Next() {
		var ing Ingredient
		if err := rows.Scan(&ing.ID, &ing.RestaurantID, &ing.Name, &ing.Unit, &ing.UnitCostCents, &ing.Supplier, &ing.YieldPct, &ing.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, ing)
	}
	return out, rows.Err()
}

func (s *Store) IngredientByID(id int64) (*Ingredient, error) {
	row := s.DB.QueryRow(`SELECT id, restaurant_id, name, unit, unit_cost_cents, supplier, yield_pct, updated_at
		FROM ingredients WHERE id = ?`, id)
	var ing Ingredient
	if err := row.Scan(&ing.ID, &ing.RestaurantID, &ing.Name, &ing.Unit, &ing.UnitCostCents, &ing.Supplier, &ing.YieldPct, &ing.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &ing, nil
}

func (s *Store) DeleteIngredient(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM ingredients WHERE id = ?`, id)
	return err
}

// --- Recipe lines ---

func (s *Store) AddRecipeLine(itemID, ingredientID int64, quantity float64) (int64, error) {
	res, err := s.DB.Exec(`INSERT INTO recipe_lines (item_id, ingredient_id, quantity) VALUES (?, ?, ?)`, itemID, ingredientID, quantity)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) DeleteRecipeLine(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM recipe_lines WHERE id = ?`, id)
	return err
}

func (s *Store) RecipeLinesByItem(itemID int64) ([]RecipeLine, error) {
	rows, err := s.DB.Query(`SELECT rl.id, rl.item_id, rl.ingredient_id, rl.quantity, i.name, i.unit, i.unit_cost_cents, i.yield_pct
		FROM recipe_lines rl JOIN ingredients i ON i.id = rl.ingredient_id
		WHERE rl.item_id = ? ORDER BY rl.id`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RecipeLine
	for rows.Next() {
		var rl RecipeLine
		if err := rows.Scan(&rl.ID, &rl.ItemID, &rl.IngredientID, &rl.Quantity, &rl.IngredientName, &rl.Unit, &rl.UnitCostCents, &rl.YieldPct); err != nil {
			return nil, err
		}
		out = append(out, rl)
	}
	return out, rows.Err()
}

// PlateCostCents computes the total ingredient cost for one item, accounting
// for each ingredient's yield/trim-loss factor (cost is grossed up by
// 100/yield_pct to reflect usable cost per purchased unit).
func (s *Store) PlateCostCents(itemID int64) (int64, error) {
	lines, err := s.RecipeLinesByItem(itemID)
	if err != nil {
		return 0, err
	}
	var total float64
	for _, l := range lines {
		yield := l.YieldPct
		if yield <= 0 {
			yield = 100
		}
		lineCost := l.UnitCostCents * l.Quantity * (100.0 / yield)
		total += lineCost
	}
	return int64(total + 0.5), nil
}
