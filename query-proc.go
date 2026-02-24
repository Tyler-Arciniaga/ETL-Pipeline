package main

import (
	"fmt"
	"strings"
)

type QueryProcessor struct {
	WorkerPool *WorkerPool
}

func (q QueryProcessor) StartQuery(input string) {
	results, err := q.WorkerPool.ProcessQuery(input)
	if err != nil {
		fmt.Println("Error fetching results for query:", err)
		return
	}

	q.PrintResults(results)
}

func (q QueryProcessor) PrintResults(results []string) {
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
		fmt.Println(q.IndentText(result, 4))
		fmt.Printf("\n%s\n\n", separator)
	}
}

// indentText indents multi-line text by given spaces
func (q QueryProcessor) IndentText(text string, spaces int) string {
	padding := strings.Repeat(" ", spaces)
	lines := strings.Split(text, "\n")

	for i := range lines {
		lines[i] = padding + lines[i]
	}

	return strings.Join(lines, "\n")
}
