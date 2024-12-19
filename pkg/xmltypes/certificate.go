package xmltypes

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Certificates []Certificate

type Certificate struct {
	CertificateCode               string                        `xml:"certificateCode,attr"`
	CertificateType               string                        `xml:"certificateType,attr"`
	ChangeType                    string                        `xml:"changeType,attr"`
	DateEnd                       FileDistTime                  `xml:"dateEnd,attr"`
	DateStart                     FileDistTime                  `xml:"dateStart,attr"`
	National                      int                           `xml:"national,attr"`
	CertificateDescriptionPeriods CertificateDescriptionPeriods `xml:"certificateDescriptionPeriod"`
}

func (certificates Certificates) BatchInsert(ctx context.Context, conn *pgx.Conn, batchSize int) error {
	insertQuery := `
	INSERT INTO certificate (certificate_code, certificate_type, change_type, date_start, date_end, national)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (certificate_code, certificate_type) DO UPDATE 
	SET change_type = EXCLUDED.change_type,
		date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		national = EXCLUDED.national
	`

	deleteQuery := `
	DELETE FROM certificate
	WHERE certificate_code = $1 AND certificate_type = $2;
	`

	batch := &pgx.Batch{}

	for i, cert := range certificates {
		switch cert.ChangeType {
		case "U": // Insert or update
			batch.Queue(insertQuery, cert.CertificateCode, cert.CertificateType, cert.ChangeType, cert.DateStart, cert.DateEnd, cert.National)

			// Queue child description periods
			if len(cert.CertificateDescriptionPeriods) > 0 {
				if err := cert.CertificateDescriptionPeriods.QueueBatch(ctx, batch, cert.CertificateCode, cert.CertificateType); err != nil {
					return fmt.Errorf("failed to queue description periods for certificate %s-%s: %w", cert.CertificateCode, cert.CertificateType, err)
				}
			}

		case "D": // Delete
			batch.Queue(deleteQuery, cert.CertificateCode, cert.CertificateType)

		default:
			return fmt.Errorf("unknown ChangeType: %s for CertificateCode: %s", cert.ChangeType, cert.CertificateCode)
		}

		if (i+1)%batchSize == 0 || i == len(certificates)-1 {
			if err := conn.SendBatch(ctx, batch).Close(); err != nil {
				return fmt.Errorf("failed to execute batch: %w", err)
			}
			batch = &pgx.Batch{}
		}
	}

	return nil
}

type CertificateDescriptionPeriods []CertificateDescriptionPeriod

type CertificateDescriptionPeriod struct {
	DateEnd                 FileDistTime            `xml:"dateEnd,attr"`
	DateStart               FileDistTime            `xml:"dateStart,attr"`
	National                int                     `xml:"national,attr"`
	SID                     int                     `xml:"SID,attr"`
	CertificateDescriptions CertificateDescriptions `xml:"certificateDescription"`
}

func (periods CertificateDescriptionPeriods) QueueBatch(ctx context.Context, batch *pgx.Batch, parentCode, parentType string) error {
	insertQuery := `
	INSERT INTO certificate_description_period (sid, parent_certificate_code, parent_certificate_type, date_start, date_end, national)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (sid) DO UPDATE 
	SET date_start = EXCLUDED.date_start,
		date_end = EXCLUDED.date_end,
		national = EXCLUDED.national;
	`

	for _, period := range periods {
		batch.Queue(insertQuery, period.SID, parentCode, parentType, period.DateStart, period.DateEnd, period.National)

		if len(period.CertificateDescriptions) > 0 {
			if err := period.CertificateDescriptions.QueueBatch(ctx, batch, period.SID); err != nil {
				return fmt.Errorf("failed to queue descriptions for SID %d: %w", period.SID, err)
			}
		}
	}
	return nil
}

type CertificateDescriptions []CertificateDescription

type CertificateDescription struct {
	Description string `xml:"description,attr"`
	LanguageID  string `xml:"languageId,attr"`
	National    int    `xml:"national,attr"`
}

func (descriptions CertificateDescriptions) QueueBatch(ctx context.Context, batch *pgx.Batch, parentSID int) error {
	insertQuery := `
	INSERT INTO certificate_description (parent_sid, description, language_id, national)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (parent_sid, language_id) DO UPDATE 
	SET description = EXCLUDED.description,
		national = EXCLUDED.national;
	`

	for _, desc := range descriptions {
		batch.Queue(insertQuery, parentSID, desc.Description, desc.LanguageID, desc.National)
	}
	return nil
}
