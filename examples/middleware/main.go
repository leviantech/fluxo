package main

import (
	"fmt"
	"log"

	"github.com/leviantech/fluxo"
)

// AuthHeader defines the structure for our authentication header
type AuthHeader struct {
	Token string `form:"token" validate:"required"` // Bind from query param ?token=...
}

// User is our domain model
type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AuthMiddleware is a type-safe middleware that binds the Token from query params
func AuthMiddleware(ctx *fluxo.Context, req AuthHeader) error {
	fmt.Printf("🔐 Middleware checking token: %s\n", req.Token)
	if req.Token != "valid-token" {
		return fluxo.Unauthorized("Invalid or missing token. Hint: use ?token=valid-token")
	}
	
	// Mock setting an authenticated user into context
	ctx.SetAuthenticatedUser(&User{ID: "user_1", Name: "John Doe"})
	return nil
}

type CreatePostRequest struct {
	Title   string `json:"title" validate:"required,min=5"`
	Content string `json:"content" validate:"required"`
}

type CreatePostResponse struct {
	ID     string `json:"id"`
	Author string `json:"author"`
	Title  string `json:"title"`
}

// CreatePostHandler is the main handler
func CreatePostHandler(ctx *fluxo.Context, req CreatePostRequest) (CreatePostResponse, error) {
	var user User
	// Retrieve user set by middleware
	if err := ctx.GetAuthenticatedUser(&user); err != nil {
		return CreatePostResponse{}, fluxo.InternalServerError("Failed to get user from context")
	}

	fmt.Printf("📝 User %s is creating a post: %s\n", user.Name, req.Title)

	return CreatePostResponse{
		ID:     "post_123",
		Author: user.Name,
		Title:  req.Title,
	}, nil
}

func main() {
	app := fluxo.New().WithSwagger("Type-Safe Middleware Demo", "1.0.0")

	// Route with type-safe middleware and main handler
	// Swagger will automatically merge fields from AuthHeader and CreatePostRequest
	app.POST("/posts", fluxo.Middleware(AuthMiddleware), fluxo.Handle(CreatePostHandler))

	fmt.Println("🚀 Middleware Demo starting on :8080")
	fmt.Println("\n🌟 Test with valid token:")
	fmt.Println(`  curl -X POST "http://localhost:8080/posts?token=valid-token" \`)
	fmt.Println(`       -H "Content-Type: application/json" \`)
	fmt.Println(`       -d '{"title":"Fluxo is Great","content":"Type-safe middleware is awesome"}'`)

	fmt.Println("\n🌟 Test with invalid token (should fail):")
	fmt.Println(`  curl -X POST "http://localhost:8080/posts?token=wrong" \`)
	fmt.Println(`       -d '{"title":"Fail","content":"fail"}'`)

	fmt.Println("\n🌟 Check Swagger docs at http://localhost:8080/docs")

	if err := app.Start(":8080"); err != nil {
		log.Fatal(err)
	}
}
