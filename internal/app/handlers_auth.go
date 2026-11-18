package app

import (
	"net/http"
	"strings"

	"menuapp/internal/auth"
	"menuapp/internal/store"
)

func (a *App) handleRoot(w http.ResponseWriter, r *http.Request) {
	u, err := a.currentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if rid, ok := auth.ActiveRestaurantID(r); ok {
		if rest, err := a.Store.RestaurantByID(rid); err == nil && rest.OwnerUserID == u.ID {
			http.Redirect(w, r, "/r/"+itoa(rest.ID)+"/dashboard", http.StatusSeeOther)
			return
		}
	}
	http.Redirect(w, r, "/restaurants", http.StatusSeeOther)
}

type authPageData struct {
	navData
	Error string
	Email string
}

func (a *App) handleRegisterForm(w http.ResponseWriter, r *http.Request) {
	a.render(w, "register.html", authPageData{})
}

func (a *App) handleRegister(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	name := strings.TrimSpace(r.FormValue("name"))
	password := r.FormValue("password")

	if email == "" || !strings.Contains(email, "@") {
		a.render(w, "register.html", authPageData{Error: "Enter a valid email address.", Email: email})
		return
	}
	if len(password) < 8 {
		a.render(w, "register.html", authPageData{Error: "Password must be at least 8 characters.", Email: email})
		return
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	id, err := a.Store.CreateUser(email, hash, name)
	if err != nil {
		if err == store.ErrDuplicate {
			a.render(w, "register.html", authPageData{Error: "An account with that email already exists.", Email: email})
			return
		}
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, auth.NewSessionCookie(id))
	http.Redirect(w, r, "/restaurants?flash=Welcome! Create your first restaurant to get started.", http.StatusSeeOther)
}

func (a *App) handleLoginForm(w http.ResponseWriter, r *http.Request) {
	a.render(w, "login.html", authPageData{})
}

func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	password := r.FormValue("password")

	u, err := a.Store.UserByEmail(email)
	if err != nil || !auth.CheckPassword(u.PasswordHash, password) {
		a.render(w, "login.html", authPageData{Error: "Incorrect email or password.", Email: email})
		return
	}
	http.SetCookie(w, auth.NewSessionCookie(u.ID))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, auth.ExpiredCookie())
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
