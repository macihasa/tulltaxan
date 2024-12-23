package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"tulltaxan/pkg/filedist"
	"tulltaxan/pkg/handlers"

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
	_, err = conn.Exec(ctx, string(ddlQuery))
	if err != nil {
		log.Fatal("unable to execute ddl query: ", err)
	}

	filedist.StartDbMaintenanceScheduler(conn)

	staticDir := "./static"
	fs := http.FileServer(http.Dir(staticDir))
	http.Handle("/", fs)
	port := "8080"

	http.Handle("/search", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.SearchHandler(w, r, conn)
	}))

	log.Printf("server listening on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("server crashed: ", err)
	}

}
