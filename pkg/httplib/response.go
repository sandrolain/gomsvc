package httplib

import "fmt"

type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ResponseErrorEnvelope struct {
	Error ResponseError `json:"error,omitempty"`
}

func GetResponseForError(err RouteError) ResponseErrorEnvelope {
	code := err.Code
	if code == "" {
		code = fmt.Sprintf("%v", err.Status)
	}
	return ResponseErrorEnvelope{
		Error: ResponseError{
			Code:    code,
			Message: err.Error(),
		},
	}
}
