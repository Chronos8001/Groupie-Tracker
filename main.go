package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	groupie := app.New()
	w := groupie.NewWindow("Groupie Tracker")
	green := color.NRGBA{R: 0, G: 180, B: 0, A: 255}
	text1 := canvas.NewText("Hello", green)
	text2 := canvas.NewText("There", green)
	text3 := canvas.NewText("General Kenobi", green)

	text1.Move(fyne.NewPos(20, 20))
	text2.Move(fyne.NewPos(20, 40))
	text3.Move(fyne.NewPos(20, 60))
	content := container.NewWithoutLayout(widget.NewLabel("Welcome to Groupie Tracker!"), text1, text2, text3)

	w.SetContent(content)
	w.ShowAndRun()
}
