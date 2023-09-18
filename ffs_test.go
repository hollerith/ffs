package main

import (
	"io/ioutil"
	"io"
	"strings"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
)

func setup() {
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	fileCount = 0
	byteCount = 0
	matchCount = 0

}

func TestWalkFunction_NestedDir(t *testing.T) {
	setup()

	tempDir := "./testwalk_nested"
	if err := os.Mkdir(tempDir, 0755); err != nil {
		t.Fatalf("Could not create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	nestedDir := filepath.Join(tempDir, "nested")
	if err := os.Mkdir(nestedDir, 0755); err != nil {
		t.Fatalf("Could not create nested directory: %v", err)
	}

	walkFn := func(path string, info os.FileInfo, err error) error {
		return nil
	}

	if err := Walk(tempDir, true, walkFn); err != nil {
		t.Fatalf("'Walk' function returned error: %v", err)
	}
}

func TestWalkFunction_DifferentFileTypes(t *testing.T) {
	setup()
	tempDir := "./testwalk_types"
	if err := os.Mkdir(tempDir, 0755); err != nil {
		t.Fatalf("Could not create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	textFilePath := filepath.Join(tempDir, "text.txt")
	if err := ioutil.WriteFile(textFilePath, []byte("Hello, world!"), 0644); err != nil {
		t.Fatalf("Could not create text file: %v", err)
	}

	walkFn := func(path string, info os.FileInfo, err error) error {
		return nil
	}

	if err := walk(tempDir, "", true, make(map[string]bool), walkFn); err != nil {
		t.Fatalf("'walk' function returned error: %v", err)
	}
}

func TestSearchSimple(t *testing.T) {
    setup()

    testDir := "./tests"

    os.Args = []string{"ffs", testDir}
    main()

    expectedFileCount := 6  // Six files should be found

    if fileCount != expectedFileCount {
        t.Errorf("Expected fileCount: %d, Got: %d", expectedFileCount, fileCount)
    }
}

func TestSearchFileFlag(t *testing.T) {
	setup()

	testDir := "./tests/fixtures"
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("Could not create temp directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	file1Path := filepath.Join(testDir, "file1.txt")
	if err := ioutil.WriteFile(file1Path, []byte("This is a sample text."), 0644); err != nil {
		t.Fatalf("Could not create file1: %v", err)
	}

	file2Path := filepath.Join(testDir, "file2.txt")
	if err := ioutil.WriteFile(file2Path, []byte("This is another sample."), 0644); err != nil {
		t.Fatalf("Could not create file2: %v", err)
	}

	// Run your FFS application with specific arguments to search for "sample"
	os.Args = []string{"ffs", testDir, "--file", "*.txt", "--verbose", "--global"}
	main()

	expectedFileCount := 2
	expectedByteCount := int64(45)
	expectedMatchCount := 0

	if fileCount != expectedFileCount {
		t.Errorf("Expected fileCount: %d, Got: %d", expectedFileCount, fileCount)
	}

	if byteCount != expectedByteCount {
		t.Errorf("Expected byteCount: %d, Got: %d", expectedByteCount, byteCount)
	}

	if matchCount != expectedMatchCount {
		t.Errorf("Expected matchCount: %d, Got: %d", expectedMatchCount, matchCount)
	}
}

func TestSearchFileFlagWithRegex(t *testing.T) {
    setup()

    testDir := "./tests/fixtures"
    if err := os.Mkdir(testDir, 0755); err != nil {
        t.Fatalf("Could not create temp directory: %v", err)
    }
    defer os.RemoveAll(testDir)

    file1Path := filepath.Join(testDir, "file1.txt")
    if err := ioutil.WriteFile(file1Path, []byte("This is a sample text."), 0644); err != nil {
        t.Fatalf("Could not create file1: %v", err)
    }

    file2Path := filepath.Join(testDir, "file2.log")
    if err := ioutil.WriteFile(file2Path, []byte("This is another sample."), 0644); err != nil {
        t.Fatalf("Could not create file2: %v", err)
    }

    // Run your FFS application with specific arguments to search using a regex pattern
    os.Args = []string{"ffs", testDir, "--file", ".*\\.txt", "--verbose", "--global"}
    main()

    expectedFileCount := 1   		// Only "file1.txt" should match the regex pattern
    expectedByteCount := int64(22)  // Total bytes from the matched file
    expectedMatchCount := 0  		// No string match

    if fileCount != expectedFileCount {
        t.Errorf("Expected fileCount: %d, Got: %d", expectedFileCount, fileCount)
    }

    if byteCount != expectedByteCount {
        t.Errorf("Expected byteCount: %d, Got: %d", expectedByteCount, byteCount)
    }

    if matchCount != expectedMatchCount {
        t.Errorf("Expected matchCount: %d, Got: %d", expectedMatchCount, matchCount)
    }
}

func TestSearchTextFlag(t *testing.T) {
	setup()

	testDir := "./tests/fixtures"
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("Could not create temp directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	file1Path := filepath.Join(testDir, "file1.txt")
	if err := ioutil.WriteFile(file1Path, []byte("This is a sample text."), 0644); err != nil {
		t.Fatalf("Could not create file1: %v", err)
	}

	file2Path := filepath.Join(testDir, "file2.txt")
	if err := ioutil.WriteFile(file2Path, []byte("This is another sample."), 0644); err != nil {
		t.Fatalf("Could not create file2: %v", err)
	}

	os.Args = []string{"ffs", testDir, "--string", "sample", "--verbose", "--global"}
	main()

	expectedFileCount := 2
	expectedByteCount := int64(45)
	expectedMatchCount := 2

	if fileCount != expectedFileCount {
		t.Errorf("Expected fileCount: %d, Got: %d", expectedFileCount, fileCount)
	}

	if byteCount != expectedByteCount {
		t.Errorf("Expected byteCount: %d, Got: %d", expectedByteCount, byteCount)
	}

	if matchCount != expectedMatchCount {
		t.Errorf("Expected matchCount: %d, Got: %d", expectedMatchCount, matchCount)
	}
}

func TestSearchTextFlag_Negative(t *testing.T) {
	setup()

	testDir := "./tests/fixtures"
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("Could not create temp directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	file1Path := filepath.Join(testDir, "file1.txt")
	if err := ioutil.WriteFile(file1Path, []byte("This is a sample text."), 0644); err != nil {
		t.Fatalf("Could not create file1: %v", err)
	}

	file2Path := filepath.Join(testDir, "file2.txt")
	if err := ioutil.WriteFile(file2Path, []byte("This is another sample."), 0644); err != nil {
		t.Fatalf("Could not create file2: %v", err)
	}

	os.Args = []string{"ffs", testDir, "--string", "notfound", "--verbose", "--global"}
	main()

	expectedFileCount := 0
	expectedByteCount := int64(0)
	expectedMatchCount := 0 // No matches should be found in this negative case

	if fileCount != expectedFileCount {
		t.Errorf("Expected fileCount: %d, Got: %d", expectedFileCount, fileCount)
	}

	if byteCount != expectedByteCount {
		t.Errorf("Expected byteCount: %d, Got: %d", expectedByteCount, byteCount)
	}

	if matchCount != expectedMatchCount {
		t.Errorf("Expected matchCount: %d, Got: %d", expectedMatchCount, matchCount)
	}
}

func TestNoFilesFound(t *testing.T) {
	setup()

	testDir := "./empty_directory" // Create an empty directory
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("Could not create temp directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	os.Args = []string{"ffs", testDir, "--string", "sample", "--verbose", "--global"}
	main()

	expectedFileCount := 0     // No files in the directory
	expectedByteCount := int64(0) // No files to count bytes from
	expectedMatchCount := 0    // No files, so no matches should be found

	if fileCount != expectedFileCount {
		t.Errorf("Expected fileCount: %d, Got: %d", expectedFileCount, fileCount)
	}

	if byteCount != expectedByteCount {
		t.Errorf("Expected byteCount: %d, Got: %d", expectedByteCount, byteCount)
	}

	if matchCount != expectedMatchCount {
		t.Errorf("Expected matchCount: %d, Got: %d", expectedMatchCount, matchCount)
	}
}

func TestSearchOnlyPath(t *testing.T) {
    setup()

    testDir := "./tests/fixtures"
    if err := os.MkdirAll(testDir, 0755); err != nil {
        t.Fatalf("Could not create temp directory: %v", err)
    }
    defer os.RemoveAll(testDir)

    // Create a main.js file that includes 'react'
    mainFilePath := filepath.Join(testDir, "main.js")
    if err := ioutil.WriteFile(mainFilePath, []byte("#find react"), 0644); err != nil {
        t.Fatalf("Could not create main.js: %v", err)
    }

    // Create a node_modules directory and a react.js file inside it
    nodeModulesDir := filepath.Join(testDir, "node_modules")
    if err := os.Mkdir(nodeModulesDir, 0755); err != nil {
        t.Fatalf("Could not create node_modules directory: %v", err)
    }

    reactFilePath := filepath.Join(nodeModulesDir, "react.js")
    if err := ioutil.WriteFile(reactFilePath, []byte("#find react"), 0644); err != nil {
        t.Fatalf("Could not create react.js: %v", err)
    }

    os.Args = []string{"ffs", testDir, "--file", ".*node_modules.*", "--string", "react", "--verbose", "--global"}
    main()

    expectedFileCount := 1          // Should only include main.js and exclude node_modules
    expectedByteCount := int64(11)  // "import react from 'react';" is 24 bytes
    expectedMatchCount := 1         // One match in main.js

    if fileCount != expectedFileCount {
        t.Errorf("Expected fileCount: %d, Got: %d", expectedFileCount, fileCount)
    }

    if byteCount != expectedByteCount {
        t.Errorf("Expected byteCount: %d, Got: %d", expectedByteCount, byteCount)
    }

    if matchCount != expectedMatchCount {
        t.Errorf("Expected matchCount: %d, Got: %d", expectedMatchCount, matchCount)
    }
}

func TestEmptySearchString(t *testing.T) {
	setup()

	testDir := "./tests/fixtures"
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("Could not create temp directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	file1Path := filepath.Join(testDir, "file1.txt")
	if err := ioutil.WriteFile(file1Path, []byte("This is a sample text."), 0644); err != nil {
		t.Fatalf("Could not create file1: %v", err)
	}

	file2Path := filepath.Join(testDir, "file2.txt")
	if err := ioutil.WriteFile(file2Path, []byte("This is another sample."), 0644); err != nil {
		t.Fatalf("Could not create file2: %v", err)
	}

	os.Args = []string{"ffs", testDir, "--string", "", "--verbose", "--global"}
	main()

	expectedFileCount := 2     // Two files in the directory
	expectedByteCount := int64(45) // Total bytes from both files
	expectedMatchCount := 0    // An empty search string should not match anything

	if fileCount != expectedFileCount {
		t.Errorf("Expected fileCount: %d, Got: %d", expectedFileCount, fileCount)
	}

	if byteCount != expectedByteCount {
		t.Errorf("Expected byteCount: %d, Got: %d", expectedByteCount, byteCount)
	}

	if matchCount != expectedMatchCount {
		t.Errorf("Expected matchCount: %d, Got: %d", expectedMatchCount, matchCount)
	}
}

func TestMetaFlag(t *testing.T) {
    setup()

    testDir := "./tests/fixtures"
    if err := os.Mkdir(testDir, 0755); err != nil {
        t.Fatalf("Could not create temp directory: %v", err)
    }
    defer os.RemoveAll(testDir)

    file1Path := filepath.Join(testDir, "file1.txt")
    if err := ioutil.WriteFile(file1Path, []byte("This is a sample text."), 0644); err != nil {
        t.Fatalf("Could not create file1: %v", err)
    }

    file2Path := filepath.Join(testDir, "file2.txt")
    if err := ioutil.WriteFile(file2Path, []byte("This is another sample."), 0644); err != nil {
        t.Fatalf("Could not create file2: %v", err)
    }

    os.Args = []string{"ffs", testDir, "--meta", "text/plain", "--verbose", "--global"}
    main()

    expectedFileCount := 2     // Two files in the directory
    expectedByteCount := int64(45) // Total bytes from both files
    expectedMatchCount := 2    // Two matches with MIME type "text/plain"

    if fileCount != expectedFileCount {
        t.Errorf("Expected fileCount: %d, Got: %d", expectedFileCount, fileCount)
    }

    if byteCount != expectedByteCount {
        t.Errorf("Expected byteCount: %d, Got: %d", expectedByteCount, byteCount)
    }

    if matchCount != expectedMatchCount {
        t.Errorf("Expected matchCount: %d, Got: %d", expectedMatchCount, matchCount)
    }
}

func TestSearchStringWithGitIgnore(t *testing.T) {
    setup()

    testDir := "./tests/fixtures"
    if err := os.Mkdir(testDir, 0755); err != nil {
        t.Fatalf("Could not create temp directory: %v", err)
    }
    defer os.RemoveAll(testDir)

    // Create a .gitignore file that excludes "file1.txt"
    gitignorePath := filepath.Join(testDir, ".gitignore")
    if err := ioutil.WriteFile(gitignorePath, []byte("file1.txt"), 0644); err != nil {
        t.Fatalf("Could not create .gitignore file: %v", err)
    }

    file1Path := filepath.Join(testDir, "file1.txt")
    if err := ioutil.WriteFile(file1Path, []byte("This is a sample text."), 0644); err != nil {
        t.Fatalf("Could not create file1: %v", err)
    }

    file2Path := filepath.Join(testDir, "file2.txt")
    if err := ioutil.WriteFile(file2Path, []byte("This is another sample."), 0644); err != nil {
        t.Fatalf("Could not create file2: %v", err)
    }

    os.Args = []string{"ffs", testDir, "--string", "sample", "--verbose"}
    main()

    expectedFileCount := 1     		// Only one file should be matched due to .gitignore
    expectedByteCount := int64(23) 	// Total bytes from the matched file
    expectedMatchCount := 1    		// One match

    if fileCount != expectedFileCount {
        t.Errorf("Expected fileCount: %d, Got: %d", expectedFileCount, fileCount)
    }

    if byteCount != expectedByteCount {
        t.Errorf("Expected byteCount: %d, Got: %d", expectedByteCount, byteCount)
    }

    if matchCount != expectedMatchCount {
        t.Errorf("Expected matchCount: %d, Got: %d", expectedMatchCount, matchCount)
    }
}


func TestSearchHex(t *testing.T) {
    setup()

    testDir := "./tests/fixtures"
    if err := os.Mkdir(testDir, 0755); err != nil {
        t.Fatalf("Could not create temp directory: %v", err)
    }
    defer os.RemoveAll(testDir)

    // Create a sample file with actual bytes that represent the string "hello" (hex: 68656c6c6f)
    hexFilePath := filepath.Join(testDir, "hex_file.txt")
    if err := ioutil.WriteFile(hexFilePath, []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f}, 0644); err != nil {
        t.Fatalf("Could not create hex_file: %v", err)
    }

    // Run ffs with --hex flag
    os.Args = []string{"ffs", testDir, "--hex", "68 65 6c 6c 6f", "--verbose", "--global"}
    main()

    expectedFileCount := 1     		// One file should be matched
    expectedByteCount := int64(5) 	// Total bytes from the matched file (hello in hex is 5 bytes)
    expectedMatchCount := 1    		// One match

    if fileCount != expectedFileCount {
        t.Errorf("Expected fileCount: %d, Got: %d", expectedFileCount, fileCount)
    }

    if byteCount != expectedByteCount {
        t.Errorf("Expected byteCount: %d, Got: %d", expectedByteCount, byteCount)
    }

    if matchCount != expectedMatchCount {
        t.Errorf("Expected matchCount: %d, Got: %d", expectedMatchCount, matchCount)
    }
}

func TestSearchMultipleWithGitIgnore(t *testing.T) {
    setup()

    testDir := "./tests/fixtures"
    if err := os.Mkdir(testDir, 0755); err != nil {
        t.Fatalf("Could not create temp directory: %v", err)
    }
    defer os.RemoveAll(testDir)

    // Create a .gitignore file that excludes "file1.txt"
    gitignorePath := filepath.Join(testDir, ".gitignore")
    if err := ioutil.WriteFile(gitignorePath, []byte("file1.txt"), 0644); err != nil {
        t.Fatalf("Could not create .gitignore file: %v", err)
    }

    file1Path := filepath.Join(testDir, "file1.txt")
    if err := ioutil.WriteFile(file1Path, []byte("This is a sample text."), 0644); err != nil {
        t.Fatalf("Could not create file1: %v", err)
    }

    file2Path := filepath.Join(testDir, "file2.txt")
    if err := ioutil.WriteFile(file2Path, []byte("This is another sample."), 0644); err != nil {
        t.Fatalf("Could not create file2: %v", err)
    }

    os.Args = []string{"ffs", testDir, "--file", "file.*", "--string", "sample", "--meta", "text", "--verbose"}
    main()

    expectedFileCount := 1     		// Only one file should be matched due to .gitignore
    expectedByteCount := int64(23) 	// Total bytes from the matched file
    expectedMatchCount := 2    		// Two matches, string match and metafield match

    if fileCount != expectedFileCount {
        t.Errorf("Expected fileCount: %d, Got: %d", expectedFileCount, fileCount)
    }

    if byteCount != expectedByteCount {
        t.Errorf("Expected byteCount: %d, Got: %d", expectedByteCount, byteCount)
    }

    if matchCount != expectedMatchCount {
        t.Errorf("Expected matchCount: %d, Got: %d", expectedMatchCount, matchCount)
    }
}

func TestSearchWithDepthFlag(t *testing.T) {
    setup()

    testDir := "./tests/fixtures"
    if err := os.Mkdir(testDir, 0755); err != nil {
        t.Fatalf("Could not create temp directory: %v", err)
    }
    defer os.RemoveAll(testDir)

    // Create a file at depth 3
    deepFilePath := filepath.Join(testDir, "dir1", "dir2", "dir3", "deep.txt")
    if err := os.MkdirAll(filepath.Dir(deepFilePath), 0755); err != nil {
        t.Fatalf("Could not create directories: %v", err)
    }
    if err := ioutil.WriteFile(deepFilePath, []byte("This is a deep file."), 0644); err != nil {
        t.Fatalf("Could not create deep file: %v", err)
    }

    // Create a file at depth 4
    tooDeepFilePath := filepath.Join(testDir, "dir1", "dir2", "dir3", "dir4", "toodeep.txt")
    if err := os.MkdirAll(filepath.Dir(tooDeepFilePath), 0755); err != nil {
        t.Fatalf("Could not create directories: %v", err)
    }
    if err := ioutil.WriteFile(tooDeepFilePath, []byte("This is a too deep file."), 0644); err != nil {
        t.Fatalf("Could not create too deep file: %v", err)
    }

    os.Args = []string{"ffs", testDir, "--string", "deep", "--verbose", "--depth", "3", "--global"}
    main()

    expectedFileCount := 1     		// Only one file at depth 3 should be matched
    expectedByteCount := int64(20) 	// Total bytes from the matched file
    expectedMatchCount := 1    		// One match

    if fileCount != expectedFileCount {
        t.Errorf("Expected fileCount: %d, Got: %d", expectedFileCount, fileCount)
    }

    if byteCount != expectedByteCount {
        t.Errorf("Expected byteCount: %d, Got: %d", expectedByteCount, byteCount)
    }

    if matchCount != expectedMatchCount {
        t.Errorf("Expected matchCount: %d, Got: %d", expectedMatchCount, matchCount)
    }
}

func TestSearchWithDepthLimit(t *testing.T) {
    setup()

    testDir := "./tests/fixtures"
    level1Dir := "./tests/fixtures/level1"
    level2Dir := "./tests/fixtures/level1/level2"
    level3Dir := "./tests/fixtures/level1/level2/level3"

    for _, dir := range []string{testDir, level1Dir, level2Dir, level3Dir} {
        if err := os.Mkdir(dir, 0755); err != nil {
            t.Fatalf("Could not create directory %s: %v", dir, err)
        }
    }

    defer os.RemoveAll(testDir)

    filePath := filepath.Join(level3Dir, "deepFile.txt")
    if err := ioutil.WriteFile(filePath, []byte("This is a deep test."), 0644); err != nil {
        t.Fatalf("Could not create deepFile: %v", err)
    }

    os.Args = []string{"ffs", testDir, "--string", "deep", "--depth", "2", "--verbose", "--global"}
    main()

    expectedFileCount := 0
    expectedByteCount := int64(0)
    expectedMatchCount := 0

    if fileCount != expectedFileCount {
        t.Errorf("Expected fileCount: %d, Got: %d", expectedFileCount, fileCount)
    }

    if byteCount != expectedByteCount {
        t.Errorf("Expected byteCount: %d, Got: %d", expectedByteCount, byteCount)
    }

    if matchCount != expectedMatchCount {
        t.Errorf("Expected matchCount: %d, Got: %d", expectedMatchCount, matchCount)
    }
}

func TestSearchWithLinks(t *testing.T) {
    setup()

    testDir := "./tests/fixtures"
    if err := os.Mkdir(testDir, 0755); err != nil {
        t.Fatalf("Could not create temp directory: %v", err)
    }
    defer os.RemoveAll(testDir)

    // Create a sample file with content
    filePath := filepath.Join(testDir, "sample_file.txt")
    if err := ioutil.WriteFile(filePath, []byte("This is a symlink test."), 0644); err != nil {
        t.Fatalf("Could not create sample_file: %v", err)
    }

    // Create a symlink to the sample file
    symlinkPath := filepath.Join(testDir, "sample_symlink.txt")
    if err := os.Symlink(filePath, symlinkPath); err != nil {
        t.Fatalf("Could not create symlink: %v", err)
    }

    // Run ffs with --links flag
    os.Args = []string{"ffs", testDir, "--string", "symlink", "--links", "--verbose", "--global"}
    main()

    expectedFileCount := 1     		// Both the file and its symlink should be matched
    expectedByteCount := int64(23) 	// Total bytes from the matched files
    expectedMatchCount := 1    		// Two matches

    if fileCount != expectedFileCount {
        t.Errorf("Expected fileCount: %d, Got: %d", expectedFileCount, fileCount)
    }

    if byteCount != expectedByteCount {
        t.Errorf("Expected byteCount: %d, Got: %d", expectedByteCount, byteCount)
    }

    if matchCount != expectedMatchCount {
        t.Errorf("Expected matchCount: %d, Got: %d", expectedMatchCount, matchCount)
    }
}

func TestSearchWithGitignoreAndLinks(t *testing.T) {
    setup()

    testDir := "./tests/fixtures"
    if err := os.Mkdir(testDir, 0755); err != nil {
        t.Fatalf("Could not create temp directory: %v", err)
    }
    defer os.RemoveAll(testDir)

    // Create a sub-directory and a symlink to that directory
    subDir := filepath.Join(testDir, "subDir")
    if err := os.Mkdir(subDir, 0755); err != nil {
        t.Fatalf("Could not create sub-directory: %v", err)
    }

    linkDir := filepath.Join(testDir, "linkDir")
    if err := os.Symlink("./subDir", linkDir); err != nil {
        t.Fatalf("Could not create symlink to directory: %v", err)
    }

    // Create a sample file in the sub-directory
    filePath := filepath.Join(subDir, "sample_file.txt")
    if err := ioutil.WriteFile(filePath, []byte("This is in a symlinked directory."), 0644); err != nil {
        t.Fatalf("Could not create sample_file: %v", err)
    }

    // Create a .gitignore file that excludes the sub-directory
    gitignorePath := filepath.Join(testDir, ".gitignore")
    if err := ioutil.WriteFile(gitignorePath, []byte("subDir"), 0644); err != nil {
        t.Fatalf("Could not create .gitignore file: %v", err)
    }

    // Run ffs with --links flag but without --global
    os.Args = []string{"ffs", testDir, "--string", "symlinked", "--links", "--verbose"}
    main()

    expectedFileCount := 1     		// One file should be matched due to links .gitignore
    expectedByteCount := int64(33) 	// 33 bytes should be matched
    expectedMatchCount := 1    		// Links bypasses the .gitignore and matches the file

    if fileCount != expectedFileCount {
        t.Errorf("Expected fileCount: %d, Got: %d", expectedFileCount, fileCount)
    }

    if byteCount != expectedByteCount {
        t.Errorf("Expected byteCount: %d, Got: %d", expectedByteCount, byteCount)
    }

    if matchCount != expectedMatchCount {
        t.Errorf("Expected matchCount: %d, Got: %d", expectedMatchCount, matchCount)
    }
}

func TestSearchWithErrorsFlag(t *testing.T) {
    setup()

    testDir := "./tests/fixtures"
    if err := os.Mkdir(testDir, 0755); err != nil {
        t.Fatalf("Could not create temp directory: %v", err)
    }
    defer os.RemoveAll(testDir)

    // Create a readable file
    readableFilePath := filepath.Join(testDir, "readable.txt")
    if err := ioutil.WriteFile(readableFilePath, []byte("This is readable."), 0644); err != nil {
        t.Fatalf("Could not create readableFile: %v", err)
    }

    // Create an unreadable file
    unreadableFilePath := filepath.Join(testDir, "unreadable.txt")
    if err := ioutil.WriteFile(unreadableFilePath, []byte("This is unreadable."), 0000); err != nil {
        t.Fatalf("Could not create unreadableFile: %v", err)
    }

    // Redirect stdout and stderr to capture messages
    oldStdout := os.Stdout
    oldStderr := os.Stderr
    r, w, _ := os.Pipe()
    os.Stdout = w
    os.Stderr = w

    os.Args = []string{"ffs", testDir, "--string", "readable", "--errors", "--verbose", "--global"}
    main()

    // Revert the stdout and stderr redirection
    w.Close()
    os.Stdout = oldStdout
    os.Stderr = oldStderr

    var buf bytes.Buffer
    io.Copy(&buf, r)
    capturedOutput := buf.String()

    // Validate the output to make sure an error message was printed
    if !strings.Contains(capturedOutput, "unreadable.txt") {
        t.Errorf("Expected an error for unreadable.txt, but did not find one in output")
    }
}

func TestSearchWithBinaryFlag(t *testing.T) {
    setup()

    testDir := "./tests/fixtures"
    if err := os.MkdirAll(testDir, 0755); err != nil {
        t.Fatalf("Could not create temp directory: %v", err)
    }
    defer os.RemoveAll(testDir)

    // Create a text file
    textFilePath := filepath.Join(testDir, "textFile.txt")
    if err := ioutil.WriteFile(textFilePath, []byte("Search me"), 0644); err != nil {
        t.Fatalf("Could not create textFile: %v", err)
    }

    // Create a binary file that also contains the string "Search"
    binaryFilePath := filepath.Join(testDir, "binaryFile.bin")
    if err := ioutil.WriteFile(binaryFilePath, []byte{0x00, 0x01, 0x02, 0x03, 'S', 'e', 'a', 'r', 'c', 'h'}, 0644); err != nil {
        t.Fatalf("Could not create binaryFile: %v", err)
    }

    os.Args = []string{"ffs", testDir, "--string", "Search", "--binary", "--verbose", "--global"}
    main()

    expectedFileCount := 2  // Should include both text and binary files
    expectedByteCount := int64(9 + 10)  // "Search me" is 9 bytes, binary file is 10 bytes
    expectedMatchCount := 2  // One match in text file, one in binary file

    if fileCount != expectedFileCount {
        t.Errorf("Expected fileCount: %d, Got: %d", expectedFileCount, fileCount)
    }

    if byteCount != expectedByteCount {
        t.Errorf("Expected byteCount: %d, Got: %d", expectedByteCount, byteCount)
    }

    if matchCount != expectedMatchCount {
        t.Errorf("Expected matchCount: %d, Got: %d", expectedMatchCount, matchCount)
    }
}
