package api

type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ResponseEnvelope[T any] struct {
	Success bool          `json:"success"`
	Payload T             `json:"payload,omitempty"`
	Error   ResponseError `json:"errors,omitempty"`
}

func GetResponse[T any](payload T) ResponseEnvelope[T] {
	return ResponseEnvelope[T]{
		Success: true,
		Payload: payload,
	}
}

func GetResponseForError(err RouteError) ResponseEnvelope[interface{}] {
	return ResponseEnvelope[interface{}]{
		Error: ResponseError{
			Code:    err.Code,
			Message: err.Error.Error(),
		},
	}
}
