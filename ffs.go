package main

import _ "net/http/pprof"

import (
	"bufio"
	"fmt"

	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/sabhiram/go-gitignore"
	"github.com/spf13/pflag"
)

type Metadata struct {
	Size     int64
	Mode     string
	Suid     bool
	Link     string
	Owner    string
	Group    string
	ModTime  string
	MimeType string
	ExifData string
	Error    string
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

var lastDir string
var fileCount int
var matchCount int
var byteCount int64

func main() {
	verbose, binary, errors, links, root, depth, filePatternRegex, stringPatternRegex, hexPatternRegex, metaPatternRegex, globalPattern, ignoreParser, tree := parseFlags()

	search := func(path string, info os.FileInfo, err error) error {
		var lastCount = matchCount

		if err != nil {
			if errors {
				fmt.Printf("Error processing file %s: %v\n", path, err)
			}
			return nil
		}

		// Check if file is a directory and if depth is reached
		if info.IsDir() {
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

		// By default only search files according to .gitignore
		if !globalPattern && (ignoreParser != nil && ignoreParser.MatchesPath(path)) {
			return nil
		}

		directory, filename := filepath.Split(path)
		directory = strings.TrimSuffix(directory, string(os.PathSeparator))
		if directory == "" {
			directory = root
		}

		// Ignore .git folders by default
		if !globalPattern && strings.Contains(path, ".git") {
			return nil
		}

		// Match filename regex pattern, optional
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

		if tree {
			depth := strings.Count(directory, string(os.PathSeparator))
			indent := strings.Repeat(" ", depth)
			fmt.Println(indent + filepath.Base(directory) + "/")
		}

		// Extract metadata and other file information
		metaData, isBinary, err := extractFileData(file)
		if err != nil {
			metaData.Error = fmt.Sprintf("Warn: %v", err)
		}

		// Check for metadata pattern match
		if metaPatternRegex != nil {
			metadataString := fmt.Sprintf("%d %s %s %s %s %s %s", metaData.Size, metaData.Mode, metaData.Owner, metaData.Group, metaData.ModTime, metaData.MimeType, metaData.ExifData)
			if metaPatternRegex.MatchString(metadataString) {
				matchCount++
			}
		}

		fi, err := os.Lstat(path)
		if err != nil {
			if errors {
				log.Printf("Error Lstat-ing %s: %v\n", path, err)
				return nil
			}
		}

		// Add link pointer to metaData
		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			linkPath, err := filepath.EvalSymlinks(path)
			if err != nil {
				if errors {
					log.Printf("Error eval-ing %s: %v\n", path, err)
					return nil
				}
			}
			metaData.Link = linkPath
		}

		// Check if file is binary and skip if set to exclude binary files
		if !binary && isBinary {
			return nil
		}

		// Scan each line of the file content
		if stringPatternRegex != nil || hexPatternRegex != nil {
			file.Seek(0, 0) // reset file pointer to the beginning of the file
			scanner := bufio.NewScanner(file)
			scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // set buffer size to 1MB
			lineNumber := 1
			hasPrintedFileDetails := false
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
					// Print results before printing the source line
					if !hasPrintedFileDetails {
						lastDir, fileCount, matchCount, byteCount = printResults(fileCount, lastDir, directory, filename, metaData, fi, byteCount, matchCount, verbose, tree, errors)
						hasPrintedFileDetails = true
					}
					if verbose {
						fmt.Printf("\x1b[38;5;221m%s\x1b[0m:\x1b[38;5;39m%d\x1b[0m:\x1b[38;5;8m%s\x1b[0m\n", path, lineNumber, replaceNonPrintable(line))
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
		} else {
			// Print results
			if (matchCount > lastCount) || (stringPatternRegex == nil && hexPatternRegex == nil && metaPatternRegex == nil) {
				lastDir, fileCount, matchCount, byteCount = printResults(fileCount, lastDir, directory, filename, metaData, fi, byteCount, matchCount, verbose, tree, errors)
			}
		}

		return nil
	}

	// Search
	err := Walk(root, links, search)

	if err != nil {
		if errors {
			fmt.Printf("Error walking directories: %v\n", err)
		}
	}

	if verbose {
		fmt.Println("\n\x1b[36m- files:\x1b[0m", fileCount)
		fmt.Printf("\x1b[36m- bytes:\x1b[0m %d (\x1b[33m%s\x1b[0m)\n", byteCount, humanizeBytes(byteCount))

		if !(stringPatternRegex == nil && hexPatternRegex == nil && metaPatternRegex == nil) {
			fmt.Println("\x1b[36m- matches:\x1b[0m", matchCount)
		}
		fmt.Printf("\n")
	}
}

func printResults(fileCount int, lastDir string, directory string, filename string, metaData Metadata, fi os.FileInfo, byteCount int64, matchCount int, verbose bool, tree bool, errors bool) (string, int, int, int64) {

	if verbose {
		// Print directory
		if fileCount == 0 || lastDir != directory {
			lastDir = directory
			// Check if the directory is a symlink
			dirInfo, err := os.Lstat(lastDir)
			if err == nil && dirInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
				// directory is a symlink, print final path in light yellow with arrow pointing to actual path in regular green
				finalPath, err := filepath.EvalSymlinks(lastDir)
				if err != nil {
					fmt.Printf("\n\033[38;5;221m%s\033[0m (could not resolve symlink):\n", lastDir)
				} else {
					fmt.Printf("\n\033[38;5;221m%s\033[0m --> \033[32m%s\033[0m:\n", lastDir, finalPath)
				}
			} else {
				fmt.Printf("\n\033[32m%s\033[0m:\n", lastDir)
			}
		}

		// Print the current file details
		sizeStr := fmt.Sprintf("%*d", sizeWidth, metaData.Size)
		modeStr := formatColumn(metaData.Mode, modeWidth)
		if metaData.Suid {
			modeStr = fmt.Sprintf("\x1b[31m%s\x1b[0m", modeStr)
		}
		ownerStr := formatColumn(metaData.Owner, ownerWidth)
		groupStr := formatColumn(metaData.Group, groupWidth)
		timeStr := formatColumn(metaData.ModTime, timeWidth)
		mimeTypeStr := formatColumn(metaData.MimeType, mimeTypeWidth)
		fileStr := filename
		if metaData.Link != "" {
			// file is a link, color it light yellow
			fileStr = fmt.Sprintf("\x1b[38;5;221m%s\x1b[0m --> %s", filename, metaData.Link)
		} else {
			if fi.Mode().Perm()&0111 != 0 {
				if fi.Mode().Perm()&0007 != 0 {
					// file is world executable, color it dark red
					fileStr = fmt.Sprintf("\x1b[38;5;124m%s\x1b[0m", filename)
				} else if fi.Mode().Perm()&0070 != 0 {
					// file is group executable, color it light red
					fileStr = fmt.Sprintf("\x1b[38;5;211m%s\x1b[0m", filename)
				} else {
					// file is owner executable, color it light pink
					fileStr = fmt.Sprintf("\x1b[38;5;219m%s\x1b[0m", filename)
				}
			} else {
				fileStr = fmt.Sprintf("\x1b[38;5;117m%s\x1b[0m", filename)
			}
			// Exclude symlinks from byteCount
			byteCount += metaData.Size
		}

		var errorStr string
		if errors {
			errorStr = fmt.Sprintf("\033[90m - %s\033[0m", metaData.Error)
		}

		fmt.Printf("%s %s %s %s %s %s %s %s\n", modeStr, ownerStr, groupStr, sizeStr, timeStr, mimeTypeStr, fileStr, errorStr)
	} else if tree {
        depth := strings.Count(directory, string(os.PathSeparator))
        indent := strings.Repeat(" ", depth)
        fmt.Println(indent + " " + filename)
	} else {
		// Default printing (neither verbose nor tree)
		fmt.Printf("%s/%s\n", directory, filename)
	}

	fileCount++

	return lastDir, fileCount, matchCount, byteCount
}

func parseFlags() (bool, bool, bool, bool, string, int, *regexp.Regexp, *regexp.Regexp, *regexp.Regexp, *regexp.Regexp, bool, ignore.IgnoreParser, bool) {
	var filePattern, stringPattern, hexPattern, metaPattern string
	var verbose, binary, errors, globalPattern, links, tree bool
	var root string
	var depth int
	var ignoreParser ignore.IgnoreParser
	var filePatternRegex, stringPatternRegex, hexPatternRegex, metaPatternRegex *regexp.Regexp
	var err error

	pflag.StringVarP(&filePattern, "file", "f", "", "regex pattern to match file names")
	pflag.StringVarP(&stringPattern, "string", "s", "", "regex pattern to match file string")
	pflag.StringVarP(&hexPattern, "hex", "x", "", "regex pattern to match hex-encoded lines")
	pflag.StringVarP(&metaPattern, "meta", "m", "", "regex pattern to match file metadata lines")
	pflag.BoolVarP(&verbose, "verbose", "v", false, "enable verbose mode")
	pflag.BoolVarP(&binary, "binary", "b", false, "exclude binary files in search")
	pflag.BoolVarP(&errors, "errors", "e", false, "print errors encountered during execution")
	pflag.BoolVarP(&links, "links", "l", false, "follow symbolic links to directories")
	pflag.BoolVarP(&globalPattern, "global", "g", false, "search all including .gitignore paths")
	pflag.BoolVarP(&tree, "tree", "t", false, "display results in a tree format")
	pflag.IntVarP(&depth, "depth", "d", -1, "depth to recurse, -1 for infinite depth")
	pflag.Parse()

	rootArgs := pflag.Args()
	if len(rootArgs) > 0 {
		homedir, _ := os.UserHomeDir()
		root = strings.Replace(rootArgs[0], "~", homedir, 1)
		if len(rootArgs) > 1 {
			root = "."
			filePattern = strings.Join(rootArgs, "|")
		} else {
			info, err := os.Stat(root)
			if os.IsNotExist(err) {
				fmt.Printf("Error: directory '%s' does not exist.\n", root)
				os.Exit(1)
			}
			if !info.IsDir() {
				filePattern = root
				root = "."
			}
		}
	} else {
		root = "."
	}

	if filePattern != "" {
		filePatternRegex, err = regexp.Compile(filePattern)
		if err != nil {
			// If the compilation fails, assume filePattern is a glob pattern and convert it
			filePattern = globToRegex(filePattern)
			filePatternRegex, err = regexp.Compile(filePattern)
			if err != nil {
				fmt.Printf("Error compiling file pattern regex: %v\n", err)
				os.Exit(1)
			}
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

	if !globalPattern {
		ignoreFilePath := filepath.Join(root, ".gitignore")
		if _, err := os.Stat(ignoreFilePath); os.IsNotExist(err) {
			if errors {
				fmt.Printf("No .gitignore file in %s\n", root)
			}
		} else {
			ignoreParser, err = ignore.CompileIgnoreFile(ignoreFilePath)
			if err != nil {
				fmt.Printf("Error parsing .gitignore file: %v\n", err)
				os.Exit(1)
			}
		}
	}

	return verbose, binary, errors, links, root, depth, filePatternRegex, stringPatternRegex, hexPatternRegex, metaPatternRegex, globalPattern, ignoreParser, tree
}
