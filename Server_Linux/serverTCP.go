package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"net"
	"os"
	"os/exec"
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
	go recCommand(reader, msgCh)
	go sendReports(n, msgCh)

	// Mantener el servidor activo hasta que se cierre
	for {
	}

	socketS.Close()
}

func recCommand(recBuffer *bufio.Reader, msgCh chan string) {
	for {
		command, _ := recBuffer.ReadString('\n')
		fmt.Println("Mensaje recibido del cliente:", command)

		command = strings.TrimRight(command, "\n")
		Scommand := strings.Split(command, ":")
		shell := exec.Command(Scommand[0], Scommand[1:]...)
		resCommand,_ := shell.Output()

		rtaComando := string(resCommand) + "\n"

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

func ValidateUser(user,passw string) bool {
	archivo, err := os.ReadFile("users.txt")
	password := sha256.Sum256([]byte(passw))
	hexapassUser := fmt.Sprintf("%x", password)
	if err != nil {
		fmt.Println("Error al abrir el archivo:", err)
		return false
	}
	contArchivo := string(archivo)
	contArchivo = strings.TrimRight(contArchivo, "\n")
	users := strings.Split(contArchivo, "\n")
	for _, lineUser := range users {
		credentials := strings.Split(lineUser, ":")
		if user == credentials[0] && hexapassUser == credentials[1] {
			return true
		}
	}
	return false
}