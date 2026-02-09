package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestProductAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	app := setupApp()

	t.Run("Create Product - Validation Error (Code too long)", func(t *testing.T) {
		body, _ := json.Marshal(CreateProductRequest{
			Code:  "TOO_LONG",
			Price: 100,
			Name:  "Test Product",
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		app.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("Create Product - Success", func(t *testing.T) {
		body, _ := json.Marshal(CreateProductRequest{
			Code:  "P001",
			Price: 100,
			Name:  "Laptop",
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		app.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var resp Product
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Code != "P001" || resp.ID == 0 {
			t.Errorf("Unexpected response: %+v", resp)
		}
	})

	t.Run("List Products", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/products", nil)
		
		app.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var resp ProductListResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Total != 1 || len(resp.Items) != 1 {
			t.Errorf("Expected 1 product, got %d total and %d items", resp.Total, len(resp.Items))
		}
	})

	t.Run("Get Product - Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/products/1", nil)
		
		app.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Update Product", func(t *testing.T) {
		body, _ := json.Marshal(UpdateProductRequest{
			Price: 150,
			Name:  "Updated Laptop",
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/products/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		app.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var resp Product
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Price != 150 || resp.Name != "Updated Laptop" {
			t.Errorf("Update failed: %+v", resp)
		}
	})

	t.Run("Delete Product", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/products/1", nil)
		
		app.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify deletion
		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodGet, "/products/1", nil)
		app.ServeHTTP(w2, req2)
		if w2.Code != http.StatusNotFound {
			t.Errorf("Expected status 404 after deletion, got %d", w2.Code)
		}
	})
}
