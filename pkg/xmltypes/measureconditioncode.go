package xmltypes

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type MeasureConditionCodes []MeasureConditionCode

type MeasureConditionCode struct {
	ChangeType                       string                           `xml:"changeType,attr"`
	ConditionCode                    string                           `xml:"conditionCode,attr"`
	DateEnd                          FileDistTime                     `xml:"dateEnd,attr"`
	DateStart                        FileDistTime                     `xml:"dateStart,attr"`
	National                         int                              `xml:"national,attr"`
	Type                             *int                             `xml:"type,attr"`
	MeasureConditionCodeDescriptions MeasureConditionCodeDescriptions `xml:"measureConditionCodeDescription"`
}

func (codes MeasureConditionCodes) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO measure_condition_code (condition_code, change_type, date_start, date_end, type, national)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (condition_code) DO UPDATE 
	SET change_type = EXCLUDED.change_type,
		date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		type = EXCLUDED.type,
		national = EXCLUDED.national;
	`

	deleteQuery := `
	DELETE FROM measure_condition_code
	WHERE condition_code = $1;
	`

	batch := &pgx.Batch{}

	for i, code := range codes {
		switch code.ChangeType {
		case "U": // Insert or update
			batch.Queue(insertQuery, code.ConditionCode, code.ChangeType, code.DateStart, code.DateEnd, code.Type, code.National)

			// Queue child descriptions
			if len(code.MeasureConditionCodeDescriptions) > 0 {
				if err := code.MeasureConditionCodeDescriptions.QueueBatch(ctx, batch, code.ConditionCode); err != nil {
					return fmt.Errorf("failed to queue descriptions for ConditionCode %s: %w", code.ConditionCode, err)
				}
			}

		case "D": // Delete
			batch.Queue(deleteQuery, code.ConditionCode)

		default:
			return fmt.Errorf("unknown ChangeType: %s for ConditionCode: %s", code.ChangeType, code.ConditionCode)
		}

		if (i+1)%batchSize == 0 || i == len(codes)-1 {
			if err := conn.SendBatch(ctx, batch).Close(); err != nil {
				return fmt.Errorf("failed to execute batch: %w", err)
			}
			batch = &pgx.Batch{}
		}
	}

	return nil
}

type MeasureConditionCodeDescriptions []MeasureConditionCodeDescription

type MeasureConditionCodeDescription struct {
	ParentConditionCode string // added type
	Description         string `xml:"description,attr"`
	LanguageID          string `xml:"languageId,attr"`
	National            int    `xml:"national,attr"`
}

func (descriptions MeasureConditionCodeDescriptions) QueueBatch(ctx context.Context, batch *pgx.Batch, parentConditionCode string) error {
	insertQuery := `
	INSERT INTO measure_condition_code_description (parent_condition_code, description, language_id, national)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (parent_condition_code, language_id) DO UPDATE 
	SET description = EXCLUDED.description,
		national = EXCLUDED.national;
	`

	for _, desc := range descriptions {
		desc.ParentConditionCode = parentConditionCode
		batch.Queue(insertQuery, desc.ParentConditionCode, desc.Description, desc.LanguageID, desc.National)
	}
	return nil
}
