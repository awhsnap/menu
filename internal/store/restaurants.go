package store

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var slugInvalid = regexp.MustCompile(`[^a-z0-9]+`)

func Slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = slugInvalid.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "restaurant"
	}
	return s
}

// CreateRestaurant picks a unique slug (appending -2, -3, ... on collision)
// and inserts the restaurant.
func (s *Store) CreateRestaurant(ownerUserID int64, name string, targetFoodCostPct float64) (*Restaurant, error) {
	base := Slugify(name)
	slug := base
	for i := 2; ; i++ {
		var exists int
		row := s.DB.QueryRow(`SELECT COUNT(1) FROM restaurants WHERE slug = ?`, slug)
		if err := row.Scan(&exists); err != nil {
			return nil, err
		}
		if exists == 0 {
			break
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
	res, err := s.DB.Exec(`INSERT INTO restaurants (owner_user_id, name, slug, target_food_cost_pct) VALUES (?, ?, ?, ?)`,
		ownerUserID, name, slug, targetFoodCostPct)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return s.RestaurantByID(id)
}

func (s *Store) RestaurantByID(id int64) (*Restaurant, error) {
	row := s.DB.QueryRow(`SELECT id, owner_user_id, name, slug, currency, target_food_cost_pct, created_at FROM restaurants WHERE id = ?`, id)
	var r Restaurant
	if err := row.Scan(&r.ID, &r.OwnerUserID, &r.Name, &r.Slug, &r.Currency, &r.TargetFoodCostPct, &r.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &r, nil
}

func (s *Store) RestaurantsByOwner(ownerUserID int64) ([]Restaurant, error) {
	rows, err := s.DB.Query(`SELECT id, owner_user_id, name, slug, currency, target_food_cost_pct, created_at
		FROM restaurants WHERE owner_user_id = ? ORDER BY created_at ASC`, ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Restaurant
	for rows.Next() {
		var r Restaurant
		if err := rows.Scan(&r.ID, &r.OwnerUserID, &r.Name, &r.Slug, &r.Currency, &r.TargetFoodCostPct, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) UpdateRestaurant(id int64, name string, targetFoodCostPct float64) error {
	_, err := s.DB.Exec(`UPDATE restaurants SET name = ?, target_food_cost_pct = ? WHERE id = ?`, name, targetFoodCostPct, id)
	return err
}
