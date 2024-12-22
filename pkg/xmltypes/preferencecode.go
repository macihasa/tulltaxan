package xmltypes

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type PreferenceCodes []PreferenceCode

type PreferenceCode struct {
	PrefCode                   int                        `xml:"prefCode,attr"`
	DateStart                  FileDistTime               `xml:"dateStart,attr"`
	ChangeType                 string                     `xml:"changeType,attr"`
	PreferenceCodeDescriptions PreferenceCodeDescriptions `xml:"preferenceCodeDescription"`
}

func (codes PreferenceCodes) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO preference_code (
		pref_code, date_start, change_type
	)
	VALUES ($1, $2, $3)
	ON CONFLICT (pref_code) DO UPDATE 
	SET date_start = EXCLUDED.date_start,
		change_type = EXCLUDED.change_type;
	`

	deleteQuery := `
	DELETE FROM preference_code
	WHERE pref_code = $1;
	`

	batch := &pgx.Batch{}

	for i, code := range codes {
		switch code.ChangeType {
		case "U": // Insert or update
			batch.Queue(insertQuery, code.PrefCode, code.DateStart, code.ChangeType)

			// Queue child descriptions
			if len(code.PreferenceCodeDescriptions) > 0 {
				if err := code.PreferenceCodeDescriptions.QueueBatch(ctx, batch, code.PrefCode); err != nil {
					return fmt.Errorf("failed to queue descriptions for PrefCode %d: %w", code.PrefCode, err)
				}
			}

		case "D": // Delete
			batch.Queue(deleteQuery, code.PrefCode)

		default:
			return fmt.Errorf("unknown ChangeType: %s for PrefCode: %d", code.ChangeType, code.PrefCode)
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

type PreferenceCodeDescriptions []PreferenceCodeDescription

type PreferenceCodeDescription struct {
	Description string `xml:"description,attr"`
	LanguageID  string `xml:"languageId,attr"`
}

func (descriptions PreferenceCodeDescriptions) QueueBatch(ctx context.Context, batch *pgx.Batch, parentPrefCode int) error {
	insertQuery := `
	INSERT INTO preference_code_description (
		parent_pref_code, description, language_id
	)
	VALUES ($1, $2, $3)
	ON CONFLICT (parent_pref_code, language_id) DO UPDATE 
	SET description = EXCLUDED.description;
	`

	for _, desc := range descriptions {
		batch.Queue(insertQuery, parentPrefCode, desc.Description, desc.LanguageID)
	}

	return nil
}
