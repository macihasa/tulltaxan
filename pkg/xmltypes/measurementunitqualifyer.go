package xmltypes

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type MeasurementUnitQualifiers []MeasurementUnitQualifier

type MeasurementUnitQualifier struct {
	ChangeType                           string                               `xml:"changeType,attr"`
	DateStart                            FileDistTime                         `xml:"dateStart,attr"`
	MeasurementUnitQualifierCode         string                               `xml:"measurementUnitQualifierCode,attr"`
	National                             int                                  `xml:"national,attr"`
	MeasurementUnitQualifierDescriptions MeasurementUnitQualifierDescriptions `xml:"measurementUnitQualifierDescription"`
}

func (qualifiers MeasurementUnitQualifiers) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO measurement_unit_qualifier (
		measurement_unit_qualifier_code, change_type, date_start, national
	)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (measurement_unit_qualifier_code) DO UPDATE 
	SET change_type = EXCLUDED.change_type,
		date_start = EXCLUDED.date_start,
		national = EXCLUDED.national;
	`

	deleteQuery := `
	DELETE FROM measurement_unit_qualifier
	WHERE measurement_unit_qualifier_code = $1;
	`

	batch := &pgx.Batch{}

	for i, qualifier := range qualifiers {
		switch qualifier.ChangeType {
		case "U": // Insert or update
			batch.Queue(insertQuery, qualifier.MeasurementUnitQualifierCode, qualifier.ChangeType, qualifier.DateStart, qualifier.National)

			// Queue child descriptions
			if len(qualifier.MeasurementUnitQualifierDescriptions) > 0 {
				if err := qualifier.MeasurementUnitQualifierDescriptions.QueueBatch(ctx, batch, qualifier.MeasurementUnitQualifierCode); err != nil {
					return fmt.Errorf("failed to queue descriptions for MeasurementUnitQualifierCode %s: %w", qualifier.MeasurementUnitQualifierCode, err)
				}
			}

		case "D": // Delete
			batch.Queue(deleteQuery, qualifier.MeasurementUnitQualifierCode)

		default:
			return fmt.Errorf("unknown ChangeType: %s for MeasurementUnitQualifierCode: %s", qualifier.ChangeType, qualifier.MeasurementUnitQualifierCode)
		}

		if (i+1)%batchSize == 0 || i == len(qualifiers)-1 {
			if err := conn.SendBatch(ctx, batch).Close(); err != nil {
				return fmt.Errorf("failed to execute batch: %w", err)
			}
			batch = &pgx.Batch{}
		}
	}

	return nil
}

type MeasurementUnitQualifierDescriptions []MeasurementUnitQualifierDescription

type MeasurementUnitQualifierDescription struct {
	ParentMeasurementUnitQualifierCode string // added type
	Description                        string `xml:"description,attr"`
	LanguageID                         string `xml:"languageId,attr"`
	National                           int    `xml:"national,attr"`
}

func (descriptions MeasurementUnitQualifierDescriptions) QueueBatch(ctx context.Context, batch *pgx.Batch, parentQualifierCode string) error {
	insertQuery := `
	INSERT INTO measurement_unit_qualifier_description (
		parent_measurement_unit_qualifier_code, description, language_id, national
	)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (parent_measurement_unit_qualifier_code, language_id) DO UPDATE 
	SET description = EXCLUDED.description,
		national = EXCLUDED.national;
	`

	for _, desc := range descriptions {
		desc.ParentMeasurementUnitQualifierCode = parentQualifierCode
		batch.Queue(insertQuery, desc.ParentMeasurementUnitQualifierCode, desc.Description, desc.LanguageID, desc.National)
	}
	return nil
}
