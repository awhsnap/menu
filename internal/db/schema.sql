-- Menu Platform MVP schema (SQLite)
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS users (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    name          TEXT NOT NULL DEFAULT '',
    created_at    TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS restaurants (
    id                   INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_user_id        INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name                 TEXT NOT NULL,
    slug                 TEXT NOT NULL UNIQUE,
    currency             TEXT NOT NULL DEFAULT 'USD',
    target_food_cost_pct REAL NOT NULL DEFAULT 30.0,
    created_at           TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_restaurants_owner ON restaurants(owner_user_id);

CREATE TABLE IF NOT EXISTS menus (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    restaurant_id INTEGER NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    name          TEXT NOT NULL,
    active        INTEGER NOT NULL DEFAULT 1,
    sort_order    INTEGER NOT NULL DEFAULT 0,
    created_at    TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_menus_restaurant ON menus(restaurant_id);

CREATE TABLE IF NOT EXISTS sections (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    menu_id    INTEGER NOT NULL REFERENCES menus(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_sections_menu ON sections(menu_id);

CREATE TABLE IF NOT EXISTS items (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    section_id  INTEGER NOT NULL REFERENCES sections(id) ON DELETE CASCADE,
    restaurant_id INTEGER NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    price_cents INTEGER NOT NULL DEFAULT 0,
    active      INTEGER NOT NULL DEFAULT 1,
    sort_order  INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_items_section ON items(section_id);
CREATE INDEX IF NOT EXISTS idx_items_restaurant ON items(restaurant_id);

CREATE TABLE IF NOT EXISTS ingredients (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    restaurant_id  INTEGER NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    name           TEXT NOT NULL,
    unit           TEXT NOT NULL DEFAULT 'unit', -- g, kg, ml, l, oz, lb, unit, tsp, tbsp, cup
    unit_cost_cents REAL NOT NULL DEFAULT 0,     -- cost per 1 `unit`, in cents (fractional: many
                                                  -- ingredients cost less than a whole cent per gram/ml)
    supplier       TEXT NOT NULL DEFAULT '',
    yield_pct      REAL NOT NULL DEFAULT 100.0,  -- usable yield after trim/loss
    created_at     TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at     TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_ingredients_restaurant ON ingredients(restaurant_id);

CREATE TABLE IF NOT EXISTS ingredient_price_history (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    ingredient_id  INTEGER NOT NULL REFERENCES ingredients(id) ON DELETE CASCADE,
    unit_cost_cents REAL NOT NULL,
    changed_at     TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_price_history_ingredient ON ingredient_price_history(ingredient_id);

CREATE TABLE IF NOT EXISTS recipe_lines (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    item_id       INTEGER NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    ingredient_id INTEGER NOT NULL REFERENCES ingredients(id) ON DELETE CASCADE,
    quantity      REAL NOT NULL DEFAULT 0 -- in the ingredient's own purchase unit
);
CREATE INDEX IF NOT EXISTS idx_recipe_lines_item ON recipe_lines(item_id);
CREATE INDEX IF NOT EXISTS idx_recipe_lines_ingredient ON recipe_lines(ingredient_id);

CREATE TABLE IF NOT EXISTS sales_imports (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    restaurant_id INTEGER NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    filename      TEXT NOT NULL DEFAULT '',
    row_count     INTEGER NOT NULL DEFAULT 0,
    matched_count INTEGER NOT NULL DEFAULT 0,
    imported_at   TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS sales_lines (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    restaurant_id INTEGER NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    import_id     INTEGER REFERENCES sales_imports(id) ON DELETE SET NULL,
    item_id       INTEGER REFERENCES items(id) ON DELETE SET NULL,
    item_name     TEXT NOT NULL,
    quantity      REAL NOT NULL DEFAULT 0,
    sale_date     TEXT NOT NULL -- YYYY-MM-DD
);
CREATE INDEX IF NOT EXISTS idx_sales_restaurant ON sales_lines(restaurant_id);
CREATE INDEX IF NOT EXISTS idx_sales_item ON sales_lines(item_id);
CREATE INDEX IF NOT EXISTS idx_sales_date ON sales_lines(sale_date);
