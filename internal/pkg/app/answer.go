package app

import (
	"net/http"
	"strconv"

	"github.com/dotunj/bequest/internal/pkg/datastore"
	"github.com/dotunj/bequest/internal/pkg/util"
	"github.com/gin-gonic/gin"
)

func (a *Application) CreateAnswer(c *gin.Context) {
	var createAnswer datastore.CreateAnswer

	if err := c.ShouldBindJSON(&createAnswer); err != nil {
		a.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	answer, err := a.answerService.CreateAnswer(c.Request.Context(), &createAnswer)
	if err != nil {
		status, message := util.NewServiceErrResponse(err)
		a.errorResponse(c, status, message)
		return
	}

	answerResponse := &datastore.AnswerResponse{
		UID:       answer.UID,
		Key:       answer.Key,
		Value:     answer.Values[0].Value,
		CreatedAt: answer.CreatedAt,
		UpdatedAt: answer.UpdatedAt,
	}

	a.successResponse(c, http.StatusCreated, "answer created successfully", answerResponse)
}

func (a *Application) FindAnswerByKey(c *gin.Context) {
	answer, err := a.answerService.FindAnswerByKey(c.Request.Context(), c.Param("key"))
	if err != nil {
		status, message := util.NewServiceErrResponse(err)
		a.errorResponse(c, status, message)
		return
	}

	// Gets the index of the most recent answer
	latestIndex := len(answer.Values) - 1

	answerResponse := &datastore.AnswerResponse{
		UID:       answer.UID,
		Key:       answer.Key,
		Value:     answer.Values[latestIndex].Value,
		CreatedAt: answer.CreatedAt,
		UpdatedAt: answer.UpdatedAt,
	}

	a.successResponse(c, http.StatusOK, "answer retrieved successfully", answerResponse)

}

func (a *Application) UpdateAnswer(c *gin.Context) {
	var updateAnswer datastore.UpdateAnswer

	if err := c.ShouldBindJSON(&updateAnswer); err != nil {
		a.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	answer, err := a.answerService.UpdateAnswer(c.Request.Context(), c.Param("key"), &updateAnswer)
	if err != nil {
		status, message := util.NewServiceErrResponse(err)
		a.errorResponse(c, status, message)
		return
	}

	// Gets the index of the most recent answer
	latestIndex := len(answer.Values) - 1

	answerResponse := &datastore.AnswerResponse{
		UID:       answer.UID,
		Key:       answer.Key,
		Value:     answer.Values[latestIndex].Value,
		CreatedAt: answer.CreatedAt,
		UpdatedAt: answer.UpdatedAt,
	}

	a.successResponse(c, http.StatusOK, "answer updated successfully", answerResponse)
}

func (a *Application) DeleteAnswer(c *gin.Context) {
	err := a.answerService.DeleteAnswer(c.Request.Context(), c.Param("key"))
	if err != nil {
		status, message := util.NewServiceErrResponse(err)
		a.errorResponse(c, status, message)
		return
	}

	a.successResponse(c, http.StatusOK, "answer deleted successfully", nil)
}

func (a *Application) FindHistoryByKey(c *gin.Context) {
	pageable := a.pagination(c)

	events, paginationData, err := a.eventService.FindHistoryByKey(c.Request.Context(), c.Param("key"), pageable)
	if err != nil {
		status, message := util.NewServiceErrResponse(err)
		a.errorResponse(c, status, message)
		return
	}

	pagedResponse := &datastore.PagedResponse{
		Content:    events,
		Pagination: &paginationData,
	}

	a.successResponse(c, http.StatusOK, "retrieved history", pagedResponse)

}

func (a *Application) pagination(c *gin.Context) datastore.Pageable {
	rawPerPage := c.Request.URL.Query().Get("perPage")
	rawPage := c.Request.URL.Query().Get("page")

	if len(rawPerPage) == 0 {
		rawPerPage = "20"
	}

	if len(rawPage) == 0 {
		rawPage = "1"
	}

	var sort = -1 // desc by default
	var perPage, page int
	var err error
	if perPage, err = strconv.Atoi(rawPerPage); err != nil {
		perPage = 20
	}

	if page, err = strconv.Atoi(rawPage); err != nil {
		page = 0
	}

	pageable := datastore.Pageable{
		Page:    page,
		PerPage: perPage,
		Sort:    sort,
	}

	return pageable
}
