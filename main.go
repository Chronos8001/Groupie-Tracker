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
	// Création de l'application Fyne
	groupie := app.New()
	w := groupie.NewWindow("Groupie Tracker")
	w.Resize(fyne.NewSize(400, 600)) // Taille de la fenêtre

	// 1. Récupération des données depuis l'API Groupie Tracker
	log.Println("Téléchargement des artistes...")
	artists, err := api.FetchArtists()
	if err != nil {
		// Si l'API ne répond pas, on affiche un message d'erreur dans la fenêtre
		log.Println("Erreur API:", err)
		w.SetContent(widget.NewLabel("Impossible de charger les données: " + err.Error()))
		w.ShowAndRun()
		return
	}

	// Variable déclarée ici pour pouvoir rappeler la liste depuis la page de détails
	var showList func()

	// 2. Fonction d'affichage des détails d'un artiste
	showDetails := func(artist api.Artist) {

		// Titre avec le nom de l'artiste
		header := widget.NewLabelWithStyle(
			artist.Name,
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true},
		)

		// Labels affichés avant le chargement réel des données
		locLabel := widget.NewLabel("Chargement des localisations...")
		locLabel.Wrapping = fyne.TextWrapWord

		dateLabel := widget.NewLabel("Chargement des dates...")
		dateLabel.Wrapping = fyne.TextWrapWord

		relLabel := widget.NewLabel("Chargement des relations...")
		relLabel.Wrapping = fyne.TextWrapWord

		// Chargement des données en arrière-plan (goroutines)
		// Cela évite de bloquer l'interface graphique

		go func() {
			data := api.FetchLocation(artist.LocationsURL)
			locLabel.SetText("Localisations :\n" + data)
		}()

		go func() {
			data := api.FetchDates(artist.ConcertDates)
			dateLabel.SetText("Dates :\n" + data)
		}()

		go func() {
			data := api.FetchRelations(artist.RelationsURL)
			relLabel.SetText("Relations :\n" + data)
		}()

		// Bouton pour revenir à la liste principale
		backBtn := widget.NewButton("Retour à la liste", func() {
			showList()
		})

		// Mise en page verticale des informations
		detailsContent := container.NewVBox(
			header,
			widget.NewSeparator(),
			locLabel,
			dateLabel,
			relLabel,
			layout.NewSpacer(), // Espace flexible pour pousser le bouton vers le bas
			backBtn,
		)

		// Ajout d'un scroll pour éviter que le texte dépasse l'écran
		scroll := container.NewVScroll(detailsContent)
		w.SetContent(scroll)
	}

	// 3. Création de la liste principale des artistes
	list := widget.NewList(
		// Nombre d'éléments dans la liste
		func() int {
			return len(artists)
		},

		// Template d'un élément de liste (un bouton ici)
		func() fyne.CanvasObject {
			btn := widget.NewButton("Nom", nil)
			btn.Alignment = widget.ButtonAlignLeading // Alignement à gauche
			return btn
		},

		// Remplissage de chaque élément avec les données réelles
		func(i widget.ListItemID, o fyne.CanvasObject) {
			artist := artists[i]
			button := o.(*widget.Button)
			button.SetText(artist.Name)

			// Lorsqu'on clique sur un artiste → afficher ses détails
			button.OnTapped = func() {
				showDetails(artist)
			}
		},
	)

	// Fonction pour afficher la vue principale (liste des artistes)
	showList = func() {
		title := widget.NewLabelWithStyle(
			"Liste des Artistes",
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true},
		)

		// Border layout : titre en haut, liste au centre
		content := container.NewBorder(title, nil, nil, nil, list)
		w.SetContent(content)
	}

	// Affichage initial
	showList()
	w.ShowAndRun()
}