package fluxo

import "testing"

func TestHTTPErrorHelpers(t *testing.T) {
    if BadRequest("x").Status != 400 { t.Fatalf("bad request") }
    if Unauthorized("x").Status != 401 { t.Fatalf("unauthorized") }
    if Forbidden("x").Status != 403 { t.Fatalf("forbidden") }
    if NotFound("x").Status != 404 { t.Fatalf("notfound") }
    if InternalServerError("x").Status != 500 { t.Fatalf("ise") }

    e := NewHTTPError(418, "teapot")
    if e.Error() == "" { t.Fatalf("error string empty") }
}
