package types

import (
	"database/sql"
	"time"
)

type DB interface {
	InsertTransaction(t Transaction) error
	QueryTransactions(keyword string, accounts []string, startTime, endTime time.Time) ([]Transaction, error)
	GetUniqueKeywords() ([]string, error)
	GetUniqueBankAccounts() ([]string, error)
	QueryTransactionsWithPagination(keyword string, accounts []string, startTime, endTime time.Time, limit, offset int, sortOrder string) ([]Transaction, error)
	GetTrendData(category string, startTime, endTime time.Time) ([]TrendData, error)
	GetAggregateData(category string, startTime, endTime time.Time) (AggregateData, error)
	Close() error
}

type Transaction struct {
	Date        string
	Description string
	Debit       sql.NullFloat64
	Credit      sql.NullFloat64
	Balance     sql.NullFloat64
	AccountID   string
}

type TrendData struct {
	Period      string  `json:"period"`
	TotalCredit float64 `json:"total_credit"`
	TotalDebit  float64 `json:"total_debit"`
}

type AggregateData struct {
	Category    string  `json:"category"`
	Total       float64 `json:"total"`
	TotalCredit float64 `json:"total_credit"`
	TotalDebit  float64 `json:"total_debit"`
}
