package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/sessions"

	_ "dating_app/docs"

	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
	"golang.org/x/crypto/bcrypt"
)

var (
	db    *sql.DB
	store = sessions.NewCookieStore([]byte("super-secret-key"))
)

type User struct {
	ID          int    `json:"id"`
	PhoneNumber string `json:"phone_number"`
	IsPremium   bool   `json:"is_premium"`
	Verified    bool   `json:"verified"`
	IsDeleted   bool   `json:"is_deleted"`
	SignupAt    string `json:"signup_at"`
	LoginAt     string `json:"login_at"`
	LogoutAt    string `json:"logout_at"`
}

type OTPResponse struct {
	OTP string `json:"otp"`
}

type Profile struct {
	ID       int    `json:"id"`
	UserID   int    `json:"-"`
	Name     string `json:"name"`
	Age      int    `json:"age"`
	Bio      string `json:"bio"`
	PhotoURL string `json:"photo_url"`
}

type Swipe struct {
	ID        int       `json:"id"`
	SwiperID  int       `json:"-"`
	ProfileID int       `json:"-"`
	SwipeType string    `json:"swipe_type"`
	SwipeDate time.Time `json:"swipe_date"`
}

type Purchase struct {
	ID           int       `json:"id"`
	UserID       int       `json:"-"`
	PurchaseDate time.Time `json:"purchase_date"`
}

