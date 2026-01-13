package groupie

// Artist représente un artiste ou groupe musical tel que défini par l'API Groupie Tracker.
// Chaque champ est mappé à une clé JSON pour faciliter le tri de sinformations.
type Artist struct {
	ID           int      `json:"id"`            // Identifiant unique de l'artiste
	Image        string   `json:"image"`         // URL de l'image officielle
	Name         string   `json:"name"`          // Nom du groupe ou artiste
	Members      []string `json:"members"`       // Liste des membres du groupe
	CreationDate int      `json:"creationDate"`  // Année de création du groupe
	FirstAlbum   string   `json:"firstAlbum"`    // Date de sortie du premier album
	LocationsURL string   `json:"locations"`     // URL vers les lieux de concert
	ConcertDates string   `json:"concertDates"`  // URL vers les dates de concert
	RelationsURL string   `json:"relations"`     // URL vers les relations lieu/date
}

// LocationData est utilisée pour trier les lieux de concert depuis l'API.
// Exemple : ["new_york", "paris", "tokyo"]
type LocationData struct {
	Locations []string `json:"locations"` // Liste brute des lieux
}

// DateData est utilisée pour trier les dates de concert depuis l'API.
// Exemple : ["2023-05-12", "2023-06-01"]
type DateData struct {
	Dates []string `json:"dates"` // Liste brute des dates
}

// RelationData permet de relier chaque lieu à ses dates de concert.
// Exemple : {"new_york": ["2023-05-12", "2023-06-01"]}
type RelationData struct {
	DatesLocations map[string][]string `json:"datesLocations"` // Mapping lieu → dates
}