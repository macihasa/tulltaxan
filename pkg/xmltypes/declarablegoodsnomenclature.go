package xmltypes

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type DeclarableGoodsNomenclatures []DeclarableGoodsNomenclature

type DeclarableGoodsNomenclature struct {
	ChangeType            string       `xml:"changeType,attr"`
	DateEnd               FileDistTime `xml:"dateEnd,attr"`
	DateStart             FileDistTime `xml:"dateStart,attr"`
	GoodsNomenclatureCode int          `xml:"goodsNomenclatureCode,attr"`
	Type                  string       `xml:"type,attr"`
}

func (nomenclatures DeclarableGoodsNomenclatures) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO declarable_goods_nomenclature (goods_nomenclature_code, change_type, date_start, date_end, type)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (goods_nomenclature_code) DO UPDATE 
	SET change_type = EXCLUDED.change_type,
		date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		type = EXCLUDED.type;
	`

	deleteQuery := `
	DELETE FROM declarable_goods_nomenclature
	WHERE goods_nomenclature_code = $1;
	`

	batch := &pgx.Batch{}

	for i, nomenclature := range nomenclatures {
		switch nomenclature.ChangeType {
		case "U": // Insert or update
			batch.Queue(insertQuery, fmt.Sprint(nomenclature.GoodsNomenclatureCode), nomenclature.ChangeType, nomenclature.DateStart, nomenclature.DateEnd, nomenclature.Type)

		case "D": // Delete
			batch.Queue(deleteQuery, fmt.Sprint(nomenclature.GoodsNomenclatureCode))

		default:
			return fmt.Errorf("unknown ChangeType: %s for GoodsNomenclatureCode: %d", nomenclature.ChangeType, nomenclature.GoodsNomenclatureCode)
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
