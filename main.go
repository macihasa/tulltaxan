package main

import (
	"context"
	"fmt"
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

	ddlQuery, err := os.ReadFile(`./sql/ddl.sql`)
	if err != nil {
		log.Fatal("unable to read ddl file content: ", err)
	}

	fmt.Println(string(ddlQuery))

	_, err = conn.Exec(ctx, string(ddlQuery))
	if err != nil {
		log.Fatal("unable to execute ddl query: ", err)
	}

	filedist.StartDbMaintenanceScheduler()

}
