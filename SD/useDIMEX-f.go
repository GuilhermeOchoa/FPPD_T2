// Construido como parte da disciplina: Sistemas Distribuidos - PUCRS - Escola Politecnica
//  Professor: Fernando Dotti  (https://fldotti.github.io/)
// Uso p exemplo:
//   go run useDIMEX-f.go 0 127.0.0.1:5000  127.0.0.1:6001  127.0.0.1:7002 ")
//   go run useDIMEX-f.go 1 127.0.0.1:5000  127.0.0.1:6001  127.0.0.1:7002 ")
//   go run useDIMEX-f.go 2 127.0.0.1:5000  127.0.0.1:6001  127.0.0.1:7002 ")
// ----------
// LANCAR N PROCESSOS EM SHELL's DIFERENTES, UMA PARA CADA PROCESSO.
// para cada processo fornecer: seu id único (0, 1, 2 ...) e a mesma lista de processos.
// o endereco de cada processo é o dado na lista, na posicao do seu id.
// no exemplo acima o processo com id=1  usa a porta 6001 para receber e as portas
// 5000 e 7002 para mandar mensagens respectivamente para processos com id=0 e 2
// -----------
// Esta versão supõe que todos processos tem acesso a um mesmo arquivo chamado "mxOUT.txt"
// Todos processos escrevem neste arquivo, usando o protocolo dimex para exclusao mutua.
// Os processos escrevem "|." cada vez que acessam o arquivo.   Assim, o arquivo com conteúdo
// correto deverá ser uma sequencia de
// |.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.
// |.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.
// |.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.|.
// etc etc ...     ....  até o usuário interromper os processos (ctl c).
// Qualquer padrao diferente disso, revela um erro.
//      |.|.|.|.|.||..|.|.|.  etc etc  por exemplo.
// Se voce retirar o protocolo dimex vai ver que o arquivo poderá entrelacar
// "|."  dos processos de diversas diferentes formas.
// Ou seja, o padrão correto acima é garantido pelo dimex.
// Ainda assim, isto é apenas um teste.  E testes são frágeis em sistemas distribuídos.

package main

import (
	"SD/DIMEX" // Importa o pacote DIMEX que contém a implementação do protocolo de exclusão mútua.
	"fmt"      // Importa o pacote fmt para realizar operações de I/O formatadas, como imprimir na tela.
	"os"       // Importa o pacote fmt para realizar operações de I/O formatadas, como imprimir na tela.
	"strconv"  // Importa o pacote strconv para converter strings para outros tipos de dados.
	"time"     // Importa o pacote time para manipular operações relacionadas a tempo, como esperas.F
)

func main() {

	if len(os.Args) < 2 { // Verifica se o número de argumentos de linha de comando é menor que 2 (o que significa que falta algum argumento).
		// Se faltar argumentos, imprime as instruções de como executar o programa corretamente.
		fmt.Println("Please specify at least one address:port!")
		fmt.Println("go run useDIMEX-f.go 0 127.0.0.1:5000  127.0.0.1:6001  127.0.0.1:7002 ")
		fmt.Println("go run useDIMEX-f.go 1 127.0.0.1:5000  127.0.0.1:6001  127.0.0.1:7002 ")
		fmt.Println("go run useDIMEX-f.go 2 127.0.0.1:5000  127.0.0.1:6001  127.0.0.1:7002 ")
		return
	}

	id, _ := strconv.Atoi(os.Args[1]) // Converte o segundo argumento de linha de comando para um inteiro, que representa o ID do processo.
	addresses := os.Args[2:]          // Armazena todos os argumentos após o segundo em um slice, que são os endereços dos outros processos.
	fmt.Print("id: ", id, "   ")      // Imprime o ID do processo.
	fmt.Println(addresses)            // Imprime os endereços dos outros processos.

	// Cria uma nova instância do módulo DIMEX com os endereços fornecidos e o ID.
	var dmx *DIMEX.DIMEX_Module = DIMEX.NewDIMEX(addresses, id, true)
	//fmt.Println(dmx)

	// abre arquivo que TODOS processos devem poder usar
	// Tenta abrir o arquivo mxOUT.txt para adicionar conteúdo ao final.
	file, err := os.OpenFile("./mxOUT.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err) // Se houver erro ao abrir o arquivo, imprime a mensagem de erro.
		return                                  // Encerra o programa se não puder abrir o arquivo.
	}
	defer file.Close() // Garante que o arquivo será fechado no final da execução da função main.

	// espera para facilitar inicializacao de todos processos (a mao)
	time.Sleep(3 * time.Second)

	// Inicia um loop infinito.
	for {
		// SOLICITA ACESSO AO DIMEX
		fmt.Println("[ APP id: ", id, " PEDE   MX ]") // Imprime que o processo está pedindo acesso exclusivo (MX).
		dmx.Req <- DIMEX.ENTER                        // Envia uma mensagem para o canal Req do módulo DIMEX para entrar na seção crítica.
		fmt.Println("[ APP id: ", id, " ESPERA MX ]") // Imprime que o processo está esperando a liberação do acesso exclusivo.
		// ESPERA LIBERACAO DO MODULO DIMEX
		<-dmx.Ind                                     // Espera receber uma mensagem no canal Ind, que indica que o acesso foi concedido.
		fmt.Println("[ APP id: ", id, " RECEBE MX ]") // Imprime que o processo recebeu acesso exclusivo.
		// A PARTIR DAQUI ESTA ACESSANDO O ARQUIVO SOZINHO
		_, err = file.WriteString("|") // Escreve o caractere '|' no arquivo, que marca a entrada na seção crítica.
		if err != nil {
			fmt.Println("Error writing to file:", err) // Se houver erro ao escrever no arquivo, imprime a mensagem de erro.
			return                                     // Encerra o programa se não puder escrever no arquivo.
		}

		fmt.Println("[ APP id: ", id, " *EM*   MX ]") // Imprime que o processo está dentro da seção crítica.

		_, err = file.WriteString(".") // Escreve o caractere '.' no arquivo, que marca a saída da seção crítica.
		if err != nil {
			fmt.Println("Error writing to file:", err) // Se houver erro ao escrever no arquivo, imprime a mensagem de erro.
			return                                     // Encerra o programa se não puder escrever no arquivo.
		}

		// AGORA VAI LIBERAR O ARQUIVO PARA OUTROS
		dmx.Req <- DIMEX.EXIT                         // Envia uma mensagem para o canal Req do módulo DIMEX para sair da seção crítica.
		fmt.Println("[ APP id: ", id, " FORA   MX ]") // Imprime que o processo saiu da seção crítica.
	}
}
