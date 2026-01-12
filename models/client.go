package groupie

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const baseURL = "https://groupietrackers.herokuapp.com/api"

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// FetchArtists récupère la liste principale
func FetchArtists() ([]Artist, error) {
	resp, err := httpClient.Get(baseURL + "/artists")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("API error: " + resp.Status)
	}

	var artists []Artist
	err = json.NewDecoder(resp.Body).Decode(&artists)
	if err != nil {
		return nil, err
	}

	return artists, nil
}

// GetFromURL télécharge le contenu d'une URL (pour les details locations/dates/relations)
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
	// Retourne le JSON brut sous forme de string pour l'affichage
	return string(body)
}

// FetchLocation récupère et formate les lieux
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

	// Nettoyage: remplace "_" par " ", met en majuscule, etc.
	formatted := []string{}
	for _, l := range loc.Locations {
		formatted = append(formatted, strings.Title(strings.ReplaceAll(l, "_", " ")))
	}
	return strings.Join(formatted, "\n")
}

// FetchDates récupère et formate les dates
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

	// Retire les "*" souvent présents dans l'API
	formatted := []string{}
	for _, date := range d.Dates {
		formatted = append(formatted, strings.ReplaceAll(date, "*", ""))
	}
	return strings.Join(formatted, "\n")
}

// FetchRelations récupère et formate les relations
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
		cleanLoc := strings.Title(strings.ReplaceAll(loc, "_", " "))
		builder.WriteString(fmt.Sprintf("%s :\n", cleanLoc))
		for _, d := range dates {
			builder.WriteString(fmt.Sprintf("  - %s\n", d))
		}
		builder.WriteString("\n")
	}
	return builder.String()
}
