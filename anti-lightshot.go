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
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"runtime"
	"strings"

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
	imgurImageID           = "https://i.imgur.com/y5Onqkp.png"
)

/////////////////////////////////////////////////////////////
//FUNCTIONS:
func eliminateNewLineCrap(text string) string {
	os := runtime.GOOS
	if os == "windows" {
		return (strings.Replace(text, "\r\n", "", -1))
	} else {
		return (strings.Replace(text, "\n", "", -1))
	}
}

func trimStringBetweenTwo(input string, startS string, endS string) (result string, found bool) {
	s := strings.Index(input, startS)
	if s == -1 {
		return result, false
	}

	newS := input[s+len(startS):]
	e := strings.Index(newS, endS)
	if e == -1 {
		return result, false
	}
	result = newS[:e]
	return result, true
}

func generateRandomImageID() string {
	var newImageId string
	for i := 0; i < 6; i++ {
		newImageId = newImageId + imageIDcharacters[rand.Intn(len(imageIDcharacters))]
	}

	return newImageId
}

func downloadImageByURL(URL string) error {
	//Get the response bytes from the url
	URL = eliminateNewLineCrap(URL)
	imageIDFromURL, _ := trimStringBetweenTwo(URL, "https://i.imgur.com/", ".png")

	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("Received non 200 response code")
	}
	//Create a empty file
	file, err := os.Create(pathToDownloadedImages + imageIDFromURL + ".png")
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

func getPageHTML(url string) string {
	// return doc.FullText()
	// resp, err := http.Get(url)
	// fmt.Println(resp.ContentLength)
	// resp.Header.Add("Accept", "application/xml")
	// resp.Header.Add("Content-Type", "application/xml; charset=utf-8")
	// if err != nil {
	// 	// handle error
	// }
	// defer resp.Body.Close()

	// htmlbody, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	panic(err)
	// }

	return "htmlbody"
}

func downloadImageByLightshotID(imageID string) error {
	// 0. create image URL
	var imageLightshotURL = "https://prnt.sc/" + imageID

	// 1. parse the HTML page
	res, err := http.Get(imageLightshotURL)
	if err != nil {
		// handle error
	}
	defer res.Body.Close()

	htmlbody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", htmlbody)

	// //Get the response bytes from the url
	// var fullImageURL = "https://prnt.sc/" + imageID
	// response, err := http.Get(fullImageURL)
	// if err != nil {
	// 	return err
	// }
	// defer response.Body.Close()

	// if response.StatusCode != 200 {
	// 	return errors.New("Received non 200 response code")
	// }
	// //Create a empty file
	// file, err := os.Create(pathToDownloadedImages + imageID + ".png")
	// if err != nil {
	// 	return err
	// }
	// defer file.Close()

	// //Write the bytes to the fiel
	// _, err = io.Copy(file, response.Body)
	// if err != nil {
	// 	return err
	// }

	return nil
}

func recordDownloadedImageID(imageID string) {
	logfile, openerr := os.OpenFile(pathToLogFile, os.O_APPEND|os.O_WRONLY, 0644)

	_, logerr := logfile.WriteString("https://prnt.sc/" + imageID + "\n")

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
	specifyImageURL := flag.String("imageurl", "", "you can specify image url here, e.g. --imageurl https://i.imgur.com/y5Onqkp.png")
	readPageHTML := flag.String("readhtml", "", "read page html of the specified address, e.g. --readhtml https://i.imgur.com/y5Onqkp.png")
	//slowMode := flag.Bool("slowmode", false, "use slowmode to pre-define image id before downloading them")
	flag.Parse()

	if *readPageHTML != "" {
		fmt.Printf("%s\n", string(getPageHTML(*readPageHTML)))
		os.Exit(3)
	} else if *specifyImageURL != "" {
		fmt.Print("you provided image url: ")
		fmt.Println(*specifyImageURL)
		downloadImageByURL(*specifyImageURL)
	} else if *specifyImageURL == "" && *howManyImages != 0 {
		fmt.Print("you specified images: ")
		fmt.Println(*howManyImages)
		downloadingProgressBar := progressbar.DefaultBytes(
			(int64(*howManyImages)),
			"Downloading",
		)
		for i := 0; i < (*howManyImages); i++ {
			newImageID = generateRandomImageID()
			if newImageID != previousImageID && newImageID != "" {
				recordDownloadedImageID(newImageID)
				downloadImageByLightshotID(newImageID)
				previousImageID = newImageID
				downloadingProgressBar.Add(1)
			} else {
				i = i - 1
			}
		}
	} else {
		fmt.Println("no command provided")
		os.Exit(3)
	}
}
