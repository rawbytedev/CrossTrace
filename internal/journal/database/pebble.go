// Copyright (C) Enviora
// This file is part of go-enviora
//
// PebbleDB implementation for StorageDB interface.
// Provides efficient key-value storage for blocks, transactions, and other data.
// Supports batch operations and is optimized for concurrent access.
//
// Usage:
//   db, err := NewPebbledb(cfg)
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

	"github.com/cockroachdb/pebble"
)

// pebbledb manages Database Insert/Deletion/Batch Operations for Pebble.
// It only handles []byte keys and values.
type pebbledb struct {
	batch *pebble.Batch
	db    *pebble.DB
}

// NewPebbledb creates a new PebbleDB instance with the given config.
// Returns a StorageDB interface or an error if initialization fails.
func NewPebbledb(cfg dbconfig.Config) (StorageDB, error) {
	opts := &pebble.Options{
		MaxOpenFiles:    5000,
		BytesPerSync:    1 << 20,
		WALBytesPerSync: 1 << 20,
	}
	if cfg.PebbleCacheSize != "" {
		size, err := dbconfig.ParseSize(cfg.PebbleCacheSize)
		if err != nil {
			return nil, err
		}
		opts.Cache = pebble.NewCache(int64(size))
	}
	db, err := pebble.Open(cfg.DataDir, opts)
	if err != nil {
		return nil, err
	}
	batch := db.NewBatch()
	if batch == nil {
		db.Close()
		return nil, err
	}
	return &pebbledb{db: db, batch: batch}, nil
}

// Put inserts or updates a key-value pair in the database.
func (p *pebbledb) Put(key []byte, data []byte) error {
	if p.db == nil {
		return pebble.ErrClosed
	}
	return p.db.Set(key, data, pebble.Sync)
}

// Get retrieves the value for a given key. Returns an error if not found.
func (p *pebbledb) Get(key []byte) ([]byte, error) {
	if p.db == nil {
		return nil, pebble.ErrClosed
	}
	val, closer, err := p.db.Get(key)
	if err != nil {
		return nil, err
	}
	defer closer.Close()
	return val, nil
}

// BatchPut adds a key-value pair to the current batch. If last is true, commits the batch.
func (p *pebbledb) BatchPut(key []byte, data []byte, last bool) error {
	if p.batch == nil {
		return pebble.ErrClosed
	}
	if err := p.batch.Set(key, data, pebble.NoSync); err != nil {
		p.batch.Close()
		return err
	}
	if last {
		if err := p.batch.Commit(pebble.Sync); err != nil {
			return err
		}
	}
	return nil
}

// BatchDel adds a delete operation to the current batch. If last is true, commits the batch.
func (p *pebbledb) BatchDel(key []byte, last bool) error {
	if p.batch == nil {
		return pebble.ErrClosed
	}
	if err := p.batch.Delete(key, pebble.NoSync); err != nil {
		p.batch.Close()
		return err
	}
	if last {
		if err := p.batch.Commit(pebble.Sync); err != nil {
			return err
		}
	}
	return nil
}

// Del deletes a key-value pair from the database.
func (p *pebbledb) Del(key []byte) error {
	if p.db == nil {
		return pebble.ErrClosed
	}
	return p.db.Delete(key, pebble.Sync)
}

// Close closes the database and releases all resources.
func (p *pebbledb) Close() error {
	if p.batch != nil {
		if err := p.batch.Commit(pebble.Sync); err != nil {
			return err
		}
		p.batch.Close()
	}
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}
