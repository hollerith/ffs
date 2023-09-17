package main

import (
	"io/ioutil"
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

func TestSearchFileFlag(t *testing.T) {
	setup()
	testDir := "./fixtures"
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
    testDir := "./fixtures"
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
	testDir := "./fixtures"
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
	testDir := "./fixtures"
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

func TestEmptySearchString(t *testing.T) {
	setup()
	testDir := "./fixtures"
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
    testDir := "./fixtures"
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
    testDir := "./fixtures"
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
