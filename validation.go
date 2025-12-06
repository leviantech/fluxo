// Copyright 2025 M Reyhan
// Licensed under the MIT License. See LICENSE for details.

package fluxo

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var (
	validate            = validator.New()
	translationRegistry = map[string]map[string]string{}
	mu                  sync.RWMutex
)

// RegisterTranslation registers a translated message for a validation tag.
// Example: fluxo.RegisterTranslation("jp", "required", "%s は必須です")
func RegisterTranslation(lang, tag, message string) {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := translationRegistry[lang]; !ok {
		translationRegistry[lang] = map[string]string{}
	}

	translationRegistry[lang][tag] = message
}

// translate returns a translated message if found.
func translate(lang, tag string, args ...any) (string, bool) {
	mu.RLock()
	defer mu.RUnlock()

	if tbl, ok := translationRegistry[lang]; ok {
		if msg, ok := tbl[tag]; ok {
			return fmt.Sprintf(msg, args...), true
		}
	}

	return "", false
}

// defaultValidationMessage replicates your original English fallback text.
func defaultValidationMessage(e validator.FieldError) string {
	field := e.Field()
	tag := e.Tag()
	param := e.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, param)
	case "numeric":
		return fmt.Sprintf("%s must be numeric", field)
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", field)
	default:
		return fmt.Sprintf("%s failed validation for %s", field, tag)
	}
}

// formatValidationError uses translation if available, fallback otherwise.
func formatValidationError(e validator.FieldError, lang string) string {
	field := e.Field()
	tag := e.Tag()
	param := e.Param()

	if param != "" {
		if msg, ok := translate(lang, tag, field, param); ok {
			return msg
		}
	} else {
		if msg, ok := translate(lang, tag, field); ok {
			return msg
		}
	}

	return defaultValidationMessage(e)
}

// validateStruct validates a struct using ctx to determine language.
func validateStruct(ctx *gin.Context, s interface{}) error {
	lang := ctx.GetHeader("Accept-Language")
	if lang == "" {
		lang = "en"
	}

	if err := validate.Struct(s); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return fmt.Errorf("validation failed: %v", err)
		}

		var messages []string
		for _, e := range validationErrors {
			messages = append(messages, formatValidationError(e, lang))
		}

		return fmt.Errorf("validation failed: %s", strings.Join(messages, "; "))
	}

	return nil
}
