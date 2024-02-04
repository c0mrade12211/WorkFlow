package userslib

import (
	"database/sql"
	"fmt"
)

func GetTasksByUserID(db *sql.DB, userID string) ([]map[string]interface{}, error) {
<<<<<<< HEAD
	rows, err := db.Query("SELECT id, created_at, description, title, iscomplete FROM tasks WHERE userid = $1 ORDER BY created_at DESC", userID)
=======
	rows, err := db.Query("SELECT id, created_at, description, title FROM tasks WHERE userid = $1 ORDER BY created_at DESC", userID)
>>>>>>> 120ebd7a25d59279905e416eea67a18cbcaed647
	if err != nil {
		fmt.Println("Error querying tasks:", err)
		return nil, err
	}
	defer rows.Close()
	tasks := []map[string]interface{}{}
	for rows.Next() {
<<<<<<< HEAD
		var taskID, taskCreatedAt, taskDescription, title, iscomplete interface{}
		err := rows.Scan(&taskID, &taskCreatedAt, &taskDescription, &title, &iscomplete)
=======
		var taskID, taskCreatedAt, taskDescription, title interface{}
		err := rows.Scan(&taskID, &taskCreatedAt, &taskDescription, &title)
>>>>>>> 120ebd7a25d59279905e416eea67a18cbcaed647
		if err != nil {
			fmt.Println("Error scanning task:", err)
			return nil, err
		}
		task := map[string]interface{}{
			"id":          taskID,
			"title":       title,
			"created_at":  taskCreatedAt,
			"description": taskDescription,
<<<<<<< HEAD
			"iscomplete":  iscomplete,
=======
>>>>>>> 120ebd7a25d59279905e416eea67a18cbcaed647
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}
