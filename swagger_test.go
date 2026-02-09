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
