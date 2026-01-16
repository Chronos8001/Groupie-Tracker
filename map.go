package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"time"
)

type GeoResult struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

var client = &http.Client{Timeout: 10 * time.Second}

// GetCoordinates : Trouve Lat/Lon via le nom de la ville
func GetCoordinates(city string) (string, string, error) {
	q := url.QueryEscape(city)
	url := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&format=json&limit=1", q)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "GroupieTracker/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var res []GeoResult
	if json.NewDecoder(resp.Body).Decode(&res) != nil || len(res) == 0 {
		return "", "", fmt.Errorf("not found")
	}
	return res[0].Lat, res[0].Lon, nil
}

// GetOSMTileURL : Calcule l'URL de l'image (Tuile) pour une position
func GetOSMTileURL(lat, lon float64, zoom int) string {
	x := int(math.Floor((lon + 180.0) / 360.0 * math.Pow(2.0, float64(zoom))))
	latRad := lat * math.Pi / 180.0
	y := int(math.Floor((1.0 - math.Log(math.Tan(latRad)+1.0/math.Cos(latRad))/math.Pi) / 2.0 * math.Pow(2.0, float64(zoom))))
	return fmt.Sprintf("https://tile.openstreetmap.org/%d/%d/%d.png", zoom, x, y)
}
