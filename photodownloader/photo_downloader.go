package photodownloader

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	ErrCreatingHousesRequest  = "error creating housesRequest"
	ErrGettingHouses          = "error getting houses"
	ErrReadingResponseBody    = "error reading response body"
	ErrUnmarshallingResponse  = "error unmarshalling response body"
	ErrDownloadingImageToFile = "error downloading image to file"
	ErrGettingImage           = "error getting image"
	ErrCreatingFile           = "error creating file for image"
	ErrCopyingDataToFile      = "error copying image data to file"
	ErrAPIUnavailable         = "API not available"
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

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// DownloadPhotosFromHouses uses the home-vision houses API to get the image urls for each house, and download
// those image to a file.
func DownloadPhotosFromHouses(ctx context.Context, client HttpClient) error {
	os.Mkdir("tmp", os.ModePerm)

	// TODO: here we could build the url using the parameters mentioned in the TODO of the main() function,
	// in order to fetch the houses specified by the user.
	housesRequest, err := http.NewRequestWithContext(ctx,
		http.MethodGet, "http://app-homevision-staging.herokuapp.com/api_project/houses?page=1&per_page=10", nil)
	if err != nil {
		return fmt.Errorf("%s. %v", ErrCreatingHousesRequest, err)
	}

	var calls int
	var response *http.Response
	for {
		response, err = client.Do(housesRequest)
		if err != nil {
			return fmt.Errorf("%s. %v", ErrGettingHouses, err)
		}
		if response.StatusCode == http.StatusOK {
			break
		}
		if response.StatusCode >= 400 && response.StatusCode < 500 {
			return fmt.Errorf("%s. Status Code: %d", ErrGettingHouses, response.StatusCode)
		}

		// TODO: this is a very simple and rudimentary retry implementation, if needed it can be made better using back off for example.
		if calls > 3 {
			return fmt.Errorf("%s. %s", ErrGettingHouses, ErrAPIUnavailable)
		}
		calls++
		time.Sleep(time.Millisecond * 100)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("%s. %v", ErrReadingResponseBody, err)
	}

	var houseAPIResponse HouseAPIResponse
	if err := json.Unmarshal(body, &houseAPIResponse); err != nil {
		return fmt.Errorf("%s. %v", ErrUnmarshallingResponse, err)
	}
	defer response.Body.Close()

	errGroup := new(errgroup.Group)

	for _, h := range houseAPIResponse.Houses {
		house := h
		errGroup.Go(func() error {
			return downloadImageAndSaveItToFile(ctx, house, client)
		})
	}

	if err := errGroup.Wait(); err != nil {
		return fmt.Errorf("%s. %v", ErrDownloadingImageToFile, err)
	}

	return nil
}

// downloadImageAndSaveItToFile receives a House and a HttpClient, then uses the house's PhotoURL to
// download the image and saves it to a file.
func downloadImageAndSaveItToFile(ctx context.Context, house House, client HttpClient) error {
	imageRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, house.PhotoURL, nil)
	if err != nil {
		return fmt.Errorf("error creating image request. %v", err)
	}

	var calls int
	var response *http.Response
	for {
		response, err = client.Do(imageRequest)
		if err != nil {
			return fmt.Errorf("%s. %v", ErrGettingImage, err)
		}
		defer response.Body.Close()

		if response.StatusCode == http.StatusOK {
			break
		}
		if response.StatusCode >= 400 && response.StatusCode < 500 {
			return fmt.Errorf("%s. Status Code: %d", ErrGettingImage, response.StatusCode)
		}

		// TODO: this is a very simple and rudimentary retry implementation, if needed it can be made better using back off for example.
		if calls > 3 {
			return fmt.Errorf("%s. %s", ErrGettingImage, ErrAPIUnavailable)
		}
		calls++
		time.Sleep(time.Millisecond * 100)
	}

	// TODO: this section creates a file locally into the tmp directory. This, however, should be addressed using an
	// object storage (e.g. Amazon S3, or MinIO). Doing so we upload the files to the OS server and then they can be
	// accessed from another application or manually.
	ext := filepath.Ext(house.PhotoURL)
	filePath := fmt.Sprintf("tmp/%d-%s%s", house.ID, house.Address, ext)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("%s. %v", ErrCreatingFile, err)
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return fmt.Errorf("%s. %v", ErrCopyingDataToFile, err)
	}

	return nil
}
