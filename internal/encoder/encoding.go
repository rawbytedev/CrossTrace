package encoder

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/rlp"
	"gopkg.in/yaml.v3"
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
	case "yaml":
		return &YamlEncoder{}
	case "json":
		return &JsonEncoder{}

	default:
		return &RLPEncoder{}
	}
}

type YamlEncoder struct{}

func (e YamlEncoder) Encode(v interface{}) ([]byte, error)    { return yaml.Marshal(v) }
func (e YamlEncoder) Decode(data []byte, v interface{}) error { return yaml.Unmarshal(data, v) }
func (e YamlEncoder) Name() string                            { return "yaml" }

type JsonEncoder struct{}

func (e JsonEncoder) Encode(v interface{}) ([]byte, error)    { return json.Marshal(v) }
func (e JsonEncoder) Decode(data []byte, v interface{}) error { return json.Unmarshal(data, v) }
func (e JsonEncoder) Name() string                            { return "json" }

type RLPEncoder struct{}

func (e RLPEncoder) Encode(v interface{}) ([]byte, error)    { return rlp.EncodeToBytes(v) }
func (e RLPEncoder) Decode(data []byte, v interface{}) error { return rlp.DecodeBytes(data, v) }
func (e RLPEncoder) Name() string                            { return "rlp" }
