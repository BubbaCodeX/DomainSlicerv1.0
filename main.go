package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

func readHosts() []string {

	filename := os.Args[1]
	if (filename == "" || len(filename) == 0) || strings.Split(filename, ".")[1] != "txt" {
		log.Fatal("Please provide a valid .txt file name")
	}

	if runtime.GOOS == "windows" {
		data, err := os.ReadFile(filename)
		if err != nil {
			panic("error reading " + filename)
		}
		return strings.Split(string(data), "\r\n")
	} else {
		data, err := os.ReadFile("uncheckedHosts.txt")
		if err != nil {
			panic("error reading uncheckedHosts.txt")
		}
		return strings.Split(string(data), "\n")
	}
}
func main() {

	//createFiles()
	loadedHosts := readHosts()
	//loadUncheckedHosts()
	checkStatus(loadedHosts)
}

func checkStatus(hosts []string) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var slice []string

	// Create an HTTP client with a timeout
	client := &http.Client{
		Timeout: 5 * time.Second, // Adjust the timeout duration as needed
	}

	for _, host := range hosts {
		wg.Add(1)
		if strings.Contains(host, "http://") || strings.Contains(host, "https://") {
			host = strings.Split(host, "/")[2]
		}
		go func(host string) {
			defer wg.Done()
			resp, err := client.Get("http://" + strings.ReplaceAll(host, "\t", ""))
			if err != nil {
				log.Println(err.Error())
				return
			}
			defer resp.Body.Close()
			mu.Lock()
			defer mu.Unlock()
			slice = append(slice, fmt.Sprintf("%d:%s\n", resp.StatusCode, host))
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
		"500": "",
		"999": "",
	}
	for i := 0; i < len(toSort); i++ {
		switch strings.Split(toSort[i], ":")[0] {
		case "200":
			hosts["200"] = (hosts["200"] + strings.Split(toSort[i], ":")[1])
		case "400":
			hosts["400"] = (hosts["400"] + strings.Split(toSort[i], ":")[1])
		case "401":
			hosts["401"] = (hosts["401"] + strings.Split(toSort[i], ":")[1])
		case "403":
			hosts["403"] = (hosts["403"] + strings.Split(toSort[i], ":")[1])
		case "404":
			hosts["404"] = (hosts["404"] + strings.Split(toSort[i], ":")[1])
		case "500":
			hosts["500"] = (hosts["500"] + strings.Split(toSort[i], ":")[1])
		default:
			hosts["999"] = (hosts["999"] + strings.Split(toSort[i], ":")[1])
		}
	}
	writeToFile(hosts)
}

func writeToFile(sorted map[string]string) {
	for key, value := range sorted {
		if len(value) != 0 {
			filePath := key + ".txt"
			if err := os.WriteFile(filePath, []byte(value), 0666); err != nil {
				log.Printf("Error writing to file %s: %s\n", filePath, err)
			} else {
				fmt.Printf("File %s created successfully.\n", filePath)
			}
		}
	}
}

func createFiles() {
	statusCodes := []string{"200", "401", "403", "404", "500", "unknown"}
	for _, code := range statusCodes {
		file, err := os.Open(code + ".txt")
		if errors.Is(err, os.ErrNotExist) {
			createfile, err := os.Create(code + ".txt")
			if err != nil {
				log.Fatal(err)
			}
			createfile.Close()
		}
		file.Close()
	}
}
