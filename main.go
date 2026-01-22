package main

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
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

// showMap affiche une carte avec les lieux de concerts de l'artiste
func showMap(artist api.Artist, w fyne.Window) {
	mapWindow := fyne.CurrentApp().NewWindow("Carte - " + artist.Name)
	mapWindow.Resize(fyne.NewSize(800, 600))

	// Récupérer les lieux
	locationsText := api.FetchLocation(artist.LocationsURL)
	locations := strings.Split(locationsText, "\n")

	if len(locations) == 0 || locationsText == "" {
		mapWindow.SetContent(widget.NewLabel("Aucun lieu de concert disponible"))
		mapWindow.Show()
		return
	}

	// Créer une liste des lieux avec leurs coordonnées
	var mapContent *fyne.Container
	locationsList := container.NewVBox()

	for _, loc := range locations {
		loc = strings.TrimSpace(loc)
		if loc == "" {
			continue
		}

		// Nettoyer le nom de la ville
		cleanLoc := strings.ReplaceAll(loc, "_", " ")
		cleanLoc = strings.ReplaceAll(cleanLoc, "-", ", ")

		locationLabel := widget.NewLabel("📍 " + cleanLoc)
		locationsList.Add(locationLabel)

		// Essayer de récupérer les coordonnées et afficher une mini-carte
		go func(cityName string, label *widget.Label) {
			lat, lon, err := GetCoordinates(cityName)
			if err == nil {
				label.SetText(cityName + fmt.Sprintf(" (%.4s, %.4s)", lat, lon))
			}
		}(cleanLoc, locationLabel)
	}

	// Prendre le premier lieu pour afficher une carte centrée
	if len(locations) > 0 {
		firstLoc := strings.TrimSpace(locations[0])
		cleanLoc := strings.ReplaceAll(firstLoc, "_", " ")
		cleanLoc = strings.ReplaceAll(cleanLoc, "-", ", ")

		go func() {
			lat, lon, err := GetCoordinates(cleanLoc)
			if err == nil {
				// Convertir lat/lon en float
				latF, _ := strconv.ParseFloat(lat, 64)
				lonF, _ := strconv.ParseFloat(lon, 64)

				// Récupérer la tuile de carte
				zoom := 4
				tileURL := GetOSMTileURL(latF, lonF, zoom)

				// Télécharger l'image
				resp, err := http.Get(tileURL)
				if err == nil {
					defer resp.Body.Close()
					mapImage := canvas.NewImageFromReader(resp.Body, "map")
					mapImage.FillMode = canvas.ImageFillContain
					mapImage.SetMinSize(fyne.NewSize(600, 400))

					mapContent = container.NewBorder(
						widget.NewLabelWithStyle("Lieux de concerts de "+artist.Name, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
						nil, nil, nil,
						container.NewHSplit(
							mapImage,
							container.NewVScroll(locationsList),
						),
					)
					mapWindow.SetContent(mapContent)
				}
			}
		}()
	}

	// Contenu par défaut pendant le chargement
	if mapContent == nil {
		mapContent = container.NewBorder(
			widget.NewLabelWithStyle("Lieux de concerts de "+artist.Name, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			nil, nil, nil,
			container.NewVScroll(locationsList),
		)
		mapWindow.SetContent(mapContent)
	}

	mapWindow.Show()
}

func main() {

	groupie := app.New()
	w := groupie.NewWindow("Groupie Tracker")
	w.Resize(fyne.NewSize(400, 600))

	// Variable pour savoir si on est sur la page de détails
	var isDetailsPage bool = false

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
		isDetailsPage = true

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

		mapBtn := widget.NewButton("Voir sur la carte", func() {
			showMap(artist, w)
		})

		backBtn := widget.NewButton("Retour (Échap)", func() { showList() })

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
				container.NewGridWithColumns(2, mapBtn, backBtn),
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
				container.NewGridWithColumns(2, mapBtn, backBtn),
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
	search.SetPlaceHolder("Rechercher un artiste... (Ctrl+F)")

	// On agrandit la barre via un container
	searchContainer := container.NewGridWrap(fyne.NewSize(260, 40), search)

	// --- 5. Bouton Filtres ---
	filterBtn := widget.NewButton("Filtres (Ctrl+M)", nil)

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
		isDetailsPage = false

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

	// --- 8. Raccourcis clavier ---
	w.Canvas().SetOnTypedKey(func(key *fyne.KeyEvent) {
		switch key.Name {
		case fyne.KeyEscape:
			// Échap: Retour à la liste principale
			if isDetailsPage {
				showList()
			}

		case fyne.KeyReturn, fyne.KeyEnter:
			// Entrée: Ouvrir le premier résultat de recherche
			if !isDetailsPage && len(filtered) > 0 {
				showDetails(filtered[0])
			}
		}
	})

	w.Canvas().AddShortcut(&fyne.ShortcutCopy{}, func(shortcut fyne.Shortcut) {})

	// Ctrl+F: Focus sur la recherche
	ctrlF := &fyne.KeyboardShortcut{
		KeyName:  fyne.KeyF,
		Modifier: fyne.KeyModifierControl,
	}
	w.Canvas().AddShortcut(ctrlF, func(shortcut fyne.Shortcut) {
		if !isDetailsPage {
			w.Canvas().Focus(search)
		}
	})

	// Ctrl+M: Afficher/masquer les filtres
	ctrlM := &fyne.KeyboardShortcut{
		KeyName:  fyne.KeyM,
		Modifier: fyne.KeyModifierControl,
	}
	w.Canvas().AddShortcut(ctrlM, func(shortcut fyne.Shortcut) {
		if !isDetailsPage {
			if filterMenu.Visible() {
				filterMenu.Hide()
			} else {
				filterMenu.Show()
			}
		}
	})

	// Ctrl+Q: Quitter l'application
	ctrlQ := &fyne.KeyboardShortcut{
		KeyName:  fyne.KeyQ,
		Modifier: fyne.KeyModifierControl,
	}
	w.Canvas().AddShortcut(ctrlQ, func(shortcut fyne.Shortcut) {
		groupie.Quit()
	})

	showList()
	w.ShowAndRun()
}
