package types

import (
	"database/sql"
	"time"
)

// DB is an interface for our database operations. It allows us to abstract away
// the specific database we are using (PostgreSQL) and makes our code more
// flexible and testable.
type DB interface {
	InsertTransaction(t Transaction) error
	QueryTransactions(keyword string, accounts []string, startTime, endTime time.Time) ([]Transaction, error)
	GetUniqueKeywords() ([]string, error)
	GetUniqueBankAccounts() ([]string, error)
	QueryTransactionsWithPagination(keyword string, accounts []string, startTime, endTime time.Time, limit, offset int, sortOrder string) ([]Transaction, error)
	Close() error
}

// Transaction is a struct representing a bank transaction.
type Transaction struct {
	Date        string
	Description string
	Debit       sql.NullFloat64
	Credit      sql.NullFloat64
	Balance     sql.NullFloat64
	AccountID   string // This will link the transaction to a specific bank account
}
