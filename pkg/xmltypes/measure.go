package xmltypes

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Measures []Measure

type Measure struct {
	SID                              int                              `xml:"SID,attr"`
	SIDAdditionalCode                *int                             `xml:"SIDAdditionalCode,attr"`
	SIDExportRefundNomenclature      *int                             `xml:"SIDExportRefundNomenclature,attr"`
	SIDGeographicalArea              int                              `xml:"SIDGeographicalArea,attr"`
	SIDGoodsNomenclature             *int                             `xml:"SIDGoodsNomenclature,attr"`
	AdditionalCodeID                 *string                          `xml:"additionalCodeId,attr"`
	AdditionalCodeType               *string                          `xml:"additionalCodeType,attr"`
	ChangeType                       string                           `xml:"changeType,attr"`
	DateEnd                          FileDistTime                     `xml:"dateEnd,attr"`
	DateStart                        FileDistTime                     `xml:"dateStart,attr"`
	Expression                       *string                          `xml:"expression,attr"`
	GeographicalAreaID               string                           `xml:"geographicalAreaId,attr"`
	GoodsNomenclatureCode            *int                             `xml:"goodsNomenclatureCode,attr"`
	JustificationRegulationID        *string                          `xml:"justificationRegulationId,attr"`
	JustificationRegulationRoleType  *int                             `xml:"justificationRegulationRoleType,attr"`
	MeasureType                      string                           `xml:"measureType,attr"`
	National                         int                              `xml:"national,attr"`
	QuotaOrderNumber                 *int                             `xml:"quotaOrderNumber,attr"`
	ReductionIndicator               *int                             `xml:"reductionIndicator,attr"`
	RegulationID                     string                           `xml:"regulationId,attr"`
	RegulationRoleType               int                              `xml:"regulationRoleType,attr"`
	StoppedFlag                      int                              `xml:"stoppedFlag,attr"`
	MeasureConditions                MeasureConditions                `xml:"measureCondition"`
	MeasureFootnoteAssociations      MeasureFootnoteAssociations      `xml:"measureFootnoteAssociation"`
	MeasureComponents                MeasureComponents                `xml:"measureComponent"`
	MeasureExcludedGeographicalAreas MeasureExcludedGeographicalAreas `xml:"measureExcludedGeographicalArea"`
	MeasurePartialTemporaryStops     MeasurePartialTemporaryStops     `xml:"measurePartialTemporaryStop"`
}

