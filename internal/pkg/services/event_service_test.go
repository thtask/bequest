package services

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/dotunj/bequest/internal/pkg/datastore"
	"github.com/dotunj/bequest/internal/pkg/datastore/mocks"
	"github.com/dotunj/bequest/internal/pkg/util"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func provideEventService(ctrl *gomock.Controller) *EventService {
	answerRepo := mocks.NewMockAnswerRepository(ctrl)
	eventRepo := mocks.NewMockEventRepository(ctrl)

	eventService := NewEventService(answerRepo, eventRepo)
	return eventService
}

func TestEventService_CreateEvent(t *testing.T) {
	type args struct {
		ctx   context.Context
		event *datastore.AnswerEvent
	}

	ctx := context.Background()
	tt := []struct {
		name        string
		args        args
		wantErr     bool
		wantErrMsg  string
		wantErrCode int
		wantEvent   *datastore.Event
		dbFn        func(e *EventService)
	}{
		{
			name: "should_create_event",
			args: args{
				ctx: ctx,
				event: &datastore.AnswerEvent{
					Answer: &datastore.Answer{
						UID:    "12345",
						Key:    "some-key",
						Values: []datastore.Value{{Value: "some-value"}},
					},
					Type: datastore.CreateEvent,
				},
			},
			dbFn: func(e *EventService) {
				eventRepo, _ := e.eventRepo.(*mocks.MockEventRepository)

				eventRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantEvent: &datastore.Event{
				Type: datastore.CreateEvent,
				Data: &datastore.EventData{
					Key:   "some-key",
					Value: "some-value",
				},
			},
		},

		{
			name: "should_fail_to_create_event",
			args: args{
				ctx: ctx,
				event: &datastore.AnswerEvent{
					Answer: &datastore.Answer{
						UID:    "12345",
						Key:    "some-key",
						Values: []datastore.Value{{Value: "some-value"}},
					},
					Type: datastore.CreateEvent,
				},
			},
			dbFn: func(e *EventService) {
				eventRepo, _ := e.eventRepo.(*mocks.MockEventRepository)

				eventRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("failed"))
			},
			wantErr:     true,
			wantErrCode: http.StatusInternalServerError,
			wantErrMsg:  "failed",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			eventService := provideEventService(ctrl)

			if tc.dbFn != nil {
				tc.dbFn(eventService)
			}

			event, err := eventService.CreateEvent(tc.args.ctx, tc.args.event)

			if tc.wantErr {
				require.NotNil(t, err)
				require.Equal(t, tc.wantErrCode, err.(*util.ServiceError).ErrCode())
				require.Equal(t, tc.wantErrMsg, err.(*util.ServiceError).Error())
				return
			}

			require.Nil(t, err)
			require.NotEmpty(t, event.UID)
			require.NotEmpty(t, event.CreatedAt)
			require.NotEmpty(t, event.UpdatedAt)
			require.Empty(t, event.DeletedAt)

			require.Equal(t, tc.wantEvent.Type, event.Type)
			require.Equal(t, tc.wantEvent.Data.Key, event.Data.Key)
			require.Equal(t, tc.wantEvent.Data.Value, event.Data.Value)

		})
	}
}

func TestEventService_FindHistoryByKey(t *testing.T) {
	type args struct {
		ctx      context.Context
		key      string
		pageable datastore.Pageable
	}

	ctx := context.Background()

	tt := []struct {
		name               string
		args               args
		dbFn               func(e *EventService)
		wantEvents         []datastore.Event
		wantPaginationData datastore.PaginationData
	}{
		{
			name: "should_find_history_by_key",
			args: args{
				ctx: ctx,
				key: "some-key",
				pageable: datastore.Pageable{
					Page:    1,
					PerPage: 10,
					Sort:    -1,
				},
			},
			dbFn: func(e *EventService) {

				answerRepo, _ := e.answerRepo.(*mocks.MockAnswerRepository)
				eventRepo, _ := e.eventRepo.(*mocks.MockEventRepository)

				answerRepo.EXPECT().FindByKey(gomock.Any(), gomock.Any()).Return(&datastore.Answer{
					UID: "12345",
					Key: "some-key",
				}, nil)

				eventRepo.EXPECT().FindManyByKey(gomock.Any(), gomock.Any(), gomock.Any()).Return([]datastore.Event{
					{UID: "12345"},
					{UID: "123456"},
				}, datastore.PaginationData{
					Total:     2,
					Page:      1,
					PerPage:   10,
					Prev:      0,
					Next:      2,
					TotalPage: 3,
				}, nil)
			},
			wantEvents: []datastore.Event{
				{UID: "12345"},
				{UID: "123456"},
			},
			wantPaginationData: datastore.PaginationData{
				Total:     2,
				Page:      1,
				PerPage:   10,
				Prev:      0,
				Next:      2,
				TotalPage: 3,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			eventService := provideEventService(ctrl)

			if tc.dbFn != nil {
				tc.dbFn(eventService)
			}

			events, paginationData, err := eventService.FindHistoryByKey(tc.args.ctx, tc.args.key, tc.args.pageable)

			require.Nil(t, err)
			require.Equal(t, tc.wantEvents, events)
			require.Equal(t, tc.wantPaginationData, paginationData)
		})
	}
}
