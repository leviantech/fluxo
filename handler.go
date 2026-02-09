// Copyright 2025 M Reyhan Fahlevi
// Licensed under the MIT License. See LICENSE for details.
package fluxo

import (
	"fmt"
	"net/http"
	"reflect"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type typesPair struct {
	req reflect.Type
	res reflect.Type
	ct  string
}

type HandlerFunc[Req any, Res any] func(ctx *Context, req Req) (Res, error)

type MiddlewareFunc[Req any] func(ctx *Context, req Req) error

var handlerTypeRegistry sync.Map

func registerHandlerTypes(h gin.HandlerFunc, req, res reflect.Type, ct string) {
	handlerTypeRegistry.Store(reflect.ValueOf(h).Pointer(), typesPair{req: req, res: res, ct: ct})
}

func lookupHandlerTypes(h gin.HandlerFunc) (reflect.Type, reflect.Type, string, bool) {
	if v, ok := handlerTypeRegistry.Load(reflect.ValueOf(h).Pointer()); ok {
		p := v.(typesPair)
		return p.req, p.res, p.ct, true
	}
	return nil, nil, "", false
}

// Handle creates a type-safe handler using gin's native functionality with automatic content-type detection
func Handle[Req any, Res any](fn HandlerFunc[Req, Res]) gin.HandlerFunc {
	var reqZero Req
	var resZero Res
	reqType := reflect.TypeOf(reqZero)
	resType := reflect.TypeOf(resZero)

	handler := func(ctx *gin.Context) {
		var req Req

		// Use gin's native binding based on content type
		if ctx.Request.Method != http.MethodGet && ctx.Request.Method != http.MethodHead && ctx.Request.ContentLength != 0 {
			contentType := ctx.ContentType()

			switch contentType {
			case gin.MIMEPOSTForm:
				if err := ctx.ShouldBind(&req); err != nil {
					ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Form binding failed: %v", err)})
					return
				}
			case gin.MIMEMultipartPOSTForm:
				if err := ctx.ShouldBind(&req); err != nil {
					ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Multipart binding failed: %v", err)})
					return
				}
			default:
				// JSON binding as default (use ShouldBindBodyWith to allow multiple reads)
				if err := ctx.ShouldBindBodyWith(&req, binding.JSON); err != nil {
					ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("JSON binding failed: %v", err)})
					return
				}
			}
		}

		// Bind query parameters using gin's native binding
		if err := ctx.ShouldBindQuery(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Query binding failed: %v", err)})
			return
		}

		// Bind path parameters using gin's native binding
		if err := ctx.ShouldBindUri(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Path binding failed: %v", err)})
			return
		}

		// Bind header parameters using gin's native binding
		if err := ctx.ShouldBindHeader(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Header binding failed: %v", err)})
			return
		}

		// Validate the request if it's a struct
		if reqType != nil && (reqType.Kind() == reflect.Struct || (reqType.Kind() == reflect.Ptr && reqType.Elem().Kind() == reflect.Struct)) {
			if err := validateStruct(ctx, &req); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Validation failed: %v", err)})
				return
			}
		}

		// Call the handler function
		res, err := fn(&Context{Context: ctx}, req)
		if err != nil {
			if httpErr, ok := err.(HTTPError); ok {
				ctx.JSON(httpErr.Status, httpErr)
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Internal server error: %v", err)})
			}
			return
		}

		// Return success response
		ctx.JSON(http.StatusOK, res)
	}

	// Determine content types based on struct tags
	contentTypes := detectContentTypes(reqType)

	// Register handler types for each detected content type
	for _, ct := range contentTypes {
		registerHandlerTypes(handler, reqType, resType, ct)
	}
	return handler
}

