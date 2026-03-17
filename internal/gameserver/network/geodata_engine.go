package network

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

const (
	geoCellFlagNone        byte  = 0x00
	geoCellFlagE           byte  = 0x01
	geoCellFlagW           byte  = 0x02
	geoCellFlagS           byte  = 0x04
	geoCellFlagN           byte  = 0x08
	geoCellFlagAll         byte  = 0x0F
	geoCellSize            int32 = 16
	geoCellHeight          int32 = 8
	geoCellIgnoreHeight    int32 = geoCellHeight * 6
	geoTipoFlatL2jL2off    byte  = 0
	geoTipoComplexL2j      byte  = 1
	geoTipoComplexL2off    byte  = 0x40
	geoTipoMultilayerL2j   byte  = 2
	geoBlockCellsX         int32 = 8
	geoBlockCellsY         int32 = 8
	geoBlockCellsTotal     int32 = geoBlockCellsX * geoBlockCellsY
	geoRegionBlocksX       int32 = 256
	geoRegionBlocksY       int32 = 256
	geoRegionCellsX        int32 = geoRegionBlocksX * geoBlockCellsX
	geoRegionCellsY        int32 = geoRegionBlocksY * geoBlockCellsY
	geoTileSize            int32 = 32768
	geoTileXMin            int32 = 16
	geoTileXMax            int32 = 26
	geoTileYMin            int32 = 10
	geoTileYMax            int32 = 25
	geoCabecalhoL2offBytes int32 = 18
)

type geodataRuntime struct {
	mu                    sync.RWMutex
	datapackPath          string
	caminhoGeodata        string
	arquivosRegiao        []string
	cacheRegiao           map[string]*geodataRegiao
	disponivel            bool
	avisoSemArquivosFeito bool
}

var geodataAtual = &geodataRuntime{}

type geodataBloco interface {
	hasGeoPos() bool
	getHeightNearest(geoX int32, geoY int32, worldZ int32) int32
	getNsweNearest(geoX int32, geoY int32, worldZ int32) byte
	getIndexBelow(geoX int32, geoY int32, worldZ int32) int32
	getHeightPorIndice(geoX int32, geoY int32, indice int32) int32
	getNswePorIndice(geoX int32, geoY int32, indice int32) byte
}

type geodataRegiao struct {
	rx     int32
	ry     int32
	blocos [][]geodataBloco
}

type geodataBlocoNulo struct{}

type geodataBlocoFlat struct {
	altura int16
	nswe   byte
}

type geodataCelulaComplexa struct {
	altura int16
	nswe   byte
}

type geodataBlocoComplexo struct {
	celulas [geoBlockCellsTotal]geodataCelulaComplexa
}

type geodataCamada struct {
	altura int16
	nswe   byte
}

type geodataBlocoMulticamadas struct {
	celulas [geoBlockCellsTotal][]geodataCamada
}

func carregarGeodata(datapackPath string) {
	if geodataAtual == nil {
		return
	}
	geodataAtual.mu.Lock()
	defer geodataAtual.mu.Unlock()
	geodataAtual.datapackPath = strings.TrimSpace(datapackPath)
	geodataAtual.caminhoGeodata = ""
	geodataAtual.arquivosRegiao = nil
	geodataAtual.cacheRegiao = map[string]*geodataRegiao{}
	geodataAtual.disponivel = false
	geodataAtual.avisoSemArquivosFeito = false
	if geodataAtual.datapackPath == "" {
		logger.Warnf("Geodata nao configurada porque o datapack nao foi informado")
		return
	}
	caminhoGeodata := filepath.Join(geodataAtual.datapackPath, "data", "geodata")
	geodataAtual.caminhoGeodata = caminhoGeodata
	arquivos, err := listarArquivosBinariosGeodata(caminhoGeodata)
	if err != nil {
		logger.Warnf("Geodata indisponivel caminho=%s erro=%v", caminhoGeodata, err)
		return
	}
	if len(arquivos) <= 0 {
		logger.Warnf("Geodata sem arquivos reais caminho=%s", caminhoGeodata)
		return
	}
	geodataAtual.arquivosRegiao = arquivos
	geodataAtual.disponivel = true
	logger.Infof("Geodata detectada caminho=%s arquivos=%d", caminhoGeodata, len(arquivos))
}

