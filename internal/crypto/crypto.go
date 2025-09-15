package crypto

import (
	"crypto/md5"
	"crypto/sha256"
)

type Hasher interface {
	Sum(data []byte) []byte
	Name() string
}

func NewHasher(name string) Hasher {
	switch name {
	case "sha256":
		return &SHA256Hasher{}
	case "md5":
		return &MD5Hasher{}
	default:
		return nil
	}
}

type SHA256Hasher struct{}

func (h SHA256Hasher) Sum(data []byte) []byte { sum := sha256.Sum256(data); return sum[:] }
func (h SHA256Hasher) Name() string             { return "sha256" }

type MD5Hasher struct{}

func (h MD5Hasher) Sum(data []byte) []byte { sum := md5.Sum(data); return sum[:] }
func (h MD5Hasher) Name() string           { return "md5" }
