package xmltypes

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type MeursingSubheadings []MeursingSubheading

type MeursingSubheading struct {
	ChangeType               string       `xml:"changeType,attr"`
	DateStart                FileDistTime `xml:"dateStart,attr"`
	Description              string       `xml:"description,attr"`
	HeadingNumber            int          `xml:"headingNumber,attr"`
	MeursingTablePlanID      int          `xml:"meursingTablePlanId,attr"`
	National                 int          `xml:"national,attr"`
	RowColumnCode            int          `xml:"rowColumnCode,attr"`
	SubheadingSequenceNumber int          `xml:"subheadingSequenceNumber,attr"`
}

func (subheadings MeursingSubheadings) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO meursing_subheading (
		heading_number, meursing_table_plan_id, row_column_code, subheading_sequence_number,
		date_start, description, national, change_type
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT (heading_number, meursing_table_plan_id, row_column_code, subheading_sequence_number) DO UPDATE 
	SET date_start = EXCLUDED.date_start,
		description = EXCLUDED.description,
		national = EXCLUDED.national,
		change_type = EXCLUDED.change_type;
	`

	deleteQuery := `
	DELETE FROM meursing_subheading
	WHERE heading_number = $1 AND meursing_table_plan_id = $2 AND row_column_code = $3 AND subheading_sequence_number = $4;
	`

	batch := &pgx.Batch{}

	for i, subheading := range subheadings {
		switch subheading.ChangeType {
		case "U": // Insert or update
			batch.Queue(insertQuery, subheading.HeadingNumber, subheading.MeursingTablePlanID, subheading.RowColumnCode,
				subheading.SubheadingSequenceNumber, subheading.DateStart, subheading.Description, subheading.National, subheading.ChangeType)

		case "D": // Delete
			batch.Queue(deleteQuery, subheading.HeadingNumber, subheading.MeursingTablePlanID, subheading.RowColumnCode, subheading.SubheadingSequenceNumber)

		default:
			return fmt.Errorf("unknown ChangeType: %s for HeadingNumber: %d, SubheadingSequenceNumber: %d", subheading.ChangeType, subheading.HeadingNumber, subheading.SubheadingSequenceNumber)
		}

		if (i+1)%batchSize == 0 || i == len(subheadings)-1 {
			if err := conn.SendBatch(ctx, batch).Close(); err != nil {
				return fmt.Errorf("failed to execute batch: %w", err)
			}
			batch = &pgx.Batch{}
		}
	}

	return nil
}