func listarArquivosBinariosGeodata(caminhoGeodata string) ([]string, error) {
	entradas, err := os.ReadDir(caminhoGeodata)
	if err != nil {
		return nil, err
	}
	arquivos := make([]string, 0, len(entradas))
	for _, entrada := range entradas {
		if entrada == nil {
			continue
		}
		if entrada.IsDir() {
			continue
		}
		nome := strings.ToLower(strings.TrimSpace(entrada.Name()))
		if nome == "" {
			continue
		}
		if strings.HasSuffix(nome, ".jpg") {
			continue
		}
		if strings.HasSuffix(nome, ".jpeg") {
			continue
		}
		if strings.HasSuffix(nome, ".png") {
			continue
		}
		if strings.HasSuffix(nome, ".gif") {
			continue
		}
		if strings.HasSuffix(nome, ".txt") {
			continue
		}
		if strings.HasSuffix(nome, ".md") {
			continue
		}
		caminhoArquivo := filepath.Join(caminhoGeodata, entrada.Name())
		arquivos = append(arquivos, caminhoArquivo)
	}
	return arquivos, nil
}

func geodataDisponivel() bool {
	if geodataAtual == nil {
		return false
	}
	geodataAtual.mu.RLock()
	defer geodataAtual.mu.RUnlock()
	return geodataAtual.disponivel
}

func getGeoX(worldX int32) int32 {
	return (worldX - mundoXMin) / geoCellSize
}

func getGeoY(worldY int32) int32 {
	return (worldY - mundoYMin) / geoCellSize
}

func getWorldX(geoX int32) int32 {
	return geoX*geoCellSize + mundoXMin + (geoCellSize / 2)
}

func getWorldY(geoY int32) int32 {
	return geoY*geoCellSize + mundoYMin + (geoCellSize / 2)
}

func getHeight(worldX int32, worldY int32, worldZ int32) int32 {
	if !geodataDisponivel() {
		logarAvisoFallbackGeodataUmaVez()
		return worldZ
	}
	geoX := getGeoX(worldX)
	geoY := getGeoY(worldY)
	bloco := obterBlocoGeodata(geoX, geoY)
	if bloco == nil {
		return worldZ
	}
	if !bloco.hasGeoPos() {
		return worldZ
	}
	return bloco.getHeightNearest(geoX, geoY, worldZ)
}

func logarAvisoFallbackGeodataUmaVez() {
	if geodataAtual == nil {
		return
	}
	geodataAtual.mu.Lock()
	defer geodataAtual.mu.Unlock()
	if geodataAtual.avisoSemArquivosFeito {
		return
	}
	geodataAtual.avisoSemArquivosFeito = true
	logger.Warnf("Geodata real ainda nao carregada, getHeight esta em fallback seguro")
}

func resolverZComGeodataOuTerritorio(x int32, y int32, zBase int32, territorio territorioSpawn) int32 {
	zReferencia := zBase
	if !zDentroDaFaixaDoTerritorio(territorio, zReferencia) {
		zReferencia = resolverZReferenciaTerritorio(territorio)
	}
	zGeodata := getHeight(x, y, zReferencia)
	if zDentroDaFaixaDoTerritorio(territorio, zGeodata) {
		return zGeodata
	}
	if zDentroDaFaixaDoTerritorio(territorio, zBase) {
		return zBase
	}
	return resolverZSeguroSpawnGlobal(territorio, zGeodata)
}

func resolverZReferenciaTerritorio(territorio territorioSpawn) int32 {
	if territorio.maxZ == 0 && territorio.minZ == 0 {
		return 0
	}
	if territorio.maxZ < territorio.minZ {
		return territorio.maxZ
	}
	return territorio.minZ + ((territorio.maxZ - territorio.minZ) / 2)
}

