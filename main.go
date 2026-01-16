package main

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	api "groupie/models"
)

func noFilterSelected(filters ...*widget.Check) bool {
	for _, f := range filters {
		if f.Checked {
			return false
		}
	}
	return true
}

func main() {

	groupie := app.New()
	w := groupie.NewWindow("Groupie Tracker")
	w.Resize(fyne.NewSize(400, 600))

	// --- 1. Fetch API ---
	log.Println("Téléchargement des artistes...")
	artists, err := api.FetchArtists()
	if err != nil {
		w.SetContent(widget.NewLabel("Erreur API: " + err.Error()))
		w.ShowAndRun()
		return
	}

	// Tri initial
	sort.Slice(artists, func(i, j int) bool {
		return strings.ToLower(artists[i].Name) < strings.ToLower(artists[j].Name)
	})

	// Liste filtrée
	filtered := make([]api.Artist, len(artists))
	copy(filtered, artists)

	var showList func()

	// --- 2. Page détails ---
	showDetails := func(artist api.Artist) {

		header := widget.NewLabelWithStyle(
			artist.Name,
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true},
		)

		var artistImage *canvas.Image
		if artist.Image != "" {
			imageURI := storage.NewURI(artist.Image)
			artistImage = canvas.NewImageFromURI(imageURI)
			artistImage.FillMode = canvas.ImageFillContain
			artistImage.SetMinSize(fyne.NewSize(300, 300))
		}

		firstAlbumLabel := widget.NewLabel("Premier Album: " + artist.FirstAlbum)
		membersLabel := widget.NewLabel("Membres: " + strings.Join(artist.Members, ", "))
		creationLabel := widget.NewLabel(fmt.Sprintf("Année de Création: %d", artist.CreationDate))

		locLabel := widget.NewLabel("Chargement des localisations...")
		dateLabel := widget.NewLabel("Chargement des dates...")
		relLabel := widget.NewLabel("Chargement des relations...")

		go func() { locLabel.SetText("Localisations:\n" + api.FetchLocation(artist.LocationsURL)) }()
		go func() { dateLabel.SetText("Dates:\n" + api.FetchDates(artist.ConcertDates)) }()
		go func() { relLabel.SetText("Relations:\n" + api.FetchRelations(artist.RelationsURL)) }()

		backBtn := widget.NewButton("Retour", func() { showList() })

		var content *fyne.Container
		if artistImage != nil {
			content = container.NewVBox(
				header,
				widget.NewSeparator(),
				artistImage,
				firstAlbumLabel,
				membersLabel,
				creationLabel,
				widget.NewSeparator(),
				locLabel,
				dateLabel,
				relLabel,
				layout.NewSpacer(),
				backBtn,
			)
		} else {
			content = container.NewVBox(
				header,
				widget.NewSeparator(),
				firstAlbumLabel,
				membersLabel,
				creationLabel,
				widget.NewSeparator(),
				locLabel,
				dateLabel,
				relLabel,
				layout.NewSpacer(),
				backBtn,
			)
		}

		w.SetContent(container.NewVScroll(content))
	}

	// --- 3. Liste principale ---
	var list *widget.List

	list = widget.NewList(
		func() int { return len(filtered) },
		func() fyne.CanvasObject {
			btn := widget.NewButton("Nom", nil)
			btn.Alignment = widget.ButtonAlignLeading
			return btn
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			artist := filtered[i]
			button := o.(*widget.Button)
			button.SetText(artist.Name)
			button.OnTapped = func() { showDetails(artist) }
		},
	)

	// --- 4. Barre de recherche ---
	search := widget.NewEntry()
	search.SetPlaceHolder("Rechercher un artiste...")

	// On agrandit la barre via un container
	searchContainer := container.NewGridWrap(fyne.NewSize(260, 40), search)

	// --- 5. Bouton Filtres ---
	filterBtn := widget.NewButton("Filtres", nil)

	filterArtist := widget.NewCheck("Artistes", nil)
	filterMembers := widget.NewCheck("Membres", nil)
	filterLocations := widget.NewCheck("Lieux", nil)
	filterFirstAlbum := widget.NewCheck("Premier album", nil)
	filterCreation := widget.NewCheck("Création", nil)

	filterMenu := container.NewVBox(
		widget.NewLabel("Filtrer par :"),
		filterArtist,
		filterMembers,
		filterLocations,
		filterFirstAlbum,
		filterCreation,
	)
	filterMenu.Hide()

	filterBtn.OnTapped = func() {
		if filterMenu.Visible() {
			filterMenu.Hide()
		} else {
			filterMenu.Show()
		}
	}

	// --- 6. Recherche + filtres fonctionnels ---
	search.OnChanged = func(text string) {

		text = strings.ToLower(text)
		filtered = filtered[:0]

		for _, a := range artists {

			match := false
			noFilter := noFilterSelected(filterArtist, filterMembers, filterLocations, filterFirstAlbum, filterCreation)

			// ARTISTES
			if filterArtist.Checked || noFilter {
				if strings.Contains(strings.ToLower(a.Name), text) {
					match = true
				}
			}

			// MEMBRES
			if filterMembers.Checked || noFilter {
				for _, m := range a.Members {
					if strings.Contains(strings.ToLower(m), text) {
						match = true
					}
				}
			}

			// LIEUX
			if filterLocations.Checked || noFilter {
				loc := strings.ToLower(api.FetchLocation(a.LocationsURL))
				if strings.Contains(loc, text) {
					match = true
				}
			}

			// PREMIER ALBUM
			if filterFirstAlbum.Checked || noFilter {
				if strings.Contains(strings.ToLower(a.FirstAlbum), text) {
					match = true
				}
			}

			// DATE DE CREATION
			if filterCreation.Checked || noFilter {
				if strings.Contains(strings.ToLower(fmt.Sprint(a.CreationDate)), text) {
					match = true
				}
			}

			if match {
				filtered = append(filtered, a)
			}
		}

		sort.Slice(filtered, func(i, j int) bool {
			return strings.ToLower(filtered[i].Name) < strings.ToLower(filtered[j].Name)
		})

		list.Refresh()
	}

	// --- 7. Layout principal ---
	showList = func() {

		title := widget.NewLabelWithStyle(
			"Liste des Artistes",
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true},
		)

		// Search large à gauche, filtre à droite
		topBar := container.NewHBox(
			searchContainer,
			layout.NewSpacer(),
			filterBtn,
		)

		content := container.NewBorder(
			container.NewVBox(
				title,
				topBar,
				filterMenu,
			),
			nil, nil, nil,
			list,
		)

		w.SetContent(content)
	}

	showList()
	w.ShowAndRun()
}