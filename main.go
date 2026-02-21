package main

import (
	"fmt"
	"io/fs"
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
	return nil
}
