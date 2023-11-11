package utils

import (
	"log"
	"net/http"
)

// Middleware type is a function that takes a http.Handler and returns another http.Handler.
type Middleware func(http.Handler) http.Handler

// applyMiddleware wraps a http.Handler with multiple middleware.
func ApplyMiddleware(h http.Handler, middleware ...Middleware) http.Handler {
	// Apply middleware in reverse order
	for _, m := range middleware {
		h = m(h)
	}
	return h
}

// enableCORS is a middleware that enables CORS for the request.
func EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set headers to allow CORS
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware logs the incoming HTTP request
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the incoming request
		log.Printf("Method: %s, URI: %s, IP: %s\n", r.Method, r.RequestURI, r.RemoteAddr)
		// Call the next handler, which can be another middleware in the chain or the final handler
		next.ServeHTTP(w, r)
	})
}