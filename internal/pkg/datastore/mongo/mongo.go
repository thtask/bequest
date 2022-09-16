package mongo

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/dotunj/bequest/internal/pkg/datastore"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	AnswerCollection = "answers"
	EventCollection  = "events"
)

type Client struct {
	DB         *mongo.Database
	AnswerRepo datastore.AnswerRepository
	EventRepo  datastore.EventRepository
}

func NewMongoRepository(dsn string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dsn))
	if err != nil {
		return nil, err
	}

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}

	name := strings.TrimPrefix(u.Path, "/")
	conn := client.Database(name, nil)

	c := &Client{
		DB:         conn,
		AnswerRepo: NewAnswerRepo(conn),
		EventRepo:  NewEventRepo(conn),
	}

	// Ensure a unique index is created for the key
	// field on the answer collection
	c.createUniqueIndex(AnswerCollection, "key")

	return c, nil
}


func (c *Client) createUniqueIndex(collectionName, fieldName string) bool {
	unique := true

	createIndexOpts := &options.IndexOptions{Unique: &unique}

	mod := mongo.IndexModel{ // index in ascending order or -1 for descending order
		Keys: bson.D{
			{Key: fieldName, Value: 1},
			{Key: "document_status", Value: 1},
		},
		Options: createIndexOpts,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := c.DB.Collection(collectionName)

	_, err := collection.Indexes().CreateOne(ctx, mod)
	if err != nil {
		logrus.WithError(err).Errorf("failed to create index on field %s in %s", fieldName, collectionName)
		return false
	}

	return true
}
