package fluxo

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
