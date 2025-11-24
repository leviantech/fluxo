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
	OpenAPI string                 `json:"openapi"`
	Info    OpenAPIInfo            `json:"info"`
	Paths   map[string]PathItem    `json:"paths"`
	Components Components           `json:"components"`
}

type OpenAPIInfo struct {
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}

type PathItem struct {
	POST *Operation `json:"post,omitempty"`
	GET  *Operation `json:"get,omitempty"`
	PUT  *Operation `json:"put,omitempty"`
	DELETE *Operation `json:"delete,omitempty"`
	PATCH *Operation `json:"patch,omitempty"`
}

type Operation struct {
	Summary     string              `json:"summary,omitempty"`
	Description string              `json:"description,omitempty"`
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
	Type        string             `json:"type,omitempty"`
	Properties  map[string]Schema  `json:"properties,omitempty"`
	Required    []string           `json:"required,omitempty"`
	Items       *Schema            `json:"items,omitempty"`
	Format      string             `json:"format,omitempty"`
	Description string             `json:"description,omitempty"`
	Example     interface{}        `json:"example,omitempty"`
}

type Components struct {
	Schemas map[string]Schema `json:"schemas,omitempty"`
}

type SwaggerGenerator struct {
	spec OpenAPISpec
}

func NewSwaggerGenerator(title, version string) *SwaggerGenerator {
	return &SwaggerGenerator{
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
	}
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

func (sg *SwaggerGenerator) AddEndpoint(method, path string, requestType, responseType reflect.Type, contentType string) {
	path = convertChiPathToOpenAPI(path)
	
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
	
    if requestType != nil && method != "GET" && method != "HEAD" {
        operation.RequestBody = &RequestBody{
            Description: "Request body",
            Content: map[string]MediaType{
                contentType: {
                    Schema: sg.generateSchema(requestType),
                },
            },
            Required: true,
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
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		
		fieldName := strings.Split(jsonTag, ",")[0]
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

func convertChiPathToOpenAPI(chiPath string) string {
	// Convert chi path parameters {param} to OpenAPI format {param}
	return chiPath
}

// serveSwaggerUI serves the Swagger UI using gin
func serveSwaggerUI(ctx *gin.Context) {
	ctx.Header("Content-Type", "text/html")
	ctx.String(http.StatusOK, fmt.Sprintf(swaggerUITemplate, "/openapi.json"))
}

const swaggerUITemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Fluxo API Documentation</title>
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