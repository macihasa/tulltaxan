package db

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

// HSCode represents a result from the HS code search
type HSCode struct {
	Code              string            `json:"code"`
	Description       string            `json:"description"`
	MeasureComponents MeasureComponents `json:"measure_components"`
}

type MeasureComponents struct {
	Certificates    []Certificate    `json:"certificates"`
	AdditionalCodes []AdditionalCode `json:"additional_codes"`
}

type Certificate string
type AdditionalCode string

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

	for i, hsCode := range results {

		comp, err := getTaricComposition(hsCode.Code)
		if err != nil {
			return nil, fmt.Errorf("getTaricComposition: %w", err)
		}

		components, err := SearchMeasureComponents(comp.HSCode, "RU", ctx, conn)
		if err != nil {
			return nil, fmt.Errorf("SearchMeasureComponents: %w", err)
		}

		if components != nil {
			results[i].MeasureComponents = *components
		}

	}

	return results, nil
}

func SearchMeasureComponents(hs string, dstCtry string, ctx context.Context, conn *pgx.Conn) (*MeasureComponents, error) {
	// Check validity of input
	if hs == "" || dstCtry == "" || len(dstCtry) != 2 {
		return nil, nil
	}

	composition, err := getTaricComposition(hs)
	if err != nil {
		return nil, fmt.Errorf("getTaricComposition: %w", err)
	}

	rows, err := conn.Query(ctx, `
	SELECT (mc.certificate_type || mc.certificate_code) AS y_code,
		(m.additional_code_type || m.additional_code_id) AS additional_code
	FROM measure m
		LEFT JOIN measure_type mt ON m.measure_type = mt.measure_type
		LEFT JOIN measure_type_description mtd ON mt.measure_type = mtd.parent_measure_type
		LEFT JOIN measure_condition mc ON m.sid = mc.parent_sid
		LEFT JOIN measure_action ma ON mc.action_code = ma.action_code
		LEFT JOIN base_regulation br ON m.regulation_id = br.regulation_id
		LEFT JOIN modification_regulation mr ON m.regulation_id = mr.modification_regulation_id
		LEFT JOIN certificate c ON mc.certificate_code = c.certificate_code
		LEFT JOIN additional_code ac ON m.additional_code_id = ac.additional_code_id
	WHERE m.goods_nomenclature_code IN (
			$1,
			$2,
			$3,
			$4
		)
		AND mt.trade_movement_code IN ('1', '2')
		AND (
			m.geographical_area_id = $5
			OR m.geographical_area_id IN (
				SELECT ga2.geographical_area_id
				FROM geographical_area ga1
					JOIN geographical_area_membership gam ON ga1.sid = gam.parent_sid
					JOIN geographical_area ga2 ON gam.sid_geographical_area_group = ga2.sid
				WHERE ga1.geographical_area_id = $5
			)
		)
		AND (
			mc.certificate_type = 'Y'
			OR m.additional_code_type != ''
		)
		AND (
			m.date_end IS NULL
			OR m.date_end > CURRENT_TIMESTAMP
		)
		AND (m.date_start < CURRENT_TIMESTAMP)
		AND (
			c.date_end IS NULL
			OR c.date_end > CURRENT_TIMESTAMP
		)
		AND (c.date_start < CURRENT_TIMESTAMP)
		AND (
			br.date_end IS NULL
			OR br.date_end > CURRENT_TIMESTAMP
		)
		AND (
			mr.date_end IS NULL
			OR mr.date_end > CURRENT_TIMESTAMP
		)
	GROUP BY y_code,
		additional_code`, composition.Chapter, composition.HSCode, composition.HSUnderNumber, composition.CNCode, dstCtry)

	if err != nil {
		return nil, fmt.Errorf("failed to query y-codes and add codes: %w", err)
	}

	components := new(MeasureComponents)

	for rows.Next() {
		certificate := new(Certificate)
		additionalCode := new(AdditionalCode)
		if err := rows.Scan(&certificate, &additionalCode); err != nil {
			if err != nil {
				return nil, fmt.Errorf("failed to scan rows into variables: %w", err)
			}
		}

		if certificate != nil {
			components.Certificates = append(components.Certificates, *certificate)
		}
		if additionalCode != nil {
			components.AdditionalCodes = append(components.AdditionalCodes, *additionalCode)
		}
	}

	return components, nil
}

type TaricComposition struct {
	Chapter       string `json:"chapter"`
	HSCode        string `json:"hs_code"`
	HSUnderNumber string `json:"hs_under_number"`
	CNCode        string `json:"cn_code"`
	Taric         string `json:"taric"`
}

// Requires 10 char hs code
func getTaricComposition(taric string) (*TaricComposition, error) {
	if len(taric) != 10 {
		return nil, fmt.Errorf("invalid length of taric: code[%s] length[%v]. expected 10", taric, len(taric))
	}
	composition := &TaricComposition{
		Chapter:       taric[:2] + "00000000",
		HSCode:        taric[:4] + "000000",
		HSUnderNumber: taric[:6] + "0000",
		CNCode:        taric[:8] + "00",
		Taric:         taric,
	}

	return composition, nil
}

// Preprocess the query to handle phrases
func preprocessQuery(input string) string {
	input = strings.TrimPrefix(input, " ")
	input = strings.TrimSuffix(input, " ")
	return strings.ReplaceAll(input, " ", " <-> ")
}
