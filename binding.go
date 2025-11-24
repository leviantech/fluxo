package fluxo

import (
	"encoding"
	"fmt"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func bindQuery(r *http.Request, target interface{}) error {
	return bindValues(r.URL.Query(), target, "query")
}

func bindPath(r *http.Request, target interface{}) error {
	params := make(map[string][]string)
	ginCtx, ok := r.Context().Value("gin").(*gin.Context)
	if ok {
		for _, p := range ginCtx.Params {
			params[p.Key] = []string{p.Value}
		}
	}
	return bindValues(params, target, "path")
}

func bindValues(values map[string][]string, target interface{}, tag string) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer")
	}

	targetElem := targetValue.Elem()
	if targetElem.Kind() != reflect.Struct {
		return fmt.Errorf("target must point to a struct")
	}

	targetType := targetElem.Type()

	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		fieldValue := targetElem.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		tagValue := field.Tag.Get(tag)
		if tagValue == "" {
			continue
		}

		tagParts := strings.Split(tagValue, ",")
		if len(tagParts) == 0 {
			continue
		}

		paramName := tagParts[0]
		if paramName == "" {
			continue
		}

		values, exists := values[paramName]
		if !exists || len(values) == 0 {
			continue
		}

		value := values[0]
		if err := setFieldValue(fieldValue, value); err != nil {
			return fmt.Errorf("failed to set field %s: %v", field.Name, err)
		}
	}

	return nil
}

func setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolVal)
	default:
		if textUnmarshaler, ok := field.Addr().Interface().(encoding.TextUnmarshaler); ok {
			return textUnmarshaler.UnmarshalText([]byte(value))
		}
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}
	return nil
}

func bindMultipartFiles(r *http.Request, target interface{}) error {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer")
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("target must point to a struct")
	}
	t := v.Type()
	fhPtr := reflect.TypeOf((*multipart.FileHeader)(nil))
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		name := f.Tag.Get("form")
		if name == "" || name == "-" {
			continue
		}
		if r.MultipartForm == nil || r.MultipartForm.File == nil {
			continue
		}
		files, ok := r.MultipartForm.File[name]
		if !ok || len(files) == 0 {
			continue
		}
		fv := v.Field(i)
		if !fv.CanSet() {
			continue
		}
		if fv.Type() == fhPtr {
			fv.Set(reflect.ValueOf(files[0]))
			continue
		}
		if fv.Kind() == reflect.Slice && fv.Type().Elem() == fhPtr {
			slice := reflect.MakeSlice(fv.Type(), len(files), len(files))
			for j := 0; j < len(files); j++ {
				slice.Index(j).Set(reflect.ValueOf(files[j]))
			}
			fv.Set(slice)
		}
	}
	return nil
}
