package fluxo

import (
	"net/http/httptest"
	"strings"
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

func TestRegisterTranslation(t *testing.T) {
	RegisterTranslation("id", "required", "%s harus diisi")
	
	w := httptest.NewRecorder()
	httpReq := httptest.NewRequest("GET", "/", nil)
	httpReq.Header.Set("Accept-Language", "id")
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httpReq

	type req struct {
		Name string `validate:"required"`
	}

	err := validateStruct(ctx, &req{})
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != "validation failed: Name harus diisi" {
		t.Fatalf("expected translated error, got %v", err)
	}
}

func TestValidationMessages(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("GET", "/", nil)

	type allTags struct {
		Numeric  string `validate:"numeric"`
		Alpha    string `validate:"alpha"`
		Alphanum string `validate:"alphanum"`
		Max      string `validate:"max=3"`
		Len      string `validate:"len=3"`
	}

	err := validateStruct(ctx, &allTags{
		Numeric:  "abc",
		Alpha:    "123",
		Alphanum: "!@#",
		Max:      "abcd",
		Len:      "ab",
	})

	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestValidation_EdgeCases(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("GET", "/", nil)

	t.Run("Default_Message", func(t *testing.T) {
		type unknownTag struct {
			Name string `validate:"url"`
		}
		err := validateStruct(ctx, &unknownTag{Name: "not-a-url"})
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "Name failed validation for url") {
			t.Errorf("expected default message, got %v", err)
		}
	})

	t.Run("Param_Translation", func(t *testing.T) {
		RegisterTranslation("en", "min", "Field %s must be at least %s")
		type minTag struct {
			Age int `validate:"min=18"`
		}
		err := validateStruct(ctx, &minTag{Age: 10})
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "Field Age must be at least 18") {
			t.Errorf("expected param translation, got %v", err)
		}
	})
}
