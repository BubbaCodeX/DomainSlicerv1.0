package main

//TODO: Can't write to file every single request, too costly. Write to all files when all requests are done until we can find a more optimized way to do it.
//TODO: Can we create a slice that we can just append the hosts with status codes attached to them? and maybe change the create file function to take this slice and create files based on the status codes that we have or maybe we
// can keep what we have and just try that at some point if we have issues with memory.
import (
	//"bufio"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func readHosts() []string {
	data, err := os.ReadFile("uncheckedHosts.txt")
	if err != nil {
		panic("error reading uncheckedHosts.txt")
	}
	sliced := strings.Split(string(data), "\r\n")
	//log.Println(sliced[1])
	return sliced
}

func main() {
	//createFiles()
	loadedHosts := readHosts()
	//loadUncheckedHosts()
	checkStatus(loadedHosts)
}

// REUSABLE REQUEST
func checkStatus(hosts []string) (slice []string) {

	for i := 0; i < len(hosts); i++ {
		resp, err := http.Get("http://" + strings.ReplaceAll(hosts[i], "\t", ""))
		if err != nil {
			log.Println(err.Error())
			continue
		}
		defer resp.Body.Close()
		//append to slice
		//slice = append(slice, ("|"+fmt.Sprint(resp.StatusCode) + "|" + hosts[i]))
		slice = append(slice, (fmt.Sprint(resp.StatusCode)+":"+hosts[i])+"\n")
	}

	sortData(slice)
	return slice
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
			hosts["200"] = (hosts["200"] + strings.Split(toSort[i], ":")[1] )
		case "400":
			hosts["400"] = (hosts["400"] + strings.Split(toSort[i], ":")[1] )
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
	/* for _, value := range hosts{
		fmt.Println(value)
	} */
}

func writeToFile(sorted map[string]string) {
	for key,value := range sorted{
	if len(value) !=0{
		if err := os.WriteFile(key+"txt", []byte(value), 0666); err != nil {
			log.Fatal(err)
		}
	}
	 /* err := os.WriteFile(key+".txt", []byte(value), 0666)
	 if err != os.ErrNotExist {
		log.Println("file not found...creating file")
		createFiles()
		continue
	 } */
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

