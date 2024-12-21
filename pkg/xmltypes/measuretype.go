package xmltypes

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type MeasureTypes []MeasureType

type MeasureType struct {
	ChangeType                     string                  `xml:"changeType,attr"`
	DateEnd                        FileDistTime            `xml:"dateEnd,attr"`
	DateStart                      FileDistTime            `xml:"dateStart,attr"`
	ExplosionLevel                 int                     `xml:"explosionLevel,attr"`
	MeasureComponentApplicableCode int                     `xml:"measureComponentApplicableCode,attr"`
	MeasureType                    string                  `xml:"measureType,attr"`
	MeasureTypeSeriesID            string                  `xml:"measureTypeSeriesId,attr"`
	National                       int                     `xml:"national,attr"`
	OrderNumberCaptureCode         int                     `xml:"orderNumberCaptureCode,attr"`
	OriginDestinationCode          int                     `xml:"originDestinationCode,attr"`
	PriorityCode                   int                     `xml:"priorityCode,attr"`
	TradeMovementCode              int                     `xml:"tradeMovementCode,attr"`
	MeasureTypeDescriptions        MeasureTypeDescriptions `xml:"measureTypeDescription"`
}

func (types MeasureTypes) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO measure_type (
		measure_type, measure_type_series_id, change_type, date_start, date_end, explosion_level,
		measure_component_applicable_code, order_number_capture_code, origin_destination_code,
		priority_code, trade_movement_code, national
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	ON CONFLICT (measure_type) DO UPDATE 
	SET measure_type_series_id = EXCLUDED.measure_type_series_id,
		change_type = EXCLUDED.change_type,
		date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		explosion_level = EXCLUDED.explosion_level,
		measure_component_applicable_code = EXCLUDED.measure_component_applicable_code,
		order_number_capture_code = EXCLUDED.order_number_capture_code,
		origin_destination_code = EXCLUDED.origin_destination_code,
		priority_code = EXCLUDED.priority_code,
		trade_movement_code = EXCLUDED.trade_movement_code,
		national = EXCLUDED.national;
	`

	deleteQuery := `
	DELETE FROM measure_type
	WHERE measure_type = $1;
	`

	batch := &pgx.Batch{}

	for i, measureType := range types {
		switch measureType.ChangeType {
		case "U": // Insert or update
			batch.Queue(insertQuery, measureType.MeasureType, measureType.MeasureTypeSeriesID, measureType.ChangeType, measureType.DateStart, measureType.DateEnd, measureType.ExplosionLevel, measureType.MeasureComponentApplicableCode, measureType.OrderNumberCaptureCode, measureType.OriginDestinationCode, measureType.PriorityCode, measureType.TradeMovementCode, measureType.National)

			// Queue child descriptions
			if len(measureType.MeasureTypeDescriptions) > 0 {
				if err := measureType.MeasureTypeDescriptions.QueueBatch(ctx, batch, measureType.MeasureType); err != nil {
					return fmt.Errorf("failed to queue descriptions for MeasureType %s: %w", measureType.MeasureType, err)
				}
			}

		case "D": // Delete
			batch.Queue(deleteQuery, measureType.MeasureType)

		default:
			return fmt.Errorf("unknown ChangeType: %s for MeasureType: %s", measureType.ChangeType, measureType.MeasureType)
		}

		if (i+1)%batchSize == 0 || i == len(types)-1 {
			if err := conn.SendBatch(ctx, batch).Close(); err != nil {
				return fmt.Errorf("failed to execute batch: %w", err)
			}
			batch = &pgx.Batch{}
		}
	}

	return nil
}

type MeasureTypeDescriptions []MeasureTypeDescription

type MeasureTypeDescription struct {
	ParentMeasureType string // added ParentMeasureType
	Description       string `xml:"description,attr"`
	LanguageID        string `xml:"languageId,attr"`
	National          int    `xml:"national,attr"`
}

func (descriptions MeasureTypeDescriptions) QueueBatch(ctx context.Context, batch *pgx.Batch, parentMeasureType string) error {
	insertQuery := `
	INSERT INTO measure_type_description (parent_measure_type, description, language_id, national)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (parent_measure_type, language_id) DO UPDATE 
	SET description = EXCLUDED.description,
		national = EXCLUDED.national;
	`

	for _, desc := range descriptions {
		desc.ParentMeasureType = parentMeasureType
		batch.Queue(insertQuery, desc.ParentMeasureType, desc.Description, desc.LanguageID, desc.National)
	}
	return nil
}
