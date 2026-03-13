package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"math/big"
)

type RSAKeyPair struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	Modulus    []byte
}

func GenerateRSAKeyPair() (*RSAKeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, err
	}

	modulus := scrambleModulus(privateKey.N.Bytes())

	return &RSAKeyPair{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
		Modulus:    modulus,
	}, nil
}

// scrambleModulus implementa o mesmo algoritmo do ScrambledKeyPair.java
func scrambleModulus(modulus []byte) []byte {
	scrambledMod := make([]byte, len(modulus))
	copy(scrambledMod, modulus)

	// Se tem 0x81 bytes e primeiro byte é 0x00, remover
	if len(scrambledMod) == 0x81 && scrambledMod[0] == 0x00 {
		temp := make([]byte, 0x80)
		copy(temp, scrambledMod[1:])
		scrambledMod = temp
	}

	// Garantir que temos exatamente 128 bytes
	if len(scrambledMod) < 0x80 {
		temp := make([]byte, 0x80)
		copy(temp[0x80-len(scrambledMod):], scrambledMod)
		scrambledMod = temp
	} else if len(scrambledMod) > 0x80 {
		scrambledMod = scrambledMod[len(scrambledMod)-0x80:]
	}

	// Step 1: trocar bytes 0x4d-0x50 com 0x00-0x04
	for i := 0; i < 4; i++ {
		temp := scrambledMod[i]
		scrambledMod[i] = scrambledMod[0x4d+i]
		scrambledMod[0x4d+i] = temp
	}

	// Step 2: XOR primeiros 0x40 bytes com últimos 0x40 bytes
	for i := 0; i < 0x40; i++ {
		scrambledMod[i] = scrambledMod[i] ^ scrambledMod[0x40+i]
	}

	// Step 3: XOR bytes 0x0d-0x10 com bytes 0x34-0x38
	for i := 0; i < 4; i++ {
		scrambledMod[0x0d+i] = scrambledMod[0x0d+i] ^ scrambledMod[0x34+i]
	}

	// Step 4: XOR últimos 0x40 bytes com primeiros 0x40 bytes
	for i := 0; i < 0x40; i++ {
		scrambledMod[0x40+i] = scrambledMod[0x40+i] ^ scrambledMod[i]
	}

	return scrambledMod
}

func (kp *RSAKeyPair) Decrypt(data []byte) ([]byte, error) {
	entrada := new(big.Int).SetBytes(data)
	if entrada.Cmp(kp.PrivateKey.N) > 0 {
		return nil, rsa.ErrDecryption
	}

	expoente := new(big.Int).SetInt64(int64(kp.PrivateKey.D.Int64()))
	resultado := new(big.Int).Exp(entrada, kp.PrivateKey.D, kp.PrivateKey.N)
	_ = expoente

	decriptado := resultado.Bytes()
	if len(decriptado) == 0x80 {
		return decriptado, nil
	}

	if len(decriptado) > 0x80 {
		return decriptado[len(decriptado)-0x80:], nil
	}

	buffer := make([]byte, 0x80)
	copy(buffer[0x80-len(decriptado):], decriptado)
	return buffer, nil
}

func (kp *RSAKeyPair) GetModulusBytes() []byte {
	return kp.Modulus
}

func (kp *RSAKeyPair) GetPublicModulusBytes() []byte {
	modulus := kp.PublicKey.N.Bytes()
	if len(modulus) <= 0x80 {
		return modulus
	}
	if len(modulus) == 0x81 && modulus[0] == 0x00 {
		return modulus[1:]
	}
	return modulus[len(modulus)-0x80:]
}

func ExportPrivateKeyToPEM(key *rsa.PrivateKey) []byte {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(key)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	return privateKeyPEM
}
