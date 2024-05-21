package model

import (
	"time"

	_ "dating_app/docs"

	_ "github.com/lib/pq"
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

type Card struct {
	UserID   int    `json:"user_id"`
	Verified bool   `json:"verified"`
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
	UserID       int       `json:"user_id"`
	PackageID    int       `json:"package_id"`
	PurchaseDate time.Time `json:"purchase_date"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Package struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Feature   string    `json:"feature"`
	Price     float64   `json:"price"`
	Currency  string    `json:"currency"`
	IsDeleted bool      `json:"is_deleted"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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

type OTPPayload struct {
	Data struct {
		OTP         string `json:"otp"`
		PhoneNumber string `json:"phone_number"`
	} `json:"data"`
}