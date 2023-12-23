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

	var filteredFiles []File
	for _, file := range files {
		// if the file is a gitignore file, add it to the list
		if file.Type == "file" && strings.HasSuffix(file.Name, ".gitignore") {
			// remove the .gitignore extension and add it to the list
			filteredFiles = append(filteredFiles, File{
				Name:        strings.TrimSuffix(file.Name, ".gitignore"),
				Type:        file.Type,
				DownloadURL: file.DownloadURL,
			})
		}
	}

	// if no gitignore files found, exit
	if len(filteredFiles) == 0 {
		fmt.Println("Error: cannot get list of gitignore files")
		return
	}

	// if the argument is 1, check if the file exists on the list
	found := true
	if flag.NArg() == 1 {
		// check if the file exists on the list
		for _, filteredFile := range filteredFiles {
			if strings.EqualFold(strings.TrimSpace(filteredFile.Name), strings.TrimSpace(flag.Arg(0))) {
				// download and save the file
				if err := downloadFile(filteredFile); err != nil {
					fmt.Println("Error downloading file:", err)
					return
				}

				// print success message
				fmt.Printf("Successfully downloaded .gitignore for %v", filteredFile.Name)
				return
			}
		}

		found = false
	}

	// print the list of files
	fmt.Println("Available gitignore file:")
	for i, filteredFile := range filteredFiles {
		// print the index and the name of the file
		fmt.Printf("%d. %s\n", i+1, filteredFile.Name)
	}

	// if the file is not found, print prompt
	if !found {
		fmt.Printf("\nGitignore file for %v not found\n", flag.Arg(0))
	}

selection:
	// wait for user input
	var selected int
	fmt.Print("Please enter the number of your choice: ")
	fmt.Scanln(&selected)

	// check if the choice is valid
	if selected < 1 || selected > len(filteredFiles) {
		fmt.Println("Error: invalid choice")
		goto selection
	}

	// download and save the file
	if err := downloadFile(filteredFiles[selected-1]); err != nil {
		fmt.Println("Error downloading file:", err)
		return
	}

	// print success message
	fmt.Printf("Successfully downloaded .gitignore for %v", filteredFiles[selected-1].Name)
}

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
