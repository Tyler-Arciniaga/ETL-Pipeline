package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//TODO try chunking together HTTP requests to Ollama to embded multiple strings

type ETL struct {
	WorkerPool *WorkerPool
}

func main() {
	flags := HandleCommandArgs()
	workerPool := NewWorkerPool(flags.Workers)
	etl := ETL{WorkerPool: workerPool}

	start := time.Now()
	if flags.Query {
		etl.StartQuery(flags.Input)
	} else {
		etl.StartIngest(flags.Dir)
	}

	elapsed := time.Since(start)
	fmt.Println("Total work took:", elapsed.Seconds(), "seconds")
}

func HandleCommandArgs() CommandFlags {
	dir := flag.String("dir", "test-data", "root folder to ingest")
	workers := flag.Uint64("workers", 1, "number of workers to embed data chunks")
	query := flag.Bool("query", false, "toggle query mode")
	input := flag.String("input", "", "input for query")
	flag.Parse()

	return CommandFlags{Dir: *dir, Workers: *workers, Query: *query, Input: *input}
}

func (e ETL) StartQuery(input string) {
	results, err := e.WorkerPool.ProcessQuery(input)
	if err != nil {
		fmt.Println("Error fetching results for query:", err)
		return
	}

	PrintResults(results)
}

func PrintResults(results []string) {
	const separator = "────────────────────────────────────────────"

	if len(results) == 0 {
		fmt.Println("\nNo results found.")
		return
	}

	fmt.Printf("\n%s\n", separator)
	fmt.Printf("Found %d result(s)\n", len(results))
	fmt.Printf("%s\n\n", separator)

	for i, result := range results {
		fmt.Printf("[%d]\n", i+1)
		fmt.Println(indentText(result, 4))
		fmt.Printf("\n%s\n\n", separator)
	}
}

// indentText indents multi-line text by given spaces
func indentText(text string, spaces int) string {
	padding := strings.Repeat(" ", spaces)
	lines := strings.Split(text, "\n")

	for i := range lines {
		lines[i] = padding + lines[i]
	}

	return strings.Join(lines, "\n")
}

func (e ETL) StartIngest(rootDir string) {
	e.WorkerPool.StartPool()
	filepath.WalkDir(rootDir, e.ingestDirectory)
	e.WorkerPool.WaitToFinish()
	fmt.Println("Finished embedding all text content!")
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

		e.WorkerPool.SubmitJob(buf[:n], filepath, chunkNum)
		chunkNum++
	}

	return nil
}
