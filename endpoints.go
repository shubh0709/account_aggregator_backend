package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/lib/pq"
)

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the JSON data
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert the user ID and password into the table
	_, err = DB.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", user.UserID, user.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate a JWT token
	token, err := generateToken(user.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	// Return a success response
	w.WriteHeader(http.StatusOK)
	// Return the JWT token
	fmt.Fprintf(w, token)
	fmt.Fprintf(w, "User ID and password stored successfully")
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Return a success response
	w.WriteHeader(http.StatusOK)
}

func protectedHandler(w http.ResponseWriter, r *http.Request) {
	// Access the user ID from the request context
	userID := r.Context().Value("UserID").(string)

	// Perform actions specific to the protected endpoint
	// For example, fetch user data from the database based on the user ID

	// Return the response
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Protected endpoint accessed by user: %s", userID)
}
