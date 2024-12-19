package xmltypes

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type DutyExpressions []DutyExpression

type DutyExpression struct {
	ChangeType                       string                     `xml:"changeType,attr"`
	DateEnd                          FileDistTime               `xml:"dateEnd,attr"`
	DateStart                        FileDistTime               `xml:"dateStart,attr"`
	DutyAmountApplicabilityCode      int                        `xml:"dutyAmountApplicabilityCode,attr"`
	DutyExpressionID                 string                     `xml:"dutyExpressionId,attr"`
	MeasurementUnitApplicabilityCode int                        `xml:"measurementUnitApplicabilityCode,attr"`
	MonetaryUnitApplicabilityCode    int                        `xml:"monetaryUnitApplicabilityCode,attr"`
	National                         int                        `xml:"national,attr"`
	DutyExpressionDescriptions       DutyExpressionDescriptions `xml:"dutyExpressionDescription"`
}

func (expressions DutyExpressions) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO duty_expression (duty_expression_id, change_type, date_start, date_end, duty_amount_applicability_code, measurement_unit_applicability_code, monetary_unit_applicability_code, national)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT (duty_expression_id) DO UPDATE 
	SET change_type = EXCLUDED.change_type,
		date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		duty_amount_applicability_code = EXCLUDED.duty_amount_applicability_code,
		measurement_unit_applicability_code = EXCLUDED.measurement_unit_applicability_code,
		monetary_unit_applicability_code = EXCLUDED.monetary_unit_applicability_code,
		national = EXCLUDED.national;
	`

	deleteQuery := `
	DELETE FROM duty_expression
	WHERE duty_expression_id = $1;
	`

	batch := &pgx.Batch{}

	for i, expr := range expressions {
		switch expr.ChangeType {
		case "U": // Insert or update
			batch.Queue(insertQuery, expr.DutyExpressionID, expr.ChangeType, expr.DateStart, expr.DateEnd, expr.DutyAmountApplicabilityCode, expr.MeasurementUnitApplicabilityCode, expr.MonetaryUnitApplicabilityCode, expr.National)

			// Queue child descriptions
			if len(expr.DutyExpressionDescriptions) > 0 {
				if err := expr.DutyExpressionDescriptions.QueueBatch(ctx, batch, expr.DutyExpressionID); err != nil {
					return fmt.Errorf("failed to queue descriptions for DutyExpressionID %s: %w", expr.DutyExpressionID, err)
				}
			}

		case "D": // Delete
			batch.Queue(deleteQuery, expr.DutyExpressionID)

		default:
			return fmt.Errorf("unknown ChangeType: %s for DutyExpressionID: %s", expr.ChangeType, expr.DutyExpressionID)
		}

		if (i+1)%batchSize == 0 || i == len(expressions)-1 {
			if err := conn.SendBatch(ctx, batch).Close(); err != nil {
				return fmt.Errorf("failed to execute batch: %w", err)
			}
			batch = &pgx.Batch{}
		}
	}

	return nil
}

type DutyExpressionDescriptions []DutyExpressionDescription

type DutyExpressionDescription struct {
	ParentDutyExpressionID string // added type
	Description            string `xml:"description,attr"`
	LanguageID             string `xml:"languageId,attr"`
	National               int    `xml:"national,attr"`
}

func (descriptions DutyExpressionDescriptions) QueueBatch(ctx context.Context, batch *pgx.Batch, parentID string) error {
	insertQuery := `
	INSERT INTO duty_expression_description (parent_duty_expression_id, description, language_id, national)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (parent_duty_expression_id, language_id) DO UPDATE 
	SET description = EXCLUDED.description,
		national = EXCLUDED.national;
	`

	for _, desc := range descriptions {
		desc.ParentDutyExpressionID = parentID // Ensure parent relationship
		batch.Queue(insertQuery, desc.ParentDutyExpressionID, desc.Description, desc.LanguageID, desc.National)
	}
	return nil
}
