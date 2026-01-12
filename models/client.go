package groupie

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
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
