package handler

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/crypto/bcrypt"

	"klowna/internal/model"
	"klowna/internal/render"
	"klowna/internal/store"
)

const (
	sessionCookieName = "klowna_session"
	sessionTTL        = 7 * 24 * time.Hour
)

type ctxKey string

const userKey ctxKey = "user"

func userFromContext(ctx context.Context) (model.User, bool) {
	u, ok := ctx.Value(userKey).(model.User)
	return u, ok
}

func (h *Handler) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || cookie.Value == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		user, err := h.store.UserBySession(r.Context(), hashToken(cookie.Value))
		if store.IsNotFound(err) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if err != nil {
			h.internalError(w, r, err)
			return
		}
		ctx := context.WithValue(r.Context(), userKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if _, ok := userFromContext(r.Context()); ok {
		http.Redirect(w, r, "/app", http.StatusSeeOther)
		return
	}
	site, err := h.store.Settings(r.Context())
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	d := render.PageData{
		Site:        site,
		Path:        "/login",
		Year:        time.Now().Year(),
		Title:       "Acceso | " + site.SiteName,
		Description: "Accede al panel de gestión de tu clínica.",
		FormError:   r.URL.Query().Get("error"),
	}
	h.renderAuth(w, "login", d)
}

func (h *Handler) LoginPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.internalError(w, r, err)
		return
	}
	email := trim(r.FormValue("email"), 160)
	password := r.FormValue("password")

	user, err := h.store.UserByEmail(r.Context(), email)
	if store.IsNotFound(err) || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		http.Redirect(w, r, "/login?error="+url.QueryEscape("Credenciales incorrectas."), http.StatusSeeOther)
		return
	}
	if err != nil && !store.IsNotFound(err) {
		h.internalError(w, r, err)
		return
	}

	token, err := newToken()
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	if err := h.store.CreateSession(r.Context(), hashToken(token), user.ID, time.Now().UTC().Add(sessionTTL)); err != nil {
		h.internalError(w, r, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(sessionTTL.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/app", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookieName); err == nil && cookie.Value != "" {
		_ = h.store.DeleteSession(r.Context(), hashToken(cookie.Value))
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (h *Handler) renderAuth(w http.ResponseWriter, page string, d render.PageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.render.RenderAuth(w, page, d); err != nil {
		slog.Error("render auth page", "page", page, "err", err)
	}
}
