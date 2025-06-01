package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

func main() {
	fmt.Println("************************************")
	fmt.Println("***SERVER TCP OPER 2025***")
	fmt.Println("************************************")
	fmt.Println("Creando socket abriendo puerto 1705")
	socketInicial, _ := net.Listen("tcp", "10.1.6.19:1705")
	fmt.Println("Socket creado:")
	fmt.Println("Esperando conexiones....!")
	socketS, _ := socketInicial.Accept()
	fmt.Println("Cliente conectado [", socketS.RemoteAddr().String(), "]")

	go recibeMensaje(&socketS) //con go los hace en paralelo
	go envReporte(&socketS)
	for {

	}
	socketS.Close()

}
func recibeMensaje(socketS *net.Conn) {
	for {
		bufRecibir := bufio.NewReader(*socketS)
		recComando, _ := bufRecibir.ReadString('\n')
		fmt.Println("Comando recibido: ", recComando) //recibe el comando

		if recComando == "bye\n" {
			break
		}

		rtaComando := "Comando ejecutado OK: " + recComando + "\n"
		//envio rta al cliente
		bufEnvio := bufio.NewWriter(*socketS)
		bufEnvio.WriteString(rtaComando)
		bufEnvio.Flush()
		fmt.Println("RTA enviada al cliente! ")
	}
}
func envReporte(socketS *net.Conn) {
	x := 0
	for {
		x++
		time.Sleep(8 * time.Second)
		bufEnvio := bufio.NewWriter(*socketS)
		reporte := fmt.Sprintf("*******Reporte linux -> [Proc=%d] - [Mem=%d] - [DD=%d] - [NumProc=%d]*****\n", 3, 5, 2, 6)
		bufEnvio.WriteString(reporte)
		bufEnvio.Flush()
		fmt.Println("Comando enviado al servidor")

	}

}
