package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"valyx/aggregator/utils"

	_ "github.com/lib/pq"
	"github.com/ztrue/tracerr"
)

var DB *sql.DB

func main() {
	// Open a connection to the database.
	db, err := setupDB()
	if err != nil {
		tracerr.Wrap(err)
		tracerr.PrintSourceColor(err)
		panic(err)
	}
	defer db.Close()
	DB = db

	_, err = db.Exec("SELECT 1 FROM users LIMIT 1")
	if err != nil {
		tracerr.Wrap(err)
		tracerr.PrintSourceColor(err)
		// Create the users table if it doesn't exist.
		_, err = db.Exec("CREATE TABLE users (id serial PRIMARY KEY, username text NOT NULL, password text NOT NULL)")
		if err != nil {
			tracerr.Wrap(err)
			tracerr.PrintSourceColor(err)
			panic(err)
		}
	}

	// Check the connection
	err = db.Ping()
	if err != nil {
		tracerr.Wrap(err)
		tracerr.PrintSourceColor(err)
		log.Fatal(err)
	}

	utils.ReadExcelFiles(db)

	// Wrap the protectedHandler with authMiddleware
	protectedHandlerWithAuth := authMiddleware(http.HandlerFunc(protectedHandler))

	// Register the wrapped handler to the protected endpoint
	http.Handle("/protected", protectedHandlerWithAuth)

	http.Handle("/user", http.HandlerFunc(createUserHandler))
	http.HandleFunc("/", healthCheckHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))

	fmt.Println("User created successfully!")
}
