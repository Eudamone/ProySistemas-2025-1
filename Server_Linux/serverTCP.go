package main

import (
	"fmt"
	"net"
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
	fmt.Println("Esperando conexiones....!")
	socketS, err := socketInicial.Accept()
	if err != nil {
		fmt.Println("Error al aceptar conexión:", err)
		return
	}
	fmt.Println("Cliente conectado [", socketS.RemoteAddr().String(), "]")
	fmt.Println("¡Conexión exitosa con el cliente!")
}
