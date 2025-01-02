package xmltypes

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type BaseRegulations []BaseRegulation

type BaseRegulation struct {
	AntidumpingRegulationID       *string      `xml:"antidumpingRegulationId,attr"`
	AntidumpingRegulationRoleType *int         `xml:"antidumpingRegulationRoleType,attr"`
	ChangeType                    string       `xml:"changeType,attr"`
	CommunityCode                 int          `xml:"communityCode,attr"`
	DateEnd                       FileDistTime `xml:"dateEnd,attr"`
	DatePublished                 FileDistTime `xml:"datePublished,attr"`
	DateStart                     FileDistTime `xml:"dateStart,attr"`
	Description                   *string      `xml:"description,attr"`
	EffectiveEndDate              FileDistTime `xml:"effectiveEndDate,attr"`
	JournalPage                   *int         `xml:"journalPage,attr"`
	National                      int          `xml:"national,attr"`
	OfficialJournalID             *string      `xml:"officialJournalId,attr"`
	RegulationApprovedFlag        int          `xml:"regulationApprovedFlag,attr"`
	RegulationGroup               string       `xml:"regulationGroup,attr"`
	RegulationID                  string       `xml:"regulationId,attr"`
	RegulationRoleType            int          `xml:"regulationRoleType,attr"`
	ReplacementIndicator          int          `xml:"replacementIndicator,attr"`
	StoppedFlag                   int          `xml:"stoppedFlag,attr"`
	Url                           *string      `xml:"url,attr"`
}

