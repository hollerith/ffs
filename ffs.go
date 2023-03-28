package main

import "net/http"
import _ "net/http/pprof"

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/spf13/pflag"

	"github.com/sabhiram/go-gitignore"
)

func main() {
	verbose, binary, errors, debugging, root, filePatternRegex, stringPatternRegex, hexPatternRegex, metaPatternRegex, ignoreParser := parseFlags()

	var fileCount int
	var matchCount int

	// Search
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if errors {
				fmt.Printf("Error processing file %s: %v\n", path, err)
			}
			return nil
		} else {
			if debugging {
				fmt.Printf("%s\n", path)
			}
		}

		if info.IsDir() {
			return nil
		}

        // Only search files according to .gitignore
        if ignoreParser != nil && ignoreParser.MatchesPath(path) {
            return nil
        }

		// Match file regex pattern
		if filePatternRegex != nil && !filePatternRegex.MatchString(path) {
			return nil
		}

		// Open the file for reading
		file, err := os.Open(path)
		if err != nil {
			if errors {
				fmt.Printf("Error opening file %s: %v\n", path, err)
			}
			return nil
		}
		defer file.Close()

		// Extract metadata and other file information
		metaData, isBinary, err := extractFileData(file)
		if err != nil {
			if errors {
				fmt.Printf("Error extracting information from the file %s: %v\n", path, err)
			}
		}

		// Check for metadata pattern match
		if metaPatternRegex != nil {
			for _, line := range metaData {
				if metaPatternRegex.MatchString(line) {
					matchCount++
					fmt.Printf("%s:%s\n", path, line)
					break
				}
			}
		}

		// Check if file is binary and skip if not set to include binary files
		if !binary && isBinary {
			if errors {
				fmt.Printf("Skipping binary file %s\n", path)
			}
			return nil
		}

		// Scan each line of the file content
		if stringPatternRegex != nil {
			file.Seek(0, 0) // reset file pointer to the beginning of the file
			scanner := bufio.NewScanner(file)
			scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // set buffer size to 1MB
			lineNumber := 1
			for scanner.Scan() {
				line := scanner.Text()
				var match bool
				if hexPatternRegex != nil {
					// Convert line to hex string and perform match on hex string
					hex := ""
					for _, b := range line {
						hex += " " + strconv.FormatInt(int64(b), 16)
					}
					match = hexPatternRegex.MatchString(hex)
				} else {
					match = stringPatternRegex.MatchString(line)
				}
				if match {
					matchCount++
					if verbose {
						fmt.Printf("%s:%d:%s\n", path, lineNumber, replaceNonPrintable(line))
					} else {
						fmt.Printf("%s\n", path)
					}
				}
				lineNumber++
			}
			if err := scanner.Err(); err != nil {
				if errors {
					fmt.Printf("Error scanning file %s: %v\n", path, err)
				}
				return nil
			}
		}

		// Print the path of every file scanned
		if verbose || (stringPatternRegex == nil && hexPatternRegex == nil && metaPatternRegex == nil) {
			fmt.Println(metaData, path)
			fileCount++
		}
		return nil
	})
	if err != nil {
		if errors {
			fmt.Printf("Error walking directories: %v\n", err)
		}
	}

	if verbose || (stringPatternRegex == nil && hexPatternRegex == nil && metaPatternRegex == nil) {
		fmt.Println("\nNumber of matches:")
		fmt.Println("- files:", fileCount)
		fmt.Println("- matches:", matchCount)
	}
}

func replaceNonPrintable(s string) string {
	b := []byte(s)
	for i, c := range b {
		if !strconv.IsPrint(rune(c)) {
			b[i] = '.'
		}
	}
	return string(b)
}

