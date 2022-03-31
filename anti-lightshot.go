package main

import (

	//required to read image from file

	_ "image/png"
	"os"

	//needed to draw text on image

	"C"
)
import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

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

func getPageHTML(url string, timeout time.Duration) (content []byte, err error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	ctx, cancel_func := context.WithTimeout(context.Background(), timeout)
	request = request.WithContext(ctx)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		cancel_func()
		return nil, err
	}

	return ioutil.ReadAll(response.Body)
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

func getActualImageLink(pathToHTML string) string {
	var imageURL string
	var parsedHTML []string

	file, _ := os.Open(pathToHTML)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		parsedHTML = append(parsedHTML, scanner.Text())
		if strings.Contains(scanner.Text(), "https://i.imgur.com/") {
			imageURL = scanner.Text()
			break
		}
	}

	//imageURL, _ = trimStringBetweenTwo(imageURL, "src=", " crossorigin=")
	fmt.Print("actual image link: ")
	fmt.Println(imageURL)
	// imageURL = imageURL[1:]
	// imageURL = imageURL[:len(imageURL)-1]

	return imageURL
}

func downloadUsingWget(url string) {
	url = eliminateNewLineCrap(url)
	downloadLightshotCommand := exec.Command("wget", url)
	_, err := downloadLightshotCommand.Output()
	if err != nil {
		fmt.Println("error downloading file")
	}
}

func checkIfFileExists(filename string) bool {
	_, err := os.Stat(filename)
	if !os.IsExist(err) {
		checkIfFileExists(filename)
	}
	return os.IsExist(err) && checkIfFileExists(filename)
}

func downloadImageByLightshotID(imageID string) error {
	// 0. create image URL
	imageID = eliminateNewLineCrap(imageID)
	var imageLightshotURL = "https://prnt.sc/" + imageID

	// 1. get file with certain ID
	var isLightshotHacked = false
	downloadLightshotCommand := exec.Command("wget", imageLightshotURL)
	_, err := downloadLightshotCommand.Output()
	if err != nil {
		fmt.Println("error downloading file")
	} else if err == nil {
		for {
			_, err := os.Stat(imageID)
			if !os.IsExist(err) {
				break
			}
		}

		isLightshotHacked = true
	}

	// 2. read their crap
	if isLightshotHacked {
		downloadImageByURL(getActualImageLink(imageID))
		e := os.Remove(imageID)
		if e != nil {
			fmt.Println("file does not exit")
			os.Exit(3)
		}
	}

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
	getImageLink := flag.String("getimglink", "", "get actual image link, e.g. --getimglink https://prnt.sc/2a93m0")
	//slowMode := flag.Bool("slowmode", false, "use slowmode to pre-define image id before downloading them")
	flag.Parse()

	if *readPageHTML != "" {
		pageHTML, _ := getPageHTML(eliminateNewLineCrap(*readPageHTML), 10*time.Second)
		fmt.Printf("%s\n", string(pageHTML))
		os.Exit(3)
	} else if *getImageLink != "" {
		downloadUsingWget(*getImageLink)
		fmt.Println(getActualImageLink((*getImageLink)[16 : len(*getImageLink)-1]))
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
