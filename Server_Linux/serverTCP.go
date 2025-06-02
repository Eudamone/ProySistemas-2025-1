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
	// Canal de cierre para el socket
	exitCh := make(chan struct{})

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
		for {
			select {
			case msg := <-msgCh:
				_, err := writer.WriteString(msg)
				if err != nil {
					fmt.Println("Error al escribir en el socket:", err)
					close(exitCh)
					return
				}
				writer.Flush()
			case <-exitCh:
				return
			}
		}
	}()

	var n int
	for {
		response, _ := reader.ReadString('\n')
		if response == "" {
			continue
		}
		response = strings.TrimRight(response, "\n")
		itemsResponse := strings.Split(response, ",")
		if len(itemsResponse) < 3 || itemsResponse[0] == "" || itemsResponse[1] == "" || itemsResponse[2] == "" {
			fmt.Printf("Parametro n:%s,usuario:%s,pass:%s", itemsResponse[0], itemsResponse[1], itemsResponse[2])
			msgCh <- "Parametros de conexion no validos.\n"
			continue
		}
		nS, user, passw := itemsResponse[0], itemsResponse[1], itemsResponse[2]
		fmt.Println("Validando usuario y contraseña...")
		if ValidateUser(user, passw) {
			fmt.Printf("Usuario %s validado correctamente.\n", user)
			msgCh <- "Autenticación exitosa.\n"
			n, _ = strconv.Atoi(nS)
			fmt.Println("Valor de n recibido del cliente:", n)
			break
		} else {
			fmt.Printf("Usuario %s no validado correctamente.\n", user)
			msgCh <- "Usuario o contraseña incorrectos.\n"
			// No cerrar el socket aquí, esperar a que el cliente reintente
		}
	}

	// Manejo del cliente en goroutines
	// Se crean dos goroutines, una para recibir mensajes y otra para enviar reportes
	go recCommand(reader, msgCh, exitCh)
	go sendReports(n, msgCh, exitCh)

	// Mantener el servidor activo hasta que se cierre
	<-exitCh
	fmt.Println("Cerrando conexión con el cliente...")
	socketS.Close()
}

func recCommand(recBuffer *bufio.Reader, msgCh chan string, exitCh chan struct{}) {
	for {
		command, err := recBuffer.ReadString('\n')
		if err != nil {
			fmt.Println("Cliente desconectado (lectura):", err)
			close(exitCh)
			return
		}
		fmt.Println("Mensaje recibido del cliente:", command)

		command = strings.TrimRight(command, "\n")
		Scommand := strings.Split(command, ":")
		shell := exec.Command(Scommand[0], Scommand[1:]...)
		resCommand, _ := shell.CombinedOutput()

		rtaComando := string(resCommand) + "[FIN]\n"

		msgCh <- rtaComando
	}
}

func sendReports(n int, msgCh chan string, exitCh chan struct{}) {
	x := 0
	ticker := time.NewTicker(time.Duration(n) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-exitCh:
			fmt.Println("Deteniendo envío de reportes...")
			return
		case <-ticker.C:
			report := fmt.Sprintf("REPORT:CPU:%d,PRC:%d,RAM:%d,DD:%d\n", x+5, x+20, x+3, x+10)
			msgCh <- report
			fmt.Println("Enviando reporte al cliente:", report)
			x++
		}
		time.Sleep(time.Duration(n) * time.Second)
		//Procesador - Procesos - Memoria - Disco
		report := fmt.Sprintf("REPORT:CPU:%d,PRC:%d,RAM:%d,DD:%d\n", x+5, x+20, x+3, x+10)
		msgCh <- report
		fmt.Println("Enviando reporte al cliente:", report)
		x++
	}
}

func ValidateUser(user, passw string) bool {
	archivo, err := os.ReadFile("Server_Linux/users.txt")
	fmt.Printf("user:%s, passw:%s\n", user, passw)
	//dir,_ := os.Getwd()
	//fmt.Println("Directorio actual: ",dir)
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
		if user == strings.TrimSpace(credentials[0]) && hexapassUser == strings.TrimSpace(credentials[1]) {
			return true
		}
	}
	return false
}
