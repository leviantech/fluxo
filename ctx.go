// Copyright 2025 M Reyhan Fahlevi
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package fluxo

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Ctx struct {
	Context  context.Context
	Response http.ResponseWriter
	Request  *http.Request
}

func NewCtx(w http.ResponseWriter, r *http.Request) Ctx {
	return Ctx{
		Context:  r.Context(),
		Response: w,
		Request:  r,
	}
}

func (c *Ctx) PathParam(key string) string {
	ginCtx, ok := c.Request.Context().Value("gin").(*gin.Context)
	if ok {
		return ginCtx.Param(key)
	}
	return ""
}

func (c *Ctx) QueryParam(key string) string {
	return c.Request.URL.Query().Get(key)
}

func (c *Ctx) JSON(status int, data interface{}) error {
	c.Response.Header().Set("Content-Type", "application/json")
	c.Response.WriteHeader(status)
	return json.NewEncoder(c.Response).Encode(data)
}

func (c *Ctx) Error(status int, message string) error {
	return c.JSON(status, HTTPError{
		Status:  status,
		Message: message,
	})
}
