package main

import (
	"bookings/internal/helpers"
	"net/http"

	"github.com/justinas/nosurf"
)

// NoSurf adds CSRF protection to all POST requests
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})

	return csrfHandler
}

// Sessionload loads and saves the session on every request
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

// Auth redirects to the login page, if user is not authenticated needs for protected routes
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !helpers.IsAuthenticated(r) {
			session.Put(r.Context(), "error", "Log in first!")
			http.Redirect(w, r, "/users/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
