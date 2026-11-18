package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// secretKey signs session cookies. In production this should come from an
// environment variable; for the MVP we generate one at startup (see main.go)
// and keep it in memory, which is fine for a single-process deployment and
// simply logs everyone out on restart.
var secretKey []byte

func SetSecret(key []byte) { secretKey = key }

type sessionClaims struct {
	UserID int64 `json:"uid"`
	Exp    int64 `json:"exp"`
}

const cookieName = "menuapp_session"
const sessionTTL = 30 * 24 * time.Hour

func sign(payload []byte) string {
	mac := hmac.New(sha256.New, secretKey)
	mac.Write(payload)
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// NewSessionCookie returns a signed, HttpOnly cookie for the given user.
func NewSessionCookie(userID int64) *http.Cookie {
	claims := sessionClaims{UserID: userID, Exp: time.Now().Add(sessionTTL).Unix()}
	payload, _ := json.Marshal(claims)
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	sig := sign(payload)
	return &http.Cookie{
		Name:     cookieName,
		Value:    encoded + "." + sig,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(sessionTTL),
	}
}

func ExpiredCookie() *http.Cookie {
	return &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	}
}

var errInvalidSession = errors.New("invalid session")

// UserIDFromRequest validates the session cookie and returns the user id.
func UserIDFromRequest(r *http.Request) (int64, error) {
	c, err := r.Cookie(cookieName)
	if err != nil {
		return 0, errInvalidSession
	}
	dot := -1
	for i := len(c.Value) - 1; i >= 0; i-- {
		if c.Value[i] == '.' {
			dot = i
			break
		}
	}
	if dot < 0 {
		return 0, errInvalidSession
	}
	encoded, sig := c.Value[:dot], c.Value[dot+1:]
	payload, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return 0, errInvalidSession
	}
	if !hmac.Equal([]byte(sign(payload)), []byte(sig)) {
		return 0, errInvalidSession
	}
	var claims sessionClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return 0, errInvalidSession
	}
	if time.Now().Unix() > claims.Exp {
		return 0, errInvalidSession
	}
	return claims.UserID, nil
}

// --- Active restaurant selection (not security sensitive on its own; every
// handler re-checks that the requesting user actually owns the restaurant id
// before using it) ---

const activeRestaurantCookie = "menuapp_active_restaurant"

func NewActiveRestaurantCookie(restaurantID int64) *http.Cookie {
	return &http.Cookie{
		Name:     activeRestaurantCookie,
		Value:    itoa(restaurantID),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(sessionTTL),
	}
}

func ActiveRestaurantID(r *http.Request) (int64, bool) {
	c, err := r.Cookie(activeRestaurantCookie)
	if err != nil || c.Value == "" {
		return 0, false
	}
	id, ok := atoi(c.Value)
	return id, ok
}

func itoa(n int64) string {
	payload, _ := json.Marshal(n)
	return string(payload)
}

func atoi(s string) (int64, bool) {
	var n int64
	if err := json.Unmarshal([]byte(s), &n); err != nil {
		return 0, false
	}
	return n, true
}
