package fluxo

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type vt struct {
	Email string `validate:"required,email"`
	Age   int    `validate:"min=18"`
}

func TestValidateStruct(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Language", "en")

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	err := validateStruct(ctx, &vt{Email: "bad", Age: 10})
	if err == nil {
		t.Fatalf("expected error")
	}

	if err2 := validateStruct(ctx, &vt{Email: "a@b.com", Age: 25}); err2 != nil {
		t.Fatalf("unexpected %v", err2)
	}
}
