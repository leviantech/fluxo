package fluxo

import "fmt"

type HTTPError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.Status, e.Message)
}

func NewHTTPError(status int, message string) HTTPError {
	return HTTPError{
		Status:  status,
		Message: message,
	}
}

func BadRequest(message string) HTTPError {
	return NewHTTPError(400, message)
}

func Unauthorized(message string) HTTPError {
	return NewHTTPError(401, message)
}

func Forbidden(message string) HTTPError {
	return NewHTTPError(403, message)
}

func NotFound(message string) HTTPError {
	return NewHTTPError(404, message)
}

func InternalServerError(message string) HTTPError {
	return NewHTTPError(500, message)
}