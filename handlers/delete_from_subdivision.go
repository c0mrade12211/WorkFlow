package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	jwt_service "todo/JWT"

	"github.com/gorilla/mux"
)

func DeleteUserFromSubdivision(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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

	row := db.QueryRow("SELECT owner FROM subdivisions WHERE owner = $1", userID)
	var owner int
	err = row.Scan(&owner)
	if err != nil {
		http.Error(w, "Failed to select user", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

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

	getSubdivisionID, err := db.Query("SELECT subdivision_id FROM subdivisions WHERE owner = $1", convertUserID)
	if err != nil {
		http.Error(w, "Failed to query subdivisions", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	defer getSubdivisionID.Close()

	var subdivisionID int
	if getSubdivisionID.Next() {
		err = getSubdivisionID.Scan(&subdivisionID)
		if err != nil {
			http.Error(w, "Failed to scan subdivision ID", http.StatusInternalServerError)
			fmt.Println(err)
			return
		}
	}

	params := mux.Vars(r)
	userIDToDelete := params["id"]

	_, err = db.Exec("UPDATE users SET subdivision = null WHERE id = $1", userIDToDelete)
	if err != nil {
		http.Error(w, "Failed to delete user from subdivision", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
}
