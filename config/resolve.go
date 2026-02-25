package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/KARTIKrocks/apikit/errors"
)

// resolve populates dst by merging sources in priority order:
// env vars > .env file values > json values > default tags.
// dst must be a non-nil pointer to a struct.
func resolve(dst any, prefix string, envFileValues, jsonValues map[string]string) error {
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return errors.BadRequest("config: dst must be a non-nil pointer to a struct")
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return errors.BadRequest("config: dst must be a pointer to a struct")
	}

	return resolveStruct(rv, prefix, "", "", envFileValues, jsonValues)
}

// resolveStruct recursively resolves struct fields.
// userPrefix is the caller-supplied prefix (e.g., "APP").
// structPrefix is the nesting prefix built from struct field names (e.g., "DB_").
// namePrefix is the struct nesting prefix for error messages.
func resolveStruct(rv reflect.Value, userPrefix, structPrefix, namePrefix string, envFileValues, jsonValues map[string]string) error {
	rt := rv.Type()

	for i := range rt.NumField() {
		field := rt.Field(i)
		fieldVal := rv.Field(i)

		if !field.IsExported() {
			continue
		}

		// Handle nested structs (no env tag).
		envTag := field.Tag.Get("env")
		if fieldVal.Kind() == reflect.Struct && envTag == "" && field.Type != reflect.TypeFor[time.Duration]() {
			// Extend structPrefix with the field name.
			nestedStructPrefix := structPrefix + strings.ToUpper(field.Name) + "_"

			nestedNamePrefix := namePrefix
			if nestedNamePrefix != "" {
				nestedNamePrefix += "." + field.Name
			} else {
				nestedNamePrefix = field.Name
			}

			if err := resolveStruct(fieldVal, userPrefix, nestedStructPrefix, nestedNamePrefix, envFileValues, jsonValues); err != nil {
				return err
			}
			continue
		}

		if envTag == "" {
			continue
		}

		// Determine the display name for error messages.
		displayName := field.Name
		if namePrefix != "" {
			displayName = namePrefix + "." + field.Name
		}

		// Look up value from sources in priority order.
		value, found := lookupValue(envTag, userPrefix, structPrefix, envFileValues, jsonValues, field)
		if !found {
			continue
		}

		if err := setField(fieldVal, field.Type, value, displayName); err != nil {
			return err
		}
	}

	return nil
}

// buildKey concatenates non-empty parts with "_".
func buildKey(parts ...string) string {
	var nonEmpty []string
	for _, p := range parts {
		if p != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	return strings.Join(nonEmpty, "_")
}

// lookupValue retrieves a configuration value from sources in priority order:
// 1. Environment variable  (key = userPrefix + structPrefix + envTag)
// 2. .env file value       (key = userPrefix + structPrefix + envTag, skipped if real env has it)
// 3. JSON values           (key = structPrefix + envTag — no user prefix)
// 4. Default tag
func lookupValue(envTag, userPrefix, structPrefix string, envFileValues, jsonValues map[string]string, field reflect.StructField) (string, bool) {
	// structPrefix ends with "_" (e.g., "DB_"); trim for key building.
	sp := strings.TrimSuffix(structPrefix, "_")

	// Env prefix combines userPrefix and struct nesting: "APP" + "DB" → "APP_DB".
	envPrefix := buildKey(userPrefix, sp)
	// JSON key uses only struct nesting: "DB" + "HOST" → "DB_HOST".
	jsonKey := buildKey(sp, envTag)

	// 1. Environment variable (highest priority).
	if val, ok := lookupEnv(envTag, envPrefix); ok {
		return val, true
	}

	// 2. .env file values.
	if val, ok := lookupEnvFile(envTag, envPrefix, envFileValues); ok {
		return val, true
	}

	// 3. JSON values.
	if jsonValues != nil {
		if val, ok := jsonValues[jsonKey]; ok {
			return val, true
		}
	}

	// 4. Default tag (lowest priority).
	if def, ok := field.Tag.Lookup("default"); ok {
		return def, true
	}

	return "", false
}

// setField parses a string value and sets it on a reflect.Value.
func setField(fieldVal reflect.Value, fieldType reflect.Type, value, displayName string) error {
	// Handle slice types.
	if fieldType.Kind() == reflect.Slice {
		return setSliceField(fieldVal, fieldType, value, displayName)
	}

	// Handle time.Duration.
	if fieldType == reflect.TypeFor[time.Duration]() {
		d, err := time.ParseDuration(value)
		if err != nil {
			return errors.BadRequest(fmt.Sprintf("config: field %s: cannot parse %q as duration", displayName, value))
		}
		fieldVal.Set(reflect.ValueOf(d))
		return nil
	}

	switch fieldType.Kind() {
	case reflect.String:
		fieldVal.SetString(value)

	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return errors.BadRequest(fmt.Sprintf("config: field %s: cannot parse %q as bool", displayName, value))
		}
		fieldVal.SetBool(b)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(value, 10, fieldType.Bits())
		if err != nil {
			return errors.BadRequest(fmt.Sprintf("config: field %s: cannot parse %q as int", displayName, value))
		}
		fieldVal.SetInt(n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(value, 10, fieldType.Bits())
		if err != nil {
			return errors.BadRequest(fmt.Sprintf("config: field %s: cannot parse %q as uint", displayName, value))
		}
		fieldVal.SetUint(n)

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, fieldType.Bits())
		if err != nil {
			return errors.BadRequest(fmt.Sprintf("config: field %s: cannot parse %q as float", displayName, value))
		}
		fieldVal.SetFloat(f)

	default:
		return errors.BadRequest(fmt.Sprintf("config: field %s: unsupported type %s", displayName, fieldType.Kind()))
	}

	return nil
}

// setSliceField parses a comma-separated string into a slice.
func setSliceField(fieldVal reflect.Value, fieldType reflect.Type, value, displayName string) error {
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	elemType := fieldType.Elem()

	slice := reflect.MakeSlice(fieldType, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)

		switch elemType.Kind() {
		case reflect.String:
			slice = reflect.Append(slice, reflect.ValueOf(part))

		case reflect.Int:
			n, err := strconv.Atoi(part)
			if err != nil {
				return errors.BadRequest(fmt.Sprintf("config: field %s: cannot parse %q as int in list", displayName, part))
			}
			slice = reflect.Append(slice, reflect.ValueOf(n))

		default:
			return errors.BadRequest(fmt.Sprintf("config: field %s: unsupported slice element type %s", displayName, elemType.Kind()))
		}
	}

	fieldVal.Set(slice)
	return nil
}
