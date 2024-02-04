package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	jwt_service "todo/JWT"
)

type ItemResponse struct {
	PriceItem   int    `json:"price"`
	Description string `json:"description"`
	ID          int    `json:"id"`
	Title       string `json:"title"`
	UniqID      int    `json:"uniq_id"`
}

func MyItems(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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
	rows, err := db.Query("SELECT itemid, uniq_id FROM buyuserinfo WHERE userid = $1", userID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to query items", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []ItemResponse

	for rows.Next() {
		var idItem int
		var uniqID int
		err = rows.Scan(&idItem, &uniqID)
		if err != nil {
			http.Error(w, "Failed to scan item", http.StatusInternalServerError)
			return
		}

		rowItem, err := db.Query("SELECT * from items WHERE id = $1", idItem)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Failed to query items", http.StatusInternalServerError)
			return
		}
		defer rowItem.Close()

		for rowItem.Next() {
			var id int
			var price int
			var description string
			var title string
			err = rowItem.Scan(&id, &price, &description, &title)
			if err != nil {
				http.Error(w, "Failed to scan item", http.StatusInternalServerError)
				return
			}
			item := ItemResponse{
				ID:          id,
				Description: description,
				Title:       title,
				UniqID:      uniqID,
				PriceItem:   price,
			}
			items = append(items, item)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResp, err := json.Marshal(items)
	if err != nil {
		http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
		return
	}
	w.Write(jsonResp)
}
