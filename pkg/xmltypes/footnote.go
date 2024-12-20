package xmltypes

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Footnotes []Footnote

type Footnote struct {
	ChangeType                 string                     `xml:"changeType,attr"`
	DateEnd                    FileDistTime               `xml:"dateEnd,attr"`
	DateStart                  FileDistTime               `xml:"dateStart,attr"`
	FootnoteID                 string                     `xml:"footnoteId,attr"`
	FootnoteType               string                     `xml:"footnoteType,attr"`
	National                   int                        `xml:"national,attr"`
	FootnoteDescriptionPeriods FootnoteDescriptionPeriods `xml:"footnoteDescriptionPeriod"`
}

func (footnotes Footnotes) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO footnote (footnote_id, footnote_type, change_type, date_start, date_end, national)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (footnote_id, footnote_type) DO UPDATE 
	SET change_type = EXCLUDED.change_type,
		date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		national = EXCLUDED.national;
	`

	deleteQuery := `
	DELETE FROM footnote
	WHERE footnote_id = $1 AND footnote_type = $2;
	`

	batch := &pgx.Batch{}

	for i, footnote := range footnotes {
		switch footnote.ChangeType {
		case "U": // Insert or update
			batch.Queue(insertQuery, footnote.FootnoteID, footnote.FootnoteType, footnote.ChangeType, footnote.DateStart, footnote.DateEnd, footnote.National)

			// Queue child description periods
			if len(footnote.FootnoteDescriptionPeriods) > 0 {
				if err := footnote.FootnoteDescriptionPeriods.QueueBatch(ctx, batch, footnote.FootnoteID, footnote.FootnoteType); err != nil {
					return fmt.Errorf("failed to queue description periods for FootnoteID %s, FootnoteType %s: %w", footnote.FootnoteID, footnote.FootnoteType, err)
				}
			}

		case "D": // Delete
			batch.Queue(deleteQuery, footnote.FootnoteID, footnote.FootnoteType)

		default:
			return fmt.Errorf("unknown ChangeType: %s for FootnoteID: %s, FootnoteType: %s", footnote.ChangeType, footnote.FootnoteID, footnote.FootnoteType)
		}

		if (i+1)%batchSize == 0 || i == len(footnotes)-1 {
			if err := conn.SendBatch(ctx, batch).Close(); err != nil {
				return fmt.Errorf("failed to execute batch: %w", err)
			}
			batch = &pgx.Batch{}
		}
	}

	return nil
}

type FootnoteDescriptionPeriods []FootnoteDescriptionPeriod

type FootnoteDescriptionPeriod struct {
	ParentFootnoteID     string               // added Parent ID
	ParentFootnoteType   string               // added Parent Type
	DateEnd              FileDistTime         `xml:"dateEnd,attr"`
	DateStart            FileDistTime         `xml:"dateStart,attr"`
	National             int                  `xml:"national,attr"`
	SID                  int                  `xml:"SID,attr"`
	FootnoteDescriptions FootnoteDescriptions `xml:"footnoteDescription"`
}

func (periods FootnoteDescriptionPeriods) QueueBatch(ctx context.Context, batch *pgx.Batch, parentFootnoteID, parentFootnoteType string) error {
	insertQuery := `
	INSERT INTO footnote_description_period (sid, parent_footnote_id, parent_footnote_type, date_start, date_end, national)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (sid) DO UPDATE 
	SET date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		national = EXCLUDED.national;
	`

	for _, period := range periods {
		period.ParentFootnoteID = parentFootnoteID
		period.ParentFootnoteType = parentFootnoteType
		batch.Queue(insertQuery, period.SID, period.ParentFootnoteID, period.ParentFootnoteType, period.DateStart, period.DateEnd, period.National)

		// Queue child descriptions
		if len(period.FootnoteDescriptions) > 0 {
			if err := period.FootnoteDescriptions.QueueBatch(ctx, batch, period.SID); err != nil {
				return fmt.Errorf("failed to queue descriptions for SID %d: %w", period.SID, err)
			}
		}
	}
	return nil
}

type FootnoteDescriptions []FootnoteDescription

type FootnoteDescription struct {
	ParentSID   int     // added Parent SID
	Description *string `xml:"description,attr"`
	LanguageID  string  `xml:"languageId,attr"`
	National    int     `xml:"national,attr"`
}

func (descriptions FootnoteDescriptions) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO footnote_description (parent_sid, description, language_id, national)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (parent_sid, language_id) DO UPDATE 
	SET description = EXCLUDED.description,
		national = EXCLUDED.national;
	`

	for _, desc := range descriptions {
		desc.ParentSID = parentSID
		batch.Queue(insertQuery, desc.ParentSID, desc.Description, desc.LanguageID, desc.National)
	}
	return nil
}
