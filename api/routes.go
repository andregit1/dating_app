package api

import (
	"database/sql"
	"net/http"

	"dating_app/api/handler"
	"dating_app/api/middleware"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	httpSwagger "github.com/swaggo/http-swagger"
)

var store = sessions.NewCookieStore([]byte("super-secret-key"))

func isUserLoggedIn(w http.ResponseWriter, r *http.Request) bool {
	session, _ := store.Get(r, "session-name")

	// Check if user is authenticated
	_, ok := session.Values["user_id"].(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}
	return true
}

func Routes(db *sql.DB) {
	// wrapHandler is a helper function that wraps an HTTP handler with a database connection and optional authentication middleware
	wrapHandler := func(handlerFunc func(db *sql.DB, w http.ResponseWriter, r *http.Request), authRequired bool) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Authentication check if required
			if authRequired {
				if !isUserLoggedIn(w, r) {
					return
				}
			}

			handlerFunc(db, w, r)
		}
	}

	http.HandleFunc("/signup", wrapHandler(handler.Signup, false))        // @router /signup [post]
	http.HandleFunc("/login", wrapHandler(handler.Login, false))          // @router /login [post]
	http.HandleFunc("/verify-otp", wrapHandler(handler.VerifyOTP, false)) // @router /verify-otp [post]
	http.HandleFunc("/swipe", wrapHandler(handler.Swipe, true))       // @router /swipe [post]
	http.HandleFunc("/purchase", wrapHandler(handler.Purchase, true)) // @router /purchase [post]
	http.HandleFunc("/cards", wrapHandler(handler.Card, true))      // @router /cards [get]
	
	// Define routes for package CRUD operations using Gorilla Mux
	router := mux.NewRouter()
	router.HandleFunc("/packages/create", wrapHandler(handler.CreatePackage, true)).Methods("POST")
	router.HandleFunc("/packages", wrapHandler(handler.GetPackage, true)).Methods("GET")
	router.HandleFunc("/packages/edit/{id}", wrapHandler(handler.UpdatePackage, true)).Methods("PUT")
	router.HandleFunc("/packages/delete/{id}", wrapHandler(handler.DeletePackage, true)).Methods("PATCH")

	// Serve Swagger UI
	http.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // The URL pointing to API definition
	))

	// Enable CORS for all routes
	http.Handle("/", middleware.EnableCORS(http.DefaultServeMux))
	http.Handle("/packages/create", middleware.EnableCORSMux(router))
	http.Handle("/packages", middleware.EnableCORSMux(router))
	http.Handle("/packages/edit/{id}", middleware.EnableCORSMux(router))
	http.Handle("/packages/delete/{id}", middleware.EnableCORSMux(router))
}