func (regulations BaseRegulations) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO base_regulation (
		regulation_id, regulation_role_type, antidumping_regulation_id, antidumping_regulation_role_type, change_type,
		community_code, date_end, date_published, date_start, description, effective_end_date, journal_page,
		national, official_journal_id, regulation_approved_flag, regulation_group, replacement_indicator, stopped_flag, url
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	ON CONFLICT (regulation_id) DO UPDATE 
	SET antidumping_regulation_id = EXCLUDED.antidumping_regulation_id,
		regulation_role_type = EXCLUDED.regulation_role_type,
		antidumping_regulation_role_type = EXCLUDED.antidumping_regulation_role_type,
		change_type = EXCLUDED.change_type,
		community_code = EXCLUDED.community_code,
		date_end = EXCLUDED.date_end,
		date_published = EXCLUDED.date_published,
		date_start = EXCLUDED.date_start,
		description = EXCLUDED.description,
		effective_end_date = EXCLUDED.effective_end_date,
		journal_page = EXCLUDED.journal_page,
		national = EXCLUDED.national,
		official_journal_id = EXCLUDED.official_journal_id,
		regulation_approved_flag = EXCLUDED.regulation_approved_flag,
		regulation_group = EXCLUDED.regulation_group,
		replacement_indicator = EXCLUDED.replacement_indicator,
		stopped_flag = EXCLUDED.stopped_flag,
		url = EXCLUDED.url;
	`

	deleteQuery := `
	DELETE FROM base_regulation
	WHERE regulation_id = $1;
	`

	batch := &pgx.Batch{}

	for i, reg := range regulations {
		switch reg.ChangeType {
		case "U": // Insert or update
			batch.Queue(insertQuery, reg.RegulationID, reg.RegulationRoleType, reg.AntidumpingRegulationID, reg.AntidumpingRegulationRoleType,
				reg.ChangeType, reg.CommunityCode, reg.DateEnd, reg.DatePublished, reg.DateStart, reg.Description, reg.EffectiveEndDate,
				reg.JournalPage, reg.National, reg.OfficialJournalID, reg.RegulationApprovedFlag, reg.RegulationGroup, reg.ReplacementIndicator,
				reg.StoppedFlag, reg.Url)

		case "D": // Delete
			batch.Queue(deleteQuery, reg.RegulationID)
		default:
			return fmt.Errorf("unknown ChangeType: %s for RegulationID: %s", reg.ChangeType, reg.RegulationID)
		}

		if (i+1)%batchSize == 0 || i == len(regulations)-1 {
			if err := conn.SendBatch(ctx, batch).Close(); err != nil {
				return fmt.Errorf("failed to execute batch: %w", err)
			}
			batch = &pgx.Batch{}
		}
	}

	return nil
}

type ModificationRegulations []ModificationRegulation

type ModificationRegulation struct {
	BaseRegulationID               string       `xml:"baseRegulationId,attr"`
	BaseRegulationRoleType         int          `xml:"baseRegulationRoleType,attr"`
	ChangeType                     string       `xml:"changeType,attr"`
	DateEnd                        FileDistTime `xml:"dateEnd,attr"`
	DatePublished                  FileDistTime `xml:"datePublished,attr"`
	DateStart                      FileDistTime `xml:"dateStart,attr"`
	Description                    *string      `xml:"description,attr"`
	EffectiveEndDate               FileDistTime `xml:"effectiveEndDate,attr"`
	JournalPage                    *int         `xml:"journalPage,attr"`
	ModificationRegulationID       string       `xml:"modificationRegulationId,attr"`
	ModificationRegulationRoleType int          `xml:"modificationRegulationRoleType,attr"`
	National                       int          `xml:"national,attr"`
	OfficialJournalID              *string      `xml:"officialJournalId,attr"`
	RegulationApprovedFlag         int          `xml:"regulationApprovedFlag,attr"`
	ReplacementIndicator           int          `xml:"replacementIndicator,attr"`
	StoppedFlag                    int          `xml:"stoppedFlag,attr"`
}

func (regs ModificationRegulations) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO modification_regulation (
		modification_regulation_id, modification_regulation_role_type, base_regulation_id, base_regulation_role_type,
		change_type, date_end, date_published, date_start, description, effective_end_date, journal_page,
		national, official_journal_id, regulation_approved_flag, replacement_indicator, stopped_flag
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	ON CONFLICT (modification_regulation_id) DO UPDATE 
	SET base_regulation_id = EXCLUDED.base_regulation_id,
		modification_regulation_role_type = EXCLUDED.modification_regulation_role_type,
		base_regulation_role_type = EXCLUDED.base_regulation_role_type,
		change_type = EXCLUDED.change_type,
		date_end = EXCLUDED.date_end,
		date_published = EXCLUDED.date_published,
		date_start = EXCLUDED.date_start,
		description = EXCLUDED.description,
		effective_end_date = EXCLUDED.effective_end_date,
		journal_page = EXCLUDED.journal_page,
		national = EXCLUDED.national,
		official_journal_id = EXCLUDED.official_journal_id,
		regulation_approved_flag = EXCLUDED.regulation_approved_flag,
		replacement_indicator = EXCLUDED.replacement_indicator,
		stopped_flag = EXCLUDED.stopped_flag;
	`

	batch := &pgx.Batch{}

	fmt.Print("Inserting modification regulations\n")
	for i, reg := range regs {
		batch.Queue(insertQuery, reg.ModificationRegulationID, reg.ModificationRegulationRoleType, reg.BaseRegulationID, reg.BaseRegulationRoleType,
			reg.ChangeType, reg.DateEnd, reg.DatePublished, reg.DateStart, reg.Description, reg.EffectiveEndDate, reg.JournalPage,
			reg.National, reg.OfficialJournalID, reg.RegulationApprovedFlag, reg.ReplacementIndicator, reg.StoppedFlag)

		if (i+1)%batchSize == 0 || i == len(regs)-1 {
			if err := conn.SendBatch(ctx, batch).Close(); err != nil {
				return fmt.Errorf("failed to execute batch: %w", err)
			}
			batch = &pgx.Batch{}
		}
	}

	return nil
}

type FullTemporaryStopRegulations []FullTemporaryStopRegulation

