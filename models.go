package main

import (
	"time"
	// "go.mongodb.org/mongo-driver/bson/primitive"
)

type MetaData struct {
	Source    string    `bson:"source"`
	ChunkNum  uint64    `bson:"chunk_num"`
	Checksum  [32]byte  `bson:"checksum"`
	Timestamp time.Time `bson:"timestamp"`
}

type DataDocument struct {
	// ID         primitive.ObjectID `bson:"_id,omitempty"`
	Embedding  []float32 `bson:"embedding"`
	RawContent string    `bson:"raw_content"`
	MetaData   MetaData  `bson:"metadata"`
}

type WorkerJob struct {
	rawContent []byte
	filePath   string
	chunkNum   uint64
}

type EmbeddingResponse struct {
	Embedding [][]float64 `json:"embeddings"`
}

type CommandFlags struct {
	Dir     string
	Workers uint64
	Query   bool
	Input   string
}
