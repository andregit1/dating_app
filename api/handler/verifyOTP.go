package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gorilla/sessions"

	_ "dating_app/docs"

	"dating_app/pkg/payload"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// @Summary Verify OTP
// @Description Verify the OTP entered by the user and create a session.
// @Tags Users
// @Accept json
// @Produce json
// @Param data body payload.OTP true "Verify OTP object"
// @Success 200 {string} string "OTP verified successfully"
// @Failure 400 {string} string "Invalid OTP"
// @Failure 500 {string} string "Internal server error"
// @Router /verify-otp [post]
func VerifyOTP(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var (
		store = sessions.NewCookieStore([]byte("super-secret-key"))
		payload payload.OTP
	)

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	phoneNumber := payload.Data.PhoneNumber
	otp := payload.Data.OTP

	var userID int
	var otpHash string
	err := db.QueryRow("SELECT id FROM users WHERE phone_number = $1", phoneNumber).Scan(&userID)
	if err != nil {
		http.Error(w, "Invalid phone number", http.StatusBadRequest)
		return
	}

	err = db.QueryRow("SELECT otp_hash FROM otp_auth WHERE user_id = $1", userID).Scan(&otpHash)
	if err != nil {
		http.Error(w, "OTP not found", http.StatusBadRequest)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(otpHash), []byte(otp))
	if err != nil {
		http.Error(w, "Invalid OTP", http.StatusBadRequest)
		return
	}

	// OTP verified, create session
	session, _ := store.Get(r, "session-name")
	session.Values["user_id"] = userID
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OTP verified successfully"))
}
