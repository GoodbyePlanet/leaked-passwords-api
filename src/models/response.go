package models

import "leaked-passwords-api/src/repository"

type CheckPasswordResponse struct {
	PasswordHash string `json:"passwordHash"`
	BreachCount  uint64 `json:"breachCount"`
	IsLeaked     bool   `json:"leaked"`
}

type HashesListResponse struct {
	Hashes        []repository.HashEntry `json:"hashes"`
	TotalReturned int                    `json:"total_returned"`
}

type HashResponse struct {
	Hash  string `json:"hash"`
	Count string `json:"count"`
}
