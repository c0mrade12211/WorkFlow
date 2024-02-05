package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"todo/handlers"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

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

func main() {
	connStr := "user=postgres password=test dbname=myapp sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	r := mux.NewRouter()

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://87aa-87-244-58-26.ngrok-free.app"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS", "DELETE"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	r.HandleFunc("/register", handlers.WithDB(handlers.RegisterHandler, db)).Methods("POST")
	r.HandleFunc("/delete-req/{user_id}", handlers.WithDB(handlers.DeleteReq, db)).Methods("GET")
	r.HandleFunc("/my-items", handlers.WithDB(handlers.MyItems, db)).Methods("GET")
	r.HandleFunc("/accept-user", handlers.WithDB(handlers.AcceptUser, db)).Methods("POST")
	r.HandleFunc("/request-invite/{subdiv_id}", handlers.WithDB(handlers.RequestForInvite, db)).Methods("GET")
	r.HandleFunc("/use-my-item/{id}", handlers.WithDB(handlers.UseMyItem, db)).Methods("DELETE")
	r.HandleFunc("/all-subdivisions", handlers.WithDB(handlers.ShowSubdivisions, db)).Methods("GET")
	r.HandleFunc("/login", handlers.WithDB(handlers.LoginHandler, db)).Methods("POST")
	r.HandleFunc("/get-me", handlers.WithDB(handlers.GetProfileHandler, db)).Methods("GET")
	r.HandleFunc("/create-task", handlers.WithDB(handlers.CreateTaskHandler, db)).Methods("POST")
	r.HandleFunc("/delete-task/{id}", handlers.WithDB(handlers.DeleteTaskHandler, db)).Methods("DELETE")
	r.HandleFunc("/check-auth", handlers.WithDB(handlers.CheckAuthHandler, db)).Methods("GET")
	r.HandleFunc("/shop", handlers.WithDB(handlers.ShopHandler, db)).Methods("GET")
	r.HandleFunc("/buy/{id}", handlers.WithDB(handlers.BuyHandler, db)).Methods("GET")
	r.HandleFunc("/create-item", handlers.WithDB(handlers.CreateItem, db)).Methods("POST")
	r.HandleFunc("/my-subdivision", handlers.WithDB(handlers.GetMySubdivision, db)).Methods("GET")
	r.HandleFunc("/change-task/{id}", handlers.WithDB(handlers.ChangeStatus, db)).Methods("GET")
	r.HandleFunc("/my-tasks", handlers.WithDB(handlers.MyTasksHandler, db)).Methods("GET")
	r.HandleFunc("/invited-list", handlers.WithDB(handlers.InvitedList, db)).Methods("GET")

	log.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", c.Handler(r)))
}
