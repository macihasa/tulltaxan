package xmltypes

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type MeursingHeadings []MeursingHeading

type MeursingHeading struct {
	HeadingNumber                       int                                 `xml:"headingNumber,attr"`
	MeursingTablePlanID                 int                                 `xml:"meursingTablePlanId,attr"`
	RowColumnCode                       int                                 `xml:"rowColumnCode,attr"`
	DateStart                           FileDistTime                        `xml:"dateStart,attr"`
	National                            int                                 `xml:"national,attr"`
	ChangeType                          string                              `xml:"changeType,attr"`
	MeursingHeadingFootnoteAssociations MeursingHeadingFootnoteAssociations `xml:"meursingHeadingFootnoteAssociation"`
	MeursingHeadingText                 MeursingHeadingTexts                `xml:"meursingHeadingText"`
}

func (headings MeursingHeadings) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO meursing_heading (
		heading_number, meursing_table_plan_id, row_column_code, date_start, national, change_type
	)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (heading_number, meursing_table_plan_id, row_column_code) DO UPDATE 
	SET date_start = EXCLUDED.date_start,
		national = EXCLUDED.national,
		change_type = EXCLUDED.change_type;
	`

	deleteQuery := `
	DELETE FROM meursing_heading
	WHERE heading_number = $1 AND meursing_table_plan_id = $2 AND row_column_code = $3;
	`

	batch := &pgx.Batch{}

	for i, heading := range headings {
		switch heading.ChangeType {
		case "U": // Insert or update
			batch.Queue(insertQuery, heading.HeadingNumber, heading.MeursingTablePlanID, heading.RowColumnCode, heading.DateStart, heading.National, heading.ChangeType)

			// Queue child footnote associations
			if len(heading.MeursingHeadingFootnoteAssociations) > 0 {
				if err := heading.MeursingHeadingFootnoteAssociations.QueueBatch(ctx, batch, heading.HeadingNumber); err != nil {
					return fmt.Errorf("failed to queue footnote associations for HeadingNumber %d: %w", heading.HeadingNumber, err)
				}
			}

			// Queue child texts
			if len(heading.MeursingHeadingText) > 0 {
				if err := heading.MeursingHeadingText.QueueBatch(ctx, batch, heading.HeadingNumber); err != nil {
					return fmt.Errorf("failed to queue texts for HeadingNumber %d: %w", heading.HeadingNumber, err)
				}
			}

		case "D": // Delete
			batch.Queue(deleteQuery, heading.HeadingNumber, heading.MeursingTablePlanID, heading.RowColumnCode)

		default:
			return fmt.Errorf("unknown ChangeType: %s for HeadingNumber: %d", heading.ChangeType, heading.HeadingNumber)
		}

		if (i+1)%batchSize == 0 || i == len(headings)-1 {
			if err := conn.SendBatch(ctx, batch).Close(); err != nil {
				return fmt.Errorf("failed to execute batch: %w", err)
			}
			batch = &pgx.Batch{}
		}
	}

	return nil
}

type MeursingHeadingFootnoteAssociations []MeursingHeadingFootnoteAssociation

type MeursingHeadingFootnoteAssociation struct {
	DateStart    FileDistTime `xml:"dateStart,attr"`
	FootnoteID   int          `xml:"footnoteId,attr"`
	FootnoteType string       `xml:"footnoteType,attr"`
	National     int          `xml:"national,attr"`
}

func (associations MeursingHeadingFootnoteAssociations) QueueBatch(ctx context.Context, batch *pgx.Batch, parentHeadingID int) error {
	insertQuery := `
	INSERT INTO meursing_heading_footnote_association (
		parent_heading_id, footnote_id, footnote_type, date_start, national
	)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (parent_heading_id, footnote_id, footnote_type) DO UPDATE 
	SET date_start = EXCLUDED.date_start,
		national = EXCLUDED.national;
	`

	for _, assoc := range associations {
		batch.Queue(insertQuery, parentHeadingID, assoc.FootnoteID, assoc.FootnoteType, assoc.DateStart, assoc.National)
	}

	return nil
}

type MeursingHeadingTexts []MeursingHeadingText

type MeursingHeadingText struct {
	Description *string `xml:"description,attr"`
	LanguageID  string  `xml:"languageId,attr"`
	National    int     `xml:"national,attr"`
}

func (texts MeursingHeadingTexts) QueueBatch(ctx context.Context, batch *pgx.Batch, parentHeadingID int) error {
	insertQuery := `
	INSERT INTO meursing_heading_text (
		parent_heading_id, description, language_id, national
	)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (parent_heading_id, language_id) DO UPDATE 
	SET description = EXCLUDED.description,
		national = EXCLUDED.national;
	`

	for _, text := range texts {
		batch.Queue(insertQuery, parentHeadingID, text.Description, text.LanguageID, text.National)
	}

	return nil
}
