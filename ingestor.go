package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type Ingestor interface {
	StartIngest(rootSource string)
	IngestSource(path string) error
	ReadSource(path string) error
}

type FileIngestor struct {
	WorkerPool *WorkerPool
}

func (f FileIngestor) StartIngest(rootDir string) {
	f.WorkerPool.StartPool()
	filepath.WalkDir(rootDir, f.ingestDirectory)
	f.WorkerPool.WaitToFinish()
}

func (f FileIngestor) IngestSource(path string) error { return nil }

func (f FileIngestor) ingestDirectory(path string, d fs.DirEntry, err error) error {
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

		err = f.ReadSource(path)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f FileIngestor) ReadSource(filepath string) error {
	fmt.Println("File:", filepath)

	file, err := os.Open(filepath)
	if err != nil {
		return err
	}

	buf := make([]byte, 4096)
	chunkNum := uint64(0)
	for {
		n, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		f.WorkerPool.SubmitJob(buf[:n], filepath, chunkNum)
		chunkNum++
	}

	return nil
}
