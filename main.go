package main

import (
	"log"
	"net/http"

	"valyx/aggregator/utils"

	_ "github.com/lib/pq"
	"github.com/ztrue/tracerr"
)

func main() {
	db, err := setupDB()
	if err != nil {
		tracerr.Wrap(err)
		tracerr.PrintSourceColor(err)
		panic(err)
	}
	defer db.Close()

	fileProcessor := utils.NewProcessor(db)
	err = fileProcessor.ReadExcelFiles("./dummyData", db)
	if err != nil {
		log.Fatalf("could not process files: %v", err)
	}

	queryService := NewService(db)

	server := NewServer(queryService)
	http.HandleFunc("/search", server.SearchHandler)
	http.HandleFunc("/userInfo", server.GetUserInfo)
	http.HandleFunc("/trend", server.TrendHandler)
	http.HandleFunc("/aggregate", server.AggregateHandler)
	log.Println("Starting server on :8080")

	runServer := &http.Server{
		Addr: ":8080",
		Handler: utils.ApplyMiddleware(http.DefaultServeMux, utils.EnableCORS, utils.LoggingMiddleware),
	}

	if err := runServer.ListenAndServe(); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
