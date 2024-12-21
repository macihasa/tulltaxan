package xmltypes

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type MonetaryExchangePeriods []MonetaryExchangePeriod

type MonetaryExchangePeriod struct {
	ChangeType           string                `xml:"changeType,attr"`
	DateEnd              FileDistTime          `xml:"dateEnd,attr"`
	DateStart            FileDistTime          `xml:"dateStart,attr"`
	MonetaryUnitCode     string                `xml:"monetaryUnitCode,attr"`
	National             int                   `xml:"national,attr"`
	SID                  int                   `xml:"SID,attr"`
	MonetaryExchangeRate MonetaryExchangeRates `xml:"monetaryExchangeRate"`
}

func (periods MonetaryExchangePeriods) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO monetary_exchange_period (
		sid, monetary_unit_code, change_type, date_start, date_end, national, is_quoted
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (sid) DO UPDATE 
	SET monetary_unit_code = EXCLUDED.monetary_unit_code,
		change_type = EXCLUDED.change_type,
		date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		national = EXCLUDED.national,
		is_quoted = EXCLUDED.is_quoted;
	`

	deleteQuery := `
	DELETE FROM monetary_exchange_period
	WHERE sid = $1;
	`

	batch := &pgx.Batch{}

	for i, period := range periods {
		switch period.ChangeType {
		case "U": // Insert or update
			batch.Queue(insertQuery, period.SID, period.MonetaryUnitCode, period.ChangeType, period.DateStart, period.DateEnd, period.National, true)

			// Queue child exchange rates
			if len(period.MonetaryExchangeRate) > 0 {
				if err := period.MonetaryExchangeRate.QueueBatch(ctx, batch, period.SID); err != nil {
					return fmt.Errorf("failed to queue monetary exchange rates for SID %d: %w", period.SID, err)
				}
			}

		case "D": // Delete
			batch.Queue(deleteQuery, period.SID)

		default:
			return fmt.Errorf("unknown ChangeType: %s for SID: %d", period.ChangeType, period.SID)
		}

		if (i+1)%batchSize == 0 || i == len(periods)-1 {
			if err := conn.SendBatch(ctx, batch).Close(); err != nil {
				return fmt.Errorf("failed to execute batch: %w", err)
			}
			batch = &pgx.Batch{}
		}
	}

	return nil
}

type UnquotedMonetaryExchangePeriods []UnquotedMonetaryExchangePeriod

type UnquotedMonetaryExchangePeriod struct {
	ChangeType            string                `xml:"changeType,attr"`
	DateEnd               FileDistTime          `xml:"dateEnd,attr"`
	DateStart             FileDistTime          `xml:"dateStart,attr"`
	MonetaryUnitCode      string                `xml:"monetaryUnitCode,attr"`
	National              int                   `xml:"national,attr"`
	SID                   int                   `xml:"SID,attr"`
	MonetaryExchangeRates MonetaryExchangeRates `xml:"unquotedMonetaryExchangeRate"`
}

func (periods UnquotedMonetaryExchangePeriods) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO monetary_exchange_period (
		sid, monetary_unit_code, change_type, date_start, date_end, national, is_quoted
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (sid) DO UPDATE 
	SET monetary_unit_code = EXCLUDED.monetary_unit_code,
		change_type = EXCLUDED.change_type,
		date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		national = EXCLUDED.national,
		is_quoted = EXCLUDED.is_quoted;
	`

	deleteQuery := `
	DELETE FROM unquoted_monetary_exchange_period
	WHERE sid = $1;
	`

	batch := &pgx.Batch{}

	for i, period := range periods {
		switch period.ChangeType {
		case "U": // Insert or update
			batch.Queue(insertQuery, period.SID, period.MonetaryUnitCode, period.ChangeType, period.DateStart, period.DateEnd, period.National, false)

			// Queue child exchange rates
			if len(period.MonetaryExchangeRates) > 0 {
				if err := period.MonetaryExchangeRates.QueueBatch(ctx, batch, period.SID); err != nil {
					return fmt.Errorf("failed to queue unquoted monetary exchange rates for SID %d: %w", period.SID, err)
				}
			}

		case "D": // Delete
			batch.Queue(deleteQuery, period.SID)

		default:
			return fmt.Errorf("unknown ChangeType: %s for SID: %d", period.ChangeType, period.SID)
		}

		if (i+1)%batchSize == 0 || i == len(periods)-1 {
			if err := conn.SendBatch(ctx, batch).Close(); err != nil {
				return fmt.Errorf("failed to execute batch: %w", err)
			}
			batch = &pgx.Batch{}
		}
	}

	return nil
}

type MonetaryExchangeRates []MonetaryExchangeRate

type MonetaryExchangeRate struct {
	ParentSID              int     // added Parent SID
	CalculationUnit        *int    `xml:"calculationUnit,attr"`
	MonetaryConversionRate float64 `xml:"monetaryConversionRate,attr"`
	MonetaryUnitCode       string  `xml:"monetaryUnitCode,attr"`
	National               int     `xml:"national,attr"`
}

func (rates MonetaryExchangeRates) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO monetary_exchange_rate (
		parent_sid, calculation_unit, monetary_conversion_rate, monetary_unit_code, national
	)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (parent_sid, monetary_unit_code) DO UPDATE 
	SET calculation_unit = EXCLUDED.calculation_unit,
		monetary_conversion_rate = EXCLUDED.monetary_conversion_rate,
		national = EXCLUDED.national;
	`

	for _, rate := range rates {
		rate.ParentSID = parentSID
		batch.Queue(insertQuery, rate.ParentSID, rate.CalculationUnit, rate.MonetaryConversionRate, rate.MonetaryUnitCode, rate.National)
	}
	return nil
}
