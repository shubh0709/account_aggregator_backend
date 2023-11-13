package main

import (
	"log"
	"net/http"

	"valyx/aggregator/utils"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"github.com/ztrue/tracerr"
)

func init() {
	viper.SetConfigFile(".env") // Set the path of your .env file
	viper.ReadInConfig()
	viper.SetDefault("PORT", "8080")
	viper.AutomaticEnv()

}

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
	http.HandleFunc("/env", server.TestEnvironmentHandler)

	serverPort := viper.GetString("PORT")
	log.Println("Starting server on " + serverPort)

	runServer := &http.Server{
		Addr:    "0.0.0.0:" + serverPort,
		Handler: utils.ApplyMiddleware(http.DefaultServeMux, utils.EnableCORS(), utils.LoggingMiddleware),
	}

	if err := runServer.ListenAndServe(); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
