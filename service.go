package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
	"valyx/aggregator/types"

	_ "github.com/lib/pq"
	"github.com/ztrue/tracerr"
)

// Service provides methods to query transactions.
// It uses the types.DB interface, so it's not directly tied to PostgreSQL.
type Service struct {
	db types.DB
}

type PostgresDB struct {
	*sql.DB
}

func setupDB() (types.DB, error) {
	const (
		host     = "localhost"
		port     = 5432
		user     = "postgres"
		password = "12345"
		dbname   = "bank"
	)

	// Create a connection string.
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	// Check the connection
	err = db.Ping()
	if err != nil {
		tracerr.Wrap(err)
		tracerr.PrintSourceColor(err)
		log.Fatal(err)
		return nil, err
	}

	return &PostgresDB{DB: db}, nil
}

// NewService creates a new query service.
func NewService(db types.DB) *Service {
	return &Service{db: db}
}

// Search performs a search on transactions based on the provided criteria.
func (s *Service) Search(keyword string, accounts []string, startTime, endTime time.Time) ([]types.Transaction, error) {
	// The actual search logic will be implemented here.
	return s.db.QueryTransactions(keyword, accounts, startTime, endTime)
}

// Gets the user related information
// GetKeywords retrieves a list of unique keywords from transaction descriptions.
func (s *Service) GetKeywords() ([]string, error) {
	return s.db.GetUniqueKeywords()
}

// GetAccounts retrieves a list of  available bank accounts.
func (s *Service) GetAllBankAccounts() ([]string, error) {
	return s.db.GetUniqueBankAccounts()
}

// SearchWithPagination performs a search with pagination.
func (s *Service) SearchWithPagination(keyword string, accounts []string, startTime, endTime time.Time, limit, offset int, sortOrder string) ([]types.Transaction, error) {
	// The actual search logic with pagination will be implemented here.
	// This will involve modifying the SQL query to include LIMIT and OFFSET.
	return s.db.QueryTransactionsWithPagination(keyword, accounts, startTime, endTime, limit, offset, sortOrder)
}

// GetTrends retrieves trend data for a given category within a date range.
func (s *Service) GetTrends(category string, startTime, endTime time.Time) ([]types.TrendData, error) {
	return s.db.GetTrendData(category, startTime, endTime)
}

// GetAggregates retrieves aggregated data for a given category.
func (s *Service) GetAggregates(category string, startTime time.Time, endTime time.Time) (types.AggregateData, error) {
	return s.db.GetAggregateData(category, startTime, endTime)
}

// GetAccounts retrieves a list of available accounts from the transactions table.
func (db *PostgresDB) GetUniqueBankAccounts() ([]string, error) {
	const query = `SELECT DISTINCT account_id FROM transactions`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying accounts: %v", err)
	}
	defer rows.Close()

	var accounts []string
	for rows.Next() {
		var accountId string
		if err := rows.Scan(&accountId); err != nil {
			return nil, fmt.Errorf("error scanning account id: %v", err)
		}
		accounts = append(accounts, accountId)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error with rows during accounts fetching: %v", err)
	}

	return accounts, nil
}

// GetUniqueKeywords retrieves a list of unique keywords from transaction descriptions.
func (db *PostgresDB) GetUniqueKeywords() ([]string, error) {
	const query = `SELECT DISTINCT description FROM transactions`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying unique keywords: %v", err)
	}
	defer rows.Close()

	var keywords []string
	for rows.Next() {
		var description string
		if err := rows.Scan(&description); err != nil {
			return nil, fmt.Errorf("error scanning description: %v", err)
		}
		// Process the description to extract keywords
		// This can be as simple as adding the description to the list, or
		// you might want to split the description into words and add them individually
		keywords = append(keywords, description)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error with rows during keyword fetching: %v", err)
	}

	return keywords, nil
}

