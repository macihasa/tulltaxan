package xmltypes

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type QuotaDefinitions []QuotaDefinition

type QuotaDefinition struct {
	SID                          int                     `xml:"SID,attr"`
	SIDQuotaOrderNumber          int                     `xml:"SIDQuotaOrderNumber,attr"`
	QuotaCriticalStateCode       string                  `xml:"quotaCriticalStateCode,attr"`
	QuotaCriticalThreshold       int                     `xml:"quotaCriticalThreshold,attr"`
	QuotaMaximumPrecision        int                     `xml:"quotaMaximumPrecision,attr"`
	QuotaOrderNumber             int                     `xml:"quotaOrderNumber,attr"`
	ChangeType                   string                  `xml:"changeType,attr"`
	Description                  *string                 `xml:"description,attr"`
	InitialVolume                float64                 `xml:"initialVolume,attr"`
	MeasurementUnitCode          *string                 `xml:"measurementUnitCode,attr"`
	MeasurementUnitQualifierCode *string                 `xml:"measurementUnitQualifierCode,attr"`
	MonetaryUnitCode             *string                 `xml:"monetaryUnitCode,attr"`
	Volume                       float64                 `xml:"volume,attr"`
	DateStart                    FileDistTime            `xml:"dateStart,attr"`
	DateEnd                      FileDistTime            `xml:"dateEnd,attr"`
	National                     int                     `xml:"national,attr"`
	QuotaBlockingPeriod          []QuotaBlockingPeriod   `xml:"quotaBlockingPeriod"`
	QuotaAssociation             []QuotaAssociation      `xml:"quotaAssociation"`
	QuotaSuspensionPeriod        []QuotaSuspensionPeriod `xml:"quotaSuspensionPeriod"`
}

func (definitions QuotaDefinitions) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO quota_definition (
		sid, sid_quota_order_number, quota_critical_state_code, quota_critical_threshold, quota_maximum_precision,
		quota_order_number, change_type, description, initial_volume, measurement_unit_code,
		measurement_unit_qualifier_code, monetary_unit_code, volume, date_start, date_end, national
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	ON CONFLICT (sid) DO UPDATE 
	SET sid_quota_order_number = EXCLUDED.sid_quota_order_number,
		quota_critical_state_code = EXCLUDED.quota_critical_state_code,
		quota_critical_threshold = EXCLUDED.quota_critical_threshold,
		quota_maximum_precision = EXCLUDED.quota_maximum_precision,
		quota_order_number = EXCLUDED.quota_order_number,
		change_type = EXCLUDED.change_type,
		description = EXCLUDED.description,
		initial_volume = EXCLUDED.initial_volume,
		measurement_unit_code = EXCLUDED.measurement_unit_code,
		measurement_unit_qualifier_code = EXCLUDED.measurement_unit_qualifier_code,
		monetary_unit_code = EXCLUDED.monetary_unit_code,
		volume = EXCLUDED.volume,
		date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		national = EXCLUDED.national;
	`

	batch := &pgx.Batch{}

	// Insert parent records first
	for i, def := range definitions {
		batch.Queue(insertQuery, def.SID, def.SIDQuotaOrderNumber, def.QuotaCriticalStateCode, def.QuotaCriticalThreshold,
			def.QuotaMaximumPrecision, def.QuotaOrderNumber, def.ChangeType, def.Description, def.InitialVolume,
			def.MeasurementUnitCode, def.MeasurementUnitQualifierCode, def.MonetaryUnitCode, def.Volume, def.DateStart,
			def.DateEnd, def.National)

		// Commit the batch every batchSize or at the end
		if (i+1)%batchSize == 0 || i == len(definitions)-1 {
			if err := conn.SendBatch(ctx, batch).Close(); err != nil {
				return fmt.Errorf("failed to insert quota definitions: %w", err)
			}
			batch = &pgx.Batch{} // Reset the batch
		}
	}

	// Insert child records after parent records are fully committed
	for _, def := range definitions {
		if len(def.QuotaBlockingPeriod) > 0 {
			if err := QuotaBlockingPeriods(def.QuotaBlockingPeriod).QueueBatch(ctx, batch, def.SID); err != nil {
				return fmt.Errorf("failed to queue blocking periods for SID %d: %w", def.SID, err)
			}
		}
		if len(def.QuotaAssociation) > 0 {
			if err := QuotaAssociations(def.QuotaAssociation).QueueBatch(ctx, batch, def.SID); err != nil {
				return fmt.Errorf("failed to queue associations for SID %d: %w", def.SID, err)
			}
		}
		if len(def.QuotaSuspensionPeriod) > 0 {
			if err := QuotaSuspensionPeriods(def.QuotaSuspensionPeriod).QueueBatch(ctx, batch, def.SID); err != nil {
				return fmt.Errorf("failed to queue suspension periods for SID %d: %w", def.SID, err)
			}
		}

		if err := conn.SendBatch(ctx, batch).Close(); err != nil {
			return fmt.Errorf("failed to insert child records: %w", err)
		}
		batch = &pgx.Batch{} // Reset the batch
	}

	return nil
}

type QuotaBlockingPeriods []QuotaBlockingPeriod
type QuotaBlockingPeriod struct {
	SID                int          `xml:"SID,attr"`
	BlockingPeriodType int          `xml:"blockingPeriodType,attr"`
	Description        *string      `xml:"description,attr"`
	DateStart          FileDistTime `xml:"dateStart,attr"`
	DateEnd            FileDistTime `xml:"dateEnd,attr"`
	National           int          `xml:"national,attr"`
}

func (periods QuotaBlockingPeriods) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO quota_blocking_period (
		sid, parent_sid, blocking_period_type, description, date_start, date_end, national
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (sid) DO UPDATE 
	SET blocking_period_type = EXCLUDED.blocking_period_type,
		description = EXCLUDED.description,
		date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		national = EXCLUDED.national;
	`

	for _, period := range periods {
		batch.Queue(insertQuery, period.SID, parentSID, period.BlockingPeriodType, period.Description, period.DateStart, period.DateEnd, period.National)
	}

	return nil
}

