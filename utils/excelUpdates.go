package utils

import (
	"database/sql"
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

func ReadExcelFiles(db *sql.DB) {
	createTable(db)

	// Process CSV files from the folder
	err := filepath.Walk("./dummyData", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".csv" {
			processCSVFile(path, db)
		}
		return nil
	})

	if err != nil {
		panic(err)
	}
}

func createTable(db *sql.DB) {
	// Create the table with appropriate types
	query := `CREATE TABLE IF NOT EXISTS accounts_combined (
        Date DATE,
        Description TEXT,
        Debit NUMERIC(10, 2),
        Credit NUMERIC(10, 2),
        Balance NUMERIC(15, 2)
    )`
	_, err := db.Exec(query)
	if err != nil {
		panic(err)
	}
}

func processCSVFile(filePath string, db *sql.DB) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
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
			panic(err)
		}

		// Process and update record in the database
		updateDatabase(record, db)
	}
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

func updateDatabase(record []string, db *sql.DB) {
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

	// Prepare SQL statement using placeholders for the values
	query := `INSERT INTO accounts_combined (Date, Description, Debit, Credit, Balance)
              VALUES ($1, $2, $3, $4, $5)`

	// Use prepared statement to handle NULL values properly
	stmt, err := db.Prepare(query)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	// Execute the prepared statement
	_, err = stmt.Exec(formattedDate, record[1], debit, credit, balance)
	if err != nil {
		panic(err)
	}
}
func stringToNull(s string) sql.NullString {
	if s == "NULL" {
		return sql.NullString{String: s, Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
