//go:build integration
// +build integration

package app

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dotunj/bequest/internal/pkg/datastore"
	"github.com/dotunj/bequest/internal/pkg/datastore/mongo"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AnswerIntegrationTestSuite struct {
	suite.Suite
	DB     *mongo.Client
	Router http.Handler
}

func (a *AnswerIntegrationTestSuite) SetupSuite() {
	a.DB = getDB(a.T())
	app := getApplication(a.T())
	a.Router = app.Routes()
}

func (a *AnswerIntegrationTestSuite) TearDownTest() {
	truncateDB(a.DB, a.T())
}

func (a *AnswerIntegrationTestSuite) Test_CreateAnswer() {
	key := uuid.NewString()
	value := uuid.NewString()

	plainBody := fmt.Sprintf(`{
		"key": "%s",
		"value": "%s"
	}`, key, value)

	body := strings.NewReader(plainBody)
	req := createRequest(http.MethodPost, "/api/v1/answers", body)

	w := httptest.NewRecorder()

	a.Router.ServeHTTP(w, req)

	require.Equal(a.T(), http.StatusCreated, w.Code)

	var answer datastore.AnswerResponse
	parseResponse(a.T(), w.Result(), &answer)

	require.Equal(a.T(), key, answer.Key)
	require.Equal(a.T(), value, answer.Value)
}

func (a *AnswerIntegrationTestSuite) Test_Does_Not_CreateAnswer_With_Existing_Key() {
	key := uuid.NewString()
	value := uuid.NewString()

	//Seed the DB with Existing Answer
	err := a.seedAnswer(key, value)
	require.Nil(a.T(), err)

	plainBody := fmt.Sprintf(`{
		"key": "%s",
		"value": "%s"
	}`, key, value)

	body := strings.NewReader(plainBody)
	req := createRequest(http.MethodPost, "/api/v1/answers", body)

	w := httptest.NewRecorder()

	a.Router.ServeHTTP(w, req)

	require.Equal(a.T(), http.StatusForbidden, w.Code)
}

func (a *AnswerIntegrationTestSuite) Test_CreateAnswer_WithNoKey() {
	value := uuid.NewString()

	plainBody := fmt.Sprintf(`{"value": "%s"}`, value)
	body := strings.NewReader(plainBody)
	req := createRequest(http.MethodPost, "/api/v1/answers", body)

	w := httptest.NewRecorder()

	a.Router.ServeHTTP(w, req)

	require.Equal(a.T(), http.StatusBadRequest, w.Code)
}

func (a *AnswerIntegrationTestSuite) Test_GetAnswer_ExistingKey() {
	key := uuid.NewString()
	value := uuid.NewString()

	//Seed the DB with Existing Answer
	err := a.seedAnswer(key, value)
	require.Nil(a.T(), err)

	req := createRequest(http.MethodGet, fmt.Sprintf("/api/v1/answers/%s", key), nil)

	w := httptest.NewRecorder()

	a.Router.ServeHTTP(w, req)

	require.Equal(a.T(), http.StatusOK, w.Code)

	var answer datastore.AnswerResponse
	parseResponse(a.T(), w.Result(), &answer)

	require.Equal(a.T(), key, answer.Key)
	require.Equal(a.T(), value, answer.Value)
}

func (a *AnswerIntegrationTestSuite) Test_GetAnswer_WithNonExistingKey() {
	key := "key"
	url := fmt.Sprintf("/api/v1/answers/%s", key)

	req := createRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	a.Router.ServeHTTP(w, req)

	require.Equal(a.T(), http.StatusNotFound, w.Code)
}

func (a *AnswerIntegrationTestSuite) Test_UpdateAnswer() {
	key := uuid.NewString()
	value := uuid.NewString()

	//Seed the DB with Existing Answer
	err := a.seedAnswer(key, value)
	require.Nil(a.T(), err)

	newValue := uuid.NewString()
	plainBody := fmt.Sprintf(`{"value": "%s"}`, newValue)
	body := strings.NewReader(plainBody)

	req := createRequest(http.MethodPut, fmt.Sprintf("/api/v1/answers/%s", key), body)

	w := httptest.NewRecorder()

	a.Router.ServeHTTP(w, req)

	require.Equal(a.T(), http.StatusOK, w.Code)

	var answer datastore.AnswerResponse
	parseResponse(a.T(), w.Result(), &answer)

	require.Equal(a.T(), key, answer.Key)
	require.Equal(a.T(), newValue, answer.Value)
}

func (a *AnswerIntegrationTestSuite) Test_UpdateAnswer_WithNonExistingKey() {
	key := uuid.NewString()
	value := uuid.NewString()

	plainBody := fmt.Sprintf(`{"value": "%s"}`, value)
	body := strings.NewReader(plainBody)

	req := createRequest(http.MethodPut, fmt.Sprintf("/api/v1/answers/%s", key), body)

	w := httptest.NewRecorder()

	a.Router.ServeHTTP(w, req)

	require.Equal(a.T(), http.StatusNotFound, w.Code)
}

func (a *AnswerIntegrationTestSuite) Test_DeleteAnswer() {
	key := uuid.NewString()
	value := uuid.NewString()

	//Seed the DB with Existing Answer
	err := a.seedAnswer(key, value)
	require.Nil(a.T(), err)

	req := createRequest(http.MethodDelete, fmt.Sprintf("/api/v1/answers/%s", key), nil)

	w := httptest.NewRecorder()

	a.Router.ServeHTTP(w, req)

	require.Equal(a.T(), http.StatusOK, w.Code)

	// Try to fetch the answer that was deleted
	_, err = a.DB.AnswerRepo.FindByKey(context.Background(), key)
	require.ErrorIs(a.T(), datastore.ErrAnswerNotFound, err)

}

func (a *AnswerIntegrationTestSuite) Test_DeleteAnswer_WithNonExistingKey() {
	key := uuid.NewString()

	req := createRequest(http.MethodDelete, fmt.Sprintf("/api/v1/answers/%s", key), nil)

	w := httptest.NewRecorder()

	a.Router.ServeHTTP(w, req)

	require.Equal(a.T(), http.StatusNotFound, w.Code)
}

func TestAnswerIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(AnswerIntegrationTestSuite))
}

func (a *AnswerIntegrationTestSuite) seedAnswer(key, value string) error {
	answer := &datastore.Answer{
		Key:            key,
		Values:         []datastore.Value{{Value: value}},
		DocumentStatus: datastore.ActiveDocumentStatus,
	}

	return a.DB.AnswerRepo.Create(context.Background(), answer)
}
