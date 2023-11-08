/*  Construido como parte da disciplina: FPPD - PUCRS - Escola Politecnica
    Professor: Fernando Dotti  (https://fldotti.github.io/)
    Modulo representando Algoritmo de Exclusão Mútua Distribuída:
    Semestre 2023/1
	Aspectos a observar:
	   mapeamento de módulo para estrutura
	   inicializacao
	   semantica de concorrência: cada evento é atômico
	   							  módulo trata 1 por vez
	Q U E S T A O
	   Além de obviamente entender a estrutura ...
	   Implementar o núcleo do algoritmo ja descrito, ou seja, o corpo das
	   funcoes reativas a cada entrada possível:
	   			handleUponReqEntry()  // recebe do nivel de cima (app)
				handleUponReqExit()   // recebe do nivel de cima (app)
				handleUponDeliverRespOk(msgOutro)   // recebe do nivel de baixo
				handleUponDeliverReqEntry(msgOutro) // recebe do nivel de baixo
*/

// O código DIMEX-Template.go define um módulo para gerenciar exclusão mútua distribuída (Distributed Mutual Exclusion - DIMEX) em um ambiente de sistemas distribuídos. Este módulo utiliza mensagens e estados para assegurar que apenas um processo por vez possa entrar em uma seção crítica (acesso a um recurso compartilhado).

package DIMEX

import (
	PP2PLink "SD/PP2PLink"
	"fmt"
	"strings"
)

// ------------------------------------------------------------------------------------
// ------- principais tipos
// ------------------------------------------------------------------------------------

// Define um tipo personalizado para os estados possíveis de um processo na exclusão mútua distribuída.
type State int // enumeracao dos estados possiveis de um processo
const (
	noMX   State = iota // não quer entrar na seção crítica (SC)
	wantMX              // quer entrar na SC
	inMX                // está na SC
)

// Define um tipo para as requisições que o módulo pode receber.
type dmxReq int // enumeracao dos estados possiveis de um processo
const (
	ENTER dmxReq = iota // solicitação para entrar na SC
	EXIT                // solicitação para sair da SC
)

type dmxResp struct { // mensagem do módulo DIMEX informando que pode acessar - pode ser somente um sinal (vazio)
	// mensagem para aplicacao indicando que pode prosseguir
}

// Estrutura principal do módulo DIMEX.
type DIMEX_Module struct {
	Req       chan dmxReq  // canal para receber pedidos da aplicação (ENTER e EXIT)
	Ind       chan dmxResp // canal para informar aplicacao que pode acessar
	addresses []string     // endereços de todos os processos, na mesma ordem
	id        int          // identificador do processo - é o indice no array de enderecos acima
	st        State        // estado deste processo na exclusao mutua distribuida
	waiting   []bool       // processos aguardando tem flag true
	lcl       int          // relogio logico local
	reqTs     int          // timestamp local da ultima requisicao deste processo
	nbrResps  int          // contador de respostas recebidas
	dbg       bool         // flag para depuração

	Pp2plink *PP2PLink.PP2PLink // acesso aa comunicacao enviar por PP2PLinq.Req  e receber por PP2PLinq.Ind
}

// ------------------------------------------------------------------------------------
// ------- inicializacao
// ------------------------------------------------------------------------------------

// Função para criar uma nova instância do módulo DIMEX.
func NewDIMEX(_addresses []string, _id int, _dbg bool) *DIMEX_Module {

	p2p := PP2PLink.NewPP2PLink(_addresses[_id], _dbg) // Cria um novo ponto-a-ponto.

	dmx := &DIMEX_Module{ // Inicializa a estrutura do DIMEX.
		Req: make(chan dmxReq, 1),
		Ind: make(chan dmxResp, 1),

		addresses: _addresses,
		id:        _id,
		st:        noMX,
		waiting:   make([]bool, len(_addresses)),
		lcl:       0,
		reqTs:     0,
		dbg:       _dbg,

		Pp2plink: p2p}

	for i := 0; i < len(dmx.waiting); i++ {
		dmx.waiting[i] = false // Inicializa o vetor de espera.
	}
	dmx.Start()               // Inicia o módulo.
	dmx.outDbg("Init DIMEX!") // Mensagem de depuração.
	return dmx                // Retorna a instância do módulo.
}

