package fluxo

import "testing"

type vt struct {
	Email string `validate:"required,email"`
	Age   int    `validate:"min=18"`
}

func TestValidateStruct(t *testing.T) {

	err := validateStruct(&vt{Email: "bad", Age: 10})
	if err == nil {
		t.Fatalf("expected error")
	}

	if err2 := validateStruct(&vt{Email: "a@b.com", Age: 25}); err2 != nil {
		t.Fatalf("unexpected %v", err2)
	}
}
