package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"net/http"
	"strings"
	"io"

	"github.com/rwcarlsen/goexif/exif"

	"github.com/spf13/pflag" 
)

func main() {
	var filePatternRegex *regexp.Regexp
	var stringPatternRegex *regexp.Regexp
    var filePattern string
    var stringPattern string
    var hexPattern string
    var metaPattern string
	var verbose bool
	var binary bool
	var errors bool
	var root string
    var fileCount int
    var matchCount int
	var err error

	pflag.StringP("file", "f", "", "regex pattern to match file names")
	pflag.StringVarP(&stringPattern,"string", "s", "", "regex pattern to match file string")
	pflag.StringVarP(&hexPattern,"hex", "h", "", "regex pattern to match hex-encoded lines")
    pflag.StringVarP(&metaPattern,"meta", "m", "", "regex pattern to match file metadata lines")
	pflag.BoolP("verbose", "v", false, "enable verbose mode")
	pflag.BoolP("binary", "b", false, "include binary files in search")
	pflag.BoolP("errors", "e", false, "print errors encountered during execution")
	pflag.Parse()
	
    pflag.Parse()

    rootArgs := pflag.Args()
    if len(rootArgs) > 0 {
        root = rootArgs[0]
    } else {
        root = "."
    }

	verbose = pflag.Lookup("verbose").Value.String() == "true"
	binary = pflag.Lookup("binary").Value.String() == "true"
	errors = pflag.Lookup("errors").Value.String() == "true"

	if filePattern != "" {
		filePatternRegex, err = regexp.Compile(filePattern)
		if err != nil {
			fmt.Println("Error compiling file pattern regex:", err)
			os.Exit(1)
		}
	}
	if stringPattern == "" && hexPattern == "" && metaPattern == "" {
		fmt.Println("Error: string, hex, or meta pattern regex is required")
		os.Exit(1)
	}

	stringPatternRegex, err = regexp.Compile(stringPattern)
	if err != nil {
		fmt.Println("Error compiling string pattern regex:", err)
		os.Exit(1)
	}

	var hexPatternRegex *regexp.Regexp
	if hexPattern != "" {
		hexPatternRegex, err = regexp.Compile(hexPattern)
		if err != nil {
			fmt.Println("Error compiling hex pattern regex:", err)
			os.Exit(1)
		}
	}

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
            if errors {
                fmt.Printf("Error processing file %s: %v\n", path, err)
            }    
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
			fileCount++
		}

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
		if metaPattern != "" {
			var metaPatternRegex *regexp.Regexp
			metaPatternRegex, err = regexp.Compile(metaPattern)
			if err != nil {
				fmt.Println("Error compiling metadata pattern regex:", err)
				os.Exit(1)
			}

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

		return nil
	})
	if err != nil {
        if errors {
            fmt.Printf("Error walking directories: %v\n", err)
        }        
	}

    if verbose {
        fmt.Println("Options chosen:")
        if filePattern != "" {
            fmt.Println("- file pattern:", filePattern)
        }
        if stringPattern != "" {
            fmt.Println("- string pattern:", stringPattern)
        }
        if hexPattern != "" {
            fmt.Println("- hex pattern:", hexPattern)
        }
        if metaPattern != "" {
            fmt.Println("- meta pattern:", metaPattern)
        }
        fmt.Println("- root directory:", root)
        fmt.Println("- binary files:", binary)
        fmt.Println("Number of matches:")
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
	if err != nil {
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
	metadata = append(metadata, fmt.Sprintf("MIME type: %s", mimeType))

	// Check if MIME type belongs to a group of known binary file types
	if strings.HasPrefix(mimeType, "application/octet-stream") ||
		strings.HasPrefix(mimeType, "application/pdf") ||
		strings.HasPrefix(mimeType, "image/") {
		isBinary = true
	}

	file.Seek(0, 0) // reset file pointer to the beginning of the file

	// Extract EXIF metadata
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
	metadata = append(metadata, fmt.Sprintf("EXIF metadata: %s", string(jsonByte)))

	return metadata, isBinary, nil
}

