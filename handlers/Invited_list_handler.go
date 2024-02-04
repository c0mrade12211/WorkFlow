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

type User struct {
	ID       int    `json:"user_id"`
	Username string `json:"username"`
}

type Subdivision struct {
	ID   int    `json:"subdivision_id"`
	Name string `json:"subdivision_name"`
}

type InvitedUser struct {
	User        User        `json:"user"`
	Subdivision Subdivision `json:"subdivision"`
}

func InvitedList(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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

	convertUserID, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, "Failed to convert user ID", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	subdivisions := []Subdivision{}
	rows, err := db.Query("SELECT subdivision_id, name FROM subdivisions WHERE owner = $1", convertUserID)
	if err != nil {
		http.Error(w, "Failed to query subdivisions", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var subdivision Subdivision
		err := rows.Scan(&subdivision.ID, &subdivision.Name)
		if err != nil {
			http.Error(w, "Failed to scan subdivision ID and name", http.StatusInternalServerError)
			fmt.Println(err)
			return
		}
		subdivisions = append(subdivisions, subdivision)
	}

	invitedUsers := []InvitedUser{}
	for _, subdivision := range subdivisions {
		rows, err := db.Query("SELECT user_id, username FROM join_sub JOIN users ON join_sub.user_id = users.id WHERE join_sub.subdivision_id = $1", subdivision.ID)
		if err != nil {
			http.Error(w, "Failed to query join_sub", http.StatusInternalServerError)
			fmt.Println(err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var invitedUser InvitedUser
			err := rows.Scan(&invitedUser.User.ID, &invitedUser.User.Username)
			if err != nil {
				http.Error(w, "Failed to scan invited user", http.StatusInternalServerError)
				fmt.Println(err)
				return
			}
			invitedUser.Subdivision = subdivision
			invitedUsers = append(invitedUsers, invitedUser)
		}
	}

	jsonResp, err := json.Marshal(invitedUsers)
	if err != nil {
		http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}