func (db *PostgresDB) createTable() error {
	// Create the table with appropriate types
	query := `CREATE TABLE IF NOT EXISTS transactions (
		Account_Id TEXT,
        Date DATE,
        Description TEXT,
        Debit NUMERIC(10, 2),
        Credit NUMERIC(10, 2),
        Balance NUMERIC(15, 2)
    )`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func (db *PostgresDB) InsertTransaction(t types.Transaction) error {
	if err := db.createTable(); err != nil {
		return fmt.Errorf("error creating transactions table: %v", err)
	}

	const query = `
        INSERT INTO transactions (account_id, date, description, debit, credit, balance)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err := db.Exec(query, t.AccountID, t.Date, t.Description, t.Debit, t.Credit, t.Balance)
	if err != nil {
		return fmt.Errorf("error inserting transaction: %v", err)
	}
	return nil
}

func (db *PostgresDB) QueryTransactions(keyword string, accounts []string, startTime, endTime time.Time) ([]types.Transaction, error) {
	var query strings.Builder
	query.WriteString(`
        SELECT account_id, date, description, debit, credit, balance
        FROM transactions
        WHERE description ILIKE $1
    `)
	params := []interface{}{"%" + keyword + "%"}
	paramID := 2

	if len(accounts) > 0 {
		query.WriteString(" AND account_id IN (")
		for i, account := range accounts {
			if i > 0 {
				query.WriteString(", ")
			}
			query.WriteString(fmt.Sprintf("$%d", paramID))
			params = append(params, account)
			paramID++
		}
		query.WriteString(")")
	}

	if !startTime.IsZero() {
		query.WriteString(fmt.Sprintf(" AND date >= $%d", paramID))
		params = append(params, startTime)
		paramID++
	}

	if !endTime.IsZero() {
		query.WriteString(fmt.Sprintf(" AND date <= $%d", paramID))
		params = append(params, endTime)
	}

	rows, err := db.Query(query.String(), params...)
	// fmt.Println(rows)
	if err != nil {
		return nil, fmt.Errorf("error querying transactions: %v", err)
	}
	defer rows.Close()

	var transactions []types.Transaction
	for rows.Next() {
		var t types.Transaction
		err := rows.Scan(&t.AccountID, &t.Date, &t.Description, &t.Debit, &t.Credit, &t.Balance)
		if err != nil {
			return nil, fmt.Errorf("error scanning transaction: %v", err)
		}
		transactions = append(transactions, t)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error with rows: %v", err)
	}

	return transactions, nil
}

// QueryTransactionsWithPagination retrieves transactions with pagination and sorting.
func (db *PostgresDB) QueryTransactionsWithPagination(keyword string, accounts []string, startTime, endTime time.Time, limit, offset int, sortOrder string) ([]types.Transaction, error) {
	var transactions []types.Transaction

	// Default sort order
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	// Start constructing the SQL query
	var queryBuilder strings.Builder
	queryBuilder.WriteString("SELECT account_id, date, description, debit, credit, balance FROM transactions WHERE 1=1")

	// Parameters slice for the query parameters
	var params []interface{}
	paramID := 1

	// Include keyword filter only if keyword is not empty
	if keyword != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND description ILIKE $%d", paramID))
		params = append(params, "%"+keyword+"%")
		paramID++
	}

	// Add account filtering if accounts are provided
	if len(accounts) > 0 {
		queryBuilder.WriteString(fmt.Sprintf(" AND account_id IN (%s)", paramPlaceholder(paramID, len(accounts))))
		for _, account := range accounts {
			params = append(params, account)
			paramID++
		}
	}

	// Add date range filtering if start and end times are not zero values
	if !startTime.IsZero() {
		queryBuilder.WriteString(fmt.Sprintf(" AND date >= $%d", paramID))
		params = append(params, startTime)
		paramID++
	}
	if !endTime.IsZero() {
		queryBuilder.WriteString(fmt.Sprintf(" AND date <= $%d", paramID))
		params = append(params, endTime)
		paramID++
	}

	// Add pagination and sorting using LIMIT and OFFSET
	queryBuilder.WriteString(fmt.Sprintf(" ORDER BY date %s LIMIT $%d OFFSET $%d", sortOrder, paramID, paramID+1))
	params = append(params, limit, offset)

	// Execute the query
	rows, err := db.Query(queryBuilder.String(), params...)
	if err != nil {
		return nil, fmt.Errorf("error querying transactions with pagination: %v", err)
	}
	defer rows.Close()

	// Iterate over the rows and scan the results into the transactions slice
	for rows.Next() {
		var t types.Transaction
		var date time.Time
		if err := rows.Scan(&t.AccountID, &date, &t.Description, &t.Debit, &t.Credit, &t.Balance); err != nil {
			return nil, fmt.Errorf("error scanning transaction row: %v", err)
		}
		// Convert the date to DD/MM/YYYY format
		t.Date = date.Format("02/01/2006")
		transactions = append(transactions, t)
	}

	// Check for any error that occurred during the iteration
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iteration error in QueryTransactionsWithPagination: %v", err)
	}

	return transactions, nil
}

// paramPlaceholder generates a string with placeholders for SQL IN clause.
func paramPlaceholder(start, count int) string {
	if count < 1 {
		return ""
	}
	placeholders := make([]string, count)
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", start+i)
	}
	return strings.Join(placeholders, ", ")
}

func (db *PostgresDB) Close() error {
	return db.DB.Close()
}

// ... other imports ...

func (db *PostgresDB) GetTrendData(category string, startTime, endTime time.Time) ([]types.TrendData, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString(`
        SELECT DATE_TRUNC('week', date) AS period, 
               COALESCE(SUM(credit), 0) AS total_credits, 
               COALESCE(SUM(debit), 0) AS total_debits
        FROM transactions
        WHERE description ILIKE $1
    `)

	params := []interface{}{"%" + category + "%"}
	paramIndex := 2

	if !startTime.IsZero() {
		queryBuilder.WriteString(fmt.Sprintf(" AND date >= $%d", paramIndex))
		params = append(params, startTime)
		paramIndex++
	}

	if !endTime.IsZero() {
		queryBuilder.WriteString(fmt.Sprintf(" AND date <= $%d", paramIndex))
		params = append(params, endTime)
		paramIndex++
	}

	queryBuilder.WriteString(" GROUP BY period ORDER BY period")

	var trends []types.TrendData
	rows, err := db.Query(queryBuilder.String(), params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var trend types.TrendData
		var period time.Time
		if err := rows.Scan(&period, &trend.TotalCredit, &trend.TotalDebit); err != nil {
			return nil, err
		}
		trend.Period = period.Format("02-01-2006") // Adjust the format as needed
		trends = append(trends, trend)
	}

	return trends, nil
}

func (db *PostgresDB) GetAggregateData(category string, startTime, endTime time.Time) (types.AggregateData, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString(`
        SELECT COALESCE(SUM(credit), 0) AS total_credits, 
               COALESCE(SUM(debit), 0) AS total_debits,
			   COALESCE(SUM(credit), 0) - COALESCE(SUM(debit), 0) AS total
        FROM transactions
        WHERE description ILIKE $1
    `)

	params := []interface{}{"%" + category + "%"}
	paramIndex := 2

	if !startTime.IsZero() {
		queryBuilder.WriteString(fmt.Sprintf(" AND date >= $%d", paramIndex))
		params = append(params, startTime)
		paramIndex++
	}

	if !endTime.IsZero() {
		queryBuilder.WriteString(fmt.Sprintf(" AND date <= $%d", paramIndex))
		params = append(params, endTime)
		paramIndex++
	}

	var aggregate types.AggregateData
	aggregate.Category = category

	err := db.QueryRow(queryBuilder.String(), params...).Scan(&aggregate.TotalCredit, &aggregate.TotalDebit, &aggregate.Total)
	if err != nil {
		if err == sql.ErrNoRows {
			// No rows were returned, but this isn't an error condition for aggregate queries
			aggregate.Total = 0
			log.Println("Error no row found ", err)

			return aggregate, nil
		}
		// Use log.Println here for debugging purposes; remove or adjust for production use
		log.Println("Error fetching aggregate data: ", err)
		return types.AggregateData{}, err
	}

	return aggregate, nil
}
