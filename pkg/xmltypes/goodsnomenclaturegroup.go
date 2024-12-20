package xmltypes

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type GoodsNomenclatureGroups []GoodsNomenclatureGroup

type GoodsNomenclatureGroup struct {
	ChangeType                         string                             `xml:"changeType,attr"`
	DateEnd                            FileDistTime                       `xml:"dateEnd,attr"`
	DateStart                          FileDistTime                       `xml:"dateStart,attr"`
	GoodsNomenclatureGroupID           string                             `xml:"goodsNomenclatureGroupId,attr"`
	GoodsNomenclatureGroupType         string                             `xml:"goodsNomenclatureGroupType,attr"`
	National                           int                                `xml:"national,attr"`
	NomenclatureGroupFacilityCode      int                                `xml:"nomenclatureGroupFacilityCode,attr"`
	GoodsNomenclatureGroupDescriptions GoodsNomenclatureGroupDescriptions `xml:"goodsNomenclatureGroupDescription"`
}

func (groups GoodsNomenclatureGroups) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO goods_nomenclature_group (goods_nomenclature_group_id, goods_nomenclature_group_type, change_type, date_start, date_end, nomenclature_group_facility_code, national)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (goods_nomenclature_group_id, goods_nomenclature_group_type) DO UPDATE 
	SET change_type = EXCLUDED.change_type,
		date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		nomenclature_group_facility_code = EXCLUDED.nomenclature_group_facility_code,
		national = EXCLUDED.national;
	`

	deleteQuery := `
	DELETE FROM goods_nomenclature_group
	WHERE goods_nomenclature_group_id = $1 AND goods_nomenclature_group_type = $2;
	`

	batch := &pgx.Batch{}

	for i, group := range groups {
		switch group.ChangeType {
		case "U": // Insert or update
			batch.Queue(insertQuery, group.GoodsNomenclatureGroupID, group.GoodsNomenclatureGroupType, group.ChangeType, group.DateStart, group.DateEnd, group.NomenclatureGroupFacilityCode, group.National)

			// Queue child descriptions
			if len(group.GoodsNomenclatureGroupDescriptions) > 0 {
				if err := group.GoodsNomenclatureGroupDescriptions.QueueBatch(ctx, batch, group.GoodsNomenclatureGroupID, group.GoodsNomenclatureGroupType); err != nil {
					return fmt.Errorf("failed to queue descriptions for GoodsNomenclatureGroupID %s, GoodsNomenclatureGroupType %s: %w", group.GoodsNomenclatureGroupID, group.GoodsNomenclatureGroupType, err)
				}
			}

		case "D": // Delete
			batch.Queue(deleteQuery, group.GoodsNomenclatureGroupID, group.GoodsNomenclatureGroupType)

		default:
			return fmt.Errorf("unknown ChangeType: %s for GoodsNomenclatureGroupID: %s, GoodsNomenclatureGroupType: %s", group.ChangeType, group.GoodsNomenclatureGroupID, group.GoodsNomenclatureGroupType)
		}

		if (i+1)%batchSize == 0 || i == len(groups)-1 {
			if err := conn.SendBatch(ctx, batch).Close(); err != nil {
				return fmt.Errorf("failed to execute batch: %w", err)
			}
			batch = &pgx.Batch{}
		}
	}

	return nil
}

type GoodsNomenclatureGroupDescriptions []GoodsNomenclatureGroupDescription

type GoodsNomenclatureGroupDescription struct {
	ParentGoodsNomenclatureGroupID   string // added Parent ID
	ParentGoodsNomenclatureGroupType string // added Parent Type
	Description                      string `xml:"description,attr"`
	LanguageID                       string `xml:"languageId,attr"`
	National                         int    `xml:"national,attr"`
}

func (descriptions GoodsNomenclatureGroupDescriptions) QueueBatch(ctx context.Context, batch *pgx.Batch, parentID, parentType string) error {
	insertQuery := `
	INSERT INTO goods_nomenclature_group_description (parent_goods_nomenclature_group_id, parent_goods_nomenclature_group_type, description, language_id, national)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (parent_goods_nomenclature_group_id, parent_goods_nomenclature_group_type, language_id) DO UPDATE 
	SET description = EXCLUDED.description,
		national = EXCLUDED.national;
	`

	for _, desc := range descriptions {
		desc.ParentGoodsNomenclatureGroupID = parentID
		desc.ParentGoodsNomenclatureGroupType = parentType
		batch.Queue(insertQuery, desc.ParentGoodsNomenclatureGroupID, desc.ParentGoodsNomenclatureGroupType, desc.Description, desc.LanguageID, desc.National)
	}
	return nil
}
