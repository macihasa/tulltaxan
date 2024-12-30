package handlers

import (
	"context"
	"fmt"
	"html"
	"io"
	"log"
	"log/slog"
	"net/http"
	"regexp"
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
		// Format descriptions into a list with levels
		descriptions := ""
		descriptions += fmt.Sprintf("<li>%s</li>", highlightText(result.Description, query))
		fmt.Fprintf(w, `
			<div class="result">
				<strong>HS Code:</strong> %s
				<ul>%s</ul>
			</div>
		`, result.Code, descriptions)
	}
}

// highlightText highlights all occurrences of the search term in the given text, case-insensitively.
func highlightText(text, term string) string {
	// Escape special HTML characters in the term to prevent HTML injection
	escapedTerm := html.EscapeString(term)

	// Compile a case-insensitive regex pattern for the term
	pattern := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(escapedTerm))

	// Replace all matches with the highlighted term
	highlighted := fmt.Sprintf(`<strong>%s</strong>`, escapedTerm)
	return pattern.ReplaceAllString(text, highlighted)
}

func IpHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("IP Endpoint hit..")
	response, err := http.Get(`https://ipinfo.io/ip`)
	if err != nil {
		slog.Error("unable to get ip adress", err)
		http.Error(w, "unable to get req for ip: "+err.Error(), http.StatusInternalServerError)
	}

	ip, err := io.ReadAll(response.Body)
	if err != nil {
		slog.Error("unable to read response body", err)
		http.Error(w, "unable to read response body: "+err.Error(), http.StatusInternalServerError)
	}

	w.Write(ip)

}
