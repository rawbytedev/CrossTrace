package encoder

import (
	"fmt"
	"testing"
)

type entry struct {
	Name string
	Data string
}

func NewEntry() *entry {
	return &entry{Name: "hello", Data: "world"}
}
func TestRLPEncode(t *testing.T) {

	test1 := NewEntry()
	enc := NewEncoder("json")
	encoded, err := enc.Encode(*test1)
	fmt.Printf("JSON format: %s", encoded)
	if err != nil {
		t.Fatal(err)
	}
	var test2 entry
	err = enc.Decode(encoded, &test2)
	if err != nil {
		t.Fatal(err)
	}
	if *test1 != test2 {
		t.Fail()
	}
}
