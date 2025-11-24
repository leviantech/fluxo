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
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
)

type App struct {
	router        *gin.Engine
	swagger       *SwaggerGenerator
	enableSwagger bool
	handlers      map[string]handlerInfo // Store handler type information
}

type handlerInfo struct {
	method      string
	path        string
	reqType     reflect.Type
	resType     reflect.Type
	contentType string
}

func New() *App {
	gin.SetMode(gin.ReleaseMode)
	return &App{
		router:        gin.New(),
		enableSwagger: false,
		handlers:      make(map[string]handlerInfo),
	}
}

func (a *App) GET(path string, handler gin.HandlerFunc) {
	// Check if this is a fluxo.Handle wrapper and extract type info if swagger is enabled
	if a.enableSwagger {
		a.captureHandlerInfo("GET", path, handler)
	}
	a.router.GET(path, handler)
}

// POST registers a POST handler
func (a *App) POST(path string, handler gin.HandlerFunc) {
	// Check if this is a fluxo.Handle wrapper and extract type info if swagger is enabled
	if a.enableSwagger {
		a.captureHandlerInfo("POST", path, handler)
	}
	a.router.POST(path, handler)
}

// PUT registers a PUT handler
func (a *App) PUT(path string, handler gin.HandlerFunc) {
	// Check if this is a fluxo.Handle wrapper and extract type info if swagger is enabled
	if a.enableSwagger {
		a.captureHandlerInfo("PUT", path, handler)
	}
	a.router.PUT(path, handler)
}

// DELETE registers a DELETE handler
func (a *App) DELETE(path string, handler gin.HandlerFunc) {
	// Check if this is a fluxo.Handle wrapper and extract type info if swagger is enabled
	if a.enableSwagger {
		a.captureHandlerInfo("DELETE", path, handler)
	}
	a.router.DELETE(path, handler)
}

// PATCH registers a PATCH handler
func (a *App) PATCH(path string, handler gin.HandlerFunc) {
	// Check if this is a fluxo.Handle wrapper and extract type info if swagger is enabled
	if a.enableSwagger {
		a.captureHandlerInfo("PATCH", path, handler)
	}
	a.router.PATCH(path, handler)
}

// Use adds middleware to the gin router
func (a *App) Use(middleware ...gin.HandlerFunc) {
	a.router.Use(middleware...)
}

// Group creates a route group with optional middleware
func (a *App) Group(path string, middleware ...gin.HandlerFunc) *gin.RouterGroup {
	return a.router.Group(path, middleware...)
}

func (a *App) Start(addr string) error {
	return a.router.Run(addr)
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}

// captureHandlerInfo attempts to extract type information from fluxo.Handle wrappers
func (a *App) captureHandlerInfo(method, path string, handler gin.HandlerFunc) {
	reqType, resType, ct, ok := lookupHandlerTypes(handler)
	if !ok {
		return
	}
	handlerKey := fmt.Sprintf("%s:%s", method, path)
	a.handlers[handlerKey] = handlerInfo{
		method:      method,
		path:        path,
		reqType:     reqType,
		resType:     resType,
		contentType: ct,
	}
	if a.swagger != nil {
		a.swagger.AddEndpoint(method, path, reqType, resType, ct)
	}
}

// WithSwagger enables swagger documentation generation and serves it at /docs
func (a *App) WithSwagger(title, version string, opts ...SwaggerOption) *App {
	a.enableSwagger = true
	a.swagger = NewSwaggerGenerator(title, version, opts...)
	a.EnableSwaggerUI("/docs")
	return a
}

// EnableSwaggerUI serves the Swagger UI at the specified path
func (a *App) EnableSwaggerUI(path string) {
	if !a.enableSwagger {
		panic("Swagger is not enabled. Call WithSwagger() first.")
	}

	// Serve the OpenAPI JSON spec (only if not already registered)
	if _, exists := a.handlers["GET:/openapi.json"]; !exists {
		a.GET("/openapi.json", func(c *gin.Context) {
			// Generate the OpenAPI spec dynamically when requested
			spec := a.swagger.Generate(a.handlers)
			c.JSON(http.StatusOK, spec)
		})
	}

	// Serve the Swagger UI
	if path != "/openapi.json" {
		a.GET(path, a.swagger.UIHandler())
	}
}
