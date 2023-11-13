package utils

import (
	"log"
	"net/http"
	"strings"

	"github.com/spf13/viper"
)

type Middleware func(http.Handler) http.Handler

func ApplyMiddleware(h http.Handler, middleware ...Middleware) http.Handler {
	for _, m := range middleware {
		h = m(h)
	}
	return h
}

func EnableCORS() func(http.Handler) http.Handler {
	allowedOrigins := viper.GetString("ALLOWED_ORIGINS")

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")

            for _, allowedOrigin := range strings.Split(allowedOrigins, ",") {
                if origin == allowedOrigin {
                    w.Header().Set("Access-Control-Allow-Origin", origin)
                    break
                }
            }

            w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
            w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusNoContent)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}


func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Method: %s, URI: %s, IP: %s\n", r.Method, r.RequestURI, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}