package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	_ "dating_app/docs"

	"dating_app/api/middleware"
	"dating_app/pkg/model"

	_ "github.com/lib/pq"
)

// SwipeHandler handles swiping left or right
// @Summary Swipe
// @Description Swipe left or right on a profile.
// @Accept json
// @Produce json
// @Param data body payload.Swipe true "Swipe object"
// @Success 201 {string} string "Swipe recorded successfully"
// @Failure 400 {string} string "Invalid request format"
// @Failure 500 {string} string "Internal server error"
// @Router /swipe [post]
func Swipe(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var swipe model.Swipe
	
	if err := json.NewDecoder(r.Body).Decode(&swipe); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the current user ID from the context
	userID, err := middleware.CurrentUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	swipe.SwiperID = userID

	// Check if user has exceeded the daily swipe limit
	if err := checkDailySwipeLimit(db, swipe.SwiperID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if user has already swiped this profile today
	if err := checkDuplicateSwipe(db, swipe.SwiperID, swipe.ProfileID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = db.Exec("INSERT INTO swipes (swiper_id, profile_id, swipe_type, swipe_date) VALUES ($1, $2, $3, $4)", swipe.SwiperID, swipe.ProfileID, swipe.SwipeType, time.Now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// checkDailySwipeLimit checks if the user has exceeded the daily swipe limit
func checkDailySwipeLimit(db *sql.DB, userID int) error {
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
func checkDuplicateSwipe(db *sql.DB, userID, profileID int) error {
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
