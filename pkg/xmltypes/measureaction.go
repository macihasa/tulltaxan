package xmltypes

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type MeasureActions []MeasureAction

type MeasureAction struct {
	ActionCode                string                    `xml:"actionCode,attr"`
	ChangeType                string                    `xml:"changeType,attr"`
	DateEnd                   FileDistTime              `xml:"dateEnd,attr"`
	DateStart                 FileDistTime              `xml:"dateStart,attr"`
	National                  int                       `xml:"national,attr"`
	MeasureActionDescriptions MeasureActionDescriptions `xml:"measureActionDescription"`
}

func (actions MeasureActions) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO measure_action (action_code, change_type, date_start, date_end, national)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (action_code) DO UPDATE 
	SET change_type = EXCLUDED.change_type,
		date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		national = EXCLUDED.national;
	`

	deleteQuery := `
	DELETE FROM measure_action
	WHERE action_code = $1;
	`

	batch := &pgx.Batch{}

	for i, action := range actions {
		switch action.ChangeType {
		case "U": // Insert or update
			batch.Queue(insertQuery, action.ActionCode, action.ChangeType, action.DateStart, action.DateEnd, action.National)

			// Queue child descriptions
			if len(action.MeasureActionDescriptions) > 0 {
				if err := action.MeasureActionDescriptions.QueueBatch(ctx, batch, action.ActionCode); err != nil {
					return fmt.Errorf("failed to queue descriptions for ActionCode %s: %w", action.ActionCode, err)
				}
			}

		case "D": // Delete
			batch.Queue(deleteQuery, action.ActionCode)

		default:
			return fmt.Errorf("unknown ChangeType: %s for ActionCode: %s", action.ChangeType, action.ActionCode)
		}

		if (i+1)%batchSize == 0 || i == len(actions)-1 {
			if err := conn.SendBatch(ctx, batch).Close(); err != nil {
				return fmt.Errorf("failed to execute batch: %w", err)
			}
			batch = &pgx.Batch{}
		}
	}

	return nil
}

type MeasureActionDescriptions []MeasureActionDescription

type MeasureActionDescription struct {
	ParentActionCode string // added type
	Description      string `xml:"description,attr"`
	LanguageID       string `xml:"languageId,attr"`
	National         int    `xml:"national,attr"`
}

func (descriptions MeasureActionDescriptions) QueueBatch(ctx context.Context, batch *pgx.Batch, parentActionCode string) error {
	insertQuery := `
	INSERT INTO measure_action_description (parent_action_code, description, language_id, national)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (parent_action_code, language_id) DO UPDATE 
	SET description = EXCLUDED.description,
		national = EXCLUDED.national;
	`

	for _, desc := range descriptions {
		desc.ParentActionCode = parentActionCode
		batch.Queue(insertQuery, desc.ParentActionCode, desc.Description, desc.LanguageID, desc.National)
	}
	return nil
}
