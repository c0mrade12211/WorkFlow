package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	jwt_service "todo/JWT"
	dbuser "todo/userslib"
)

func MyTasksHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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
	tasks, err := dbuser.GetTasksByUserID(db, userID)
	if err != nil {
		http.Error(w, "Failed to get tasks", http.StatusInternalServerError)
		return
	}
	resp := make([]map[string]interface{}, len(tasks))
	for i, task := range tasks {
		resp[i] = task
	}
	json.NewEncoder(w).Encode(resp)
}
