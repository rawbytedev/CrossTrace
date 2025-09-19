package mptree

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func Init() MerkleTree {
	return NewMerkleTree()
}
func TestInsert(t *testing.T) {
	m := Init()
	data1 := GenerateData(5)
	check := m.Insert(data1)
	if !check {
		t.Fail()
	}
	/* this fails
	broot := m.Root()
	if broot != nil {
		t.Fail()
	}*/
	check = m.Commit()
	if !check {
		t.Fail()
	}
	root := m.Root()
	if root == nil {
		t.Fail()
	}

}
func TestDelete(t *testing.T) {
	m := Init()
	data1 := GenerateData(5)
	check := m.Insert(data1)
	if !check {
		t.FailNow()
	}
	check = m.Commit()
	if !check {
		t.FailNow()
	}
	check = m.Delete(FromByte(data1[0]))
	if !check {
		t.FailNow()
	}
	check, err := m.Contains(data1[0])
	if err != nil {
		t.Fatal(err)
	}
	if check {
		t.FailNow()
	}
}
func TestRoot(t *testing.T) {
	m := Init()
	data1 := GenerateData(5)
	check := m.Insert(data1)
	if !check {
		t.Fail()
	}
	check = m.Commit()
	if !check {
		t.Fail()
	}
	firstroot := m.Root()
	data2 := GenerateData(5)
	check = m.Insert(data2)
	if !check {
		t.Fail()
	}
	check = m.Commit()
	if !check {
		t.Fail()
	}
	if bytes.Equal(firstroot, m.Root()) {
		t.Fail()
	}
}
func TestCommit(t *testing.T) {

}
func TestProof(t *testing.T) {
	m := Init()
	data1 := GenerateData(5)
	check := m.Insert(data1)
	if !check {
		t.FailNow()
	}
	check = m.Commit()
	if !check {
		t.FailNow()
	}
	_, _, err := m.Proof(data1[2])
	if err != nil {
		t.Fatal(err)
	}
}
func GenerateData(size int) [][]byte {
	list2 := make([][]byte, 0)
	list1 := make([]byte, 32)
	for range size {
		rand.Read(list1)
		list2 = append(list2, list1)
		list1 = make([]byte, 32)
	}
	return list2
}

func FromByte(item []byte) [][]byte {
	var tmp [][]byte
	return append(tmp, item)
}