// Middleware creates a type-safe middleware using gin's native functionality with automatic content-type detection
func Middleware[Req any](fn MiddlewareFunc[Req]) gin.HandlerFunc {
	var reqZero Req
	reqType := reflect.TypeOf(reqZero)

	handler := func(ctx *gin.Context) {
		var req Req

		// Use gin's native binding based on content type
		if ctx.Request.Method != http.MethodGet && ctx.Request.Method != http.MethodHead && ctx.Request.ContentLength != 0 {
			contentType := ctx.ContentType()

			switch contentType {
			case gin.MIMEPOSTForm:
				if err := ctx.ShouldBind(&req); err != nil {
					ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Form binding failed: %v", err)})
					ctx.Abort()
					return
				}
			case gin.MIMEMultipartPOSTForm:
				if err := ctx.ShouldBind(&req); err != nil {
					ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Multipart binding failed: %v", err)})
					ctx.Abort()
					return
				}
			default:
				// JSON binding as default (use ShouldBindBodyWith to allow multiple reads)
				if err := ctx.ShouldBindBodyWith(&req, binding.JSON); err != nil {
					ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("JSON binding failed: %v", err)})
					ctx.Abort()
					return
				}
			}
		}

		// Bind query parameters using gin's native binding
		if err := ctx.ShouldBindQuery(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Query binding failed: %v", err)})
			ctx.Abort()
			return
		}

		// Bind path parameters using gin's native binding
		if err := ctx.ShouldBindUri(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Path binding failed: %v", err)})
			ctx.Abort()
			return
		}

		// Bind header parameters using gin's native binding
		if err := ctx.ShouldBindHeader(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Header binding failed: %v", err)})
			ctx.Abort()
			return
		}

		// Validate the request if it's a struct
		if reqType != nil && (reqType.Kind() == reflect.Struct || (reqType.Kind() == reflect.Ptr && reqType.Elem().Kind() == reflect.Struct)) {
			if err := validateStruct(ctx, &req); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Validation failed: %v", err)})
				ctx.Abort()
				return
			}
		}

		// Call the middleware function
		err := fn(&Context{Context: ctx}, req)
		if err != nil {
			if httpErr, ok := err.(HTTPError); ok {
				ctx.JSON(httpErr.Status, httpErr)
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Internal server error: %v", err)})
			}
			ctx.Abort()
			return
		}

		ctx.Next()
	}

	// Determine content types based on struct tags
	contentTypes := detectContentTypes(reqType)

	// Register middleware types for each detected content type (use nil for response type)
	for _, ct := range contentTypes {
		registerHandlerTypes(handler, reqType, nil, ct)
	}
	return handler
}

// detectContentTypes analyzes struct tags to determine appropriate content types
func detectContentTypes(reqType reflect.Type) []string {
	if reqType == nil || reqType.Kind() != reflect.Struct {
		return []string{"application/json"}
	}

	var hasJSON, hasForm, hasFile bool

	// Analyze struct fields
	for i := 0; i < reqType.NumField(); i++ {
		field := reqType.Field(i)

		// Check for json tags
		if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
			hasJSON = true
		}

		// Check for form tags
		if formTag := field.Tag.Get("form"); formTag != "" && formTag != "-" {
			hasForm = true
		}

		// Check for file upload fields
		if field.Type.String() == "*multipart.FileHeader" ||
			field.Type.String() == "[]*multipart.FileHeader" {
			hasFile = true
		}
	}

	// Determine content types based on analysis
	var contentTypes []string

	if hasFile {
		// If there are file fields, must use multipart
		contentTypes = append(contentTypes, "multipart/form-data")
	} else if hasForm {
		// If there are form tags, support both form and JSON
		contentTypes = append(contentTypes, "application/x-www-form-urlencoded")
		if hasJSON {
			contentTypes = append(contentTypes, "application/json")
		}
	} else if hasJSON {
		// If only JSON tags, use JSON
		contentTypes = append(contentTypes, "application/json")
	} else {
		// Default to JSON
		contentTypes = append(contentTypes, "application/json")
	}

	return contentTypes
}
