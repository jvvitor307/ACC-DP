package avro

import (
	"fmt"
	"reflect"
	"strings"

	avrolib "github.com/hamba/avro/v2"

	"acc-dp/producer/internal/domain/event"
	"acc-dp/producer/internal/source/acc_shm"
)

type Serializer struct {
	physicsSchema  avrolib.Schema
	graphicsSchema avrolib.Schema
	staticSchema   avrolib.Schema
}

func NewSerializer() (*Serializer, error) {
	physicsSchema, err := parseEventSchema("acc_physics_event", "acc_physics_payload", reflect.TypeOf(acc_shm.PhysicsPage{}))
	if err != nil {
		return nil, fmt.Errorf("create serializer: %w", err)
	}

	graphicsSchema, err := parseEventSchema("acc_graphics_event", "acc_graphics_payload", reflect.TypeOf(acc_shm.GraphicsPage{}))
	if err != nil {
		return nil, fmt.Errorf("create serializer: %w", err)
	}

	staticSchema, err := parseEventSchema("acc_static_event", "acc_static_payload", reflect.TypeOf(acc_shm.StaticPage{}))
	if err != nil {
		return nil, fmt.Errorf("create serializer: %w", err)
	}

	return &Serializer{
		physicsSchema:  physicsSchema,
		graphicsSchema: graphicsSchema,
		staticSchema:   staticSchema,
	}, nil
}

func (s *Serializer) SerializePhysics(evt *event.PhysicsEvent) ([]byte, error) {
	if evt == nil {
		return nil, fmt.Errorf("serialize physics: event is nil")
	}

	datum, err := toDatum(evt)
	if err != nil {
		return nil, fmt.Errorf("serialize physics: %w", err)
	}

	payload, err := avrolib.Marshal(s.physicsSchema, datum)
	if err != nil {
		return nil, fmt.Errorf("serialize physics: %w", err)
	}

	return payload, nil
}

func (s *Serializer) SerializeGraphics(evt *event.GraphicsEvent) ([]byte, error) {
	if evt == nil {
		return nil, fmt.Errorf("serialize graphics: event is nil")
	}

	datum, err := toDatum(evt)
	if err != nil {
		return nil, fmt.Errorf("serialize graphics: %w", err)
	}

	payload, err := avrolib.Marshal(s.graphicsSchema, datum)
	if err != nil {
		return nil, fmt.Errorf("serialize graphics: %w", err)
	}

	return payload, nil
}

func (s *Serializer) SerializeStatic(evt *event.StaticEvent) ([]byte, error) {
	if evt == nil {
		return nil, fmt.Errorf("serialize static: event is nil")
	}

	datum, err := toDatum(evt)
	if err != nil {
		return nil, fmt.Errorf("serialize static: %w", err)
	}

	payload, err := avrolib.Marshal(s.staticSchema, datum)
	if err != nil {
		return nil, fmt.Errorf("serialize static: %w", err)
	}

	return payload, nil
}

func parseEventSchema(eventName, payloadName string, payloadType reflect.Type) (avrolib.Schema, error) {
	schemaText, err := buildEventSchema(eventName, payloadName, payloadType)
	if err != nil {
		return nil, fmt.Errorf("build schema %s: %w", eventName, err)
	}

	schema, err := avrolib.Parse(schemaText)
	if err != nil {
		return nil, fmt.Errorf("parse schema %s: %w", eventName, err)
	}

	return schema, nil
}

func toDatum(value any) (any, error) {
	if value == nil {
		return nil, fmt.Errorf("datum value is nil")
	}

	return reflectToDatum(reflect.ValueOf(value))
}

func reflectToDatum(value reflect.Value) (any, error) {
	if !value.IsValid() {
		return nil, fmt.Errorf("invalid datum value")
	}

	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil, fmt.Errorf("nil pointer datum")
		}
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.Struct:
		out := make(map[string]any)
		typ := value.Type()
		for i := range typ.NumField() {
			field := typ.Field(i)
			if !field.IsExported() {
				continue
			}

			fieldValue, err := reflectToDatum(value.Field(i))
			if err != nil {
				return nil, fmt.Errorf("field %s: %w", field.Name, err)
			}

			out[fieldName(field)] = fieldValue
		}

		return out, nil
	case reflect.Array, reflect.Slice:
		length := value.Len()
		out := make([]any, length)
		for i := range length {
			itemValue, err := reflectToDatum(value.Index(i))
			if err != nil {
				return nil, fmt.Errorf("index %d: %w", i, err)
			}
			out[i] = itemValue
		}

		return out, nil
	case reflect.Bool:
		return value.Bool(), nil
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int:
		return int32(value.Int()), nil
	case reflect.Int64:
		return value.Int(), nil
	case reflect.Uint8, reflect.Uint16:
		return int32(value.Uint()), nil
	case reflect.Uint32, reflect.Uint64, reflect.Uint:
		return int64(value.Uint()), nil
	case reflect.Float32:
		return float32(value.Float()), nil
	case reflect.Float64:
		return value.Float(), nil
	case reflect.String:
		return value.String(), nil
	default:
		return nil, fmt.Errorf("unsupported datum kind %s", value.Kind())
	}
}

func fieldName(field reflect.StructField) string {
	tag := field.Tag.Get("avro")
	if tag == "" {
		return field.Name
	}

	parts := strings.Split(tag, ",")
	if len(parts) == 0 || parts[0] == "" || parts[0] == "-" {
		return field.Name
	}

	return parts[0]
}
