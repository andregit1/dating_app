package utils

import (
	"database/sql"
	"fmt"
	"math/rand"

	_ "dating_app/docs"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// Generate a 6-digit OTP
func GenerateOTP() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

// Hash the OTP
func HashOTP(otp string) (string, error) {
	hashedOTP, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedOTP), nil
}

// Save the hashed OTP in the database
func SaveOTP(db *sql.DB, userID int, otp string) error {
	otpHash, err := HashOTP(otp)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO otp_auth (user_id, otp_hash) VALUES ($1, $2)", userID, otpHash)
	return err
}
