package crypto

import (
	"crypto/cipher"
	"golang.org/x/crypto/blowfish"
)

type BlowfishCipher struct {
	cipher cipher.Block
}

func NewBlowfishCipher(key []byte) (*BlowfishCipher, error) {
	c, err := blowfish.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &BlowfishCipher{cipher: c}, nil
}

func (b *BlowfishCipher) Encrypt(data []byte) []byte {
	encrypted := make([]byte, len(data))
	for i := 0; i < len(data); i += 8 {
		b.cipher.Encrypt(encrypted[i:i+8], data[i:i+8])
	}
	return encrypted
}

func (b *BlowfishCipher) Decrypt(data []byte) []byte {
	decrypted := make([]byte, len(data))
	for i := 0; i < len(data); i += 8 {
		b.cipher.Decrypt(decrypted[i:i+8], data[i:i+8])
	}
	return decrypted
}
