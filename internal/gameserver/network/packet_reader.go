package network

import (
	"encoding/binary"
	"unicode/utf16"
)

type leitorPacket struct {
	dados  []byte
	offset int
}

func novoLeitorPacket(dados []byte) *leitorPacket {
	return &leitorPacket{dados: dados, offset: 1}
}

func (l *leitorPacket) lerC() byte {
	if l.offset >= len(l.dados) {
		return 0
	}
	valor := l.dados[l.offset]
	l.offset++
	return valor
}

func (l *leitorPacket) lerH() uint16 {
	if l.offset+2 > len(l.dados) {
		return 0
	}
	valor := binary.LittleEndian.Uint16(l.dados[l.offset : l.offset+2])
	l.offset += 2
	return valor
}

func (l *leitorPacket) lerD() uint32 {
	if l.offset+4 > len(l.dados) {
		return 0
	}
	valor := binary.LittleEndian.Uint32(l.dados[l.offset : l.offset+4])
	l.offset += 4
	return valor
}

func (l *leitorPacket) lerB(tamanho int) []byte {
	if tamanho <= 0 {
		return []byte{}
	}
	if l.offset+tamanho > len(l.dados) {
		tamanho = len(l.dados) - l.offset
	}
	valor := make([]byte, tamanho)
	copy(valor, l.dados[l.offset:l.offset+tamanho])
	l.offset += tamanho
	return valor
}

func (l *leitorPacket) lerS() string {
	codigo := make([]uint16, 0)
	for l.offset+1 < len(l.dados) {
		valor := binary.LittleEndian.Uint16(l.dados[l.offset : l.offset+2])
		l.offset += 2
		if valor == 0 {
			break
		}
		codigo = append(codigo, valor)
	}
	return string(utf16.Decode(codigo))
}
