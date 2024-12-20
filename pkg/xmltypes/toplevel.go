package xmltypes

import (
	"context"
	"database/sql/driver"
	"encoding/xml"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

type Export struct {
	ID         string     `xml:"id"`
	ExportType string     `xml:"exportType"`
	Parameters Parameters `xml:"parameters"`
	Items      Items      `xml:"items"`
}

type Parameters struct {
	QueryDateStart *string `xml:"queryDateStart"`
}

type FileDistItem interface {
	BatchInsert(context.Context, *pgx.Conn, int) error
}

type Items struct {
	AdditionalCodes              AdditionalCodes              `xml:"additionalCode"`
	Certificates                 Certificates                 `xml:"certificate"`
	CodeTypes                    CodeTypes                    `xml:"codeType"`
	DeclarableGoodsNomenclatures DeclarableGoodsNomenclatures `xml:"declarableGoodsNomenclature"`
	DutyExpressions              DutyExpressions              `xml:"dutyExpression"`
	ExportRefundNomenclatures    ExportRefundNomenclatures    `xml:"exportRefundNomenclature"`
	// Footnotes                       Footnotes                       `xml:"footnote"`
	// GeographicalAreas               GeographicalAreas               `xml:"geographicalArea"`
	// GoodsNomenclatureGroups         GoodsNomenclatureGroups         `xml:"goodsNomenclatureGroup"`
	// GoodsNomenclatures              GoodsNomenclatures              `xml:"goodsNomenclature"`
	// Record                          []Record                        `xml:"record"`
	// LookupTables                    LookupTables                    `xml:"lookupTable"`
	// MeasureActions                  MeasureActions                  `xml:"measureAction"`
	// MeasureConditionCodes           MeasureConditionCodes           `xml:"measureConditionCode"`
	// MeasureTypes                    MeasureTypes                    `xml:"measureType"`
	// Measures                        Measures                        `xml:"measure"`
	// MeasurementUnitQualifiers       MeasurementUnitQualifiers       `xml:"measurementUnitQualifier"`
	// MeasurementUnit                 []MeasurementUnit               `xml:"measurementUnit"`
	// Measurements                    Measurements                    `xml:"measurement"`
	// MeursingAdditionalCode          []MeursingAdditionalCode        `xml:"meursingAdditionalCode"`
	// MeursingHeading                 []MeursingHeading               `xml:"meursingHeading"`
	// MeursingSubheading              []MeursingSubheading            `xml:"meursingSubheading"`
	// MeursingTablePlan               []MeursingTablePlan             `xml:"meursingTablePlan"`
	// MonetaryExchangePeriods         MonetaryExchangePeriods         `xml:"monetaryExchangePeriod"`
	// PreferenceCode                  []PreferenceCode                `xml:"preferenceCode"`
	// QuotaBalanceEvent               []QuotaBalanceEvent             `xml:"quotaBalanceEvent"`
	// QuotaDefinition                 []QuotaDefinition               `xml:"quotaDefinition"`
	// QuotaUnblockingEvent            []QuotaUnblockingEvent          `xml:"quotaUnblockingEvent"`
	// QuotaCriticalEvent              []QuotaCriticalEvent            `xml:"quotaCriticalEvent"`
	// QuotaExhaustionEvent            []QuotaExhaustionEvent          `xml:"quotaExhaustionEvent"`
	// QuotaReopeningEvent             []QuotaReopeningEvent           `xml:"quotaReopeningEvent"`
	// QuotaUnsuspensionEvent          []QuotaUnsuspensionEvent        `xml:"quotaUnsuspensionEvent"`
	// QuotaOrderNumber                []QuotaOrderNumber              `xml:"quotaOrderNumber"`
	// BaseRegulation                  BaseRegulations                 `xml:"baseRegulation"`
	// ModificationRegulation          ModificationRegulations         `xml:"modificationRegulation"`
	// FullTemporaryStopRegulation     FullTemporaryStopRegulations    `xml:"fullTemporaryStopRegulation"`
	// TaxCode                         []TaxCode                       `xml:"taxCode"`
	// UnquotedMonetaryExchangePeriods UnquotedMonetaryExchangePeriods `xml:"unquotedMonetaryExchangePeriod"`
}

// FileDistTime is a custom type that wraps time.Time for XML and database use.
type FileDistTime struct {
	time.Time // Embed time.Time directly
}

// UnmarshalXMLAttr allows parsing the time from XML attributes.
func (ft *FileDistTime) UnmarshalXMLAttr(attr xml.Attr) error {
	if attr.Value == "" {
		*ft = FileDistTime{Time: time.Time{}} // Zero time
		return nil
	}
	parsedTime, err := time.Parse("2006-01-02", attr.Value)
	if err != nil {
		return err
	}
	ft.Time = parsedTime
	return nil
}

// Value implements the driver.Valuer interface for database compatibility.
func (ft FileDistTime) Value() (driver.Value, error) {
	if ft.Time.IsZero() {
		return nil, nil // Return NULL for zero time
	}
	return ft.Time, nil
}

// Scan implements the sql.Scanner interface to retrieve time.Time from the database.
func (ft *FileDistTime) Scan(value interface{}) error {
	if value == nil {
		ft.Time = time.Time{}
		return nil
	}
	if t, ok := value.(time.Time); ok {
		ft.Time = t
		return nil
	}
	return nil
}

// String for debugging/logging purposes.
func (ft FileDistTime) String() string {
	if ft.Time.IsZero() {
		return "nil"
	}
	return ft.Time.Format(time.RFC3339)
}

// FileDistTimeStamp is a string type for handling XML timestamp strings.
// The XML format prevents us from using the time.Time type directly.
type FileDistTimeStamp string

// Parse converts the timestamp string to time.Time in Sweden's timezone.
func (ct FileDistTimeStamp) Value() *time.Time {
	if ct == "" {
		return nil
	}

	const layout = "2006-01-02T15:04:05" // Tulltaxan XML timestamp format
	location, err := time.LoadLocation("Europe/Stockholm")
	if err != nil {
		log.Fatalf("Failed to load location: %v", err)
	}

	t, err := time.Parse(layout, string(ct))
	if err != nil {
		log.Fatalf("Failed to parse time: %v", err)
	}

	t = t.In(location)
	return &t
}
