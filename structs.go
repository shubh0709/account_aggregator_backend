package main

import "github.com/dgrijalva/jwt-go"

type User struct {
	UserID   string `json:"userid"`
	Password string `json:"password"`
}

type Claims struct {
	UserID string `json:"userid"`
	jwt.StandardClaims
}
