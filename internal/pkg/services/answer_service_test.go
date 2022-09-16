package services

import (
	"context"
	"net/http"
	"sync"
	"testing"

	"github.com/dotunj/bequest/internal/pkg/datastore"
	"github.com/dotunj/bequest/internal/pkg/datastore/mocks"
	"github.com/dotunj/bequest/internal/pkg/util"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func provideAnswerService(ctrl *gomock.Controller) *AnswerService {
	answerRepo := mocks.NewMockAnswerRepository(ctrl)
	eventRepo := mocks.NewMockEventRepository(ctrl)
	eventService := NewEventService(answerRepo, eventRepo)
	answerService := NewAnswerService(answerRepo, eventService)

	return answerService
}

func TestAnswerService_CreateAnswer(t *testing.T) {
	type args struct {
		ctx context.Context
		req *datastore.CreateAnswer
	}

	wg := sync.WaitGroup{}

	ctx := context.Background()
	tt := []struct {
		name        string
		args        args
		wantErr     bool
		wantErrMsg  string
		wantErrCode int
		wantAnswer  *datastore.Answer
		dbFn        func(a *AnswerService)
	}{
		{
			name: "should_create_answer",
			args: args{
				ctx: ctx,
				req: &datastore.CreateAnswer{
					Key:   "some-key",
					Value: "some-value",
				},
			},
			dbFn: func(a *AnswerService) {
				answerRepo, _ := a.answerRepo.(*mocks.MockAnswerRepository)
				eventRepo, _ := a.eventService.eventRepo.(*mocks.MockEventRepository)

				wg.Add(1)
				answerRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

				eventRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ *datastore.Event) error {
					defer wg.Done()
					return nil
				})

			},
			wantAnswer: &datastore.Answer{
				Key:            "some-key",
				Values:         []datastore.Value{{Value: "some-value"}},
				DocumentStatus: datastore.ActiveDocumentStatus,
			},
		},

		{
			name: "should_fail_to_create_answer_for_duplicate_key",
			args: args{
				ctx: ctx,
				req: &datastore.CreateAnswer{
					Key:   "some-key",
					Value: "some-value",
				},
			},
			dbFn: func(a *AnswerService) {
				answerRepo, _ := a.answerRepo.(*mocks.MockAnswerRepository)

				answerRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(datastore.ErrDuplicateKey)
			},
			wantErr:     true,
			wantErrCode: http.StatusForbidden,
			wantErrMsg:  datastore.ErrDuplicateKey.Error(),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			answerService := provideAnswerService(ctrl)

			if tc.dbFn != nil {
				tc.dbFn(answerService)
			}

			answer, err := answerService.CreateAnswer(tc.args.ctx, tc.args.req)
			wg.Wait()

			if tc.wantErr {
				require.NotNil(t, err)
				require.Equal(t, tc.wantErrCode, err.(*util.ServiceError).ErrCode())
				require.Equal(t, tc.wantErrMsg, err.(*util.ServiceError).Error())
				return
			}

			require.Nil(t, err)
			require.NotEmpty(t, answer.UID)
			require.NotEmpty(t, answer.CreatedAt)
			require.NotEmpty(t, answer.UpdatedAt)
			require.Empty(t, answer.DeletedAt)

			require.Equal(t, tc.wantAnswer.Key, answer.Key)
			require.Equal(t, tc.wantAnswer.Values[0].Value, answer.Values[0].Value)
		})
	}
}

func TestAnswerService_FindAnswerByKey(t *testing.T) {
	type args struct {
		ctx context.Context
		key string
	}

	ctx := context.Background()
	tt := []struct {
		name        string
		args        args
		wantErr     bool
		wantErrMsg  string
		wantErrCode int
		wantAnswer  *datastore.Answer
		dbFn        func(a *AnswerService)
	}{
		{
			name: "should_find_answer_by_key",
			args: args{
				ctx: ctx,
				key: "some-key",
			},
			dbFn: func(a *AnswerService) {
				answerRepo, _ := a.answerRepo.(*mocks.MockAnswerRepository)

				answerRepo.EXPECT().FindByKey(gomock.Any(), gomock.Any()).Return(&datastore.Answer{
					UID:    "12345",
					Key:    "some-key",
					Values: []datastore.Value{{Value: "some-value"}},
				}, nil)
			},
			wantAnswer: &datastore.Answer{
				UID:    "12345",
				Key:    "some-key",
				Values: []datastore.Value{{Value: "some-value"}},
			},
	},

		{
			name: "should_fail_to_find_answer_by_key",
			args: args{
				ctx: ctx,
				key: "some-key",
			},
			dbFn: func(a *AnswerService) {
				answerRepo, _ := a.answerRepo.(*mocks.MockAnswerRepository)

				answerRepo.EXPECT().FindByKey(gomock.Any(), gomock.Any()).Return(nil, datastore.ErrAnswerNotFound)
			},
			wantErr:     true,
			wantErrCode: http.StatusNotFound,
			wantErrMsg:  datastore.ErrAnswerNotFound.Error(),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			answerService := provideAnswerService(ctrl)

			if tc.dbFn != nil {
				tc.dbFn(answerService)
			}

			answer, err := answerService.FindAnswerByKey(tc.args.ctx, tc.args.key)

			if tc.wantErr {
				require.NotNil(t, err)
				require.Equal(t, tc.wantErrCode, err.(*util.ServiceError).ErrCode())
				require.Equal(t, tc.wantErrMsg, err.(*util.ServiceError).Error())
				return
			}

			require.Equal(t, tc.wantAnswer.UID, answer.UID)
			require.Equal(t, tc.wantAnswer.Key, answer.Key)
			require.Equal(t, tc.wantAnswer.Values[0].Value, answer.Values[0].Value)
		})
	}
}

