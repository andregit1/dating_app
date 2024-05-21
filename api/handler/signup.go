package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"

	_ "dating_app/docs"

	"dating_app/pkg/payload"
	"dating_app/pkg/response"
	"dating_app/pkg/utils"

	_ "github.com/lib/pq"
)

// SignupHandler handles user registration
// @Summary Register a new user
// @Description Register a new user with the provided phone number.
// @Tags Users
// @Accept json
// @Produce json
// @Param data body payload.Entry true "Signup Object"
// @Success 201 {object} response.OTP "OTP generated successfully"
// @Failure 400 {string} string "Invalid request format"
// @Failure 500 {string} string "Internal server error"
// @Router /signup [post]
func Signup(db *sql.DB, w http.ResponseWriter, r *http.Request) {
  var payload payload.Entry

  if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
      http.Error(w, err.Error(), http.StatusBadRequest)
      return
  }

  phoneNumber := payload.Data.PhoneNumber

  var userID int // Assuming your user ID column is of type SERIAL or BIGSERIAL

  // Inserting into the users table and returning the user ID
  err := db.QueryRow("INSERT INTO users (phone_number) VALUES ($1) RETURNING id", phoneNumber).Scan(&userID)
  if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
  }

  otp := utils.GenerateOTP()
  err = utils.SaveOTP(db, userID, otp)
  if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
  }

  // Here, you would send the OTP to the user's phone number via an SMS service

  w.WriteHeader(http.StatusCreated)
  json.NewEncoder(w).Encode(response.OTP{OTP: otp})
}
