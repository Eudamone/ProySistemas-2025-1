package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// closeOnce es una variable para asegurar que el socket se cierra una sola vez
var closeOnce sync.Once

func main() {
	proy := app.New()
	window := proy.NewWindow("Cliente SO")
	window.Resize(fyne.NewSize(400, 400))

	// Canales para la sincronización de la interfaz
	commandCh := make(chan string)
	reportCh := make(chan string)
	// Socket para la conexión con el servidor
	var socketC net.Conn
	var err error
	var writer *bufio.Writer

	// Cerrar el socket al cerrar la ventana
	window.SetOnClosed(func() {
		if socketC != nil {
			socketC.Close()
		}
	})

	// Pantalla de parametros
	// Entrada para valor de IP
	ipEntry := widget.NewEntry()
	ipEntry.SetPlaceHolder("IP del servidor")
	regexIP := `^(25[0-5]|2[0-4][0-9]|[0-1]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[0-1]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[0-1]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[0-1]?[0-9][0-9]?)$`
	ipEntry.Validator = validation.NewRegexp(regexIP, "Dirección IP invalida")
	// Entrada para valor de puerto
	portEntry := widget.NewEntry()
	portEntry.SetPlaceHolder("Puerto del servidor")
	regexPort := `^([0-9]{1,4}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])$`
	portEntry.Validator = validation.NewRegexp(regexPort, "Puerto invalido")
	// Entrada para valor del parametro n para tiempo de actualización de reportes
	nEntry := widget.NewEntry()
	nEntry.SetPlaceHolder("Tiempo de actualización reportes")
	regexN := `^([0-9]?[0-9][0-9]?)$`
	nEntry.Validator = validation.NewRegexp(regexN, "Parametro n invalido")
	// Entrada para usuario
	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder("Ingrese su usuario")
	regexUser := `^[a-zA-Z0-9_]{3,20}$`
	userEntry.Validator = validation.NewRegexp(regexUser, "Usuario invalido (3-20 caracteres, alfanuméricos o guiones bajos)")
	// Entrada para contraseña
	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder("Ingrese su contraseña")
	regexPass := `^[a-zA-Z0-9_!@#$%^&*()]{4,20}$`
	passEntry.Validator = validation.NewRegexp(regexPass, "Contraseña invalida (4-20 caracteres, alfanuméricos o símbolos permitidos)")

	// Botón para realizar la conexión
	connectBtt := widget.NewButton("Conectar", nil)
	validateBtt := widget.NewButton("Validar usuario", nil)

	// Formulario de conexión
	connForm := container.NewVBox(
		widget.NewLabel("Bienvenido"),
		ipEntry,
		portEntry,
		nEntry,
		userEntry,
		passEntry,
		connectBtt,
	)

	// Dialog de confirmación para cerrar la aplicación
	exitDialog := dialog.NewCustom("Notificación:", "Aceptar", widget.NewLabel("Número máximo de intentos alcanzado.\nCerrando aplicación."), window)
	exitDialog.SetOnClosed(func() {
		window.Close()
	})

	//Elementos de reporte
	cpuLabel := widget.NewLabel("CPU: -")
	prcLabel := widget.NewLabel("Procesos: -")
	ramLabel := widget.NewLabel("RAM: -")
	diskLabel := widget.NewLabel("Disco: -")

	// Logica de conexión al servidor
	connectBtt.OnTapped = func() {
		//Validar si no hay errores en las entradas de ip y puerto
		if (ipEntry.Validate() != nil) || (portEntry.Validate() != nil) || (nEntry.Validate() != nil) || (userEntry.Validate() != nil) || (passEntry.Validate() != nil) {
			dialog.ShowError(errors.New("por favor, corrija los errores en las entradas"), window)
			return
		} else if connectBtt.Text == "Validar usuario" {
			// Si el botón es "Validar usuario", se valida el usuario y contraseña
			if userEntry.Validate() != nil || passEntry.Validate() != nil {
				dialog.ShowError(errors.New("por favor, corrija los errores en las entradas"), window)
				return
			}
			connectBtt.Disable()
			// Si las entradas son válidas, se procede a validar el usuario y contraseña
			writer.WriteString(nEntry.Text + "," + userEntry.Text + "," + passEntry.Text + "\n")
			writer.Flush()
			response := <-commandCh
			fmt.Println("Respuesta del servidor:", response)
			response = strings.TrimSpace(response)
			if response == "Número máximo de intentos alcanzado. Cerrando conexión." {
				exitDialog.Show()
			} else if response == "" || response == "Usuario o contraseña incorrectos." || response == "Parametros de conexion no validos." {
				dialog.ShowError(errors.New("usuario y/o contraseña invalidos"), window)
				// Se habilitan las entradas para que el usuario pueda corregir
				userEntry.SetText("")
				passEntry.SetText("")
				userEntry.Enable()
				passEntry.Enable()
				connectBtt.Enable()
				return
			} else {
				fmt.Println("Respuesta del servidor -> ", response)
				dialog.ShowInformation("Validación exitosa", "Usuario y contraseña validados correctamente.", window)
				for _, w := range []fyne.Disableable{userEntry, passEntry} {
					w.Disable()
				}
				// Si la conexión es exitosa, se procede a mostrar la interfaz principal
				window.Resize(fyne.NewSize(800, 600))
				window.SetContent(MainInterface(window, &socketC, cpuLabel, prcLabel, ramLabel, diskLabel, commandCh))

				go updateReports([]*widget.Label{cpuLabel, prcLabel, ramLabel, diskLabel}, reportCh)

				window.SetTitle("Cliente SO - " + ipEntry.Text + ":" + portEntry.Text)
				//Para evitar que avance en el codigo y se intente conectar al servidor
				return
			}

		}

		elements := []fyne.Disableable{
			ipEntry,
			portEntry,
			nEntry,
			connectBtt,
		}

		for _, w := range elements {
			w.Disable()
		}

		connectBtt.SetText("Espere....")
		connectBtt.Disable()

		go func() {
			socketC, err = net.Dial("tcp", ipEntry.Text+":"+portEntry.Text)
			fyne.Do(func() {
				if err != nil {
					dialog.ShowError(err, window)
					for _, w := range elements {
						w.Enable()
					}
					connectBtt.SetText("Conectar")
					return
				}

				// Se envían parametros de conexión al servidor
				writer = bufio.NewWriter(socketC)

				go interfaceSocket(&socketC, commandCh, reportCh)

				writer.WriteString(nEntry.Text + "," + userEntry.Text + "," + passEntry.Text + "\n")
				writer.Flush()

				response := <-commandCh
				response = strings.TrimSpace(response)
				if response == "Número máximo de intentos alcanzado. Cerrando conexión." {
					exitDialog.Show()
				} else if response == "" || response == "Usuario o contraseña incorrectos." || response == "Parametros de conexion no validos." {
					dialog.ShowError(errors.New("usuario y/o contraseña invalidos"), window)
					// Se habilitan las entradas para que el usuario pueda corregir
					userEntry.SetText("")
					passEntry.SetText("")
					userEntry.Enable()
					passEntry.Enable()
					connectBtt.SetText("Validar usuario")
					connectBtt.Enable()
					return
				} else {
					fmt.Println("Respuesta del servidor -> ", response)
					dialog.ShowInformation("Validación exitosa", "Usuario y contraseña validados correctamente.", window)
				}

				// Si la conexión es exitosa, se procede a mostrar la interfaz principal
				window.Resize(fyne.NewSize(800, 600))
				window.SetContent(MainInterface(window, &socketC, cpuLabel, prcLabel, ramLabel, diskLabel, commandCh))

				go updateReports([]*widget.Label{cpuLabel, prcLabel, ramLabel, diskLabel}, reportCh)

				window.SetTitle("Cliente SO - " + ipEntry.Text + ":" + portEntry.Text)
			})

		}()
	}

	validateBtt.OnTapped = func() {
		if (userEntry.Validate() != nil) || (passEntry.Validate() != nil) {
			dialog.ShowError(errors.New("por favor, corrija los errores en las entradas"), window)
			return
		}

	}

	window.SetContent(connForm)
	window.ShowAndRun()
}