type FullTemporaryStopRegulation struct {
	ChangeType                         string                             `xml:"changeType,attr"`
	DateEnd                            FileDistTime                       `xml:"dateEnd,attr"`
	DatePublished                      FileDistTime                       `xml:"datePublished,attr"`
	DateStart                          FileDistTime                       `xml:"dateStart,attr"`
	Description                        string                             `xml:"description,attr"`
	EffectiveEndDate                   FileDistTime                       `xml:"effectiveEndDate,attr"`
	FtsRegulationID                    string                             `xml:"ftsRegulationId,attr"`
	FtsRegulationRoleType              int                                `xml:"ftsRegulationRoleType,attr"`
	JournalPage                        *int                               `xml:"journalPage,attr"`
	National                           int                                `xml:"national,attr"`
	OfficialJournalID                  *string                            `xml:"officialJournalId,attr"`
	RegulationApprovedFlag             int                                `xml:"regulationApprovedFlag,attr"`
	ReplacementIndicator               int                                `xml:"replacementIndicator,attr"`
	StoppedFlag                        int                                `xml:"stoppedFlag,attr"`
	FullTemporaryStopRegulationActions FullTemporaryStopRegulationActions `xml:"fullTemporaryStopRegulationAction"`
}

type FullTemporaryStopRegulationActions []FullTemporaryStopRegulationAction

type FullTemporaryStopRegulationAction struct {
	National                  int    `xml:"national,attr"`
	StoppedRegulationID       string `xml:"stoppedRegulationId,attr"`
	StoppedRegulationRoleType int    `xml:"stoppedRegulationRoleType,attr"`
}

func (regs FullTemporaryStopRegulations) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO full_temporary_stop_regulation (
		fts_regulation_id, fts_regulation_role_type, change_type, date_start, date_end, date_published,
		description, effective_end_date, journal_page, national, official_journal_id, regulation_approved_flag,
		replacement_indicator, stopped_flag
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	ON CONFLICT (fts_regulation_id) DO UPDATE 
	SET change_type = EXCLUDED.change_type,
		fts_regulation_role_type = EXCLUDED.fts_regulation_role_type,
		date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		date_published = EXCLUDED.date_published,
		description = EXCLUDED.description,
		effective_end_date = EXCLUDED.effective_end_date,
		journal_page = EXCLUDED.journal_page,
		national = EXCLUDED.national,
		official_journal_id = EXCLUDED.official_journal_id,
		regulation_approved_flag = EXCLUDED.regulation_approved_flag,
		replacement_indicator = EXCLUDED.replacement_indicator,
		stopped_flag = EXCLUDED.stopped_flag;
	`

	actionInsertQuery := `
	INSERT INTO full_temporary_stop_regulation_action (
		fts_regulation_id, stopped_regulation_id, stopped_regulation_role_type, national
	)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (fts_regulation_id, stopped_regulation_id, stopped_regulation_role_type) DO UPDATE 
	SET national = EXCLUDED.national;
	`

	batch := &pgx.Batch{}

	for i, reg := range regs {
		batch.Queue(insertQuery, reg.FtsRegulationID, reg.FtsRegulationRoleType, reg.ChangeType, reg.DateStart, reg.DateEnd,
			reg.DatePublished, reg.Description, reg.EffectiveEndDate, reg.JournalPage, reg.National, reg.OfficialJournalID,
			reg.RegulationApprovedFlag, reg.ReplacementIndicator, reg.StoppedFlag)

		// Queue child actions
		for _, action := range reg.FullTemporaryStopRegulationActions {
			batch.Queue(actionInsertQuery, reg.FtsRegulationID, action.StoppedRegulationID, action.StoppedRegulationRoleType, action.National)
		}

		if (i+1)%batchSize == 0 || i == len(regs)-1 {
			if err := conn.SendBatch(ctx, batch).Close(); err != nil {
				return fmt.Errorf("failed to execute batch: %w", err)
			}
			batch = &pgx.Batch{}
		}
	}

	return nil
}
