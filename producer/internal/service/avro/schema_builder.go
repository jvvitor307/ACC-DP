package avro

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func buildEventSchema(eventName, payloadName string, payloadType reflect.Type) (string, error) {
	payloadSchema, err := buildRecordSchema(payloadName, payloadType)
	if err != nil {
		return "", fmt.Errorf("build payload schema: %w", err)
	}

	eventSchema := map[string]any{
		"type": "record",
		"name": eventName,
		"fields": []any{
			map[string]any{"name": "event_id", "type": "string"},
			map[string]any{"name": "event_time", "type": "long"},
			map[string]any{"name": "ingestion_time", "type": "long"},
			map[string]any{"name": "usuario_id", "type": "string"},
			map[string]any{"name": "username", "type": "string"},
			map[string]any{"name": "source", "type": "string"},
			map[string]any{"name": "payload", "type": payloadSchema},
		},
	}

	encoded, err := json.Marshal(eventSchema)
	if err != nil {
		return "", fmt.Errorf("marshal event schema: %w", err)
	}

	return string(encoded), nil
}

func buildRecordSchema(name string, typ reflect.Type) (map[string]any, error) {
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("record %s must be struct, got %s", name, typ.Kind())
	}

	fields := make([]any, 0, typ.NumField())
	for i := range typ.NumField() {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldType, err := mapType(field.Type)
		if err != nil {
			return nil, fmt.Errorf("map field %s.%s: %w", typ.Name(), field.Name, err)
		}

		fields = append(fields, map[string]any{
			"name": field.Name,
			"type": fieldType,
		})
	}

	return map[string]any{
		"type":   "record",
		"name":   name,
		"fields": fields,
	}, nil
}

func mapType(typ reflect.Type) (any, error) {
	switch typ.Kind() {
	case reflect.Bool:
		return "boolean", nil
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Uint8, reflect.Uint16:
		return "int", nil
	case reflect.Int64, reflect.Uint32, reflect.Uint64, reflect.Uint:
		return "long", nil
	case reflect.Float32:
		return "float", nil
	case reflect.Float64:
		return "double", nil
	case reflect.String:
		return "string", nil
	case reflect.Array, reflect.Slice:
		itemType, err := mapType(typ.Elem())
		if err != nil {
			return nil, err
		}

		return map[string]any{
			"type":  "array",
			"items": itemType,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported kind %s", typ.Kind())
	}
}
