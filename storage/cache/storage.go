package cache

import (
	"time"

	"github.com/thamaji/nkdb"
)

func New(storage nkdb.Storage, expires time.Duration) *Storage {
	return &Storage{
		storage: storage,
		expires: expires,
	}
}

type Storage struct {
	storage nkdb.Storage
	expires time.Duration

	updated time.Time
	records [][]string
}

func (storage *Storage) Load() ([][]string, error) {
	now := time.Now()

	if storage.records != nil && now.Sub(storage.updated) < storage.expires {
		return storage.records, nil
	}

	records, err := storage.storage.Load()
	if err != nil {
		return nil, err
	}

	storage.records = records
	storage.updated = now

	return records, nil
}

func (storage *Storage) Save(records [][]string) error {
	if err := storage.storage.Save(records); err != nil {
		storage.records = nil
		return err
	}

	storage.records = records

	return nil
}
