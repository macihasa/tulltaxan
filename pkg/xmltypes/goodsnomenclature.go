package xmltypes

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type GoodsNomenclatures []GoodsNomenclature

type GoodsNomenclature struct {
	SID                                   int                                   `xml:"SID,attr"`
	ChangeType                            string                                `xml:"changeType,attr"`
	DateEnd                               FileDistTime                          `xml:"dateEnd,attr"`
	DateStart                             FileDistTime                          `xml:"dateStart,attr"`
	GoodsNomenclatureCode                 int                                   `xml:"goodsNomenclatureCode,attr"`
	National                              int                                   `xml:"national,attr"`
	ProductLineSuffix                     int                                   `xml:"productLineSuffix,attr"`
	StatisticalIndicator                  int                                   `xml:"statisticalIndicator,attr"`
	GoodsNomenclatureIndents              GoodsNomenclatureIndents              `xml:"goodsNomenclatureIndent"`
	GoodsNomenclatureDescriptionPeriods   GoodsNomenclatureDescriptionPeriods   `xml:"goodsNomenclatureDescriptionPeriod"`
	GoodsNomenclatureFootnoteAssociations GoodsNomenclatureFootnoteAssociations `xml:"goodsNomenclatureFootnoteAssociation"`
	GoodsNomenclatureGroupMemberships     GoodsNomenclatureGroupMemberships     `xml:"goodsNomenclatureGroupMembership"`
}

func (nomenclatures GoodsNomenclatures) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO goods_nomenclature (sid, goods_nomenclature_code, product_line_suffix, statistical_indicator, change_type, date_start, date_end, national)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT (sid) DO UPDATE 
	SET goods_nomenclature_code = EXCLUDED.goods_nomenclature_code,
		product_line_suffix = EXCLUDED.product_line_suffix,
		statistical_indicator = EXCLUDED.statistical_indicator,
		change_type = EXCLUDED.change_type,
		date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		national = EXCLUDED.national;
	`

	deleteQuery := `
	DELETE FROM goods_nomenclature
	WHERE sid = $1;
	`

	batch := &pgx.Batch{}

	for i, nomenclature := range nomenclatures {
		switch nomenclature.ChangeType {
		case "U": // Insert or update
			batch.Queue(insertQuery, nomenclature.SID, fmt.Sprint(nomenclature.GoodsNomenclatureCode), nomenclature.ProductLineSuffix, nomenclature.StatisticalIndicator, nomenclature.ChangeType, nomenclature.DateStart, nomenclature.DateEnd, nomenclature.National)

			// Queue child elements
			if len(nomenclature.GoodsNomenclatureIndents) > 0 {
				if err := nomenclature.GoodsNomenclatureIndents.QueueBatch(ctx, batch, nomenclature.SID); err != nil {
					return fmt.Errorf("failed to queue indents for SID %d: %w", nomenclature.SID, err)
				}
			}
			if len(nomenclature.GoodsNomenclatureDescriptionPeriods) > 0 {
				if err := nomenclature.GoodsNomenclatureDescriptionPeriods.QueueBatch(ctx, batch, nomenclature.SID); err != nil {
					return fmt.Errorf("failed to queue description periods for SID %d: %w", nomenclature.SID, err)
				}
			}
			if len(nomenclature.GoodsNomenclatureFootnoteAssociations) > 0 {
				if err := nomenclature.GoodsNomenclatureFootnoteAssociations.QueueBatch(ctx, batch, nomenclature.SID); err != nil {
					return fmt.Errorf("failed to queue footnote associations for SID %d: %w", nomenclature.SID, err)
				}
			}
			if len(nomenclature.GoodsNomenclatureGroupMemberships) > 0 {
				if err := nomenclature.GoodsNomenclatureGroupMemberships.QueueBatch(ctx, batch, nomenclature.SID); err != nil {
					return fmt.Errorf("failed to queue group memberships for SID %d: %w", nomenclature.SID, err)
				}
			}

		case "D": // Delete
			batch.Queue(deleteQuery, nomenclature.SID)

		default:
			return fmt.Errorf("unknown ChangeType: %s for SID: %d", nomenclature.ChangeType, nomenclature.SID)
		}

		if (i+1)%batchSize == 0 || i == len(nomenclatures)-1 {
			if err := conn.SendBatch(ctx, batch).Close(); err != nil {
				return fmt.Errorf("failed to execute batch: %w", err)
			}
			batch = &pgx.Batch{}
		}
	}

	return nil
}

type GoodsNomenclatureIndents []GoodsNomenclatureIndent

type GoodsNomenclatureIndent struct {
	SID             int          `xml:"SID,attr"`
	ParentSID       int          // addded type
	DateEnd         FileDistTime `xml:"dateEnd,attr"`
	DateStart       FileDistTime `xml:"dateStart,attr"`
	National        int          `xml:"national,attr"`
	QuantityIndents int          `xml:"quantityIndents,attr"`
}

func (indents GoodsNomenclatureIndents) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO goods_nomenclature_indent (sid, parent_sid, date_start, date_end, quantity_indents, national)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (sid) DO UPDATE 
	SET date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		quantity_indents = EXCLUDED.quantity_indents,
		national = EXCLUDED.national;
	`

	for _, indent := range indents {
		indent.ParentSID = parentSID
		batch.Queue(insertQuery, indent.SID, indent.ParentSID, indent.DateStart, indent.DateEnd, indent.QuantityIndents, indent.National)
	}
	return nil
}

