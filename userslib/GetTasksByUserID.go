package test

import (
	"database/sql"
	"encoding/json"
)

var db *sql.DB

func GetTasksByUserID(userID string) ([]byte, error) {
	rows, err := db.Query("SELECT id, title, iscomplete, created_at, description, userid FROM tasks WHERE userid = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rowUsername := db.QueryRow("SELECT username FROM users WHERE id = $1", userID)
	var username string
	err = rowUsername.Scan(&username)
	if err != nil {
		return nil, err
	}
	resp := make(map[string]interface{})
	resp["user"] = map[string]string{
		"username": username,
		"userid":   userID,
	}
	tasks := []map[string]interface{}{}
	for rows.Next() {
		var task map[string]interface{}
		var taskID, taskTitle, taskIsComplete, taskCreatedAt, taskDescription, taskUserID interface{}
		err := rows.Scan(&taskID, &taskTitle, &taskIsComplete, &taskCreatedAt, &taskDescription, &taskUserID)
		if err != nil {
			return nil, err
		}
		task = make(map[string]interface{})
		task["id"] = taskID
		task["title"] = taskTitle
		task["iscomplete"] = taskIsComplete
		task["created_at"] = taskCreatedAt
		task["description"] = taskDescription
		task["userid"] = taskUserID
		task["username"] = username
		tasks = append(tasks, task)
	}
	resp["tasks"] = tasks
	jsonData, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}
