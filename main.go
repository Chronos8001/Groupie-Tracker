package main

import (
	"fmt"
	"image/color"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
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

// createCard crée une card stylisée avec un fond et des bordures arrondies
func createCard(content fyne.CanvasObject) *fyne.Container {
	bg := canvas.NewRectangle(color.NRGBA{R: 40, G: 40, B: 50, A: 255})
	return container.NewStack(bg, container.NewPadded(content))
}

// createInfoLabel crée un label stylisé pour les informations
func createInfoLabel(icon, text string) *widget.RichText {
	rt := widget.NewRichText(
		&widget.TextSegment{Text: icon + " ", Style: widget.RichTextStyle{SizeName: theme.SizeNameHeadingText}},
		&widget.TextSegment{Text: text},
	)
	return rt
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

		locationLabel := widget.NewRichTextWithText(cleanLoc)
		locationCard := createCard(locationLabel)
		locationsList.Add(locationCard)

		// Essayer de récupérer les coordonnées et afficher une mini-carte
		go func(cityName string, label *widget.RichText) {
			lat, lon, err := GetCoordinates(cityName)
			if err == nil {
				label.ParseMarkdown(cityName + fmt.Sprintf(" (%.4s, %.4s)", lat, lon))
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
	groupie.Settings().SetTheme(theme.DarkTheme())

	w := groupie.NewWindow("Groupie Tracker")
	w.Resize(fyne.NewSize(90, 70))
	w.CenterOnScreen()

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

		// Header avec titre stylisé
		header := widget.NewRichTextFromMarkdown("# " + artist.Name)
		header.Wrapping = fyne.TextWrapWord

		// Image de l'artiste avec style
		var artistImage *canvas.Image
		if artist.Image != "" {
			imageURI := storage.NewURI(artist.Image)
			artistImage = canvas.NewImageFromURI(imageURI)
			artistImage.FillMode = canvas.ImageFillContain
			artistImage.SetMinSize(fyne.NewSize(350, 350))
		}

		// Informations principales
		firstAlbumLabel := createInfoLabel("", "Premier Album: "+artist.FirstAlbum)
		membersLabel := createInfoLabel("", "Membres: "+strings.Join(artist.Members, ", "))
		creationLabel := createInfoLabel("", fmt.Sprintf("Année de Création: %d", artist.CreationDate))

		// Card pour les infos principales
		infoCard := createCard(container.NewVBox(
			firstAlbumLabel,
			widget.NewSeparator(),
			membersLabel,
			widget.NewSeparator(),
			creationLabel,
		))

		// Labels pour les données asynchrones
		locLabel := widget.NewLabel("Chargement des localisations...")
		dateLabel := widget.NewLabel("Chargement des dates...")
		relLabel := widget.NewLabel("Chargement des relations...")

		locLabel.Wrapping = fyne.TextWrapWord
		dateLabel.Wrapping = fyne.TextWrapWord
		relLabel.Wrapping = fyne.TextWrapWord

		go func() {
			locData := api.FetchLocation(artist.LocationsURL)
			locLabel.SetText("Localisations:\n" + locData)
		}()
		go func() {
			dateData := api.FetchDates(artist.ConcertDates)
			dateLabel.SetText("Dates:\n" + dateData)
		}()
		go func() {
			relData := api.FetchRelations(artist.RelationsURL)
			relLabel.SetText("Relations:\n" + relData)
		}()

		// Cards pour les sections de données
		locCard := createCard(locLabel)
		dateCard := createCard(dateLabel)
		relCard := createCard(relLabel)

		// Boutons avec style amélioré
		mapBtn := widget.NewButton("Voir sur la carte", func() {
			showMap(artist, w)
		})
		mapBtn.Importance = widget.HighImportance

		backBtn := widget.NewButton("Retour (Échap)", func() { showList() })

		buttonBar := container.NewGridWithColumns(2, mapBtn, backBtn)

		// Organisation du contenu
		var content *fyne.Container
		if artistImage != nil {
			imageCard := createCard(container.NewCenter(artistImage))
			content = container.NewVBox(
				container.NewPadded(header),
				imageCard,
				container.NewPadded(widget.NewSeparator()),
				infoCard,
				container.NewPadded(widget.NewSeparator()),
				locCard,
				dateCard,
				relCard,
				container.NewPadded(widget.NewSeparator()),
				buttonBar,
			)
		} else {
			content = container.NewVBox(
				container.NewPadded(header),
				infoCard,
				container.NewPadded(widget.NewSeparator()),
				locCard,
				dateCard,
				relCard,
				container.NewPadded(widget.NewSeparator()),
				buttonBar,
			)
		}

		w.SetContent(container.NewPadded(container.NewVScroll(content)))
	}

	// --- 3. Liste principale ---
	var list *widget.List

	list = widget.NewList(
		func() int { return len(filtered) },
		func() fyne.CanvasObject {
			label := widget.NewLabel("Nom")
			label.TextStyle = fyne.TextStyle{Bold: false}
			return container.NewPadded(label)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			artist := filtered[i]
			containerObj := o.(*fyne.Container)
			label := containerObj.Objects[0].(*widget.Label)
			label.SetText(artist.Name)
		},
	)
	list.OnSelected = func(id widget.ListItemID) {
		if id < len(filtered) {
			showDetails(filtered[id])
		}
	}

	// --- 4. Barre de recherche ---
	search := widget.NewEntry()
	search.SetPlaceHolder("Rechercher un artiste... (Ctrl+F)")

	// On agrandit la barre via un container
	searchContainer := container.NewPadded(search)

	// --- 5. Bouton Filtres ---
	filterBtn := widget.NewButton("Filtres (Ctrl+M)", nil)
	filterBtn.Importance = widget.MediumImportance

	filterArtist := widget.NewCheck("Artistes", nil)
	filterMembers := widget.NewCheck("Membres", nil)
	filterLocations := widget.NewCheck("Lieux", nil)
	filterFirstAlbum := widget.NewCheck("Premier album", nil)
	filterCreation := widget.NewCheck("Création", nil)

	filterMenuContent := container.NewVBox(
		widget.NewLabelWithStyle("Filtrer par :", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		filterArtist,
		filterMembers,
		filterLocations,
		filterFirstAlbum,
		filterCreation,
	)
	filterMenu := createCard(filterMenuContent)
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

		// Header avec titre et compteur
		title := canvas.NewText("Groupie Tracker", color.White)
		title.TextSize = 24
		title.TextStyle = fyne.TextStyle{Bold: true}
		title.Alignment = fyne.TextAlignCenter

		resultCount := widget.NewLabel(fmt.Sprintf("%d artiste(s)", len(filtered)))
		resultCount.Alignment = fyne.TextAlignCenter
		resultCount.TextStyle = fyne.TextStyle{Italic: true}

		headerBox := container.NewVBox(
			title,
			resultCount,
		)

		// Search large à gauche, filtre à droite
		topBar := container.NewBorder(
			nil, nil, nil, filterBtn,
			searchContainer,
		)

		// Mettre à jour le compteur après chaque recherche
		oldOnChanged := search.OnChanged
		search.OnChanged = func(text string) {
			oldOnChanged(text)
			resultCount.SetText(fmt.Sprintf("%d artiste(s)", len(filtered)))
		}

		content := container.NewBorder(
			container.NewVBox(
				headerBox,
				widget.NewSeparator(),
				container.NewPadded(topBar),
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
	ctrlF := &desktop.CustomShortcut{
		KeyName:  fyne.KeyF,
		Modifier: fyne.KeyModifierControl,
	}
	w.Canvas().AddShortcut(ctrlF, func(shortcut fyne.Shortcut) {
		if !isDetailsPage {
			w.Canvas().Focus(search)
		}
	})

	// Ctrl+M: Afficher/masquer les filtres
	ctrlM := &desktop.CustomShortcut{
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
	ctrlQ := &desktop.CustomShortcut{
		KeyName:  fyne.KeyQ,
		Modifier: fyne.KeyModifierControl,
	}
	w.Canvas().AddShortcut(ctrlQ, func(shortcut fyne.Shortcut) {
		groupie.Quit()
	})

	showList()
	w.ShowAndRun()
}
