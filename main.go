package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func main() {
	dir := "test-data" // TODO change from hardcoded file directory to specified by user args
	filepath.WalkDir(dir, ingestDirectory)
}

func ingestDirectory(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	fmt.Println("Filepath:", path)
	fmt.Println("Is Directory:", d.IsDir())

	if !d.IsDir() {
		//DirEntry is a file...
		ext := filepath.Ext(filepath.Base(path))
		if ext != ".txt" && ext != ".md" {
			return nil
		} //skip files that are not txt or md

		err = readFile(path)
		if err != nil {
			return err
		}
	}

	return nil
}

func readFile(filepath string) error {
	fmt.Println("File:", filepath)

	file, err := os.Open(filepath)
	if err != nil {
		return err
	}

	var checksum [32]byte
	buf := make([]byte, 4096)
	for {
		n, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		checksum = sha256.Sum256(buf[:n])
		fmt.Println("Checksum:", checksum)
	}

	return nil
}
