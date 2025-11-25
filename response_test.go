package fluxo

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestResponseHelpers(t *testing.T) {
    rr := httptest.NewRecorder()
    _ = JSON(rr, http.StatusOK, map[string]string{"ok":"1"})
    if rr.Code != http.StatusOK { t.Fatalf("status=%d", rr.Code) }
    var m map[string]string
    _ = json.Unmarshal(rr.Body.Bytes(), &m)
    if m["ok"] != "1" { t.Fatalf("json") }

    rr = httptest.NewRecorder()
    _ = Success(rr, map[string]int{"n":2})
    if rr.Code != http.StatusOK { t.Fatalf("success") }

    rr = httptest.NewRecorder()
    _ = Created(rr, map[string]int{"n":3})
    if rr.Code != http.StatusCreated { t.Fatalf("created") }

    rr = httptest.NewRecorder()
    _ = NoContent(rr)
    if rr.Code != http.StatusNoContent { t.Fatalf("nocontent") }
}
