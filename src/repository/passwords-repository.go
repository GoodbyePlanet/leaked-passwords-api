package repository

import (
	"errors"
	"github.com/dgraph-io/badger/v4"
	"log/slog"
	"os"
	"strings"
)

type PasswordsRepository struct {
	database *badger.DB
	logger   *slog.Logger
}

type HashEntry struct {
	Key   string
	Value string
}

func NewPasswordsRepository(database *badger.DB) *PasswordsRepository {
	return &PasswordsRepository{
		database: database,
		logger:   slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
}

func (repo *PasswordsRepository) Save(blob []byte) {
	lines := strings.Split(string(blob), "\r\n")

	err := repo.database.Update(func(txn *badger.Txn) error {
		for _, line := range lines {
			if line == "" {
				continue
			}

			parts := strings.SplitN(line, ":", 2)

			if len(parts) != 2 {
				continue
			}

			hash := parts[0]
			count := parts[1]

			if err := txn.Set([]byte(hash), []byte(count)); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		repo.logger.Error("Failed to store data in Badger", slog.Any("error", err))
	}
}

func (repo *PasswordsRepository) GetByHash(passwordHash string) (*HashEntry, error) {
	repo.logger.Info("Get by password hash", "hash", passwordHash)
	var entry *HashEntry

	err := repo.database.View(func(txn *badger.Txn) error {
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

func (repo *PasswordsRepository) GetAll() ([]HashEntry, error) {
	repo.logger.Info("Getting all entries from DB")
	var entries []HashEntry

	err := repo.database.View(func(txn *badger.Txn) error {
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
