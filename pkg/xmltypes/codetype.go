package xmltypes

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type CodeTypes []CodeType

type CodeType struct {
	ID                   string               `xml:"id,attr"`
	ChangeType           string               `xml:"changeType,attr"`
	CodeTypeID           string               `xml:"codeTypeId,attr"`
	DateEnd              FileDistTime         `xml:"dateEnd,attr"`
	DateStart            FileDistTime         `xml:"dateStart,attr"`
	ExportImportType     string               `xml:"exportImportType,attr"`
	MeasureTypeSeriesID  *string              `xml:"measureTypeSeriesId,attr"`
	National             int                  `xml:"national,attr"`
	CodeTypeDescriptions CodeTypeDescriptions `xml:"codeTypeDescription"`
}

func (codeTypes CodeTypes) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO code_type (id, code_type_id, change_type, date_start, date_end, export_import_type, measure_type_series_id, national)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT (id) DO UPDATE 
	SET code_type_id = EXCLUDED.code_type_id,
		change_type = EXCLUDED.change_type,
		date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		export_import_type = EXCLUDED.export_import_type,
		measure_type_series_id = EXCLUDED.measure_type_series_id,
		national = EXCLUDED.national;
	`

	deleteQuery := `
	DELETE FROM code_type
	WHERE id = $1;
	`

	batch := &pgx.Batch{}

	for i, codeType := range codeTypes {
		switch codeType.ChangeType {
		case "U": // Insert or update
			batch.Queue(insertQuery, codeType.ID, codeType.CodeTypeID, codeType.ChangeType, codeType.DateStart, codeType.DateEnd, codeType.ExportImportType, codeType.MeasureTypeSeriesID, codeType.National)

			// Queue child descriptions
			if len(codeType.CodeTypeDescriptions) > 0 {
				if err := codeType.CodeTypeDescriptions.QueueBatch(ctx, batch, codeType.ID); err != nil {
					return fmt.Errorf("failed to queue descriptions for CodeType ID %s: %w", codeType.ID, err)
				}
			}

		case "D": // Delete
			batch.Queue(deleteQuery, codeType.ID)

		default:
			return fmt.Errorf("unknown ChangeType: %s for CodeType ID: %s", codeType.ChangeType, codeType.ID)
		}

		if (i+1)%batchSize == 0 || i == len(codeTypes)-1 {
			if err := conn.SendBatch(ctx, batch).Close(); err != nil {
				return fmt.Errorf("failed to execute batch: %w", err)
			}
			batch = &pgx.Batch{}
		}
	}

	return nil
}

type CodeTypeDescriptions []CodeTypeDescription

type CodeTypeDescription struct {
	ParentID    string // added type
	Description string `xml:"description,attr"`
	LanguageID  string `xml:"languageId,attr"`
	National    int    `xml:"national,attr"`
}

func (descriptions CodeTypeDescriptions) QueueBatch(ctx context.Context, batch *pgx.Batch, parentID string) error {
	insertQuery := `
	INSERT INTO code_type_description (parent_id, description, language_id, national)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (parent_id, language_id) DO UPDATE 
	SET description = EXCLUDED.description,
		national = EXCLUDED.national;
	`

	for _, desc := range descriptions {
		desc.ParentID = parentID // Ensure parent relationship
		batch.Queue(insertQuery, desc.ParentID, desc.Description, desc.LanguageID, desc.National)
	}
	return nil
}
