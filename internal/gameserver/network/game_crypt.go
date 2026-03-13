package network

type gameCrypt struct {
	inKey     [16]byte
	outKey    [16]byte
	habilitado bool
}

func novoGameCrypt() *gameCrypt {
	return &gameCrypt{}
}

func (g *gameCrypt) setKey(chave []byte) {
	copy(g.inKey[:], chave)
	copy(g.outKey[:], chave)
}

func (g *gameCrypt) decrypt(raw []byte, offset, size int) {
	if !g.habilitado {
		return
	}
	temp := 0
	for i := 0; i < size; i++ {
		temp2 := int(raw[offset+i] & 0xFF)
		raw[offset+i] = byte(temp2 ^ int(g.inKey[i&15]) ^ temp)
		temp = temp2
	}
	old := int(g.inKey[8]) & 0xff
	old |= int(g.inKey[9])<<8 & 0xff00
	old |= int(g.inKey[10])<<16 & 0xff0000
	old |= int(g.inKey[11])<<24 & 0xff000000
	old += size
	g.inKey[8] = byte(old & 0xff)
	g.inKey[9] = byte(old >> 8 & 0xff)
	g.inKey[10] = byte(old >> 16 & 0xff)
	g.inKey[11] = byte(old >> 24 & 0xff)
}

func (g *gameCrypt) encrypt(raw []byte, offset, size int) {
	if !g.habilitado {
		g.habilitado = true
		return
	}
	temp := 0
	for i := 0; i < size; i++ {
		temp2 := int(raw[offset+i] & 0xFF)
		temp = temp2 ^ int(g.outKey[i&15]) ^ temp
		raw[offset+i] = byte(temp)
	}
	old := int(g.outKey[8]) & 0xff
	old |= int(g.outKey[9])<<8 & 0xff00
	old |= int(g.outKey[10])<<16 & 0xff0000
	old |= int(g.outKey[11])<<24 & 0xff000000
	old += size
	g.outKey[8] = byte(old & 0xff)
	g.outKey[9] = byte(old >> 8 & 0xff)
	g.outKey[10] = byte(old >> 16 & 0xff)
	g.outKey[11] = byte(old >> 24 & 0xff)
}
