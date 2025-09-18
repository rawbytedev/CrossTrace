// Copyright 2017 Cameron Bergoon
// Licensed under the MIT License, see LICENCE file for details.
// a modified version @rawbytedev
package merkletree

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"slices"
)

// Note it must be hashed before being passed to node
// by default sha256 is used to build the tree but for leafs any hash is accepted
func isEqual(a, b []byte) bool {
	return bytes.Equal(a, b)
}

// Content represents the data that is stored and verified by the tree. A type that
// implements this interface can be used as an item in the tree.
type Content interface {
	CalculateHash() ([]byte, error)
	Equals(other Content) (bool, error)
}

// MerkleTree is the container for the tree. It holds a pointer to the root of the tree,
// a list of pointers to the leaf nodes, and the merkle root.
type MerkleTree struct {
	Root         *Node
	merkleRoot   []byte
	Leafs        []*Node
	hashStrategy func() hash.Hash
}

// Node represents a node, root, or leaf in the tree. It stores pointers to its immediate
// relationships, a hash, the content stored if it is a leaf, and other metadata.

type Node struct {
	Tree   *MerkleTree
	Parent *Node
	Left   *Node
	Right  *Node
	leaf   bool
	dup    bool
	Hash   []byte
	//data   []byte // removed
}

// verifyNode walks down the tree until hitting a leaf, calculating the hash at each level
// and returning the resulting hash of Node n.
func (n *Node) verifyNode() ([]byte, error) {
	if n.leaf {
		return n.Hash, nil
	}
	rightBytes, err := n.Right.verifyNode()
	if err != nil {
		return nil, err
	}

	leftBytes, err := n.Left.verifyNode()
	if err != nil {
		return nil, err
	}

	h := n.Tree.hashStrategy()
	if _, err := h.Write(append(leftBytes, rightBytes...)); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// NewTree creates a new Merkle Tree using the content data.
func NewTree(data [][]byte) (*MerkleTree, error) {
	var defaultHashStrategy = sha256.New
	t := &MerkleTree{
		hashStrategy: defaultHashStrategy,
	}
	root, leafs, err := buildWithContent(data, t)
	if err != nil {
		return nil, err
	}
	t.Root = root
	t.Leafs = leafs
	t.merkleRoot = root.Hash
	return t, nil
}

// NewTreeWithHashStrategy creates a new Merkle Tree using the content cs using the provided hash
// strategy. Note that the hash type used in the type that implements the Content interface must
// match the hash type profided to the tree.
func NewTreeWithHashStrategy(data [][]byte, hashStrategy func() hash.Hash) (*MerkleTree, error) {
	t := &MerkleTree{
		hashStrategy: hashStrategy,
	}
	root, leafs, err := buildWithContent(data, t)
	if err != nil {
		return nil, err
	}
	t.Root = root
	t.Leafs = leafs
	t.merkleRoot = root.Hash
	return t, nil
}

// GetMerklePath: Get Merkle path and indexes(left leaf or right leaf)
func (m *MerkleTree) GetMerklePath(data []byte) ([][]byte, []int64, error) {
	for _, current := range m.Leafs {
		ok := isEqual(current.Hash, data)
		if ok {
			currentParent := current.Parent
			var merklePath [][]byte
			var index []int64
			for currentParent != nil {
				if bytes.Equal(currentParent.Left.Hash, current.Hash) {
					merklePath = append(merklePath, currentParent.Right.Hash)
					index = append(index, 1) // right leaf
				} else {
					merklePath = append(merklePath, currentParent.Left.Hash)
					index = append(index, 0) // left leaf
				}
				current = currentParent
				currentParent = currentParent.Parent
			}
			return merklePath, index, nil
		}
	}
	return nil, nil, nil
}

// buildWithContent is a helper function that for a given set of Contents, generates a
// corresponding tree and returns the root node, a list of leaf nodes, and a possible error.
// Returns an error if cs contains no Contents.
func buildWithContent(data [][]byte, t *MerkleTree) (*Node, []*Node, error) {
	if len(data) == 0 {
		return nil, nil, errors.New("error: cannot construct tree with no content")
	}
	var leafs []*Node
	for _, c := range data {
		leafs = append(leafs, &Node{
			Hash: c,
			leaf: true,
			Tree: t,
		})
	}
	if len(leafs)%2 == 1 {
		duplicate := &Node{
			Hash: leafs[len(leafs)-1].Hash,
			leaf: true,
			dup:  true,
			Tree: t,
		}
		leafs = append(leafs, duplicate)
	}
	root, err := buildIntermediate(leafs, t)
	if err != nil {
		return nil, nil, err
	}

	return root, leafs, nil
}

// buildIntermediate is a helper function that for a given list of leaf nodes, constructs
// the intermediate and root levels of the tree. Returns the resulting root node of the tree.
func buildIntermediate(nl []*Node, t *MerkleTree) (*Node, error) {
	var nodes []*Node
	for i := 0; i < len(nl); i += 2 {
		h := t.hashStrategy()
		var left, right int = i, i + 1
		if i+1 == len(nl) {
			right = i
		}
		chash := append(nl[left].Hash, nl[right].Hash...)
		if _, err := h.Write(chash); err != nil {
			return nil, err
		}
		n := &Node{
			Left:  nl[left],
			Right: nl[right],
			Hash:  h.Sum(nil),
			Tree:  t,
		}
		nodes = append(nodes, n)
		nl[left].Parent = n
		nl[right].Parent = n
		if len(nl) == 2 {
			return n, nil
		}
	}
	return buildIntermediate(nodes, t)
}

// MerkleRoot returns the unverified Merkle Root (hash of the root node) of the tree.
func (m *MerkleTree) MerkleRoot() []byte {
	return m.merkleRoot
}

// RebuildTree is a helper function that will rebuild the tree reusing only the content that
// it holds in the leaves.
func (m *MerkleTree) RebuildTree() error {
	var data [][]byte
	for _, c := range m.Leafs {
		data = append(data, c.Hash)
	}
	root, leafs, err := buildWithContent(data, m)
	if err != nil {
		return err
	}
	m.Root = root
	m.Leafs = leafs
	m.merkleRoot = root.Hash
	return nil
}

// Removal of elements from merkle tree
func (m *MerkleTree) Delete(data [][]byte) error {
	if len(data) == 0 {
		return errors.New("error: cannot construct tree with no content")
	}
	leafs := m.Leafs
	// Remove content that can't be deleted from the tree
	data = slices.DeleteFunc(data, func(s []byte) bool {
		vr, err := m.VerifyContent(s)
		if err != nil {
			return false // if error, do not delete}
		}
		return !vr // if not verified, do not delete
	})
	for _, c := range data {
		leafs = slices.DeleteFunc(leafs, func(s *Node) bool {
			return isEqual(s.Hash, c)
		})
	}
	// build Tree
	return m.Build(leafs)
}

// custom Rebuild this will allow inserting of content into the merkle tree dynamically
func (m *MerkleTree) Insert(data [][]byte) error {
	if len(data) == 0 {
		return errors.New("error: cannot construct tree with no content")
	}
	leafs := m.Leafs
	for _, c := range data {
		leafs = append(leafs, &Node{
			Hash: c,
			leaf: true,
			Tree: m,
		})

	}
	// building logic will sep this later
	return m.Build(leafs)
}

// VerifyTree verify tree validates the hashes at each level of the tree and returns true if the
// resulting hash at the root of the tree matches the resulting root hash; returns false otherwise.
func (m *MerkleTree) VerifyTree() (bool, error) {
	calculatedMerkleRoot, err := m.Root.verifyNode()
	if err != nil {
		return false, err
	}
	if bytes.Equal(m.merkleRoot, calculatedMerkleRoot) {
		return true, nil
	}
	return false, nil
}

// Builder helper function uses the provided leafs to build a tree
func (m *MerkleTree) Build(leafs []*Node) error {
	if len(leafs)%2 == 1 {
		duplicate := &Node{
			Hash: leafs[len(leafs)-1].Hash,
			leaf: true,
			dup:  true,
			Tree: m,
		}
		leafs = append(leafs, duplicate)
	}
	root, err := buildIntermediate(leafs, m)
	if err != nil {
		return err
	}
	m.Root = root
	m.Leafs = leafs
	m.merkleRoot = root.Hash
	return nil
}

// VerifyContent indicates whether a given content is in the tree and the hashes are valid for that content.
// Returns true if the expected Merkle Root is equivalent to the Merkle root calculated on the critical path
// for a given content. Returns true if valid and false otherwise.
func (m *MerkleTree) VerifyContent(hash []byte) (bool, error) {
	for _, l := range m.Leafs {
		if isEqual(l.Hash, hash) {
			return true, nil
		}
	}
	return false, nil
}

// String returns a string representation of the node.
func (n *Node) String() string {
	return fmt.Sprintf("%t %t %v", n.leaf, n.dup, n.Hash)
}

// String returns a string representation of the tree. Only leaf nodes are included
// in the output.
func (m *MerkleTree) String() string {
	s := ""
	for _, l := range m.Leafs {
		s += fmt.Sprint(l)
		s += "\n"
	}
	return s
}
