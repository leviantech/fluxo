package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/leviantech/fluxo"
)

// Todo represents a task in our API
type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title" validate:"required,min=3"`
	Completed bool   `json:"completed"`
}

// Global in-memory store for demonstration
var (
	todos = make(map[int]Todo)
	nextID = 1
	mu     sync.Mutex
)

// Request/Response types
type CreateTodoRequest struct {
	Title string `json:"title" validate:"required,min=3"`
}

type UpdateTodoRequest struct {
	ID        int    `uri:"id" validate:"required"`
	Title     string `json:"title" validate:"required,min=3"`
	Completed bool   `json:"completed"`
}

type GetTodoRequest struct {
	ID int `uri:"id" validate:"required"`
}

type ListTodosResponse struct {
	Data []Todo `json:"data"`
}

// Middleware: API Key Authentication
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-Api-Key")
		if apiKey != "secret-token" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized: invalid API key"})
			c.Abort()
			return
		}
		c.Next()
	}
}

type ListTodosRequest struct{}

// Handlers
func listTodosHandler(ctx *fluxo.Context, req ListTodosRequest) (ListTodosResponse, error) {
	mu.Lock()
	defer mu.Unlock()

	var result []Todo
	for _, todo := range todos {
		result = append(result, todo)
	}
	// Return empty slice instead of nil to ensure [] in JSON
	if result == nil {
		result = []Todo{}
	}
	return ListTodosResponse{Data: result}, nil
}

func createTodoHandler(ctx *fluxo.Context, req CreateTodoRequest) (Todo, error) {
	mu.Lock()
	defer mu.Unlock()

	newTodo := Todo{
		ID:        nextID,
		Title:     req.Title,
		Completed: false,
	}
	todos[nextID] = newTodo
	nextID++

	return newTodo, nil
}

func getTodoHandler(ctx *fluxo.Context, req GetTodoRequest) (Todo, error) {
	mu.Lock()
	defer mu.Unlock()

	todo, ok := todos[req.ID]
	if !ok {
		return Todo{}, fluxo.NotFound(fmt.Sprintf("todo with ID %d not found", req.ID))
	}
	return todo, nil
}

func updateTodoHandler(ctx *fluxo.Context, req UpdateTodoRequest) (Todo, error) {
	mu.Lock()
	defer mu.Unlock()

	todo, ok := todos[req.ID]
	if !ok {
		return Todo{}, fluxo.NotFound(fmt.Sprintf("todo with ID %d not found", req.ID))
	}

	todo.Title = req.Title
	todo.Completed = req.Completed
	todos[req.ID] = todo

	return todo, nil
}

func deleteTodoHandler(ctx *fluxo.Context, req GetTodoRequest) (gin.H, error) {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := todos[req.ID]; !ok {
		return nil, fluxo.NotFound(fmt.Sprintf("todo with ID %d not found", req.ID))
	}

	delete(todos, req.ID)
	return gin.H{"message": "deleted successfully"}, nil
}

func setupApp() *fluxo.App {
	app := fluxo.New().WithSwagger("Todo Advanced API", "1.0.0")

	// Global middleware
	app.Use(gin.Logger())
	app.Use(gin.Recovery())

	// Public routes
	app.GET("/todos", fluxo.Handle(listTodosHandler))

	// Protected routes using API Key
	protected := app.Group("/api", APIKeyAuth())
	{
		protected.POST("/todos", fluxo.Handle(createTodoHandler))
		protected.GET("/todos/:id", fluxo.Handle(getTodoHandler))
		protected.PUT("/todos/:id", fluxo.Handle(updateTodoHandler))
		protected.DELETE("/todos/:id", fluxo.Handle(deleteTodoHandler))
	}

	return app
}

func main() {
	app := setupApp()
	fmt.Println("🚀 Todo API starting on :8080")
	if err := app.Start(":8080"); err != nil {
		log.Fatal(err)
	}
}
