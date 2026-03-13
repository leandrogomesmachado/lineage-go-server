package crypto

var StaticBlowfishKey = []byte{
	0x6b, 0x60, 0xcb, 0x5b,
	0x82, 0xce, 0x90, 0xb1,
	0xcc, 0x2b, 0x6c, 0x55,
	0x6c, 0x6c, 0x6c, 0x6c,
}

type LoginCrypt struct {
	staticCrypt *NewCrypt
	crypt       *NewCrypt
	isStatic    bool
}

func NewLoginCrypt(key []byte) (*LoginCrypt, error) {
	staticCrypt, err := NewNewCrypt(StaticBlowfishKey)
	if err != nil {
		return nil, err
	}

	crypt, err := NewNewCrypt(key)
	if err != nil {
		return nil, err
	}

	return &LoginCrypt{
		staticCrypt: staticCrypt,
		crypt:       crypt,
		isStatic:    true,
	}, nil
}

func (lc *LoginCrypt) Decrypt(data []byte, offset, size int) bool {
	if err := lc.crypt.Decrypt(data, offset, size); err != nil {
		return false
	}
	return VerifyChecksum(data, offset, size)
}

func (lc *LoginCrypt) Encrypt(data []byte, offset, size int) (int, error) {
	size += 4

	if lc.isStatic {
		size += 4
		size += 8 - size%8
		
		xorKey := GenerateRandomXORKey()
		EncXORPass(data, offset, size, xorKey)
		
		if err := lc.staticCrypt.Encrypt(data, offset, size); err != nil {
			return 0, err
		}
		
		lc.isStatic = false
	} else {
		size += 8 - size%8
		AppendChecksum(data, offset, size)
		
		if err := lc.crypt.Encrypt(data, offset, size); err != nil {
			return 0, err
		}
	}

	return size, nil
}
