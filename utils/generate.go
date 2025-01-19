package main

// generate checksums for all files in the static folder
// and write them to a file in their designated folder

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func GenerateSums() error {
	// open the file for writing
	f, err := os.Create("static/sums.txt")
	if err != nil {
		return fmt.Errorf("could not create sums file: %w", err)
	}
	defer f.Close()

	// walk the static folder
	err = filepath.Walk("static", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("could not walk path %s: %w", path, err)
		}

		// skip directories
		if info.IsDir() {
			return nil
		}

		// skip the sums.txt file
		if path == "static/sums.txt" {
			return nil
		}

		// open the file
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("could not open file %s: %w", path, err)
		}
		defer file.Close()

		// create a new hash
		hash := sha256.New()

		// copy the file to the hash
		_, err = io.Copy(hash, file)
		if err != nil {
			return fmt.Errorf("could not copy file to hash: %w", err)
		}

		// get the sum
		sum := hex.EncodeToString(hash.Sum(nil))

		// write the sum to the file
		relPath, err := filepath.Rel("static", path)
		if err != nil {
			return fmt.Errorf("could not get relative path: %w", err)
		}
		_, err = fmt.Fprintf(f, "%s %s\n", sum, relPath)
		if err != nil {
			return fmt.Errorf("could not write sum to file: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("could not walk static folder: %w", err)
	}

	return nil
}

func main() {
	GenerateSums()
}
