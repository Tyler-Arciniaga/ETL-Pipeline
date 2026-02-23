package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/joho/godotenv"
)

func (w *WorkerPool) ProcessQuery(input string) ([]string, error) {
	if err := godotenv.Load(); err != nil {
		slog.Error("loading .env", "err", err)
	}

	err := w.ConnectToDB()
	if err != nil {
		return nil, err
	}

	embeddings := w.GetQueryEmbedding(input)
	results, err := w.QueryDB(embeddings)
	if err != nil {
		return nil, err
	}

	formattedResults := w.FormatResults(results)
	return formattedResults, nil
}

func (w *WorkerPool) FormatResults(results []DataDocument) []string {
	newRes := make([]string, len(results))
	for i, v := range results {
		newRes[i] = v.RawContent
	}

	return newRes
}

func (w *WorkerPool) QueryDB(embeddings []float32) ([]DataDocument, error) {
	coll := w.DB_Client.Database("TextData").Collection("LocalFileData")
	pipeline := mongo.Pipeline{
		{{Key: "$vectorSearch",
			Value: bson.M{
				"index":         "vector_index",
				"path":          "embedding",
				"queryVector":   embeddings,
				"numCandidates": 5,
				"limit":         3,
			}}},
	}

	cursor, err := coll.Aggregate(context.TODO(), pipeline, options.Aggregate())
	if err != nil {
		return nil, fmt.Errorf("Error aggregating documents: %e", err)
	}
	defer cursor.Close(context.Background())

	var results []DataDocument
	if err = cursor.All(context.Background(), &results); err != nil {
		return nil, fmt.Errorf("Error bundling document results: %e", err)
	}

	return results, nil
}

func (w *WorkerPool) GetQueryEmbedding(input string) []float32 {
	client := &http.Client{Timeout: 30 * time.Second}
	embeddingResp := w.CreateVectorEmbedding(input, client)
	return w.ToFloat32Slice(embeddingResp.Embedding[0])
}
