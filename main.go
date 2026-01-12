package fyne

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func main() {
	groupie := app.New()
	w := groupie.NewWindow("Groupie Tracker")

	w.SetContent(widget.NewLabel("Welcome to Groupie Tracker!"))
	w.ShowAndRun()
}
