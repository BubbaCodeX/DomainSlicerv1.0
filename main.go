package main

//TODO: ADD A THREAD LIMITER TO THE PROGRAM TO STABILIZE THE RESULT OUTPUTS AND MINIMIZE RACE CONDITIONS
import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

func readHosts() []string {

	args := os.Args
	if len(args) != 2 {
		log.Println("Please enter the correct number of arguments!")
		return nil
	}
	filename, err := filepath.Abs(args[1])
	if err != nil {
		log.Println("Error getting absolute path:", err)
		return nil
	}
	if filename == "" || len(filename) == 0 || filepath.Ext(filename) != ".txt" {
		log.Println("Please provide a valid .txt file name")
		return nil
	}

	if runtime.GOOS == "windows" {
		data, err := os.ReadFile(filename)
		if err != nil {
			log.Println("Error reading file:", err)
			return nil
		}
		return strings.Split(string(data), "\r\n")
	} else {
		data, err := os.ReadFile(filename)
		if err != nil {
			log.Println("Error reading file:", err)
			return nil
		}
		return strings.Split(string(data), "\n")
	}

}

func main() {
	printAsciiArt()
	loadedHosts := readHosts()
	checkStatus(loadedHosts)
}

func printAsciiArt() {
	fmt.Println(`
	····································································································
	:▓█████▄  ▒█████   ███▄ ▄███▓ ▄▄▄       ██▓ ███▄    █   ██████  ██▓     ██▓ ▄████▄ ▓█████  ██▀███  :
	:▒██▀ ██▌▒██▒  ██▒▓██▒▀█▀ ██▒▒████▄    ▓██▒ ██ ▀█   █ ▒██    ▒ ▓██▒    ▓██▒▒██▀ ▀█ ▓█   ▀ ▓██ ▒ ██▒:
	:░██   █▌▒██░  ██▒▓██    ▓██░▒██  ▀█▄  ▒██▒▓██  ▀█ ██▒░ ▓██▄   ▒██░    ▒██▒▒▓█    ▄▒███   ▓██ ░▄█ ▒:
	:░▓█▄   ▌▒██   ██░▒██    ▒██ ░██▄▄▄▄██ ░██░▓██▒  ▐▌██▒  ▒   ██▒▒██░    ░██░▒▓▓▄ ▄██▒▓█  ▄ ▒██▀▀█▄  :
	:░▒████▓ ░ ████▓▒░▒██▒   ░██▒ ▓█   ▓██▒░██░▒██░   ▓██░▒██████▒▒░██████▒░██░▒ ▓███▀ ░▒████▒░██▓ ▒██▒:
	: ▒▒▓  ▒ ░ ▒░▒░▒░ ░ ▒░   ░  ░ ▒▒   ▓▒█░░▓  ░ ▒░   ▒ ▒ ▒ ▒▓▒ ▒ ░░ ▒░▓  ░░▓  ░ ░▒ ▒  ░░ ▒░ ░░ ▒▓ ░▒▓░:
	: ░ ▒  ▒   ░ ▒ ▒░ ░  ░      ░  ▒   ▒▒ ░ ▒ ░░ ░░   ░ ▒░░ ░▒  ░ ░░ ░ ▒  ░ ▒ ░  ░  ▒   ░ ░  ░  ░▒ ░ ▒░:
	: ░ ░  ░ ░ ░ ░ ▒  ░      ░     ░   ▒    ▒ ░   ░   ░ ░ ░  ░  ░    ░ ░    ▒ ░░          ░     ░░   ░ :
	:   ░        ░ ░         ░         ░  ░ ░           ░       ░      ░  ░ ░  ░ ░        ░  ░   ░     :
	: ░                                                                        ░            BubbaCode  :
	····································································································`)
}

func checkStatus(hosts []string) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var slice []string
	fmt.Println("SCAN STARTED.....")
	// Create an HTTP client with a timeout
	client := &http.Client{
		Timeout: 25 * time.Second, // Adjust the timeout duration as needed
	}

	for _, host := range hosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			// Parse the URL to extract the hostname
			parsedURL, err := url.Parse(host)
			if err != nil {
				//log.Printf("Error parsing URL: %v\n", err)
				return
			}
			hostname := parsedURL.Hostname()
			// Perform HTTP GET request using the hostname
			resp, err := client.Get("http://" + hostname)
			if err != nil {
				//log.Printf("Error performing GET request to %s: %v\n", hostname, err)
				return
			}
			defer resp.Body.Close()
			statusLine := fmt.Sprintf("%d:%s\n", resp.StatusCode, hostname)
			mu.Lock()
			slice = append(slice, statusLine)
			mu.Unlock()
		}(host)
	}

	wg.Wait()
	sortData(slice)
}

func sortData(toSort []string) {
	hosts := map[string]string{
		"200": "",
		"400": "",
		"401": "",
		"403": "",
		"404": "",
		"429": "",
		"999": "",
	}
	statusCount := make(map[string]int) // Map to store count of each status code

	for i := 0; i < len(toSort); i++ {
		statusCode := strings.Split(toSort[i], ":")[0]
		host := strings.Split(toSort[i], ":")[1]

		switch statusCode {
		case "200", "400", "401", "403", "404", "429", "999":
			hosts[statusCode] += host
			statusCount[statusCode]++
		default:
			hosts["999"] += host
			statusCount["999"]++
		}
	}

	writeToFile(hosts)
	//output results of scan
	for code, count := range statusCount {
		fmt.Printf("%d domains have resolved to status code %s\n", count, code)
	}
	fmt.Println("You can find the results of the scan in the this directory")
}

func writeToFile(sorted map[string]string) {
	for key, value := range sorted {
		if len(value) != 0 {
			filePath := key + ".txt"
			if err := os.WriteFile(filePath, []byte(value+"\n"), 0666); err != nil {
				log.Printf("Error writing to file %s: %s\n", filePath, err)
			} else {
				//test file creation

			}

		}
	}
}
