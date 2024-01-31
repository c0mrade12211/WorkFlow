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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	err = db.QueryRow("INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id", user.Username, string(hashedPassword)).Scan(&user.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to add user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{
		"user": map[string]interface{}{
			"username": user.Username,
			"id":       user.ID,
			"balance":  0, // Используйте тип int для целочисленного значения
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
