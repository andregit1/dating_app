package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"dating_app/pkg/model"
	"dating_app/pkg/payload"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

func isUserLoggedIn(w http.ResponseWriter, r *http.Request) bool {
	var store = sessions.NewCookieStore([]byte("super-secret-key"))
	session, _ := store.Get(r, "session-name")

	// Check if user is authenticated
	_, ok := session.Values["user_id"].(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}
	return true
}

// @Summary Create a new package
// @Description Create a new package.
// @Tags Packages
// @Accept json
// @Produce json
// @Param data body payload.Package true "Package object"
// @Success 201 {string} string "Package created successfully"
// @Failure 400 {string} string "Invalid request format"
// @Failure 500 {string} string "Internal server error"
// @Router /packages/create [post]
func CreatePackage(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var pkg payload.Package

	if !isUserLoggedIn(w, r) {
			return
	}

	if err := json.NewDecoder(r.Body).Decode(&pkg); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
	}

	createdAt := time.Now()

	_, err := db.Exec("INSERT INTO packages (name, feature, price, currency, created_at) VALUES ($1, $2, $3, $4, $5)", pkg.Data.Name, pkg.Data.Feature, pkg.Data.Price, pkg.Data.Currency, createdAt)
	if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
	}

	w.WriteHeader(http.StatusCreated)
}

// @Summary Get all packages
// @Description Retrieve all packages.
// @Tags Packages
// @Accept json
// @Produce json
// @Success 200 {array} model.Package "List of packages"
// @Failure 500 {string} string "Internal server error"
// @Router /packages [get]
func GetPackage(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if !isUserLoggedIn(w, r) {
			return
	}

	rows, err := db.Query("SELECT id, name, feature, price, currency, is_deleted, created_at, updated_at FROM packages WHERE is_deleted = false")
	if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
	}
	defer rows.Close()

	var packages []model.Package
	for rows.Next() {
			var pkg model.Package
			err := rows.Scan(&pkg.ID, &pkg.Name, &pkg.Feature, &pkg.Price, &pkg.Currency, &pkg.IsDeleted, &pkg.CreatedAt, &pkg.UpdatedAt)
			if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
			}
			packages = append(packages, pkg)
	}

	if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
	}

	json.NewEncoder(w).Encode(packages)
}

// @Summary Update a package
// @Description Update an existing package by ID.
// @Tags Packages
// @Accept json
// @Produce json
// @Param id path integer true "Package ID"
// @Param data body payload.Package true "Package object"
// @Success 200 {string} string "Package updated successfully"
// @Failure 400 {string} string "Invalid request format"
// @Failure 404 {string} string "Package not found"
// @Failure 500 {string} string "Internal server error"
// @Router /packages/edit/{id} [put]
func UpdatePackage(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if !isUserLoggedIn(w, r) {
			return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
			http.Error(w, "Package ID is required", http.StatusBadRequest)
			return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
			http.Error(w, "Invalid package ID", http.StatusBadRequest)
			return
	}

	var pkg payload.Package
	if err := json.NewDecoder(r.Body).Decode(&pkg); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
	}

	updatedAt := time.Now()

	result, err := db.Exec("UPDATE packages SET name = $1, feature = $2, price = $3, currency = $4, updated_at = $5 WHERE id = $6 AND is_deleted = false", pkg.Data.Name, pkg.Data.Feature, pkg.Data.Price, pkg.Data.Currency, updatedAt, id)
	if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
	}

	if rowsAffected == 0 {
			http.Error(w, "Package not found", http.StatusNotFound)
			return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary Soft delete a package
// @Description Soft delete a package by setting is_deleted field to true.
// @Tags Packages
// @Accept json
// @Produce json
// @Param id path integer true "Package ID"
// @Success 204 {string} string "Package deleted successfully"
// @Failure 404 {string} string "Package not found"
// @Failure 500 {string} string "Internal server error"
// @Router /packages/delete/{id} [patch]
func DeletePackage(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if !isUserLoggedIn(w, r) {
			return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
			http.Error(w, "Package ID is required", http.StatusBadRequest)
			return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
			http.Error(w, "Invalid package ID", http.StatusBadRequest)
			return
	}

	updatedAt := time.Now()

	result, err := db.Exec("UPDATE packages SET is_deleted = true, updated_at = $1 WHERE id = $2", updatedAt, id)
	if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
	}

	if rowsAffected == 0 {
			http.Error(w, "Package not found", http.StatusNotFound)
			return
	}

	w.WriteHeader(http.StatusNoContent)
}
