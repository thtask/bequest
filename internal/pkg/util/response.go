package util

import "net/http"

type ServiceError struct {
	errCode int
	errMsg  error
}

func NewServiceError(errCode int, errMsg error) *ServiceError {
	return &ServiceError{errCode: errCode, errMsg: errMsg}
}

func (s *ServiceError) Error() string {
	return s.errMsg.Error()
}

func (s *ServiceError) ErrCode() int {
	return s.errCode
}

func NewServiceErrResponse(err error) (int, string) {
	var msg string
	statusCode := http.StatusBadRequest

	switch v := err.(type) {
	case *ServiceError:
		msg = v.Error()
		statusCode = v.errCode
	case error:
		msg = v.Error()
	}

	return statusCode, msg
}
