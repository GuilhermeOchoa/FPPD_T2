// Construido como parte da disciplina: Sistemas Distribuidos - PUCRS - Escola Politecnica
//  Professor: Fernando Dotti  (https://fldotti.github.io/)
/* fmt.Println("go run useDIMEX.go 0 127.0.0.1:5000  127.0.0.1:6001  127.0.0.1:7002 ")
   fmt.Println("go run useDIMEX.go 1 127.0.0.1:5000  127.0.0.1:6001  127.0.0.1:7002 ")
   fmt.Println("go run useDIMEX.go 2 127.0.0.1:5000  127.0.0.1:6001  127.0.0.1:7002 ")
/*
  LANCAR N PROCESSOS EM SHELL's DIFERENTES, UMA PARA CADA PROCESSO.
  para cada processo: seu id único e a mesma lista de processos.
  o endereco de cada processo é o dado na lista, na posicao do seu id.
  no exemplo acima o processo com id=1  usa a porta 6001 para receber e as portas
  5000 e 7002 para mandar mensagens respectivamente para processos com id=0 e 2
*/

package main

import (
	"SD/DIMEX" // Importa o pacote DIMEX que contém a implementação do protocolo de exclusão mútua.
	"fmt"      // Importa o pacote fmt para realizar operações de I/O formatadas, como imprimir na tela.
	"os"       // Importa o pacote fmt para realizar operações de I/O formatadas, como imprimir na tela.
	"strconv"  // Importa o pacote strconv para converter strings para outros tipos de dados.
	"time"     // Importa o pacote time para manipular operações relacionadas a tempo, como esperas.
)

// O processo também entra em um loop infinito, pedindo e liberando o acesso exclusivo usando o módulo DIMEX. A principal diferença é que não há operações de escrita em arquivo, e a espera inicial é de 5 segundos 
func main() {

	if len(os.Args) < 2 { // Verifica se o número de argumentos de linha de comando é menor que 2 (o que significa que falta algum argumento).
		// Se faltar argumentos, imprime as instruções de como executar o programa corretamente.
		fmt.Println("Please specify at least one address:port!")
		fmt.Println("go run useDIMEX.go 0 127.0.0.1:5000  127.0.0.1:6001  127.0.0.1:7002 ")
		fmt.Println("go run useDIMEX.go 1 127.0.0.1:5000  127.0.0.1:6001  127.0.0.1:7002 ")
		fmt.Println("go run useDIMEX.go 2 127.0.0.1:5000  127.0.0.1:6001  127.0.0.1:7002 ")
		return
	}

	id, _ := strconv.Atoi(os.Args[1]) // Converte o segundo argumento de linha de comando para um inteiro, que representa o ID do processo.
	addresses := os.Args[2:]          // Armazena todos os argumentos após o segundo em um slice, que são os endereços dos outros processos.
	// fmt.Print("id: ", id, "   ") fmt.Println(addresses)

	var dmx *DIMEX.DIMEX_Module = DIMEX.NewDIMEX(addresses, id, true)
	fmt.Println(dmx)

	time.Sleep(5 * time.Second)

	for {
		fmt.Println("[ APP id: ", id, " PEDE   MX ]")
		dmx.Req <- DIMEX.ENTER
		fmt.Println("[ APP id: ", id, " ESPERA MX ]")
		<-dmx.Ind //
		fmt.Println("[ APP id: ", id, " *EM*   MX ]")
		dmx.Req <- DIMEX.EXIT //
		fmt.Println("[ APP id: ", id, " FORA   MX ]")
	}
}
