package encoder

import (
	"encoding/json"
)

type Encoder interface {
	Encode(v interface{}) ([]byte, error)
	Decode(data []byte, v interface{}) error
	Name() string
}

func NewEncoder(name string) Encoder {
	switch name {
	/*case "rlp":
		return &RLPEncoder{}
	case "yaml":
		return &YAMLEncoder{}*/
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

type YAMLEncoder struct{}

/*
func (e YAMLEncoder) Encode(v interface{}) ([]byte, error)    { return yaml.EncodeToBytes(v) }
func (e YAMLEncoder) Decode(data []byte, v interface{}) error { return yaml.DecodeBytes(data, v) }
func (e YAMLEncoder) Name() string                            { return "yaml" }
*/
type RLPEncoder struct{}

/*
func (e RLPEncoder) Encode(v interface{}) ([]byte, error)    { return rlp.EncodeToBytes(v) }
func (e RLPEncoder) Decode(data []byte, v interface{}) error { return rlp.DecodeBytes(data, v) }
func (e RLPEncoder) Name() string                            { return "rlp" }
*/
