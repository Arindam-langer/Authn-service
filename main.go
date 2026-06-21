package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Arindam-langer/governance-service/handlers"
	"github.com/Arindam-langer/governance-service/internal/db"
	"github.com/Arindam-langer/governance-service/middleware"
	"github.com/Arindam-langer/governance-service/routes"
	"github.com/joho/godotenv"
)

const (
	listenAddr   string        = "localhost:8080"
	ReadTimeout  time.Duration = 10 * time.Second
	WriteTimeout time.Duration = 10 * time.Second
)

func main() {
	_ = godotenv.Load()

	store, err := db.New(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("could not connect to db: %v", err)
	}
	defer store.Close()
	h := handlers.New(store)
	router := routes.Init(h)

	chain := middleware.LoggingMiddleware(middleware.UpdateHeader(router))
	s := &http.Server{
		Addr:           listenAddr,
		Handler:        chain,
		ReadTimeout:    ReadTimeout,
		WriteTimeout:   WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	fmt.Println("server listening on ", listenAddr)
	err = s.ListenAndServe()
	if err != nil {
		log.Fatalf("failed to run the server %v", err)
	}
}
