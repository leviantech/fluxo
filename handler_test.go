package fluxo

import (
    "bytes"
    "mime/multipart"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/gin-gonic/gin"
)

type htCreateUserReq struct {
    Email string `json:"email" validate:"required,email"`
    Name  string `json:"name" validate:"required,min=2"`
    Age   int    `json:"age" validate:"min=18,max=120"`
}
type htCreateUserRes struct {
    ID    string `json:"id"`
    Email string `json:"email"`
    Name  string `json:"name"`
    Age   int    `json:"age"`
}

func htCreateUser(ctx *Context, req htCreateUserReq) (htCreateUserRes, error) {
    if req.Name == "error" {
        return htCreateUserRes{}, BadRequest("bad")
    }
    return htCreateUserRes{ID: "u1", Email: req.Email, Name: req.Name, Age: req.Age}, nil
}

type htLoginReq struct {
    Username string `form:"username" validate:"required"`
    Password string `form:"password" validate:"required"`
}
type htLoginRes struct { Token string `json:"token"` }

func htLogin(ctx *Context, req htLoginReq) (htLoginRes, error) { return htLoginRes{Token: req.Username + ":ok"}, nil }

type htUploadReq struct {
    Title string                `form:"title" validate:"required"`
    File  *multipart.FileHeader `form:"file"`
}
type htUploadRes struct { Name string `json:"name"` }
func htUpload(ctx *Context, req htUploadReq) (htUploadRes, error) { return htUploadRes{Name: req.File.Filename}, nil }

type htGetReq struct {
    ID    string `uri:"id"`
    Limit int    `form:"limit"`
}
type htGetRes struct { ID string `json:"id"` }
func htGet(ctx *Context, req htGetReq) (htGetRes, error) { return htGetRes{ID: req.ID}, nil }

func TestHandle_JSON_Validation_ErrorMapping(t *testing.T) {
    gin.SetMode(gin.TestMode)
    app := New().WithSwagger("Test API", "1.0.0")
    app.POST("/users", Handle(htCreateUser))

    w := httptest.NewRecorder()
    r := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"email":"a@b.com","name":"Ab","age":20}`))
    r.Header.Set("Content-Type", "application/json")
    app.ServeHTTP(w, r)
    if w.Code != http.StatusOK { t.Fatalf("status=%d", w.Code) }

    w2 := httptest.NewRecorder()
    r2 := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"email":"bad","name":"A","age":17}`))
    r2.Header.Set("Content-Type", "application/json")
    app.ServeHTTP(w2, r2)
    if w2.Code != http.StatusBadRequest { t.Fatalf("status=%d", w2.Code) }

    app.POST("/err", Handle(func(ctx *Context, req htCreateUserReq) (htCreateUserRes, error) { return htCreateUserRes{}, NotFound("no") }))
    w3 := httptest.NewRecorder()
    r3 := httptest.NewRequest(http.MethodPost, "/err", strings.NewReader(`{"email":"a@b.com","name":"Ab","age":20}`))
    r3.Header.Set("Content-Type", "application/json")
    app.ServeHTTP(w3, r3)
    if w3.Code != http.StatusNotFound { t.Fatalf("status=%d", w3.Code) }
}

func TestHandle_Form_Multipart_Query_Path(t *testing.T) {
    gin.SetMode(gin.TestMode)
    app := New().WithSwagger("Test API", "1.0.0")
    app.POST("/login", Handle(htLogin))
    app.POST("/upload", Handle(htUpload))
    app.GET("/users/:id", Handle(htGet))

    w := httptest.NewRecorder()
    r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("username=foo&password=bar"))
    r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    app.ServeHTTP(w, r)
    if w.Code != http.StatusOK { t.Fatalf("status=%d", w.Code) }

    buf := &bytes.Buffer{}
    mw := multipart.NewWriter(buf)
    _ = mw.WriteField("title", "t")
    fw, _ := mw.CreateFormFile("file", "x.txt")
    _, _ = fw.Write([]byte("hi"))
    _ = mw.Close()
    w2 := httptest.NewRecorder()
    r2 := httptest.NewRequest(http.MethodPost, "/upload", buf)
    r2.Header.Set("Content-Type", mw.FormDataContentType())
    app.ServeHTTP(w2, r2)
    if w2.Code != http.StatusOK { t.Fatalf("status=%d", w2.Code) }

    w3 := httptest.NewRecorder()
    r3 := httptest.NewRequest(http.MethodGet, "/users/123?limit=5", nil)
    app.ServeHTTP(w3, r3)
    if w3.Code != http.StatusOK { t.Fatalf("status=%d", w3.Code) }
}
