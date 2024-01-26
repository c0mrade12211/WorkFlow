package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"

	jwt_service "todo/jwt"
	"todo/models"

	_ "github.com/lib/pq"
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
		AllowedOrigins:   []string{"https://e395-87-244-58-26.ngrok-free.app"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS", "DELETE"},
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
	r.HandleFunc("/my-subdivision", getTasksInSubdivisionHandler).Methods("GET")

	log.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", c.Handler(r)))
}
func getTasksInSubdivisionHandler(w http.ResponseWriter, r *http.Request) {
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

	rowUserBalance := db.QueryRow("SELECT balance FROM users WHERE id = $1", userID)
	var userBalance int
	err = rowUserBalance.Scan(&userBalance)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to get user balance", http.StatusInternalServerError)
		return
	}

	params := mux.Vars(r)
	itemID := params["id"]
	rowItem := db.QueryRow("SELECT price FROM items WHERE id = $1", itemID)
	var itemPrice int
	err = rowItem.Scan(&itemPrice)
	if err != nil {
		http.Error(w, "Failed to get item price", http.StatusInternalServerError)
		return
	}

	if itemPrice > userBalance {
		http.Error(w, "Insufficient balance", http.StatusForbidden)
		return
	}

	updateUserBalance, err := db.Exec("UPDATE users SET balance = balance - $1 WHERE id = $2", itemPrice, userID)
	if err != nil {
		http.Error(w, "Failed to update user balance", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := updateUserBalance.RowsAffected()
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
	var priceItem int
	var descriptionItem string
	var titleItem string
	err = row.Scan(&idItem, &priceItem, &descriptionItem, &titleItem)
	if err != nil {
		http.Error(w, "User does not exist", http.StatusForbidden)
		return
	}

	item := map[string]interface{}{
		"id":          idItem,
		"price":       priceItem,
		"description": descriptionItem,
		"title":       titleItem,
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

	row := db.QueryRow("SELECT username, balance FROM users WHERE id = $1", userID)
	var username string
	var balance int
	err = row.Scan(&username, &balance)
	if err != nil {
		http.Error(w, "User does not exist", http.StatusForbidden)
		return
	}

	resp := map[string]interface{}{
		"user": map[string]interface{}{
			"username": username,
			"id":       userID,
			"tasks":    tasks,
			"balance":  balance,
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

	err = db.QueryRow("INSERT INTO users (username, password, subdivision) VALUES ($1, $2, $3) RETURNING id", user.Username, string(hashedPassword), user.Subdivision).Scan(&user.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to add user", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	tasks := []interface{}{}
	resp := map[string]interface{}{
		"user": map[string]interface{}{
			"username": user.Username,
			"id":       fmt.Sprintf("%d", user.ID),
			"tasks":    tasks,
			"balance":  "0",
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
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	row := db.QueryRow("SELECT id, password, balance FROM users WHERE username = $1", user.Username)
	var userID int
	var dbPassword string
	var balance int
	err = row.Scan(&userID, &dbPassword, &balance)
	if err != nil {
		http.Error(w, "User does not exist", http.StatusForbidden)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(user.Password))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	tasks, err := GetTasksByUserID(fmt.Sprintf("%d", userID))
	if err != nil {
		http.Error(w, "Failed to get tasks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	token, err := jwt_service.GenerateJWT(fmt.Sprintf("%d", userID), "")
	if err != nil {
		http.Error(w, "Failed to generate JWT", http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"user": map[string]interface{}{
			"username": user.Username,
			"id":       fmt.Sprintf("%d", userID),
			"tasks":    tasks,
			"balance":  balance,
		},
		"token": token,
	}

	json.NewEncoder(w).Encode(resp)
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
