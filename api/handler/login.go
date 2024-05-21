package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"

	_ "dating_app/docs"

	"dating_app/pkg/model"
	"dating_app/pkg/payload"
	"dating_app/pkg/response"
	"dating_app/pkg/utils"

	_ "github.com/lib/pq"
)

// LoginHandler handles user login
// @Summary Login
// @Description Login with the provided phone number and get OTP.
// @Tags Users
// @Accept json
// @Produce json
// @Param data body payload.Entry true "Login Object"
// @Success 200 {object} response.OTP "OTP generated successfully"
// @Failure 400 {string} string "Invalid request format"
// @Failure 401 {string} string "Invalid phone number"
// @Router /login [post]
func Login(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var payload payload.Entry

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var user model.User
	phoneNumber := payload.Data.PhoneNumber
	err := db.QueryRow("SELECT id FROM users WHERE phone_number = $1", phoneNumber).Scan(&user.ID)
	if err != nil {
		http.Error(w, "invalid phone number", http.StatusUnauthorized)
		return
	}

	otp := utils.GenerateOTP()
	err = utils.SaveOTP(db, user.ID, otp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Here, you would send the OTP to the user's phone number via an SMS service

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response.OTP{OTP: otp})
}
