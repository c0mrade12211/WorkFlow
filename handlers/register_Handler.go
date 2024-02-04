package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	jwt_service "todo/JWT"
	models "todo/models"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var existingUsername string
	err = db.QueryRow("SELECT username FROM users WHERE username = $1", user.Username).Scan(&existingUsername)
	if err == nil {
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	} else if err != sql.ErrNoRows {
		http.Error(w, "Failed to check for existing username", http.StatusInternalServerError)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	role := "user"
	err = db.QueryRow("INSERT INTO users (username, password, role) VALUES ($1, $2, $3) RETURNING id", user.Username, string(hashedPassword), role).Scan(&user.ID)
	if err != nil {
		http.Error(w, "Failed to add user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{
		"user": map[string]interface{}{
			"username": user.Username,
			"id":       user.ID,
			"balance":  0,
			"role":     role,
		},
	}
	token, err := jwt_service.GenerateJWT(fmt.Sprintf("%d", user.ID), user.Username)
	if err != nil {
		http.Error(w, "Failed to generate JWT", http.StatusInternalServerError)
		return
	}
	resp["token"] = token
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
		return
	}
	w.Write(jsonResp)
}
