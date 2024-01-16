package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"
)

type Task struct {
	ID       int    `json:"id"`
	User     string `json:"user"`
	Title    string `json:"title"`
	Complete bool   `json:"complete"`
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

func isTokenValid(tokenString string, secretKey []byte) bool {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unsupported signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
	if err != nil {
		fmt.Println("Error parsing token:", err)
		return false
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		expirationTime := time.Unix(int64(claims["exp"].(float64)), 0)
		if expirationTime.Before(time.Now()) {
			fmt.Println("Token has expired")
			return false
		}
		return true
	} else {
		fmt.Println("Invalid token")
		return false
	}
}

func generateJWT(userID, username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	sampleSecretKey := []byte("test")
	claims := token.Claims.(jwt.MapClaims)
	claims["ID"] = userID
	claims["username"] = username
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()
	tokenString, err := token.SignedString(sampleSecretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func parseJWT(tokenString string) (string, error) {
	sampleSecretKey := []byte("test")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unsupported signing method: %v", token.Header["alg"])
		}
		return sampleSecretKey, nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["ID"].(string)
		return userID, nil
	} else {
		return "", fmt.Errorf("invalid token")
	}
}

func main() {
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
	})

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
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

		token, err := generateJWT(fmt.Sprintf("%d", user.ID), user.Username)
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

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
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

		token, err := generateJWT(fmt.Sprintf("%d", dbID), user.Username)
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

	http.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		tmpl, _ := template.ParseFiles("test_service.html")
		tmpl.Execute(w, nil)
	})

	http.HandleFunc("/check-auth", func(w http.ResponseWriter, r *http.Request) {
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
		userID, err := parseJWT(tokenString)
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

		token, err := generateJWT(userID, "")
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
	log.Fatal(http.ListenAndServe(":8080", c.Handler(http.DefaultServeMux)))
}
