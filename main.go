package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"

	_ "dating_app/docs"

	"dating_app/api"

	_ "github.com/lib/pq"
)

// @title Dating App API
// @description This is a sample dating app API.
// @version 1.0
// @host localhost:8080
// @BasePath /
func main() {
	var err error
	var db  *sql.DB
	db, err = sql.Open("postgres", "user=root password=123123123 dbname=dating_app sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Setup HTTP routes
	api.Routes(db)

	// Start the HTTP server
	serverAddr := "localhost:8080"
	go func() {
		log.Printf("Server is starting and listening on %s", serverAddr)
		if err := http.ListenAndServe(serverAddr, nil); err != nil {
			log.Fatalf("Error starting server: %s", err)
		}
	}()

	// Wait for server shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	log.Println("Server stopped gracefully")
}
