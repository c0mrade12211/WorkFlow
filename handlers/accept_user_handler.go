package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	jwt_service "todo/JWT"
)

type AcceptStruct struct {
	UserID        int `json:"user_id"`
	SubdivisionID int `json:"subdivision_id"`
}

func AcceptUser(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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

	row, err := db.Query("SELECT owner FROM subdivisions WHERE owner = $1", userID)
	if err != nil {
		http.Error(w, "Failed to select user", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	defer row.Close()

	var owner int
	row.Next()
	row.Scan(&owner)

	convertUserID, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, "Failed to convert user ID", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	if owner != convertUserID {
		http.Error(w, "You are not the owner of this subdivision", http.StatusUnauthorized)
		return
	}

	var accept AcceptStruct
	err = json.NewDecoder(r.Body).Decode(&accept)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE users SET subdivision = $1 WHERE id = $2", accept.SubdivisionID, accept.UserID)
	if err != nil {
		http.Error(w, "Failed to update subdivision", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	_, err = db.Exec("DELETE FROM join_sub WHERE subdivision_id = $1 AND user_id = $2", accept.SubdivisionID, accept.UserID)
	if err != nil {
		http.Error(w, "Failed to delete join_sub", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	fmt.Println("ALLL USER CREATE AND DELETE from join_sub")
}
