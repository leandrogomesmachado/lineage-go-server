package network

const (
	mundoXMin     = -131072
	mundoXMax     = 229375
	mundoYMin     = -262144
	mundoYMax     = 294911
	mundoZMax     = 16410
	tamanhoRegiao = 2048
	raioVisibilidade = 2500
	desyncMaximoValidate = 300
)

func clampInt32(valor int32, minimo int32, maximo int32) int32 {
	if valor < minimo {
		return minimo
	}
	if valor > maximo {
		return maximo
	}
	return valor
}

func calcularRegiaoX(x int32) int32 {
	return (clampInt32(x, mundoXMin, mundoXMax) - mundoXMin) / tamanhoRegiao
}

func calcularRegiaoY(y int32) int32 {
	return (clampInt32(y, mundoYMin, mundoYMax) - mundoYMin) / tamanhoRegiao
}

func normalizarPosicaoMundo(x int32, y int32, z int32) (int32, int32, int32) {
	xAjustado := clampInt32(x, mundoXMin, mundoXMax)
	yAjustado := clampInt32(y, mundoYMin, mundoYMax)
	zAjustado := clampInt32(z, -20000, mundoZMax)
	return xAjustado, yAjustado, zAjustado
}

func corrigirPosicaoPorGeodataInicial(origemX int32, origemY int32, origemZ int32, destinoX int32, destinoY int32, destinoZ int32) (int32, int32, int32) {
	destinoX, destinoY, destinoZ = normalizarPosicaoMundo(destinoX, destinoY, destinoZ)
	if distancia3D(origemX, origemY, origemZ, destinoX, destinoY, destinoZ) <= 1200 {
		return destinoX, destinoY, destinoZ
	}
	return origemX, origemY, origemZ
}

func posicaoNoRaioVisivel(origem *playerAtivo, alvo *playerAtivo) bool {
	if origem == nil {
		return false
	}
	if alvo == nil {
		return false
	}
	if diferencaAbsolutaInt32(origem.regiaoX, alvo.regiaoX) > 1 {
		return false
	}
	if diferencaAbsolutaInt32(origem.regiaoY, alvo.regiaoY) > 1 {
		return false
	}
	if distancia3D(origem.x, origem.y, origem.z, alvo.x, alvo.y, alvo.z) > raioVisibilidade {
		return false
	}
	return true
}

func diferencaAbsolutaInt32(a int32, b int32) int32 {
	if a >= b {
		return a - b
	}
	return b - a
}
