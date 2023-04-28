package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var body string
var houses io.ReadCloser

func TestDownloadPhotosFromHousesOK(t *testing.T) {
	setup()

	image := "hola"
	photo := io.NopCloser(strings.NewReader(image))
	photo2 := io.NopCloser(strings.NewReader(image))

	mockClient := mockClient{
		responses: []*http.Response{
			{
				StatusCode: http.StatusOK,
				Body:       houses,
			},
			{
				StatusCode: http.StatusOK,
				Body:       photo,
			},
			{
				StatusCode: http.StatusOK,
				Body:       photo2,
			},
		},
	}
	err := DownloadPhotosFromHouses(context.Background(), &mockClient)
	assert.Nil(t, err)

	dat, err := os.ReadFile("tmp/0-house-0.jpg")
	assert.Nil(t, err)
	assert.Equal(t, image, string(dat))

	dat, err = os.ReadFile("tmp/1-house-1.jpg")
	assert.Nil(t, err)
	assert.Equal(t, image, string(dat))
}

func TestDownloadPhotosFromHousesFailCreatingRequest(t *testing.T) {
	mockClient := mockClient{}
	err := DownloadPhotosFromHouses(nil, &mockClient)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), ErrCreatingHousesRequest)
}

func TestDownloadPhotosFromHousesFailGettingHouses(t *testing.T) {
	aMockClient := mockClient{
		err:        fmt.Errorf("mock error"),
		failHouses: true,
	}
	err := DownloadPhotosFromHouses(context.Background(), &aMockClient)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), ErrGettingHouses)
	assert.Contains(t, err.Error(), "mock error")

	aMockClient = mockClient{
		responses: []*http.Response{
			{
				StatusCode: http.StatusBadRequest,
			},
		},
	}
	err = DownloadPhotosFromHouses(context.Background(), &aMockClient)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), ErrGettingHouses)
	assert.Contains(t, err.Error(), strconv.Itoa(http.StatusBadRequest))

	aMockClient = mockClient{
		responses: []*http.Response{
			{
				StatusCode: http.StatusServiceUnavailable,
			},
			{
				StatusCode: http.StatusServiceUnavailable,
			},
			{
				StatusCode: http.StatusServiceUnavailable,
			},
			{
				StatusCode: http.StatusServiceUnavailable,
			},
			{
				StatusCode: http.StatusServiceUnavailable,
			},
		},
	}
	err = DownloadPhotosFromHouses(context.Background(), &aMockClient)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), ErrGettingHouses)
	assert.Contains(t, err.Error(), ErrAPIUnavailable)
}

func TestDownloadPhotosFromHousesFailReadingResponseBody(t *testing.T) {
	mockClient := mockClient{
		responses: []*http.Response{
			{
				StatusCode: http.StatusOK,
				Body:       errReader(1),
			},
		},
	}
	err := DownloadPhotosFromHouses(context.Background(), &mockClient)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), ErrReadingResponseBody)
	assert.Contains(t, err.Error(), "mock error")
}

func TestDownloadPhotosFromHousesFailUnmarshallingBody(t *testing.T) {
	mockClient := mockClient{
		responses: []*http.Response{
			{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("mock body")),
			},
		},
	}
	err := DownloadPhotosFromHouses(context.Background(), &mockClient)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), ErrUnmarshallingResponse)
}

func TestDownloadPhotosFromHousesFailGettingImage(t *testing.T) {
	setup()
	mockClient := mockClient{
		responses: []*http.Response{
			{
				StatusCode: http.StatusOK,
				Body:       houses,
			},
		},
		err:        fmt.Errorf("mock error"),
		failPhotos: true,
	}
	err := DownloadPhotosFromHouses(context.Background(), &mockClient)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), ErrGettingImage)
}

func TestDownloadPhotosFromHousesFailGettingImageBadRequest(t *testing.T) {
	setup()
	mockClient := mockClient{
		responses: []*http.Response{
			{
				StatusCode: http.StatusOK,
				Body:       houses,
			},
			{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader("mock body")),
			},
			{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("mock body")),
			},
		},
	}
	err := DownloadPhotosFromHouses(context.Background(), &mockClient)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), ErrGettingImage)
	assert.Contains(t, err.Error(), strconv.Itoa(http.StatusBadRequest))
}

