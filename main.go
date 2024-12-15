package main

import (
	"context"
	"log"
	"os"
	"tulltaxan/pkg/filedist"

	"github.com/jackc/pgx/v5"
)

func main() {
	// Fetch database connection string from environment variables
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	log.Println("Starting the application...")

	// Connect to the database
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer conn.Close(ctx)

	log.Println("Successfully connected to the database.")

	// Verify the connection with a simple query
	var result string
	err = conn.QueryRow(ctx, "SELECT 'Connected tbo PostgreSQL!'").Scan(&result)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	filedist.StartDbMaintenanceScheduler()

}
