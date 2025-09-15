package crypto

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"testing"
)

func TestSha256(t *testing.T) {
	Crypt := NewHasher("sha256")
	data := "hello"
	if Crypt.Sum([]byte(data)) == nil {
		t.Fail()
	}
	data2 := "hi"
	tes1 := Crypt.Sum([]byte(data2))
	tes2 := sha256.Sum256([]byte(data2))
	if !bytes.Equal(tes1, tes2[:]) {
		t.Fail()
	}

}
func TestMd5(t *testing.T) {
	Crypt := NewHasher("md5")
	data := "hello" // change to Random Generator
	if Crypt.Sum([]byte(data)) == nil {
		t.Fail()
	}
	data2 := "hi" // change to Random Generator
	tes1 := Crypt.Sum([]byte(data2))
	tes2 := md5.Sum([]byte(data2))
	if !bytes.Equal(tes1, tes2[:]) {
		t.Fail()
	}

}