func (measures Measures) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO measure (
		sid, sid_additional_code, sid_export_refund_nomenclature, sid_geographical_area,
		sid_goods_nomenclature, additional_code_id, additional_code_type, change_type,
		date_end, date_start, expression, geographical_area_id, goods_nomenclature_code,
		justification_regulation_id, justification_regulation_role_type, measure_type,
		national, quota_order_number, reduction_indicator, regulation_id, regulation_role_type, stopped_flag
	)
	VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22
	)
	ON CONFLICT (sid) DO UPDATE 
	SET sid_additional_code = EXCLUDED.sid_additional_code,
		sid_export_refund_nomenclature = EXCLUDED.sid_export_refund_nomenclature,
		sid_geographical_area = EXCLUDED.sid_geographical_area,
		sid_goods_nomenclature = EXCLUDED.sid_goods_nomenclature,
		additional_code_id = EXCLUDED.additional_code_id,
		additional_code_type = EXCLUDED.additional_code_type,
		change_type = EXCLUDED.change_type,
		date_end = EXCLUDED.date_end,
		date_start = EXCLUDED.date_start,
		expression = EXCLUDED.expression,
		geographical_area_id = EXCLUDED.geographical_area_id,
		goods_nomenclature_code = EXCLUDED.goods_nomenclature_code,
		justification_regulation_id = EXCLUDED.justification_regulation_id,
		justification_regulation_role_type = EXCLUDED.justification_regulation_role_type,
		measure_type = EXCLUDED.measure_type,
		national = EXCLUDED.national,
		quota_order_number = EXCLUDED.quota_order_number,
		reduction_indicator = EXCLUDED.reduction_indicator,
		regulation_id = EXCLUDED.regulation_id,
		regulation_role_type = EXCLUDED.regulation_role_type,
		stopped_flag = EXCLUDED.stopped_flag;
	`

	deleteQuery := `
	DELETE FROM measure
	WHERE sid = $1;
	`

	batch := &pgx.Batch{}

	for i, measure := range measures {
		switch measure.ChangeType {
		case "U": // Insert or update
			batch.Queue(insertQuery, measure.SID, measure.SIDAdditionalCode, measure.SIDExportRefundNomenclature, measure.SIDGeographicalArea, measure.SIDGoodsNomenclature, measure.AdditionalCodeID, measure.AdditionalCodeType, measure.ChangeType, measure.DateEnd, measure.DateStart, measure.Expression, measure.GeographicalAreaID, fmt.Sprint(measure.GoodsNomenclatureCode), measure.JustificationRegulationID, measure.JustificationRegulationRoleType, measure.MeasureType, measure.National, measure.QuotaOrderNumber, measure.ReductionIndicator, measure.RegulationID, measure.RegulationRoleType, measure.StoppedFlag)

			// Queue child elements (already implemented in previous methods)
			if len(measure.MeasureConditions) > 0 {
				if err := measure.MeasureConditions.QueueBatch(ctx, batch, measure.SID); err != nil {
					return fmt.Errorf("failed to queue measure conditions for SID %d: %w", measure.SID, err)
				}
			}

			if len(measure.MeasureFootnoteAssociations) > 0 {
				if err := measure.MeasureFootnoteAssociations.QueueBatch(ctx, batch, measure.SID); err != nil {
					return fmt.Errorf("failed to queue measure footnotes for SID %d: %w", measure.SID, err)
				}
			}

			if len(measure.MeasureComponents) > 0 {
				if err := measure.MeasureComponents.QueueBatch(ctx, batch, measure.SID); err != nil {
					return fmt.Errorf("failed to queue measure components for SID %d: %w", measure.SID, err)
				}
			}

			if len(measure.MeasureExcludedGeographicalAreas) > 0 {
				if err := measure.MeasureExcludedGeographicalAreas.QueueBatch(ctx, batch, measure.SID); err != nil {
					return fmt.Errorf("failed to queue excluded geographical areas for SID %d: %w", measure.SID, err)
				}
			}

			if len(measure.MeasurePartialTemporaryStops) > 0 {
				if err := measure.MeasurePartialTemporaryStops.QueueBatch(ctx, batch, measure.SID); err != nil {
					return fmt.Errorf("failed to queue partial temporary stops for SID %d: %w", measure.SID, err)
				}
			}

		case "D": // Delete
			batch.Queue(deleteQuery, measure.SID)

		default:
			return fmt.Errorf("unknown ChangeType: %s for SID: %d", measure.ChangeType, measure.SID)
		}

		if (i+1)%batchSize == 0 || i == len(measures)-1 {
			if err := conn.SendBatch(ctx, batch).Close(); err != nil {
				return fmt.Errorf("failed to execute batch: %w", err)
			}
			batch = &pgx.Batch{}
		}
	}

	return nil
}

type MeasureConditions []MeasureCondition

type MeasureCondition struct {
	ParentSID                    int                        // added ParentSID
	SID                          int                        `xml:"SID,attr"`
	ActionCode                   int                        `xml:"actionCode,attr"`
	CertificateCode              *string                    `xml:"certificateCode,attr"`
	CertificateType              *string                    `xml:"certificateType,attr"`
	ConditionCodeID              string                     `xml:"conditionCodeId,attr"`
	DutyAmount                   *float64                   `xml:"dutyAmount,attr"`
	Expression                   *string                    `xml:"expression,attr"`
	MeasurementUnitCode          *string                    `xml:"measurementUnitCode,attr"`
	MeasurementUnitQualifierCode *string                    `xml:"measurementUnitQualifierCode,attr"`
	MonetaryUnitCode             *string                    `xml:"monetaryUnitCode,attr"`
	National                     int                        `xml:"national,attr"`
	SequenceNumber               int                        `xml:"sequenceNumber,attr"`
	MeasureConditionComponent    MeasureConditionComponents `xml:"measureConditionComponent"`
}

func (conditions MeasureConditions) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO measure_condition (
		sid, parent_sid, condition_code_id, sequence_number, action_code, certificate_code,
		certificate_type, duty_amount, expression, measurement_unit_code, measurement_unit_qualifier_code, monetary_unit_code, national
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	ON CONFLICT (sid) DO UPDATE 
	SET condition_code_id = EXCLUDED.condition_code_id,
		sequence_number = EXCLUDED.sequence_number,
		action_code = EXCLUDED.action_code,
		certificate_code = EXCLUDED.certificate_code,
		certificate_type = EXCLUDED.certificate_type,
		duty_amount = EXCLUDED.duty_amount,
		expression = EXCLUDED.expression,
		measurement_unit_code = EXCLUDED.measurement_unit_code,
		measurement_unit_qualifier_code = EXCLUDED.measurement_unit_qualifier_code,
		monetary_unit_code = EXCLUDED.monetary_unit_code,
		national = EXCLUDED.national;
	`

	for _, condition := range conditions {
		condition.ParentSID = parentSID
		batch.Queue(insertQuery, condition.SID, condition.ParentSID, condition.ConditionCodeID, condition.SequenceNumber, condition.ActionCode, condition.CertificateCode, condition.CertificateType, condition.DutyAmount, condition.Expression, condition.MeasurementUnitCode, condition.MeasurementUnitQualifierCode, condition.MonetaryUnitCode, condition.National)

		// Queue child components
		if len(condition.MeasureConditionComponent) > 0 {
			if err := condition.MeasureConditionComponent.QueueBatch(ctx, batch, condition.SID); err != nil {
				return fmt.Errorf("failed to queue condition components for Condition SID %d: %w", condition.SID, err)
			}
		}
	}
	return nil
}

