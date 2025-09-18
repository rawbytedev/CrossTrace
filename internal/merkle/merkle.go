package mptree

import "crosstrace/internal/merkle/merkletree"

// item represent an hashed representation of the data
type MerkleTree interface {
	Insert(item [][]byte) bool
	Delete(item [][]byte) bool
	Commit() bool
	Contains(item [][]byte) (bool, error)
	Proof(item []byte) ([][]byte, []int64 ,error)
	Root() []byte
}

type customMerkle struct {
	merkletree.Merkle
}

func NewMerkleTree() MerkleTree {
	return &customMerkle{}
}
