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
func NewBadgerdb(cfg dbconfig.Config) (StorageDB, error) {
	opts := badger.DefaultOptions(cfg.DataDir)
	if cfg.BadgerValueLogSize != "" {
		size, err := dbconfig.ParseSize(cfg.BadgerValueLogSize)
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
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, data)
	})
}

// Get retrieves the value for a given key. Returns an error if not found.
func (b *badgerdb) Get(key []byte) ([]byte, error) {
	if b.db == nil {
		return nil, badger.ErrInvalidRequest
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
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

// BatchPut adds a key-value pair to the current batch. If last is true, flushes the batch.
// Returns error if batch operation fails.
func (b *badgerdb) BatchPut(key []byte, data []byte, last bool) error {
	if b.batch == nil {
		return badger.ErrInvalidRequest
	}
	err := b.batch.Set(key, data)
	if err != nil {
		b.batch.Cancel()
		return err
	}
	if last {
		if err := b.batch.Flush(); err != nil {
			return err
		}
	}
	return nil
}

// BatchDel adds a delete operation to the current batch. If last is true, flushes the batch.
func (b *badgerdb) BatchDel(key []byte, last bool) error {
	if b.batch == nil {
		return badger.ErrInvalidRequest
	}
	err := b.batch.Delete(key)
	if err != nil {
		b.batch.Cancel()
		return err
	}
	if last {
		if err := b.batch.Flush(); err != nil {
			return err
		}
	}
	return nil
}

// Close closes the database and releases all resources.
func (b *badgerdb) Close() error {
	if b.batch != nil {
		if err := b.batch.Flush(); err != nil {
			return err
		}
		b.batch.Cancel()
	}
	if b.db != nil {
		return b.db.Close()
	}
	return nil
}
