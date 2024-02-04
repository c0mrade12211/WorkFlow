package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	jwt_service "todo/JWT"
	models "todo/models"
	dbuser "todo/userslib"

	"golang.org/x/crypto/bcrypt"
)

func LoginHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	row := db.QueryRow("SELECT id, password, balance, role FROM users WHERE username = $1", user.Username)
	var userID int
	var role string
	var dbPassword string
	var balance int
	err = row.Scan(&userID, &dbPassword, &balance, &role)
	if err != nil {
		http.Error(w, "User does not exist", http.StatusForbidden)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(user.Password))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	tasks, err := dbuser.GetTasksByUserID(db, fmt.Sprintf("%d", userID))
	if err != nil {
		http.Error(w, "Failed to get tasks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	token, err := jwt_service.GenerateJWT(fmt.Sprintf("%d", userID), "")
	if err != nil {
		http.Error(w, "Failed to generate JWT", http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"user": map[string]interface{}{
			"username": user.Username,
			"role":     role,
			"id":       fmt.Sprintf("%d", userID),
			"tasks":    tasks,
			"balance":  balance,
		},
		"token": token,
	}

	json.NewEncoder(w).Encode(resp)
}
