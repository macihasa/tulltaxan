package xmltypes

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type MeursingAdditionalCodes []MeursingAdditionalCode

type MeursingAdditionalCode struct {
	SID                         int                         `xml:"SID,attr"`
	AdditionalCodeID            string                      `xml:"additionalCodeId,attr"`
	DateStart                   FileDistTime                `xml:"dateStart,attr"`
	DateEnd                     FileDistTime                `xml:"dateEnd,attr"`
	National                    int                         `xml:"national,attr"`
	ChangeType                  string                      `xml:"changeType,attr"`
	MeursingTableCellComponents MeursingTableCellComponents `xml:"meursingTableCellComponent"`
}

func (codes MeursingAdditionalCodes) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO meursing_additional_code (
		meursing_table_plan_id, additional_code_id, date_start, date_end, national, change_type
	)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (meursing_table_plan_id) DO UPDATE 
	SET additional_code_id = EXCLUDED.additional_code_id,
		date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		national = EXCLUDED.national,
		change_type = EXCLUDED.change_type;
	`

	deleteQuery := `
	DELETE FROM meursing_additional_code
	WHERE meursing_table_plan_id = $1;
	`

	batch := &pgx.Batch{}

	for i, code := range codes {
		switch code.ChangeType {
		case "U": // Insert or update
			batch.Queue(insertQuery, code.SID, code.AdditionalCodeID, code.DateStart, nil, code.National, code.ChangeType)

			// Queue child components
			if len(code.MeursingTableCellComponents) > 0 {
				if err := code.MeursingTableCellComponents.QueueBatch(ctx, batch, code.SID); err != nil {
					return fmt.Errorf("failed to queue cell components for SID %d: %w", code.SID, err)
				}
			}

		case "D": // Delete
			batch.Queue(deleteQuery, code.SID)

		default:
			return fmt.Errorf("unknown ChangeType: %s for SID: %d", code.ChangeType, code.SID)
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

type MeursingTableCellComponents []MeursingTableCellComponent

type MeursingTableCellComponent struct {
	MeursingTablePlanID      int          `xml:"meursingTablePlanId,attr"`
	HeadingNumber            int          `xml:"headingNumber,attr"`
	RowColumnCode            int          `xml:"rowColumnCode,attr"`
	SubheadingSequenceNumber int          `xml:"subheadingSequenceNumber,attr"`
	DateStart                FileDistTime `xml:"dateStart,attr"`
	National                 int          `xml:"national,attr"`
}

func (components MeursingTableCellComponents) QueueBatch(ctx context.Context, batch *pgx.Batch, parentTablePlanID int) error {
	insertQuery := `
	INSERT INTO meursing_table_cell_component (
		parent_table_plan_id, heading_number, row_column_code, subheading_sequence_number, date_start, national
	)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (parent_table_plan_id, heading_number, row_column_code, subheading_sequence_number) DO UPDATE 
	SET date_start = EXCLUDED.date_start,
		national = EXCLUDED.national;
	`

	for _, component := range components {
		component.MeursingTablePlanID = parentTablePlanID
		batch.Queue(insertQuery, component.MeursingTablePlanID, component.HeadingNumber, component.RowColumnCode, component.SubheadingSequenceNumber, component.DateStart, component.National)
	}

	return nil
}