type GoodsNomenclatureDescriptionPeriods []GoodsNomenclatureDescriptionPeriod

type GoodsNomenclatureDescriptionPeriod struct {
	SID                           int                           `xml:"SID,attr"`
	ParentSID                     int                           // added type
	DateEnd                       FileDistTime                  `xml:"dateEnd,attr"`
	DateStart                     FileDistTime                  `xml:"dateStart,attr"`
	National                      int                           `xml:"national,attr"`
	GoodsNomenclatureDescriptions GoodsNomenclatureDescriptions `xml:"goodsNomenclatureDescription"`
}

func (periods GoodsNomenclatureDescriptionPeriods) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO goods_nomenclature_description_period (sid, parent_sid, date_start, date_end, national)
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
		if len(period.GoodsNomenclatureDescriptions) > 0 {
			if err := period.GoodsNomenclatureDescriptions.QueueBatch(ctx, batch, period.SID); err != nil {
				return fmt.Errorf("failed to queue descriptions for SID %d: %w", period.SID, err)
			}
		}
	}
	return nil
}

type GoodsNomenclatureDescriptions []GoodsNomenclatureDescription

type GoodsNomenclatureDescription struct {
	ParentSID   int    // added type
	Description string `xml:"description,attr"`
	LanguageID  string `xml:"languageId,attr"`
	National    int    `xml:"national,attr"`
}

func (descriptions GoodsNomenclatureDescriptions) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO goods_nomenclature_description (parent_sid, description, language_id, national)
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

type GoodsNomenclatureFootnoteAssociations []GoodsNomenclatureFootnoteAssociation

type GoodsNomenclatureFootnoteAssociation struct {
	ParentSID    int          // added type
	DateEnd      FileDistTime `xml:"dateEnd,attr"`
	DateStart    FileDistTime `xml:"dateStart,attr"`
	FootnoteID   int          `xml:"footnoteId,attr"`
	FootnoteType string       `xml:"footnoteType,attr"`
	National     int          `xml:"national,attr"`
}

func (associations GoodsNomenclatureFootnoteAssociations) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO goods_nomenclature_footnote_association (parent_sid, date_start, date_end, footnote_id, footnote_type, national)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (parent_sid, footnote_id, footnote_type) DO UPDATE 
	SET date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		national = EXCLUDED.national;
	`

	for _, assoc := range associations {
		assoc.ParentSID = parentSID
		batch.Queue(insertQuery, assoc.ParentSID, assoc.DateStart, assoc.DateEnd, assoc.FootnoteID, assoc.FootnoteType, assoc.National)
	}
	return nil
}

type GoodsNomenclatureGroupMemberships []GoodsNomenclatureGroupMembership

type GoodsNomenclatureGroupMembership struct {
	ParentSID                  int          // added type
	DateEnd                    FileDistTime `xml:"dateEnd,attr"`
	DateStart                  FileDistTime `xml:"dateStart,attr"`
	GoodsNomenclatureGroupID   string       `xml:"goodsNomenclatureGroupId,attr"`
	GoodsNomenclatureGroupType string       `xml:"goodsNomenclatureGroupType,attr"`
	National                   int          `xml:"national,attr"`
}

func (memberships GoodsNomenclatureGroupMemberships) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO goods_nomenclature_group_membership (parent_sid, date_start, date_end, goods_nomenclature_group_id, goods_nomenclature_group_type, national)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (parent_sid, goods_nomenclature_group_id, goods_nomenclature_group_type) DO UPDATE 
	SET date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		national = EXCLUDED.national;
	`

	for _, membership := range memberships {
		membership.ParentSID = parentSID
		batch.Queue(insertQuery, membership.ParentSID, membership.DateStart, membership.DateEnd, membership.GoodsNomenclatureGroupID, membership.GoodsNomenclatureGroupType, membership.National)
	}
	return nil
}
