package network

import (
	"fmt"
	"strings"

	"github.com/leandrogomesmachado/l2raptors-go/pkg/logger"
)

type npcGlobalRuntime struct {
	objID         int32
	npcID         int32
	idTemplate    int32
	nome          string
	titulo        string
	tipo          string
	tipoAI        string
	ehMonster     bool
	nivel         int32
	hpMaximo      int32
	mpMaximo      int32
	pAtk          int32
	pDef          int32
	mAtk          int32
	mDef          int32
	crit          int32
	aggroRange    int32
	origemX       int32
	origemY       int32
	origemZ       int32
	alvoObjID     int32
	x             int32
	y             int32
	z             int32
	heading       int32
	regiaoX       int32
	regiaoY       int32
	ultimoMoveX   int32
	ultimoMoveY   int32
	ultimoMoveZ   int32
	radiusColisao float64
	heightColisao float64
	runSpd        int32
	walkSpd       int32
	pAtkSpd       int32
	mAtkSpd       int32
	rHand         int32
	lHand         int32
	canMove       bool
	canBeAttacked bool
}

func (n *npcGlobalRuntime) atualizarRegiao() {
	if n == nil {
		return
	}
	n.regiaoX = calcularRegiaoX(n.x)
	n.regiaoY = calcularRegiaoY(n.y)
}

func construirNpcsGlobaisDoSpawn() []*npcGlobalRuntime {
	templates := obterTemplatesSpawnGlobal()
	resultado := make([]*npcGlobalRuntime, 0)
	proximoObjID := int32(700000000)
	for _, maker := range templates.makers {
		territorio, ok := templates.territorios[maker.territorioNome]
		if !ok {
			continue
		}
		for _, npcMaker := range maker.npcs {
			templateNpc, okNpc := obterTemplateNpc(npcMaker.npcID)
			if !okNpc {
				continue
			}
			total := npcMaker.total
			if total <= 0 {
				total = 1
			}
			for indice := int32(0); indice < total; indice++ {
				x, y, z, heading := resolverPosicaoSpawnGlobal(territorio, npcMaker.pos, indice)
				npc := &npcGlobalRuntime{objID: proximoObjID, npcID: npcMaker.npcID, idTemplate: templateNpc.idTemplate, nome: templateNpc.nome, titulo: templateNpc.titulo, tipo: templateNpc.tipo, tipoAI: maker.tipoAI, ehMonster: templateNpc.ehMonster(), nivel: templateNpc.nivel, hpMaximo: templateNpc.hp, mpMaximo: templateNpc.mp, pAtk: templateNpc.pAtk, pDef: templateNpc.pDef, mAtk: templateNpc.mAtk, mDef: templateNpc.mDef, crit: templateNpc.crit, aggroRange: templateNpc.aggroRange, origemX: x, origemY: y, origemZ: z, x: x, y: y, z: z, heading: heading, ultimoMoveX: x, ultimoMoveY: y, ultimoMoveZ: z, radiusColisao: templateNpc.radius, heightColisao: templateNpc.height, runSpd: templateNpc.runSpd, walkSpd: templateNpc.walkSpd, pAtkSpd: templateNpc.pAtkSpd, mAtkSpd: templateNpc.mAtkSpd, rHand: templateNpc.rHand, lHand: templateNpc.lHand, canMove: templateNpc.canMove, canBeAttacked: templateNpc.canBeAttacked}
				npc.atualizarRegiao()
				resultado = append(resultado, npc)
				proximoObjID++
			}
		}
	}
	return resultado
}

func resolverPosicaoSpawnGlobal(territorio territorioSpawn, pos string, indice int32) (int32, int32, int32, int32) {
	if pos != "" {
		partes := strings.Split(pos, ";")
		if len(partes) >= 4 {
			x := parseInt32Seguro(partes[0])
			y := parseInt32Seguro(partes[1])
			z := parseInt32Seguro(partes[2])
			heading := parseInt32Seguro(partes[3])
			return normalizarPosicaoGlobal(x, y, z, heading)
		}
	}
	quantidade := int32(len(territorio.nos))
	if quantidade <= 0 {
		return 0, 0, 0, 0
	}
	indiceBase := indice % quantidade
	baseX := territorio.nos[indiceBase].x
	baseY := territorio.nos[indiceBase].y
	baseZ := (territorio.minZ + territorio.maxZ) / 2
	deslocamentoFaixa := indice / quantidade
	deslocamentoX := (deslocamentoFaixa % 3) * 40
	deslocamentoY := ((deslocamentoFaixa / 3) % 3) * 40
	return normalizarPosicaoGlobal(baseX+deslocamentoX, baseY+deslocamentoY, baseZ, 0)
}

func normalizarPosicaoGlobal(x int32, y int32, z int32, heading int32) (int32, int32, int32, int32) {
	xAjustado, yAjustado, zAjustado := normalizarPosicaoMundo(x, y, z)
	return xAjustado, yAjustado, zAjustado, heading
}

func (g *gameServer) inicializarNpcsGlobais() {
	if g == nil {
		return
	}
	if g.mundo == nil {
		return
	}
	g.mundo.limparNpcs()
	npcs := construirNpcsGlobaisDoSpawn()
	amostraLogada := 0
	for _, npc := range npcs {
		g.mundo.registrarNpc(npc)
		if amostraLogada >= 5 {
			continue
		}
		logger.Infof("NPC global instanciado objID=%d npcID=%d idTemplate=%d nome=%s pos=(%d,%d,%d) heading=%d", npc.objID, npc.npcID, npc.idTemplate, npc.nome, npc.x, npc.y, npc.z, npc.heading)
		amostraLogada++
	}
	logger.Infof("NPCs globais inicializados: %d", len(npcs))
}

func nomeNpcGlobalLog(npc *npcGlobalRuntime) string {
	if npc == nil {
		return "npc_global_nil"
	}
	return fmt.Sprintf("%s(%d)", npc.nome, npc.npcID)
}

func (n *npcGlobalRuntime) aplicarPosicao(x int32, y int32, z int32) {
	if n == nil {
		return
	}
	n.x = x
	n.y = y
	n.z = z
	n.atualizarRegiao()
}

func (n *npcGlobalRuntime) aplicarPosicaoComHeading(x int32, y int32, z int32, heading int32) {
	if n == nil {
		return
	}
	n.x = x
	n.y = y
	n.z = z
	n.heading = heading
	n.atualizarRegiao()
}
