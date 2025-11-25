package main

import (
	"fmt"
	"log"
	"mime/multipart"

	"github.com/gin-gonic/gin"
	"github.com/leviantech/fluxo"
)

// JSON Request/Response types
type CreateUserRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Age      int    `json:"age" validate:"min=1,max=120"`
	IsActive bool   `json:"is_active"`
}

type CreateUserResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Age       int    `json:"age"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

// Form Request type
type LoginFormRequest struct {
	Username string `form:"username" validate:"required"`
	Password string `form:"password" validate:"required"`
	Remember bool   `form:"remember"`
}

type LoginFormResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
	Message string `json:"message"`
}

// Multipart form with file upload
type UploadRequest struct {
	Title       string                `form:"title" validate:"required"`
	Description string                `form:"description"`
	File        *multipart.FileHeader `form:"file" validate:"required"`
}

type UploadResponse struct {
	Success     bool   `json:"success"`
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
}

// Path parameter example
type GetUserRequest struct {
	ID string `uri:"id" validate:"required"`
}

type GetUserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Query parameter example
type SearchRequest struct {
	Query string `form:"q" validate:"required"`
	Limit int    `form:"limit" validate:"max=100"`
	Page  int    `form:"page"`
}

type SearchResponse struct {
	Results []string `json:"results"`
	Total   int      `json:"total"`
	Page    int      `json:"page"`
}

func main() {
	fmt.Println("🚀 Starting Fluxo Framework - Full Gin Integration Demo")
	fmt.Println("=====================================================")

	app := fluxo.New()

	// Enable swagger documentation first (before defining routes)
	app.WithSwagger("Fluxo API Demo", "1.0.0",
		fluxo.WithSwaggerPageTitle("A demo API for Fluxo Framework"))

	// Add some gin middleware
	app.Use(gin.Logger())
	app.Use(gin.Recovery())

	// JSON endpoints
	app.POST("/api/users", fluxo.Handle(createUserHandler))
	app.GET("/api/users/:id", fluxo.Handle(getUserHandler))
	app.GET("/api/search", fluxo.Handle(searchHandler))

	// Form endpoints
	app.POST("/login", fluxo.Handle(loginHandler))

	// Multipart form endpoints
	app.POST("/upload", fluxo.Handle(uploadHandler))

	// Create a route group for admin endpoints
	admin := app.Group("/admin", gin.BasicAuth(gin.Accounts{
		"admin": "password",
	}))
	admin.GET("/dashboard", fluxo.Handle(adminDashboardHandler))

	// Swagger UI is already enabled by WithSwagger, no need to call EnableSwaggerUI again

	fmt.Println("\n🌟 Available endpoints:")
	fmt.Println("  POST /api/users          - Create user (JSON)")
	fmt.Println("  GET  /api/users/:id      - Get user by ID (JSON + path param)")
	fmt.Println("  GET  /api/search?q=...   - Search users (JSON + query param)")
	fmt.Println("  POST /login              - Login (form data)")
	fmt.Println("  POST /upload             - Upload file (multipart form)")
	fmt.Println("  GET  /admin/dashboard    - Admin dashboard (basic auth)")
	fmt.Println("  GET  /docs               - Swagger UI documentation")
	fmt.Println("  GET  /openapi.json       - OpenAPI specification")

	fmt.Println("\n📝 Try these commands:")
	fmt.Println("  # Create user")
	fmt.Println(`  curl -X POST http://localhost:8080/api/users \`)
	fmt.Println(`       -H 'Content-Type: application/json' \`)
	fmt.Println(`       -d '{"name":"Alice","email":"alice@example.com","age":25}'`)
	fmt.Println("")
	fmt.Println("  # Login with form")
	fmt.Println(`  curl -X POST http://localhost:8080/login \`)
	fmt.Println(`       -d 'username=admin&password=secret&remember=true'`)
	fmt.Println("")
	fmt.Println("  # Get user with path param")
	fmt.Println("  curl http://localhost:8080/api/users/123")
	fmt.Println("")
	fmt.Println("  # Search with query params")
	fmt.Println(`  curl 'http://localhost:8080/api/search?q=alice&limit=10&page=1'`)

	fmt.Println("\n🚀 Starting server on :8080...")
	if err := app.Start(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// Handlers
func createUserHandler(ctx *fluxo.Context, req CreateUserRequest) (CreateUserResponse, error) {
	fmt.Printf("Creating user: %+v\n", req)

	return CreateUserResponse{
		ID:        "12345",
		Name:      req.Name,
		Email:     req.Email,
		Age:       req.Age,
		IsActive:  req.IsActive,
		CreatedAt: "2024-01-01T00:00:00Z",
	}, nil
}

func getUserHandler(ctx *fluxo.Context, req GetUserRequest) (GetUserResponse, error) {
	fmt.Printf("Getting user with ID: %s\n", req.ID)

	return GetUserResponse{
		ID:    req.ID,
		Name:  "John Doe",
		Email: "john@example.com",
	}, nil
}

func searchHandler(ctx *fluxo.Context, req SearchRequest) (SearchResponse, error) {
	fmt.Printf("Searching for: %s (limit: %d, page: %d)\n", req.Query, req.Limit, req.Page)

	if req.Limit == 0 {
		req.Limit = 10
	}

	results := []string{
		fmt.Sprintf("Result 1 for '%s'", req.Query),
		fmt.Sprintf("Result 2 for '%s'", req.Query),
		fmt.Sprintf("Result 3 for '%s'", req.Query),
	}

	return SearchResponse{
		Results: results,
		Total:   len(results),
		Page:    req.Page,
	}, nil
}

func loginHandler(ctx *fluxo.Context, req LoginFormRequest) (LoginFormResponse, error) {
	fmt.Printf("Login attempt for user: %s\n", req.Username)

	if req.Username == "admin" && req.Password == "secret" {
		return LoginFormResponse{
			Success: true,
			Token:   "abc123xyz789",
			Message: "Login successful",
		}, nil
	}

	return LoginFormResponse{
		Success: false,
		Token:   "",
		Message: "Invalid credentials",
	}, nil
}

func uploadHandler(ctx *fluxo.Context, req UploadRequest) (UploadResponse, error) {
	fmt.Printf("Uploading file: %s (title: %s)\n", req.File.Filename, req.Title)

	return UploadResponse{
		Success:     true,
		Filename:    req.File.Filename,
		Size:        1024, // Mock size
		ContentType: "application/octet-stream",
	}, nil
}

func adminDashboardHandler(ctx *fluxo.Context, req interface{}) (gin.H, error) {
	user := ctx.MustGet(gin.AuthUserKey).(string)
	fmt.Printf("Admin dashboard accessed by: %s\n", user)

	return gin.H{
		"message": "Welcome to admin dashboard",
		"user":    user,
		"stats": gin.H{
			"total_users":     150,
			"active_sessions": 23,
			"server_uptime":   "99.9%",
		},
	}, nil
}
