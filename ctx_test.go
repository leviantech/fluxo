package fluxo

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type user struct {
	ID string `json:"id"`
}

func assertPanic(t *testing.T, f func(), expected string) {
	t.Helper()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic but got none")
		}
		if r != expected {
			t.Fatalf("expected panic %v but got %v", expected, r)
		}
	}()

	f()
}

func TestAuthenticateUser(t *testing.T) {
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)
	ctx := Context{ginCtx}
	ctx.SetAuthenticatedUser(user{ID: "123"})

	parsedUser := user{}
	err := ctx.GetAuthenticatedUser(&parsedUser)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if parsedUser.ID != "123" {
		t.Errorf("expected 123, got %s", parsedUser.ID)
	}

	invalidType := struct {
		ID int `json:"id"`
	}{}

	err = ctx.GetAuthenticatedUser(&invalidType)
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	assertPanic(t, func() { ctx.GetAuthenticatedUser(invalidType) }, "target must be a pointer")
}
