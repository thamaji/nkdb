package csv

import (
	"encoding/csv"
	"os"
)

func New(path string) *Storage {
	return &Storage{
		path: path,
	}
}

type Storage struct {
	path string
}

func (storage *Storage) Load() ([][]string, error) {
	file, err := os.OpenFile(storage.path, os.O_RDONLY, 0)
	if err != nil {
		if os.IsNotExist(err) {
			return [][]string{}, nil
		}

		return nil, err
	}
	defer file.Close()

	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (storage *Storage) Save(records [][]string) error {
	file, err := os.OpenFile(storage.path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0664)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		if err := os.MkdirAll(storage.path, 0775); err != nil {
			return err
		}

		file, err = os.OpenFile(storage.path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0664)
		if err != nil {
			return err
		}
	}
	defer file.Close()

	if err := csv.NewWriter(file).WriteAll(records); err != nil {
		return err
	}

	if err := file.Sync(); err != nil {
		return err
	}

	return nil
}
