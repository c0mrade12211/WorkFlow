package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	jwt_service "todo/JWT"

	"github.com/gorilla/mux"
)

func BuyHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader == "" {
		http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
		fmt.Println("hello")
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

	rowUserBalance := db.QueryRow("SELECT balance FROM users WHERE id = $1", userID)
	var userBalance int
	err = rowUserBalance.Scan(&userBalance)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to get user balance", http.StatusInternalServerError)
		return
	}

	params := mux.Vars(r)
	itemID := params["id"]
	rowItem := db.QueryRow("SELECT price FROM items WHERE id = $1", itemID)
	var itemPrice int
	err = rowItem.Scan(&itemPrice)
	if err != nil {
		http.Error(w, "Failed to get item price", http.StatusInternalServerError)
		return
	}

	if itemPrice > userBalance {
		http.Error(w, "Insufficient balance", http.StatusForbidden)
		return
	}

	updateUserBalance, err := db.Exec("UPDATE users SET balance = balance - $1 WHERE id = $2", itemPrice, userID)
	if err != nil {
		http.Error(w, "Failed to update user balance", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := updateUserBalance.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to get number of rows affected", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Failed to update user balance", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("INSERT INTO buyuserinfo (userid, itemid) VALUES ($1, $2)", userID, itemID)
	if err != nil {
		http.Error(w, "Failed to insert user item", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("OK"))
}
