package filedist

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"database/sql"
	_ "embed"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func StartDbMaintenanceScheduler() {
	pubKey, err := downloadPublicKey(`https://distr.tullverket.se/tulltaxan/Tulltaxan_Fildistribution.asc`)
	if err != nil {
		slog.Error("downloadPublicKey", "error", err)
	}

	// Perform initial database maintenance immediately
	err = performDbMaintenance(pubKey)
	if err != nil {
		slog.Error("Error during initial database maintenance", "error", err)
	}

	// Start daily database maintenance loop
	go func() {
		for {
			// Calculate the duration until the next 11:30 PM
			now := time.Now()
			nextRun := time.Date(now.Year(), now.Month(), now.Day(), 23, 30, 0, 0, now.Location())

			slog.Info("db maintainer waiting on next run:", "next run", nextRun.String())
			// If the next run time is in the past (i.e., it's already after 11:30 PM today), schedule it for tomorrow
			if nextRun.Before(now) {
				nextRun = nextRun.Add(24 * time.Hour)
			}

			timeUntilNextRun := time.Until(nextRun)

			// Sleep until the next run time
			time.Sleep(timeUntilNextRun)

			// Perform scheduled database maintenance at 11:30 PM
			err := performDbMaintenance(pubKey)
			if err != nil {
				slog.Error("Error during scheduled database maintenance", "error", err)
			}
		}
	}()
}

// downloadPublicKey fetches the public key from the given URL.
func downloadPublicKey(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download public key: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read public key response: %w", err)
	}

	return string(body), nil
}

// performDbMaintenance performs the necessary maintenance tasks on the database.
// It downloads new files from the distribution, processes them into the database, and sets the active codes.
func performDbMaintenance(pubKey string) error {
	slog.Info("Starting database maintenance")

	slog.Info("Downloading new files from distribution")
	// Download new files from distribution
	totFiles, err := downloadAndPrepareNewFiles("https://distr.tullverket.se/tulltaxan/xml/tot/", pubKey)
	if err != nil {
		return fmt.Errorf("error fetching tot files: %w", err)
	}

	difFiles, err := downloadAndPrepareNewFiles("https://distr.tullverket.se/tulltaxan/xml/dif/", pubKey)
	if err != nil {
		return fmt.Errorf("error fetching dif files: %w", err)
	}

	slog.Info("Files downloaded successfully", "totFiles", len(totFiles), "difFiles", len(difFiles))

	return nil
}

// downloadAndPrepareNewFiles retrieves and prepares a list of files from a specified distribution URL.
// This function downloads, decrypts, and decompresses files that are available in the distribution.
// If a file already exists in the output directory, it is skipped. It returns a list of the resulting files.
func downloadAndPrepareNewFiles(distUrl, pubKey string) ([]string, error) {
	slog.Debug("Retrieving files", "URL", distUrl)

	response, err := http.Get(distUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to get file list, url: [%v] err: %w", distUrl, err)
	}

	// Parse the HTML content for pgp files
	fileList, err := parseHtmlForPgpAnchors(html.NewTokenizer(response.Body))
	if err != nil {
		return nil, fmt.Errorf("parseHtmlForPgpAnchors: %w", err)
	}

	fileList, err = sortFilesByDate(fileList)
	if err != nil {
		return nil, fmt.Errorf("sortFilesByDate: %w", err)
	}
	fmt.Println(fileList)

	// // Query already inserted filenames
	// insertedFileNames, err := getInsertedFileNames(db)
	// if err != nil {
	// 	return nil, fmt.Errorf("getInsertedFileNames: %w", err)
	// }

	// // filter out inserted files to only download new ones.
	// fileList = filterOutInsertedFiles(fileList, insertedFileNames)

	// Iterate through file list and download all files for the current month
	for _, v := range fileList {
		// Construct the url for the file
		fileUrl, err := url.JoinPath(distUrl, filepath.ToSlash(filepath.Base(v)))
		if err != nil {
			return nil, fmt.Errorf("unable to construct valid URL for tot file dist, url: [%v] err: %w", fileUrl, err)
		}

		// Download and prepare the file to a usable format
		gzReader, err := fetchDecryptedReader(fileUrl, pubKey)
		if err != nil {
			return nil, fmt.Errorf("downloadAndPrepareFile: %w", err)
		}

		// Save the resulting file name in the list
		slog.Info("Downloaded and prepared dist file", "filename", v)

		gzReader.Close()
	}

	return fileList, nil
}

