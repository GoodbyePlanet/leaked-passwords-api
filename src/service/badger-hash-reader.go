package service

import (
	"errors"
	"github.com/dgraph-io/badger/v4"
)

type BadgerHashReader struct {
	database *badger.DB
}

type HashEntry struct {
	Key   string
	Value string
}

func NewBadgerHashReader(database *badger.DB) *BadgerHashReader {
	return &BadgerHashReader{database: database}
}

func (db *BadgerHashReader) GetByHash(passwordHash string) (*HashEntry, error) {
	var entry *HashEntry

	err := db.database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(passwordHash))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				entry = nil
				return nil
			}
			return err
		}

		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		entry = &HashEntry{
			Key:   passwordHash,
			Value: string(val),
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return entry, nil
}

func (db *BadgerHashReader) GetAll() ([]HashEntry, error) {
	var entries []HashEntry

	err := db.database.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.KeyCopy(nil)

			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}

			entries = append(entries, HashEntry{
				Key:   string(key),
				Value: string(val),
			})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return entries, nil
}
