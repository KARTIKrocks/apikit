package dbx

import (
	"database/sql"
	"reflect"
	"sync"
)

// fieldInfo describes a single struct field mapped to a database column.
type fieldInfo struct {
	index []int // reflect field index (supports embedded structs)
}

// typeMapping holds the column-name → field mapping for a struct type.
type typeMapping struct {
	fields map[string]fieldInfo // db tag value → field info
}

// cache stores computed type mappings keyed by reflect.Type.
var cache sync.Map // reflect.Type → *typeMapping

// getMapping returns the column→field mapping for T, using a cached value if available.
// Panics if T is not a struct (or pointer to struct).
func getMapping[T any]() *typeMapping {
	t := reflect.TypeFor[T]()

	// Dereference pointer type to get the struct type.
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		panic("dbx: type parameter T must be a struct, got " + t.Kind().String())
	}

	if cached, ok := cache.Load(t); ok {
		return cached.(*typeMapping)
	}

	m := buildMapping(t)
	actual, _ := cache.LoadOrStore(t, m)
	return actual.(*typeMapping)
}

// buildMapping constructs a typeMapping by inspecting struct tags.
func buildMapping(t reflect.Type) *typeMapping {
	m := &typeMapping{fields: make(map[string]fieldInfo)}
	collectFields(t, nil, m)
	return m
}

// collectFields recursively collects fields from t and embedded structs.
func collectFields(t reflect.Type, parentIndex []int, m *typeMapping) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		index := make([]int, len(parentIndex)+1)
		copy(index, parentIndex)
		index[len(parentIndex)] = i

		// Handle embedded structs: recurse into them.
		// Anonymous fields may have unexported type names (e.g., lowercase struct)
		// but still contain exported fields, so check Anonymous before IsExported.
		// However, embedded *pointer* types that are unexported cannot be initialized
		// via reflect at scan time, so we only recurse into pointer embeds when the
		// field is exported (i.e., the type name is capitalized).
		if f.Anonymous {
			ft := f.Type
			isPtr := ft.Kind() == reflect.Pointer
			if isPtr {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				// Skip unexported pointer embeds — reflect can't Set them.
				if isPtr && !f.IsExported() {
					continue
				}
				collectFields(ft, index, m)
				continue
			}
		}

		// Skip unexported non-anonymous fields.
		if !f.IsExported() {
			continue
		}

		tag := f.Tag.Get("db")
		if tag == "" || tag == "-" {
			continue
		}

		// First field wins for duplicate tags (consistent with encoding/json).
		if _, exists := m.fields[tag]; !exists {
			m.fields[tag] = fieldInfo{index: index}
		}
	}
}

// columnMapping holds pre-computed scan layout for a specific set of columns
// against a specific struct type. It is built once per query and reused across rows.
type columnMapping struct {
	// fieldIndexes[i] is the reflect field index for column i,
	// or nil if the column has no matching struct field.
	fieldIndexes [][]int
}

// buildColumnMapping pre-computes the column→field index mapping.
func buildColumnMapping(columns []string, m *typeMapping) *columnMapping {
	cm := &columnMapping{
		fieldIndexes: make([][]int, len(columns)),
	}
	for i, col := range columns {
		if fi, ok := m.fields[col]; ok {
			cm.fieldIndexes[i] = fi.index
		}
	}
	return cm
}

// scanInto builds scan destinations into targets for the given struct value.
// targets must be pre-allocated with len(columns). discard is a caller-owned
// scratch variable for unmatched columns (avoids shared mutable state).
// This avoids per-row allocation of the targets slice.
func scanInto(val reflect.Value, cm *columnMapping, targets []any, discard *sql.RawBytes) {
	for i, idx := range cm.fieldIndexes {
		if idx == nil {
			targets[i] = discard
		} else {
			if len(idx) > 1 {
				initEmbeddedPtrs(val, idx)
			}
			targets[i] = val.FieldByIndex(idx).Addr().Interface()
		}
	}
}

// initEmbeddedPtrs allocates nil embedded pointer fields along the
// index path so that FieldByIndex won't panic on nil dereference.
func initEmbeddedPtrs(val reflect.Value, index []int) {
	for i := 1; i < len(index); i++ {
		f := val.FieldByIndex(index[:i])
		if f.Kind() == reflect.Pointer && f.IsNil() && f.CanSet() {
			f.Set(reflect.New(f.Type().Elem()))
		}
	}
}
