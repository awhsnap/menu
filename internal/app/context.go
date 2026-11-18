package app

import (
	"errors"
	"net/http"
	"strconv"

	"menuapp/internal/auth"
	"menuapp/internal/store"
)

var errUnauthorized = errors.New("unauthorized")
var errForbidden = errors.New("forbidden")

// currentUser loads the signed-in user from the session cookie, or returns
// errUnauthorized if there isn't one.
func (a *App) currentUser(r *http.Request) (*store.User, error) {
	uid, err := auth.UserIDFromRequest(r)
	if err != nil {
		return nil, errUnauthorized
	}
	u, err := a.Store.UserByID(uid)
	if err != nil {
		return nil, errUnauthorized
	}
	return u, nil
}

// requireUser redirects to /login and returns ok=false if there's no session.
func (a *App) requireUser(w http.ResponseWriter, r *http.Request) (*store.User, bool) {
	u, err := a.currentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return nil, false
	}
	return u, true
}

// requireRestaurant loads the user and the {rid} path param's restaurant,
// and verifies the user owns it. Writes an appropriate error response and
// returns ok=false if anything fails.
func (a *App) requireRestaurant(w http.ResponseWriter, r *http.Request) (*store.User, *store.Restaurant, bool) {
	u, ok := a.requireUser(w, r)
	if !ok {
		return nil, nil, false
	}
	ridStr := r.PathValue("rid")
	rid, err := strconv.ParseInt(ridStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return nil, nil, false
	}
	rest, err := a.Store.RestaurantByID(rid)
	if err != nil {
		http.NotFound(w, r)
		return nil, nil, false
	}
	if rest.OwnerUserID != u.ID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return nil, nil, false
	}
	return u, rest, true
}

func pathID(r *http.Request, name string) (int64, bool) {
	v := r.PathValue(name)
	id, err := strconv.ParseInt(v, 10, 64)
	return id, err == nil
}
