// Copyright (C) Enviora
// This file is part of go-enviora
//
// BadgerDB implementation for StorageDB interface.
// Provides efficient key-value storage for blocks, transactions, and other data.
// Supports batch operations and is optimized for concurrent access.
//
// Usage:
//   db, err := NewBadgerdb(cfg)
//   err = db.Put(key, value)
//   value, err := db.Get(key)
//   err = db.Del(key)
//   err = db.Close()
//
// Batch operations:
//   err = db.BatchPut(key, value, last)
//   err = db.BatchDel(key, last)

package database

import (
	dbconfig "crosstrace/internal/configs"

	"github.com/dgraph-io/badger/v4"
)

// badgerdb manages Database Insert/Deletion/Batch Operations.
// It only handles []byte keys and values.
type badgerdb struct {
	batch *badger.WriteBatch
	db    *badger.DB
}

// NewBadgerdb creates a new BadgerDB instance with the given config.
// Returns a StorageDB interface or an error if initialization fails.
func NewBadgerdb(cfg dbconfig.JournalConfig) (StorageDB, error) {
	opts := badger.DefaultOptions(cfg.DBPath)
	if cfg.LogSize != "" {
		size, err := dbconfig.ParseSize(cfg.LogSize)
		if err != nil {
			return nil, err
		}
		opts = opts.WithValueLogFileSize(int64(size))
	}

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	batch := db.NewWriteBatch()
	if batch == nil {
		db.Close()
		return nil, badger.ErrInvalidRequest
	}
	return &badgerdb{db: db, batch: batch}, nil
}

// Put inserts or updates a key-value pair in the database.
func (b *badgerdb) Put(key []byte, data []byte) error {
	if b.db == nil {
		return badger.ErrInvalidRequest
	}
	if key == nil {
		return ErrEmptydbKey
	}
	if data == nil {
		return ErrEmptydbValue
	}
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, data)
	})
}

// Get retrieves the value for a given key. Returns an error if not found.
func (b *badgerdb) Get(key []byte) ([]byte, error) {
	if b.db == nil {
		return nil, badger.ErrInvalidRequest
	}
	if key == nil {
		return nil, ErrEmptydbKey
	}
	var data []byte
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			data = append([]byte{}, val...)
			return nil
		})
	})
	return data, err
}

// Del deletes a key-value pair from the database.
func (b *badgerdb) Del(key []byte) error {
	if b.db == nil {
		return badger.ErrInvalidRequest
	}
	if key == nil {
		return ErrEmptydbKey
	}
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

// BatchPut adds a key-value pair to the current batch.
// Returns error if batch operation fails.
func (b *badgerdb) BatchPut(key []byte, data []byte) error {
	if b.batch == nil {
		return badger.ErrInvalidRequest
	}
	if key != nil && data != nil {
		err := b.batch.Set(key, data)
		if err != nil {
			b.batch.Cancel()
			return err
		}
	}
	if key == nil && data == nil {
		if err := b.batch.Flush(); err != nil {
			return err
		}
		return nil
	}
	if key == nil {
		return ErrEmptydbKey
	}
	if data == nil {
		return ErrEmptydbValue
	}

	return nil
}

// BatchDel adds a delete operation to the current batch.
func (b *badgerdb) BatchDel(key []byte) error {
	if b.batch == nil {
		return badger.ErrInvalidRequest
	}
	if key != nil {
		err := b.batch.Delete(key)
		if err != nil {
			b.batch.Cancel()
			return err
		}
	}

	if err := b.batch.Flush(); err != nil {
		return err
	}
	return nil
}

// Close closes the database and releases all resources.
func (b *badgerdb) Close() error {
	if b.batch != nil {
		// needs to works on plugging this or keeping records of batch
		// this is a fast workaround but error not reported if
		// flush faces errors
		if err := b.batch.Flush(); err != nil {
			b.batch.Cancel()
		}
	}
	if b.db != nil {
		return b.db.Close()
	}
	return nil
}
