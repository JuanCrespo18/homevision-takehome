package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type House struct {
	ID        int    `json:"id,omitempty"`
	Address   string `json:"address,omitempty"`
	HomeOwner string `json:"homeowner,omitempty"`
	Price     int    `json:"price,omitempty"`
	PhotoURL  string `json:"photoURL,omitempty"`
}

type HouseAPIResponse struct {
	Houses  []House `json:"houses,omitempty"`
	Ok      bool    `json:"ok,omitempty"`
	Message string  `json:"message,omitempty"`
}

func main() {
	start := time.Now()
	client := http.Client{}

	housesRequest, err := http.NewRequest(
		http.MethodGet, "http://app-homevision-staging.herokuapp.com/api_project/houses?page=1&per_page=10", nil)
	if err != nil {
		log.Fatalf("error creating housesRequest. %v", err)
	}

	var houseAPIResponse HouseAPIResponse
	for {
		response, err := client.Do(housesRequest)
		if err != nil {
			log.Fatalf("error getting houses. %v", err)
		}

		body, err := io.ReadAll(response.Body)
		if err != nil {
			log.Fatalf("error reading response body. %v", err)
		}

		if err := json.Unmarshal(body, &houseAPIResponse); err != nil {
			log.Fatalf("error unmarshalling response body. %v", err)
		}
		response.Body.Close()

		if houseAPIResponse.Ok {
			break
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(len(houseAPIResponse.Houses))

	for _, house := range houseAPIResponse.Houses {
		go downloadImageAndSaveItToFile(&wg, house, client)
	}

	wg.Wait()
	log.Printf("finished in %s", time.Since(start).Round(time.Millisecond).String())
}

func downloadImageAndSaveItToFile(wg *sync.WaitGroup, house House, client http.Client) {
	imageRequest, err := http.NewRequest(http.MethodGet, house.PhotoURL, nil)
	if err != nil {
		log.Fatalf("error creating image request. %v", err)
	}

	response, err := client.Do(imageRequest)
	if err != nil {
		log.Fatalf("error getting image. %v", err)
	}
	defer response.Body.Close()

	ext := filepath.Ext(house.PhotoURL)
	filePath := fmt.Sprintf("tmp/%d-%s%s", house.ID, house.Address, ext)
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("error creating file for image. %v", err)
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		log.Fatalf("error copying image data to file. %v", err)
	}

	log.Printf("finished house %d", house.ID)
	wg.Done()
}
