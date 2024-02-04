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

func UseMyItem(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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

	params := mux.Vars(r)
	item_uniq_id := params["id"]
	tokenString := authHeaderParts[1]
	userID, err := jwt_service.ParseJWT(tokenString)
	if err != nil {
		http.Error(w, "Invalid JWT token", http.StatusUnauthorized)
		return
	}
	row, err := db.Query("SELECT userid FROM buyuserinfo WHERE uniq_id = $1", item_uniq_id)
	if err != nil {
		http.Error(w, "Failed to get buyuserinfo", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	defer row.Close()
	var creater_item int
	for row.Next() {
		err := row.Scan(&creater_item)
		if err != nil {
			http.Error(w, "Failed to get buyuserinfo", http.StatusInternalServerError)
			fmt.Println(err)
			return
		}
	}
	i, err := strconv.Atoi(userID)
	if i != creater_item {
		http.Error(w, "You are not the owner of this item", http.StatusUnauthorized)
		return
	} else {
		delete_row, err := db.Query("DELETE FROM buyuserinfo WHERE uniq_id = $1", item_uniq_id)
		if err != nil {
			http.Error(w, "Failed to delete buyuserinfo", http.StatusInternalServerError)
			fmt.Println(err)
			return
		}
		defer delete_row.Close()

	}
}
