package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
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
	predefinedImageBuffer  []string // dynamic slice of image IDs
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

func generateRandomImageID() (string, int) {
	var newImageId string
	var iter = 0
	for {
		newImageId = ""
		for i := 0; i < 6; i++ {
			newImageId = newImageId + imageIDcharacters[rand.Intn(len(imageIDcharacters))]
		}

		if checkEntireLogFile(newImageId) {
			break
		} else {
			iter++
		}
	}

	return newImageId, iter
}

func downloadUsingWget(url string) {
	url = eliminateNewLineCrap(url)
	downloadLightshotCommand := exec.Command("wget", url)
	_, err := downloadLightshotCommand.Output()
	if err != nil {
		fmt.Println("error downloading file")
	}
}

func readfile(filename string) (string, []string) {
	var contentAsSingleString string
	var contentLineByLine []string

	file, _ := os.Open(filename)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		contentLineByLine = append(contentLineByLine, scanner.Text())
		contentAsSingleString = contentAsSingleString + (scanner.Text() + "\n")
	}

	return contentAsSingleString, contentLineByLine
}

func checkIfFileExists(filename string) bool {
	_, err := os.Stat(filename)
	for {
		_, err = os.Stat(filename)
		if os.IsExist(err) {
			break
		}
		fmt.Println("waiting for file")
	}
	return true
}

func getPageHTML(url string) (content string) {
	// download file using wget
	downloadUsingWget(url)

	// read file content
	filecontent, _ := readfile(eliminateNewLineCrap(url[16:]))

	// delete file
	os.Remove(eliminateNewLineCrap(url[16:]))

	return filecontent
}

func downloadImageByURL(URL string, lightshotID string) error {
	//Get the response bytes from the url
	URL = eliminateNewLineCrap(URL)
	imageFormat := URL[len(URL)-4:]
	if imageFormat[0] != '.' {
		imageFormat = "." + imageFormat
	}

	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("Received non 200 response code")
	}
	//Create a empty file
	file, err := os.Create(pathToDownloadedImages + lightshotID + imageFormat)
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

func GetStringInBetweenTwoString(str string, startS string, endS string) (result string, found bool) {
	s := strings.Index(str, startS)
	if s == -1 {
		return result, false
	}
	newS := str[s+len(startS):]
	e := strings.Index(newS, endS)
	if e == -1 {
		return result, false
	}
	result = newS[:e]
	return result, true
}

func getActualImageLink(parsedhtml string) (string, bool) {
	// 1. get string between those values
	/*
		example #1: <img class="no-click screenshot-image" src="https://image.prntscr.com/image/74j1YlFTSn61-yVmpszTlw.png" crossorigin="anonymous" alt="Lightshot screenshot" id="screenshot-image" image-id="r68h04">
		example #2: https://i.imgur.com/
	*/
	imageURL, isFound := trimStringBetweenTwo(parsedhtml, "https://i.imgur.com/", " ")

	if len(imageURL) < 6 || !isFound {
		imageURL = ""
	} else if isFound {
		imageURL = ("https://i.imgur.com/" + imageURL[:len(imageURL)-3])
	}

	return imageURL, isFound
}

func downloadImageByLightshotID(imageID string) (bool, error) {
	// 0. create image URL
	imageID = eliminateNewLineCrap(imageID)
	var imageLightshotURL = "https://prnt.sc/" + imageID
	var error error

	actualImageURL, isFound := getActualImageLink(getPageHTML(eliminateNewLineCrap(imageLightshotURL)))
	if isFound {
		error = downloadImageByURL(actualImageURL, imageID)
		if error != nil {
			fmt.Println("error downloading image")
			os.Exit(3)
		}
	} else {
		isFound = false
	}

	return isFound, error
}

func checkEntireLogFile(input string) (nomatch bool) {
	_, logSlice := readfile(pathToLogFile)
	currentlog := make([]string, len(logSlice))
	copy(currentlog, logSlice)
	nomatch = true

	for i := 0; i < len(currentlog); i++ {
		nomatch = nomatch && !(currentlog[i] == input)
	}

	return nomatch
}

func predefineImageBuffer(imagecount int) {
	_, logSlice := readfile(pathToLogFile)
	currentlog := make([]string, len(logSlice))
	copy(currentlog, logSlice)

	var newImageID string
	//var iterationsNeeded int

	for i := 0; i < imagecount; i++ {
		newImageID, _ = generateRandomImageID()

		logLength := len(currentlog)
		switch logLength {
		case 0:
			predefinedImageBuffer = append(predefinedImageBuffer, newImageID)
			previousImageID = newImageID
		default:
			if newImageID != previousImageID && newImageID != "" {
				predefinedImageBuffer = append(predefinedImageBuffer, newImageID)
				previousImageID = newImageID
			} else {
				i = i - 1
			}
		}
	}
}

func recordDownloadedImageID(imageID string) {
	logfile, openerr := os.OpenFile(pathToLogFile, os.O_APPEND|os.O_WRONLY, 0644)

	//_, logerr := logfile.WriteString("https://prnt.sc/" + imageID + "\n")
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
	specifyImageURL := flag.String("imageurl", "", "you can specify image url here, e.g. --imageurl https://i.imgur.com/y5Onqkp.png")
	readPageHTML := flag.String("readhtml", "", "read page html of the specified address, e.g. --readhtml https://i.imgur.com/y5Onqkp.png")
	getImageLink := flag.String("getimglink", "", "get actual image link, e.g. --getimglink https://prnt.sc/2a93m0")
	//slowMode := flag.Bool("slowmode", false, "use slowmode to pre-define image id before downloading them")
	flag.Parse()

	if *readPageHTML != "" {
		fmt.Printf("%s\n", getPageHTML(eliminateNewLineCrap(*readPageHTML)))
		os.Exit(3)
	} else if *getImageLink != "" {
		fmt.Println(getActualImageLink(getPageHTML(eliminateNewLineCrap(*getImageLink))))
	} else if *specifyImageURL != "" {
		fmt.Print("you provided image url: ")
		fmt.Println(*specifyImageURL)
		downloadImageByURL(*specifyImageURL, "imageIDshouldBeHere")
	} else if *specifyImageURL == "" && *howManyImages != 0 {
		var totalImagesDownloaded int
		fmt.Print("you specified images: ")
		fmt.Println(*howManyImages)
		predefineImageBuffer(*howManyImages)
		downloadingProgressBar := progressbar.DefaultBytes(
			(int64(*howManyImages)),
			"Downloading",
		)
		for i := 0; i < (*howManyImages); i++ {
			recordDownloadedImageID(predefinedImageBuffer[i])
			isFound, _ := downloadImageByLightshotID(predefinedImageBuffer[i])
			downloadingProgressBar.Add(1)

			if isFound {
				totalImagesDownloaded = totalImagesDownloaded + 1
				isFound = false
			}
		}

		fmt.Print("Success rate: ")
		fmt.Print(100 * float32(totalImagesDownloaded) / float32(*howManyImages))
		fmt.Println(" %")

	} else {
		fmt.Println("no command provided")
		os.Exit(3)
	}
}
