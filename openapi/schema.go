package openapi

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// MarshalJSON folds Nullable into the "type" keyword as ["<type>", "null"],
// per JSON Schema 2020-12 (OpenAPI 3.1). A nullable $ref can't express that
// the same way — a sibling "type" next to "$ref" isn't reliably honored by
// tooling — so it's wrapped as {"anyOf": [{"$ref": ...}, {"type": "null"}]}
// instead. Everything else marshals like a plain struct.
func (s *Schema) MarshalJSON() ([]byte, error) {
	type alias Schema // avoid recursing into MarshalJSON
	switch {
	case s == nil:
		return []byte("null"), nil
	case !s.Nullable:
		return json.Marshal((*alias)(s))
	case s.Ref != "":
		// registry.schemaFor never sets other fields alongside Ref, so a
		// bare {$ref} is all there is to preserve here.
		return json.Marshal(&Schema{AnyOf: []*Schema{{Ref: s.Ref}, {Type: "null"}}})
	case s.Type == "":
		return json.Marshal((*alias)(s)) // unconstrained schema already matches null
	default:
		out := struct {
			*alias
			Type []string `json:"type"`
		}{alias: (*alias)(s), Type: []string{s.Type, "null"}}
		return json.Marshal(out)
	}
}

// registry accumulates named component schemas as they're discovered, keyed
// by the name used in "#/components/schemas/<name>".
type registry struct {
	schemas map[string]*Schema      // name -> schema
	types   map[reflect.Type]string // type -> assigned name (dedup + collision tracking)
	byName  map[string]reflect.Type // name -> owning type (collision detection)
}

func newRegistry() *registry {
	return &registry{
		schemas: make(map[string]*Schema),
		types:   make(map[reflect.Type]string),
		byName:  make(map[string]reflect.Type),
	}
}

// schemaFor returns the schema for t, registering it as a named component if
// t is a defined (non-anonymous) struct type, or a ref to that component if
// it was already registered.
func (reg *registry) schemaFor(t reflect.Type) *Schema {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Struct:
		if t.Name() == "" {
			// Anonymous struct: inline it, no component.
			return reg.structSchema(t)
		}
		if t == reflectTimeType {
			return &Schema{Type: "string", Format: "date-time"}
		}
		if name, ok := reg.types[t]; ok {
			return &Schema{Ref: "#/components/schemas/" + name}
		}
		name := reg.reserveName(t)
		// Reserve the name before recursing so self-referential/cyclic
		// struct graphs terminate instead of looping forever.
		reg.types[t] = name
		reg.schemas[name] = reg.structSchema(t)
		return &Schema{Ref: "#/components/schemas/" + name}

	case reflect.Slice, reflect.Array:
		if t.Elem().Kind() == reflect.Uint8 {
			return &Schema{Type: "string", Format: "byte"} // []byte -> base64
		}
		return &Schema{Type: "array", Items: reg.schemaFor(t.Elem())}

	case reflect.Map:
		return &Schema{Type: "object", AdditionalProperties: reg.schemaFor(t.Elem())}

	default:
		return primitiveSchema(t)
	}
}

// reserveName returns a name unique within the registry, qualifying with the
// package path on collision (two distinct types sharing a bare type name).
func (reg *registry) reserveName(t reflect.Type) string {
	name := t.Name()
	if owner, exists := reg.byName[name]; exists && owner != t {
		name = strings.ReplaceAll(t.PkgPath(), "/", "_") + "_" + t.Name()
	}
	reg.byName[name] = t
	return name
}

// structSchema builds an inline object schema for a struct type, flattening
// anonymous (embedded) fields the way encoding/json promotes them.
func (reg *registry) structSchema(t reflect.Type) *Schema {
	s := &Schema{Type: "object", Properties: map[string]*Schema{}}

	for i := range t.NumField() {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		if field.Anonymous && jsonTag == "" {
			embedded := field.Type
			for embedded.Kind() == reflect.Pointer {
				embedded = embedded.Elem()
			}
			if embedded.Kind() == reflect.Struct && embedded != reflectTimeType {
				inner := reg.structSchema(embedded)
				for name, propSchema := range inner.Properties {
					s.Properties[name] = propSchema
				}
				s.Required = append(s.Required, inner.Required...)
				continue
			}
		}

		name, omitEmpty := jsonFieldInfo(field)
		fieldSchema := reg.schemaFor(field.Type)
		if field.Type.Kind() == reflect.Pointer {
			fieldSchema.Nullable = true
		}

		required := applyValidateTag(fieldSchema, field.Type, field.Tag.Get("validate"))
		s.Properties[name] = fieldSchema
		if required && !omitEmpty {
			s.Required = append(s.Required, name)
		}
	}

	return s
}

// jsonFieldInfo returns the wire name and whether "omitempty" is set, falling
// back to the Go field name like request/tag_validator.go's jsonFieldName.
func jsonFieldInfo(field reflect.StructField) (name string, omitEmpty bool) {
	tag := field.Tag.Get("json")
	if tag == "" {
		return field.Name, false
	}
	parts := strings.Split(tag, ",")
	name = parts[0]
	if name == "" {
		name = field.Name
	}
	for _, opt := range parts[1:] {
		if opt == "omitempty" {
			omitEmpty = true
		}
	}
	return name, omitEmpty
}

var reflectTimeType = reflect.TypeFor[time.Time]()

