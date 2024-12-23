package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"tulltaxan/pkg/db"

	"github.com/jackc/pgx/v5"
)

// SearchHandler processes HTMX search requests
func SearchHandler(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {

	// Extract the 'q' query parameter
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	// Log the query for debugging
	log.Printf("Received query: %s", query)

	// Query the database using the updated db package function
	results, err := db.SearchHSCodes(context.Background(), conn, query)
	if err != nil {
		http.Error(w, "Database query failed", http.StatusInternalServerError)
		log.Printf("Database error: %v", err)
		return
	}

	// Write results as HTML for HTMX
	w.Header().Set("Content-Type", "text/html")
	for _, result := range results {
		// Format descriptions into a list
		descriptions := ""
		for _, desc := range result.Descriptions {
			descriptions += fmt.Sprintf("<li>%s (%s)</li>", desc.Description, desc.LanguageID)
		}
		fmt.Fprintf(w, `
			<div class="result">
				<strong>HS Code:</strong> %s
				<ul>%s</ul>
			</div>
		`, result.Code, descriptions)
	}
}