func extractFileData(file *os.File) ([]string, bool, error) {
	var metadata []string
	isBinary := false

	// Get file size
	fileInfo, err := file.Stat()
	if err == nil {
		mode := os.FileMode(fileInfo.Mode())
		size := fileInfo.Size()
		metadata = append(metadata, fmt.Sprintf("\U0001f512 %s", mode.String()))
		metadata = append(metadata, fmt.Sprintf("\U0001f4be %12d ", size))
	} else {
		return metadata, isBinary, nil
	}

	// Determine number of bytes to read
	numBytes := fileInfo.Size()
	if numBytes > 512 {
		numBytes = 512
	}

	// Extract MIME type
	buf := make([]byte, numBytes)
	_, err = file.Read(buf)
	if err != nil {
		return metadata, isBinary, err
	}
	mimeType := http.DetectContentType(buf)
	metadata = append(metadata, fmt.Sprintf("\U0001f4c4 %s \t", mimeType))

	// Check if MIME type belongs to a group of known binary file types
	if strings.HasPrefix(mimeType, "application/octet-stream") ||
		strings.HasPrefix(mimeType, "application/pdf") ||
		strings.HasPrefix(mimeType, "image/") {
		isBinary = true
	}

	// If the file is not an image type return without exifdata
	if !strings.HasPrefix(mimeType, "image/") {
		return metadata, isBinary, nil
	}

	// Extract EXIF metadata
	file.Seek(0, 0) // reset file pointer to the beginning of the file

	exifData, err := exif.Decode(file)
	if err != nil {
		if err == io.EOF {
			return metadata, isBinary, fmt.Errorf("EOF reached while reading file")
		}
		return metadata, isBinary, err
	}

	// Convert EXIF metadata to JSON string
	jsonByte, err := exifData.MarshalJSON()
	if err != nil {
		return metadata, isBinary, err
	}
	metadata = append(metadata, fmt.Sprintf("\U0001f4f7 %s", string(jsonByte)))

	return metadata, isBinary, nil
}

func parseFlags() (bool, bool, bool, bool, string, *regexp.Regexp, *regexp.Regexp, *regexp.Regexp, *regexp.Regexp, ignore.IgnoreParser) {
	var filePattern, stringPattern, hexPattern, metaPattern string
	var verbose, binary, errors, gitPattern, debugging bool
	var root string
	var ignoreParser ignore.IgnoreParser

	pflag.StringVarP(&filePattern, "file", "f", "", "regex pattern to match file names")
	pflag.StringVarP(&stringPattern, "string", "s", "", "regex pattern to match file string")
	pflag.StringVarP(&hexPattern, "hex", "h", "", "regex pattern to match hex-encoded lines")
	pflag.StringVarP(&metaPattern, "meta", "m", "", "regex pattern to match file metadata lines")

	pflag.BoolVarP(&verbose, "verbose", "v", false, "enable verbose mode")
	pflag.BoolVarP(&binary, "binary", "b", true, "include binary files in search")
	pflag.BoolVarP(&errors, "errors", "e", false, "print errors encountered during execution")
	pflag.BoolVarP(&debugging, "debugging", "d", false, "set debugging and trace during execution")
	pflag.BoolVarP(&gitPattern, "gitignore", "g", false, "search according to .gitignore")
	pflag.Parse()

	rootArgs := pflag.Args()
	if len(rootArgs) > 0 {
		root = rootArgs[0]
		_, err := os.Stat(root)
		if os.IsNotExist(err) {
			fmt.Printf("Error: root directory '%s' does not exist.\n", root)
			os.Exit(1)
		}
	} else {
		root = "."
	}

	if debugging {
		debug.SetGCPercent(25)
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	var filePatternRegex, stringPatternRegex, hexPatternRegex, metaPatternRegex *regexp.Regexp
	var err error

	if filePattern != "" {
		filePatternRegex, err = regexp.Compile(filePattern)
		if err != nil {
			fmt.Printf("Error compiling file pattern regex: %v\n", err)
			os.Exit(1)
		}
	}

	if stringPattern != "" {
		stringPatternRegex, err = regexp.Compile(stringPattern)
		if err != nil {
			fmt.Printf("Error compiling string pattern regex: %v\n", err)
			os.Exit(1)
		}
	}

	if hexPattern != "" {
		hexPatternRegex, err = regexp.Compile(hexPattern)
		if err != nil {
			fmt.Printf("Error compiling hex pattern regex: %v\n", err)
			os.Exit(1)
		}
	}

	if metaPattern != "" {
		metaPatternRegex, err = regexp.Compile(metaPattern)
		if err != nil {
			fmt.Printf("Error compiling metadata pattern regex: %v\n", err)
			os.Exit(1)
		}
	}

	if gitPattern {
		ignoreFilePath := filepath.Join(root, ".gitignore")
		ignoreParser, err = ignore.CompileIgnoreFile(ignoreFilePath)
		if err != nil {
			fmt.Printf("Error parsing .gitignore file: %v\n", err)
			os.Exit(1)
		}
	}

	return verbose, binary, errors, debugging, root, filePatternRegex, stringPatternRegex, hexPatternRegex, metaPatternRegex, ignoreParser
}
