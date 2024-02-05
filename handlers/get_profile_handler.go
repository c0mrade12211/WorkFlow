package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	jwt_service "todo/JWT"
)

type User_get struct {
	Balance     string          `json:"balance"`
	ID          int             `json:"id"`
	Role        string          `json:"role"`
	Username    string          `json:"username"`
	Subdivision Subdivision_get `json:"subdivision"`
}

type Subdivision_get struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

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
	row := db.QueryRow("SELECT username, balance, role, subdivision FROM users WHERE id = $1", userID)
	var username string
	var balance string
	var role string
	var subdivisionID int
	err = row.Scan(&username, &balance, &role, &subdivisionID)
	if err != nil {
		http.Error(w, "User does not exist", http.StatusForbidden)
		return
	}
	subdivisionInfo := Subdivision_get{}
	err = db.QueryRow("SELECT name FROM subdivisions WHERE subdivision_id = $1", subdivisionID).Scan(&subdivisionInfo.Name)
	if err != nil {
		http.Error(w, "Failed to get subdivision info", http.StatusInternalServerError)
		return
	}
	subdivisionInfo.ID = subdivisionID
	convert_userid, err := strconv.Atoi(userID)
	user := User_get{
		Username:    username,
		ID:          convert_userid,
		Balance:     balance,
		Role:        role,
		Subdivision: subdivisionInfo,
	}
	resp := map[string]interface{}{
		"user": user,
	}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}