func zDentroDaFaixaDoTerritorio(territorio territorioSpawn, z int32) bool {
	if territorio.maxZ == 0 && territorio.minZ == 0 {
		return true
	}
	if territorio.maxZ < territorio.minZ {
		return true
	}
	margem := int32(128)
	if z < territorio.minZ-margem {
		return false
	}
	if z > territorio.maxZ+margem {
		return false
	}
	return true
}

func obterBlocoGeodata(geoX int32, geoY int32) geodataBloco {
	regiao := obterRegiaoGeodataPorGeo(geoX, geoY)
	if regiao == nil {
		return nil
	}
	bx := (geoX / geoBlockCellsX) % geoRegionBlocksX
	by := (geoY / geoBlockCellsY) % geoRegionBlocksY
	if bx < 0 || bx >= geoRegionBlocksX {
		return nil
	}
	if by < 0 || by >= geoRegionBlocksY {
		return nil
	}
	return regiao.blocos[bx][by]
}

func obterRegiaoGeodataPorGeo(geoX int32, geoY int32) *geodataRegiao {
	rx := geoX/geoRegionCellsX + geoTileXMin
	ry := geoY/geoRegionCellsY + geoTileYMin
	return obterRegiaoGeodata(rx, ry)
}

func obterRegiaoGeodata(rx int32, ry int32) *geodataRegiao {
	if geodataAtual == nil {
		return nil
	}
	if rx < geoTileXMin || rx > geoTileXMax {
		return nil
	}
	if ry < geoTileYMin || ry > geoTileYMax {
		return nil
	}
	chave := fmt.Sprintf("%d_%d", rx, ry)
	geodataAtual.mu.RLock()
	if geodataAtual.cacheRegiao != nil {
		regiaoEmCache := geodataAtual.cacheRegiao[chave]
		if regiaoEmCache != nil {
			geodataAtual.mu.RUnlock()
			return regiaoEmCache
		}
	}
	caminhoGeodata := geodataAtual.caminhoGeodata
	geodataAtual.mu.RUnlock()
	regiaoCarregada, err := carregarRegiaoGeodataL2off(caminhoGeodata, rx, ry)
	if err != nil {
		logger.Warnf("Falha ao carregar regiao de geodata regiao=%s erro=%v", chave, err)
		return nil
	}
	geodataAtual.mu.Lock()
	defer geodataAtual.mu.Unlock()
	if geodataAtual.cacheRegiao == nil {
		geodataAtual.cacheRegiao = map[string]*geodataRegiao{}
	}
	regiaoEmCache := geodataAtual.cacheRegiao[chave]
	if regiaoEmCache != nil {
		return regiaoEmCache
	}
	geodataAtual.cacheRegiao[chave] = regiaoCarregada
	return regiaoCarregada
}

func carregarRegiaoGeodataL2off(caminhoGeodata string, rx int32, ry int32) (*geodataRegiao, error) {
	if strings.TrimSpace(caminhoGeodata) == "" {
		return nil, fmt.Errorf("caminho de geodata vazio")
	}
	nomeArquivo := fmt.Sprintf("%d_%d_conv.dat", rx, ry)
	caminhoArquivo := filepath.Join(caminhoGeodata, nomeArquivo)
	dados, err := os.ReadFile(caminhoArquivo)
	if err != nil {
		return nil, err
	}
	if len(dados) <= int(geoCabecalhoL2offBytes) {
		return nil, fmt.Errorf("arquivo de geodata invalido nome=%s tamanho=%d", nomeArquivo, len(dados))
	}
	leitor := &leitorBytesGeodata{dados: dados, posicao: geoCabecalhoL2offBytes}
	blocos := make([][]geodataBloco, geoRegionBlocksX)
	for bx := int32(0); bx < geoRegionBlocksX; bx++ {
		blocos[bx] = make([]geodataBloco, geoRegionBlocksY)
		for by := int32(0); by < geoRegionBlocksY; by++ {
			tipo, erroTipo := leitor.lerUint16()
			if erroTipo != nil {
				return nil, fmt.Errorf("falha ao ler tipo de bloco nome=%s bx=%d by=%d erro=%w", nomeArquivo, bx, by, erroTipo)
			}
			bloco, erroBloco := lerBlocoGeodataL2off(leitor, tipo)
			if erroBloco != nil {
				return nil, fmt.Errorf("falha ao ler bloco nome=%s bx=%d by=%d tipo=%d erro=%w", nomeArquivo, bx, by, tipo, erroBloco)
			}
			blocos[bx][by] = bloco
		}
	}
	return &geodataRegiao{rx: rx, ry: ry, blocos: blocos}, nil
}

