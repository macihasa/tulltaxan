package xmltypes

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type ExportRefundNomenclatures []ExportRefundNomenclature

type ExportRefundNomenclature struct {
	AdditionalCodeType                           int                                          `xml:"additionalCodeType,attr"`
	ChangeType                                   string                                       `xml:"changeType,attr"`
	DateEnd                                      FileDistTime                                 `xml:"dateEnd,attr"`
	DateStart                                    FileDistTime                                 `xml:"dateStart,attr"`
	ExportRefundCode                             int                                          `xml:"exportRefundCode,attr"`
	GoodsNomenclatureCode                        int                                          `xml:"goodsNomenclatureCode,attr"`
	National                                     int                                          `xml:"national,attr"`
	ProductLineSuffix                            int                                          `xml:"productLineSuffix,attr"`
	SID                                          int                                          `xml:"SID,attr"`
	SIDGoodsNomenclature                         int                                          `xml:"SIDGoodsNomenclature,attr"`
	ExportRefundNomenclatureIndents              ExportRefundNomenclatureIndents              `xml:"exportRefundNomenclatureIndent"`
	ExportRefundNomenclatureDescriptionPeriods   ExportRefundNomenclatureDescriptionPeriods   `xml:"exportRefundNomenclatureDescriptionPeriod"`
	ExportRefundNomenclatureFootnoteAssociations ExportRefundNomenclatureFootnoteAssociations `xml:"exportRefundNomenclatureFootnoteAssociation"`
}

func (nomenclatures ExportRefundNomenclatures) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO export_refund_nomenclature (sid, goods_nomenclature_code, additional_code_type, export_refund_code, product_line_suffix, sid_goods_nomenclature, change_type, date_start, date_end, national)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	ON CONFLICT (sid) DO UPDATE 
	SET goods_nomenclature_code = EXCLUDED.goods_nomenclature_code,
		additional_code_type = EXCLUDED.additional_code_type,
		export_refund_code = EXCLUDED.export_refund_code,
		product_line_suffix = EXCLUDED.product_line_suffix,
		sid_goods_nomenclature = EXCLUDED.sid_goods_nomenclature,
		change_type = EXCLUDED.change_type,
		date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		national = EXCLUDED.national;
	`

	deleteQuery := `
	DELETE FROM export_refund_nomenclature
	WHERE sid = $1;
	`

	batch := &pgx.Batch{}

	for i, nomenclature := range nomenclatures {
		switch nomenclature.ChangeType {
		case "U": // Insert or update
			batch.Queue(insertQuery, nomenclature.SID, fmt.Sprint(nomenclature.GoodsNomenclatureCode), fmt.Sprint(nomenclature.AdditionalCodeType), nomenclature.ExportRefundCode, nomenclature.ProductLineSuffix, nomenclature.SIDGoodsNomenclature, nomenclature.ChangeType, nomenclature.DateStart, nomenclature.DateEnd, nomenclature.National)

			// Queue child elements
			if len(nomenclature.ExportRefundNomenclatureIndents) > 0 {
				if err := nomenclature.ExportRefundNomenclatureIndents.QueueBatch(ctx, batch, nomenclature.SID); err != nil {
					return fmt.Errorf("failed to queue indents for SID %d: %w", nomenclature.SID, err)
				}
			}

			if len(nomenclature.ExportRefundNomenclatureDescriptionPeriods) > 0 {
				if err := nomenclature.ExportRefundNomenclatureDescriptionPeriods.QueueBatch(ctx, batch, nomenclature.SID); err != nil {
					return fmt.Errorf("failed to queue description periods for SID %d: %w", nomenclature.SID, err)
				}
			}

			if len(nomenclature.ExportRefundNomenclatureFootnoteAssociations) > 0 {
				if err := nomenclature.ExportRefundNomenclatureFootnoteAssociations.QueueBatch(ctx, batch, nomenclature.SID); err != nil {
					return fmt.Errorf("failed to queue footnote associations for SID %d: %w", nomenclature.SID, err)
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

type ExportRefundNomenclatureIndents []ExportRefundNomenclatureIndent

type ExportRefundNomenclatureIndent struct {
	ParentSID       int          // added Parent SID
	DateStart       FileDistTime `xml:"dateStart,attr"`
	National        int          `xml:"national,attr"`
	QuantityIndents int          `xml:"quantityIndents,attr"`
	SID             int          `xml:"SID,attr"`
}

func (indents ExportRefundNomenclatureIndents) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO export_refund_nomenclature_indent (sid, parent_sid, date_start, quantity_indents, national)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (sid) DO UPDATE 
	SET date_start = EXCLUDED.date_start,
		quantity_indents = EXCLUDED.quantity_indents,
		national = EXCLUDED.national;
	`

	for _, indent := range indents {
		indent.ParentSID = parentSID
		batch.Queue(insertQuery, indent.SID, indent.ParentSID, indent.DateStart, indent.QuantityIndents, indent.National)
	}
	return nil
}

type ExportRefundNomenclatureDescriptionPeriods []ExportRefundNomenclatureDescriptionPeriod

type ExportRefundNomenclatureDescriptionPeriod struct {
	ParentSID                            int                                  // added Parent SID
	DateStart                            FileDistTime                         `xml:"dateStart,attr"`
	National                             int                                  `xml:"national,attr"`
	SID                                  int                                  `xml:"SID,attr"`
	ExportRefundNomenclatureDescriptions ExportRefundNomenclatureDescriptions `xml:"exportRefundNomenclatureDescription"`
}

func (periods ExportRefundNomenclatureDescriptionPeriods) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO export_refund_nomenclature_description_period (sid, parent_sid, date_start, national)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (sid) DO UPDATE 
	SET date_start = EXCLUDED.date_start,
		national = EXCLUDED.national;
	`

	for _, period := range periods {
		period.ParentSID = parentSID
		batch.Queue(insertQuery, period.SID, period.ParentSID, period.DateStart, period.National)

		// Queue child descriptions
		if len(period.ExportRefundNomenclatureDescriptions) > 0 {
			if err := period.ExportRefundNomenclatureDescriptions.QueueBatch(ctx, batch, period.SID); err != nil {
				return fmt.Errorf("failed to queue descriptions for SID %d: %w", period.SID, err)
			}
		}
	}
	return nil
}

type ExportRefundNomenclatureDescriptions []ExportRefundNomenclatureDescription

type ExportRefundNomenclatureDescription struct {
	ParentSID   int    // added Parent SID
	Description string `xml:"description,attr"`
	LanguageID  string `xml:"languageId,attr"`
	National    int    `xml:"national,attr"`
}

func (descriptions ExportRefundNomenclatureDescriptions) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO export_refund_nomenclature_description (parent_sid, description, language_id, national)
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

type ExportRefundNomenclatureFootnoteAssociations []ExportRefundNomenclatureFootnoteAssociation

type ExportRefundNomenclatureFootnoteAssociation struct {
	ParentSID    int          // added Parent SID
	DateEnd      FileDistTime `xml:"dateEnd,attr"`
	DateStart    FileDistTime `xml:"dateStart,attr"`
	FootnoteID   int          `xml:"footnoteId,attr"`
	FootnoteType string       `xml:"footnoteType,attr"`
	National     int          `xml:"national,attr"`
}

func (associations ExportRefundNomenclatureFootnoteAssociations) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO export_refund_nomenclature_footnote_association (parent_sid, date_start, date_end, footnote_id, footnote_type, national)
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
