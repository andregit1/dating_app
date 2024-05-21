package middleware

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type contextKey string

const userIDKey contextKey = "userID"

var store = sessions.NewCookieStore([]byte("super-secret-key"))

// Authentication middleware to check if the user is authenticated
func Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session-name")

		// Check if user is authenticated
		userID, ok := session.Values["user_id"].(int)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Store userID in context
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// CurrentUserID retrieves the current user ID from the context
func CurrentUserID(r *http.Request) (int, error) {
	userID, ok := r.Context().Value(userIDKey).(int)
	if !ok {
		return 0, http.ErrNoCookie
	}
	return userID, nil
}

func SetupSession(store *sessions.CookieStore) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					session, _ := store.Get(r, "session-name")
					userID, ok := session.Values["user_id"].(int)
					if !ok || userID == 0 {
							http.Error(w, "Unauthorized", http.StatusUnauthorized)
							return
					}
					next.ServeHTTP(w, r)
			})
	}
}
