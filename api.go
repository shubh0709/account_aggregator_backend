package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"valyx/aggregator/types"
)

type Server struct {
	QueryService *Service
}

func NewServer(queryService *Service) *Server {

	return &Server{
		QueryService: queryService,
	}
}

func (s *Server) SearchHandler(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	accounts := r.URL.Query()["accounts"]
	sortOrder := r.URL.Query().Get("sort")
	start, end := r.URL.Query().Get("start"), r.URL.Query().Get("end")
	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		pageStr = "1"
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		http.Error(w, "Invalid page parameter. It must be a number.", http.StatusBadRequest)
		return
	}

	limit := 30
	offset := (page - 1) * limit
	if start != "" {
		start = start + "T00:00:00Z"
	}
	if end != "" {
		end = end + "T00:00:00Z"
	}
	startTime, endTime, err := parseTimeRange(start, end)
	if err != nil {
		errorMsg := fmt.Sprintf("Invalid date format: %v. Please use YYYY-MM-DD.", err.Error())
		http.Error(w, errorMsg, http.StatusBadRequest)
		return
	}

	results, err := s.QueryService.SearchWithPagination(keyword, accounts, startTime, endTime, limit, offset, sortOrder)
	if err != nil {
		http.Error(w, "Failed to perform search", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, "Failed to encode results", http.StatusInternalServerError)
		return
	}
}

func parseTimeRange(start, end string) (startTime, endTime time.Time, err error) {
	if start != "" {
		startTime, err = time.Parse(time.RFC3339, start)
		if err != nil {
			return
		}
	}
	if end != "" {
		endTime, err = time.Parse(time.RFC3339, end)
		if err != nil {
			return
		}
	}
	return
}

func (s *Server) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	keywords, err := s.QueryService.GetKeywords()
	if err != nil {
		http.Error(w, "Failed to fetch keywords", http.StatusInternalServerError)
		return
	}

	bankAccounts, err := s.QueryService.GetAllBankAccounts()
	if err != nil {
		http.Error(w, "Failed to fetch accounts", http.StatusInternalServerError)
		return
	}

	data := types.UserInfo{
		Keywords:     keywords,
		BankAccounts: bankAccounts,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode keywords", http.StatusInternalServerError)
		return
	}
}

func (s *Server) TrendHandler(w http.ResponseWriter, r *http.Request) {

	keyword := r.URL.Query().Get("keyword")
	start, end := r.URL.Query().Get("start"), r.URL.Query().Get("end")

	if start != "" {
		start = start + "T00:00:00Z"
	}
	if end != "" {
		end = end + "T00:00:00Z"
	}
	startTime, endTime, err := parseTimeRange(start, end)
	if err != nil {
		errorMsg := fmt.Sprintf("Invalid date format: %v. Please use YYYY-MM-DD.", err.Error())
		http.Error(w, errorMsg, http.StatusBadRequest)
		return
	}

	trendData, err := s.QueryService.GetTrends(keyword, startTime, endTime)
	if err != nil {
		http.Error(w, "Failed to fetch trend data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(trendData); err != nil {
		http.Error(w, "Failed to encode trend data", http.StatusInternalServerError)
		return
	}
}

func (s *Server) AggregateHandler(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	start, end := r.URL.Query().Get("start"), r.URL.Query().Get("end")

	if start != "" {
		start = start + "T00:00:00Z"
	}
	if end != "" {
		end = end + "T00:00:00Z"
	}
	startTime, endTime, err := parseTimeRange(start, end)
	if err != nil {
		errorMsg := fmt.Sprintf("Invalid date format: %v. Please use YYYY-MM-DD.", err.Error())
		http.Error(w, errorMsg, http.StatusBadRequest)
		return
	}

	aggregateData, err := s.QueryService.GetAggregates(keyword, startTime, endTime)
	if err != nil {
		http.Error(w, "Failed to fetch aggregate data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(aggregateData); err != nil {
		http.Error(w, "Failed to encode aggregate data", http.StatusInternalServerError)
		return
	}
}