type MeasureConditionComponents []MeasureConditionComponent

type MeasureConditionComponent struct {
	ParentSID                    int      // added ParentSID
	DutyAmount                   *float64 `xml:"dutyAmount,attr"`
	DutyExpressionID             int      `xml:"dutyExpressionId,attr"`
	MeasurementUnitCode          *string  `xml:"measurementUnitCode,attr"`
	MeasurementUnitQualifierCode *string  `xml:"measurementUnitQualifierCode,attr"`
	MonetaryUnitCode             *string  `xml:"monetaryUnitCode,attr"`
	National                     int      `xml:"national,attr"`
}

func (components MeasureConditionComponents) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO measure_condition_component (
		parent_sid, duty_amount, duty_expression_id, measurement_unit_code, measurement_unit_qualifier_code, monetary_unit_code, national
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (parent_sid, duty_expression_id) DO UPDATE 
	SET duty_amount = EXCLUDED.duty_amount,
		measurement_unit_code = EXCLUDED.measurement_unit_code,
		measurement_unit_qualifier_code = EXCLUDED.measurement_unit_qualifier_code,
		monetary_unit_code = EXCLUDED.monetary_unit_code,
		national = EXCLUDED.national;
	`

	for _, component := range components {
		component.ParentSID = parentSID

		// Handle potential nil values for duty_amount
		dutyAmount := float64(0.0)
		if component.DutyAmount != nil {
			dutyAmount = *component.DutyAmount
		}

		batch.Queue(insertQuery, component.ParentSID, dutyAmount, component.DutyExpressionID, component.MeasurementUnitCode, component.MeasurementUnitQualifierCode, component.MonetaryUnitCode, component.National)
	}
	return nil
}

type MeasureFootnoteAssociations []MeasureFootnoteAssociation

type MeasureFootnoteAssociation struct {
	ParentSID    int    // added ParentSID
	FootnoteID   string `xml:"footnoteId,attr"`
	FootnoteType string `xml:"footnoteType,attr"`
	National     int    `xml:"national,attr"`
}

func (associations MeasureFootnoteAssociations) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO measure_footnote_association (parent_sid, footnote_id, footnote_type, national)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (parent_sid, footnote_id, footnote_type) DO UPDATE 
	SET national = EXCLUDED.national;
	`

	for _, assoc := range associations {
		assoc.ParentSID = parentSID
		batch.Queue(insertQuery, assoc.ParentSID, assoc.FootnoteID, assoc.FootnoteType, assoc.National)
	}
	return nil
}

