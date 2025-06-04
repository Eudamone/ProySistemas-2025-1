package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func main() {
	parameters := initParams()
	fmt.Printf("Creando servidor TCP en %s:%s\n", parameters[0], parameters[1])
	// IP a 0.0.0.0 para aceptar conexiones de cualquier IP
	socketInicial, err := net.Listen("tcp", parameters[0]+":"+parameters[1])
	if err != nil {
		fmt.Println("Error al abrir el puerto:", err)
		return
	}
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
	limit, _ := strconv.Atoi(parameters[2])
	for i := 0; i < limit; i++ {
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
		if ValidateUser(parameters[3], user, passw) {
			fmt.Printf("Usuario %s validado correctamente.\n", user)
			msgCh <- "Autenticación exitosa.[FIN]\n"
			n, _ = strconv.Atoi(nS)
			fmt.Println("Valor de n recibido del cliente:", n)
			break
		} else if i == limit-1 {
			fmt.Println("Número máximo de intentos alcanzado. Cerrando conexión.")
			msgCh <- "Número máximo de intentos alcanzado. Cerrando conexión.[FIN]\n"
			// Dar tiempo para que el cliente reciba el mensaje
			time.Sleep(3 * time.Second)
			// Cerrar el canal de mensajes y el socket
			close(exitCh)
			socketS.Close()
			return
		} else {
			fmt.Printf("Usuario %s no validado correctamente.\n", user)
			msgCh <- "Usuario o contraseña incorrectos.[FIN]\n"
			// No cerrar el socket aquí, esperar a que el cliente reintente
		}
	}

	shell := exec.Command("bash")
	shellIn, _ := shell.StdinPipe()
	shellOut, _ := shell.StdoutPipe()
	shellErr, _ := shell.StderrPipe()
	shell.Start()

	// Lectores persistentes para la salida estándar y de error del shell
	go func() {
		scanner := bufio.NewScanner(shellOut)
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				msgCh <- line + "\n"
			}
		}
	}()

	go func() {
		scanner := bufio.NewScanner(shellErr)
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				msgCh <- line + "\n"
			}
		}
	}()

	// Manejo del cliente en goroutines
	// Se crean dos goroutines, una para recibir mensajes y otra para enviar reportes
	go recCommand(reader, shellIn, msgCh, exitCh)
	go sendReports(n, msgCh, exitCh)

	// Mantener el servidor activo hasta que se cierre
	<-exitCh
	fmt.Println("Cerrando conexión con el cliente...")
	socketS.Close()
}

func recCommand(recBuffer *bufio.Reader, shellIn io.WriteCloser, msgCh chan string, exitCh chan struct{}) {
	for {
		command, err := recBuffer.ReadString('\n')
		if err != nil {
			fmt.Println("Cliente desconectado (lectura):", err)
			close(exitCh)
			return
		}
		command = strings.TrimSpace(command)
		if command == "" {
			continue
		}
		fmt.Println("Comando recibido del cliente:", command)
		_, err = shellIn.Write([]byte(command + "\necho \"[FIN]\"\n"))
		if err != nil {
			msgCh <- "Error al enviar comando a shell.\n[FIN]\n"
		}

		msgCh <- "[FIN]\n"
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
			report := fmt.Sprintf("REPORT:CPU:%s%%;PRC:%s;RAM:%s%%;DD:%s%%\n", cpu, prc, ram, disk)
			msgCh <- report
			fmt.Println("Enviando reporte al cliente:", report)
		}
	}
}

func ValidateUser(rute, user, passw string) bool {
	archivo, err := os.ReadFile("Server_Linux/" + rute)
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

func initParams() []string {
	file, err := os.ReadFile("Server_Linux/conf.txt")
	if err != nil {
		fmt.Println("Error al leer el archivo de configuración:", err)
		return []string{}
	}
	lines := strings.Split(string(file), "\n")
	var params []string
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}
		value := strings.TrimSpace(parts[1])
		params = append(params, value)
	}
	return params
}
