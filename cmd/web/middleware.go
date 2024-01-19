package main

import (
	"nastenka_udalosti/internal/helpers"
	"net/http"

	"github.com/justinas/nosurf"
)

// NoSurt přídá CSRF ochranu na každý POST request
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

// SessionLoad nahraje a uloží relaci na každý request
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

// Auth ověří, že je uživatel přihlášen
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, _ := helpers.IsAuthenticated(r)
		if !id {
			session.Put(r.Context(), "error", "Nejprve se musíte přihlásit!")
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Verified ověří, jestli uživatel je ověřen
func Verified(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ver, _ := helpers.IsVerified(r)
		if !ver {
			session.Put(r.Context(), "error", "Tvůj účet není ještě schválen!")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Verified ověří, jestli uživatel je má dostatečná práva Admin (access_level = 3)
func Admin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		adminAccessLevel, _ := helpers.GetAdminAccessLevel(r)
		if !adminAccessLevel {
			session.Put(r.Context(), "error", "Nemáte dostatečná práva")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func IsAuthor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		IsAuthor := false // TODO: Tady dodělej helper
		if !IsAuthor {
			http.Redirect(w, r, "/", http.StatusSeeOther) //TODO dodělej přesměrování funkce z helperu bude vracet src a bool -> přesměruje /dashboard/{src}/posts/my-events
			return
		}
		next.ServeHTTP(w, r)
	})
}
