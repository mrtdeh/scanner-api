package engines

import (
	"crypto/sha256"
	"encoding/hex"
)

type HashGeneratorEngine interface {
	GenerateHash(input string) string
	GetSalt() string
}

type hashGeneratorEngine struct {
	salt string
}

func NewHashGeneratorEngine(salt string) HashGeneratorEngine {
	return &hashGeneratorEngine{salt: salt}
}

func (r *hashGeneratorEngine) GenerateHash(input string) string {
	hasher := sha256.New()
	hasher.Write([]byte(input + r.salt))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (r *hashGeneratorEngine) GetSalt() string {
	return r.salt
}
