// Copyright (C) Enviora
// This file is part of go-enviora
//
// StorageDB is a generic key-value database interface for blockchain storage backends.
// Implementations must provide efficient and safe access to persistent storage for blocks, transactions, and state.
//
// Methods:
//
//	Put(key, data): Insert or update a key-value pair.
//	Get(key): Retrieve the value for a given key.
//	Del(key): Delete a key-value pair.
//	BatchPut(key, data, last): Add a key-value pair to a batch; flush if last is true.
//	BatchDel(key, last): Add a delete operation to a batch; flush if last is true.
//	Close(): Close the database and release all resources.
package database

// StorageDB defines the interface for a pluggable key-value store.
type StorageDB interface {
	// Put inserts or updates a key-value pair in the database.
	Put(key []byte, data []byte) error
	// Get retrieves the value for a given key. Returns an error if not found.
	Get(key []byte) ([]byte, error)
	// Del deletes a key-value pair from the database.
	Del(key []byte) error
	// BatchPut adds a key-value pair to the current batch. If last is true, flushes the batch.
	BatchPut(key []byte, data []byte) error
	// BatchDel adds a delete operation to the current batch. If last is true, flushes the batch.
	BatchDel(key []byte) error
	// Close closes the database and releases all resources.
	Close() error
}