// parseHtmlForPgpAnchors scans HTML content from the provided tokenizer for anchor tags linking to .pgp files.
// It returns a slice of the href values of these links. The function stops parsing when it encounters an end-of-file
// or an error, returning the collected links up to that point along with any error encountered.
func parseHtmlForPgpAnchors(tokenizer *html.Tokenizer) (anchorRefs []string, err error) {
	anchorRefs = make([]string, 0, 100)
	for {
		// iterate through html
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.StartTagToken:
			// Check if it's a anchor tag
			token := tokenizer.Token()
			if token.DataAtom == atom.A {
				for _, attribute := range token.Attr {
					// Ensure attribute is a href to a pgp file. Add it to result if so.
					if attribute.Key == "href" && filepath.Ext(attribute.Val) == ".pgp" {
						anchorRefs = append(anchorRefs, attribute.Val)
					}
				}
			}
		case html.ErrorToken:
			err := tokenizer.Err()
			// EOF = Parsing finished and all refs are added to slice.
			if err == io.EOF {
				return anchorRefs, nil
			} else {
				return anchorRefs, fmt.Errorf("tokenizer error occured; %w", err)
			}
		}
	}
}

// getInsertedFileNames sends a select query to the inserted_files table in DB
// and returns a slice of all filenames it retrieved.
func getInsertedFileNames(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT file_name FROM inserted_files;")
	if err != nil {
		return nil, err
	}

	var alreadyImportedFileNames []string
	for rows.Next() {
		filename := new(sql.NullString)
		rows.Scan(filename)

		if filename.Valid {
			alreadyImportedFileNames = append(alreadyImportedFileNames, filename.String)
		}
	}

	return alreadyImportedFileNames, nil
}

// filterOutInsertedFiles filters out files from the distribution list that already exist locally.
func filterOutInsertedFiles(distributionFiles []string, alreadyImportedFileNames []string) []string {
	alreadyImportedFileMap := make(map[string]bool)

	// Create a map of local files for quick lookup
	for _, file := range alreadyImportedFileNames {
		alreadyImportedFileMap[filepath.Base(file)] = true
	}

	newFiles := []string{}
	for _, file := range distributionFiles {
		// If the file is not in the local map, it is new
		if !alreadyImportedFileMap[strings.TrimSuffix(filepath.Base(file), ".gz.pgp")] {
			newFiles = append(newFiles, file)
		} else {
			slog.Debug("Skipping filedist file, already exists locally", "filename", file)
		}

	}

	return newFiles
}

// sortFilesByDate sorts a list of filenames by their embedded dates in ascending order.
func sortFilesByDate(files []string) ([]string, error) {
	// Create a slice of file-date pairs
	type fileDate struct {
		filename string
		date     time.Time
	}

	fileDates := make([]fileDate, 0, len(files))

	// Extract dates
	for _, file := range files {
		date, err := extractDate(file)
		if err != nil {
			return nil, fmt.Errorf("failed to extract date from %s: %w", file, err)
		}
		fileDates = append(fileDates, fileDate{filename: file, date: date})
	}

	// Sort by date
	sort.Slice(fileDates, func(i, j int) bool {
		return fileDates[i].date.Before(fileDates[j].date)
	})

	// Extract sorted filenames
	sortedFiles := make([]string, len(files))
	for i, fd := range fileDates {
		sortedFiles[i] = fd.filename
	}

	return sortedFiles, nil
}

