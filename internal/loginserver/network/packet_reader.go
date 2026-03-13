package network

import (
	"encoding/binary"
	"unicode/utf16"
)

type leitorPacketCliente struct {
	dados  []byte
	offset int
}

func novoLeitorPacketCliente(dados []byte) *leitorPacketCliente {
	return &leitorPacketCliente{
		dados:  dados,
		offset: 1,
	}
}

func (l *leitorPacketCliente) lerC() byte {
	valor := l.dados[l.offset]
	l.offset++
	return valor
}

func (l *leitorPacketCliente) lerH() uint16 {
	valor := binary.LittleEndian.Uint16(l.dados[l.offset : l.offset+2])
	l.offset += 2
	return valor
}

func (l *leitorPacketCliente) lerD() uint32 {
	valor := binary.LittleEndian.Uint32(l.dados[l.offset : l.offset+4])
	l.offset += 4
	return valor
}

func (l *leitorPacketCliente) lerB(tamanho int) []byte {
	resultado := make([]byte, tamanho)
	copy(resultado, l.dados[l.offset:l.offset+tamanho])
	l.offset += tamanho
	return resultado
}

func (l *leitorPacketCliente) lerS() string {
	inicio := l.offset
	fim := inicio
	for fim < len(l.dados)-1 {
		if l.dados[fim] == 0 && l.dados[fim+1] == 0 {
			break
		}
		fim += 2
	}

	bytesTexto := l.dados[inicio:fim]
	l.offset = fim + 2
	if len(bytesTexto) == 0 {
		return ""
	}

	unidades := make([]uint16, 0, len(bytesTexto)/2)
	for i := 0; i+1 < len(bytesTexto); i += 2 {
		unidades = append(unidades, binary.LittleEndian.Uint16(bytesTexto[i:i+2]))
	}
	return string(utf16.Decode(unidades))
}
