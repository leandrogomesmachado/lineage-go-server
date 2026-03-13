package network

import (
	"bytes"
	"encoding/binary"
	"unicode/utf16"
)

type escritorPacketServidor struct {
	buffer bytes.Buffer
}

func novoEscritorPacketServidor() *escritorPacketServidor {
	return &escritorPacketServidor{}
}

func (e *escritorPacketServidor) escreverC(valor byte) {
	e.buffer.WriteByte(valor)
}

func (e *escritorPacketServidor) escreverH(valor uint16) {
	bytesValor := make([]byte, 2)
	binary.LittleEndian.PutUint16(bytesValor, valor)
	e.buffer.Write(bytesValor)
}

func (e *escritorPacketServidor) escreverD(valor uint32) {
	bytesValor := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytesValor, valor)
	e.buffer.Write(bytesValor)
}

func (e *escritorPacketServidor) escreverB(valor []byte) {
	e.buffer.Write(valor)
}

func (e *escritorPacketServidor) escreverS(valor string) {
	if valor == "" {
		e.buffer.WriteByte(0)
		e.buffer.WriteByte(0)
		return
	}

	for _, unidade := range utf16.Encode([]rune(valor)) {
		bytesValor := make([]byte, 2)
		binary.LittleEndian.PutUint16(bytesValor, unidade)
		e.buffer.Write(bytesValor)
	}
		e.buffer.WriteByte(0)
		e.buffer.WriteByte(0)
}

func (e *escritorPacketServidor) bytes() []byte {
	resultado := make([]byte, e.buffer.Len())
	copy(resultado, e.buffer.Bytes())
	return resultado
}
