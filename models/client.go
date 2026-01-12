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

func FetchLocation() ([]Artist, error) {
	resp, err := httpClient.Get(baseURL + "/artists")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("API error: " + resp.Status)
	}

	var LocationsURL []Artist
	err = json.NewDecoder(resp.Body).Decode(&LocationsURL)
	if err != nil {
		return nil, err
	}

	return LocationsURL, nil
}

func FetchDates() ([]Artist, error) {
	resp, err := httpClient.Get(baseURL + "/artists")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("API error: " + resp.Status)
	}

	var DatesURL []Artist
	err = json.NewDecoder(resp.Body).Decode(&DatesURL)
	if err != nil {
		return nil, err
	}

	return DatesURL, nil
}

func FetchRelations() ([]Artist, error) {
	resp, err := httpClient.Get(baseURL + "/artists")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("API error: " + resp.Status)
	}

	var RelationsURL []Artist
	err = json.NewDecoder(resp.Body).Decode(&RelationsURL)
	if err != nil {
		return nil, err
	}

	return RelationsURL, nil
}
