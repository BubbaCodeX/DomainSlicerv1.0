package network

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/TwiN/go-color"
	"github.com/schollz/progressbar/v3" // Import the progressbar package
	"github.com/spf13/cobra"
)

var (
	hostFilePath string
	hosts            = []string{}
	numWorkers   int = 5
	sortedHosts      = map[int][]string{}
	mutex        sync.Mutex
	scope_target string
	countHosts   = map[string]int{}
)

// ParseCmd represents the Parse command
var ParseCmd = &cobra.Command{
	Use:   "Parse",
	Short: "Begin parsing a list of hosts",
	Long:  `Begin parsing a list of hosts and sorting them based on the http responses.`,

	Run: func(cmd *cobra.Command, args []string) {
		printAsciiArt()
		fmt.Println("SCAN STARTED.....")
		loadFile(handleHostFilePath(hostFilePath)) // load the list of hosts

		// Initialize the progress bar
		bar := progressbar.NewOptions(len(hosts),
			progressbar.OptionEnableColorCodes(true),
			progressbar.OptionSetWidth(10),
			progressbar.OptionSetDescription("[cyan]Processing"),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "=",
				SaucerPadding: " ",
				SaucerHead:    ">",

				BarStart: "[",
				BarEnd:   "]",
			}))

		// Create a channel to send tasks to workers
		taskCh := make(chan string)

		// Create a wait group to wait for all workers to finish
		var wg sync.WaitGroup

		if numWorkers < 1 {
			numWorkers = 1 // Minimum of 1 worker
			fmt.Println("Minimum of 1 worker set")
		}

		// Start workers
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for host := range taskCh {
					if resp, err := SendRequest(host); err != nil {
						// Handle error
						//fmt.Printf("Error sending request to host %s: %v\n", host, err)
						//Increment progress bar
						mutex.Lock()
						bar.Add(1)
						mutex.Unlock()
					} else {
						// Sort data
						sortData(host, resp.StatusCode)
						// Increment progress bar
						mutex.Lock()
						bar.Add(1)
						mutex.Unlock()
					}
				}
			}()
		}

		// Send tasks to workers
		go func() {
			for _, host := range hosts {
				taskCh <- host
			}
			close(taskCh) // Close the task channel after sending all tasks
		}()

		// Wait for all tasks to complete
		wg.Wait()

		printKeyOccurrences()
		createFoldersAndFiles(sortedHosts)
	},
}

func init() {
	ParseCmd.Flags().StringVarP(&hostFilePath, "filepath", "f", "", "Path to a file containing a list of hosts")
	ParseCmd.Flags().IntVarP(&numWorkers, "workers", "w", 10, "Amount of workers to use, default is 10. Higher numbers will use more memory and results may be less reliable.")
	if err := ParseCmd.MarkFlagRequired("filepath"); err != nil {
		fmt.Println("You must specify a filepath")
	}

}

func createClient(host string) (*http.Response, error) {
	timeout := 10 * time.Second
	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get("http://" + handleHostString(host))
	if err != nil {
		return nil, err // Return error
	}
	client.CloseIdleConnections()
	return resp, nil
}

// handle the host string to minimize errors
func handleHostString(host string) string {
	host = strings.TrimSpace(host)
	host = strings.ToLower(host)
	return host
}

// handles the input for the file path the user enters and makes sure the file path is an absolute path
func handleHostFilePath(file string) string {

	if filepath.IsAbs(file) {
		return file
	} else {
		// turn the path to an absolute path
		newPath, err := filepath.Abs(file)
		if err != nil {
			fmt.Println("Please provide a valid file path")
		}
		if !strings.HasSuffix(file, ".txt") {
			fmt.Println("Please provide a valid file path")
		}
		return newPath
	}
}

func loadFile(path string) {
	file, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("There was an error loading the file:", err)
		return
	}
	var osDetect string
	if runtime.GOOS == "windows" {
		osDetect = "\r\n"
	} else {
		osDetect = "\n"
	}
	//DEBUGGING println("loaded file")
	hosts = strings.Split(string(file), osDetect)
}