// ------------------------------------------------------------------------------------
// ------- nucleo do funcionamento
// ------------------------------------------------------------------------------------

// Inicia o loop principal que lida com eventos.
func (module *DIMEX_Module) Start() {
	go func() { // Inicia uma nova goroutine.
		for { // Loop infinito.
			select { // Espera por um evento.
			case dmxR := <-module.Req: // Se um evento vier do canal Req.
				if dmxR == ENTER { // Se for um pedido de entrada na SC.
					module.outDbg("app pede mx")
					module.handleUponReqEntry() // Chama a função de tratamento de entrada. ENTRADA DO ALGORITMO

				} else if dmxR == EXIT { // Se for um pedido de saída da SC.
					module.outDbg("app libera mx")
					module.handleUponReqExit() // Chama a função de tratamento de saída. ENTRADA DO ALGORITMO
				}

			case msgOutro := <-module.Pp2plink.Ind: // Se um evento vier de outro processo.
				//fmt.Printf("dimex recebe da rede: ", msgOutro)
				// Avalia o conteúdo da mensagem recebida e chama a função de tratamento correspondente.
				if strings.Contains(msgOutro.Message, "respOK") {
					module.outDbg("         <<<---- responde! " + msgOutro.Message)
					module.handleUponDeliverRespOk(msgOutro) // Trata uma resposta de OK. ENTRADA DO ALGORITMO

				} else if strings.Contains(msgOutro.Message, "reqEntry") {
					module.outDbg("          <<<---- pede??  " + msgOutro.Message)
					module.handleUponDeliverReqEntry(msgOutro) // Trata uma solicitação de entrada. ENTRADA DO ALGORITMO

				}
			}
		}
	}()
}

// ------------------------------------------------------------------------------------
// ------- tratamento de pedidos vindos da aplicacao
// ------- UPON ENTRY
// ------- UPON EXIT
// ------------------------------------------------------------------------------------

func (module *DIMEX_Module) handleUponReqEntry() {
	/*
					upon event [ dmx, Entry  |  r ]  do
		    			lts.ts++
		    			myTs := lts
		    			resps := 0
		    			para todo processo p
							trigger [ pl , Send | [ reqEntry, r, myTs ]
		    			estado := queroSC
	*/

	// Aumenta o carimbo de tempo local
	module.lcl++
	// Define o time stamp local da última solicitação deste processo
	module.reqTs = module.lcl
	// Inicializa o contador de respostas
	module.nbrResps = 0
	// Para cada processo p
	for i, addr := range module.addresses {
		if i == module.id {
			//module.outDbg("Não envia para si mesmo: " + addr)
			continue // Não envie solicitação para si mesmo
		}
		// Envia uma mensagem [reqEntry, r, myTs] para o processo p
		message := fmt.Sprintf("reqEntry, %d, %d", module.id, module.reqTs)
		module.sendToLink(addr, message, "reqEntry")
		// Atualiza o estado para "wantMX" (quer a exclusão mútua)
		module.st = wantMX
	}
}
func (module *DIMEX_Module) handleUponReqExit() {
	/*
						upon event [ dmx, Exit  |  r  ]  do
		       				para todo [p, r, ts ] em waiting
		          				trigger [ pl, Send | p , [ respOk, r ]  ]
		    				estado := naoQueroSC
							waiting := {}
	*/
	// Para todos os processos em waiting
	for i := 0; i < len(module.waiting); i++ {
		if module.waiting[i] {
			// Envia uma mensagem [respOk, r] para o processo i
			message := fmt.Sprintf("respOK, %d", module.id)
			//message := fmt.Sprintf("respOK")
			module.sendToLink(module.addresses[i], message, "Exit Response")
		}
	}

	// Atualiza o estado para "noMX" (não está na exclusão mútua)
	module.st = noMX

	// Limpa a lista de processos esperando
	for i := 0; i < len(module.waiting); i++ {
		module.waiting[i] = false
	}

}

