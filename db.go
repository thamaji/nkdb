package nkdb

import (
	"errors"
	"sync"
)

func New(key int, storage Storage) *DB {
	return &DB{
		key:     key,
		storage: storage,
		mutex:   &sync.RWMutex{},
	}
}

type DB struct {
	key     int
	storage Storage
	mutex   *sync.RWMutex
}

func (db *DB) Keys() ([]string, error) {
	const Op = "Keys"

	db.mutex.RLock()
	defer db.mutex.RUnlock()

	records, err := db.storage.Load()
	if err != nil {
		return nil, &Error{Type: ErrInternal, Op: Op, Key: "", Err: err}
	}

	keys := make([]string, 0, len(records))

	for i := range records {
		if db.key >= len(records[i]) {
			continue // error?
		}

		keys = append(keys, records[i][db.key])
	}

	return keys, nil
}

func (db *DB) Get(key string) ([]string, error) {
	const Op = "Get"

	db.mutex.RLock()
	defer db.mutex.RUnlock()

	records, err := db.storage.Load()
	if err != nil {
		return nil, &Error{Type: ErrInternal, Op: Op, Key: key, Err: err}
	}

	for i := range records {
		if db.key >= len(records[i]) {
			continue // error?
		}

		if key == records[i][db.key] {
			record := make([]string, len(records[i]))
			copy(record, records[i])
			return record, nil
		}
	}

	return nil, &Error{Type: ErrNotExist, Op: Op, Key: key, Err: errors.New("record dose not exist")}
}

func (db *DB) Set(key string, record []string) error {
	const Op = "Set"

	if db.key >= len(record) {
		return &Error{Type: ErrInvalid, Op: Op, Key: key, Err: errors.New("record must have a key field")}
	}

	if key != record[db.key] {
		return &Error{Type: ErrInvalid, Op: Op, Key: key, Err: errors.New("key must be equals key field")}
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	records, err := db.storage.Load()
	if err != nil {
		return &Error{Type: ErrInternal, Op: Op, Key: key, Err: err}
	}

	overwrite := false
	for i := range records {
		if db.key >= len(records[i]) {
			continue // error?
		}

		if key == records[i][db.key] {
			records[i] = record
			overwrite = true
			break
		}
	}

	if !overwrite {
		records = append(records, record)
	}

	if err := db.storage.Save(records); err != nil {
		return &Error{Type: ErrInternal, Op: Op, Key: key, Err: err}
	}

	return nil
}

func (db *DB) Delete(key string) error {
	const Op = "Delete"

	db.mutex.Lock()
	defer db.mutex.Unlock()

	records, err := db.storage.Load()
	if err != nil {
		return &Error{Type: ErrInternal, Op: Op, Key: key, Err: err}
	}

	for i := range records {
		if db.key >= len(records[i]) {
			continue
		}

		if key == records[i][db.key] {
			records = append(records[:i], records[i+1:]...)
			break
		}
	}

	if err := db.storage.Save(records); err != nil {
		return &Error{Type: ErrInternal, Op: Op, Key: key, Err: err}
	}

	return nil
}

func (db *DB) View(f func(*Tx) error) error {
	const Op = "View"

	tx, err := db.Begin(false)
	if err != nil {
		return &Error{Type: ErrInternal, Op: Op, Key: "", Err: err}
	}
	defer tx.Rollback()

	if err := f(tx); err != nil {
		return err
	}

	return nil
}

func (db *DB) Update(f func(*Tx) error) error {
	const Op = "Update"

	tx, err := db.Begin(true)
	if err != nil {
		return &Error{Type: ErrInternal, Op: Op, Key: "", Err: err}
	}
	defer tx.Rollback()

	if err := f(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return &Error{Type: ErrInternal, Op: Op, Key: "", Err: err}
	}

	return nil
}

func (db *DB) Begin(writable bool) (*Tx, error) {
	const Op = "Begin"

	if writable {
		db.mutex.Lock()
	} else {
		db.mutex.RLock()
	}

	records, err := db.storage.Load()
	if err != nil {
		if writable {
			db.mutex.Unlock()
		} else {
			db.mutex.RUnlock()
		}
		return nil, &Error{Type: ErrInternal, Op: Op, Key: "", Err: err}
	}

	tx := &Tx{
		db:       db,
		writable: writable,
		keys:     make([]string, 0, len(records)),
		records:  make(map[string][]string, len(records)),
	}

	for i := range records {
		if db.key >= len(records[i]) {
			continue // error?
		}

		record := records[i]
		key := record[db.key]
		tx.keys = append(tx.keys, key)
		tx.records[key] = record
	}

	return tx, nil
}

type Tx struct {
	db       *DB
	writable bool
	keys     []string
	records  map[string][]string
}

func (tx *Tx) Commit() error {
	const Op = "Commit"

	if !tx.writable {
		return &Error{Type: ErrInvalid, Op: Op, Key: "", Err: errors.New("invalid operation")}
	}

	records := make([][]string, 0, len(tx.keys))
	for _, key := range tx.keys {
		records = append(records, tx.records[key])
	}

	if err := tx.db.storage.Save(records); err != nil {
		return &Error{Type: ErrInternal, Op: Op, Key: "", Err: err}
	}

	return nil
}

func (tx *Tx) Rollback() {
	if tx.writable {
		tx.db.mutex.Unlock()
	} else {
		tx.db.mutex.RUnlock()
	}
}

func (tx *Tx) Keys() ([]string, error) {
	out := make([]string, len(tx.keys))
	copy(out, tx.keys)
	return out, nil
}

func (tx *Tx) Get(key string) ([]string, error) {
	const Op = "Get"

	record, ok := tx.records[key]
	if !ok {
		return nil, &Error{Type: ErrNotExist, Op: Op, Key: key, Err: errors.New("record dose not exist")}
	}

	out := make([]string, len(record))
	copy(out, record)
	return out, nil
}

func (tx *Tx) Set(key string, record []string) error {
	const Op = "Set"

	if !tx.writable {
		return &Error{Type: ErrInvalid, Op: Op, Key: "", Err: errors.New("invalid operation")}
	}

	if tx.db.key >= len(record) {
		return &Error{Type: ErrInvalid, Op: Op, Key: key, Err: errors.New("record must have a key field")}
	}

	if key != record[tx.db.key] {
		return &Error{Type: ErrInvalid, Op: Op, Key: key, Err: errors.New("key must be equals key field")}
	}

	if _, ok := tx.records[key]; !ok {
		tx.keys = append(tx.keys, key)
	}

	tx.records[key] = record

	return nil
}

func (tx *Tx) Delete(key string) error {
	const Op = "Delete"

	if !tx.writable {
		return &Error{Type: ErrInvalid, Op: Op, Key: "", Err: errors.New("invalid operation")}
	}

	if _, ok := tx.records[key]; !ok {
		return &Error{Type: ErrNotExist, Op: Op, Key: key, Err: errors.New("record dose not exist")}
	}

	for i := range tx.keys {
		if key == tx.keys[i] {
			tx.keys = append(tx.keys[:i], tx.keys[i+1:]...)
			break
		}
	}

	delete(tx.records, key)

	return nil
}
