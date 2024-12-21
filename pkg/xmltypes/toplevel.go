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
	Footnotes                    Footnotes                    `xml:"footnote"`
	GeographicalAreas            GeographicalAreas            `xml:"geographicalArea"`
	GoodsNomenclatureGroups      GoodsNomenclatureGroups      `xml:"goodsNomenclatureGroup"`
	GoodsNomenclatures           GoodsNomenclatures           `xml:"goodsNomenclature"`
	// Record                          []Record                        `xml:"record"`
	LookupTables              LookupTables              `xml:"lookupTable"`
	MeasureActions            MeasureActions            `xml:"measureAction"`
	MeasureConditionCodes     MeasureConditionCodes     `xml:"measureConditionCode"`
	MeasureTypes              MeasureTypes              `xml:"measureType"`
	Measures                  Measures                  `xml:"measure"`
	MeasurementUnitQualifiers MeasurementUnitQualifiers `xml:"measurementUnitQualifier"`
	MeasurementUnits          MeasurementUnits          `xml:"measurementUnit"`
	Measurements              Measurements              `xml:"measurement"`
	// MeursingAdditionalCode          []MeursingAdditionalCode        `xml:"meursingAdditionalCode"`
	// MeursingHeading                 []MeursingHeading               `xml:"meursingHeading"`
	// MeursingSubheading              []MeursingSubheading            `xml:"meursingSubheading"`
	// MeursingTablePlan               []MeursingTablePlan             `xml:"meursingTablePlan"`
	MonetaryExchangePeriods MonetaryExchangePeriods `xml:"monetaryExchangePeriod"`
	// PreferenceCode                  []PreferenceCode                `xml:"preferenceCode"`
	// QuotaBalanceEvent               []QuotaBalanceEvent             `xml:"quotaBalanceEvent"`
	// QuotaDefinition                 []QuotaDefinition               `xml:"quotaDefinition"`
	// QuotaUnblockingEvent            []QuotaUnblockingEvent          `xml:"quotaUnblockingEvent"`
	// QuotaCriticalEvent              []QuotaCriticalEvent            `xml:"quotaCriticalEvent"`
	// QuotaExhaustionEvent            []QuotaExhaustionEvent          `xml:"quotaExhaustionEvent"`
	// QuotaReopeningEvent             []QuotaReopeningEvent           `xml:"quotaReopeningEvent"`
	// QuotaUnsuspensionEvent          []QuotaUnsuspensionEvent        `xml:"quotaUnsuspensionEvent"`
	// QuotaOrderNumber                []QuotaOrderNumber              `xml:"quotaOrderNumber"`
	BaseRegulation              BaseRegulations              `xml:"baseRegulation"`
	ModificationRegulation      ModificationRegulations      `xml:"modificationRegulation"`
	FullTemporaryStopRegulation FullTemporaryStopRegulations `xml:"fullTemporaryStopRegulation"`
	// TaxCode                         []TaxCode                       `xml:"taxCode"`
	UnquotedMonetaryExchangePeriods UnquotedMonetaryExchangePeriods `xml:"unquotedMonetaryExchangePeriod"`
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

type QuotaDefinition struct {
	ChangeType                   string                  `xml:"changeType,attr"`
	DateEnd                      FileDistTime            `xml:"dateEnd,attr"`
	DateStart                    FileDistTime            `xml:"dateStart,attr"`
	Description                  *string                 `xml:"description,attr"`
	InitialVolume                float64                 `xml:"initialVolume,attr"`
	MeasurementUnitCode          *string                 `xml:"measurementUnitCode,attr"`
	MeasurementUnitQualifierCode *string                 `xml:"measurementUnitQualifierCode,attr"`
	MonetaryUnitCode             *string                 `xml:"monetaryUnitCode,attr"`
	National                     int                     `xml:"national,attr"`
	QuotaCriticalStateCode       string                  `xml:"quotaCriticalStateCode,attr"`
	QuotaCriticalThreshold       int                     `xml:"quotaCriticalThreshold,attr"`
	QuotaMaximumPrecision        int                     `xml:"quotaMaximumPrecision,attr"`
	QuotaOrderNumber             int                     `xml:"quotaOrderNumber,attr"`
	SID                          int                     `xml:"SID,attr"`
	SIDQuotaOrderNumber          int                     `xml:"SIDQuotaOrderNumber,attr"`
	Volume                       float64                 `xml:"volume,attr"`
	QuotaBlockingPeriod          []QuotaBlockingPeriod   `xml:"quotaBlockingPeriod"`
	QuotaAssociation             []QuotaAssociation      `xml:"quotaAssociation"`
	QuotaSuspensionPeriod        []QuotaSuspensionPeriod `xml:"quotaSuspensionPeriod"`
}

type QuotaBlockingPeriod struct {
	BlockingPeriodType int          `xml:"blockingPeriodType,attr"`
	DateEnd            FileDistTime `xml:"dateEnd,attr"`
	DateStart          FileDistTime `xml:"dateStart,attr"`
	Description        *string      `xml:"description,attr"`
	National           int          `xml:"national,attr"`
	SID                int          `xml:"SID,attr"`
}

