package main

import (
	"flag"
	"fmt"
	"time"
)

//TODO try chunking together HTTP requests to Ollama to embded multiple strings

type ETL struct {
	Ingestor       Ingestor
	QueryProcessor QueryProcessor
}

func main() {
	flags := HandleCommandArgs()
	workerPool := NewWorkerPool(flags.Workers)

	start := time.Now()

	var etl ETL
	if flags.Query {
		etl.QueryProcessor = QueryProcessor{WorkerPool: workerPool}
		etl.StartQuery(flags.Input)
	} else {
		if flags.Web {
			// etl.Ingestor = WebIngestor{}
		} else {
			etl.Ingestor = FileIngestor{WorkerPool: workerPool}
			etl.StartIngest(flags.Dir)
		}
	}

	elapsed := time.Since(start)
	fmt.Println("Total work took:", elapsed.Seconds(), "seconds")
}

func HandleCommandArgs() CommandFlags {
	dir := flag.String("dir", "test-data", "root folder to ingest")
	workers := flag.Uint64("workers", 1, "number of workers to embed data chunks")
	query := flag.Bool("query", false, "toggle query mode")
	input := flag.String("input", "", "input for query")
	web := flag.Bool("web", false, "toggle web ingest mode")
	flag.Parse()

	return CommandFlags{Dir: *dir, Workers: *workers, Query: *query, Input: *input, Web: *web}
}

func (e ETL) StartIngest(rootSource string) {
	e.Ingestor.StartIngest(rootSource)
	fmt.Println("Finished embedding all text content!")
}

func (e ETL) StartQuery(input string) {
	e.QueryProcessor.StartQuery(input)
	fmt.Println("Finished retrieving related results!")
}
