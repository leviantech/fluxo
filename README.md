# Fluxo — A Lightweight, Delightful Go Web Framework

Fluxo focuses on developer productivity: type‑safe generic handlers, automatic binding (JSON, query, path, form, multipart), built‑in validation, and automatic Swagger/OpenAPI—without boilerplate. Built on top of **gin** for maximum performance and ecosystem compatibility.

## Key Features
- **Type‑safe generic handlers**: `Handle` (automatic content-type detection)
- **Automatic binding** from multiple sources: `json`, `query`, `path`, `form`, `multipart`
- **Built‑in validation** via `go-playground/validator`
- **Automatic Swagger/OpenAPI** generation with `WithSwagger(title, version)`
- **Full gin integration** - native gin.Context and middleware support
- **Route groups** with shared middleware
- **Zero configuration**, no code generation
- **Production-ready** with gin's battle-tested HTTP engine

## Install
```bash
go get github.com/leviantech/fluxo
```

## Quick Start
```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/leviantech/fluxo"
)

type CreateUserReq struct {
    Email string `json:"email" validate:"required,email"`
    Name  string `json:"name"  validate:"required,min=2"`
    Age   int    `json:"age"   validate:"min=18,max=120"`
}

type CreateUserRes struct {
    ID    string `json:"id"`
    Email string `json:"email"`
    Name  string `json:"name"`
    Age   int    `json:"age"`
}

func CreateUser(ctx *gin.Context, req CreateUserReq) (CreateUserRes, error) {
    return CreateUserRes{ID: "user_123", Email: req.Email, Name: req.Name, Age: req.Age}, nil
}

func main() {
    app := fluxo.New().WithSwagger("User API", "1.0.0")
    app.POST("/users", fluxo.Handle(CreateUser))
    app.Start(":8080")
}
```

## Binding Sources
- JSON body: `json:"..."`
- Query string: `query:"..."`
- Path params: `uri:"..."` (gin native)
- Form & Multipart: `form:"..."`

Example with query + path:
```go
type ListReq struct {
    Page int    `query:"page"`
    ID   string `uri:"id"`      // gin native path parameter
}
```

## Content‑Type Automatic Detection
The unified `Handle` function automatically detects content-type and binds accordingly:

```go
// JSON (application/json) - default
app.POST("/users", fluxo.Handle(CreateUser))

// Form (application/x-www-form-urlencoded) - auto-detected
type LoginReq struct {
    Username string `form:"username" validate:"required"`
    Password string `form:"password" validate:"required"`
}
type LoginRes struct { Token string `json:"token"` }
app.POST("/login", fluxo.Handle(func(ctx *gin.Context, req LoginReq) (LoginRes, error) {
    return LoginRes{Token: "ok"}, nil
}))

// Multipart (multipart/form-data) with file upload - auto-detected
// Single file: *multipart.FileHeader; Multiple files: []*multipart.FileHeader
type UploadReq struct {
    Title string                   `form:"title" validate:"required"`
    File  *multipart.FileHeader    `form:"file"`
    Files []*multipart.FileHeader  `form:"files"`
}
type UploadRes struct { URL string `json:"url"` }
app.POST("/upload", fluxo.Handle(func(ctx *gin.Context, req UploadReq) (UploadRes, error) {
    return UploadRes{URL: "https://example.com/file"}, nil
}))
```

## Validation
- Use `validate:"..."` tags (e.g. `required`, `email`, `min`, `max`, `len`).
- Validation errors return HTTP 400 with formatted messages.

## Gin Integration & Middleware
Fluxo is built on top of **gin**, giving you access to gin's powerful ecosystem:

```go
app := fluxo.New()

// Add gin middleware
app.Use(gin.Logger())
app.Use(gin.Recovery())

// Route groups with middleware
admin := app.Group("/admin", gin.BasicAuth(gin.Accounts{
    "admin": "password",
}))
admin.GET("/dashboard", fluxo.Handle(adminHandler))

// Access gin.Context directly in handlers
func MyHandler(ctx *gin.Context, req MyRequest) (MyResponse, error) {
    // Use any gin.Context method
    clientIP := ctx.ClientIP()
    userAgent := ctx.GetHeader("User-Agent")
    
    return MyResponse{Message: "Hello from gin!"}, nil
}
```

## Automatic Swagger/OpenAPI
- Enable with `app.WithSwagger("Title", "Version")`
- UI: `http://localhost:8080/docs`
- Spec: `http://localhost:8080/openapi.json`
- Fluxo detects `Req`/`Res` types and chooses proper `contentType` (JSON/Form/Multipart) automatically.
- Full OpenAPI 3.0 specification with validation rules

## Performance & Ecosystem
Built on **gin** - one of the fastest Go web frameworks:
- **High performance** HTTP router
- **Battle-tested** in production environments
- **Rich middleware ecosystem** - CORS, rate limiting, authentication, etc.
- **JSON serialization** with optimized libraries
- **Memory efficient** with sync.Pool

## Why Fluxo?
- **Type-safe** handlers with Go generics
- **Zero boilerplate** - automatic binding and validation
- **One handler style** across all content-types (JSON, Form, Multipart) with automatic detection
- **Always-in-sync** API documentation
- **Gin-powered** for maximum performance and ecosystem compatibility
- **Production-ready** with built-in error handling and logging

## License
Apache 2.0 — see `LICENSE`.