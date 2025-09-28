package services

import (
	"bytes"
	"encoding/csv"
	"io"
	"regexp"
	"slices"
)

type (
	Parser interface {
		Parse([]byte) ([]byte, error)
	}
	CSVParser struct{}
)

// NewCSVParser returns a new instance of CSVParser.
func NewCSVParser() Parser { return &CSVParser{} }

// Parse reads CSV data, appends a "has_email" column, and returns the modified CSV.
// It marks each row with "true" if any field contains a valid email address.
func (p *CSVParser) Parse(data []byte) ([]byte, error) {

	r := csv.NewReader(bytes.NewReader(data))
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Regular expression to match email addresses
	emailRegexp := regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	isHeader := true
	for {
		record, err := r.Read()
		if err == io.EOF {
			break // End of file reached
		}
		if err != nil {
			return nil, err
		}

		if isHeader {
			// Append "has_email" column to header row
			record = append(record, "has_email")
			isHeader = false
		} else {
			// Check if any field in the row matches the email regex
			hasEmail := "false"
			if slices.ContainsFunc(record, emailRegexp.MatchString) {
				hasEmail = "true"
			}
			record = append(record, hasEmail)
		}
		// Write the modified record to output
		w.Write(record)
	}
	// Flush any buffered data to ensure it's written
	w.Flush()
	return buf.Bytes(), nil
}
