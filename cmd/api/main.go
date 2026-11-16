package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/chrisamar/cea-go-ar-exception/internal/httpapi"
	"github.com/chrisamar/cea-go-ar-exception/internal/report"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("load .env: %v", err)
	}

	dsn, err := oracleDSNFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("oracle", dsn)
	if err != nil {
		log.Fatalf("open Oracle connection: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("connect to Oracle: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8012"
	}

	server := &http.Server{
		Addr:              "127.0.0.1:" + port,
		Handler:           httpapi.New(report.NewOracleRunner(db)),
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("API listening on http://%s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
