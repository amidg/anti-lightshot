package main

import (

	//required to read image from file

	_ "image/png"
	"os"

	//needed to draw text on image

	"C"
)
import (
	"errors"
	"flag"
	"io"
	"math/rand"
	"net/http"

	"github.com/schollz/progressbar/v3"
)

/////////////////////////////////////////////
//GLOBAL VARIABLES
var (
	pathToDownloadedImages = "./DownloadedContent/"
	pathToLogFile          = "./downloadlog.txt"
	imageIDcharacters      = [36]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}
	downloadedImageBuffer  []string // dynamic slice of image IDs
	newImageID             string
	previousImageID        string
)

/////////////////////////////////////////////////////////////
//FUNCTIONS:
func generateRandomImageID() string {
	var newImageId string
	for i := 0; i < 6; i++ {
		newImageId = newImageId + imageIDcharacters[rand.Intn(len(imageIDcharacters))]
	}

	return newImageId
}

func downloadImageByName(imageID string) error {
	//Get the response bytes from the url
	var fullImageURL = "https://prnt.sc/" + imageID
	response, err := http.Get(fullImageURL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("Received non 200 response code")
	}
	//Create a empty file
	file, err := os.Create(pathToDownloadedImages + imageID)
	if err != nil {
		return err
	}
	defer file.Close()

	//Write the bytes to the fiel
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func recordDownloadedImageID(imageID string) {
	logfile, openerr := os.OpenFile(pathToLogFile, os.O_APPEND|os.O_WRONLY, 0644)

	_, logerr := logfile.WriteString(imageID + "\n")

	if openerr != nil || logerr != nil {
		println("file error")
		os.Exit(3)
	}

	logfile.Close()
}

/*
1. Build Linux executable
	go build -o anti_lightshot anti-lightshot.go

2. Build Windows executable
	g++ -pthread main.cpp wolfimggen.a -o wolfwisdomgenerator
*/

func main() {
	howManyImages := flag.Int("howmanyimages", 0, "specify how many images to download")
	flag.Parse()

	downloadingProgressBar := progressbar.DefaultBytes(
		(int64(*howManyImages)),
		"Downloading",
	)
	for i := 0; i < (*howManyImages); i++ {
		newImageID = generateRandomImageID()
		if newImageID != previousImageID {
			recordDownloadedImageID(newImageID)
			downloadImageByName(newImageID)
			previousImageID = newImageID
			downloadingProgressBar.Add(1)
		} else {
			i = i - 1
		}
	}
}
