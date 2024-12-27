package main

import (
	"context"
	"fmt"
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

	// Constants
	const ddlFile = "./sql/ddl.sql"
	const ddlViewsFile = "./sql/ddl_views.sql"

	// Connect to the database
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer conn.Close(ctx)

	if err := insertSQLFiles([]string{ddlFile, ddlViewsFile}, ctx, conn); err != nil {
		log.Fatal("insertSQLFiles: ", err)
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

func insertSQLFiles(paths []string, ctx context.Context, conn *pgx.Conn) error {
	for _, path := range paths {
		query, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("unable to read ddl file %s: %w", path, err)
		}
		_, err = conn.Exec(ctx, string(query))
		if err != nil {
			return fmt.Errorf("unable to execute query for file %s: %w", path, err)
		}
	}

	return nil
}
