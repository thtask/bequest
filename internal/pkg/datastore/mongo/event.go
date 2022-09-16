package mongo

import (
	"context"

	"github.com/dotunj/bequest/internal/pkg/datastore"
	pager "github.com/gobeam/mongo-go-pagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type EventRepo struct {
	client *mongo.Collection
}

func NewEventRepo(db *mongo.Database) *EventRepo {
	return &EventRepo{
		client: db.Collection(EventCollection),
	}
}

func (e *EventRepo) Create(ctx context.Context, event *datastore.Event) error {
	_, err := e.client.InsertOne(ctx, event)
	return err
}

func (e *EventRepo) FindManyByKey(ctx context.Context, key string, pageable datastore.Pageable) ([]datastore.Event, datastore.PaginationData, error) {
	var events []datastore.Event

	filter := bson.M{
		"document_status": datastore.ActiveDocumentStatus,
		"data.key":        key,
	}

	paginatedData, err := pager.New(e.client).Context(ctx).Limit(int64(pageable.PerPage)).Page(int64(pageable.Page)).Sort("created_at", pageable.Sort).Filter(filter).Decode(&events).Find()
	if err != nil {
		return events, datastore.PaginationData{}, err
	}

	if events == nil {
		events = make([]datastore.Event, 0)
	}

	return events, datastore.PaginationData(paginatedData.Pagination), nil
}
