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

func (db *PostgresDB) Close() error {
	return db.DB.Close()
}