func lerBlocoGeodataL2off(leitor *leitorBytesGeodata, tipo uint16) (geodataBloco, error) {
	if tipo == uint16(geoTipoFlatL2jL2off) {
		altura, errAltura := leitor.lerInt16()
		if errAltura != nil {
			return nil, errAltura
		}
		_, errDummy := leitor.lerUint16()
		if errDummy != nil {
			return nil, errDummy
		}
		return &geodataBlocoFlat{altura: altura, nswe: geoCellFlagAll}, nil
	}
	if tipo == uint16(geoTipoComplexL2off) {
		bloco := &geodataBlocoComplexo{}
		for indice := int32(0); indice < geoBlockCellsTotal; indice++ {
			data, errData := leitor.lerUint16()
			if errData != nil {
				return nil, errData
			}
			bloco.celulas[indice] = decodificarCelulaCompactada(data)
		}
		return bloco, nil
	}
	bloco := &geodataBlocoMulticamadas{}
	for indice := int32(0); indice < geoBlockCellsTotal; indice++ {
		quantidadeCamadas, errCamadas := leitor.lerUint16()
		if errCamadas != nil {
			return nil, errCamadas
		}
		if quantidadeCamadas <= 0 {
			return nil, fmt.Errorf("quantidade de camadas invalida=%d", quantidadeCamadas)
		}
		camadas := make([]geodataCamada, 0, quantidadeCamadas)
		for camada := uint16(0); camada < quantidadeCamadas; camada++ {
			data, errData := leitor.lerUint16()
			if errData != nil {
				return nil, errData
			}
			celula := decodificarCelulaCompactada(data)
			camadas = append(camadas, geodataCamada{altura: celula.altura, nswe: celula.nswe})
		}
		bloco.celulas[indice] = camadas
	}
	return bloco, nil
}

func decodificarCelulaCompactada(data uint16) geodataCelulaComplexa {
	nswe := byte(data & 0x000F)
	altura := int16(int16(data&0xFFF0) >> 1)
	return geodataCelulaComplexa{altura: altura, nswe: nswe}
}

type leitorBytesGeodata struct {
	dados   []byte
	posicao int32
}

func (l *leitorBytesGeodata) lerUint16() (uint16, error) {
	if l == nil {
		return 0, fmt.Errorf("leitor nulo")
	}
	if l.posicao < 0 {
		return 0, fmt.Errorf("posicao invalida")
	}
	if int(l.posicao)+2 > len(l.dados) {
		return 0, fmt.Errorf("fim do arquivo")
	}
	valor := binary.LittleEndian.Uint16(l.dados[l.posicao : l.posicao+2])
	l.posicao += 2
	return valor, nil
}

func (l *leitorBytesGeodata) lerInt16() (int16, error) {
	valor, err := l.lerUint16()
	if err != nil {
		return 0, err
	}
	return int16(valor), nil
}

func (b *geodataBlocoNulo) hasGeoPos() bool {
	return false
}

func (b *geodataBlocoNulo) getHeightNearest(geoX int32, geoY int32, worldZ int32) int32 {
	return worldZ
}

func (b *geodataBlocoNulo) getNsweNearest(geoX int32, geoY int32, worldZ int32) byte {
	return geoCellFlagAll
}

func (b *geodataBlocoNulo) getIndexBelow(geoX int32, geoY int32, worldZ int32) int32 {
	return 0
}

