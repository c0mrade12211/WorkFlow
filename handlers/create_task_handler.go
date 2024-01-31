package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	jwt_service "todo/JWT"

	models "todo/models"
)

func CreateTaskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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

	var task models.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	tokenString := authHeaderParts[1]
	userID, err := jwt_service.ParseJWT(tokenString)
	if err != nil {
		http.Error(w, "Invalid JWT token", http.StatusUnauthorized)
		return
	}

	_, err = db.Exec("INSERT INTO tasks (title, description, userid) VALUES ($1, $2, $3)", task.Title, task.Description, userID)
	if err != nil {
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	row := db.QueryRow("SELECT created_at, id FROM tasks WHERE userid = $1 ORDER BY created_at DESC LIMIT 1", userID)
	var createdAt string
	var dbID int
	err = row.Scan(&createdAt, &dbID)
	if err != nil {
		http.Error(w, "Error retrieving task", http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"user": map[string]interface{}{
			"username": "",
			"id":       userID,
			"tasks": []map[string]interface{}{
				{
					"created_at":  createdAt,
					"description": task.Description,
					"id":          dbID,
				},
			},
		},
	}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
		return
	}

	w.Write(jsonResp)
}
