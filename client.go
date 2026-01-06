package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

const baseURL = "https://groupietrackers.herokuapp.com/api"

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

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