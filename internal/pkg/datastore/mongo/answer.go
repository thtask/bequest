package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/dotunj/bequest/internal/pkg/datastore"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AnswerRepo struct {
	client *mongo.Collection
}

func NewAnswerRepo(db *mongo.Database) *AnswerRepo {
	return &AnswerRepo{
		client: db.Collection(AnswerCollection),
	}
}

func (a *AnswerRepo) Create(ctx context.Context, answer *datastore.Answer) error {
	_, err := a.client.InsertOne(ctx, answer)
	if mongo.IsDuplicateKeyError(err) {
		return datastore.ErrDuplicateKey
	}

	return err
}

func (a *AnswerRepo) FindByKey(ctx context.Context, key string) (*datastore.Answer, error) {
	answer := &datastore.Answer{}
	filter := bson.M{"key": key, "document_status": datastore.ActiveDocumentStatus}

	err := a.client.FindOne(ctx, filter).Decode(answer)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return answer, datastore.ErrAnswerNotFound
	}

	return answer, err
}

func (a *AnswerRepo) Update(ctx context.Context, answer *datastore.Answer, value *datastore.Value) (*datastore.Answer, error) {
	filter := bson.M{"key": answer.Key, "document_status": datastore.ActiveDocumentStatus}
	update := bson.M{
		"$push": bson.M{
			"values": value,
		},
		"$set": bson.M{
			"updated_at": primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	_, err := a.client.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	answer, err = a.FindByKey(ctx, answer.Key)
	if err != nil {
		return nil, err
	}

	return answer, nil
}

func (a *AnswerRepo) Delete(ctx context.Context, answer *datastore.Answer) error {
	filter := bson.M{"key": answer.Key, "document_status": datastore.ActiveDocumentStatus}
	update := bson.M{
		"$set": bson.M{
			"document_status": datastore.DeletedDocumentStatus,
			"deleted_at":      primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	_, err := a.client.UpdateOne(ctx, filter, update)
	return err
}
