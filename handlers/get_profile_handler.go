package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	jwt_service "todo/JWT"
)

type UserGet struct {
	Balance     string          `json:"balance"`
	ID          string          `json:"id"`
	Role        string          `json:"role"`
	Username    string          `json:"username"`
	Subdivision *SubdivisionGet `json:"subdivision"`
}

type SubdivisionGet struct {
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
	var subdivisionID sql.NullInt64
	err = row.Scan(&username, &balance, &role, &subdivisionID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "User does not exist", http.StatusForbidden)
		return
	}

	var subdivisionInfo *SubdivisionGet
	if subdivisionID.Valid {
		subdivisionRow := db.QueryRow("SELECT name FROM subdivisions WHERE subdivision_id = $1", subdivisionID.Int64)
		var subdivisionName string
		err = subdivisionRow.Scan(&subdivisionName)
		if err != nil {
			http.Error(w, "Failed to get subdivision info", http.StatusInternalServerError)
			return
		}
		subdivisionInfo = &SubdivisionGet{
			Name: subdivisionName,
			ID:   int(subdivisionID.Int64),
		}
	} else {
		subdivisionInfo = nil
	}

	user := UserGet{
		Username:    username,
		ID:          userID,
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
