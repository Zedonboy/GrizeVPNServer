package signer

import (
	"crypto"
	"crypto/hmac"
)

type Signer struct {
	key []byte
}

func NewSigner(key string) *Signer {
	return &Signer{
		key: []byte(key),
	}
}

func (s *Signer) Sign(mssg []byte) ([]byte, error) {
	mac := hmac.New(crypto.SHA224.New, s.key)
	_, err := mac.Write(mssg)
	if err != nil {
		return nil, err
	}
	sig := mac.Sum(nil)
	return sig, nil

}
func (s *Signer) Verify(message []byte, inputHash []byte) bool {
	mac := hmac.New(crypto.SHA256.New, s.key)
	_, err := mac.Write(message)
	if err != nil {
		return false
	}
	expectedHash := mac.Sum(nil)
	return hmac.Equal(expectedHash, inputHash)
}
