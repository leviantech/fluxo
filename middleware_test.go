package fluxo

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMiddleware_TypeSafe(t *testing.T) {
	gin.SetMode(gin.TestMode)
	app := New()

	type AuthReq struct {
		Token string `form:"token" validate:"required"`
	}

	type MainReq struct {
		Name string `json:"name" validate:"required"`
	}

	type Res struct {
		Msg string `json:"msg"`
	}

	authMid := Middleware(func(ctx *Context, req AuthReq) error {
		if req.Token != "secret" {
			return Unauthorized("bad token")
		}
		ctx.Set("user", "alice")
		return nil
	})

	mainHand := Handle(func(ctx *Context, req MainReq) (Res, error) {
		user := ctx.GetString("user")
		return Res{Msg: "hello " + req.Name + " from " + user}, nil
	})

	app.POST("/test", authMid, mainHand)

	t.Run("Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/test?token=wrong", bytes.NewBufferString(`{"name":"world"}`))
		r.Header.Set("Content-Type", "application/json")
		app.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/test?token=secret", bytes.NewBufferString(`{"name":"world"}`))
		r.Header.Set("Content-Type", "application/json")
		app.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var resp Res
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Msg != "hello world from alice" {
			t.Errorf("unexpected message: %s", resp.Msg)
		}
	})
}

func TestSwagger_MergedTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	app := New().WithSwagger("Merge Test", "1.0")

	type QueryReq struct {
		Page int `form:"page"`
	}
	type BodyReq struct {
		Data string `json:"data"`
	}

	app.POST("/merge", 
		Middleware(func(ctx *Context, req QueryReq) error { return nil }),
		Handle(func(ctx *Context, req BodyReq) (gin.H, error) { return gin.H{"ok": true}, nil }),
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	app.ServeHTTP(w, r)

	var m map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &m)

	paths := m["paths"].(map[string]interface{})
	mergePath := paths["/merge"].(map[string]interface{})
	postOp := mergePath["post"].(map[string]interface{})

	// Should have parameters from QueryReq
	params := postOp["parameters"].([]interface{})
	if len(params) == 0 {
		t.Error("expected parameters from QueryReq")
	}

	// Should have requestBody from BodyReq
	if postOp["requestBody"] == nil {
		t.Error("expected requestBody from BodyReq")
	}
}

func TestSwagger_HeaderMerge(t *testing.T) {
	gin.SetMode(gin.TestMode)
	app := New().WithSwagger("Header Merge Test", "1.0")

	type HeaderReq struct {
		Token string `header:"Authorization" validate:"required"`
	}
	type BodyReq struct {
		Name string `json:"name"`
	}

	app.POST("/header-merge",
		Middleware(func(ctx *Context, req HeaderReq) error { return nil }),
		Handle(func(ctx *Context, req BodyReq) (gin.H, error) { return gin.H{"ok": true}, nil }),
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	app.ServeHTTP(w, r)

	var m map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &m)

	paths := m["paths"].(map[string]interface{})
	pathItem := paths["/header-merge"].(map[string]interface{})
	op := pathItem["post"].(map[string]interface{})

	params := op["parameters"].([]interface{})
	foundHeader := false
	for _, p := range params {
		param := p.(map[string]interface{})
		if param["in"] == "header" && param["name"] == "Authorization" {
			foundHeader = true
			if param["required"] != true {
				t.Error("expected Authorization header to be required")
			}
		}
	}
	if !foundHeader {
		t.Error("expected Authorization header in swagger")
	}
}

func TestMiddleware_HeaderBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)
	app := New()

	type HeaderReq struct {
		Auth string `header:"X-Auth-Token" validate:"required"`
	}

	app.GET("/header", Middleware(func(ctx *Context, req HeaderReq) error {
		if req.Auth != "pass" {
			return Unauthorized("bad header")
		}
		return nil
	}), Handle(func(ctx *Context, req any) (gin.H, error) {
		return gin.H{"ok": true}, nil
	}))

	t.Run("Missing Header", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/header", nil)
		app.ServeHTTP(w, r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("Valid Header", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/header", nil)
		r.Header.Set("X-Auth-Token", "pass")
		app.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
		}
	})
}
