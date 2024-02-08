package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type TaskResponse struct {
	Creator     CreatorResponse   `json:"creator"`
	Description string            `json:"description"`
	Title       string            `json:"title"`
	ID          int               `json:"id"`
	IsComplete  bool              `json:"iscomplete"`
	CreatedAt   string            `json:"created_at"`
	Comments    []CommentResponse `json:"comments"`
}

type CreatorResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type CommentResponse struct {
	ID        int             `json:"id"`
	Comment   string          `json:"comment"`
	User      CreatorResponse `json:"user"`
	CreatedAt string          `json:"createdAt"`
}

func TaskInfo(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader == "" {
		http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	taskID := vars["task_id"]

	taskIDInt, err := strconv.Atoi(taskID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var task TaskResponse

	err = db.QueryRow("SELECT userid, username, role FROM tasks JOIN users ON tasks.userid = users.id WHERE tasks.id = $1", taskIDInt).Scan(&task.Creator.ID, &task.Creator.Username, &task.Creator.Role)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to get creator info", http.StatusInternalServerError)
		return
	}

	err = db.QueryRow("SELECT description, title, iscomplete, created_at FROM tasks WHERE id = $1", taskIDInt).Scan(&task.Description, &task.Title, &task.IsComplete, &task.CreatedAt)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to get task info", http.StatusInternalServerError)
		return
	}

	rows, err := db.Query("SELECT id, comment, who_create, created_at FROM comments WHERE id_task = $1", taskIDInt)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to get comments", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var comment CommentResponse
		var userID int
		err = rows.Scan(&comment.ID, &comment.Comment, &userID, &comment.CreatedAt)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Failed to scan comment", http.StatusInternalServerError)
			return
		}

		err = db.QueryRow("SELECT id, username, role FROM users WHERE id = $1", userID).Scan(&comment.User.ID, &comment.User.Username, &comment.User.Role)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Failed to get user info for comment", http.StatusInternalServerError)
			return
		}
		task.ID = taskIDInt
		task.Comments = append(task.Comments, comment)
	}

	json.NewEncoder(w).Encode(task)
}
