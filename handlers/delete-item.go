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

func DeleteItem(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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

	vars := mux.Vars(r)
	getItemid := vars["item_id"]

	itemID, err := strconv.Atoi(getItemid)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM items WHERE id = $1", itemID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to delete item", http.StatusInternalServerError)
		return
	}

	fmt.Println("[+] Request delete item")
}
