// Package store holds all hand-written SQL queries for the menu platform.
// No ORM: the schema is small enough that plain database/sql is clearer and
// has zero extra dependencies.
package store

import "database/sql"

type Store struct {
	DB *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{DB: db}
}
