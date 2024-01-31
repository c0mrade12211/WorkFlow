package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	jwt_service "todo/JWT"
	dbuser "todo/userslib"

	"github.com/gorilla/mux"
)

func DeleteTaskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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
	taskID := params["id"]
	tokenString := authHeaderParts[1]
	userID, err := jwt_service.ParseJWT(tokenString)
	if err != nil {
		http.Error(w, "Invalid JWT token", http.StatusUnauthorized)
		return
	}

	_, err = db.Exec("DELETE FROM tasks WHERE id = $1 AND userid = $2", taskID, userID)
	if err != nil {
		http.Error(w, "Failed to delete task", http.StatusInternalServerError)
		fmt.Println(err)
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
