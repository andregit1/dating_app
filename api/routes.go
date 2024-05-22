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

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isUserLoggedIn(w, r) {
			return
		}
		next.ServeHTTP(w, r)
	})
}

func Routes(db *sql.DB) {
	// Create a new router
	router := mux.NewRouter()

	// Public routes
	router.HandleFunc("/signup", handler.Signup(db)).Methods("POST")
	router.HandleFunc("/login", handler.Login(db)).Methods("POST")
	router.HandleFunc("/verify-otp", handler.VerifyOTP(db)).Methods("POST")

	// Create a subrouter for authenticated routes
	authenticatedRouter := router.NewRoute().Subrouter()
	authenticatedRouter.Use(authMiddleware)

	// Define authenticated routes
	authenticatedRouter.HandleFunc("/swipe", handler.Swipe(db)).Methods("POST")
	authenticatedRouter.HandleFunc("/purchase", handler.Purchase(db)).Methods("POST")
	authenticatedRouter.HandleFunc("/cards", handler.Card(db)).Methods("GET")

	// Create a subrouter for package-related routes that require authentication
	packagesRouter := router.PathPrefix("/packages").Subrouter()
	packagesRouter.Use(authMiddleware)

	// Define package-related routes using the packagesRouter
	packagesRouter.HandleFunc("/create", handler.CreatePackage(db)).Methods("POST")
	packagesRouter.HandleFunc("", handler.GetPackage(db)).Methods("GET")
	packagesRouter.HandleFunc("/edit/{id}", handler.UpdatePackage(db)).Methods("PUT")
	packagesRouter.HandleFunc("/delete/{id}", handler.DeletePackage(db)).Methods("PATCH")

	// Enable CORS for all routes
	corsRouter := middleware.EnableCORSMux(router)

	// Use the corsRouter for handling requests
	http.Handle("/", corsRouter)

	// Serve Swagger UI
	http.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // The URL pointing to API definition
	))
}
