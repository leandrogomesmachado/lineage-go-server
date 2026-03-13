package crypto

import (
	"crypto/rand"
	"encoding/binary"
	"math/big"

	"golang.org/x/crypto/blowfish"
)

type NewCrypt struct {
	encryptCipher *blowfish.Cipher
	decryptCipher *blowfish.Cipher
}

func NewNewCrypt(key []byte) (*NewCrypt, error) {
	encryptCipher, err := blowfish.NewCipher(key)
	if err != nil {
		return nil, err
	}

	decryptCipher, err := blowfish.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return &NewCrypt{
		encryptCipher: encryptCipher,
		decryptCipher: decryptCipher,
	}, nil
}

func converterBlocoL2ParaPadrao(origem []byte) [8]byte {
	var convertido [8]byte
	esquerda := binary.LittleEndian.Uint32(origem[:4])
	direita := binary.LittleEndian.Uint32(origem[4:8])
	binary.BigEndian.PutUint32(convertido[:4], esquerda)
	binary.BigEndian.PutUint32(convertido[4:8], direita)
	return convertido
}

func converterBlocoPadraoParaL2(origem []byte) [8]byte {
	var convertido [8]byte
	esquerda := binary.BigEndian.Uint32(origem[:4])
	direita := binary.BigEndian.Uint32(origem[4:8])
	binary.LittleEndian.PutUint32(convertido[:4], esquerda)
	binary.LittleEndian.PutUint32(convertido[4:8], direita)
	return convertido
}

func (nc *NewCrypt) Decrypt(data []byte, offset, size int) error {
	result := make([]byte, size)
	count := size / 8

	for i := 0; i < count; i++ {
		blocoEntrada := converterBlocoL2ParaPadrao(data[offset+i*8 : offset+i*8+8])
		var blocoSaida [8]byte
		nc.decryptCipher.Decrypt(blocoSaida[:], blocoEntrada[:])
		blocoL2 := converterBlocoPadraoParaL2(blocoSaida[:])
		copy(result[i*8:i*8+8], blocoL2[:])
	}

	copy(data[offset:offset+size], result)
	return nil
}

func (nc *NewCrypt) Encrypt(data []byte, offset, size int) error {
	result := make([]byte, size)
	count := size / 8

	for i := 0; i < count; i++ {
		blocoEntrada := converterBlocoL2ParaPadrao(data[offset+i*8 : offset+i*8+8])
		var blocoSaida [8]byte
		nc.encryptCipher.Encrypt(blocoSaida[:], blocoEntrada[:])
		blocoL2 := converterBlocoPadraoParaL2(blocoSaida[:])
		copy(result[i*8:i*8+8], blocoL2[:])
	}

	copy(data[offset:offset+size], result)
	return nil
}

func VerifyChecksum(data []byte, offset, size int) bool {
	if (size&3) != 0 || size <= 4 {
		return false
	}

	var chksum uint32 = 0
	count := size - 4

	for i := offset; i < offset+count; i += 4 {
		check := uint32(data[i]) |
			uint32(data[i+1])<<8 |
			uint32(data[i+2])<<16 |
			uint32(data[i+3])<<24
		chksum ^= check
	}

	i := offset + count
	check := uint32(data[i]) |
		uint32(data[i+1])<<8 |
		uint32(data[i+2])<<16 |
		uint32(data[i+3])<<24

	return check == chksum
}

func AppendChecksum(data []byte, offset, size int) {
	var chksum uint32 = 0
	count := size - 4

	for i := offset; i < offset+count; i += 4 {
		ecx := uint32(data[i]) |
			uint32(data[i+1])<<8 |
			uint32(data[i+2])<<16 |
			uint32(data[i+3])<<24
		chksum ^= ecx
	}

	i := offset + count
	data[i] = byte(chksum & 0xff)
	data[i+1] = byte((chksum >> 8) & 0xff)
	data[i+2] = byte((chksum >> 16) & 0xff)
	data[i+3] = byte((chksum >> 24) & 0xff)
}

func EncXORPass(data []byte, offset, size int, key uint32) {
	stop := size - 8
	pos := 4 + offset
	ecx := key

	for pos < stop {
		edx := uint32(data[pos]) |
			uint32(data[pos+1])<<8 |
			uint32(data[pos+2])<<16 |
			uint32(data[pos+3])<<24

		ecx += edx
		edx ^= ecx

		data[pos] = byte(edx & 0xff)
		data[pos+1] = byte((edx >> 8) & 0xff)
		data[pos+2] = byte((edx >> 16) & 0xff)
		data[pos+3] = byte((edx >> 24) & 0xff)
		pos += 4
	}

	data[pos] = byte(ecx & 0xff)
	data[pos+1] = byte((ecx >> 8) & 0xff)
	data[pos+2] = byte((ecx >> 16) & 0xff)
	data[pos+3] = byte((ecx >> 24) & 0xff)
}

func GenerateRandomXORKey() uint32 {
	n, _ := rand.Int(rand.Reader, big.NewInt(0x7FFFFFFF))
	return uint32(n.Int64())
}
