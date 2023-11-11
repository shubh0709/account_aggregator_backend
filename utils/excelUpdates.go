package utils

import (
	"database/sql"
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"valyx/aggregator/types"

	_ "github.com/lib/pq"
)

type Processor struct {
	db types.DB
}

// New creates a new processor with dependencies.
func NewProcessor(db types.DB) *Processor {
	return &Processor{db: db}
}

func (p *Processor) ReadExcelFiles(path string, db types.DB) error {
	// createTable(db)

	// Process CSV files from the folder
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".csv" {
			accountId := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
			if err := p.processCSVFile(path, db, accountId); err != nil {
				return err
			}
		}
		return nil
	})
}

func (p *Processor) processCSVFile(filePath string, db types.DB, accountId string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Read() // Skip header line
	for {
		// Read each record
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Process and update record in the database
		if err := p.processData(record, accountId); err != nil {
			return err
		}
	}

	return nil
}

func (p *Processor) processData(record []string, accountId string) error {
	// Parse the date from the CSV format to Go's time.Time type
	parsedDate, err := time.Parse("02/01/2006", record[0])
	if err != nil {
		panic(err)
	}
	// Format the date to PostgreSQL's accepted format
	formattedDate := parsedDate.Format("2006-01-02")

	// Convert empty strings to SQL null values for numeric fields
	debit := stringToNullNumeric(record[2])
	credit := stringToNullNumeric(record[3])
	balance := stringToNullNumeric(record[4])

	transaction := types.Transaction{
		AccountID:   accountId,
		Date:        formattedDate,
		Description: record[1],
		Debit:       debit,
		Credit:      credit,
		Balance:     balance,
	}

	return p.db.InsertTransaction(transaction)
}

func stringToNullNumeric(s string) sql.NullFloat64 {
	if s == "" {
		return sql.NullFloat64{Float64: 0, Valid: false}
	}
	var f float64
	var err error
	if f, err = strconv.ParseFloat(s, 64); err != nil {
		panic(err)
	}
	return sql.NullFloat64{Float64: f, Valid: true}
}

func stringToNull(s string) sql.NullString {
	if s == "NULL" {
		return sql.NullString{String: s, Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
