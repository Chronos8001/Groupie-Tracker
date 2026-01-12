package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	api "groupie/models"
)

func main() {
	groupie := app.New()
	w := groupie.NewWindow("Groupie Tracker")
	w.Resize(fyne.NewSize(400, 600))

	// 1. Récupération des données
	log.Println("Téléchargement des artistes...")
	artists, err := api.FetchArtists()
	if err != nil {
		log.Println("Erreur API:", err)
		w.SetContent(widget.NewLabel("Impossible de charger les données: " + err.Error()))
		w.ShowAndRun()
		return
	}

	var showList func() // Déclaration préalable pour pouvoir rappeler la liste depuis le détail

	// 2. Fonction pour afficher les détails d'un artiste
	showDetails := func(artist api.Artist) {
		// Création des labels avec les informations.
		// Note: Si vos champs dans la struct Artist sont différents, adaptez les noms ici.
		header := widget.NewLabelWithStyle(artist.Name, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

		// Création de labels "en chargement"
		locLabel := widget.NewLabel("Chargement des localisations...")
		locLabel.Wrapping = fyne.TextWrapWord

		dateLabel := widget.NewLabel("Chargement des dates...")
		dateLabel.Wrapping = fyne.TextWrapWord

		relLabel := widget.NewLabel("Chargement des relations...")
		relLabel.Wrapping = fyne.TextWrapWord

		// On lance les récupérations de données dans des goroutines pour ne pas bloquer l'interface
		go func() {
			data := api.GetFromURL(artist.LocationsURL)
			locLabel.SetText("Localisations :\n" + data)
		}()
		go func() {
			data := api.GetFromURL(artist.ConcertDates)
			dateLabel.SetText("Dates :\n" + data)
		}()
		go func() {
			data := api.GetFromURL(artist.RelationsURL)
			relLabel.SetText("Relations :\n" + data)
		}()

		// Bouton Retour
		backBtn := widget.NewButton("Retour à la liste", func() {
			showList()
		})

		// Layout des détails : On empile tout verticalement
		detailsContent := container.NewVBox(
			header,
			widget.NewSeparator(),
			locLabel,
			dateLabel,
			relLabel,
			layout.NewSpacer(), // Pousse le bouton vers le bas si on utilise un Border layout, sinon espace vide
			backBtn,
		)

		// On met le contenu dans un ScrollContainer au cas où il y a beaucoup de texte
		scroll := container.NewVScroll(detailsContent)
		w.SetContent(scroll)
	}

	// 3. Configuration de la liste principale
	list := widget.NewList(
		func() int {
			return len(artists)
		},
		func() fyne.CanvasObject {
			// On remplace NewLabel par NewButton pour le clic
			// On aligne le texte à gauche pour faire "style liste"
			btn := widget.NewButton("Nom", nil)
			btn.Alignment = widget.ButtonAlignLeading
			return btn
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			artist := artists[i]
			button := o.(*widget.Button)
			button.SetText(artist.Name)
			// Au clic, on appelle la fonction de détails
			button.OnTapped = func() {
				showDetails(artist)
			}
		},
	)

	// Fonction pour afficher la vue "Liste"
	showList = func() {
		title := widget.NewLabelWithStyle("Liste des Artistes", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		content := container.NewBorder(title, nil, nil, nil, list)
		w.SetContent(content)
	}

	// Afficher la liste au démarrage
	showList()
	w.ShowAndRun()
}
