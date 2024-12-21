package xmltypes

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type LookupTables []LookupTable

type LookupTable struct {
	SID                    int                     `xml:"SID,attr"`
	ChangeType             string                  `xml:"changeType,attr"`
	DateStart              FileDistTime            `xml:"dateStart,attr"`
	Interpolate            bool                    `xml:"interpolate,attr"`
	MaxInterval            int                     `xml:"maxInterval,attr"`
	MinInterval            int                     `xml:"minInterval,attr"`
	TableID                string                  `xml:"tableId,attr"`
	LookupTableItem        LookupTableItems        `xml:"lookupTableItem"`
	LookupTableDescription LookupTableDescriptions `xml:"lookupTableDescription"`
}

func (tables LookupTables) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO lookup_table (sid, table_id, change_type, date_start, interpolate, max_interval, min_interval)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (sid) DO UPDATE 
	SET table_id = EXCLUDED.table_id,
		change_type = EXCLUDED.change_type,
		date_start = EXCLUDED.date_start,
		interpolate = EXCLUDED.interpolate,
		max_interval = EXCLUDED.max_interval,
		min_interval = EXCLUDED.min_interval;
	`

	deleteQuery := `
	DELETE FROM lookup_table
	WHERE sid = $1;
	`

	batch := &pgx.Batch{}

	for i, table := range tables {
		switch table.ChangeType {
		case "U": // Insert or update
			batch.Queue(insertQuery, table.SID, table.TableID, table.ChangeType, table.DateStart, table.Interpolate, table.MaxInterval, table.MinInterval)

			// Queue child items
			if len(table.LookupTableItem) > 0 {
				if err := table.LookupTableItem.QueueBatch(ctx, batch, table.SID); err != nil {
					return fmt.Errorf("failed to queue lookup table items for SID %d: %w", table.SID, err)
				}
			}

			// Queue child descriptions
			if len(table.LookupTableDescription) > 0 {
				if err := table.LookupTableDescription.QueueBatch(ctx, batch, table.SID); err != nil {
					return fmt.Errorf("failed to queue lookup table descriptions for SID %d: %w", table.SID, err)
				}
			}

		case "D": // Delete
			batch.Queue(deleteQuery, table.SID)

		default:
			return fmt.Errorf("unknown ChangeType: %s for SID: %d", table.ChangeType, table.SID)
		}

		if (i+1)%batchSize == 0 || i == len(tables)-1 {
			if err := conn.SendBatch(ctx, batch).Close(); err != nil {
				return fmt.Errorf("failed to execute batch: %w", err)
			}
			batch = &pgx.Batch{}
		}
	}

	return nil
}

type LookupTableItems []LookupTableItem

type LookupTableItem struct {
	ParentSID int     // added type
	Threshold float64 `xml:"threshold,attr"`
	Value     float64 `xml:"value,attr"`
}

func (items LookupTableItems) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO lookup_table_item (parent_sid, threshold, value)
	VALUES ($1, $2, $3)
	ON CONFLICT (parent_sid, threshold, value) DO NOTHING;
	`

	for _, item := range items {
		item.ParentSID = parentSID
		batch.Queue(insertQuery, item.ParentSID, item.Threshold, item.Value)
	}
	return nil
}

type LookupTableDescriptions []LookupTableDescription

type LookupTableDescription struct {
	ParentSID   int    // added type
	Description string `xml:"description,attr"`
	LanguageID  string `xml:"languageId,attr"`
}

func (descriptions LookupTableDescriptions) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO lookup_table_description (parent_sid, description, language_id)
	VALUES ($1, $2, $3)
	ON CONFLICT (parent_sid, language_id) DO UPDATE 
	SET description = EXCLUDED.description;
	`

	for _, desc := range descriptions {
		desc.ParentSID = parentSID
		batch.Queue(insertQuery, desc.ParentSID, desc.Description, desc.LanguageID)
	}
	return nil
}
