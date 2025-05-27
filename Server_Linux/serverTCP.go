package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

func main() {
	proy := app.New()
	window := proy.NewWindow("Proyecto Sistemas Operativos")
	window.Resize(fyne.NewSize(800, 600))

	title := canvas.NewText("Proyecto Sistemas Operativos 2025-1", color.White)
	title.TextSize = 30
	text1 := canvas.NewText("Comment 1", color.White)
	text2 := canvas.NewText("Comment 2", color.White)

	contTitle := container.New(layout.NewHBoxLayout(), layout.NewSpacer(), title, layout.NewSpacer())
	contCode := container.New(layout.NewHBoxLayout(), text1, layout.NewSpacer(), text2)
	//contReport := container.New(layout.NewHBoxLayout(), layout.NewSpacer(), title, layout.NewSpacer())

	// Contenedor principal
	contMain := container.New(layout.NewVBoxLayout(), contTitle, contCode)

	window.SetContent(contMain)
	window.ShowAndRun()
}
