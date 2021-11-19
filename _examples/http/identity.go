package main

import (
	"context"
	"net/http"
)

type Identity interface {
	Middleware(http.Handler) http.Handler

	User(*http.Request) string
}

type BasicAuth struct{}

func (auth *BasicAuth) key() *BasicAuth {
	return auth
}

func (auth *BasicAuth) User(r *http.Request) string {
	user := r.Context().Value(auth.key())
	if user == nil {
		return ""
	}

	return user.(string)
}

func (auth *BasicAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if ok {
			if !checkPassword(user, password) {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), auth.key(), user))
		}

		next.ServeHTTP(w, r)
	})
}
