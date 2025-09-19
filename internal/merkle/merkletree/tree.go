package merkletree

import (
	"sync"
)
/*
This implementation is a helper to faciliate the usage of merkle tree
*/
type Merkle struct {
	m       *MerkleTree
	mu      sync.RWMutex
	pending [][]byte // list of pending data
}

// NewMerkle
func NewMerkleTree(maxitem int) *Merkle {
	return &Merkle{}
}

// Insert help function/ Pending handler
func (m *Merkle) Insert(item [][]byte) bool {
	if m.pending == nil {
		m.pending = make([][]byte, 0)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pending = append(m.pending, item...)
	return true

}
// Commit 
func (m *Merkle) Commit() bool {
	if m.m != nil {
		err := m.m.Insert(m.pending)
		return err == nil
	}
	var err error
	m.m, err = NewTree(m.pending)
	return err == nil
}

func (m *Merkle) Clean() {
	m.pending = m.pending[:0]
}

// Delete a sinle item or a list of item from merkle tree
// true = delete success; false= error
func (m *Merkle) Delete(item [][]byte) bool {
	err := m.m.Delete(item)
	return err == nil
}

// check whether an items is in the merkle tree or not
// true = is in tree; false = not in tree
func (m *Merkle) Contains(item []byte) (bool, error) {
	vr, err := m.m.VerifyContent(item)
	if err != nil {
		return vr, err
	}
	return vr, nil
}
func (m *Merkle) Root() []byte {
	return m.m.Root.Hash
}
func (m *Merkle) Proof(item []byte) ([][]byte, []int64, error) {
	vr, err := m.Contains(item)
	if err != nil {
		return nil, nil, err
	}
	if vr {
		return m.m.GetMerklePath(item)
	}
	return nil, nil, nil
}

// turns []byte to [][]byte
func FromByte(item []byte) [][]byte {
	var tmp [][]byte
	return append(tmp, item)
}

// turns [][]byte into []byte
func ToByte(item [][]byte) []byte {
	for _, i := range item {
		return i
	}
	return nil
}
