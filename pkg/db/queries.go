package db

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
)

// HSCode represents a result from the HS code search
type HSCode struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// SearchHSCodes queries the materialized view for matching HS codes
func SearchHSCodes(ctx context.Context, conn *pgx.Conn, query string) ([]HSCode, error) {
	if query == "" {
		return nil, errors.New("query string cannot be empty")
	}
	// Query the materialized view
	processedQuery := preprocessQuery(query)
	rows, err := conn.Query(ctx, `
		SELECT
			cn_code,
			descriptions
		FROM
			mv_goods_nomenclature_search
		WHERE
			search_vector @@ to_tsquery('swedish', $1)
		LIMIT 20;
	`, processedQuery)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Parse the results into a slice of HSCode
	var results []HSCode
	for rows.Next() {
		var hsCode HSCode
		if err := rows.Scan(&hsCode.Code, &hsCode.Description); err != nil {
			return nil, err
		}

		results = append(results, hsCode)
	}

	return results, nil
}

// Preprocess the query to handle phrases
func preprocessQuery(input string) string {
	input = strings.TrimPrefix(input, " ")
	input = strings.TrimSuffix(input, " ")
	return strings.ReplaceAll(input, " ", " <-> ")
}
