package groupie

import (
	"encoding/json" // Pour décoder les réponses JSON de l'API
	"errors"        // Pour gérer les erreurs personnalisées
	"fmt"           // Pour formater les chaînes de caractères
	"io"            // Pour lire le corps des réponses HTTP
	"net/http"      // Pour effectuer les requêtes HTTP
	"strings"       // Pour manipuler les chaînes (nettoyage, formatage)
	"time"          // Pour gérer les délais de requêtes
)

// URL de base de l'API Groupie Tracker
const baseURL = "https://groupietrackers.herokuapp.com/api"

// Client HTTP avec timeout de 10 secondes pour éviter les blocages
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// FetchArtists récupère la liste des artistes depuis l'API
// Elle renvoie un tableau d'objets Artist ou une erreur
func FetchArtists() ([]Artist, error) {
	resp, err := httpClient.Get(baseURL + "/artists")
	if err != nil {
		return nil, err // Erreur réseau ou requête
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("API error: " + resp.Status) // Erreur côté serveur
	}

	var artists []Artist
	err = json.NewDecoder(resp.Body).Decode(&artists)
	if err != nil {
		return nil, err // Erreur de décodage JSON
	}

	return artists, nil
}

// GetFromURL récupère le contenu brut d'une URL
// Utile pour afficher directement les données sans traitement
func GetFromURL(url string) string {
	resp, err := httpClient.Get(url)
	if err != nil {
		return "Erreur de chargement: " + err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "Erreur serveur: " + resp.Status
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "Erreur lecture: " + err.Error()
	}

	return string(body) // Retourne le JSON brut sous forme de string
}

// FetchLocation récupère les lieux de concert et les formate
// Exemple : "new_york" devient "New York"
func FetchLocation(url string) string {
	resp, err := httpClient.Get(url)
	if err != nil {
		return "Erreur: " + err.Error()
	}
	defer resp.Body.Close()

	var loc LocationData
	if err := json.NewDecoder(resp.Body).Decode(&loc); err != nil {
		return "Erreur lecture données"
	}

	formatted := []string{}
	for _, l := range loc.Locations {
		// Nettoyage : remplace "_" par " ", met la première lettre en majuscule
		formatted = append(formatted, strings.Title(strings.ReplaceAll(l, "_", " ")))
	}
	return strings.Join(formatted, "\n") // Retourne une liste formatée
}

// FetchDates récupère les dates de concert et les nettoie
// Supprime les caractères parasites comme "*"
func FetchDates(url string) string {
	resp, err := httpClient.Get(url)
	if err != nil {
		return "Erreur: " + err.Error()
	}
	defer resp.Body.Close()

	var d DateData
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return "Erreur lecture données"
	}

	formatted := []string{}
	for _, date := range d.Dates {
		formatted = append(formatted, strings.ReplaceAll(date, "*", ""))
	}
	return strings.Join(formatted, "\n") // Liste propre des dates
}

// FetchRelations récupère les relations entre lieux et dates
// Formatage lisible : chaque lieu suivi des dates associées
func FetchRelations(url string) string {
	resp, err := httpClient.Get(url)
	if err != nil {
		return "Erreur: " + err.Error()
	}
	defer resp.Body.Close()

	var rel RelationData
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "Erreur lecture données"
	}

	var builder strings.Builder
	for loc, dates := range rel.DatesLocations {
		// Nettoyage du nom du lieu
		cleanLoc := strings.Title(strings.ReplaceAll(loc, "_", " "))
		builder.WriteString(fmt.Sprintf("%s :\n", cleanLoc))
		for _, d := range dates {
			builder.WriteString(fmt.Sprintf("  - %s\n", d)) // Ajout des dates
		}
		builder.WriteString("\n")
	}
	return builder.String() // Format final lisible
}

// nouveaux

// FetchImage récupère l'URL de l'image d'un artiste
func FetchImage(url string) (string, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", errors.New("API error: " + resp.Status)
	}

	var artist Artist
	err = json.NewDecoder(resp.Body).Decode(&artist)
	if err != nil {
		return "", err
	}

	return artist.Image, nil
}

// FetchFirstAlbum récupère la date du premier album d'un artiste
func FetchFirstAlbum(url string) (string, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", errors.New("API error: " + resp.Status)
	}

	var artist Artist
	err = json.NewDecoder(resp.Body).Decode(&artist)
	if err != nil {
		return "", err
	}

	return artist.FirstAlbum, nil
}

// FetchMembers récupère la liste des membres d'un artiste
func FetchMembers(url string) ([]string, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("API error: " + resp.Status)
	}

	var artist Artist
	err = json.NewDecoder(resp.Body).Decode(&artist)
	if err != nil {
		return nil, err
	}

	return artist.Members, nil
}

// FetchCreationDate récupère l'année de création d'un artiste
func FetchCreationDate(url string) (int, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, errors.New("API error: " + resp.Status)
	}

	var artist Artist
	err = json.NewDecoder(resp.Body).Decode(&artist)
	if err != nil {
		return 0, err
	}

	return artist.CreationDate, nil
}
