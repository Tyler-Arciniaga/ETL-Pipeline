package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type WorkerPool struct {
	DB_Client  *mongo.Client
	JobChan    chan (WorkerJob)
	NumWorkers uint64
	wg         sync.WaitGroup
}

type EmbeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

func NewWorkerPool(numWorkers uint64) *WorkerPool {
	return &WorkerPool{JobChan: make(chan WorkerJob, 100), NumWorkers: numWorkers, wg: sync.WaitGroup{}}
}

func (w *WorkerPool) StartPool() {
	if err := godotenv.Load(); err != nil {
		slog.Error("loading .env", "err", err)
	}

	err := w.ConnectToDB()
	if err != nil {
		slog.Error("Error connecting to DB", "err", err)
	}
	for i := range w.NumWorkers {
		go w.Work(i)
	}
}

func (w *WorkerPool) ConnectToDB() error {
	// Use the SetServerAPIOptions() method to set the version of the Stable API on the client
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)

	db_conn := os.Getenv("DB_CONN")
	opts := options.Client().ApplyURI(db_conn).SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server
	client, err := mongo.Connect(opts)
	if err != nil {
		return err
	}

	// Send a ping to confirm a successful connection
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		return err
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")

	w.DB_Client = client
	return nil
}

func (w *WorkerPool) SubmitJob(rawContent []byte, filepath string, chunkNum uint64) {
	w.wg.Add(1)
	job := WorkerJob{rawContent: rawContent, filePath: filepath, chunkNum: chunkNum}
	w.JobChan <- job
}

func (w *WorkerPool) WaitToFinish() {
	w.StopPool()
	w.wg.Wait()

	fmt.Println("All jobs finished!")

	if err := w.DB_Client.Disconnect(context.TODO()); err != nil {
		slog.Error("disconnecting database client", "err", err)
	}
}

func (w *WorkerPool) StopPool() {
	close(w.JobChan)

}

func (w *WorkerPool) Work(id uint64) {
	fmt.Printf("Worker %d running!\n", id)
	client := &http.Client{Timeout: 30 * time.Second}
	for job := range w.JobChan {
		embeddingResp := w.CreateVectorEmbedding(string(job.rawContent), client)
		checksum := w.CreateChecksum(job.rawContent)
		w.UpsertData(embeddingResp, checksum, job)
		w.wg.Done()
	}
}

func (w *WorkerPool) UpsertData(embeddingResp EmbeddingResponse, checksum [32]byte, job WorkerJob) {
	metadata := MetaData{Source: job.filePath, ChunkNum: job.chunkNum, Checksum: checksum, Timestamp: time.Now()}
	doc := DataDocument{Embedding: w.ToFloat32Slice(embeddingResp.Embedding[0]), RawContent: string(job.rawContent), MetaData: metadata}
	coll := w.DB_Client.Database("TextData").Collection("LocalFileData")

	_, err := coll.InsertOne(context.TODO(), doc)
	if err != nil {
		slog.Error("Error inserting document into DB", "err", err)
	}
}

func (w *WorkerPool) ToFloat32Slice(input []float64) []float32 {
	out := make([]float32, len(input))
	for i, v := range input {
		out[i] = float32(v)
	}

	return out
}

func (w *WorkerPool) CreateVectorEmbedding(input string, client *http.Client) EmbeddingResponse {
	url := "http://localhost:11434/api/embed"
	embeddingReq := EmbeddingRequest{Model: "qwen3-embedding:4b", Input: input}
	postBody, err := json.Marshal(embeddingReq)
	if err != nil {
		slog.Error("marshalling embedding request", "err", err)
		return EmbeddingResponse{}
	}

	requestBody := bytes.NewBuffer(postBody)

	resp, err := client.Post(url, "application/json", requestBody)
	if err != nil {
		slog.Error("sending post request to Ollama embedding api", "err", err)
		return EmbeddingResponse{}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("error reading response body", "err", err)
		return EmbeddingResponse{}
	}

	// fmt.Println("Got response:", string(body))
	var embeddings EmbeddingResponse
	err = json.Unmarshal(body, &embeddings)
	if err != nil {
		slog.Error("error unmarshalling embedding response", "err", err)
		return EmbeddingResponse{}
	}

	return embeddings
}

func (w *WorkerPool) CreateChecksum(bytes []byte) [32]byte {
	return sha256.Sum256(bytes)
}