func (b *geodataBlocoNulo) getHeightPorIndice(geoX int32, geoY int32, indice int32) int32 {
	return 0
}

func (b *geodataBlocoNulo) getNswePorIndice(geoX int32, geoY int32, indice int32) byte {
	return geoCellFlagAll
}

func (b *geodataBlocoFlat) hasGeoPos() bool {
	return true
}

func (b *geodataBlocoFlat) getHeightNearest(geoX int32, geoY int32, worldZ int32) int32 {
	if b == nil {
		return worldZ
	}
	return int32(b.altura)
}

func (b *geodataBlocoFlat) getNsweNearest(geoX int32, geoY int32, worldZ int32) byte {
	if b == nil {
		return geoCellFlagAll
	}
	return b.nswe
}

func (b *geodataBlocoFlat) getIndexBelow(geoX int32, geoY int32, worldZ int32) int32 {
	if b == nil {
		return -1
	}
	return 0
}

func (b *geodataBlocoFlat) getHeightPorIndice(geoX int32, geoY int32, indice int32) int32 {
	if b == nil {
		return 0
	}
	if indice != 0 {
		return 0
	}
	return int32(b.altura)
}

func (b *geodataBlocoFlat) getNswePorIndice(geoX int32, geoY int32, indice int32) byte {
	if b == nil {
		return geoCellFlagAll
	}
	if indice != 0 {
		return geoCellFlagAll
	}
	return b.nswe
}

func (b *geodataBlocoComplexo) hasGeoPos() bool {
	return true
}

func (b *geodataBlocoComplexo) getHeightNearest(geoX int32, geoY int32, worldZ int32) int32 {
	if b == nil {
		return worldZ
	}
	indice := calcularIndiceCelulaBloco(geoX, geoY)
	if indice < 0 || indice >= geoBlockCellsTotal {
		return worldZ
	}
	return int32(b.celulas[indice].altura)
}

func (b *geodataBlocoComplexo) getNsweNearest(geoX int32, geoY int32, worldZ int32) byte {
	if b == nil {
		return geoCellFlagAll
	}
	indice := calcularIndiceCelulaBloco(geoX, geoY)
	if indice < 0 || indice >= geoBlockCellsTotal {
		return geoCellFlagAll
	}
	return b.celulas[indice].nswe
}

func (b *geodataBlocoComplexo) getIndexBelow(geoX int32, geoY int32, worldZ int32) int32 {
	if b == nil {
		return -1
	}
	indice := calcularIndiceCelulaBloco(geoX, geoY)
	if indice < 0 || indice >= geoBlockCellsTotal {
		return -1
	}
	altura := int32(b.celulas[indice].altura)
	if altura > worldZ {
		return -1
	}
	return 0
}

func (b *geodataBlocoComplexo) getHeightPorIndice(geoX int32, geoY int32, indice int32) int32 {
	if b == nil {
		return 0
	}
	if indice != 0 {
		return 0
	}
	indiceCelula := calcularIndiceCelulaBloco(geoX, geoY)
	if indiceCelula < 0 || indiceCelula >= geoBlockCellsTotal {
		return 0
	}
	return int32(b.celulas[indiceCelula].altura)
}

func (b *geodataBlocoComplexo) getNswePorIndice(geoX int32, geoY int32, indice int32) byte {
	if b == nil {
		return geoCellFlagAll
	}
	if indice != 0 {
		return geoCellFlagAll
	}
	indiceCelula := calcularIndiceCelulaBloco(geoX, geoY)
	if indiceCelula < 0 || indiceCelula >= geoBlockCellsTotal {
		return geoCellFlagAll
	}
	return b.celulas[indiceCelula].nswe
}

func (b *geodataBlocoMulticamadas) hasGeoPos() bool {
	return true
}

