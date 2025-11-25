// Copyright 2025 M Reyhan Fahlevi
// Licensed under the MIT License. See LICENSE for details.
package fluxo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

type OpenAPISpec struct {
	OpenAPI    string              `json:"openapi"`
	Info       OpenAPIInfo         `json:"info"`
	Paths      map[string]PathItem `json:"paths"`
	Components Components          `json:"components"`
}

type OpenAPIInfo struct {
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}

type PathItem struct {
	POST   *Operation `json:"post,omitempty"`
	GET    *Operation `json:"get,omitempty"`
	PUT    *Operation `json:"put,omitempty"`
	DELETE *Operation `json:"delete,omitempty"`
	PATCH  *Operation `json:"patch,omitempty"`
}

type Operation struct {
	Summary     string              `json:"summary,omitempty"`
	Description string              `json:"description,omitempty"`
	Parameters  []Parameter         `json:"parameters,omitempty"`
	RequestBody *RequestBody        `json:"requestBody,omitempty"`
	Responses   map[string]Response `json:"responses"`
}

type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Content     map[string]MediaType `json:"content"`
	Required    bool                 `json:"required,omitempty"`
}

type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

type MediaType struct {
	Schema Schema `json:"schema"`
}

type Schema struct {
	Type        string            `json:"type,omitempty"`
	Properties  map[string]Schema `json:"properties,omitempty"`
	Required    []string          `json:"required,omitempty"`
	Items       *Schema           `json:"items,omitempty"`
	Format      string            `json:"format,omitempty"`
	Description string            `json:"description,omitempty"`
	Example     interface{}       `json:"example,omitempty"`
}

type Components struct {
	Schemas map[string]Schema `json:"schemas,omitempty"`
}

type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required"`
	Schema      Schema `json:"schema"`
}

type SwaggerGenerator struct {
	spec      OpenAPISpec
	pageTitle string
}

type SwaggerOption func(*SwaggerGenerator)

func WithSwaggerDescription(desc string) SwaggerOption {
	return func(sg *SwaggerGenerator) {
		sg.spec.Info.Description = desc
	}
}

func WithSwaggerPageTitle(title string) SwaggerOption {
	return func(sg *SwaggerGenerator) {
		sg.pageTitle = title
	}
}

func NewSwaggerGenerator(title, version string, opts ...SwaggerOption) *SwaggerGenerator {
	sg := &SwaggerGenerator{
		spec: OpenAPISpec{
			OpenAPI: "3.0.0",
			Info: OpenAPIInfo{
				Title:       title,
				Version:     version,
				Description: "Auto-generated API documentation",
			},
			Paths: make(map[string]PathItem),
			Components: Components{
				Schemas: make(map[string]Schema),
			},
		},
		pageTitle: title,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(sg)
		}
	}
	return sg
}

// Generate returns the OpenAPI spec as a map (for JSON serialization)
func (sg *SwaggerGenerator) Generate(handlers map[string]handlerInfo) map[string]interface{} {
	// Process all handlers to build the spec
	for _, info := range handlers {
		sg.AddEndpoint(info.method, info.path, info.reqType, info.resType, info.contentType)
	}

	// Convert to map for JSON serialization
	result := make(map[string]interface{})
	data, _ := json.Marshal(sg.spec)
	json.Unmarshal(data, &result)
	return result
}

// detectSwaggerContentTypes analyzes struct tags to determine appropriate content types for swagger
func (sg *SwaggerGenerator) detectSwaggerContentTypes(requestType reflect.Type) []string {
	if requestType == nil {
		return []string{"application/json"}
	}

	var hasJSON, hasForm, hasFile bool

	// Analyze struct fields
	for i := 0; i < requestType.NumField(); i++ {
		field := requestType.Field(i)

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

// generateParameters creates OpenAPI parameters for both query and path parameters
func (sg *SwaggerGenerator) generateParameters(requestType reflect.Type, path string) []Parameter {
	if requestType == nil {
		return nil
	}

	var parameters []Parameter

	// Extract path parameters from the path string (e.g., :id -> id)
	pathParams := extractPathParameters(path)

	for i := 0; i < requestType.NumField(); i++ {
		field := requestType.Field(i)

		// Check for path parameters (uri tags in gin)
		if uriTag := field.Tag.Get("uri"); uriTag != "" && uriTag != "-" {
			paramName := strings.Split(uriTag, ",")[0]
			if paramName == "" {
				continue
			}

			param := Parameter{
				Name:     paramName,
				In:       "path",
				Required: true, // Path parameters are always required
				Schema:   sg.generateSchema(field.Type),
			}

			parameters = append(parameters, param)
			continue
		}

		// Check for query parameters (form tags in gin)
		if formTag := field.Tag.Get("form"); formTag != "" && formTag != "-" {
			paramName := strings.Split(formTag, ",")[0]
			if paramName == "" {
				continue
			}

			// Skip if this is also a path parameter
			if contains(pathParams, paramName) {
				continue
			}

			param := Parameter{
				Name:     paramName,
				In:       "query",
				Required: false, // Query params are typically optional
				Schema:   sg.generateSchema(field.Type),
			}

			// Check if field is required based on validation tags
			if validateTag := field.Tag.Get("validate"); validateTag != "" {
				if strings.Contains(validateTag, "required") {
					param.Required = true
				}
			}

			parameters = append(parameters, param)
		}
	}

	return parameters
}

// extractPathParameters extracts parameter names from path like /users/:id -> [id]
func extractPathParameters(path string) []string {
	var params []string
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			paramName := strings.TrimPrefix(part, ":")
			params = append(params, paramName)
		}
	}
	return params
}

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (sg *SwaggerGenerator) AddEndpoint(method, path string, requestType, responseType reflect.Type, contentType string) {

	operation := &Operation{
		Summary: fmt.Sprintf("%s %s", method, path),
		Responses: map[string]Response{
			"200": {
				Description: "Success",
				Content: map[string]MediaType{
					"application/json": {
						Schema: sg.generateSchema(responseType),
					},
				},
			},
			"400": {
				Description: "Bad Request",
				Content: map[string]MediaType{
					"application/json": {
						Schema: Schema{
							Type: "object",
							Properties: map[string]Schema{
								"status":  {Type: "integer"},
								"message": {Type: "string"},
							},
						},
					},
				},
			},
		},
	}

	if requestType != nil {
		if method == "GET" || method == "HEAD" {
			// For GET/HEAD requests, add query parameters and path parameters
			operation.Parameters = sg.generateParameters(requestType, path)
		} else {
			// For other methods, add request body
			contentTypes := sg.detectSwaggerContentTypes(requestType)

			operation.RequestBody = &RequestBody{
				Description: "Request body",
				Content:     make(map[string]MediaType),
				Required:    true,
			}

			// Add each detected content type
			for _, ct := range contentTypes {
				operation.RequestBody.Content[ct] = MediaType{
					Schema: sg.generateSchema(requestType),
				}
			}
		}
	}

	pathItem, exists := sg.spec.Paths[path]
	if !exists {
		pathItem = PathItem{}
	}

	switch method {
	case "POST":
		pathItem.POST = operation
	case "GET":
		pathItem.GET = operation
	case "PUT":
		pathItem.PUT = operation
	case "DELETE":
		pathItem.DELETE = operation
	case "PATCH":
		pathItem.PATCH = operation
	}

	sg.spec.Paths[path] = pathItem
}

