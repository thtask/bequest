package app

import (
	"github.com/dotunj/bequest/internal/pkg/datastore/mongo"
	"github.com/dotunj/bequest/internal/pkg/services"
)

type Application struct {
	DB            *mongo.Client
	answerService *services.AnswerService
	eventService  *services.EventService
}

func NewApplication(dsn string) (*Application, error) {
	db, err := mongo.NewMongoRepository(dsn)
	if err != nil {
		return nil, err
	}

	eventService := services.NewEventService(db.AnswerRepo, db.EventRepo)
	answerService := services.NewAnswerService(db.AnswerRepo, eventService)

	a := &Application{
		DB:            db,
		eventService:  eventService,
		answerService: answerService,
	}
	
	return a, nil
}
