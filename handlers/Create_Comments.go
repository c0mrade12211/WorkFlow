package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	jwt_service "todo/JWT"

	"github.com/gorilla/mux"
)

type Comment struct {
	Comment string `json:"comment"`
}

func CreateComments(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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
	item_id := params["task_id"]
	comment := Comment{}
	err = json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	_, err = db.Exec("INSERT INTO comments (comment, who_create, id_task) VALUES ($1, $2, $3)", comment.Comment, userID, item_id)
	if err != nil {
		http.Error(w, "Failed to create comments", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	w.Write([]byte("OK comment will be create "))

}
