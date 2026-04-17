package service

import (
	"crypto/rand"
	"math/big"
)

const (
	alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type Generator interface {
	Generate() (string, error)
}

type RandomGenerator struct {
	length int
}

func NewRandomGenerator(length int) *RandomGenerator {
	return &RandomGenerator{length: length}
}

func (r *RandomGenerator) Generate() (string, error) {
	short := make([]byte, r.length)
	alphabetLen := big.NewInt(int64(len(alphabet)))

	for i := 0; i < r.length; i++ {
		n, err := rand.Int(rand.Reader, alphabetLen)
		if err != nil {
			return "", err
		}
		short[i] = alphabet[n.Int64()]
	}

	return string(short), nil
}
