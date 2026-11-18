package store

import (
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("not found")
var ErrDuplicate = errors.New("already exists")

func (s *Store) CreateUser(email, passwordHash, name string) (int64, error) {
	res, err := s.DB.Exec(`INSERT INTO users (email, password_hash, name) VALUES (?, ?, ?)`,
		email, passwordHash, name)
	if err != nil {
		if isUniqueErr(err) {
			return 0, ErrDuplicate
		}
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) UserByEmail(email string) (*User, error) {
	row := s.DB.QueryRow(`SELECT id, email, password_hash, name, created_at FROM users WHERE email = ?`, email)
	var u User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (s *Store) UserByID(id int64) (*User, error) {
	row := s.DB.QueryRow(`SELECT id, email, password_hash, name, created_at FROM users WHERE id = ?`, id)
	var u User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func isUniqueErr(err error) bool {
	// mattn/go-sqlite3 returns a *sqlite3.Error; comparing strings keeps this
	// file dependency-free of the driver's error type.
	return err != nil && (contains(err.Error(), "UNIQUE constraint failed") || contains(err.Error(), "constraint failed"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	n, m := len(s), len(substr)
	if m == 0 {
		return 0
	}
	for i := 0; i+m <= n; i++ {
		if s[i:i+m] == substr {
			return i
		}
	}
	return -1
}
