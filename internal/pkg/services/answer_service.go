package services

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/dotunj/bequest/internal/pkg/datastore"
	"github.com/dotunj/bequest/internal/pkg/util"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AnswerService struct {
	answerRepo   datastore.AnswerRepository
	eventService *EventService
}

func NewAnswerService(answerRepo datastore.AnswerRepository, eventService *EventService) *AnswerService {
	return &AnswerService{
		answerRepo:   answerRepo,
		eventService: eventService,
	}
}

func (a *AnswerService) CreateAnswer(ctx context.Context, req *datastore.CreateAnswer) (*datastore.Answer, error) {
	answer := &datastore.Answer{
		ID:             primitive.NewObjectID(),
		UID:            uuid.NewString(),
		Key:            req.Key,
		Values:         []datastore.Value{{Value: req.Value}},
		CreatedAt:      primitive.NewDateTimeFromTime(time.Now()),
		UpdatedAt:      primitive.NewDateTimeFromTime(time.Now()),
		DocumentStatus: datastore.ActiveDocumentStatus,
	}

	err := a.answerRepo.Create(ctx, answer)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, datastore.ErrDuplicateKey) {
			statusCode = http.StatusForbidden
		}
		return nil, util.NewServiceError(statusCode, err)
	}

	go a.broadcastEvent(answer, datastore.CreateEvent)

	return answer, nil
}

func (a *AnswerService) FindAnswerByKey(ctx context.Context, key string) (*datastore.Answer, error) {
	answer, err := a.answerRepo.FindByKey(ctx, key)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, datastore.ErrAnswerNotFound) {
			statusCode = http.StatusNotFound
		}
		return nil, util.NewServiceError(statusCode, err)
	}

	return answer, nil
}

func (a *AnswerService) UpdateAnswer(ctx context.Context, key string, req *datastore.UpdateAnswer) (*datastore.Answer, error) {
	value := &datastore.Value{Value: req.Value}

	answer, err := a.FindAnswerByKey(ctx, key)
	if err != nil {
		return nil, err
	}

	answer, err = a.answerRepo.Update(ctx, answer, value)
	if err != nil {
		return nil, util.NewServiceError(http.StatusInternalServerError, err)
	}

	go a.broadcastEvent(answer, datastore.UpdateEvent)
	return answer, nil
}

func (a *AnswerService) DeleteAnswer(ctx context.Context, key string) error {
	answer, err := a.FindAnswerByKey(ctx, key)
	if err != nil {
		return err
	}

	err = a.answerRepo.Delete(ctx, answer)

	if err != nil {
		return util.NewServiceError(http.StatusInternalServerError, err)
	}

	go a.broadcastEvent(answer, datastore.DeleteEvent)
	return nil
}

func (a *AnswerService) broadcastEvent(answer *datastore.Answer, eventType datastore.EventType) {
	ev := &datastore.AnswerEvent{Answer: answer, Type: eventType}
	_, err := a.eventService.CreateEvent(context.Background(), ev)
	if err != nil {
		logrus.WithError(err).Errorf("failed to create event")
	}

}
