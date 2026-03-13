package network

import "math"

type locSpawnInicial struct {
	x int32
	y int32
	z int32
}

type templatePersonagemInicial struct {
	classID         int32
	race            int32
	str             int32
	dex             int32
	con             int32
	intel           int32
	wit             int32
	men             int32
	x               int32
	y               int32
	z               int32
	maxHp           int32
	maxMp           int32
	maxCp           int32
	pAtk            int32
	pDef            int32
	mAtk            int32
	mDef            int32
	runSpd          int32
	walkSpd         int32
	swimSpd         int32
	pAtkSpd         int32
	mAtkSpd         int32
	radiusMasculino float64
	radiusFeminino  float64
	heightMasculino float64
	heightFeminino  float64
	spawns          []locSpawnInicial
}

var templatesPersonagemInicial = []templatePersonagemInicial{
	{classID: 0, race: 0, str: 40, dex: 30, con: 43, intel: 21, wit: 11, men: 25, x: -71338, y: 258271, z: -3104, maxHp: 80, maxMp: 30, maxCp: 32, pAtk: 4, pDef: 80, mAtk: 6, mDef: 41, runSpd: 120, walkSpd: 80, swimSpd: 50, pAtkSpd: 300, mAtkSpd: 333, radiusMasculino: 9, radiusFeminino: 8, heightMasculino: 23, heightFeminino: 23.5, spawns: []locSpawnInicial{{x: -71338, y: 258271, z: -3104}, {x: -71417, y: 258270, z: -3104}, {x: -71453, y: 258305, z: -3104}, {x: -71467, y: 258378, z: -3104}}},
	{classID: 10, race: 0, str: 22, dex: 21, con: 27, intel: 41, wit: 20, men: 39, x: -90875, y: 248162, z: -3570, maxHp: 101, maxMp: 40, maxCp: 51, pAtk: 3, pDef: 54, mAtk: 6, mDef: 41, runSpd: 120, walkSpd: 80, swimSpd: 50, pAtkSpd: 333, mAtkSpd: 333, radiusMasculino: 7.5, radiusFeminino: 6.5, heightMasculino: 22.8, heightFeminino: 22.5, spawns: []locSpawnInicial{{x: -90875, y: 248162, z: -3570}, {x: -90954, y: 248118, z: -3570}, {x: -90918, y: 248070, z: -3570}, {x: -90890, y: 248027, z: -3570}}},
	{classID: 18, race: 1, str: 36, dex: 35, con: 36, intel: 23, wit: 14, men: 26, x: 46045, y: 41251, z: -3440, maxHp: 89, maxMp: 30, maxCp: 36, pAtk: 4, pDef: 80, mAtk: 6, mDef: 41, runSpd: 122, walkSpd: 85, swimSpd: 50, pAtkSpd: 300, mAtkSpd: 333, radiusMasculino: 7.5, radiusFeminino: 7.5, heightMasculino: 24, heightFeminino: 23, spawns: []locSpawnInicial{{x: 46045, y: 41251, z: -3440}, {x: 46117, y: 41247, z: -3440}, {x: 46182, y: 41198, z: -3440}, {x: 46115, y: 41141, z: -3440}, {x: 46048, y: 41141, z: -3440}, {x: 45978, y: 41196, z: -3440}}},
	{classID: 25, race: 1, str: 21, dex: 24, con: 25, intel: 37, wit: 23, men: 40, x: 46045, y: 41251, z: -3440, maxHp: 104, maxMp: 40, maxCp: 52, pAtk: 3, pDef: 54, mAtk: 6, mDef: 41, runSpd: 122, walkSpd: 85, swimSpd: 50, pAtkSpd: 333, mAtkSpd: 333, radiusMasculino: 7.5, radiusFeminino: 7.5, heightMasculino: 23.5, heightFeminino: 22.5, spawns: []locSpawnInicial{{x: 46045, y: 41251, z: -3440}, {x: 46117, y: 41247, z: -3440}, {x: 46182, y: 41198, z: -3440}, {x: 46115, y: 41141, z: -3440}, {x: 46048, y: 41141, z: -3440}, {x: 45978, y: 41196, z: -3440}}},
	{classID: 31, race: 2, str: 41, dex: 34, con: 32, intel: 25, wit: 12, men: 26, x: 28295, y: 11063, z: -4224, maxHp: 94, maxMp: 30, maxCp: 38, pAtk: 4, pDef: 80, mAtk: 6, mDef: 41, runSpd: 122, walkSpd: 85, swimSpd: 50, pAtkSpd: 300, mAtkSpd: 333, radiusMasculino: 7.5, radiusFeminino: 7, heightMasculino: 24, heightFeminino: 23.5, spawns: []locSpawnInicial{{x: 28295, y: 11063, z: -4224}, {x: 28302, y: 11008, z: -4224}, {x: 28377, y: 10916, z: -4224}, {x: 28456, y: 10997, z: -4224}, {x: 28461, y: 11044, z: -4224}, {x: 28395, y: 11127, z: -4224}}},
	{classID: 38, race: 2, str: 23, dex: 23, con: 24, intel: 44, wit: 19, men: 37, x: 28295, y: 11063, z: -4224, maxHp: 106, maxMp: 40, maxCp: 53, pAtk: 3, pDef: 54, mAtk: 6, mDef: 41, runSpd: 122, walkSpd: 85, swimSpd: 50, pAtkSpd: 333, mAtkSpd: 333, radiusMasculino: 7.5, radiusFeminino: 7, heightMasculino: 24, heightFeminino: 23.5, spawns: []locSpawnInicial{{x: 28295, y: 11063, z: -4224}, {x: 28302, y: 11008, z: -4224}, {x: 28377, y: 10916, z: -4224}, {x: 28456, y: 10997, z: -4224}, {x: 28461, y: 11044, z: -4224}, {x: 28395, y: 11127, z: -4224}}},
	{classID: 44, race: 3, str: 40, dex: 26, con: 47, intel: 18, wit: 12, men: 27, x: -56733, y: -113459, z: -690, maxHp: 80, maxMp: 30, maxCp: 32, pAtk: 4, pDef: 80, mAtk: 6, mDef: 41, runSpd: 117, walkSpd: 80, swimSpd: 50, pAtkSpd: 300, mAtkSpd: 333, radiusMasculino: 11, radiusFeminino: 7, heightMasculino: 28, heightFeminino: 25, spawns: []locSpawnInicial{{x: -56733, y: -113459, z: -690}, {x: -56686, y: -113470, z: -690}, {x: -56728, y: -113610, z: -690}, {x: -56693, y: -113610, z: -690}, {x: -56743, y: -113757, z: -690}, {x: -56682, y: -113730, z: -690}}},
	{classID: 49, race: 3, str: 27, dex: 24, con: 31, intel: 31, wit: 15, men: 42, x: -56733, y: -113459, z: -690, maxHp: 95, maxMp: 40, maxCp: 47, pAtk: 3, pDef: 54, mAtk: 6, mDef: 41, runSpd: 117, walkSpd: 80, swimSpd: 50, pAtkSpd: 333, mAtkSpd: 333, radiusMasculino: 7, radiusFeminino: 8, heightMasculino: 24, heightFeminino: 26, spawns: []locSpawnInicial{{x: -56733, y: -113459, z: -690}, {x: -56686, y: -113470, z: -690}, {x: -56728, y: -113610, z: -690}, {x: -56693, y: -113610, z: -690}, {x: -56743, y: -113757, z: -690}, {x: -56682, y: -113730, z: -690}}},
	{classID: 53, race: 4, str: 39, dex: 29, con: 45, intel: 20, wit: 10, men: 27, x: 108644, y: -173947, z: -400, maxHp: 80, maxMp: 30, maxCp: 32, pAtk: 4, pDef: 80, mAtk: 6, mDef: 41, runSpd: 126, walkSpd: 87, swimSpd: 50, pAtkSpd: 300, mAtkSpd: 333, radiusMasculino: 9, radiusFeminino: 5, heightMasculino: 18.5, heightFeminino: 19, spawns: []locSpawnInicial{{x: 108644, y: -173947, z: -400}, {x: 108678, y: -174002, z: -400}, {x: 108505, y: -173964, z: -400}, {x: 108512, y: -174026, z: -400}, {x: 108549, y: -174075, z: -400}, {x: 108576, y: -174122, z: -400}}},
}

func obterTemplatePersonagemInicial(classID int32) (templatePersonagemInicial, bool) {
	for _, template := range templatesPersonagemInicial {
		if template.classID == classID {
			return template, true
		}
	}
	return templatePersonagemInicial{}, false
}

func (t templatePersonagemInicial) obterColisao(sexo int32) (float64, float64) {
	if sexo == 0 {
		return t.radiusMasculino, t.heightMasculino
	}
	return t.radiusFeminino, t.heightFeminino
}

func (t templatePersonagemInicial) obterSpawnInicial(seletor int32) locSpawnInicial {
	if len(t.spawns) == 0 {
		return locSpawnInicial{x: t.x, y: t.y, z: t.z}
	}
	indice := int(seletor)
	if indice < 0 {
		indice = -indice
	}
	indice = indice % len(t.spawns)
	return t.spawns[indice]
}

func (t templatePersonagemInicial) obterCpMaximoPorNivel(nivel int32) int32 {
	if nivel <= 1 {
		return t.maxCp
	}
	valorCalculado := float64(t.maxCp) + (float64(nivel-1) * (float64(t.maxCp) * 0.12))
	return int32(math.Round(valorCalculado))
}
