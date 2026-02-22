package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
)

type WorkerPool struct {
	JobChan    chan ([]byte)
	DoneChans  []chan (byte)
	NumWorkers uint64
	wg         sync.WaitGroup
}

type EmbeddingRequest struct {
	Model      string `json:"model"`
	Input      []byte `json:"input"`
	Dimensions uint64 `json:"dimensions"`
}

func NewWorkerPool(numWorkers uint64) *WorkerPool {
	return &WorkerPool{JobChan: make(chan ([]byte)), DoneChans: make([]chan (byte), 0), NumWorkers: numWorkers, wg: sync.WaitGroup{}}
}

func (w *WorkerPool) StartPool() {
	for i := range w.NumWorkers {
		doneChan := make(chan (byte))
		w.DoneChans = append(w.DoneChans, doneChan)
		go w.Work(i, doneChan)
	}
}

func (w *WorkerPool) SubmitJob(job []byte) {
	w.wg.Add(1)
	w.JobChan <- job
}

func (w *WorkerPool) WaitToFinish() {
	w.wg.Wait()
	w.StopPool()
	fmt.Println("All jobs finished!")
}

func (w *WorkerPool) StopPool() {
	for _, done := range w.DoneChans {
		go func() {
			done <- 1
		}()
	}
}

func (w *WorkerPool) Work(id uint64, doneChan chan (byte)) {
	fmt.Printf("Worker %d running!\n", id)
	select {
	case job := <-w.JobChan:
		fmt.Printf("Worker %d started a job\n", id)
		w.CreateVectorEmbedding(job)
	case <-doneChan:
		return
	}
}

func (w *WorkerPool) CreateVectorEmbedding(buf []byte) {
	url := "http://localhost:11434/api/embed"
	embeddingReq := EmbeddingRequest{Model: "qwen3-embedding:4b", Input: buf, Dimensions: 1000}
	postBody, err := json.Marshal(embeddingReq)
	if err != nil {
		slog.Error("marshalling embedding request", "err", err)
		return
	}

	requestBody := bytes.NewBuffer(postBody)

	resp, err := http.Post(url, "application/json", requestBody)
	if err != nil {
		slog.Error("sending post request to Ollama embedding api", "err", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("error reading response body", "err", err)
		return
	}

	fmt.Println("Got response:", string(body))
	w.wg.Done()
}
