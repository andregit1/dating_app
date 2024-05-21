package payload

import (
	_ "dating_app/docs"

	_ "github.com/lib/pq"
)

type Entry struct {
	Data struct {
		PhoneNumber string `json:"phone_number" example:"1234567890"`
	} `json:"data"`
}

type OTP struct {
	Data struct {
		OTP         string `json:"otp" example:"123456"`
		PhoneNumber string `json:"phone_number" example:"1234567890"`
	} `json:"data"`
}

type Swipe struct {
	Data struct {
		SwiperID  int    `json:"swiper_id" example:"123"`
		ProfileID int    `json:"profile_id" example:"456"`
		SwipeType string `json:"swipe_type" example:"like"`
	}
}

type Package struct {
	Data struct {
		Name      string  `json:"name" example:"Sample Package"`
		Feature   string  `json:"feature" example:"Sample Feature"`
		Price     float64 `json:"price" example:"9.99"`
		Currency  string  `json:"currency" example:"USD"`
	}
}
