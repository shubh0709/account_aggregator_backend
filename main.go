package main

import (
	"log"
	"net/http"

	"valyx/aggregator/utils"

	_ "github.com/lib/pq"
	"github.com/ztrue/tracerr"
)

func main() {
	// Open a connection to the database.
	db, err := setupDB()
	if err != nil {
		tracerr.Wrap(err)
		tracerr.PrintSourceColor(err)
		panic(err)
	}
	defer db.Close()

	// Processor setup to read and store CSV data.
	fileProcessor := utils.NewProcessor(db)
	err = fileProcessor.ReadExcelFiles("./dummyData", db)
	if err != nil {
		log.Fatalf("could not process files: %v", err)
	}

	// Create the query service.
	queryService := NewService(db)

	// Set up the HTTP server.
	server := NewServer(queryService)
	http.HandleFunc("/search", server.SearchHandler)
	http.HandleFunc("/userInfo", server.GetUserInfo)
	http.HandleFunc("/trend", server.TrendHandler)
	http.HandleFunc("/aggregate", server.AggregateHandler)
	// Start the server.
	log.Println("Starting server on :8080")

	// Create your server instance
	runServer := &http.Server{
		Addr: ":8080",
		// Use applyMiddleware to wrap your handlers with the desired middleware
		Handler: utils.ApplyMiddleware(http.DefaultServeMux, utils.EnableCORS, utils.LoggingMiddleware),
	}

	if err := runServer.ListenAndServe(); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
