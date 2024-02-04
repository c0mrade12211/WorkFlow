package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	jwt_service "todo/JWT"

	"github.com/gorilla/mux"
)

func RequestForInvite(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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
	params := mux.Vars(r)
	subdivision_id := params["subdiv_id"]

	row, err := db.Query("INSERT INTO join_sub (subdivision_id, user_id) VALUES ($1, $2)", subdivision_id, userID)
	if err != nil {
		http.Error(w, "Failed to insert request", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	defer row.Close()

	fmt.Fprintf(w, "Request sent")
	return
}
