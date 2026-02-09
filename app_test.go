package fluxo

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestApp_Routes_Group_Middleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	app := New().WithSwagger("Test", "1.0.0")

	app.Use(gin.Logger(), gin.Recovery())

	admin := app.Group("/admin", gin.BasicAuth(gin.Accounts{"admin": "password"}))
	admin.GET("/dashboard", Handle(func(ctx *Context, req interface{}) (gin.H, error) { return gin.H{"ok": true}, nil }))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/admin/dashboard", nil)
	app.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestApp_Methods(t *testing.T) {
	gin.SetMode(gin.TestMode)
	app := New()

	handler := Handle(func(ctx *Context, req struct{}) (gin.H, error) {
		return gin.H{"ok": true}, nil
	})

	app.PUT("/put", handler)
	app.DELETE("/delete", handler)
	app.PATCH("/patch", handler)

	methods := []string{"PUT", "DELETE", "PATCH"}
	paths := []string{"/put", "/delete", "/patch"}

	for i, method := range methods {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(method, paths[i], nil)
		app.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 for %s, got %d", method, w.Code)
		}
	}
}

func TestApp_Methods_WithSwagger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	app := New().WithSwagger("Test", "1.0.0")

	handler := Handle(func(ctx *Context, req struct{}) (gin.H, error) {
		return gin.H{"ok": true}, nil
	})

	app.PUT("/put", handler)
	app.DELETE("/delete", handler)
	app.PATCH("/patch", handler)

	// Verify they are in the spec
	spec := app.swagger.Generate(app.handlers)
	paths := spec["paths"].(map[string]interface{})
	if _, ok := paths["/put"]; !ok {
		t.Error("expected /put in swagger spec")
	}
	if _, ok := paths["/delete"]; !ok {
		t.Error("expected /delete in swagger spec")
	}
	if _, ok := paths["/patch"]; !ok {
		t.Error("expected /patch in swagger spec")
	}
}

func TestApp_Start(t *testing.T) {
	app := New()
	// Start in a goroutine on a random port
	go func() {
		_ = app.Start(":0")
	}()
	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)
}

func TestSwagger_UI_Disabled_Panic(t *testing.T) {
	app := New()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	app.EnableSwaggerUI("/docs")
}
