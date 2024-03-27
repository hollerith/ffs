package main

import (
    "net/http"
	"fmt"
	"strings"
	"strconv"
	"github.com/rwcarlsen/goexif/exif"
	"os"
	"os/user"
	"io"
	"syscall"
	"path/filepath"
)

func replaceNonPrintable(s string) string {
	b := []byte(s)
	for i, c := range b {
		if !strconv.IsPrint(rune(c)) {
			b[i] = '.'
		}
	}
	return string(b)
}

// Utility function to convert bytes to a human-readable format
func humanizeBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

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
	if !strings.HasPrefix(metadata.MimeType, "text/") {
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

// Utility function to format a column with a fixed width
func formatColumn(s string, width int) string {
	if len(s) <= width {
		return s + strings.Repeat(" ", width-len(s))
	} else {
		return s[:width-3] + "..."
	}
}

func walk(filename string, linkDirname string, followLinks bool, visited map[string]bool, walkFn filepath.WalkFunc) error {
    symWalkFunc := func(path string, info os.FileInfo, err error) error {
        if fname, err := filepath.Rel(filename, path); err == nil {
            path = filepath.Join(linkDirname, fname)
        } else {
            return err
        }

        if err == nil && info.Mode()&os.ModeSymlink == os.ModeSymlink && followLinks {
            finalPath, err := filepath.EvalSymlinks(path)
            if err != nil {
                return err
            }

            finalPath = filepath.Clean(finalPath) // clean up the final path

            if visited[finalPath] {
                // already visited this directory, skip it
                return nil
            }

            visited[finalPath] = true

            finalInfo, err := os.Lstat(finalPath)
            if err != nil {
                return walkFn(path, info, err)
            }

            if finalInfo.IsDir() {
                return walk(finalPath, path, followLinks, visited, walkFn)
            }
        }

        return walkFn(path, info, err)
    }

    return filepath.Walk(filename, symWalkFunc)
}

func Walk(path string, followLinks bool, walkFn filepath.WalkFunc) error {
    visited := make(map[string]bool) // create visited map
    return walk(path, path, followLinks, visited, walkFn)
}