func MainInterface(w fyne.Window, socketC *net.Conn, cpuLabel, prcLabel, ramLabel, diskLabel *widget.Label, commandCh chan string) fyne.CanvasObject {
	// Richtext para mostrar el reporte de comandos sistema
	richOutput := widget.NewRichText()
	richOutput.Wrapping = fyne.TextWrapWord //Propiedad para bloquear el contenido en sentido horizontal
	richOutput.Segments = []widget.RichTextSegment{}
	// Contenedor para el reporte de comandos de scroll
	scrollOutput := container.NewScroll(richOutput)

	input := widget.NewEntry()
	input.SetPlaceHolder("Ingrese un comando")
	input.OnSubmitted = func(text string) {
		if text == "" {
			return
		} else if text == "bye" {
			// Cerrar la aplicación
			if socketC != nil {
				(*socketC).Close()
			}
			w.Close()
			return
		} else if text == "cls" {
			richOutput.Segments = []widget.RichTextSegment{}
			richOutput.Refresh()
			input.SetText("")
			return
		}

		//Enviar el comando al servidor
		sendComand(socketC, text, richOutput,scrollOutput,commandCh)
		input.SetText("")
	}

	terminalBox := container.NewBorder(nil, input, nil, nil, scrollOutput)
	// Layout Vertical
	reportBox := container.NewVBox(widget.NewLabel("Reporte del sistema:"), cpuLabel, prcLabel, ramLabel, diskLabel)
	//Dividir pantalla
	content := container.NewHSplit(terminalBox, reportBox)
	content.Offset = 0.7
	return content
}

