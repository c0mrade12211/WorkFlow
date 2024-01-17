package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"

	jwt_service "todo/jwt"
	lib "todo/userslib"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"
)

type Task struct {
	Creator     string `json:"creator"`
	Description string `json:"description"`
	Title       string `json:"title"`
	ID          string `json:"id"`
	IsComplete  string `json:"iscomplete"`
	CreatedAt   string `json:"created_at"`
	UserID      string `json:"userid"`
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

var db *sql.DB

func init() {
	connStr := "user=postgres password=test dbname=myapp sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	r := mux.NewRouter()
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
	})
	r.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var user User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}
		err = db.QueryRow("INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id", user.Username, string(hashedPassword)).Scan(&user.ID)
		if err != nil {
			http.Error(w, "Failed to add user", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]interface{})
		resp["user"] = map[string]string{
			"username": user.Username,
			"id":       fmt.Sprintf("%d", user.ID),
		}
		token, err := jwt_service.GenerateJWT(fmt.Sprintf("%d", user.ID), user.Username)
		if err != nil {
			http.Error(w, "Failed to generate JWT", http.StatusInternalServerError)
			return
		}
		resp["token"] = token
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
			return
		}
		w.Write(jsonResp)
	})
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var user User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		row := db.QueryRow("SELECT id, password FROM users WHERE username = $1", user.Username)
		var dbID int
		var dbPassword string
		err = row.Scan(&dbID, &dbPassword)
		if err != nil {
			http.Error(w, "User does not exist", http.StatusForbidden)
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(user.Password))
		if err != nil {
			http.Error(w, "Invalid password", http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]interface{})
		resp["user"] = map[string]string{
			"username": user.Username,
			"id":       fmt.Sprintf("%d", dbID),
		}
		token, err := jwt_service.GenerateJWT(fmt.Sprintf("%d", dbID), user.Username)
		if err != nil {
			http.Error(w, "Failed to generate JWT", http.StatusInternalServerError)
			return
		}
		resp["token"] = token
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
			return
		}
		w.Write(jsonResp)
	})
	r.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		tmpl, _ := template.ParseFiles("test_service.html")
		tmpl.Execute(w, nil)
	})
	r.HandleFunc("/create-task", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
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
		var task Task
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
		row := db.QueryRow("SELECT created_at, id, iscomplete FROM tasks WHERE userid = $1", userID)
		var createdAt string
		var dbID int
		var isComplete bool
		err = row.Scan(&createdAt, &dbID, &isComplete)
		if err != nil {
			http.Error(w, "Error retrieving task", http.StatusInternalServerError)
			return
		}
		rowUsername := db.QueryRow("SELECT username FROM users WHERE id = $1", userID)
		var username string
		err = rowUsername.Scan(&username)
		if err != nil {
			http.Error(w, "Failed to retrieve username", http.StatusInternalServerError)
			return
		}
		resp := make(map[string]interface{})
		resp["user"] = map[string]string{
			"username": username,
			"id":       fmt.Sprintf("%d", userID),
		}
		resp["task_id"] = dbID
		resp["title"] = task.Title
		resp["description"] = task.Description
		resp["created_at"] = createdAt
		resp["is_complete"] = isComplete
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
			return
		}
		w.Write(jsonResp)
	})
	r.HandleFunc("/show-tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
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
			http.Error(w, "Failed to parse JWT token", http.StatusUnauthorized)
			return
		}
		jsonData, err := lib.GetTasksByUserID(userID)
		if err != nil {
			http.Error(w, "Failed to get tasks", http.StatusInternalServerError)
			fmt.Println(err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	})
	r.HandleFunc("/check-auth", func(w http.ResponseWriter, r *http.Request) {
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
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]interface{})
		row := db.QueryRow("SELECT username FROM users WHERE id = $1", userID)
		var username string
		err = row.Scan(&username)
		if err != nil {
			http.Error(w, "Failed to retrieve username", http.StatusInternalServerError)
			return
		}
		token, err := jwt_service.GenerateJWT(userID, "")
		if err != nil {
			http.Error(w, "Failed to generate JWT", http.StatusInternalServerError)
			return
		}
		resp["user"] = map[string]string{
			"username": username,
			"id":       userID,
		}
		resp["token"] = token
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
			return
		}
		w.Write(jsonResp)
	})

	log.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", c.Handler(r)))
}
