// Copyright 2025 M Reyhan Fahlevi
// Licensed under the MIT License. See LICENSE for details.
package fluxo

import (
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin"
)

const (
	authenticatedUserKey = "authenticated_user"
)

type Context struct {
	*gin.Context
}

func (c *Context) SetAuthenticatedUser(user any) {
	c.Set(authenticatedUserKey, user)
}

func (c *Context) GetAuthenticatedUser(target any) error {
	v, exists := c.Get(authenticatedUserKey)
	if !exists {
		return fmt.Errorf("authenticated user not found")
	}

	targetVal := reflect.ValueOf(target)
	if targetVal.Kind() != reflect.Pointer {
		panic("target must be a pointer")
	}

	sourceVal := reflect.ValueOf(v)

	if sourceVal.Type().AssignableTo(targetVal.Elem().Type()) {
		targetVal.Elem().Set(sourceVal)
		return nil
	}

	return fmt.Errorf("authenticated user type mismatch")
}

func (c *Context) Lang() string {
	lang := c.GetHeader("Accept-Language")
	if lang == "" {
		return "en"
	}
	return lang
}
