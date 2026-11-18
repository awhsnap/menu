// Command menuapp runs the Restaurant Menu Platform MVP: a single Go binary
// serving a server-rendered dashboard backed by SQLite. See README.md.
package main

import (
	"crypto/rand"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"menuapp/internal/app"
	"menuapp/internal/auth"
	"menuapp/internal/db"
	"menuapp/internal/store"
)

//go:embed web/templates/*.html
var templatesFS embed.FS

//go:embed web/static/*
var staticFS embed.FS

func main() {
	dbPath := envOr("MENUAPP_DB", "data/menuapp.db")
	addr := envOr("MENUAPP_ADDR", ":8080")

	if err := os.MkdirAll(dirOf(dbPath), 0o755); err != nil {
		log.Fatalf("create db dir: %v", err)
	}

	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer database.Close()

	secret, err := sessionSecret()
	if err != nil {
		log.Fatalf("session secret: %v", err)
	}
	auth.SetSecret(secret)

	s := store.New(database)

	webTemplates, err := fs.Sub(templatesFS, "web")
	if err != nil {
		log.Fatalf("templates fs: %v", err)
	}
	webStatic, err := fs.Sub(staticFS, "web/static")
	if err != nil {
		log.Fatalf("static fs: %v", err)
	}

	a, err := app.New(s, webTemplates)
	if err != nil {
		log.Fatalf("init app: %v", err)
	}

	log.Printf("menuapp listening on %s (db: %s)", addr, dbPath)
	if err := http.ListenAndServe(addr, a.Routes(webStatic)); err != nil {
		log.Fatal(err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func dirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[:i]
		}
	}
	return "."
}

// sessionSecret loads a persisted HMAC key from data/session.key, generating
// one on first run so cookies survive restarts (but not a fresh ./data wipe).
func sessionSecret() ([]byte, error) {
	path := envOr("MENUAPP_DB", "data/menuapp.db")
	keyPath := dirOf(path) + "/session.key"
	if b, err := os.ReadFile(keyPath); err == nil && len(b) == 32 {
		return b, nil
	}
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	if err := os.WriteFile(keyPath, key, 0o600); err != nil {
		return nil, err
	}
	return key, nil
}
