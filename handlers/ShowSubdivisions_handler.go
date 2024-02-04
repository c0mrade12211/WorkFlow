package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	jwt_service "todo/JWT"
)

func ShowSubdivisions(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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
	fmt.Println(userID + " Request for all subdivisions")

	checkReq, err := db.Query("SELECT subdivision_id FROM join_sub WHERE user_id = $1", userID)
	if err != nil {
		http.Error(w, "Failed to query subdivisions", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	defer checkReq.Close()

	isInvited := false
	if checkReq.Next() {
		isInvited = true
	}

	rows, err := db.Query("SELECT name, subdivision_id, owner FROM subdivisions")
	if err != nil {
		http.Error(w, "Failed to get subdivisions", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	subdivisions := []map[string]interface{}{}
	for rows.Next() {
		var subdivisionName, subdivisionID, subdivisionOwner interface{}
		err := rows.Scan(&subdivisionName, &subdivisionID, &subdivisionOwner)
		if err != nil {
			http.Error(w, "Failed to scan subdivision", http.StatusInternalServerError)
			return
		}

		subdivision := map[string]interface{}{
			"name":           subdivisionName,
			"subdivision_id": subdivisionID,
			"owner":          subdivisionOwner,
			"isinvited":      isInvited,
		}

		subdivisions = append(subdivisions, subdivision)
	}

	jsonResp, err := json.Marshal(subdivisions)
	if err != nil {
		http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}
