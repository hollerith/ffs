package main

import "net/http"
import _ "net/http/pprof"

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/spf13/pflag"

	"github.com/sabhiram/go-gitignore"
)

type Metadata struct {
	Size     int64
	Mode     string
	Suid     bool
	Owner    string
	Group    string
	ModTime  string
	MimeType string
	ExifData string
}

const (
	sizeWidth      = 10
	modeWidth      = 12
	ownerWidth     = 12
	groupWidth     = 12
	timeWidth      = 19
	mimeTypeWidth  = 30
	pathWidth      = 50
	truncateLength = 3
)

func main() {
	verbose, binary, errors, debugging, links, root, depth, filePatternRegex, stringPatternRegex, hexPatternRegex, metaPatternRegex, ignoreParser := parseFlags()

	var lastDir string
	var fileCount int
	var matchCount int
	var byteCount int64

	// Search
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		var lastCount = matchCount
		directory, filename := filepath.Split(path)

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

		// Check if file is a directory and if depth is reached
		if info.IsDir() && links {
			relPath, err := filepath.Rel(root, path)
			if err != nil {
				if errors {
					fmt.Printf("Error getting relative path for directory %s: %v\n", path, err)
				}
				return nil
			}
			if depth >= 0 && strings.Count(relPath, string(os.PathSeparator)) >= depth && relPath != "." {
				return filepath.SkipDir
			}
			return nil
		}

		// Only search files according to .gitignore
		if ignoreParser != nil && ignoreParser.MatchesPath(path) {
			return nil
		}

		// Match filename regex pattern, optional TODO add a flag to match whole path
		if filePatternRegex != nil && !filePatternRegex.MatchString(filename) {
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
			return nil
		}

		// Check for metadata pattern match
		if metaPatternRegex != nil {
			metadataString := fmt.Sprintf("%d %s %s %s %s %s %s", metaData.Size, metaData.Mode, metaData.Owner, metaData.Group, metaData.ModTime, metaData.MimeType, metaData.ExifData)
			if metaPatternRegex.MatchString(metadataString) {
				matchCount++
			}
		}

		byteCount += metaData.Size

		// Check if file is binary and skip if not set to include binary files
		if !binary && isBinary {
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
		if (matchCount > lastCount) || (stringPatternRegex == nil && hexPatternRegex == nil && metaPatternRegex == nil) {
			if fileCount == 0 || directory != lastDir {
				lastDir = directory
				fmt.Printf("\n\033[32m%s\033[0m:\n", lastDir)
			}

			sizeStr := fmt.Sprintf("%*d", sizeWidth, metaData.Size)
			modeStr := formatColumn(metaData.Mode, modeWidth)
			if metaData.Suid {
				modeStr = fmt.Sprintf("\x1b[31m%s\x1b[0m", modeStr)
			}
			ownerStr := formatColumn(metaData.Owner, ownerWidth)
			groupStr := formatColumn(metaData.Group, groupWidth)
			timeStr := formatColumn(metaData.ModTime, timeWidth)
			mimeTypeStr := formatColumn(metaData.MimeType, mimeTypeWidth)
			fileStr := formatColumn(filename, pathWidth)

			if verbose {
				fmt.Printf("%s %s %s %s %s %s %s\n", modeStr, ownerStr, groupStr, sizeStr, timeStr, mimeTypeStr, fileStr)
			} else {
				fmt.Printf("%s\n", fileStr)
			}
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
		fmt.Println("\n- files:", fileCount)
		fmt.Println("- bytes:", byteCount)

		if !(stringPatternRegex == nil && hexPatternRegex == nil && metaPatternRegex == nil) {
			fmt.Println("- matches:", matchCount)
		}
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

// Truncate or pad string to fit within width with ellipsis
func formatColumn(s string, width int) string {
	if len(s) <= width {
		return s + strings.Repeat(" ", width-len(s))
	} else {
		return s[:width-3] + "..."
	}
}

// extract file details
func extractFileData(file *os.File) (Metadata, bool, error) {
	var metadata Metadata
	isBinary := false

	// Get file size, mode, owner, and group
	fileInfo, err := file.Stat()
	if err == nil {
		metadata.Size = fileInfo.Size()
		metadata.Mode = fileInfo.Mode().String()
		metadata.Suid = (fileInfo.Mode()&os.ModeSetuid) != 0 && (fileInfo.Mode()&os.ModePerm) >= 04000

		// Get owner and group ids
		uid := fileInfo.Sys().(*syscall.Stat_t).Uid
		gid := fileInfo.Sys().(*syscall.Stat_t).Gid

		// Get owner and group names
		u, err := user.LookupId(fmt.Sprintf("%d", uid))
		if err == nil {
			metadata.Owner = fmt.Sprintf("%d - %s", uid, u.Username)
		} else {
			metadata.Owner = fmt.Sprintf("%d", uid)
		}

		g, err := user.LookupGroupId(fmt.Sprintf("%d", gid))
		if err == nil {
			metadata.Group = fmt.Sprintf("%d - %s", gid, g.Name)
		} else {
			metadata.Group = fmt.Sprintf("%d", gid)
		}

		// Get file mod time
		modTime := fileInfo.ModTime().Format("2006-01-02 15:04:05")
		metadata.ModTime = modTime
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
	metadata.MimeType = http.DetectContentType(buf)

	// Check if MIME type belongs to a group of known binary file types
	if strings.HasPrefix(metadata.MimeType, "application/octet-stream") ||
		strings.HasPrefix(metadata.MimeType, "application/pdf") ||
		strings.HasPrefix(metadata.MimeType, "image/") {
		isBinary = true
	}

	// If the file is not an image type return without exifdata
	if !strings.HasPrefix(metadata.MimeType, "image/") {
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
	metadata.ExifData = string(jsonByte)

	return metadata, isBinary, nil
}

func parseFlags() (bool, bool, bool, bool, bool, string, int, *regexp.Regexp, *regexp.Regexp, *regexp.Regexp, *regexp.Regexp, ignore.IgnoreParser) {
	var filePattern, stringPattern, hexPattern, metaPattern string
	var verbose, binary, errors, gitPattern, debugging, links bool
	var root string
	var depth int
	var ignoreParser ignore.IgnoreParser

	pflag.StringVarP(&filePattern, "file", "f", "", "regex pattern to match file names")
	pflag.StringVarP(&stringPattern, "string", "s", "", "regex pattern to match file string")
	pflag.StringVarP(&hexPattern, "hex", "h", "", "regex pattern to match hex-encoded lines")
	pflag.StringVarP(&metaPattern, "meta", "m", "", "regex pattern to match file metadata lines")

	pflag.BoolVarP(&verbose, "verbose", "v", false, "enable verbose mode")
	pflag.BoolVarP(&binary, "binary", "b", true, "include binary files in search")
	pflag.BoolVarP(&errors, "errors", "e", false, "print errors encountered during execution")
	pflag.BoolVarP(&debugging, "debugging", "t", false, "set debugging and trace during execution")
	pflag.BoolVarP(&links, "links", "l", true, "follow symbolic links to directories")
	pflag.BoolVarP(&gitPattern, "gitignore", "g", false, "search according to .gitignore")

	pflag.IntVarP(&depth, "depth", "d", -1, "depth to recurse, -1 for infinite depth")

	pflag.Parse()

	rootArgs := pflag.Args()
	if len(rootArgs) > 0 {
		homedir, _ := os.UserHomeDir()
		root = strings.Replace(rootArgs[0], "~", homedir, 1)
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
		if _, err := os.Stat(ignoreFilePath); os.IsNotExist(err) {
			if errors {
				fmt.Printf(".gitignore file not found in %s, ignoring gitPattern flag\n", root)
			}
		} else {
			ignoreParser, err = ignore.CompileIgnoreFile(ignoreFilePath)
			if err != nil {
				fmt.Printf("Error parsing .gitignore file: %v\n", err)
				os.Exit(1)
			}
		}
	}

	return verbose, binary, errors, debugging, links, root, depth, filePatternRegex, stringPatternRegex, hexPatternRegex, metaPatternRegex, ignoreParser
}
