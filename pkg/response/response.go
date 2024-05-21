package response

import (
	_ "dating_app/docs"

	_ "github.com/lib/pq"
)

type OTP struct {
	OTP string `json:"otp"`
}
