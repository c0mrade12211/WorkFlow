package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func ChangeStatus(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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

	//tokenString := authHeaderParts[1]
	/**userID, err := jwt_service.ParseJWT(tokenString)
	if err != nil {
		http.Error(w, "Invalid JWT token", http.StatusUnauthorized)
		return
	}
	**/
	taskID := mux.Vars(r)["id"]
	value_now, err := db.Query("SELECT iscomplete FROM tasks WHERE id = $1", taskID)

	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to update task", http.StatusInternalServerError)
		return
	}
	defer value_now.Close()
	var iscomplete bool
	for value_now.Next() {
		err = value_now.Scan(&iscomplete)
		if err != nil {
			http.Error(w, "Failed to scan item", http.StatusInternalServerError)
			fmt.Println(err)
			return
		}
	}
	if iscomplete == true {
		row, err := db.Query("UPDATE tasks SET iscomplete = false WHERE id = $1", taskID)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Failed to update task", http.StatusInternalServerError)
			return
		}
		defer row.Close()
	} else {
		row, err := db.Query("UPDATE tasks SET iscomplete = true WHERE id = $1", taskID)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Failed to update task", http.StatusInternalServerError)
			return
		}
		defer row.Close()
	}

}
