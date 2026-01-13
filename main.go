package main

import (
	"log"
	"sort"
	"strings"

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
	w.Resize(fyne.NewSize(400, 600))

	// 1. Récupération des artistes depuis l'API
	log.Println("Téléchargement des artistes...")
	artists, err := api.FetchArtists()
	if err != nil {
		w.SetContent(widget.NewLabel("Erreur API: " + err.Error()))
		w.ShowAndRun()
		return
	}

	// 2. Tri alphabétique initial (A → Z)
	sort.Slice(artists, func(i, j int) bool {
		return strings.ToLower(artists[i].Name) < strings.ToLower(artists[j].Name)
	})

	// Liste filtrée (celle réellement affichée)
	filtered := make([]api.Artist, len(artists))
	copy(filtered, artists)

	var showList func() // Déclaration pour pouvoir rappeler la liste

	// 3. Page de détails d'un artiste
	showDetails := func(artist api.Artist) {

		// Titre avec nom de l'artiste
		header := widget.NewLabelWithStyle(
			artist.Name,
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true},
		)

		// Labels affichés avant chargement réel
		locLabel := widget.NewLabel("Chargement des localisations...")
		dateLabel := widget.NewLabel("Chargement des dates...")
		relLabel := widget.NewLabel("Chargement des relations...")

		locLabel.Wrapping = fyne.TextWrapWord
		dateLabel.Wrapping = fyne.TextWrapWord
		relLabel.Wrapping = fyne.TextWrapWord

		// Chargement asynchrone pour ne pas bloquer l'interface
		go func() { locLabel.SetText("Localisations:\n" + api.FetchLocation(artist.LocationsURL)) }()
		go func() { dateLabel.SetText("Dates:\n" + api.FetchDates(artist.ConcertDates)) }()
		go func() { relLabel.SetText("Relations:\n" + api.FetchRelations(artist.RelationsURL)) }()

		// Bouton retour
		backBtn := widget.NewButton("Retour", func() { showList() })

		// Mise en page verticale
		content := container.NewVBox(
			header,
			widget.NewSeparator(),
			locLabel,
			dateLabel,
			relLabel,
			layout.NewSpacer(),
			backBtn,
		)

		w.SetContent(container.NewVScroll(content))
	}

	// 4. Liste principale des artistes
	var list *widget.List

	list = widget.NewList(
		// Nombre d'éléments affichés = taille de la liste filtrée
		func() int { return len(filtered) },

		// Template d'un élément (un bouton)
		func() fyne.CanvasObject {
			btn := widget.NewButton("Nom", nil)
			btn.Alignment = widget.ButtonAlignLeading
			return btn
		},

		// Remplissage d'un élément avec les données réelles
		func(i widget.ListItemID, o fyne.CanvasObject) {
			artist := filtered[i]
			button := o.(*widget.Button)
			button.SetText(artist.Name)

			// Clic → page de détails
			button.OnTapped = func() { showDetails(artist) }
		},
	)

	// 5. Barre de recherche dynamique (façon Google)
	search := widget.NewEntry()
	search.SetPlaceHolder("Rechercher un artiste...")

	search.OnChanged = func(text string) {

		// On convertit en minuscule pour comparer proprement les caractères ASCII
		text = strings.ToLower(text)

		// On vide la liste filtrée
		filtered = filtered[:0]

		// On parcourt la liste complète
		for _, a := range artists {
			name := strings.ToLower(a.Name)

			// Recherche façon Google :
			// si le nom contient ce que l'utilisateur tape → on garde
			if strings.Contains(name, text) {
				filtered = append(filtered, a)
			}
		}

		// TRI ALPHABÉTIQUE DES RÉSULTATS FILTRÉS
		sort.Slice(filtered, func(i, j int) bool {
			return strings.ToLower(filtered[i].Name) < strings.ToLower(filtered[j].Name)
		})

		// Mise à jour visuelle
		list.Refresh()
	}
	// 6. Affichage principal (liste + recherche)
	showList = func() {

		title := widget.NewLabelWithStyle(
			"Liste des Artistes",
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true},
		)

		// Border layout :
		// Haut = titre + barre de recherche
		// Centre = liste filtrée
		content := container.NewBorder(
			container.NewVBox(title, search),
			nil, nil, nil,
			list,
		)

		w.SetContent(content)
	}

	// Affichage initial
	showList()
	w.ShowAndRun()
}