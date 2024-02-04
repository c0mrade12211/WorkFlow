package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	jwt_service "todo/JWT"
)

type Item struct {
	PriceItem   int    `json:"price"`
	Description string `json:"description"`
	ID          int    `json:"id"`
	Title       string `json:"title"`
}

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

	rows, err := db.Query("SELECT id, price, description, title FROM items")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to query items", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []Item

	for rows.Next() {
		var idItem int
		var priceItem int
		var descriptionItem string
		var titleItem string

		err = rows.Scan(&idItem, &priceItem, &descriptionItem, &titleItem)
		if err != nil {
			http.Error(w, "Failed to scan item", http.StatusInternalServerError)
			return
		}

		item := Item{
			ID:          idItem,
			Description: descriptionItem,
			Title:       titleItem,
			PriceItem:   priceItem,
		}

		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, "Error occurred while iterating over items", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResp, err := json.Marshal(items)
	if err != nil {
		http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
		return
	}

	w.Write(jsonResp)
}