func SendRequest(host string) (*http.Response, error) {
	parsedURL, parseError := url.Parse(host)
	if parseError != nil {
		fmt.Println("There was a problem parsing the URL: " + host)
		return nil, parseError // Return parse error
	}

	resp, err := createClient(parsedURL.Hostname() + ":" + parsedURL.Port())
	if err != nil {
		// Handle error appropriately
		//fmt.Println("Error connecting to host:" + host)
		return nil, err // Return connection error
	}

	return resp, nil
}

func sortData(host string, statusCode int) {
	mutex.Lock()
	defer mutex.Unlock()
	addHostOccurrences(
		handleHostScope(host),
		countHosts,
	)
	sortedHosts[statusCode] = append(sortedHosts[statusCode], host)
}

func printKeyOccurrences() {

	fmt.Println("\nOccurrences of each key (status code) in sortedHosts:")
	for statusCode, hosts := range sortedHosts {
		fmt.Printf("Status Code %d: %d Hosts found!\n", statusCode, len(hosts))
	}

}
func createFoldersAndFiles(sortedHosts map[int][]string) error {
	mutex.Lock()
	defer mutex.Unlock()
	scope_target = findKeyWithHighestCount(countHosts)
	// Create a directory to store the files
	directory := "./" + scope_target
	err := os.Mkdir(directory, os.ModePerm)
	if err != nil || err == os.ErrExist {
		//fmt.Println("that directory already exists")
	}

	// Iterate over sortedHosts
	for statusCode, hosts := range sortedHosts {
		// Create a folder for each status code
		folderPath := filepath.Join(directory, fmt.Sprintf("%d", statusCode))
		err := os.Mkdir(folderPath, os.ModePerm)
		if err != nil || err == os.ErrExist {
			//fmt.Println(folderPath)
		}
		//DEBUGGING FOR SCOPE TARGET ASSIGNMENT fmt.Println(scope_target)

		// Create a text file inside the folder with the same name as the status code
		filePath := filepath.Join(folderPath, fmt.Sprintf("%s-%d.txt", scope_target, statusCode))

		//fmt.Println("Created file: " + filePath)

		file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err == os.ErrNotExist {
			file, err = os.Create(filePath)
			if err != nil {
				return err
			}

		} else if err != nil {
			return err
		}

		defer file.Close()

		//TODO:  Add seperators in the file name to make it easier to read and tell you what scan resulsts are from what dates
		// Write host names to the text file

		file.WriteString("SCAN DATE: " + time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC).Format("2006-01-02") + "\n")

		for _, host := range hosts {
			_, err := file.WriteString(host + "\n")
			if err != nil {
				fmt.Println(err)
				return err
			}

		}
		file.WriteString("-------------------------------------------------------------------------------------------------------------\n")
	}
	fmt.Println("Results have been written to the DomainSlicer directory!")
	return nil
}

func findKeyWithHighestCount(data map[string]int) string {
	var highestCount int
	var keyWithHighestCount string

	for key, count := range data {
		if count > highestCount {
			highestCount = count
			keyWithHighestCount = key
		}
	}

	return keyWithHighestCount
}
func handleHostScope(host string) string {
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimPrefix(host, "https://") // this shouldnt ever be needed but just in case
	host = strings.TrimPrefix(host, "www.")     // this shouldnt ever be needed but just in case
	host = strings.Split(host, ".com")[0]
	if strings.Count(host, ".") > 0 {
		host = strings.Split(host, ".")[strings.Count(host, ".")]
		return host
	}
	return host
}

func addHostOccurrences(host string, hostMap map[string]int) {
	// Normalize the host string to ensure consistency (e.g., case-insensitive)
	host = strings.ToLower(host)

	// Check if the host already exists in the map
	if _, ok := hostMap[host]; ok {
		// Increment the count if the host exists
		hostMap[host]++
	} else {
		// Add the host to the map with an initial count of 1 if it doesn't exist
		hostMap[host] = 1
	}
}
func printAsciiArt() {

	fmt.Println(color.InPurpleOverBlue(`
	··································································································
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
	····································································..............................`))
}