// ------------------------------------------------------------------------------------
// ------- tratamento de mensagens de outros processos
// ------- UPON respOK
// ------- UPON reqEntry
// ------------------------------------------------------------------------------------

func (module *DIMEX_Module) handleUponDeliverRespOk(msgOutro PP2PLink.PP2PLink_Ind_Message) {
	/*
						upon event [ pl, Deliver | p, [ respOk, r ] ]
		      				resps++
		      				se resps = N
		    				então trigger [ dmx, Deliver | free2Access ]
		  					    estado := estouNaSC

	*/
	module.outDbg(msgOutro.Message)
	module.outDbg("entrou no handleUponDeliverRespOk")
	//atualiza o contador de respostas
	module.nbrResps++
	if module.nbrResps == len(module.addresses)-1 {
		module.Ind <- dmxResp{}
		module.st = inMX
	}

}

func (module *DIMEX_Module) handleUponDeliverReqEntry(msgOutro PP2PLink.PP2PLink_Ind_Message) {
	// outro processo quer entrar na SC
	/*
						upon event [ pl, Deliver | p, [ reqEntry, r, rts ]  do
		     				se (estado == naoQueroSC)   OR
		        				 (estado == QueroSC AND  myTs >  ts)
							então  trigger [ pl, Send | p , [ respOk, r ]  ]
		 					senão
		        				se (estado == estouNaSC) OR
		           					 (estado == QueroSC AND  myTs < ts)
		        				então  postergados := postergados + [p, r ]
		     					lts.ts := max(lts.ts, rts.ts)
	*/
	//module.outDbg(msgOutro.Message)
	var p, ts int
	_, err := fmt.Sscanf(msgOutro.Message, "reqEntry, %d, %d", &p, &ts)
	if err != nil {
		module.outDbg("Erro ao extrair valores da mensagem")
		// Manipule o erro conforme necessário
		return
	}
	module.outDbg(fmt.Sprintf("p: %d, ts: %d", p, ts))
	module.outDbg(fmt.Sprintf("module.st: %d, module.reqTs: %d", module.st, module.reqTs))
	if module.st == noMX || (module.st == wantMX && (module.reqTs > ts || (module.reqTs == ts && module.id > p))) {
		message := fmt.Sprintf("respOK, %d", module.id)
		module.sendToLink(module.addresses[p], message, "Request Entry Response")
	} else {

		module.outDbg("Adicionando processo na lista de espera" + msgOutro.Message)
		module.waiting[p] = true
		module.outDbg(fmt.Sprintf("module.waiting: %v", module.waiting))
		module.lcl = max(module.lcl, ts)
		module.outDbg(fmt.Sprintf("module.lcl: %d", module.lcl))
	}
}

// ------------------------------------------------------------------------------------
// ------- funcoes de ajuda
// ------------------------------------------------------------------------------------

// Envia uma mensagem para um endereço específico.
func (module *DIMEX_Module) sendToLink(address string, content string, space string) {
	module.outDbg(space + " ---->>>>   to: " + address + "     msg: " + content)
	module.Pp2plink.Req <- PP2PLink.PP2PLink_Req_Message{
		To:      address,
		Message: content}
}

// Compara dois pares de identificador e timestamp para determinar a precedência.
func before(oneId, oneTs, othId, othTs int) bool {
	if oneTs < othTs {
		return true
	} else if oneTs > othTs {
		return false
	} else {
		return oneId < othId
	}
}

// Imprime mensagens de depuração se o modo de depuração estiver habilitado.
func (module *DIMEX_Module) outDbg(s string) {
	if module.dbg {
		fmt.Println(". . . . . . . . . . . . [ DIMEX : " + s + " ]")
	}
}
