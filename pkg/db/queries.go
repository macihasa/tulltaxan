package db

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
)

// HSCode represents a result from the HS code search
type HSCode struct {
	Code         string
	Descriptions []struct {
		Description string
		LanguageID  string
	}
}

// SearchHSCodes queries the materialized view for matching HS codes
func SearchHSCodes(ctx context.Context, conn *pgx.Conn, query string) ([]HSCode, error) {
	if query == "" {
		return nil, errors.New("query string cannot be empty")
	}

	// Query the materialized view
	rows, err := conn.Query(ctx, `
		SELECT
			hs_code,
			descriptions
		FROM
			mv_goods_nomenclature_search
		WHERE
			to_tsvector('english', hs_code || ' ' || flattened_descriptions) @@ plainto_tsquery($1)
		LIMIT 50;
	`, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Parse the results into a slice of HSCode
	var results []HSCode
	for rows.Next() {
		var hsCode HSCode
		var descriptions []byte // Descriptions are stored as JSONB in the database
		if err := rows.Scan(&hsCode.Code, &descriptions); err != nil {
			return nil, err
		}

		// Decode the JSONB descriptions into Go structs
		if err := json.Unmarshal(descriptions, &hsCode.Descriptions); err != nil {
			return nil, err
		}

		results = append(results, hsCode)
	}

	return results, nil
}