func primitiveSchema(t reflect.Type) *Schema {
	switch t.Kind() {
	case reflect.String:
		return &Schema{Type: "string"}
	case reflect.Bool:
		return &Schema{Type: "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &Schema{Type: "integer"}
	case reflect.Float32, reflect.Float64:
		return &Schema{Type: "number"}
	case reflect.Interface:
		return &Schema{} // any -> unconstrained
	default:
		return &Schema{Type: "string"}
	}
}

// applyValidateTag mutates s according to the comma-separated `validate` tag
// rules, mirroring request/tag_validator.go's rule set and its "min/max mean
// length for string/slice/map, value for numbers" duality. It returns whether
// "required" was present.
func applyValidateTag(s *Schema, fieldType reflect.Type, tag string) (required bool) {
	if tag == "" || tag == "-" {
		return false
	}

	isLength, isNumeric := lengthAndNumericKind(fieldType)

	for _, rule := range strings.Split(tag, ",") {
		rule = strings.TrimSpace(rule)
		name, arg, _ := strings.Cut(rule, "=")

		switch name {
		case "required":
			required = true
		case "omitempty", "":
			// no schema effect
		default:
			if fn, ok := schemaRuleFuncs[name]; ok {
				fn(s, arg, isLength, isNumeric)
			}
		}
	}

	return required
}

// lengthAndNumericKind reports which half of the "min/max mean length for
// string/slice/map, value for numbers" duality a field's (possibly
// pointer-wrapped) kind falls into.
func lengthAndNumericKind(fieldType reflect.Type) (isLength, isNumeric bool) {
	kind := fieldType.Kind()
	for kind == reflect.Pointer {
		kind = fieldType.Elem().Kind()
	}
	switch kind {
	case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
		return true, false
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return false, true
	default:
		return false, false
	}
}

// schemaRule applies a single validate rule to s, given param (the text after
// '=', empty for param-less rules) and the field's length/numeric kind.
type schemaRule func(s *Schema, param string, isLength, isNumeric bool)

// schemaRuleFuncs is the dispatch table mapping a `validate` rule name to its
// schema effect. Rules with no schema representation (e.g. "ne") are simply
// absent — an unrecognized rule is silently ignored here, unlike
// request.ValidateStruct's tag engine, since schema generation is best-effort
// documentation rather than enforcement.
var schemaRuleFuncs = map[string]schemaRule{
	"email":      func(s *Schema, _ string, _, _ bool) { s.Format = "email" },
	"url":        func(s *Schema, _ string, _, _ bool) { s.Format = "uri" },
	"uuid":       func(s *Schema, _ string, _, _ bool) { s.Format = "uuid" },
	"e164":       func(s *Schema, _ string, _, _ bool) { s.Pattern = `^\+[1-9]\d{1,14}$` },
	"alpha":      func(s *Schema, _ string, _, _ bool) { s.Pattern = `^[A-Za-z]+$` },
	"alphanum":   func(s *Schema, _ string, _, _ bool) { s.Pattern = `^[A-Za-z0-9]+$` },
	"numeric":    func(s *Schema, _ string, _, _ bool) { s.Pattern = `^[0-9]+$` },
	"contains":   func(s *Schema, arg string, _, _ bool) { s.Pattern = ".*" + regexp.QuoteMeta(arg) + ".*" },
	"startswith": func(s *Schema, arg string, _, _ bool) { s.Pattern = "^" + regexp.QuoteMeta(arg) },
	"endswith":   func(s *Schema, arg string, _, _ bool) { s.Pattern = regexp.QuoteMeta(arg) + "$" },
	"oneof":      func(s *Schema, arg string, _, _ bool) { s.Enum = strings.Fields(arg) },
	"eq":         func(s *Schema, arg string, _, _ bool) { s.Enum = []string{arg} },
	"len": func(s *Schema, arg string, isLength, _ bool) {
		if n, err := strconv.Atoi(arg); isLength && err == nil {
			s.MinLength, s.MaxLength = &n, &n
		}
	},
	"min": func(s *Schema, arg string, isLength, isNumeric bool) {
		applyBound(s, arg, isLength, isNumeric, boundMin, false)
	},
	"gte": func(s *Schema, arg string, isLength, isNumeric bool) {
		applyBound(s, arg, isLength, isNumeric, boundMin, false)
	},
	"max": func(s *Schema, arg string, isLength, isNumeric bool) {
		applyBound(s, arg, isLength, isNumeric, boundMax, false)
	},
	"lte": func(s *Schema, arg string, isLength, isNumeric bool) {
		applyBound(s, arg, isLength, isNumeric, boundMax, false)
	},
	"gt": func(s *Schema, arg string, isLength, isNumeric bool) {
		applyBound(s, arg, isLength, isNumeric, boundMin, true)
	},
	"lt": func(s *Schema, arg string, isLength, isNumeric bool) {
		applyBound(s, arg, isLength, isNumeric, boundMax, true)
	},
}

type boundKind int

const (
	boundMin boundKind = iota
	boundMax
)

func applyBound(s *Schema, arg string, isLength, isNumeric bool, kind boundKind, exclusive bool) {
	switch {
	case isLength:
		n, err := strconv.Atoi(arg)
		if err != nil {
			return
		}
		if kind == boundMin {
			s.MinLength = &n
		} else {
			s.MaxLength = &n
		}
	case isNumeric:
		f, err := strconv.ParseFloat(arg, 64)
		if err != nil {
			return
		}
		switch {
		case kind == boundMin && exclusive:
			s.ExclusiveMinimum = &f
		case kind == boundMin:
			s.Minimum = &f
		case kind == boundMax && exclusive:
			s.ExclusiveMaximum = &f
		default:
			s.Maximum = &f
		}
	}
}
