package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	jwt_service "todo/JWT"
)

func GetProfileHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader == "" {
		http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
		return
	}

	authHeaderParts := strings.Split(authorizationHeader, " ")
	if len(authHeaderParts) != 2 || authHeaderParts[0] != "Bearer" {
		http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
		return
	}

	tokenString := authHeaderParts[1]
	userID, err := jwt_service.ParseJWT(tokenString)
	if err != nil {
		http.Error(w, "Invalid JWT token", http.StatusUnauthorized)
		return
	}

	row := db.QueryRow("SELECT username, balance, role FROM users WHERE id = $1", userID)
	var username string
	var balance int
	var role string
	err = row.Scan(&username, &balance, &role)
	if err != nil {
		http.Error(w, "User does not exist", http.StatusForbidden)
		return
	}

	resp := map[string]interface{}{
		"user": map[string]interface{}{
			"username": username,
			"id":       userID,

			"balance": balance,
			"role":    role,
		},
	}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}
