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
	ipEntry := widget.NewEntry()
	ipEntry.SetPlaceHolder("IP del servidor")
	regexIP := `^(25[0-5]|2[0-4][0-9]|[0-1]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[0-1]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[0-1]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[0-1]?[0-9][0-9]?)$`
	ipEntry.Validator = validation.NewRegexp(regexIP,"Dirección IP invalida")
	portEntry := widget.NewEntry()
	portEntry.SetPlaceHolder("Puerto del servidor")
	regexPort := `^([0-9]{1,4}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])$`
	portEntry.Validator = validation.NewRegexp(regexPort,"Puerto invalido")

	connectBtt := widget.NewButton("Conectar",nil)

	connectBtt.OnTapped = func() {
		//Validar si no hay errores en las entradas de ip y puerto
		if (ipEntry.Validate() != nil) || (portEntry.Validate() != nil){
			return
		}

		connectBtt.SetText("Espere....")
		connectBtt.Disable()
		
		go func(){
			socketC,err := net.Dial("tcp",ipEntry.Text+":"+portEntry.Text)
			fyne.Do(func(){
				if err!= nil{
					dialog.ShowError(err,window)
					connectBtt.SetText("Conectar")
					connectBtt.Enable()
					return
				}
				defer socketC.Close()

				window.Resize(fyne.NewSize(800,600))
				window.SetContent(MainInterface(window))
			})
			
		}()	
	}

	connForm := container.NewVBox(
		widget.NewLabel("Bienvenido"),
		ipEntry,
		portEntry,
		connectBtt,
	)
	

	window.SetContent(connForm)
	window.ShowAndRun()
}

func MainInterface(w fyne.Window) fyne.CanvasObject {
	// Entrada e historial de comandos
	output := widget.NewMultiLineEntry()
	output.SetPlaceHolder("Salida del servidor....")
	output.Disable() //Solo lectura
	output.Wrapping = fyne.TextWrapWord //Propiedad para bloquear el contenido en sentido horizontal
	output.TextStyle.Bold = true
	output.TextStyle.Monospace = true

	//scroll := container.NewScroll(output)

	input := widget.NewEntry()
	input.SetPlaceHolder("Ingrese un comando")
	input.OnSubmitted = func(text string) {
		if text == "" {
			return
		}else if text == "cls"{
			output.SetText("")
			input.SetText("")
			return
		}

		// Aquí luego enviarías el comando al servidor
		simulatedResponse := fmt.Sprintf("Comando ejecutado: %s\nResultado simulado\n", text)
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