func interfaceSocket(socketC *net.Conn, commandCh, reportCh chan string) {
	reader := bufio.NewReader(*socketC)
	var bufferOutput string
	for {
		linesOutput, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error al leer del socket:", err)
			closeOnce.Do(func() {
				close(commandCh)
				close(reportCh)
			})
			return
		}

		if strings.HasPrefix(linesOutput, "REPORT:") {
			// Si es un reporte se envia directamente al canal de reportes
			reporte := strings.TrimSpace(linesOutput[len("REPORT:"):])
			reportCh <- reporte
			continue
		}

		bufferOutput += linesOutput

		if strings.Contains(bufferOutput, "[FIN]") {
			msg := strings.ReplaceAll(bufferOutput, "[FIN]", "")
			msg = strings.TrimSpace(msg)
			bufferOutput = "" // Reiniciar el buffer para la siguiente lectura

			commandCh <- msg
		}

	}
}

func sendComand(socketC *net.Conn, command string,richOutput *widget.RichText,scroll *container.Scroll,commandCh chan string) {
	//Enviar comando al servidor
	writer := bufio.NewWriter(*socketC)
	writer.WriteString(command + "\n")
	writer.Flush()
	//Esperar la respuesta del canal
	response := <-commandCh

	// Agregar el comando y la respuesta al richtext
	richOutput.Segments = append(richOutput.Segments,
		&widget.TextSegment{Text: "> " + command + "\n",Style: widget.RichTextStyle{ColorName: "white",TextStyle: fyne.TextStyle{Bold: true}}},
		&widget.TextSegment{Text: response + "\n", Style: widget.RichTextStyle{TextStyle: fyne.TextStyle{Monospace: true}}},
	)
	richOutput.Refresh()
	scroll.ScrollToBottom()
}

func updateReports(elementos []*widget.Label, reportCh chan string) {
	for report := range reportCh {
		//report = strings.TrimSpace(report)
		valores := strings.Split(report, ";")
		fyne.Do(func() {
			for i, val := range valores {
				elementos[i].SetText(val)
			}
		})
	}
}
