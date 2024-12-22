package xmltypes

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type MeursingTablePlans []MeursingTablePlan

type MeursingTablePlan struct {
	ChangeType          string       `xml:"changeType,attr"`
	DateStart           FileDistTime `xml:"dateStart,attr"`
	MeursingTablePlanID int          `xml:"meursingTablePlanId,attr"`
	National            int          `xml:"national,attr"`
}

func (plans MeursingTablePlans) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO meursing_table_plan (
		meursing_table_plan_id, date_start, national, change_type
	)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (meursing_table_plan_id) DO UPDATE 
	SET date_start = EXCLUDED.date_start,
		national = EXCLUDED.national,
		change_type = EXCLUDED.change_type;
	`

	deleteQuery := `
	DELETE FROM meursing_table_plan
	WHERE meursing_table_plan_id = $1;
	`

	batch := &pgx.Batch{}

	for i, plan := range plans {
		switch plan.ChangeType {
		case "U": // Insert or update
			batch.Queue(insertQuery, plan.MeursingTablePlanID, plan.DateStart, plan.National, plan.ChangeType)

		case "D": // Delete
			batch.Queue(deleteQuery, plan.MeursingTablePlanID)

		default:
			return fmt.Errorf("unknown ChangeType: %s for MeursingTablePlanID: %d", plan.ChangeType, plan.MeursingTablePlanID)
		}

		if (i+1)%batchSize == 0 || i == len(plans)-1 {
			if err := conn.SendBatch(ctx, batch).Close(); err != nil {
				return fmt.Errorf("failed to execute batch: %w", err)
			}
			batch = &pgx.Batch{}
		}
	}

	return nil
}
