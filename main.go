package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
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

	api := NewAPI()
	//localhost:8080/api?q=san%20francisco
	http.HandleFunc("/api", api.Handler)

	http.ListenAndServe(os.Getenv("PORT"), nil)
}

type API struct {
	cache *redis.Client
}

func NewAPI() *API {
	var opts *redis.Options

	if os.Getenv("LOCAL") == "true" {
		redisAddress := fmt.Sprintf("%s:6379", os.Getenv("REDIS_URL"))
		opts = &redis.Options{
			Addr:     redisAddress,
			Password: "", // no password set
			DB:       0,  // use default DB
		}
	} else {
		buildOpts, err := redis.ParseURL(os.Getenv("REDIS_URL"))
		if err != nil {
			panic(err)
		}

		opts = buildOpts
	}
	rdb := redis.NewClient(opts)

	return &API{
		cache: rdb,
	}
}

func (a *API) Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Hitted endpoint")
	city := r.URL.Query().Get("city")
	data, isCached, err := a.getData(r.Context(), city)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := APIResponse{
		Cache: isCached,
		Data:  data,
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		fmt.Printf("Error on Encoding reponse : %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

}

func (a *API) getData(ctx context.Context, city string) ([]NominatimResponse, bool, error) {

	value, err := a.cache.Get(ctx, city).Result()
	if err == redis.Nil {
		escape := url.PathEscape(city)

		address := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&format=json", escape)

		resp, err := http.Get(address)

		if err != nil {
			return nil, false, err
		}

		data := make([]NominatimResponse, 0)

		err = json.NewDecoder(resp.Body).Decode(&data)

		if err != nil {
			return nil, false, err
		}
		dataBytes, err := json.Marshal(data)
		if err != nil {
			return nil, false, err
		}
		//Set value in Redis
		err = a.cache.Set(ctx, city, bytes.NewBuffer(dataBytes).Bytes(), time.Second*15).Err()
		if err != nil {
			return nil, false, err
		}
		//Return value
		return data, false, err
	} else if err != nil {
		fmt.Printf("Error on calling redis: %v\n", err)
		return nil, false, err
	} else {
		data := make([]NominatimResponse, 0)
		err := json.Unmarshal(bytes.NewBufferString(value).Bytes(), &data)
		if err != nil {
			return nil, false, err
		}
		return data, true, err
	}

}
