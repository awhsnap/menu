package app

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"

	"menuapp/internal/store"
)

type App struct {
	Store     *store.Store
	templates map[string]*template.Template
}

func New(s *store.Store, webFS fs.FS) (*App, error) {
	a := &App{Store: s}
	if err := a.loadTemplates(webFS); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *App) Routes(staticFS fs.FS) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(staticFS)))

	// Public
	mux.HandleFunc("GET /{$}", a.handleRoot)
	mux.HandleFunc("GET /register", a.handleRegisterForm)
	mux.HandleFunc("POST /register", a.handleRegister)
	mux.HandleFunc("GET /login", a.handleLoginForm)
	mux.HandleFunc("POST /login", a.handleLogin)
	mux.HandleFunc("POST /logout", a.handleLogout)

	// Restaurant list / create / switch (user-scoped, not restaurant-scoped)
	mux.HandleFunc("GET /restaurants", a.handleRestaurantsList)
	mux.HandleFunc("POST /restaurants", a.handleRestaurantCreate)
	mux.HandleFunc("POST /restaurants/{rid}/select", a.handleRestaurantSelect)
	mux.HandleFunc("GET /restaurants/{rid}/select", a.handleRestaurantSelect)

	// Restaurant-scoped pages
	mux.HandleFunc("GET /r/{rid}/dashboard", a.handleDashboard)
	mux.HandleFunc("GET /r/{rid}/settings", a.handleSettings)
	mux.HandleFunc("POST /r/{rid}/settings", a.handleSettingsSave)

	mux.HandleFunc("GET /r/{rid}/menus", a.handleMenus)
	mux.HandleFunc("POST /r/{rid}/menus", a.handleMenuCreate)
	mux.HandleFunc("GET /r/{rid}/menus/{mid}", a.handleMenuDetail)
	mux.HandleFunc("POST /r/{rid}/menus/{mid}/sections", a.handleSectionCreate)
	mux.HandleFunc("POST /r/{rid}/sections/{sid}/items", a.handleItemCreate)

	mux.HandleFunc("GET /r/{rid}/items/{id}", a.handleItemDetail)
	mux.HandleFunc("POST /r/{rid}/items/{id}/update", a.handleItemUpdate)
	mux.HandleFunc("POST /r/{rid}/items/{id}/toggle", a.handleItemToggle)
	mux.HandleFunc("POST /r/{rid}/items/{id}/delete", a.handleItemDelete)
	mux.HandleFunc("POST /r/{rid}/items/{id}/recipe", a.handleRecipeLineAdd)
	mux.HandleFunc("POST /r/{rid}/recipe-lines/{lid}/delete", a.handleRecipeLineDelete)
	mux.HandleFunc("GET /r/{rid}/items/{id}/forecast", a.handleItemForecast)

	mux.HandleFunc("GET /r/{rid}/ingredients", a.handleIngredients)
	mux.HandleFunc("POST /r/{rid}/ingredients", a.handleIngredientCreate)
	mux.HandleFunc("POST /r/{rid}/ingredients/{id}/update", a.handleIngredientUpdate)
	mux.HandleFunc("POST /r/{rid}/ingredients/{id}/delete", a.handleIngredientDelete)

	mux.HandleFunc("GET /r/{rid}/sales", a.handleSales)
	mux.HandleFunc("POST /r/{rid}/sales/import", a.handleSalesImport)

	mux.HandleFunc("GET /r/{rid}/analytics", a.handleAnalytics)

	return logging(mux)
}

func logging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		h.ServeHTTP(w, r)
	})
}
