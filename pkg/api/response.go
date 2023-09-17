package api

type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ResponseErrorEnvelope struct {
	Error ResponseError `json:"errors,omitempty"`
}

func GetResponseForError(err RouteError) ResponseErrorEnvelope {
	return ResponseErrorEnvelope{
		Error: ResponseError{
			Code:    err.Code,
			Message: err.Error(),
		},
	}
}
