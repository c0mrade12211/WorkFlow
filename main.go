package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

type Task struct {
	ID       string `json:"id"`
	User     string `json:"user"`
	Title    string `json:"title"`
	Complete bool   `json:"complete"`
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Session struct {
	ID        string
	ExpiresAt time.Time
}

var sessions = make(map[string]Session)

func generateClientSecret(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(bytes)
	secret := strings.TrimRight(encoded, "=")
	return secret, nil
}

func createSession(username string) (string, error) {
	sessionID, err := generateClientSecret(32)
	if err != nil {
		return "", err
	}
	expiresAt := time.Now().Add(24 * time.Hour)
	session := Session{
		ID:        sessionID,
		ExpiresAt: expiresAt,
	}
	sessions[sessionID] = session
	return sessionID, nil
}

func getSession(sessionID string) (Session, bool) {
	session, ok := sessions[sessionID]
	return session, ok
}

func main() {
	db, err := leveldb.OpenFile("users.db", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
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

		err = db.Put([]byte(user.Username), []byte(user.Password), nil)
		if err != nil {
			http.Error(w, "Failed to add user", http.StatusInternalServerError)
			return
		}
		length := 32
		clientSecret, err := generateClientSecret(length)
		clientID, err := rand.Int(rand.Reader, big.NewInt(10000000))
		if err != nil {
			fmt.Println("Failed to generate client secret:", err)
			return
		}
		fmt.Println("Generated client secret:", clientSecret)
		fmt.Println("Generated client ID:", clientID)
		fmt.Fprintf(w, "User added successfully")
	})

	http.HandleFunc("/show-users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		iter := db.NewIterator(nil, nil)
		defer iter.Release()
		for iter.Next() {
			fmt.Fprintf(w, "%s\n", iter.Key())
			fmt.Fprintf(w, "%s\n", iter.Value())
		}
		if err := iter.Error(); err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(w, "All users")
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
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

		value, err := db.Get([]byte(user.Username), nil)
		if err != nil {
			http.Error(w, "User does not exist", http.StatusForbidden)
			return
		}

		if string(value) != user.Password {
			http.Error(w, "Invalid password", http.StatusForbidden)
			return
		}

		sessionID, err := createSession(user.Username)
		if err != nil {
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}

		cookie := http.Cookie{
			Name:     "session",
			Value:    sessionID,
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		}
		http.SetCookie(w, &cookie)

		fmt.Fprintf(w, "Logged in successfully")
	})

	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, "Session not found", http.StatusUnauthorized)
			return
		}

		sessionID := cookie.Value
		_, ok := getSession(sessionID)
		if !ok {
			http.Error(w, "Invalid session", http.StatusUnauthorized)
			return
		}

		delete(sessions, sessionID)

		cookie = &http.Cookie{
			Name:    "session",
			Value:   "",
			Expires: time.Now(),
		}
		http.SetCookie(w, cookie)

		fmt.Fprintf(w, "Logged out successfully")
	})

	http.HandleFunc("/protected", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, "Session not found", http.StatusUnauthorized)
			return
		}

		sessionID := cookie.Value
		session, ok := getSession(sessionID)
		if !ok || session.ExpiresAt.Before(time.Now()) {
			http.Error(w, "Invalid session", http.StatusUnauthorized)
			return
		}

		fmt.Fprintf(w, "Protected content")
	})
	http.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		tmpl, _ := template.ParseFiles("test_service.html")
		tmpl.Execute(w, nil)
	})

	log.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
