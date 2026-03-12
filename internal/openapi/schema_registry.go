package openapi

import (
	"encoding/json"
	"reflect"
	"strings"
)

var schemaRegistry = map[string]Schema{}

func registerSchema(v any) string {

	t := reflect.TypeOf(v)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	name := t.Name()

	// already registered
	if _, exists := schemaRegistry[name]; exists {
		return name
	}

	schema := schemaFromStruct(t)

	schemaRegistry[name] = schema

	return name
}

func schemaFromStruct(t reflect.Type) Schema {

	props := map[string]Schema{}

	for i := 0; i < t.NumField(); i++ {

		field := t.Field(i)

		jsonTag := field.Tag.Get("json")
		name := strings.Split(jsonTag, ",")[0]

		if name == "" || name == "-" {
			continue
		}

		props[name] = goTypeToSchema(field.Type)
	}

	return Schema{
		Type:       "object",
		Properties: props,
	}
}

func goTypeToSchema(t reflect.Type) Schema {

	switch t.Kind() {

	case reflect.String:
		return Schema{Type: "string"}

	case reflect.Int, reflect.Int64:
		return Schema{Type: "integer"}

	case reflect.Float32, reflect.Float64:
		return Schema{Type: "number"}

	case reflect.Bool:
		return Schema{Type: "boolean"}

	case reflect.Struct:

		// Handle shopspring decimal as a number
		if t.String() == "decimal.Decimal" {
			return Schema{
				Type:   "string",
				Format: "decimal",
			}
		}

		if t.String() == "time.Time" {
			return Schema{
				Type:   "string",
				Format: "date-time",
			}
		}

		name := registerSchema(reflect.New(t).Interface())

		return Schema{
			Ref: "#/components/schemas/" + name,
		}

	case reflect.Slice:

		// json.RawMessage ([]byte) should be a free-form object, not an array
		if t == reflect.TypeFor[json.RawMessage]() {
			return Schema{Type: "object"}
		}

		item := goTypeToSchema(t.Elem())

		return Schema{
			Type:  "array",
			Items: &item,
		}
	}

	return Schema{Type: "string"}
}
