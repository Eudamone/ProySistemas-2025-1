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
	socketInicial, err := net.Listen("tcp", "192.168.3.192:1705")
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
			msgCh <- "Parametros de conexion no validos.[FIN]\n"
			continue
		}
		nS, user, passw := itemsResponse[0], itemsResponse[1], itemsResponse[2]
		fmt.Println("Validando usuario y contraseña...")
		if ValidateUser(user, passw) {
			fmt.Printf("Usuario %s validado correctamente.\n", user)
			msgCh <- "Autenticación exitosa.[FIN]\n"
			n, _ = strconv.Atoi(nS)
			fmt.Println("Valor de n recibido del cliente:", n)
			break
		} else {
			fmt.Printf("Usuario %s no validado correctamente.\n", user)
			msgCh <- "Usuario o contraseña incorrectos.[FIN]\n"
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
	ticker := time.NewTicker(time.Duration(n) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-exitCh:
			fmt.Println("Deteniendo envío de reportes...")
			return
		case <-ticker.C:
			// CPU: porcentaje de uso promedio en 1 segundo
			cpuOut, _ := exec.Command("sh", "-c", `LC_NUMERIC=C top -bn1 | grep "Cpu(s)" | awk -F',' '{for(i=1;i<=NF;i++){if($i~"id"){split($i,a," "); print 100-a[1]}}}'`).Output()
			cpu := strings.TrimSpace(string(cpuOut))
			if cpu == "" {
				cpu = "-"
			}
			cpu = strings.ReplaceAll(cpu, ",", ".")

			// Procesos: cantidad de procesos en ejecución
			prcOut, _ := exec.Command("sh", "-c", `ps -e --no-headers | wc -l`).Output()
			prc := strings.TrimSpace(string(prcOut))
			if prc == "" {
				prc = "-"
			}

			// RAM: porcentaje de uso de memoria
			ramOut, _ := exec.Command("sh", "-c", `LC_NUMERIC=C free | grep Mem | awk '{printf("%.2f", $3/$2 * 100)}'`).Output()
			ram := strings.TrimSpace(string(ramOut))
			if ram == "" {
				ram = "-"
			}

			// Disco: porcentaje de uso del disco raíz
			diskOut, _ := exec.Command("sh", "-c", `df / | tail -1 | awk '{print $5}' | tr -d '%'`).Output()
			disk := strings.TrimSpace(string(diskOut))
			if disk == "" {
				disk = "-"
			}

			// Formato esperado por el cliente: CPU,PRC,RAM,DD
			report := fmt.Sprintf("REPORT:CPU:%s,PRC:%s,RAM:%s,DD:%s\n", cpu, prc, ram, disk)
			msgCh <- report
			fmt.Println("Enviando reporte al cliente:", report)
		}
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