func (sg *SwaggerGenerator) generateSchema(t reflect.Type) Schema {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if isFileHeader(t) {
		return Schema{Type: "string", Format: "binary"}
	}

	switch t.Kind() {
	case reflect.String:
		return Schema{Type: "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return Schema{Type: "integer", Format: "int64"}
	case reflect.Float32, reflect.Float64:
		return Schema{Type: "number", Format: "double"}
	case reflect.Bool:
		return Schema{Type: "boolean"}
	case reflect.Struct:
		return sg.generateStructSchema(t)
	case reflect.Slice:
		it := t.Elem()
		if it.Kind() == reflect.Ptr {
			it = it.Elem()
		}
		if isFileHeader(it) {
			return Schema{Type: "array", Items: &Schema{Type: "string", Format: "binary"}}
		}
		return Schema{Type: "array", Items: &Schema{Type: "object"}}
	default:
		return Schema{Type: "object"}
	}
}

func isFileHeader(t reflect.Type) bool {
	return t.PkgPath() == "mime/multipart" && t.Name() == "FileHeader"
}

func (sg *SwaggerGenerator) generateStructSchema(t reflect.Type) Schema {
	schemaName := t.Name()
	if schemaName == "" {
		schemaName = "Anonymous"
	}

	// Check if we already have this schema
	if existing, ok := sg.spec.Components.Schemas[schemaName]; ok {
		return existing
	}

	schema := Schema{
		Type:       "object",
		Properties: make(map[string]Schema),
		Required:   []string{},
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Try to get field name from json tag first, then form tag
		fieldName := ""
		jsonTag := field.Tag.Get("json")
		formTag := field.Tag.Get("form")

		if jsonTag != "" && jsonTag != "-" {
			fieldName = strings.Split(jsonTag, ",")[0]
		} else if formTag != "" && formTag != "-" {
			fieldName = strings.Split(formTag, ",")[0]
		}

		if fieldName == "" {
			continue
		}

		fieldSchema := sg.generateSchema(field.Type)

		// Add validation info
		if validateTag := field.Tag.Get("validate"); validateTag != "" {
			fieldSchema.Description = "Validation: " + validateTag

			// Parse basic validation rules
			if strings.Contains(validateTag, "email") {
				fieldSchema.Format = "email"
			}
			if strings.Contains(validateTag, "required") {
				schema.Required = append(schema.Required, fieldName)
			}
		}

		schema.Properties[fieldName] = fieldSchema
	}

	// Store the schema for reuse
	sg.spec.Components.Schemas[schemaName] = schema

	return schema
}

func (sg *SwaggerGenerator) GetSpec() OpenAPISpec {
	return sg.spec
}

func (sg *SwaggerGenerator) GetJSON() ([]byte, error) {
	return json.MarshalIndent(sg.spec, "", "  ")
}

// serveSwaggerUI serves the Swagger UI using gin
func (sg *SwaggerGenerator) UIHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Content-Type", "text/html")
		title := sg.pageTitle
		if title == "" {
			title = sg.spec.Info.Title
		}
		ctx.String(http.StatusOK, fmt.Sprintf(swaggerUITemplate, title, "/openapi.json"))
	}
}

const swaggerUITemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>%s</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.9.0/swagger-ui.css">
    <style>
        html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin: 0; background: #fafafa; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            window.ui = SwaggerUIBundle({
                url: "%s",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>
`
