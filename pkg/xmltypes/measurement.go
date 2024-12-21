package xmltypes

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Measurements []Measurement

type Measurement struct {
	ChangeType                   string       `xml:"changeType,attr"`
	DateEnd                      FileDistTime `xml:"dateEnd,attr"`
	DateStart                    FileDistTime `xml:"dateStart,attr"`
	MeasurementUnitCode          string       `xml:"measurementUnitCode,attr"`
	MeasurementUnitQualifierCode string       `xml:"measurementUnitQualifierCode,attr"`
	National                     int          `xml:"national,attr"`
}

func (measurements Measurements) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO measurement (
		measurement_unit_code, measurement_unit_qualifier_code, change_type, date_start, date_end, national
	)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (measurement_unit_code, measurement_unit_qualifier_code) DO UPDATE 
	SET change_type = EXCLUDED.change_type,
		date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		national = EXCLUDED.national;
	`

	deleteQuery := `
	DELETE FROM measurement
	WHERE measurement_unit_code = $1 AND measurement_unit_qualifier_code = $2;
	`

	batch := &pgx.Batch{}

	for i, measurement := range measurements {
		switch measurement.ChangeType {
		case "U": // Insert or update
			batch.Queue(insertQuery, measurement.MeasurementUnitCode, measurement.MeasurementUnitQualifierCode, measurement.ChangeType, measurement.DateStart, measurement.DateEnd, measurement.National)

		case "D": // Delete
			batch.Queue(deleteQuery, measurement.MeasurementUnitCode, measurement.MeasurementUnitQualifierCode)

		default:
			return fmt.Errorf("unknown ChangeType: %s for MeasurementUnitCode: %s, MeasurementUnitQualifierCode: %s", measurement.ChangeType, measurement.MeasurementUnitCode, measurement.MeasurementUnitQualifierCode)
		}

		if (i+1)%batchSize == 0 || i == len(measurements)-1 {
			if err := conn.SendBatch(ctx, batch).Close(); err != nil {
				return fmt.Errorf("failed to execute batch: %w", err)
			}
			batch = &pgx.Batch{}
		}
	}

	return nil
}
