package datastore

import (
	"context"
)

//go:generate mockgen --source repository.go --destination mocks/repository.go -package mocks
type AnswerRepository interface {
	Create(ctx context.Context, answer *Answer) error
	FindByKey(ctx context.Context, key string) (*Answer, error)
	Update(ctx context.Context, answer *Answer, value *Value) (*Answer, error)
	Delete(ctx context.Context, answer *Answer) error
}

type EventRepository interface {
	Create(ctx context.Context, event *Event) error
	FindManyByKey(ctx context.Context, key string, pageable Pageable) ([]Event, PaginationData, error)
}
