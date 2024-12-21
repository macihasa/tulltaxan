package xmltypes

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type MeasurementUnits []MeasurementUnit

type MeasurementUnit struct {
	MeasurementUnitCode        string                       `xml:"measurementUnitCode,attr"`
	DateStart                  FileDistTime                 `xml:"dateStart,attr"`
	DateEnd                    FileDistTime                 `xml:"dateEnd,attr"`
	National                   int                          `xml:"national,attr"`
	NationalAbbreviation       string                       `xml:"nationalAbbreviation"`
	ChangeType                 string                       `xml:"changeType,attr"`
	MeasurementUnitDescription []MeasurementUnitDescription `xml:"measurementUnitDescription"`
}

func (units MeasurementUnits) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO measurement_unit (
		measurement_unit_code, date_start, date_end, national, national_abbreviation, change_type
	)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (measurement_unit_code) DO UPDATE 
	SET date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		national = EXCLUDED.national,
		national_abbreviation = EXCLUDED.national_abbreviation,
		change_type = EXCLUDED.change_type;
	`

	deleteQuery := `
	DELETE FROM measurement_unit
	WHERE measurement_unit_code = $1;
	`

	batch := &pgx.Batch{}

	for i, unit := range units {
		switch unit.ChangeType {
		case "U": // Insert or update
			batch.Queue(insertQuery, unit.MeasurementUnitCode, unit.DateStart, unit.DateEnd, unit.National, unit.NationalAbbreviation, unit.ChangeType)

			// Queue child descriptions
			if len(unit.MeasurementUnitDescription) > 0 {
				if err := MeasurementUnitDescriptions(unit.MeasurementUnitDescription).QueueBatch(ctx, batch, unit.MeasurementUnitCode); err != nil {
					return fmt.Errorf("failed to queue descriptions for MeasurementUnitCode %s: %w", unit.MeasurementUnitCode, err)
				}
			}

		case "D": // Delete
			batch.Queue(deleteQuery, unit.MeasurementUnitCode)

		default:
			return fmt.Errorf("unknown ChangeType: %s for MeasurementUnitCode: %s", unit.ChangeType, unit.MeasurementUnitCode)
		}

		if (i+1)%batchSize == 0 || i == len(units)-1 {
			if err := conn.SendBatch(ctx, batch).Close(); err != nil {
				return fmt.Errorf("failed to execute batch: %w", err)
			}
			batch = &pgx.Batch{}
		}
	}

	return nil
}

type MeasurementUnitDescriptions []MeasurementUnitDescription

type MeasurementUnitDescription struct {
	Description string `xml:"description,attr"`
	LanguageID  string `xml:"languageId,attr"`
	National    int    `xml:"national,attr"`
}

func (descriptions MeasurementUnitDescriptions) QueueBatch(ctx context.Context, batch *pgx.Batch, parentUnitCode string) error {
	insertQuery := `
	INSERT INTO measurement_unit_description (
		parent_unit_code, description, language_id, national
	)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (parent_unit_code, language_id) DO UPDATE 
	SET description = EXCLUDED.description,
		national = EXCLUDED.national;
	`

	for _, desc := range descriptions {
		batch.Queue(insertQuery, parentUnitCode, desc.Description, desc.LanguageID, desc.National)
	}
	return nil
}
