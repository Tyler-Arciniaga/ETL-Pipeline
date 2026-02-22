package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type ETL struct {
	WorkerPool *WorkerPool
}

func main() {
	workerPool := NewWorkerPool(3)
	etl := ETL{WorkerPool: workerPool}
	etl.Start()
}

func (e ETL) Start() {
	e.WorkerPool.StartPool()
	dir := "test-data" // TODO change from hardcoded file directory to specified by user args
	filepath.WalkDir(dir, e.ingestDirectory)
}

func (e ETL) ingestDirectory(path string, d fs.DirEntry, err error) error {
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

		err = e.readFile(path)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e ETL) readFile(filepath string) error {
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

		e.WorkerPool.SubmitJob(buf[:n])
		checksum = sha256.Sum256(buf[:n])
		fmt.Println("Checksum:", checksum)
	}

	return nil
}
