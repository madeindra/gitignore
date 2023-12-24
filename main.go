package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// github api to get the list of gitignore files
const contentURL = "https://api.github.com/repos/github/gitignore/contents"

func main() {
	// parse flags
	flag.Parse()

	// check if there are too many arguments
	if flag.NArg() > 1 {
		fmt.Println("Error: too many arguments")
		return
	}

	// get the latest list of files
	files, err := getList()
	if err != nil {
		fmt.Println("Error fetching file list:", err)
		return
	}

	// filter the list of files
	filteredFiles := filterList(files)

	// if the argument is 1, check if the file exists on the list
	found := false
	if flag.NArg() == 1 {
		for _, filteredFile := range filteredFiles {
			if strings.EqualFold(strings.TrimSpace(filteredFile.Name), strings.TrimSpace(flag.Arg(0))) {
				// download and save the file
				if err := downloadFile(filteredFile); err != nil {
					fmt.Println("Error downloading file:", err)
					return
				}

				// print success message
				fmt.Printf("Successfully downloaded .gitignore for \"%v\"", filteredFile.Name)
				return
			}
		}
	}

	// print the list of files
	fmt.Println("Available gitignore files:")
	for i, filteredFile := range filteredFiles {
		// print the index and the name of the file
		fmt.Printf("%d. %s\n", i+1, filteredFile.Name)
	}

	// when the argument is 1, the file was not found, warn the user
	if !found {
		fmt.Printf("\nWarning: gitignore file for \"%v\" not found\n", flag.Arg(0))
	}

	// wait for user input
	var selected int
	fmt.Print("Please enter the number of your choice: ")
	fmt.Scanln(&selected)

	// check if the choice is valid
	if selected < 1 || selected > len(filteredFiles) {
		fmt.Println("Error: invalid choice")
		return
	}

	// download and save the file
	if err := downloadFile(filteredFiles[selected-1]); err != nil {
		fmt.Println("Error downloading file:", err)
		return
	}

	// print success message
	fmt.Printf("Successfully downloaded .gitignore for %v", filteredFiles[selected-1].Name)
}

// struct for github api response
type File struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	DownloadURL string `json:"download_url"`
}

func getList() ([]File, error) {
	// get the list of files from github
	resp, err := http.Get(contentURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var files []File
	if err = json.Unmarshal(body, &files); err != nil {
		return nil, err
	}

	return files, nil
}

func filterList(files []File) []File {
	// filter the list of files
	var filteredFiles []File
	for _, file := range files {
		// only include files that are not directories and end with .gitignore
		if file.Type != "dir" && strings.HasSuffix(file.Name, ".gitignore") {
			filteredFiles = append(filteredFiles, file)
		}
	}

	return filteredFiles
}

func downloadFile(file File) error {
	// get the file from github
	resp, err := http.Get(file.DownloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// create the file if it doesn't exist
	newFile, err := os.Create(".gitignore")
	if err != nil {
		return err
	}
	defer newFile.Close()

	// write the file
	if _, err = io.Copy(newFile, resp.Body); err != nil {
		return err
	}

	return nil
}