type QuotaAssociations []QuotaAssociation
type QuotaAssociation struct {
	SIDSubQuota  int     `xml:"SIDSubQuota,attr"`
	RelationType string  `xml:"relationType,attr"`
	Coefficient  float64 `xml:"coefficient,attr"`
	National     int     `xml:"national,attr"`
}

func (associations QuotaAssociations) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO quota_association (
		sid_sub_quota, parent_sid, relation_type, coefficient, national
	)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (sid_sub_quota) DO UPDATE 
	SET relation_type = EXCLUDED.relation_type,
		coefficient = EXCLUDED.coefficient,
		national = EXCLUDED.national;
	`

	for _, assoc := range associations {
		batch.Queue(insertQuery, assoc.SIDSubQuota, parentSID, assoc.RelationType, assoc.Coefficient, assoc.National)
	}

	return nil
}

type QuotaSuspensionPeriods []QuotaSuspensionPeriod
type QuotaSuspensionPeriod struct {
	SID         int          `xml:"SID,attr"`
	Description *string      `xml:"description,attr"`
	DateStart   FileDistTime `xml:"dateStart,attr"`
	DateEnd     FileDistTime `xml:"dateEnd,attr"`
	National    int          `xml:"national,attr"`
}

func (periods QuotaSuspensionPeriods) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO quota_suspension_period (
		sid, parent_sid, description, date_start, date_end, national
	)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (sid) DO UPDATE 
	SET description = EXCLUDED.description,
		date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		national = EXCLUDED.national;
	`

	for _, period := range periods {
		batch.Queue(insertQuery, period.SID, parentSID, period.Description, period.DateStart, period.DateEnd, period.National)
	}

	return nil
}

// type QuotaBalanceEvent struct {
// 	ChangeType             string            `xml:"changeType,attr"`
// 	EndOccurrenceTimestamp FileDistTimeStamp `xml:"endOccurrenceTimestamp,attr"`
// 	ImportedAmount         float64           `xml:"importedAmount,attr"`
// 	LastImportDate         FileDistTime      `xml:"lastImportDate,attr"`
// 	National               int               `xml:"national,attr"`
// 	NewBalance             float64           `xml:"newBalance,attr"`
// 	OccurrenceTimestamp    FileDistTimeStamp `xml:"occurrenceTimestamp,attr"`
// 	OldBalance             float64           `xml:"oldBalance,attr"`
// 	SIDQuotaDefinition     int               `xml:"SIDQuotaDefinition,attr"`
// }

