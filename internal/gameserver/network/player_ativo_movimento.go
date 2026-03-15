package network

func (p *playerAtivo) iniciarMovimento(destinoX int32, destinoY int32, destinoZ int32, heading int32) {
	if p == nil {
		return
	}
	p.movendo = true
	p.origemMovX = p.x
	p.origemMovY = p.y
	p.origemMovZ = p.z
	p.destinoX = destinoX
	p.destinoY = destinoY
	p.destinoZ = destinoZ
	p.heading = heading
	p.ultimoMoveX = p.x
	p.ultimoMoveY = p.y
	p.ultimoMoveZ = p.z
}

func (p *playerAtivo) pararMovimento() {
	if p == nil {
		return
	}
	p.movendo = false
	p.origemMovX = p.x
	p.origemMovY = p.y
	p.origemMovZ = p.z
	p.destinoX = p.x
	p.destinoY = p.y
	p.destinoZ = p.z
	p.ultimoMoveX = p.x
	p.ultimoMoveY = p.y
	p.ultimoMoveZ = p.z
}

func (p *playerAtivo) estaMovendo() bool {
	if p == nil {
		return false
	}
	return p.movendo
}