type Preference struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	DateMode       bool      `json:"date_mode"`
	BFFMode        bool      `json:"bff_mode"`
	PreferredGender string   `json:"preferred_gender"`
	MinAge         int       `json:"min_age"`
	MaxAge         int       `json:"max_age"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Payload struct {
	Data struct {
		PhoneNumber string `json:"phone_number"`
	} `json:"data"`
}

// type OTPPayload struct {
// 	Data struct {
// 		OTP string `json:"otp"`
// 	} `json:"data"`
// }

// Generate a 6-digit OTP
func generateOTP() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

// Hash the OTP
func hashOTP(otp string) (string, error) {
	hashedOTP, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedOTP), nil
}

// Save the hashed OTP in the database
func saveOTP(userID int, otp string) error {
	otpHash, err := hashOTP(otp)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO otp_auth (user_id, otp_hash) VALUES ($1, $2)", userID, otpHash)
	return err
}

// @title Dating App API
// @description This is a sample dating app API.
// @version 1.0
// @host localhost:8080
// @BasePath /
func main() {
	var err error
	db, err = sql.Open("postgres", "user=root password=123123123 dbname=dating_app sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Define HTTP routes and their corresponding handlers
	http.HandleFunc("/signup", signupHandler)     // @router /signup [post]
	http.HandleFunc("/login", loginHandler)       // @router /login [post]
	http.HandleFunc("/verify-otp", verifyOTPHandler) // @router /verify-otp [post]
	http.HandleFunc("/swipe", swipeHandler)       // @router /swipe [post]
	http.HandleFunc("/purchase", purchaseHandler) // @router /purchase [post]

	// Protected route for getting users
	http.Handle("/users", authMiddleware(http.HandlerFunc(getUsersHandler))) // @router /users [get]

	// Serve Swagger UI
	http.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // The url pointing to API definition
	))

	// Start the HTTP server
	serverAddr := "localhost:8080"
	go func() {
		log.Printf("Server is starting and listening on %s", serverAddr)
		if err := http.ListenAndServe(serverAddr, nil); err != nil {
			log.Fatalf("Error starting server: %s", err)
		}
	}()

	// Wait for server shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	log.Println("Server stopped gracefully")
}

// SignupHandler handles user registration
// @Summary Register a new user
// @Description Register a new user with the provided phone number.
// @Accept json
// @Produce json
// @Param data body Payload true "Signup Object"
// @Success 201 {object} OTPResponse "OTP generated successfully"
// @Failure 400 {string} string "Invalid request format"
// @Failure 500 {string} string "Internal server error"
// @Router /signup [post]
func signupHandler(w http.ResponseWriter, r *http.Request) {
	var payload Payload

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	phoneNumber := payload.Data.PhoneNumber

	result, err := db.Exec("INSERT INTO users (phone_number) VALUES ($1)", phoneNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	otp := generateOTP()
	err = saveOTP(int(userID), otp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Here, you would send the OTP to the user's phone number via an SMS service

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(OTPResponse{OTP: otp})
}

// LoginHandler handles user login
// @Summary Login
// @Description Login with the provided phone number and get OTP.
// @Accept json
// @Produce json
// @Param data body Payload true "Login Object"
// @Success 200 {object} OTPResponse "OTP generated successfully"
// @Failure 400 {string} string "Invalid request format"
// @Failure 401 {string} string "Invalid phone number"
// @Router /login [post]
func loginHandler(w http.ResponseWriter, r *http.Request) {
	var payload Payload

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var dbUser User
	phoneNumber := payload.Data.PhoneNumber
	err := db.QueryRow("SELECT id FROM users WHERE phone_number = $1", phoneNumber).Scan(&dbUser.ID)
	if err != nil {
		http.Error(w, "invalid phone number", http.StatusUnauthorized)
		return
	}

	otp := generateOTP()
	err = saveOTP(dbUser.ID, otp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Here, you would send the OTP to the user's phone number via an SMS service

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(OTPResponse{OTP: otp})
}

// VerifyOTPHandler handles OTP verification
// @Summary Verify OTP
// @Description Verify the OTP entered by the user and create a session.
// @Accept json
// @Produce json
// @Param data body map[string]string true "OTP verification payload"
// @Success 200 {string} string "OTP verified successfully"
// @Failure 400 {string} string "Invalid OTP"
// @Failure 500 {string} string "Internal server error"
// @Router /verify-otp [post]
func verifyOTPHandler(w http.ResponseWriter, r *http.Request) {
	var payload map[string]string

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	phoneNumber := payload["phone_number"]
	otp := payload["otp"]

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

// swipeHandler handles swiping left or right
// @Summary Swipe
// @Description Swipe left or right on a profile.
// @Accept json
// @Produce json
// @Param swipe body Swipe true "Swipe object"
// @Success 201 {string} string "Swipe recorded successfully"
// @Failure 400 {string} string "Invalid request format"
// @Failure 500 {string} string "Internal server error"
// @Router /swipe [post]
func swipeHandler(w http.ResponseWriter, r *http.Request) {
	var swipe Swipe

	if err := json.NewDecoder(r.Body).Decode(&swipe); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if user has exceeded the daily swipe limit
	if err := checkDailySwipeLimit(swipe.SwiperID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if user has already swiped this profile today
	if err := checkDuplicateSwipe(swipe.SwiperID, swipe.ProfileID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("INSERT INTO swipes (swiper_id, profile_id, swipe_type, swipe_date) VALUES ($1, $2, $3, $4)", swipe.SwiperID, swipe.ProfileID, swipe.SwipeType, time.Now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// checkDailySwipeLimit checks if the user has exceeded the daily swipe limit
func checkDailySwipeLimit(userID int) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM swipes WHERE swiper_id = $1 AND swipe_date >= current_date", userID).Scan(&count)
	if err != nil {
		return err
	}

	if count >= 10 {
		return errors.New("daily swipe limit exceeded")
	}

	return nil
}

// checkDuplicateSwipe checks if the user has already swiped the profile on the same day
func checkDuplicateSwipe(userID, profileID int) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM swipes WHERE swiper_id = $1 AND profile_id = $2 AND swipe_date >= current_date", userID, profileID).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("profile already swiped by the user today")
	}

	return nil
}

// purchaseHandler handles premium membership purchase
// @Summary Purchase premium
// @Description Purchase premium membership.
// @Accept json
// @Produce json
// @Param purchase body Purchase true "Purchase object"
// @Success 201 {string} string "Purchase successful"
// @Failure 400 {string} string "Invalid request format"
// @Failure 500 {string} string "Internal server error"
// @Router /purchase [post]
func purchaseHandler(w http.ResponseWriter, r *http.Request) {
	var purchase Purchase

	if err := json.NewDecoder(r.Body).Decode(&purchase); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Perform the premium purchase action
	_, err := db.Exec("INSERT INTO purchases (user_id, purchase_date) VALUES ($1, $2)", purchase.UserID, time.Now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// getUsersHandler handles retrieving users based on preferences
// @Summary Get a list of users based on preferences
// @Description Get a list of users based on the logged-in user's preferences.
// @Accept json
// @Produce json
// @Success 200 {array} User "List of users matching preferences"
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Internal server error"
// @Router /users [get]
func getUsersHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")

	// Check if user is authenticated
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	preferences, err := getUserPreferences(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	users, err := getUsersBasedOnPreferences(preferences)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

// getUserPreferences retrieves the preferences of the logged-in user
func getUserPreferences(userID int) (Preference, error) {
	var preferences Preference

	err := db.QueryRow("SELECT id, user_id, date_mode, bff_mode, preferred_gender, min_age, max_age, created_at, updated_at FROM preferences WHERE user_id = $1", userID).Scan(
		&preferences.ID, &preferences.UserID, &preferences.DateMode, &preferences.BFFMode, &preferences.PreferredGender, &preferences.MinAge, &preferences.MaxAge, &preferences.CreatedAt, &preferences.UpdatedAt)
	if err != nil {
		return preferences, err
	}
	return preferences, nil
}

// getUsersBasedOnPreferences retrieves a list of users based on the preferences
func getUsersBasedOnPreferences(preferences Preference) ([]User, error) {
	query := "SELECT id, phone_number, is_premium, verified, is_deleted, signup_at, login_at, logout_at FROM users WHERE is_deleted = FALSE AND id != $1"
	args := []interface{}{preferences.UserID}

	if preferences.PreferredGender != "" && preferences.PreferredGender != "both" {
		query += " AND gender = $2"
		args = append(args, preferences.PreferredGender)
	}

	if preferences.MinAge > 0 {
		query += " AND age >= $3"
		args = append(args, preferences.MinAge)
	}

	if preferences.MaxAge > 0 {
		query += " AND age <= $4"
		args = append(args, preferences.MaxAge)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User

		if err := rows.Scan(&user.ID, &user.PhoneNumber, &user.IsPremium, &user.Verified, &user.IsDeleted, &user.SignupAt, &user.LoginAt, &user.LogoutAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// authMiddleware is a middleware to check if the user is authenticated
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session-name")

		// Check if user is authenticated
		if _, ok := session.Values["user_id"].(int); !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
