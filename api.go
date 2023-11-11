package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"valyx/aggregator/types"
)

// Server holds dependencies for a HTTP server.
type Server struct {
	QueryService *Service
}

// NewServer creates a new HTTP server with dependencies.
func NewServer(queryService *Service) *Server {

	return &Server{
		QueryService: queryService,
	}
}

// SearchHandler handles the /search endpoint.
func (s *Server) SearchHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request parameters
	fmt.Println(r.URL.Query())
	keyword := r.URL.Query().Get("keyword")
	accounts := r.URL.Query()["accounts"] // This is how to get []string query params
	start, end := r.URL.Query().Get("start"), r.URL.Query().Get("end")

	start = start + "T00:00:00Z"
	end = end + "T00:00:00Z"
	// Convert start and end to time.Time
	startTime, endTime, err := parseTimeRange(start, end)
	if err != nil {
		errorMsg := fmt.Sprintf("Invalid date format: %v. Please use YYYY-MM-DD.", err.Error())
		http.Error(w, errorMsg, http.StatusBadRequest)
		return
	}

	// Perform search
	results, err := s.QueryService.Search(keyword, accounts, startTime, endTime)
	if err != nil {
		http.Error(w, "Failed to perform search", http.StatusInternalServerError)
		return
	}

	// Respond with results
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, "Failed to encode results", http.StatusInternalServerError)
		return
	}
}

// parseTimeRange parses the start and end query parameters into time.Time.
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
