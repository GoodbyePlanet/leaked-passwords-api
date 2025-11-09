package db

import (
	"log"

	"github.com/dgraph-io/badger/v4"
)

var DB *badger.DB

func Init(path string) *badger.DB {
	opts := badger.DefaultOptions(path)

	var err error
	DB, err = badger.Open(opts)
	if err != nil {
		log.Fatalf("failed to open BadgerDB: %v", err)
	}

	return DB
}
