// Copyright 2025 M Reyhan Fahlevi
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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