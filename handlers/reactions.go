package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
)

func HandleReaction(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := getCurrentUser(r, db)
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		targetType := r.FormValue("type") // "post" or "comment"
		targetID := r.FormValue("target_id")
		reactionType := r.FormValue("reaction") // "like" or "dislike"

		id, err := strconv.Atoi(targetID)
		if err != nil {
			http.Error(w, "Invalid target ID", http.StatusBadRequest)
			return
		}

		// Delete existing reaction if any
		_, err = db.Exec(`DELETE FROM reactions 
						 WHERE user_id = ? AND target_type = ? AND target_id = ?`,
			user.ID, targetType, id)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		// Insert new reaction
		_, err = db.Exec(`INSERT INTO reactions 
						 WHERE user_id = ? AND target_type = ? AND target_id = ?`,
			user.ID, targetType, id, reactionType)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	}
}

func getCurrentUser(r *http.Request, db *sql.DB) *User {
	// Get session token from cookie
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil
	}

	// Look up user by session token
	var user User
	err = db.QueryRow("SELECT id, username FROM users WHERE session_token = ?", cookie.Value).Scan(&user.ID, &user.Username)
	if err != nil {
		return nil
	}

	return &user
}
