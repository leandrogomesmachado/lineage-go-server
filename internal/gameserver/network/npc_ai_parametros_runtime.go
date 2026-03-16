package network

func (n *npcGlobalRuntime) obterNpcIntAiParamOuPadrao(nome string, valorPadrao int32) int32 {
	if n == nil {
		return valorPadrao
	}
	return n.aiParams.obterInteiroOuPadrao(nome, valorPadrao)
}

func (n *npcGlobalRuntime) obterNpcStringAiParamOuPadrao(nome string, valorPadrao string) string {
	if n == nil {
		return valorPadrao
	}
	return n.aiParams.obterStringOuPadrao(nome, valorPadrao)
}
