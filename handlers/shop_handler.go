package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	jwt_service "todo/JWT"
)

func ShopHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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

	fmt.Println(userID)

	row := db.QueryRow("SELECT id, price, description, title FROM items")
	var idItem int
	var priceItem int
	var descriptionItem string
	var titleItem string
	err = row.Scan(&idItem, &priceItem, &descriptionItem, &titleItem)
	if err != nil {
		http.Error(w, "User does not exist", http.StatusForbidden)
		return
	}

	item := map[string]interface{}{
		"id":          idItem,
		"price":       priceItem,
		"description": descriptionItem,
		"title":       titleItem,
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResp, err := json.Marshal(item)
	if err != nil {
		http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
		return
	}

	w.Write(jsonResp)
}
