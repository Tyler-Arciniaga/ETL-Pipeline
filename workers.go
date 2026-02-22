package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type WorkerPool struct {
	JobChan    chan ([]byte)
	NumWorkers uint64
	wg         sync.WaitGroup
}

type EmbeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

func NewWorkerPool(numWorkers uint64) *WorkerPool {
	return &WorkerPool{JobChan: make(chan []byte, 100), NumWorkers: numWorkers, wg: sync.WaitGroup{}}
}

func (w *WorkerPool) StartPool() {
	for i := range w.NumWorkers {
		go w.Work(i)
	}
}

func (w *WorkerPool) SubmitJob(job []byte) {
	w.wg.Add(1)
	w.JobChan <- job
}

func (w *WorkerPool) WaitToFinish() {
	w.StopPool()
	w.wg.Wait()
	fmt.Println("All jobs finished!")
}

func (w *WorkerPool) StopPool() {
	close(w.JobChan)
}

func (w *WorkerPool) Work(id uint64) {
	fmt.Printf("Worker %d running!\n", id)
	client := &http.Client{Timeout: 30 * time.Second}
	for job := range w.JobChan {
		w.CreateVectorEmbedding(job, client)
		w.wg.Done()
	}
}

func (w *WorkerPool) CreateVectorEmbedding(buf []byte, client *http.Client) {
	url := "http://localhost:11434/api/embed"
	embeddingReq := EmbeddingRequest{Model: "qwen3-embedding:4b", Input: string(buf)}
	postBody, err := json.Marshal(embeddingReq)
	if err != nil {
		slog.Error("marshalling embedding request", "err", err)
		return
	}

	requestBody := bytes.NewBuffer(postBody)

	resp, err := client.Post(url, "application/json", requestBody)
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
}
