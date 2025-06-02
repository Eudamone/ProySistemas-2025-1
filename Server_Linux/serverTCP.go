package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

func main() {
	fmt.Println("Creando socket abriendo puerto 1705")
	// IP a 0.0.0.0 para aceptar conexiones de cualquier IP
	socketInicial, err := net.Listen("tcp", "0.0.0.0:1705")
	if err != nil {
		fmt.Println("Error al abrir el puerto:", err)
		return
	}
	fmt.Println("Socket creado:")
	// Canal para escritura de mensajes
	msgCh := make(chan string)

	fmt.Println("Esperando conexiones....!")
	socketS, err := socketInicial.Accept()
	if err != nil {
		fmt.Println("Error al aceptar conexión:", err)
		return
	}
	fmt.Println("Cliente conectado [", socketS.RemoteAddr().String(), "]")
	fmt.Println("¡Conexión exitosa con el cliente!")

	reader := bufio.NewReader(socketS)
	writer := bufio.NewWriter(socketS)
	go func() {
		for msg := range msgCh {
			writer.WriteString(msg)
			writer.Flush()
		}
	}()

	response, _ := reader.ReadString('\n')
	msgCh <- "Repuesta Servidor: " + response + "\n"

	n, _ := strconv.Atoi(strings.TrimSpace(response))
	fmt.Println("Valor de n recibido del cliente:", n)

	// Manejo del cliente en goroutines
	// Se crean dos goroutines, una para recibir mensajes y otra para enviar reportes
	go recMessage(reader, msgCh)
	go sendReports(n, msgCh)

	// Mantener el servidor activo hasta que se cierre
	for {
	}

	socketS.Close()
}

func recMessage(recBuffer *bufio.Reader, msgCh chan string) {
	for {
		recComando, _ := recBuffer.ReadString('\n')
		fmt.Println("Mensaje recibido del cliente:", recComando)

		rtaComando := "Comando recibido: " + recComando + "\n"

		msgCh <- rtaComando
		fmt.Println(rtaComando)
	}
}

func sendReports(n int, msgCh chan string) {
	x := 0
	for {
		time.Sleep(time.Duration(n) * time.Second)
		//Procesador - Procesos - Memoria - Disco
		report := fmt.Sprintf("REPORT:CPU:%d,PRC:%d,RAM:%d,DD:%d\n", x+5, x+20, x+3, x+10)
		msgCh <- report
		fmt.Println("Enviando reporte al cliente:", report)
		x++
	}
}