type QuotaBalanceEvent struct {
	ChangeType             string            `xml:"changeType,attr"`
	EndOccurrenceTimestamp FileDistTimeStamp `xml:"endOccurrenceTimestamp,attr"`
	ImportedAmount         float64           `xml:"importedAmount,attr"`
	LastImportDate         FileDistTime      `xml:"lastImportDate,attr"`
	National               int               `xml:"national,attr"`
	NewBalance             float64           `xml:"newBalance,attr"`
	OccurrenceTimestamp    FileDistTimeStamp `xml:"occurrenceTimestamp,attr"`
	OldBalance             float64           `xml:"oldBalance,attr"`
	SIDQuotaDefinition     int               `xml:"SIDQuotaDefinition,attr"`
}

type QuotaCriticalEvent struct {
	ChangeType             string            `xml:"changeType,attr"`
	CriticalDate           FileDistTime      `xml:"criticalDate,attr"`
	EndOccurrenceTimestamp FileDistTimeStamp `xml:"endOccurrenceTimestamp,attr"`
	National               int               `xml:"national,attr"`
	OccurrenceTimestamp    FileDistTimeStamp `xml:"occurrenceTimestamp,attr"`
	QuotaCriticalStateCode string            `xml:"quotaCriticalStateCode,attr"`
	SIDQuotaDefinition     int               `xml:"SIDQuotaDefinition,attr"`
}

type QuotaExhaustionEvent struct {
	ChangeType             string            `xml:"changeType,attr"`
	EndOccurrenceTimestamp FileDistTimeStamp `xml:"endOccurrenceTimestamp,attr"`
	ExhaustionDate         FileDistTime      `xml:"exhaustionDate,attr"`
	National               int               `xml:"national,attr"`
	OccurrenceTimestamp    FileDistTimeStamp `xml:"occurrenceTimestamp,attr"`
	SIDQuotaDefinition     int               `xml:"SIDQuotaDefinition,attr"`
}

type QuotaReopeningEvent struct {
	ChangeType             string            `xml:"changeType,attr"`
	EndOccurrenceTimestamp FileDistTimeStamp `xml:"endOccurrenceTimestamp,attr"`
	National               int               `xml:"national,attr"`
	OccurrenceTimestamp    FileDistTimeStamp `xml:"occurrenceTimestamp,attr"`
	ReopeningDate          FileDistTime      `xml:"reopeningDate,attr"`
	SIDQuotaDefinition     int               `xml:"SIDQuotaDefinition,attr"`
}

type QuotaOrderNumber struct {
	ChangeType             string                   `xml:"changeType,attr"`
	DateEnd                FileDistTime             `xml:"dateEnd,attr"`
	DateStart              FileDistTime             `xml:"dateStart,attr"`
	National               int                      `xml:"national,attr"`
	QuotaOrderNumber       int                      `xml:"quotaOrderNumber,attr"`
	SID                    int                      `xml:"SID,attr"`
	QuotaOrderNumberOrigin []QuotaOrderNumberOrigin `xml:"quotaOrderNumberOrigin"`
}

type QuotaOrderNumberOrigin struct {
	DateEnd                         FileDistTime                      `xml:"dateEnd,attr"`
	DateStart                       FileDistTime                      `xml:"dateStart,attr"`
	GeographicalAreaID              string                            `xml:"geographicalAreaId,attr"`
	National                        int                               `xml:"national,attr"`
	SID                             int                               `xml:"SID,attr"`
	SIDGeographicalArea             int                               `xml:"SIDGeographicalArea,attr"`
	QuotaOrderNumberOriginExclusion []QuotaOrderNumberOriginExclusion `xml:"quotaOrderNumberOriginExclusion"`
}

type QuotaUnblockingEvent struct {
	ChangeType             string            `xml:"changeType,attr"`
	EndOccurrenceTimestamp FileDistTimeStamp `xml:"endOccurrenceTimestamp,attr"`
	National               int               `xml:"national,attr"`
	OccurrenceTimestamp    FileDistTimeStamp `xml:"occurrenceTimestamp,attr"`
	SIDQuotaDefinition     int               `xml:"SIDQuotaDefinition,attr"`
	UnblockingDate         FileDistTime      `xml:"unblockingDate,attr"`
}

type QuotaAssociation struct {
	Coefficient  float64 `xml:"coefficient,attr"`
	National     int     `xml:"national,attr"`
	RelationType string  `xml:"relationType,attr"`
	SIDSubQuota  int     `xml:"SIDSubQuota,attr"`
}

type MeursingAdditionalCode struct {
	AdditionalCodeID           int                          `xml:"additionalCodeId,attr"`
	ChangeType                 string                       `xml:"changeType,attr"`
	DateStart                  FileDistTime                 `xml:"dateStart,attr"`
	National                   int                          `xml:"national,attr"`
	SID                        int                          `xml:"SID,attr"`
	MeursingTableCellComponent []MeursingTableCellComponent `xml:"meursingTableCellComponent"`
}

