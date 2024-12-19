// Using this file as an example on how the types are used in the application
package xmltypes

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type AdditionalCodes []AdditionalCode

type AdditionalCode struct {
	SID                                int                                `xml:"SID,attr"`
	AdditionalCodeID                   string                             `xml:"additionalCodeId,attr"`
	AdditionalCodeType                 string                             `xml:"additionalCodeType,attr"`
	ChangeType                         string                             `xml:"changeType,attr"`
	DateEnd                            FileDistTime                       `xml:"dateEnd,attr"`
	DateStart                          FileDistTime                       `xml:"dateStart,attr"`
	National                           int                                `xml:"national,attr"`
	AdditionalCodeDescriptionPeriods   AdditionalCodeDescriptionPeriods   `xml:"additionalCodeDescriptionPeriod"`
	AdditionalCodeFootnoteAssociations AdditionalCodeFootnoteAssociations `xml:"additionalCodeFootnoteAssociation"`
}

func (codes AdditionalCodes) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO additional_code (sid, additional_code_id, additional_code_type, change_type, date_start, date_end, national)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (sid) DO UPDATE 
	SET additional_code_id = EXCLUDED.additional_code_id,
		additional_code_type = EXCLUDED.additional_code_type,
		change_type = EXCLUDED.change_type,
		date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		national = EXCLUDED.national;
	`

	deleteQuery := `
	DELETE FROM additional_codes
	WHERE sid = $1;
	`

	batch := &pgx.Batch{}

	for i, code := range codes {
		switch code.ChangeType {
		case "U": // Add or update record

			batch.Queue(insertQuery, code.SID, code.AdditionalCodeID, code.AdditionalCodeType, code.ChangeType, code.DateStart, code.DateEnd, code.National)

			// Queue child description periods
			if len(code.AdditionalCodeDescriptionPeriods) > 0 {
				if err := code.AdditionalCodeDescriptionPeriods.QueueBatch(ctx, batch, code.SID); err != nil {
					return fmt.Errorf("failed to queue description periods for SID %d: %w", code.SID, err)
				}
			}

			// Queue child footnote associations
			if len(code.AdditionalCodeFootnoteAssociations) > 0 {
				if err := code.AdditionalCodeFootnoteAssociations.QueueBatch(ctx, batch, code.SID); err != nil {
					return fmt.Errorf("failed to queue footnote associations for SID %d: %w", code.SID, err)
				}
			}

		case "D": // Delete record
			fmt.Println("DELETING RECORD WITH SID: ", code.SID)
			batch.Queue(deleteQuery, code.SID)
		default:
			// Log an error or skip invalid ChangeType
			return fmt.Errorf("unknown ChangeType: %s for SID: %d", code.ChangeType, code.SID)
		}

		// Execute the batch when reaching batchSize or end of data
		if (i+1)%batchSize == 0 || i == len(codes)-1 {
			results := conn.SendBatch(ctx, batch)
			if err := results.Close(); err != nil {
				return fmt.Errorf("failed to execute batch: %w", err)
			}
			batch = &pgx.Batch{} // Reset the batch
		}
	}

	return nil
}

type AdditionalCodeDescriptionPeriods []AdditionalCodeDescriptionPeriod

type AdditionalCodeDescriptionPeriod struct {
	SID                        int                        `xml:"SID,attr"`
	ParentSID                  int                        // added type
	DateEnd                    FileDistTime               `xml:"dateEnd,attr"`
	DateStart                  FileDistTime               `xml:"dateStart,attr"`
	National                   int                        `xml:"national,attr"`
	AdditionalCodeDescriptions AdditionalCodeDescriptions `xml:"additionalCodeDescription"`
}

func (periods AdditionalCodeDescriptionPeriods) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO additional_code_description_period (sid, parent_sid, date_start, date_end, national)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (sid) DO UPDATE 
	SET date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		national = EXCLUDED.national;
	`

	for _, period := range periods {
		period.ParentSID = parentSID // Ensure parent relationship
		batch.Queue(insertQuery, period.SID, period.ParentSID, period.DateStart, period.DateEnd, period.National)

		// Queue child descriptions
		if len(period.AdditionalCodeDescriptions) > 0 {
			if err := period.AdditionalCodeDescriptions.QueueBatch(ctx, batch, period.SID); err != nil {
				return fmt.Errorf("failed to queue descriptions for SID %d: %w", period.SID, err)
			}
		}
	}
	return nil
}

type AdditionalCodeDescriptions []AdditionalCodeDescription

type AdditionalCodeDescription struct {
	ParentSID   int    // added type
	Description string `xml:"description,attr"`
	LanguageID  string `xml:"languageId,attr"`
	National    int    `xml:"national,attr"`
}

func (descriptions AdditionalCodeDescriptions) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO additional_code_description (parent_sid, description, language_id, national)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (parent_sid, language_id) DO UPDATE 
	SET description = EXCLUDED.description,
		national = EXCLUDED.national;
	`

	for _, desc := range descriptions {
		desc.ParentSID = parentSID // Ensure parent relationship
		batch.Queue(insertQuery, desc.ParentSID, desc.Description, desc.LanguageID, desc.National)
	}
	return nil
}

type AdditionalCodeFootnoteAssociations []AdditionalCodeFootnoteAssociation

type AdditionalCodeFootnoteAssociation struct {
	ParentSID    int          // added type
	DateEnd      FileDistTime `xml:"dateEnd,attr"`
	DateStart    FileDistTime `xml:"dateStart,attr"`
	FootnoteID   int          `xml:"footnoteId,attr"`
	FootnoteType string       `xml:"footnoteType,attr"`
	National     int          `xml:"national,attr"`
}

func (associations AdditionalCodeFootnoteAssociations) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO additional_code_footnote_association (parent_sid, date_start, date_end, footnote_id, footnote_type, national)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (parent_sid, footnote_id, footnote_type) DO UPDATE 
	SET date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		national = EXCLUDED.national;
	`

	for _, assoc := range associations {
		assoc.ParentSID = parentSID // Set parent relationship
		batch.Queue(insertQuery, assoc.ParentSID, assoc.DateStart, assoc.DateEnd, assoc.FootnoteID, assoc.FootnoteType, assoc.National)
	}

	return nil
}