type MeasureComponents []MeasureComponent

type MeasureComponent struct {
	ParentSID                    int      // added ParentSID
	DutyAmount                   *float64 `xml:"dutyAmount,attr"`
	DutyExpressionID             int      `xml:"dutyExpressionId,attr"`
	MeasurementUnitCode          *string  `xml:"measurementUnitCode,attr"`
	MeasurementUnitQualifierCode *string  `xml:"measurementUnitQualifierCode,attr"`
	MonetaryUnitCode             *string  `xml:"monetaryUnitCode,attr"`
	National                     int      `xml:"national,attr"`
}

func (components MeasureComponents) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO measure_component (parent_sid, duty_amount, duty_expression_id, measurement_unit_code, measurement_unit_qualifier_code, monetary_unit_code, national)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (parent_sid, duty_expression_id) DO UPDATE 
	SET duty_amount = EXCLUDED.duty_amount,
		measurement_unit_code = EXCLUDED.measurement_unit_code,
		measurement_unit_qualifier_code = EXCLUDED.measurement_unit_qualifier_code,
		monetary_unit_code = EXCLUDED.monetary_unit_code,
		national = EXCLUDED.national;
	`

	for _, comp := range components {
		comp.ParentSID = parentSID
		batch.Queue(insertQuery, comp.ParentSID, comp.DutyAmount, comp.DutyExpressionID, comp.MeasurementUnitCode, comp.MeasurementUnitQualifierCode, comp.MonetaryUnitCode, comp.National)
	}
	return nil
}

type MeasureExcludedGeographicalAreas []MeasureExcludedGeographicalArea

type MeasureExcludedGeographicalArea struct {
	ParentSID           int    // added ParentSID
	GeographicalAreaID  string `xml:"geographicalAreaId,attr"`
	National            int    `xml:"national,attr"`
	SIDGeographicalArea int    `xml:"SIDGeographicalArea,attr"`
}

func (areas MeasureExcludedGeographicalAreas) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO measure_excluded_geographical_area (parent_sid, geographical_area_id, sid_geographical_area, national)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (parent_sid, geographical_area_id, sid_geographical_area) DO UPDATE 
	SET national = EXCLUDED.national;
	`

	for _, area := range areas {
		area.ParentSID = parentSID
		batch.Queue(insertQuery, area.ParentSID, area.GeographicalAreaID, area.SIDGeographicalArea, area.National)
	}
	return nil
}

type MeasurePartialTemporaryStops []MeasurePartialTemporaryStop

type MeasurePartialTemporaryStop struct {
	ParentSID          int    // added ParentSID
	RegulationID       string `xml:"regulationId,attr"`
	RegulationRoleType int    `xml:"regulationRoleType,attr"`
	National           int    `xml:"national,attr"`
}

func (stops MeasurePartialTemporaryStops) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO measure_partial_temporary_stop (parent_sid, regulation_id, regulation_role_type, national)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (parent_sid) DO UPDATE 
	SET regulation_id = EXCLUDED.regulation_id,
		regulation_role_type = EXCLUDED.regulation_role_type,
		national = EXCLUDED.national;
	`

	for _, stop := range stops {
		stop.ParentSID = parentSID
		batch.Queue(insertQuery, stop.ParentSID, stop.RegulationID, stop.RegulationRoleType, stop.National)
	}
	return nil
}