// extractDate extracts the date (YYMMDD) from the filename and parses it into a time.Time.
func extractDate(filename string) (time.Time, error) {
	// Split the filename on underscores
	parts := strings.Split(filename, "_")
	if len(parts) < 2 {
		return time.Time{}, fmt.Errorf("invalid filename format: %s", filename)
	}

	// The date is the last part before the extension
	lastPart := parts[len(parts)-1]
	dateStr := strings.Split(lastPart, ".")[0]

	// Parse the date in YYMMDD format
	return time.Parse("060102", dateStr)
}

// fetchDecryptedReader downloads, decrypts, and decompresses a PGP-encrypted, GZIP-compressed file from a given URL.
// It returns an io.Reader for the decompressed content.
func fetchDecryptedReader(url, pubKey string) (io.ReadCloser, error) {
	slog.Info("Downloading file", "url", url)

	// Download the file
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("unable to get PGP file: %w", err)
	}
	defer response.Body.Close()

	// Decrypt and decompress
	gzReader, err := decryptAndExtractGzippedFile(pubKey, response.Body)
	if err != nil {
		return nil, fmt.Errorf("decryptAndExtractGzippedFile: %w", err)
	}

	return gzReader, nil
}

// VerifyAndDecryptGzippedPgpFile verifies the PGP signature of a gzipped file,
// then decrypts and decompresses its contents.
func decryptAndExtractGzippedFile(pubKey string, signedFile io.Reader) (io.ReadCloser, error) {
	// Import the public key
	entityList, err := importPublicKey(pubKey)
	if err != nil {
		return nil, fmt.Errorf("importPublicKey: %w", err)
	}

	// Decode the armored signature
	block, err := armor.Decode(signedFile)
	if err != nil {
		return nil, fmt.Errorf("unable to decode signed file: %w", err)
	}

	// Initialize pgp reader
	md, err := openpgp.ReadMessage(block.Body, entityList, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize pgp reader: %w", err)
	}

	// Wrap with gzip reader
	reader, err := gzip.NewReader(md.UnverifiedBody)
	if err != nil {
		return nil, fmt.Errorf("unable to initalize gzip reader: %w", err)
	}

	return reader, nil
}

// Function to import a public key
func importPublicKey(pubKey string) (openpgp.EntityList, error) {
	block, err := armor.Decode(bytes.NewReader([]byte(pubKey)))
	if err != nil {
		return nil, fmt.Errorf("unable to decode public key: %w", err)
	}

	if block.Type != "PGP PUBLIC KEY BLOCK" {
		return nil, fmt.Errorf("block type is not [PGP PUBLIC KEY BLOCK]")
	}

	entityList, err := openpgp.ReadKeyRing(block.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read keyring: %w", err)
	}

	return entityList, nil
}

// cleanDifFile removes all lines containing <record> tags from a dif file.
// It expects line-feeds to be used as row delimiters in the xml string.
func cleanDifFile(f io.Reader) (*bufio.Reader, error) {

	content, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("unable to read filedist file content: %w", err)
	}

	strContentWithoutRecordTags, err := removeRecordTags(string(content))
	if err != nil {
		return nil, fmt.Errorf("unable to remove record tags from string: %w", err)
	}

	return bufio.NewReader(strings.NewReader(strContentWithoutRecordTags)), nil
}

// removeRecordTags removes all lines containing <record> tags from an xml string.
// It expects line-feeds to be used as row delimiters in the xml string.
func removeRecordTags(xmlStr string) (xmlWithoutRecords string, err error) {
	result := strings.Builder{}
	rows := strings.Split(xmlStr, "\n")
	if len(rows) == 0 {
		return "", fmt.Errorf("no rows found in xml string")
	}

	for _, line := range rows {
		if strings.Contains(line, "<record") || strings.Contains(line, "</record>") {
			continue
		}
		result.WriteString(line)
	}

	return result.String(), nil
}