// type QuotaCriticalEvent struct {
// 	ChangeType             string            `xml:"changeType,attr"`
// 	CriticalDate           FileDistTime      `xml:"criticalDate,attr"`
// 	EndOccurrenceTimestamp FileDistTimeStamp `xml:"endOccurrenceTimestamp,attr"`
// 	National               int               `xml:"national,attr"`
// 	OccurrenceTimestamp    FileDistTimeStamp `xml:"occurrenceTimestamp,attr"`
// 	QuotaCriticalStateCode string            `xml:"quotaCriticalStateCode,attr"`
// 	SIDQuotaDefinition     int               `xml:"SIDQuotaDefinition,attr"`
// }

// type QuotaExhaustionEvent struct {
// 	ChangeType             string            `xml:"changeType,attr"`
// 	EndOccurrenceTimestamp FileDistTimeStamp `xml:"endOccurrenceTimestamp,attr"`
// 	ExhaustionDate         FileDistTime      `xml:"exhaustionDate,attr"`
// 	National               int               `xml:"national,attr"`
// 	OccurrenceTimestamp    FileDistTimeStamp `xml:"occurrenceTimestamp,attr"`
// 	SIDQuotaDefinition     int               `xml:"SIDQuotaDefinition,attr"`
// }

// type QuotaReopeningEvent struct {
// 	ChangeType             string            `xml:"changeType,attr"`
// 	EndOccurrenceTimestamp FileDistTimeStamp `xml:"endOccurrenceTimestamp,attr"`
// 	National               int               `xml:"national,attr"`
// 	OccurrenceTimestamp    FileDistTimeStamp `xml:"occurrenceTimestamp,attr"`
// 	ReopeningDate          FileDistTime      `xml:"reopeningDate,attr"`
// 	SIDQuotaDefinition     int               `xml:"SIDQuotaDefinition,attr"`
// }

// type QuotaOrderNumber struct {
// 	ChangeType             string                   `xml:"changeType,attr"`
// 	DateEnd                FileDistTime             `xml:"dateEnd,attr"`
// 	DateStart              FileDistTime             `xml:"dateStart,attr"`
// 	National               int                      `xml:"national,attr"`
// 	QuotaOrderNumber       int                      `xml:"quotaOrderNumber,attr"`
// 	SID                    int                      `xml:"SID,attr"`
// 	QuotaOrderNumberOrigin []QuotaOrderNumberOrigin `xml:"quotaOrderNumberOrigin"`
// }

// type QuotaOrderNumberOrigin struct {
// 	DateEnd                         FileDistTime                      `xml:"dateEnd,attr"`
// 	DateStart                       FileDistTime                      `xml:"dateStart,attr"`
// 	GeographicalAreaID              string                            `xml:"geographicalAreaId,attr"`
// 	National                        int                               `xml:"national,attr"`
// 	SID                             int                               `xml:"SID,attr"`
// 	SIDGeographicalArea             int                               `xml:"SIDGeographicalArea,attr"`
// 	QuotaOrderNumberOriginExclusion []QuotaOrderNumberOriginExclusion `xml:"quotaOrderNumberOriginExclusion"`
// }

// type QuotaUnblockingEvent struct {
// 	ChangeType             string            `xml:"changeType,attr"`
// 	EndOccurrenceTimestamp FileDistTimeStamp `xml:"endOccurrenceTimestamp,attr"`
// 	National               int               `xml:"national,attr"`
// 	OccurrenceTimestamp    FileDistTimeStamp `xml:"occurrenceTimestamp,attr"`
// 	SIDQuotaDefinition     int               `xml:"SIDQuotaDefinition,attr"`
// 	UnblockingDate         FileDistTime      `xml:"unblockingDate,attr"`
// }

// type QuotaUnsuspensionEvent struct {
// 	ChangeType             string            `xml:"changeType,attr"`
// 	EndOccurrenceTimestamp FileDistTimeStamp `xml:"endOccurrenceTimestamp,attr"`
// 	National               int               `xml:"national,attr"`
// 	OccurrenceTimestamp    FileDistTimeStamp `xml:"occurrenceTimestamp,attr"`
// 	SIDQuotaDefinition     int               `xml:"SIDQuotaDefinition,attr"`
// 	UnsuspensionDate       FileDistTime      `xml:"unsuspensionDate,attr"`
// }

// type QuotaOrderNumberOriginExclusion struct {
// 	GeographicalAreaID  string `xml:"geographicalAreaId,attr"`
// 	National            int    `xml:"national,attr"`
// 	SIDGeographicalArea int    `xml:"SIDGeographicalArea,attr"`
// }
