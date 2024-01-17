package models

type Task struct {
	Creator     string `json:"creator"`
	Description string `json:"description"`
	Title       string `json:"title"`
	ID          string `json:"id"`
	IsComplete  string `json:"iscomplete"`
	CreatedAt   string `json:"created_at"`
	UserID      string `json:"userid"`
}
