//go:build integration
// +build integration

package app

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dotunj/bequest/internal/pkg/datastore/mongo"
)

func getDB(t *testing.T) *mongo.Client {
	db, err := mongo.NewMongoRepository(getTestMongoDSN())
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}

	return db
}

func getApplication(t *testing.T) *Application {
	app, err := NewApplication(getTestMongoDSN())
	if err != nil {
		t.Fatalf("failed to get application: %v", err)
	}

	return app
}

func createRequest(method, url string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, url, body)
	req.Header.Add("Content-Type", "application/json")

	return req
}

func truncateDB(db *mongo.Client, t *testing.T) {
	err := db.DB.Drop(context.TODO())
	if err != nil {
		t.Fatalf("failed to truncate db: %v", err)
	}
}

func parseResponse(t *testing.T, w *http.Response, response interface{}) {
	body, err := io.ReadAll(w.Body)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	sR := struct {
		Success bool            `json:"success"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}{}

	err = json.Unmarshal(body, &sR)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = json.Unmarshal(sR.Data, response)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func getTestMongoDSN() string {
	return os.Getenv("TEST_MONGO_DSN")
}
