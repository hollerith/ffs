package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/pflag"
)

func main() {
	var filePatternRegex *regexp.Regexp
	var contentsPatternRegex *regexp.Regexp
	var verbose bool
	var binary bool
	var root string // variable to hold the value of the root parameter

	pflag.StringP("file", "f", "", "regex pattern to match file names")
	pflag.StringP("match", "m", "", "regex pattern to match file contents")
	pflag.BoolP("verbose", "v", false, "enable verbose mode")
	pflag.BoolP("binary", "b", false, "include binary files in search")
	pflag.StringVarP(&root, "root", "r", ".", "root directory to start the search from") // add the root parameter
	pflag.Parse()

	filePattern := pflag.Lookup("file").Value.String()
	contentsPattern := pflag.Lookup("match").Value.String()
	verbose = pflag.Lookup("verbose").Value.String() == "true"
	binary = pflag.Lookup("binary").Value.String() == "true"

	if filePattern != "" {
		var err error
		filePatternRegex, err = regexp.Compile(filePattern)
		if err != nil {
			fmt.Println("Error compiling file pattern regex:", err)
			os.Exit(1)
		}
	}
	if contentsPattern == "" {
		fmt.Println("Error: contents pattern regex is required")
		os.Exit(1)
	}
	var err error
	contentsPatternRegex, err = regexp.Compile(contentsPattern)
	if err != nil {
		fmt.Println("Error compiling contents pattern regex:", err)
		os.Exit(1)
	}

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error processing file %s: %v\n", path, err)
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if filePatternRegex != nil && !filePatternRegex.MatchString(path) {
			return nil
		}

		if verbose {
			fmt.Println("Matching file:", path)
		}

		file, err := os.Open(path)
		if err != nil {
			fmt.Printf("Error opening file %s: %v\n", path, err)
			return nil
		}
		defer file.Close()

		// Check if file is binary and skip if not set to include binary files
		if !binary {
			head := make([]byte, 512) // read the first 512 bytes of the file
			_, err = file.Read(head)
			if err != nil {
				fmt.Printf("Error reading file %s: %v\n", path, err)
				return nil
			}
			// check if there are any nulbytes in the head of the file
			if bytes.Contains(head, []byte{0}) {
				return nil
			}
			_, err = file.Seek(0, 0) // reset file pointer to the beginning of the file
			if err != nil {
				fmt.Printf("Error resetting file pointer for %s: %v\n", path, err)
				return nil
			}
		}

		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // set buffer size to 1MB
		lineNumber := 1
		for scanner.Scan() {
			line := scanner.Text()
			if contentsPatternRegex.MatchString(line) {
				fmt.Printf("%s:%d:%s\n", path, lineNumber, line)
			}
			lineNumber++
		}
		if err := scanner.Err(); err != nil {
			fmt.Printf("Error scanning file %s: %v\n", path, err)
			return nil
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directories: %v\n", err)
	}
}
