package network

import (
	"bytes"
	"encoding/binary"
	"unicode/utf16"
)

type escritorPacket struct {
	buffer *bytes.Buffer
}

func novoEscritorPacket() *escritorPacket {
	return &escritorPacket{buffer: &bytes.Buffer{}}
}

func (e *escritorPacket) escreverC(valor byte) {
	e.buffer.WriteByte(valor)
}

func (e *escritorPacket) escreverH(valor uint16) {
	binary.Write(e.buffer, binary.LittleEndian, valor)
}

func (e *escritorPacket) escreverD(valor uint32) {
	binary.Write(e.buffer, binary.LittleEndian, valor)
}

func (e *escritorPacket) escreverQ(valor uint64) {
	binary.Write(e.buffer, binary.LittleEndian, valor)
}

func (e *escritorPacket) escreverF(valor float64) {
	binary.Write(e.buffer, binary.LittleEndian, valor)
}

func (e *escritorPacket) escreverB(valor []byte) {
	e.buffer.Write(valor)
}

func (e *escritorPacket) escreverS(valor string) {
	for _, runeValor := range utf16.Encode([]rune(valor)) {
		binary.Write(e.buffer, binary.LittleEndian, runeValor)
	}
	binary.Write(e.buffer, binary.LittleEndian, uint16(0))
}

func (e *escritorPacket) bytes() []byte {
	return e.buffer.Bytes()
}
