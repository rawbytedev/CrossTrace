package merkletree

import (
	"crypto/rand"
	"testing"
)

func TestInsert(t *testing.T) {
	m := NewMerkleTree(10)
	list := generatelist(5)
	for _, item := range list {
		m.Insert(FromByte(item))
	}
	m.Commit()
	for _, item := range list {
		vr, err := m.Contains(item)
		if err != nil {
			t.Log(err)
			t.Fail()
		}
		if !vr {
			t.Fail()
		}

	}
	sid, path, err := m.Proof(list[0])
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log(sid)
	t.Log(path)
}

func generatelist(size int) [][]byte {
	list2 := make([][]byte, 0)
	list1 := make([]byte, 32)
	for range size {
		rand.Read(list1)
		list2 = append(list2, list1)
		list1 = make([]byte, 32)
	}
	return list2
}