type MeursingTableCellComponent struct {
	DateStart                FileDistTime `xml:"dateStart,attr"`
	HeadingNumber            int          `xml:"headingNumber,attr"`
	MeursingTablePlanID      int          `xml:"meursingTablePlanId,attr"`
	National                 int          `xml:"national,attr"`
	RowColumnCode            int          `xml:"rowColumnCode,attr"`
	SubheadingSequenceNumber int          `xml:"subheadingSequenceNumber,attr"`
}

type MeursingHeading struct {
	ChangeType                         string                              `xml:"changeType,attr"`
	DateStart                          FileDistTime                        `xml:"dateStart,attr"`
	HeadingNumber                      int                                 `xml:"headingNumber,attr"`
	MeursingTablePlanID                int                                 `xml:"meursingTablePlanId,attr"`
	National                           int                                 `xml:"national,attr"`
	RowColumnCode                      int                                 `xml:"rowColumnCode,attr"`
	MeursingHeadingFootnoteAssociation *MeursingHeadingFootnoteAssociation `xml:"meursingHeadingFootnoteAssociation"`
	MeursingHeadingText                []MeursingHeadingText               `xml:"meursingHeadingText"`
}

type MeursingHeadingFootnoteAssociation struct {
	DateStart    FileDistTime `xml:"dateStart,attr"`
	FootnoteID   int          `xml:"footnoteId,attr"`
	FootnoteType string       `xml:"footnoteType,attr"`
	National     int          `xml:"national,attr"`
}

type MeursingHeadingText struct {
	Description *string `xml:"description,attr"`
	LanguageID  string  `xml:"languageId,attr"`
	National    int     `xml:"national,attr"`
}

type MeursingSubheading struct {
	ChangeType               string       `xml:"changeType,attr"`
	DateStart                FileDistTime `xml:"dateStart,attr"`
	Description              string       `xml:"description,attr"`
	HeadingNumber            int          `xml:"headingNumber,attr"`
	MeursingTablePlanID      int          `xml:"meursingTablePlanId,attr"`
	National                 int          `xml:"national,attr"`
	RowColumnCode            int          `xml:"rowColumnCode,attr"`
	SubheadingSequenceNumber int          `xml:"subheadingSequenceNumber,attr"`
}

type MeursingTablePlan struct {
	ChangeType          string       `xml:"changeType,attr"`
	DateStart           FileDistTime `xml:"dateStart,attr"`
	MeursingTablePlanID int          `xml:"meursingTablePlanId,attr"`
	National            int          `xml:"national,attr"`
}

type PreferenceCode struct {
	ChangeType                string                      `xml:"changeType,attr"`
	DateStart                 FileDistTime                `xml:"dateStart,attr"`
	PrefCode                  int                         `xml:"prefCode,attr"`
	PreferenceCodeDescription []PreferenceCodeDescription `xml:"preferenceCodeDescription"`
}

type PreferenceCodeDescription struct {
	Description string `xml:"description,attr"`
	LanguageID  string `xml:"languageId,attr"`
}

type QuotaSuspensionPeriod struct {
	DateEnd     FileDistTime `xml:"dateEnd,attr"`
	DateStart   FileDistTime `xml:"dateStart,attr"`
	Description *string      `xml:"description,attr"`
	National    int          `xml:"national,attr"`
	SID         int          `xml:"SID,attr"`
}

type QuotaUnsuspensionEvent struct {
	ChangeType             string            `xml:"changeType,attr"`
	EndOccurrenceTimestamp FileDistTimeStamp `xml:"endOccurrenceTimestamp,attr"`
	National               int               `xml:"national,attr"`
	OccurrenceTimestamp    FileDistTimeStamp `xml:"occurrenceTimestamp,attr"`
	SIDQuotaDefinition     int               `xml:"SIDQuotaDefinition,attr"`
	UnsuspensionDate       FileDistTime      `xml:"unsuspensionDate,attr"`
}

type QuotaOrderNumberOriginExclusion struct {
	GeographicalAreaID  string `xml:"geographicalAreaId,attr"`
	National            int    `xml:"national,attr"`
	SIDGeographicalArea int    `xml:"SIDGeographicalArea,attr"`
}

type TaxCode struct {
	ChangeType         string               `xml:"changeType,attr"`
	DateStart          FileDistTime         `xml:"dateStart,attr"`
	National           int                  `xml:"national,attr"`
	SID                int                  `xml:"SID,attr"`
	TaxCode            string               `xml:"taxCode,attr"`
	TaxCodeDescription []TaxCodeDescription `xml:"taxCodeDescription"`
}

type TaxCodeDescription struct {
	Description string `xml:"description,attr"`
	LanguageID  string `xml:"languageId,attr"`
	SID         int    `xml:"SID,attr"`
}
