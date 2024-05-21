package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/sessions"

	"dating_app/pkg/model"
)

var (
	cardShown = make(map[int]time.Time)
)

// getCardsHandler handles retrieving cards based on preferences
// @Summary Get a list of cards based on user preferences
// @Description Get a list of cards based on the logged-in user's preferences.
// @Accept json
// @Produce json
// @Success 200 {array} model.Card "List of cards matching user's preferences"
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Internal server error"
// @Router /cards [get]
func Card(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var store = sessions.NewCookieStore([]byte("super-secret-key"))
	session, _ := store.Get(r, "session-name")

	// Check if card is authenticated
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	preferences, err := getCardPreferences(db, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cards, err := getCardsBasedOnPreferences(db, preferences)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cards)
}

// getCardPreferences retrieves the preferences of the logged-in card
func getCardPreferences(db *sql.DB, userID int) (model.Preference, error) {
	var preferences model.Preference

	err := db.QueryRow("SELECT id, user_id, date_mode, bff_mode, preferred_gender, min_age, max_age, created_at, updated_at FROM preferences WHERE user_id = $1", userID).Scan(
		&preferences.ID, &preferences.UserID, &preferences.DateMode, &preferences.BFFMode, &preferences.PreferredGender, &preferences.MinAge, &preferences.MaxAge, &preferences.CreatedAt, &preferences.UpdatedAt)
	if err != nil {
		return preferences, err
	}
	return preferences, nil
}

// getCardsBasedOnPreferences retrieves a list of cards based on the preferences
func getCardsBasedOnPreferences(db *sql.DB, preferences model.Preference) ([]model.Card, error) {
	query := `
		SELECT u.id, u.phone_number, u.is_premium, u.verified, u.is_deleted, u.signup_at, u.login_at, u.logout_at, p.name, p.age, p.bio, p.photo_url
		FROM users u
		JOIN profiles p ON u.id = p.user_id
		WHERE u.is_deleted = FALSE AND u.id != $1
	`

	args := []interface{}{preferences.UserID}

	if preferences.PreferredGender != "" && preferences.PreferredGender != "both" {
		query += " AND p.gender = $2"
		args = append(args, preferences.PreferredGender)
	}

	if preferences.MinAge > 0 {
		query += " AND p.age >= $3"
		args = append(args, preferences.MinAge)
	}

	if preferences.MaxAge > 0 {
		query += " AND p.age <= $4"
		args = append(args, preferences.MaxAge)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []model.Card
	
	for rows.Next() {
		var card model.Card

		if err := rows.Scan(&card.UserID, &card.Verified, &card.Name, &card.Age, &card.Bio, &card.PhotoURL); err != nil {
			return nil, err
		}

		// Check if the card has been shown today
		if !isCardShownToday(card.UserID) {
			cards = append(cards, card)
			// Log that the card has been shown
			logCardShown(card.UserID)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return cards, nil
}

// logCardShown logs the card shown
func logCardShown(cardID int) {
	cardShown[cardID] = time.Now().Truncate(24 * time.Hour)
}

// isCardShownToday checks if the card has been shown today
func isCardShownToday(cardID int) bool {
	t, ok := cardShown[cardID]
	if !ok {
		return false
	}
	return t.Equal(time.Now().Truncate(24 * time.Hour))
}
