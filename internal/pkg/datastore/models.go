package datastore

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrAnswerNotFound = errors.New("answer not found")
	ErrDuplicateKey   = errors.New("an answer with this key already exists")
)

type DocumentStatus string
type EventType string

const (
	ActiveDocumentStatus  DocumentStatus = "Active"
	DeletedDocumentStatus DocumentStatus = "Deleted"
)

const (
	CreateEvent EventType = "create"
	UpdateEvent EventType = "update"
	DeleteEvent EventType = "delete"
)

type Answer struct {
	ID     primitive.ObjectID `json:"-" bson:"_id"`
	UID    string             `json:"uid" bson:"uid"`
	Key    string             `json:"key" bson:"key"`
	Values []Value            `json:"values" bson:"values"`

	CreatedAt      primitive.DateTime `json:"created_at" bson:"created_at"`
	UpdatedAt      primitive.DateTime `json:"updated_at" bson:"updated_at"`
	DeletedAt      primitive.DateTime `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	DocumentStatus DocumentStatus     `json:"document_status" bson:"document_status"`
}

type Event struct {
	ID   primitive.ObjectID `json:"-" bson:"_id"`
	UID  string             `json:"uid" bson:"uid"`
	Type EventType          `json:"event" bson:"event"`
	Data *EventData         `json:"data" bson:"data"`

	CreatedAt      primitive.DateTime `json:"created_at" bson:"created_at"`
	UpdatedAt      primitive.DateTime `json:"updated_at" bson:"updated_at"`
	DeletedAt      primitive.DateTime `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	DocumentStatus DocumentStatus     `json:"document_status" bson:"document_status"`
}

type AnswerEvent struct {
	Answer *Answer   `json:"answer"`
	Type   EventType `json:"event"`
}

type Pageable struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Sort    int `json:"sort"`
}

type PaginationData struct {
	Total     int64 `json:"total"`
	Page      int64 `json:"page"`
	PerPage   int64 `json:"perPage"`
	Prev      int64 `json:"prev"`
	Next      int64 `json:"next"`
	TotalPage int64 `json:"totalPage"`
}

type EventData struct {
	Key   string `json:"key" bson:"key"`
	Value string `json:"value" bson:"value"`
}

type Value struct {
	Value string `json:"value" bson:"value"`
}

type CreateAnswer struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

type UpdateAnswer struct {
	Value string `json:"value" binding:"required"`
}

type AnswerResponse struct {
	UID       string             `json:"uid"`
	Key       string             `jsn:"key"`
	Value     string             `json:"value"`
	CreatedAt primitive.DateTime `json:"created_at"`
	UpdatedAt primitive.DateTime `json:"updated_at"`
}

type PagedResponse struct {
	Content    interface{}     `json:"content"`
	Pagination *PaginationData `json:"pagination"`
}
