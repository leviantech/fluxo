package fluxo

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "reflect"
    "strings"
    "testing"
    mimeMultipart "mime/multipart"

    "github.com/gin-gonic/gin"
)

func TestSwagger_Title_Description_UI(t *testing.T) {
    gin.SetMode(gin.TestMode)
    app := New().WithSwagger("Docs Title", "1.0.0", WithSwaggerDescription("Desc"), WithSwaggerPageTitle("Page Title"))
    app.POST("/users", Handle(func(ctx *Context, req struct{ Name string `json:"name"` }) (struct{ OK bool `json:"ok"` }, error) { return struct{ OK bool `json:"ok"` }{true}, nil }))

    w := httptest.NewRecorder()
    r := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
    app.ServeHTTP(w, r)
    if w.Code != http.StatusOK { t.Fatalf("status=%d", w.Code) }
    var m map[string]interface{}
    _ = json.Unmarshal(w.Body.Bytes(), &m)
    info := m["info"].(map[string]interface{})
    if info["title"].(string) != "Docs Title" { t.Fatalf("title") }
    if info["description"].(string) != "Desc" { t.Fatalf("desc") }

    w2 := httptest.NewRecorder()
    r2 := httptest.NewRequest(http.MethodGet, "/docs", nil)
    app.ServeHTTP(w2, r2)
    if w2.Code != http.StatusOK { t.Fatalf("status=%d", w2.Code) }
    if !strings.Contains(w2.Body.String(), "<title>Page Title</title>") { t.Fatalf("ui title") }
}

func TestSwagger_ContentTypes_Parameters(t *testing.T) {
    sg := NewSwaggerGenerator("t", "v")
    type S struct {
        A string `json:"a"`
        B string `form:"b"`
        F *mimeMultipart.FileHeader `form:"file"`
    }
    cts := sg.detectSwaggerContentTypes(reflect.TypeOf(S{}))
    // Must contain multipart/form-data
    found := false
    for _, ct := range cts { if ct == "multipart/form-data" { found = true; break } }
    if !found { t.Fatalf("missing multipart") }

    type P struct {
        ID string `uri:"id"`
        Q  int    `form:"q"`
    }
    params := sg.generateParameters(reflect.TypeOf(P{}), "/items/:id")
    if len(params) == 0 { t.Fatalf("no params") }
}

func TestSwagger_NestedSlice(t *testing.T) {
	sg := NewSwaggerGenerator("t", "v")
	type Item struct {
		Name string `json:"name"`
	}
	type Res struct {
		Items []Item `json:"items"`
	}
	
	schema := sg.generateSchema(reflect.TypeOf(Res{}))
	itemsSchema := schema.Properties["items"]
	if itemsSchema.Type != "array" {
		t.Fatalf("expected array, got %s", itemsSchema.Type)
	}
	
	// Check if items schema has properties (it should, but current implementation might just say "object")
	if itemsSchema.Items == nil {
		t.Fatalf("expected items schema to be non-nil")
	}
	
	// This is where it's expected to fail if the bug exists
	if itemsSchema.Items.Type == "object" && len(itemsSchema.Items.Properties) == 0 {
		t.Errorf("nested slice items have no properties, probably just generic object")
	}
}

func TestSwagger_GetSpec_GetJSON(t *testing.T) {
	sg := NewSwaggerGenerator("Test", "1.0")
	spec := sg.GetSpec()
	if spec.Info.Title != "Test" {
		t.Fatalf("expected Test, got %s", spec.Info.Title)
	}

	jsonBytes, err := sg.GetJSON()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if len(jsonBytes) == 0 {
		t.Fatalf("expected non-empty JSON")
	}
}

func TestSwagger_SchemaTypes(t *testing.T) {
	sg := NewSwaggerGenerator("Test", "1.0")
	type AllTypes struct {
		Int     int     `json:"int"`
		Float   float64 `json:"float"`
		Bool    bool    `json:"bool"`
		Pointer *string `json:"pointer"`
	}

	schema := sg.generateSchema(reflect.TypeOf(AllTypes{}))
	if schema.Properties["int"].Type != "integer" {
		t.Errorf("expected integer, got %s", schema.Properties["int"].Type)
	}
	if schema.Properties["float"].Type != "number" {
		t.Errorf("expected number, got %s", schema.Properties["float"].Type)
	}
	if schema.Properties["bool"].Type != "boolean" {
		t.Errorf("expected boolean, got %s", schema.Properties["bool"].Type)
	}
	if schema.Properties["pointer"].Type != "string" {
		t.Errorf("expected string for pointer, got %s", schema.Properties["pointer"].Type)
	}
}

func TestSwagger_UI_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	t.Run("With_PageTitle", func(t *testing.T) {
		app := New().WithSwagger("Test", "1.0.0", WithSwaggerPageTitle("Custom Title"))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/docs", nil)
		app.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("Without_PageTitle", func(t *testing.T) {
		sg := NewSwaggerGenerator("Test", "1.0.0")
		// Manually trigger UIHandler
		handler := sg.UIHandler()
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		handler(c)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
}

func TestSwagger_Internal_Helpers(t *testing.T) {
	t.Run("contains", func(t *testing.T) {
		if !contains([]string{"a", "b"}, "a") {
			t.Error("expected true")
		}
		if contains([]string{"a", "b"}, "c") {
			t.Error("expected false")
		}
	})
}

func TestSwagger_Schema_EdgeCases(t *testing.T) {
	sg := NewSwaggerGenerator("Test", "1.0.0")

	t.Run("Recursive_Schema", func(t *testing.T) {
		type Node struct {
			Name string `json:"name"`
			Next *Node  `json:"next"`
		}
		schema := sg.generateSchema(reflect.TypeOf(Node{}))
		if schema.Type != "object" {
			t.Errorf("expected object, got %s", schema.Type)
		}
	})

	t.Run("Primitive_Types", func(t *testing.T) {
		types := []reflect.Type{
			reflect.TypeOf(""),
			reflect.TypeOf(1),
			reflect.TypeOf(1.1),
			reflect.TypeOf(true),
		}
		for _, rt := range types {
			schema := sg.generateSchema(rt)
			if schema.Type == "" {
				t.Errorf("expected type for %v", rt)
			}
		}
	})

	t.Run("Comprehensive_Struct", func(t *testing.T) {
		type ComplexReq struct {
			ID      int    `uri:"id" validate:"required"`
			Search  string `form:"q"`
			Token   string `header:"X-Token" validate:"required"`
			Data    struct {
				Value string `json:"value"`
			} `json:"data"`
			Tags    []string `json:"tags"`
			Active  bool     `json:"active"`
			Price   float64  `json:"price"`
		}
		sg.AddEndpoint("POST", "/test/:id", []reflect.Type{reflect.TypeOf(ComplexReq{})}, nil, "application/json")
		spec := sg.GetSpec()
		if _, ok := spec.Paths["/test/:id"]; !ok {
			t.Error("expected /test/:id in spec")
		}
	})
}
