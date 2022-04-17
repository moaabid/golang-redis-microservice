package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type APIResponse struct {
	Cache bool                `json:"cache"`
	Data  []NominatimResponse `json:"Data"`
}

type NominatimResponse struct {
	PlaceID     int      `json:"place_id"`
	Licence     string   `json:"licence"`
	OsmType     string   `json:"osm_type"`
	OsmID       int      `json:"osm_id"`
	Boundingbox []string `json:"boundingbox"`
	Lat         string   `json:"lat"`
	Lon         string   `json:"lon"`
	DisplayName string   `json:"display_name"`
	Class       string   `json:"class"`
	Type        string   `json:"type"`
	Importance  float64  `json:"importance"`
}

func main() {
	fmt.Println("Starting Server")

	//localhost:8080/api?q=san%20francisco
	http.HandleFunc("/api", Handler)

	http.ListenAndServe(":8080", nil)
}

func Handler(w http.ResponseWriter, r *http.Request) {

	city := r.URL.Query().Get("city")
	data, err := getData(city)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := APIResponse{
		Cache: false,
		Data:  data,
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		fmt.Printf("Error on Encoding reponse : %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

}

func getData(city string) ([]NominatimResponse, error) {
	escape := url.PathEscape(city)

	address := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&format=json", escape)

	resp, err := http.Get(address)

	if err != nil {
		return nil, err
	}

	data := make([]NominatimResponse, 0)

	err = json.NewDecoder(resp.Body).Decode(&data)

	if err != nil {
		return nil, err
	}

	return data, err
}
