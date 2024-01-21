package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	jwt_service "todo/jwt"

	"todo/models"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func GetTasksByUserID(userID string) ([]map[string]interface{}, error) {
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
		AllowedOrigins:   []string{"https://7078-95-26-28-58.ngrok-free.app"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})
	r.HandleFunc("/register", registerHandler).Methods("POST")
	r.HandleFunc("/login", loginHandler).Methods("POST")
	r.HandleFunc("/get-me", getProfileHandler).Methods("GET")
	r.HandleFunc("/create-task", createTaskHandler).Methods("POST")
	r.HandleFunc("/delete-task/{id}", deleteTaskHandler).Methods("DELETE")
	r.HandleFunc("/check-auth", checkAuthHandler).Methods("GET")
	r.HandleFunc("/shop", shopHandler).Methods("GET")
	r.HandleFunc("/buy/{id}", buyHandler).Methods("POST")
	log.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", c.Handler(r)))
}

func buyHandler(w http.ResponseWriter, r *http.Request) {
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
	row_user_balance := db.QueryRow("SELECT balance FROM users WHERE id = $1", userID)
	var user_balance int
	err = row_user_balance.Scan(&user_balance)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to get user balance", http.StatusInternalServerError)
		return
	}
	params := mux.Vars(r)
	itemID := params["id"]
	row_item := db.QueryRow("SELECT price FROM items WHERE id = $1", itemID)
	var itemPrice int
	err = row_item.Scan(&itemPrice)
	if err != nil {
		http.Error(w, "Failed to get item price", http.StatusInternalServerError)
		return
	}
	if itemPrice > user_balance {
		http.Error(w, "Insufficient balance", http.StatusForbidden)
		return
	}
	updateUserbalance, err := db.Exec("UPDATE users SET balance = balance - $1 WHERE id = $2", itemPrice, userID)
	if err != nil {
		http.Error(w, "Failed to update user balance", http.StatusInternalServerError)
		return
	}
	rowsAffected, err := updateUserbalance.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to get number of rows affected", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Failed to update user balance", http.StatusInternalServerError)
		return
	}
	_, err = db.Exec("INSERT INTO buyuserinfo (userid, itemid) VALUES ($1, $2)", userID, itemID)
	if err != nil {
		http.Error(w, "Failed to insert user item", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("OK"))
}
func shopHandler(w http.ResponseWriter, r *http.Request) {
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
	fmt.Println(userID)
	row := db.QueryRow("SELECT id, price, description, title FROM items")
	var idItem int
	var PriceItem int
	var DescriptionItem string
	var TitleItem string

	err = row.Scan(&idItem, &PriceItem, &DescriptionItem, &TitleItem)
	if err != nil {
		http.Error(w, "User does not exist", http.StatusForbidden)
		return
	}
	item := map[string]interface{}{
		"id":          idItem,
		"price":       PriceItem,
		"description": DescriptionItem,
		"title":       TitleItem,
	}
	w.Header().Set("Content-Type", "application/json")
	jsonResp, err := json.Marshal(item)
	if err != nil {
		http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
		return
	}
	w.Write(jsonResp)
}

func getProfileHandler(w http.ResponseWriter, r *http.Request) {
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
	tasks, err := GetTasksByUserID(userID)
	if err != nil {
		http.Error(w, "Failed to get tasks", http.StatusInternalServerError)
		return
	}
	row := db.QueryRow("SELECT username FROM users WHERE id = $1", userID)
	var username string
	err = row.Scan(&username)
	if err != nil {
		http.Error(w, "User does not exist", http.StatusForbidden)
		return
	}
	resp := map[string]interface{}{
		"user": map[string]interface{}{
			"username": username,
			"id":       userID,
			"tasks":    tasks,
		},
	}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}
func registerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "https://7078-95-26-28-58.ngrok-free.app")

	var user models.User
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
		fmt.Println(err)
		http.Error(w, "Failed to add user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{
		"user": map[string]string{
			"username": user.Username,
			"id":       fmt.Sprintf("%d", user.ID),
		},
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
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "https://7078-95-26-28-58.ngrok-free.app")

	var user models.User
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
	tasks, err := GetTasksByUserID(fmt.Sprintf("%d", dbID))
	if err != nil {
		http.Error(w, "Failed to get tasks", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{
		"user": map[string]interface{}{
			"username": user.Username,
			"id":       fmt.Sprintf("%d", dbID),
			"tasks":    tasks,
		},
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
	cookie := &http.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	}

	http.SetCookie(w, cookie)

	w.Write(jsonResp)
}

func createTaskHandler(w http.ResponseWriter, r *http.Request) {
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
			"id":       fmt.Sprintf("%d", userID),
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

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
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

	w.WriteHeader(http.StatusOK)
}

func checkAuthHandler(w http.ResponseWriter, r *http.Request) {
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

	tasks, err := GetTasksByUserID(fmt.Sprintf("%d", userID))
	if err != nil {
		http.Error(w, "Failed to get tasks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{
		"user": map[string]interface{}{
			"username": "",
			"id":       fmt.Sprintf("%d", userID),
			"tasks":    tasks,
		},
	}

	token, err := jwt_service.GenerateJWT(fmt.Sprintf("%d", userID), "")
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
}
