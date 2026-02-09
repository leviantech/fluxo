package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/leviantech/fluxo"
)

// AuthHeader defines the structure for our authentication headers
type AuthHeader struct {
	Authorization string `header:"Authorization" validate:"required"`
	XCustomID     string `header:"X-Custom-ID"`
}

// AuthMiddleware is a type-safe middleware that binds the Authorization header
func AuthMiddleware(ctx *fluxo.Context, req AuthHeader) error {
	fmt.Printf("🔐 Middleware checking Authorization: %s\n", req.Authorization)
	
	// Support Bearer token
	if !strings.HasPrefix(req.Authorization, "Bearer ") {
		return fluxo.Unauthorized("Missing Bearer token")
	}
	
	token := strings.TrimPrefix(req.Authorization, "Bearer ")
	if token != "secret-token" {
		return fluxo.Unauthorized("Invalid token")
	}
	
	fmt.Printf("🆔 Custom ID from header: %s\n", req.XCustomID)
	
	return nil
}

type PingResponse struct {
	Message string `json:"message"`
}

func PingHandler(ctx *fluxo.Context, req any) (PingResponse, error) {
	return PingResponse{Message: "pong"}, nil
}

func main() {
	app := fluxo.New().WithSwagger("Header Auth API", "1.0.0")

	// Protected route using header-based middleware
	app.GET("/ping", fluxo.Middleware(AuthMiddleware), fluxo.Handle(PingHandler))

	fmt.Println("🚀 Header Auth Demo starting on :8080")
	fmt.Println("\n🌟 Test with valid token:")
	fmt.Println(`  curl http://localhost:8080/ping \`)
	fmt.Println(`       -H "Authorization: Bearer secret-token" \`)
	fmt.Println(`       -H "X-Custom-ID: my-app-123"`)

	fmt.Println("\n🌟 Check Swagger docs at http://localhost:8080/docs")

	if err := app.Start(":8080"); err != nil {
		log.Fatal(err)
	}
}
