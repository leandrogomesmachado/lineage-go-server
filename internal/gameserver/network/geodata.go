package network

import "math"

const (
	mundoXMin            = -131072
	mundoXMax            = 229375
	mundoYMin            = -262144
	mundoYMax            = 294911
	mundoZMax            = 16410
	tamanhoRegiao        = 2048
	raioVisibilidade     = 2500
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

func getValidLocationCurta(origemX int32, origemY int32, origemZ int32, destinoX int32, destinoY int32, destinoZ int32) (int32, int32, int32) {
	origemX, origemY, origemZ = normalizarPosicaoMundo(origemX, origemY, origemZ)
	destinoX, destinoY, destinoZ = normalizarPosicaoMundo(destinoX, destinoY, destinoZ)
	if destinoX == origemX && destinoY == origemY {
		return destinoX, destinoY, getHeight(destinoX, destinoY, destinoZ)
	}
	if !geodataDisponivel() {
		return destinoX, destinoY, getHeight(destinoX, destinoY, destinoZ)
	}
	geoOrigemX := getGeoX(origemX)
	geoOrigemY := getGeoY(origemY)
	blocoOrigem := obterBlocoGeodata(geoOrigemX, geoOrigemY)
	if blocoOrigem == nil || !blocoOrigem.hasGeoPos() {
		return destinoX, destinoY, getHeight(destinoX, destinoY, destinoZ)
	}
	geoOrigemZ := blocoOrigem.getHeightNearest(geoOrigemX, geoOrigemY, origemZ)
	nsweAtual := blocoOrigem.getNsweNearest(geoOrigemX, geoOrigemY, geoOrigemZ)
	geoDestinoX := getGeoX(destinoX)
	geoDestinoY := getGeoY(destinoY)
	geoDestinoZ := getHeight(destinoX, destinoY, destinoZ)
	gridX := origemX &^ int32(0x0F)
	gridY := origemY &^ int32(0x0F)
	passoGeoX := geoOrigemX
	passoGeoY := geoOrigemY
	deltaX := destinoX - origemX
	deltaY := destinoY - origemY
	declive := 0.0
	if deltaX != 0 {
		declive = float64(deltaY) / float64(deltaX)
	}
	direcaoX, direcaoY, passoGridX, passoGridY, offsetBordaX, offsetBordaY := resolverDirecaoGeo(geoDestinoX-passoGeoX, geoDestinoY-passoGeoY)
	for passoGeoX != geoDestinoX || passoGeoY != geoDestinoY {
		checkX := gridX + offsetBordaX
		checkY := origemY
		if deltaX != 0 {
			checkY = origemY + int32(math.Round(declive*float64(checkX-origemX)))
		}
		dirMovimento := byte(0)
		if passoGridX != 0 && getGeoY(checkY) == passoGeoY {
			gridX += passoGridX
			passoGeoX += direcaoX
			dirMovimento = resolverNsweDirecao(direcaoX, 0)
		}
		if dirMovimento == 0 {
			checkY = gridY + offsetBordaY
			checkX = origemX
			if deltaY != 0 {
				checkX = origemX + int32(math.Round(float64(checkY-origemY)/float64(deltaY)*float64(deltaX)))
			}
			checkX = clampInt32(checkX, gridX, gridX+15)
			gridY += passoGridY
			passoGeoY += direcaoY
			dirMovimento = resolverNsweDirecao(0, direcaoY)
		}
		if passoGeoX < 0 || passoGeoY < 0 {
			return checkX, checkY, geoOrigemZ
		}
		if (nsweAtual & dirMovimento) == 0 {
			return checkX, checkY, geoOrigemZ
		}
		blocoSeguinte := obterBlocoGeodata(passoGeoX, passoGeoY)
		if blocoSeguinte == nil || !blocoSeguinte.hasGeoPos() {
			return checkX, checkY, geoOrigemZ
		}
		indiceAbaixo := blocoSeguinte.getIndexBelow(passoGeoX, passoGeoY, geoOrigemZ+geoCellIgnoreHeight)
		if indiceAbaixo < 0 {
			return checkX, checkY, geoOrigemZ
		}
		geoOrigemZ = blocoSeguinte.getHeightPorIndice(passoGeoX, passoGeoY, indiceAbaixo)
		nsweAtual = blocoSeguinte.getNswePorIndice(passoGeoX, passoGeoY, indiceAbaixo)
	}
	if geoOrigemZ != geoDestinoZ {
		return origemX, origemY, origemZ
	}
	return destinoX, destinoY, geoDestinoZ
}

func corrigirPosicaoPorGeodataInicial(origemX int32, origemY int32, origemZ int32, destinoX int32, destinoY int32, destinoZ int32) (int32, int32, int32) {
	return getValidLocationCurta(origemX, origemY, origemZ, destinoX, destinoY, destinoZ)
}

func resolverDirecaoGeo(deltaGeoX int32, deltaGeoY int32) (int32, int32, int32, int32, int32, int32) {
	direcaoX := int32(0)
	direcaoY := int32(0)
	passoGridX := int32(0)
	passoGridY := int32(0)
	offsetBordaX := int32(0)
	offsetBordaY := int32(0)
	if deltaGeoX > 0 {
		direcaoX = 1
		passoGridX = 16
		offsetBordaX = 16
	}
	if deltaGeoX < 0 {
		direcaoX = -1
		passoGridX = -16
	}
	if deltaGeoY > 0 {
		direcaoY = 1
		passoGridY = 16
		offsetBordaY = 16
	}
	if deltaGeoY < 0 {
		direcaoY = -1
		passoGridY = -16
	}
	return direcaoX, direcaoY, passoGridX, passoGridY, offsetBordaX, offsetBordaY
}

func resolverNsweDirecao(deltaGeoX int32, deltaGeoY int32) byte {
	if deltaGeoX > 0 {
		return geoCellFlagE
	}
	if deltaGeoX < 0 {
		return geoCellFlagW
	}
	if deltaGeoY > 0 {
		return geoCellFlagS
	}
	if deltaGeoY < 0 {
		return geoCellFlagN
	}
	return geoCellFlagAll
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

func posicaoNpcNoRaioVisivel(origem *playerAtivo, alvo *npcGlobalRuntime) bool {
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
