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
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	log.Printf("Received query: %s", query)

	// Fetch the results from the database
	results, err := db.SearchHSCodes(context.Background(), conn, query)
	if err != nil {
		http.Error(w, "Database query failed", http.StatusInternalServerError)
		log.Printf("Database error: %v", err)
		return
	}

	// Respond with HTML content
	w.Header().Set("Content-Type", "text/html")
	for _, result := range results {
		certificatesHTML := ""
		for _, cert := range result.MeasureComponents.Certificates {
			certificatesHTML += fmt.Sprintf("<li>%s</li>", html.EscapeString(string(cert)))
		}

		additionalCodesHTML := ""
		for _, code := range result.MeasureComponents.AdditionalCodes {
			additionalCodesHTML += fmt.Sprintf("<li>%s</li>", html.EscapeString(string(code)))
		}

		// Construct the result HTML with nested lists
		fmt.Fprintf(w, `
            <div class="result">
                <strong>HS Code:</strong> %s
                <ul>
                    <li><strong>Description:</strong> %s</li>
                    <li><strong>Certificates:</strong>
                        <ul>%s</ul>
                    </li>
                    <li><strong>Additional Codes:</strong>
                        <ul>%s</ul>
                    </li>
                </ul>
            </div>
        `,
			html.EscapeString(result.Code),
			highlightText(html.EscapeString(result.Description), query),
			certificatesHTML,
			additionalCodesHTML,
		)
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
