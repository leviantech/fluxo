package fluxo

import (
	"fmt"
	"net/http"
	"reflect"
	"sync"

	"github.com/gin-gonic/gin"
)

type typesPair struct {
	req reflect.Type
	res reflect.Type
	ct  string
}

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
func Handle[Req any, Res any](fn func(ctx *gin.Context, req Req) (Res, error)) gin.HandlerFunc {
	var reqZero Req
	var resZero Res
	reqType := reflect.TypeOf(reqZero)
	resType := reflect.TypeOf(resZero)

	handler := func(ctx *gin.Context) {
		var req Req

		// Use gin's native binding based on content type
		if ctx.Request.Method != http.MethodGet && ctx.Request.Method != http.MethodHead {
			contentType := ctx.ContentType()
			
			switch contentType {
			case "application/x-www-form-urlencoded":
				if err := ctx.ShouldBind(&req); err != nil {
					ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Form binding failed: %v", err)})
					return
				}
			case "multipart/form-data":
				if err := ctx.ShouldBind(&req); err != nil {
					ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Multipart binding failed: %v", err)})
					return
				}
			default:
				// JSON binding as default
				if err := ctx.ShouldBindJSON(&req); err != nil {
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

		// Validate the request
		if err := validateStruct(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Validation failed: %v", err)})
			return
		}

		// Call the handler function
		res, err := fn(ctx, req)
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

	// Register with appropriate content type based on common usage
	registerHandlerTypes(handler, reqType, resType, "application/json")
	return handler
}
