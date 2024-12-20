package xmltypes

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type GeographicalAreas []GeographicalArea

type GeographicalArea struct {
	SID                                int                                `xml:"SID,attr"`
	SIDParentGroup                     int                                `xml:"SIDParentGroup,attr"`
	ChangeType                         string                             `xml:"changeType,attr"`
	DateEnd                            FileDistTime                       `xml:"dateEnd,attr"`
	DateStart                          FileDistTime                       `xml:"dateStart,attr"`
	GeographicalAreaCode               int                                `xml:"geographicalAreaCode,attr"`
	GeographicalAreaID                 string                             `xml:"geographicalAreaId,attr"`
	National                           int                                `xml:"national,attr"`
	GeographicalAreaMemberships        GeographicalAreaMemberships        `xml:"geographicalAreaMembership"`
	GeographicalAreaDescriptionPeriods GeographicalAreaDescriptionPeriods `xml:"geographicalAreaDescriptionPeriod"`
}

func (areas GeographicalAreas) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO geographical_area (sid, sid_parent_group, change_type, date_start, date_end, geographical_area_code, geographical_area_id, national)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT (sid) DO UPDATE 
	SET sid_parent_group = EXCLUDED.sid_parent_group,
		change_type = EXCLUDED.change_type,
		date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		geographical_area_code = EXCLUDED.geographical_area_code,
		geographical_area_id = EXCLUDED.geographical_area_id,
		national = EXCLUDED.national;
	`

	deleteQuery := `
	DELETE FROM geographical_area
	WHERE sid = $1;
	`

	batch := &pgx.Batch{}

	for i, area := range areas {
		switch area.ChangeType {
		case "U": // Insert or update
			batch.Queue(insertQuery, area.SID, area.SIDParentGroup, area.ChangeType, area.DateStart, area.DateEnd, area.GeographicalAreaCode, area.GeographicalAreaID, area.National)

			// Queue child memberships
			if len(area.GeographicalAreaMemberships) > 0 {
				if err := area.GeographicalAreaMemberships.QueueBatch(ctx, batch, area.SID); err != nil {
					return fmt.Errorf("failed to queue memberships for SID %d: %w", area.SID, err)
				}
			}

			// Queue child description periods
			if len(area.GeographicalAreaDescriptionPeriods) > 0 {
				if err := area.GeographicalAreaDescriptionPeriods.QueueBatch(ctx, batch, area.SID); err != nil {
					return fmt.Errorf("failed to queue description periods for SID %d: %w", area.SID, err)
				}
			}

		case "D": // Delete
			batch.Queue(deleteQuery, area.SID)

		default:
			return fmt.Errorf("unknown ChangeType: %s for SID: %d", area.ChangeType, area.SID)
		}

		if (i+1)%batchSize == 0 || i == len(areas)-1 {
			if err := conn.SendBatch(ctx, batch).Close(); err != nil {
				return fmt.Errorf("failed to execute batch: %w", err)
			}
			batch = &pgx.Batch{}
		}
	}

	return nil
}

type GeographicalAreaMemberships []GeographicalAreaMembership

type GeographicalAreaMembership struct {
	ParentSID                int          // added type
	DateEnd                  FileDistTime `xml:"dateEnd,attr"`
	DateStart                FileDistTime `xml:"dateStart,attr"`
	National                 int          `xml:"national,attr"`
	SIDGeographicalAreaGroup int          `xml:"SIDGeographicalAreaGroup,attr"`
}

func (memberships GeographicalAreaMemberships) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO geographical_area_membership (sid_geographical_area_group, parent_sid, date_start, date_end, national)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (sid_geographical_area_group, parent_sid, date_start) DO UPDATE 
	SET date_end = EXCLUDED.date_end,
		national = EXCLUDED.national;
	`

	for _, membership := range memberships {
		membership.ParentSID = parentSID
		batch.Queue(insertQuery, membership.SIDGeographicalAreaGroup, membership.ParentSID, membership.DateStart, membership.DateEnd, membership.National)
	}
	return nil
}

type GeographicalAreaDescriptionPeriods []GeographicalAreaDescriptionPeriod

type GeographicalAreaDescriptionPeriod struct {
	SID                          int                          `xml:"SID,attr"`
	ParentSID                    int                          // added type
	DateEnd                      FileDistTime                 `xml:"dateEnd,attr"`
	DateStart                    FileDistTime                 `xml:"dateStart,attr"`
	National                     int                          `xml:"national,attr"`
	GeographicalAreaDescriptions GeographicalAreaDescriptions `xml:"geographicalAreaDescription"`
}

func (periods GeographicalAreaDescriptionPeriods) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO geographical_area_description_period (sid, parent_sid, date_start, date_end, national)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (sid) DO UPDATE 
	SET date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		national = EXCLUDED.national;
	`

	for _, period := range periods {
		period.ParentSID = parentSID
		batch.Queue(insertQuery, period.SID, period.ParentSID, period.DateStart, period.DateEnd, period.National)

		// Queue child descriptions
		if len(period.GeographicalAreaDescriptions) > 0 {
			if err := period.GeographicalAreaDescriptions.QueueBatch(ctx, batch, period.SID); err != nil {
				return fmt.Errorf("failed to queue descriptions for SID %d: %w", period.SID, err)
			}
		}
	}
	return nil
}

type GeographicalAreaDescriptions []GeographicalAreaDescription

type GeographicalAreaDescription struct {
	ParentSID   int    // added type
	Description string `xml:"description,attr"`
	LanguageID  string `xml:"languageId,attr"`
	National    int    `xml:"national,attr"`
}

func (descriptions GeographicalAreaDescriptions) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO geographical_area_description (parent_sid, description, language_id, national)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (parent_sid, language_id) DO UPDATE 
	SET description = EXCLUDED.description,
		national = EXCLUDED.national;
	`

	for _, desc := range descriptions {
		desc.ParentSID = parentSID
		batch.Queue(insertQuery, desc.ParentSID, desc.Description, desc.LanguageID, desc.National)
	}
	return nil
}