func (b *geodataBlocoMulticamadas) getHeightNearest(geoX int32, geoY int32, worldZ int32) int32 {
	if b == nil {
		return worldZ
	}
	indice := calcularIndiceCelulaBloco(geoX, geoY)
	if indice < 0 || indice >= geoBlockCellsTotal {
		return worldZ
	}
	camadas := b.celulas[indice]
	if len(camadas) <= 0 {
		return worldZ
	}
	melhorAltura := int32(camadas[0].altura)
	melhorDistancia := int32(math.MaxInt32)
	for _, camada := range camadas {
		distancia := diferencaAbsolutaInt32(int32(camada.altura), worldZ)
		if distancia > melhorDistancia {
			continue
		}
		melhorDistancia = distancia
		melhorAltura = int32(camada.altura)
		continue
	}
	return melhorAltura
}

func (b *geodataBlocoMulticamadas) getNsweNearest(geoX int32, geoY int32, worldZ int32) byte {
	if b == nil {
		return geoCellFlagAll
	}
	indice := calcularIndiceCelulaBloco(geoX, geoY)
	if indice < 0 || indice >= geoBlockCellsTotal {
		return geoCellFlagAll
	}
	camadas := b.celulas[indice]
	if len(camadas) <= 0 {
		return geoCellFlagAll
	}
	melhorIndice := encontrarIndiceCamadaMaisProxima(camadas, worldZ)
	if melhorIndice < 0 {
		return geoCellFlagAll
	}
	return camadas[melhorIndice].nswe
}

func (b *geodataBlocoMulticamadas) getIndexBelow(geoX int32, geoY int32, worldZ int32) int32 {
	if b == nil {
		return -1
	}
	indice := calcularIndiceCelulaBloco(geoX, geoY)
	if indice < 0 || indice >= geoBlockCellsTotal {
		return -1
	}
	camadas := b.celulas[indice]
	if len(camadas) <= 0 {
		return -1
	}
	melhorIndice := int32(-1)
	melhorAltura := int32(-2147483648)
	for indiceCamada, camada := range camadas {
		altura := int32(camada.altura)
		if altura > worldZ {
			continue
		}
		if melhorIndice >= 0 && altura <= melhorAltura {
			continue
		}
		melhorIndice = int32(indiceCamada)
		melhorAltura = altura
	}
	return melhorIndice
}

func (b *geodataBlocoMulticamadas) getHeightPorIndice(geoX int32, geoY int32, indice int32) int32 {
	if b == nil {
		return 0
	}
	indiceCelula := calcularIndiceCelulaBloco(geoX, geoY)
	if indiceCelula < 0 || indiceCelula >= geoBlockCellsTotal {
		return 0
	}
	camadas := b.celulas[indiceCelula]
	if indice < 0 || int(indice) >= len(camadas) {
		return 0
	}
	return int32(camadas[indice].altura)
}

func (b *geodataBlocoMulticamadas) getNswePorIndice(geoX int32, geoY int32, indice int32) byte {
	if b == nil {
		return geoCellFlagAll
	}
	indiceCelula := calcularIndiceCelulaBloco(geoX, geoY)
	if indiceCelula < 0 || indiceCelula >= geoBlockCellsTotal {
		return geoCellFlagAll
	}
	camadas := b.celulas[indiceCelula]
	if indice < 0 || int(indice) >= len(camadas) {
		return geoCellFlagAll
	}
	return camadas[indice].nswe
}

func encontrarIndiceCamadaMaisProxima(camadas []geodataCamada, worldZ int32) int32 {
	if len(camadas) <= 0 {
		return -1
	}
	melhorIndice := int32(0)
	melhorDistancia := int32(math.MaxInt32)
	for indiceCamada, camada := range camadas {
		distancia := diferencaAbsolutaInt32(int32(camada.altura), worldZ)
		if distancia > melhorDistancia {
			continue
		}
		melhorIndice = int32(indiceCamada)
		melhorDistancia = distancia
	}
	return melhorIndice
}

func calcularIndiceCelulaBloco(geoX int32, geoY int32) int32 {
	cx := geoX % geoBlockCellsX
	if cx < 0 {
		cx += geoBlockCellsX
	}
	cy := geoY % geoBlockCellsY
	if cy < 0 {
		cy += geoBlockCellsY
	}
	return cx*geoBlockCellsY + cy
}
