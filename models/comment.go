package models

type Comments struct {
	ID        int    `json:"id"`
	Comment   string `json:"comment"`
	User      User   `json:"user"`
	CreatedAt string `json:"createdAt"`
}
