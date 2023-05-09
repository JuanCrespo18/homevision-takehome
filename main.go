package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/JuanCrespo18/homevision-takehome/photodownloader"
)

// TODO: if needed, we could add parameters for running the script, that would allow, for example, to specify
// of which and how many houses the user wants to download the images (e.g. parameters houses_count and offset)
func main() {
	start := time.Now()
	client := http.Client{}

	if err := photodownloader.DownloadPhotosFromHouses(context.Background(), &client); err != nil {
		log.Fatal(err)
	}

	log.Printf("finished in %s", time.Since(start).Round(time.Millisecond).String())
}
