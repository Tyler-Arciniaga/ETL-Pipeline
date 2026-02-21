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
	if !d.IsDir() {
		ext := filepath.Ext(filepath.Base(path))
		if ext != ".txt" && ext != ".md" {
			return nil
		} //skip files that are not txt or md
	}

	fmt.Println("Filepath:", path)
	fmt.Println("File:", filepath.Base(path))
	fmt.Println("Is Directory:", d.IsDir())
	return nil
}
