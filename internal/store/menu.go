package store

import (
	"database/sql"
	"errors"
)

// --- Menus ---

func (s *Store) CreateMenu(restaurantID int64, name string) (int64, error) {
	res, err := s.DB.Exec(`INSERT INTO menus (restaurant_id, name) VALUES (?, ?)`, restaurantID, name)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) MenusByRestaurant(restaurantID int64) ([]Menu, error) {
	rows, err := s.DB.Query(`SELECT id, restaurant_id, name, active, sort_order FROM menus WHERE restaurant_id = ? ORDER BY sort_order, id`, restaurantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Menu
	for rows.Next() {
		var m Menu
		var active int
		if err := rows.Scan(&m.ID, &m.RestaurantID, &m.Name, &active, &m.SortOrder); err != nil {
			return nil, err
		}
		m.Active = active != 0
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) MenuByID(id int64) (*Menu, error) {
	row := s.DB.QueryRow(`SELECT id, restaurant_id, name, active, sort_order FROM menus WHERE id = ?`, id)
	var m Menu
	var active int
	if err := row.Scan(&m.ID, &m.RestaurantID, &m.Name, &active, &m.SortOrder); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	m.Active = active != 0
	return &m, nil
}

// --- Sections ---

func (s *Store) CreateSection(menuID int64, name string) (int64, error) {
	var maxOrder sql.NullInt64
	s.DB.QueryRow(`SELECT MAX(sort_order) FROM sections WHERE menu_id = ?`, menuID).Scan(&maxOrder)
	order := int(maxOrder.Int64) + 1
	res, err := s.DB.Exec(`INSERT INTO sections (menu_id, name, sort_order) VALUES (?, ?, ?)`, menuID, name, order)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) SectionsByMenu(menuID int64) ([]Section, error) {
	rows, err := s.DB.Query(`SELECT id, menu_id, name, sort_order FROM sections WHERE menu_id = ? ORDER BY sort_order, id`, menuID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Section
	for rows.Next() {
		var sec Section
		if err := rows.Scan(&sec.ID, &sec.MenuID, &sec.Name, &sec.SortOrder); err != nil {
			return nil, err
		}
		out = append(out, sec)
	}
	return out, rows.Err()
}

func (s *Store) SectionByID(id int64) (*Section, error) {
	row := s.DB.QueryRow(`SELECT id, menu_id, name, sort_order FROM sections WHERE id = ?`, id)
	var sec Section
	if err := row.Scan(&sec.ID, &sec.MenuID, &sec.Name, &sec.SortOrder); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &sec, nil
}

// --- Items ---

func (s *Store) CreateItem(restaurantID, sectionID int64, name, description string, priceCents int64) (int64, error) {
	res, err := s.DB.Exec(`INSERT INTO items (restaurant_id, section_id, name, description, price_cents) VALUES (?, ?, ?, ?, ?)`,
		restaurantID, sectionID, name, description, priceCents)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) ItemsBySection(sectionID int64) ([]Item, error) {
	rows, err := s.DB.Query(`SELECT id, section_id, restaurant_id, name, description, price_cents, active, sort_order, updated_at
		FROM items WHERE section_id = ? ORDER BY sort_order, id`, sectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanItems(rows)
}

func (s *Store) ItemsByRestaurant(restaurantID int64) ([]Item, error) {
	rows, err := s.DB.Query(`SELECT id, section_id, restaurant_id, name, description, price_cents, active, sort_order, updated_at
		FROM items WHERE restaurant_id = ? ORDER BY id`, restaurantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanItems(rows)
}

func scanItems(rows *sql.Rows) ([]Item, error) {
	var out []Item
	for rows.Next() {
		var it Item
		var active int
		if err := rows.Scan(&it.ID, &it.SectionID, &it.RestaurantID, &it.Name, &it.Description, &it.PriceCents, &active, &it.SortOrder, &it.UpdatedAt); err != nil {
			return nil, err
		}
		it.Active = active != 0
		out = append(out, it)
	}
	return out, rows.Err()
}

func (s *Store) ItemByID(id int64) (*Item, error) {
	row := s.DB.QueryRow(`SELECT id, section_id, restaurant_id, name, description, price_cents, active, sort_order, updated_at
		FROM items WHERE id = ?`, id)
	var it Item
	var active int
	if err := row.Scan(&it.ID, &it.SectionID, &it.RestaurantID, &it.Name, &it.Description, &it.PriceCents, &active, &it.SortOrder, &it.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	it.Active = active != 0
	return &it, nil
}

func (s *Store) UpdateItem(id int64, name, description string, priceCents int64) error {
	_, err := s.DB.Exec(`UPDATE items SET name=?, description=?, price_cents=?, updated_at=datetime('now') WHERE id=?`,
		name, description, priceCents, id)
	return err
}

// ToggleItemActive flips the 86 state (available/unavailable) and returns the new state.
func (s *Store) ToggleItemActive(id int64) (bool, error) {
	_, err := s.DB.Exec(`UPDATE items SET active = NOT active, updated_at = datetime('now') WHERE id = ?`, id)
	if err != nil {
		return false, err
	}
	it, err := s.ItemByID(id)
	if err != nil {
		return false, err
	}
	return it.Active, nil
}

func (s *Store) DeleteItem(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM items WHERE id = ?`, id)
	return err
}
