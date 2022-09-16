package services

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/dotunj/bequest/internal/pkg/datastore"
	"github.com/dotunj/bequest/internal/pkg/util"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EventService struct {
	answerRepo datastore.AnswerRepository
	eventRepo  datastore.EventRepository
}

func NewEventService(answerRepo datastore.AnswerRepository, eventRepo datastore.EventRepository) *EventService {
	return &EventService{
		answerRepo: answerRepo,
		eventRepo:  eventRepo,
	}
}

func (e *EventService) FindHistoryByKey(ctx context.Context, key string, pageable datastore.Pageable) ([]datastore.Event, datastore.PaginationData, error) {
	answer, err := e.answerRepo.FindByKey(ctx, key)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, datastore.ErrAnswerNotFound) {
			statusCode = http.StatusNotFound
		}
		return nil, datastore.PaginationData{}, util.NewServiceError(statusCode, err)
	}

	events, pagination, err := e.eventRepo.FindManyByKey(ctx, answer.Key, pageable)
	if err != nil {
		return nil, datastore.PaginationData{}, err
	}

	return events, pagination, nil
}

func (e *EventService) CreateEvent(ctx context.Context, answerEvent *datastore.AnswerEvent) (*datastore.Event, error) {
	answer := answerEvent.Answer

	event := &datastore.Event{
		ID:   primitive.NewObjectID(),
		UID:  uuid.NewString(),
		Type: answerEvent.Type,
		Data: &datastore.EventData{
			Key:   answer.Key,
			Value: answer.Values[len(answer.Values)-1].Value,
		},
		CreatedAt:      primitive.NewDateTimeFromTime(time.Now()),
		UpdatedAt:      primitive.NewDateTimeFromTime(time.Now()),
		DocumentStatus: datastore.ActiveDocumentStatus,
	}

	err := e.eventRepo.Create(ctx, event)
	if err != nil {
		return nil, util.NewServiceError(http.StatusInternalServerError, err)
	}

	return event, nil
}
