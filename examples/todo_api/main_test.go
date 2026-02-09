package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestTodoAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	app := setupApp()

	apiKey := "secret-token"

	t.Run("Create Todo - Unauthorized", func(t *testing.T) {
		body, _ := json.Marshal(CreateTodoRequest{Title: "Test Todo"})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/todos", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		app.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("Create Todo - Success", func(t *testing.T) {
		body, _ := json.Marshal(CreateTodoRequest{Title: "Test Todo"})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/todos", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Api-Key", apiKey)
		
		app.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var resp Todo
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Title != "Test Todo" {
			t.Errorf("Expected title 'Test Todo', got '%s'", resp.Title)
		}
		if resp.ID == 0 {
			t.Error("Expected non-zero ID")
		}
	})

	t.Run("List Todos", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/todos", nil)
		
		app.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var resp ListTodosResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp.Data) == 0 {
			t.Error("Expected at least one todo in the list")
		}
	})

	t.Run("Get Todo - Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/todos/1", nil)
		req.Header.Set("X-Api-Key", apiKey)
		
		app.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Get Todo - Not Found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/todos/999", nil)
		req.Header.Set("X-Api-Key", apiKey)
		
		app.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("Update Todo", func(t *testing.T) {
		body, _ := json.Marshal(UpdateTodoRequest{
			Title:     "Updated Title",
			Completed: true,
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/api/todos/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Api-Key", apiKey)
		
		app.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var resp Todo
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Title != "Updated Title" || !resp.Completed {
			t.Errorf("Update failed: %+v", resp)
		}
	})

	t.Run("Delete Todo", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/api/todos/1", nil)
		req.Header.Set("X-Api-Key", apiKey)
		
		app.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify deletion
		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodGet, "/api/todos/1", nil)
		req2.Header.Set("X-Api-Key", apiKey)
		app.ServeHTTP(w2, req2)
		if w2.Code != http.StatusNotFound {
			t.Errorf("Expected status 404 after deletion, got %d", w2.Code)
		}
	})
}
