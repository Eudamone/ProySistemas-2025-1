package main

import (
	"fmt"
	"net"

	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/widget"
)

func main() {
	proy := app.New()
	window := proy.NewWindow("Cliente SO")
	window.Resize(fyne.NewSize(400,400))

	// Pantalla de parametros
	// Entrada para valor de IP 
	ipEntry := widget.NewEntry()
	ipEntry.SetPlaceHolder("IP del servidor")
	regexIP := `^(25[0-5]|2[0-4][0-9]|[0-1]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[0-1]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[0-1]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[0-1]?[0-9][0-9]?)$`
	ipEntry.Validator = validation.NewRegexp(regexIP,"Dirección IP invalida")
	// Entrada para valor de puerto
	portEntry := widget.NewEntry()
	portEntry.SetPlaceHolder("Puerto del servidor")
	regexPort := `^([0-9]{1,4}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])$`
	portEntry.Validator = validation.NewRegexp(regexPort,"Puerto invalido")
	// Entrada para valor del parametro n para tiempo de actualización de reportes
	nEntry := widget.NewEntry()
	nEntry.SetPlaceHolder("Tiempo de actualización reportes")
	regexN := `^([0-9]?[0-9][0-9]?)$`
	nEntry.Validator = validation.NewRegexp(regexN,"Parametro n invalido")

	// Botón para realizar la conexión
	connectBtt := widget.NewButton("Conectar",nil)

	// Logica de conexión al servidor
	connectBtt.OnTapped = func() {
		//Validar si no hay errores en las entradas de ip y puerto
		if (ipEntry.Validate() != nil) || (portEntry.Validate() != nil) || (nEntry.Validate() != nil){
			return
		}

		elements := []fyne.Disableable{
			ipEntry,
			portEntry,
			nEntry,
			connectBtt,
		}
		
		for _,w := range elements {
			w.Disable()
		}

		connectBtt.SetText("Espere....")
		connectBtt.Disable()
		
		
		go func(){
			socketC,err := net.Dial("tcp",ipEntry.Text+":"+portEntry.Text)
			fyne.Do(func(){
				if err!= nil{
					dialog.ShowError(err,window)
					for _,w := range elements {
						w.Enable()
					}
					connectBtt.SetText("Conectar")
					return
				}

				// Se envía el parámetro n al servidor
				socketC.Write([]byte(nEntry.Text))

				// Si la conexión es exitosa, se procede a mostrar la interfaz principal
				window.Resize(fyne.NewSize(800,600))
				window.SetContent(MainInterface(window,&socketC))

				window.SetTitle("Cliente SO - " + ipEntry.Text + ":" + portEntry.Text)
				window.SetOnClosed(func() {
					if socketC != nil {
						socketC.Close()
					}
				})
			})
			
		}()	
	}

	connForm := container.NewVBox(
		widget.NewLabel("Bienvenido"),
		ipEntry,
		portEntry,
		nEntry,
		connectBtt,
	)

	window.SetContent(connForm)
	window.ShowAndRun()
}

func MainInterface(w fyne.Window,socketC *net.Conn) fyne.CanvasObject {
	// Entrada e historial de comandos
	output := widget.NewMultiLineEntry()
	output.SetPlaceHolder("Salida del servidor....")
	output.Disable() //Solo lectura
	output.Wrapping = fyne.TextWrapWord //Propiedad para bloquear el contenido en sentido horizontal
	output.TextStyle.Bold = true
	output.TextStyle.Monospace = true


	input := widget.NewEntry()

	input.SetPlaceHolder("Ingrese un comando")
	input.OnSubmitted = func(text string) {
		if text == "" {
			return
		}else if text == "bye" {
			// Cerrar la aplicación
			if socketC != nil {
				(*socketC).Close()
			}
			w.Close()
			return
		}else if text == "cls"{
			output.SetText("")
			input.SetText("")
			return
		}

		// Enviar el comando al servidor
		(*socketC).Write([]byte(text))

		// Recibir la respuesta del servidor
		buffer := make([]byte, 1024)
		responseS,err :=  (*socketC).Read(buffer)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		simulatedResponse := string(buffer[:responseS])
		fmt.Println("Respuesta del servidor:", simulatedResponse)
		output.SetText(output.Text + "\n> " + text + "\n" + simulatedResponse)

		//Forzar el scroll a bajar
		focussItem := w.Canvas().Focused()
		if focussItem == nil || focussItem != output{
			output.CursorRow = len(output.Text)-1
		}

		input.SetText("")
	}

	terminalBox := container.NewBorder(nil, input, nil, nil, output)

	// Panel de reportes
	cpuLabel := widget.NewLabel("CPU: 0%")
	ramLabel := widget.NewLabel("RAM: 5%")

	// Layout Vertical
	reportBox := container.NewVBox(widget.NewLabel("Reporte del sistema:"), cpuLabel, ramLabel)

	//Dividir pantalla
	content := container.NewHSplit(terminalBox, reportBox)
	content.Offset = 0.7
	return content
}