func TestDownloadPhotosFromHousesFailGettingImageUnavailableAPI(t *testing.T) {
	body = `
		{
			"houses": [
				{
					"id": 0,
					"address": "house-0",
					"homeowner": "Nicole Bone",
					"price": 105124,
					"photoURL": "https://image.shutterstock.com/image-photo/big-custom-made-luxury-house-260nw-374099713.jpg"
				}
			],
			"ok": true
		}
	`
	houses = io.NopCloser(strings.NewReader(body))
	mockClient := mockClient{
		responses: []*http.Response{
			{
				StatusCode: http.StatusOK,
				Body:       houses,
			},
			{
				StatusCode: http.StatusServiceUnavailable,
				Body:       io.NopCloser(strings.NewReader("mock body")),
			},
			{
				StatusCode: http.StatusServiceUnavailable,
				Body:       io.NopCloser(strings.NewReader("mock body")),
			},
			{
				StatusCode: http.StatusServiceUnavailable,
				Body:       io.NopCloser(strings.NewReader("mock body")),
			},
			{
				StatusCode: http.StatusServiceUnavailable,
				Body:       io.NopCloser(strings.NewReader("mock body")),
			},
			{
				StatusCode: http.StatusServiceUnavailable,
				Body:       io.NopCloser(strings.NewReader("mock body")),
			},
		},
	}
	err := DownloadPhotosFromHouses(context.Background(), &mockClient)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), ErrGettingImage)
	assert.Contains(t, err.Error(), ErrAPIUnavailable)
}

func TestDownloadPhotosFromHousesFailCreatingFile(t *testing.T) {
	body = `
		{
			"houses": [
				{
					"id": 0,
					"address": "asd/house-0",
					"homeowner": "Nicole Bone",
					"price": 105124,
					"photoURL": "https://image.shutterstock.com/image-photo/big-custom-made-luxury-house-260nw-374099713.jpg"
				}
			],
			"ok": true
		}
	`
	houses = io.NopCloser(strings.NewReader(body))

	image := "hola"
	photo := io.NopCloser(strings.NewReader(image))

	mockClient := mockClient{
		responses: []*http.Response{
			{
				StatusCode: http.StatusOK,
				Body:       houses,
			},
			{
				StatusCode: http.StatusOK,
				Body:       photo,
			},
		},
	}
	err := DownloadPhotosFromHouses(context.Background(), &mockClient)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), ErrCreatingFile)
}

func TestDownloadPhotosFromHousesFailCopyingImageToFile(t *testing.T) {
	body = `
		{
			"houses": [
				{
					"id": 0,
					"address": "house-0",
					"homeowner": "Nicole Bone",
					"price": 105124,
					"photoURL": "https://image.shutterstock.com/image-photo/big-custom-made-luxury-house-260nw-374099713.jpg"
				}
			],
			"ok": true
		}
	`
	houses = io.NopCloser(strings.NewReader(body))
	mockClient := mockClient{
		responses: []*http.Response{
			{
				StatusCode: http.StatusOK,
				Body:       houses,
			},
			{
				StatusCode: http.StatusOK,
				Body:       errReader(1),
			},
		},
	}
	err := DownloadPhotosFromHouses(context.Background(), &mockClient)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), ErrCopyingDataToFile)
}

func setup() {
	body = `
		{
			"houses": [
				{
					"id": 0,
					"address": "house-0",
					"homeowner": "Nicole Bone",
					"price": 105124,
					"photoURL": "https://image.shutterstock.com/image-photo/big-custom-made-luxury-house-260nw-374099713.jpg"
				},
				{
					"id": 1,
					"address": "house-1",
					"homeowner": "Rheanna Walsh",
					"price": 161856,
					"photoURL": "https://media-cdn.tripadvisor.com/media/photo-s/09/7c/a2/1f/patagonia-hostel.jpg"
				}
			],
			"ok": true
		}
	`
	houses = io.NopCloser(strings.NewReader(body))
}

type mockClient struct {
	responses  []*http.Response
	err        error
	failHouses bool
	failPhotos bool
	counter    int
}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	if m.failHouses {
		return nil, m.err
	}

	if m.counter == 0 {
		m.counter++
		return m.responses[0], nil
	}

	if m.failPhotos {
		return nil, m.err
	}

	response := m.responses[m.counter]
	m.counter++
	return response, nil
}

type errReader int

func (r errReader) Close() error {
	return fmt.Errorf("mock error")
}

func (errReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("mock error")
}