func TestAnswerService_UpdateAnswer(t *testing.T) {
	type args struct {
		ctx context.Context
		key string
		req *datastore.UpdateAnswer
	}

	wg := sync.WaitGroup{}

	ctx := context.Background()
	tt := []struct {
		name        string
		args        args
		wantErr     bool
		wantErrMsg  string
		wantErrCode int
		wantAnswer  *datastore.Answer
		dbFn        func(a *AnswerService)
	}{
		{
			name: "should_update_answer",
			args: args{
				ctx: ctx,
				key: "some-key",
				req: &datastore.UpdateAnswer{
					Value: "new-answer",
				},
			},
			dbFn: func(a *AnswerService) {
				answerRepo, _ := a.answerRepo.(*mocks.MockAnswerRepository)
				eventRepo, _ := a.eventService.eventRepo.(*mocks.MockEventRepository)

				wg.Add(1)
				answerRepo.EXPECT().FindByKey(gomock.Any(), gomock.Any()).Return(&datastore.Answer{
					UID:    "12345",
					Key:    "some-key",
					Values: []datastore.Value{{Value: "some-value"}},
				}, nil)

				answerRepo.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Return(&datastore.Answer{
					UID:    "12345",
					Key:    "some-key",
					Values: []datastore.Value{{Value: "some-value"}, {Value: "new-answer"}},
				}, nil)

				eventRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ *datastore.Event) error {
					defer wg.Done()
					return nil
				})
			},
			wantAnswer: &datastore.Answer{
				UID:    "12345",
				Key:    "some-key",
				Values: []datastore.Value{{Value: "some-value"}, {Value: "new-answer"}},
			},
		},

		{
			name: "should_fail_to_update_answer_for_non_existent_key",
			args: args{
				ctx: ctx,
				key: "some-key",
				req: &datastore.UpdateAnswer{
					Value: "new-answer",
				},
			},
			dbFn: func(a *AnswerService) {
				answerRepo, _ := a.answerRepo.(*mocks.MockAnswerRepository)

				answerRepo.EXPECT().FindByKey(gomock.Any(), gomock.Any()).Return(nil, datastore.ErrAnswerNotFound)
			},
			wantErr:     true,
			wantErrCode: http.StatusNotFound,
			wantErrMsg:  datastore.ErrAnswerNotFound.Error(),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			answerService := provideAnswerService(ctrl)

			if tc.dbFn != nil {
				tc.dbFn(answerService)
			}

			answer, err := answerService.UpdateAnswer(tc.args.ctx, tc.args.key, tc.args.req)
			wg.Wait()

			if tc.wantErr {
				require.NotNil(t, err)
				require.Equal(t, tc.wantErrCode, err.(*util.ServiceError).ErrCode())
				require.Equal(t, tc.wantErrMsg, err.(*util.ServiceError).Error())
				return
			}

			require.Nil(t, err)

			require.Equal(t, tc.wantAnswer.UID, answer.UID)
			require.Equal(t, tc.wantAnswer.Key, answer.Key)
			require.Equal(t, tc.wantAnswer.Values[1].Value, answer.Values[1].Value)
		})
	}
}

func TestAnswerService_DeleteAnswer(t *testing.T) {
	type args struct {
		ctx context.Context
		key string
	}

	wg := sync.WaitGroup{}

	ctx := context.Background()
	tt := []struct {
		name        string
		args        args
		wantErr     bool
		wantErrMsg  string
		wantErrCode int
		wantAnswer  *datastore.Answer
		dbFn        func(a *AnswerService)
	}{
		{
			name: "should_delete_answer",
			args: args{
				ctx: ctx,
				key: "some-key",
			},
			dbFn: func(a *AnswerService) {
				answerRepo, _ := a.answerRepo.(*mocks.MockAnswerRepository)
				eventRepo, _ := a.eventService.eventRepo.(*mocks.MockEventRepository)

				wg.Add(1)
				answerRepo.EXPECT().FindByKey(gomock.Any(), gomock.Any()).Return(&datastore.Answer{
					UID:    "12345",
					Key:    "some-key",
					Values: []datastore.Value{{Value: "some-value"}},
				}, nil)

				answerRepo.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil)

				eventRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ *datastore.Event) error {
					defer wg.Done()
					return nil
				})
			},
		},

		{
			name: "should_fail_to_delete_answer_for_non_existent_key",
			args: args{
				ctx: ctx,
				key: "some-key",
			},
			dbFn: func(a *AnswerService) {
				answerRepo, _ := a.answerRepo.(*mocks.MockAnswerRepository)

				answerRepo.EXPECT().FindByKey(gomock.Any(), gomock.Any()).Return(nil, datastore.ErrAnswerNotFound)
			},
			wantErr:     true,
			wantErrCode: http.StatusNotFound,
			wantErrMsg:  datastore.ErrAnswerNotFound.Error(),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			answerService := provideAnswerService(ctrl)

			if tc.dbFn != nil {
				tc.dbFn(answerService)
			}

			err := answerService.DeleteAnswer(tc.args.ctx, tc.args.key)
			wg.Wait()

			if tc.wantErr {
				require.NotNil(t, err)
				require.Equal(t, tc.wantErrCode, err.(*util.ServiceError).ErrCode())
				require.Equal(t, tc.wantErrMsg, err.(*util.ServiceError).Error())
				return
			}

			require.Nil(t, err)
		})
	}
}
