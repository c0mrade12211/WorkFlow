package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	jwt_service "todo/JWT"
)

func GetTasksInSubdivisionHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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

	rows, err := db.Query(`
		SELECT tasks.id, tasks.created_at, tasks.description, tasks.title, users.subdivision
		FROM tasks
		INNER JOIN users ON tasks.userid = users.id
		WHERE users.id = $1 OR users.subdivision = (SELECT subdivision FROM users WHERE id = $1)
		ORDER BY tasks.created_at DESC`, userID)
	if err != nil {
		http.Error(w, "Failed to get tasks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tasks := []map[string]interface{}{}
	for rows.Next() {
		var taskID, taskCreatedAt, taskDescription, taskTitle, taskSubdivision interface{}
		err := rows.Scan(&taskID, &taskCreatedAt, &taskDescription, &taskTitle, &taskSubdivision)
		if err != nil {
			http.Error(w, "Failed to scan task", http.StatusInternalServerError)
			return
		}

		task := map[string]interface{}{
			"id":          taskID,
			"created_at":  taskCreatedAt,
			"description": taskDescription,
			"title":       taskTitle,
		}

		if taskSubdivision != nil {
			task["subdivision"] = taskSubdivision.(string)
		} else {
			task["subdivision"] = ""
		}

		tasks = append(tasks, task)
	}

	jsonResp, err := json.Marshal(tasks)
	if err != nil {
		http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}
