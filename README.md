# Menu Platform — MVP

A single-binary restaurant menu management dashboard: register, manage
multiple restaurants, build menus, cost out recipes, import sales, and see
demand forecasts + menu-engineering analytics (stars / plowhorses / puzzles
/ dogs).

## Stack (and one deviation you should know about)

- **Go** server, **SQLite** storage (`github.com/mattn/go-sqlite3`, cgo-based).
- Server-rendered HTML (`html/template`) with plain forms — no JS framework,
  no build step. A couple of charts (forecast bars, the menu-engineering
  scatter) are generated as inline SVG by the server, so there's no chart
  library dependency either.
- Auth is email + password (bcrypt) with a signed, HttpOnly session cookie
  (HMAC-SHA256, no sessions table, no third-party session library).

**Deviation from the brief:** this was built with Go's standard library
`net/http` router (Go 1.22+'s pattern-matching `ServeMux`) instead of **Gin**.
The sandbox this was built in only allows outbound Go module fetches from
`github.com` directly — the Go module proxy and vanity-import domains
(`golang.org/x/...`, `modernc.org/...`) were blocked — and Gin's default JSON
encoder (`bytedance/sonic`) pulls in dependencies hosted on blocked domains
(`golang.org/x/arch`, `nullprogram.com/...`) with no clean workaround. Rather
than ship something untested, I built the MVP on `net/http` instead, which
gives equivalent routing/middleware for this app's needs. If you want Gin
specifically, it's a mechanical swap — the business logic lives in
`internal/store`, `internal/forecast`, and `internal/analytics`, completely
decoupled from the router; only `internal/app/*` (route registration +
handler signatures) would need touching, and `go get github.com/gin-gonic/gin`
should resolve fine on a normal, non-sandboxed network.

## Run it

```bash
go run .
# or
go build -o menuapp . && ./menuapp
```

Then open http://localhost:8080, register, and create your first restaurant
(a starter menu + section are seeded automatically).

Config via environment variables (both optional):

- `MENUAPP_DB` — SQLite file path (default `data/menuapp.db`)
- `MENUAPP_ADDR` — listen address (default `:8080`)

The database file and a generated session-signing key are created under
`data/` on first run. Delete that directory to reset everything (including
logging everyone out).

## What's implemented

- **Accounts**: register/login/logout, bcrypt password hashing, signed
  session cookie (30-day expiry).
- **Multi-restaurant**: one account owns many restaurants; a switcher in the
  nav bar sets the "active" restaurant (also stored in a cookie for `/`
  redirects). Every restaurant-scoped route re-checks ownership server-side.
- **Menu editor**: menus → sections → items, inline add forms, 86
  (available/unavailable) toggle, edit, delete.
- **Ingredient library & recipe costing**: ingredients with purchase unit,
  cost per unit (fractional-cent precision — important, since a lot of
  ingredient costs are sub-cent per gram/ml), supplier, and a yield/trim-loss
  %. Recipes attach ingredients + quantities to an item; plate cost and
  margin are computed live, with price-change history logged per ingredient.
- **Margin alerts**: the dashboard flags any item whose plate cost exceeds
  the restaurant's target food-cost % (set in Settings).
- **Sales import**: CSV upload (`item,quantity,date` columns, flexible
  header names/order) matched to menu items by name (case-insensitive).
  Unmatched rows are still stored, not dropped.
- **Demand forecasting**: per item, a 7-day-trailing moving average blended
  with a linear trend fit over the daily sales series (gaps zero-filled),
  rendered as a bar chart (history vs. projected). Pure Go, no ML dependency.
- **Menu engineering**: the classic 2×2 popularity/profitability matrix
  (Star / Plowhorse / Puzzle / Dog) across all items with sales data, plus a
  scatter-plot visualization.
- Unit tests for the forecast engine and the menu-engineering classifier
  (`go test ./...`).

## What's intentionally out of scope for this MVP

Matches the PRD's non-goals / later-priority items: Google OAuth (this MVP
uses plain email+password instead, per your choice), staff roles/invites,
multi-location switching, time-based/scheduled menus, QR codes & public menu
pages, PDF export, Excel import/export, POS integrations, and the v2
AI-assisted features (description generation, pruning suggestions, photo
import). The data model (separate `sales_lines`, `ingredient_price_history`,
etc.) is intentionally close to the PRD's so those are straightforward to
layer on later.

## Project layout

```
main.go                     entrypoint: wiring, embedded templates/static
internal/db/                SQLite connection + embedded schema.sql
internal/store/             hand-written SQL queries, one file per entity
internal/auth/               password hashing, signed-cookie sessions
internal/forecast/          moving-average + trend forecasting (pure logic)
internal/analytics/         menu-engineering classifier (pure logic)
internal/app/                HTTP routes, handlers, template rendering, SVG charts
web/templates/               html/template pages (one layout + one file per page)
web/static/                  style.css
```

## A couple of known simplifications (fine for an MVP, worth knowing)

- Recipe quantities are entered in the ingredient's own purchase unit (no
  g↔oz style unit conversion yet) — keeps costing correct without building
  out a conversion table first.
- Sales-to-item matching is exact name match (case-insensitive). A CSV with
  slightly different item names won't auto-link; the PRD's sales-import
  flow for v1.x would add a mapping/preview step.
- SQLite is opened with a single connection (`SetMaxOpenConns(1)`) — correct
  and simple for an MVP with one writer; fine well beyond pilot scale for a
  single-tenant-per-restaurant app, per the PRD's own single-node trade-off
  notes.
