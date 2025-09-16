package encoder

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/rlp"
)

type Encoder interface {
	Encode(v interface{}) ([]byte, error)
	Decode(data []byte, v interface{}) error
	Name() string
}

func NewEncoder(name string) Encoder {
	switch name {
	case "rlp":
		return &RLPEncoder{}
	case "json":
		return &JsonEncoder{}
	default:
		return &JsonEncoder{}
	}
}

type JsonEncoder struct{}

func (e JsonEncoder) Encode(v interface{}) ([]byte, error)    { return json.Marshal(v) }
func (e JsonEncoder) Decode(data []byte, v interface{}) error { return json.Unmarshal(data, v) }
func (e JsonEncoder) Name() string                            { return "json" }


type RLPEncoder struct{}


func (e RLPEncoder) Encode(v interface{}) ([]byte, error)    { return rlp.EncodeToBytes(v) }
func (e RLPEncoder) Decode(data []byte, v interface{}) error { return rlp.DecodeBytes(data, v) }
func (e RLPEncoder) Name() string                            { return "rlp" }

