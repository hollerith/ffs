package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"net/http"
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
	var verbose bool
	var binary bool
	var root string
    var fileCount int
    var matchCount int
    var metaPattern string

	pflag.StringP("file", "f", "", "regex pattern to match file names")
	pflag.StringVarP(&stringPattern,"string", "s", "", "regex pattern to match file string")
	pflag.StringVarP(&hexPattern,"hex", "h", "", "regex pattern to match hex-encoded lines")
	pflag.BoolP("verbose", "v", false, "enable verbose mode")
	pflag.BoolP("binary", "b", false, "include binary files in search")
    pflag.StringVarP(&metaPattern,"meta", "m", "", "regex pattern to match file metadata lines")
    pflag.Parse()

    rootArgs := pflag.Args()
    if len(rootArgs) > 0 {
        root = rootArgs[0]
    } else {
        root = "."
    }

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
	if stringPattern == "" && hexPattern == "" && metaPattern == "" {
		fmt.Println("Error: string, hex, or meta pattern regex is required")
		os.Exit(1)
	}

	var err error
	stringPatternRegex, err = regexp.Compile(stringPattern)
	if err != nil {
		fmt.Println("Error compiling string pattern regex:", err)
		os.Exit(1)
	}

	var hexPatternRegex *regexp.Regexp
	if hexPattern != "" {
		var err error
		hexPatternRegex, err = regexp.Compile(hexPattern)
		if err != nil {
			fmt.Println("Error compiling hex pattern regex:", err)
			os.Exit(1)
		}
	}

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
            if verbose {
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
            if verbose {
                fmt.Printf("Error opening file %s: %v\n", path, err)
            }            
			return nil
		}
		defer file.Close()

		// Check for metadata pattern match
		if metaPattern != "" {
			metaData, err := extractMetadata(file)
			if err != nil {
				if verbose {
					fmt.Printf("Error extracting metadata from file %s: %v\n", path, err)
				}
			}

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
		if !binary {
			// Get file size
			fileInfo, err := file.Stat()
			if err != nil {
                if verbose {
                    fmt.Printf("Error getting file info for %s: %v\n", path, err)
                }                
				return nil
			}

			// Determine number of bytes to read
			numBytes := fileInfo.Size()
			if numBytes > 512 {
				numBytes = 512
			}

			head := make([]byte, numBytes) // read the head
			_, err = file.Read(head)
			if err != nil {
				if verbose {
					fmt.Printf("Error reading file %s: %v\n", path, err)
					return nil
				}
			}
			// check if there are any nulbytes in the head of the file
			if bytes.Contains(head, []byte{0}) {
                if verbose {
                    fmt.Printf("Skipping binary file %s\n", path)
                }        
				return nil
			}
			_, err = file.Seek(0, 0) // reset file pointer to the beginning of the file
			if err != nil {
                if verbose {
                    fmt.Printf("Error resetting file pointer for %s: %v\n", path, err)
                }        
				return nil
			}
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
            if verbose {
                fmt.Printf("Error scanning file %s: %v\n", path, err)
            }            
			return nil
		}

		return nil
	})
	if err != nil {
        if verbose {
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

func extractMetadata(file *os.File) ([]string, error) {
	var metadata []string

	// Extract MIME type
	buf := make([]byte, 512)
	_, err := file.Read(buf)
	if err != nil {
		return metadata, err
	}
	mimeType := http.DetectContentType(buf)
	metadata = append(metadata, fmt.Sprintf("MIME type: %s", mimeType))

	file.Seek(0, 0) // reset file pointer to the beginning of the file

	// Extract EXIF metadata
	exifData, err := exif.Decode(file)
	if err != nil {
		if err == io.EOF {
			return metadata, fmt.Errorf("EOF reached while reading file")
		}
		return metadata, err
	}
	
	// Convert EXIF metadata to JSON string
	jsonByte, err := exifData.MarshalJSON()
	if err != nil {
		return metadata, err
	}
	metadata = append(metadata, fmt.Sprintf("EXIF metadata: %s", string(jsonByte)))

	return metadata, nil
}
