package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"dating_app/api/middleware"
	"dating_app/pkg/model"

	_ "github.com/lib/pq"
)

// @Summary Purchase premium
// @Description Purchase premium membership.
// @Accept json
// @Produce json
// @Param purchase body model.Purchase true "Purchase object"
// @Success 201 {string} string "Purchase successful"
// @Failure 400 {string} string "Invalid request format"
// @Failure 500 {string} string "Internal server error"
// @Router /purchase [post]
func Purchase(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var purchase model.Purchase
		
		if err := json.NewDecoder(r.Body).Decode(&purchase); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Get the current user ID from the context
		userID := middleware.CurrentUserID(r)
		purchase.UserID = userID

		// Ensure the purchase_date and created_at fields are set properly
		purchase.PurchaseDate = time.Now()
		purchase.CreatedAt = time.Now()

		// Perform the premium purchase action
		_, err := db.Exec("INSERT INTO purchases (user_id, package_id, purchase_date, created_at) VALUES ($1, $2, $3, $4)", purchase.UserID, purchase.PackageID, purchase.PurchaseDate, purchase.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}
