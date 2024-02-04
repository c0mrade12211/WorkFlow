package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	jwt_service "todo/JWT"
)

type Item_create struct {
	Title       string `json:"title"`
	Price       int    `json:"price"`
	Description string `json:"description"`
}

func CreateItem(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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
	checkAdmin, err := db.Query("SELECT role FROM users WHERE id = $1", userID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to query items", http.StatusInternalServerError)
		return
	}
	defer checkAdmin.Close()
	for checkAdmin.Next() {
		var role string
		err = checkAdmin.Scan(&role)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Failed to scan item", http.StatusInternalServerError)
			return
		}
		if role != "admin" {
			http.Error(w, "You are not admin", http.StatusForbidden)
			return
		}
	}
	item := Item_create{}
	err = json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	fmt.Println(item)
	rows, err := db.Query("INSERT INTO items (title, price, description) VALUES ($1, $2, $3) RETURNING id", item.Title, item.Price, item.Description)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to query items", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

}
