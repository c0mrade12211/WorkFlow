package userslib

import (
	"database/sql"
	"fmt"
)

func GetTasksByUserID(db *sql.DB, userID string) ([]map[string]interface{}, error) {
	rows, err := db.Query("SELECT id, created_at, description, title FROM tasks WHERE userid = $1 ORDER BY created_at DESC", userID)
	if err != nil {
		fmt.Println("Error querying tasks:", err)
		return nil, err
	}
	defer rows.Close()
	tasks := []map[string]interface{}{}
	for rows.Next() {
		var taskID, taskCreatedAt, taskDescription, title interface{}
		err := rows.Scan(&taskID, &taskCreatedAt, &taskDescription, &title)
		if err != nil {
			fmt.Println("Error scanning task:", err)
			return nil, err
		}
		task := map[string]interface{}{
			"id":          taskID,
			"title":       title,
			"created_at":  taskCreatedAt,
			"description": taskDescription,
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}
