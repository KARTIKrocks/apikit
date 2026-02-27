package dbx

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"testing"

	"github.com/KARTIKrocks/apikit/errors"
	"github.com/KARTIKrocks/apikit/sqlbuilder"
)

// --- Mapping / scan.go tests ---

type basicStruct struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type withPointer struct {
	ID    int     `db:"id"`
	Email *string `db:"email"`
}

type withSkip struct {
	ID      int    `db:"id"`
	Skipped string `db:"-"`
	NoTag   string
}

type withEmbedded struct {
	basicStruct
	Extra string `db:"extra"`
}

type innerEmbed struct {
	Val int `db:"val"`
}

// withEmbeddedPtr uses an unexported pointer embed — fields inside
// are unreachable via reflect.Set, so they should be skipped.
type withEmbeddedPtr struct {
	*innerEmbed
	Name string `db:"name"`
}

// InnerExported is an exported embedded type for pointer embed tests.
type InnerExported struct {
	Val int `db:"val"`
}

// withExportedEmbeddedPtr uses an exported pointer embed — scannable.
type withExportedEmbeddedPtr struct {
	*InnerExported
	Name string `db:"name"`
}

type withUnexported struct {
	ID   int    `db:"id"`
	name string `db:"name"` //nolint:unused // intentionally unexported for test
}

type withDuplicateTag struct {
	First  string `db:"dup"`
	Second string `db:"dup"`
}

func TestGetMapping_BasicStruct(t *testing.T) {
	m := getMapping[basicStruct]()

	if len(m.fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(m.fields))
	}
	if _, ok := m.fields["id"]; !ok {
		t.Fatal("missing field mapping for 'id'")
	}
	if _, ok := m.fields["name"]; !ok {
		t.Fatal("missing field mapping for 'name'")
	}
}

func TestGetMapping_PointerFields(t *testing.T) {
	m := getMapping[withPointer]()

	if _, ok := m.fields["email"]; !ok {
		t.Fatal("missing field mapping for 'email'")
	}
}

func TestGetMapping_SkippedFields(t *testing.T) {
	m := getMapping[withSkip]()

	if len(m.fields) != 1 {
		t.Fatalf("expected 1 field, got %d: %v", len(m.fields), m.fields)
	}
	if _, ok := m.fields["id"]; !ok {
		t.Fatal("missing field mapping for 'id'")
	}
	if _, ok := m.fields["-"]; ok {
		t.Fatal("db:\"-\" field should not be mapped")
	}
}

func TestGetMapping_EmbeddedStruct(t *testing.T) {
	m := getMapping[withEmbedded]()

	if len(m.fields) != 3 {
		t.Fatalf("expected 3 fields, got %d: %v", len(m.fields), m.fields)
	}
	for _, name := range []string{"id", "name", "extra"} {
		if _, ok := m.fields[name]; !ok {
			t.Errorf("missing field mapping for %q", name)
		}
	}
}

func TestGetMapping_EmbeddedPointerStruct_Exported(t *testing.T) {
	m := getMapping[withExportedEmbeddedPtr]()

	if len(m.fields) != 2 {
		t.Fatalf("expected 2 fields, got %d: %v", len(m.fields), m.fields)
	}
	if _, ok := m.fields["val"]; !ok {
		t.Error("missing field mapping for 'val'")
	}
	if _, ok := m.fields["name"]; !ok {
		t.Error("missing field mapping for 'name'")
	}
}

func TestGetMapping_EmbeddedPointerStruct_UnexportedSkipped(t *testing.T) {
	m := getMapping[withEmbeddedPtr]()

	// Unexported pointer embed's fields should be skipped.
	if len(m.fields) != 1 {
		t.Fatalf("expected 1 field, got %d: %v", len(m.fields), m.fields)
	}
	if _, ok := m.fields["name"]; !ok {
		t.Fatal("missing field mapping for 'name'")
	}
	if _, ok := m.fields["val"]; ok {
		t.Fatal("unexported pointer embed field 'val' should not be mapped")
	}
}

func TestGetMapping_UnexportedSkipped(t *testing.T) {
	m := getMapping[withUnexported]()

	if len(m.fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(m.fields))
	}
	if _, ok := m.fields["id"]; !ok {
		t.Fatal("missing field mapping for 'id'")
	}
}

func TestGetMapping_DuplicateTag_FirstWins(t *testing.T) {
	m := getMapping[withDuplicateTag]()

	if len(m.fields) != 1 {
		t.Fatalf("expected 1 field for duplicate tag, got %d", len(m.fields))
	}
	fi := m.fields["dup"]
	// First field (index 0) should win.
	if len(fi.index) != 1 || fi.index[0] != 0 {
		t.Errorf("expected first field to win, got index %v", fi.index)
	}
}

func TestGetMapping_NonStructPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for non-struct T, got none")
		}
		msg, ok := r.(string)
		if !ok || msg != "dbx: type parameter T must be a struct, got int" {
			t.Errorf("unexpected panic message: %v", r)
		}
	}()
	getMapping[int]()
}

func TestGetMapping_CacheConsistency(t *testing.T) {
	m1 := getMapping[basicStruct]()
	m2 := getMapping[basicStruct]()

	if m1 != m2 {
		t.Error("expected same pointer from cache, got different mappings")
	}
}

func TestBuildColumnMapping(t *testing.T) {
	m := getMapping[basicStruct]()

	columns := []string{"name", "unknown", "id"}
	cm := buildColumnMapping(columns, m)

	if len(cm.fieldIndexes) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(cm.fieldIndexes))
	}
	if cm.fieldIndexes[0] == nil {
		t.Error("expected non-nil index for 'name'")
	}
	if cm.fieldIndexes[1] != nil {
		t.Error("expected nil index for 'unknown'")
	}
	if cm.fieldIndexes[2] == nil {
		t.Error("expected non-nil index for 'id'")
	}
}

func TestScanInto_MatchedColumns(t *testing.T) {
	m := getMapping[basicStruct]()
	var item basicStruct
	val := reflect.ValueOf(&item).Elem()

	columns := []string{"id", "name"}
	cm := buildColumnMapping(columns, m)
	targets := make([]any, len(columns))
	var discard sql.RawBytes
	scanInto(val, cm, targets, &discard)

	// Verify they point into the struct.
	idPtr, ok := targets[0].(*int)
	if !ok {
		t.Fatalf("expected *int for 'id' target, got %T", targets[0])
	}
	namePtr, ok := targets[1].(*string)
	if !ok {
		t.Fatalf("expected *string for 'name' target, got %T", targets[1])
	}

	// Write through pointers and verify struct is updated.
	*idPtr = 42
	*namePtr = "alice"
	if item.ID != 42 || item.Name != "alice" {
		t.Errorf("scan targets don't write into struct: got %+v", item)
	}
}

func TestScanInto_UnmatchedColumns(t *testing.T) {
	m := getMapping[basicStruct]()
	var item basicStruct
	val := reflect.ValueOf(&item).Elem()

	columns := []string{"id", "unknown_col", "name"}
	cm := buildColumnMapping(columns, m)
	targets := make([]any, len(columns))
	var discard sql.RawBytes
	scanInto(val, cm, targets, &discard)

	// The middle target should be the shared discard pointer.
	if _, ok := targets[1].(*sql.RawBytes); !ok {
		t.Errorf("expected *sql.RawBytes for unmatched column, got %T", targets[1])
	}
}

func TestScanInto_ColumnOrderIndependent(t *testing.T) {
	m := getMapping[basicStruct]()
	var item basicStruct
	val := reflect.ValueOf(&item).Elem()

	// Reversed column order compared to struct field order.
	columns := []string{"name", "id"}
	cm := buildColumnMapping(columns, m)
	targets := make([]any, len(columns))
	var discard sql.RawBytes
	scanInto(val, cm, targets, &discard)

	namePtr, ok := targets[0].(*string)
	if !ok {
		t.Fatalf("expected *string for 'name', got %T", targets[0])
	}
	idPtr, ok := targets[1].(*int)
	if !ok {
		t.Fatalf("expected *int for 'id', got %T", targets[1])
	}

	*namePtr = "bob"
	*idPtr = 7
	if item.Name != "bob" || item.ID != 7 {
		t.Errorf("column-order-independent scan failed: got %+v", item)
	}
}

func TestScanInto_ReusesTargetSlice(t *testing.T) {
	m := getMapping[basicStruct]()
	columns := []string{"id", "name"}
	cm := buildColumnMapping(columns, m)
	targets := make([]any, len(columns))
	var discard sql.RawBytes

	// Simulate two rows reusing the same targets slice.
	var item1 basicStruct
	scanInto(reflect.ValueOf(&item1).Elem(), cm, targets, &discard)
	*(targets[0].(*int)) = 1
	*(targets[1].(*string)) = "first"

	var item2 basicStruct
	scanInto(reflect.ValueOf(&item2).Elem(), cm, targets, &discard)
	*(targets[0].(*int)) = 2
	*(targets[1].(*string)) = "second"

	// Each struct should have its own values.
	if item1.ID != 1 || item1.Name != "first" {
		t.Errorf("item1 corrupted: %+v", item1)
	}
	if item2.ID != 2 || item2.Name != "second" {
		t.Errorf("item2 corrupted: %+v", item2)
	}
}

func TestScanInto_EmbeddedPointerInitialized(t *testing.T) {
	m := getMapping[withExportedEmbeddedPtr]()
	var item withExportedEmbeddedPtr // item.InnerExported is nil
	val := reflect.ValueOf(&item).Elem()

	columns := []string{"val", "name"}
	cm := buildColumnMapping(columns, m)
	targets := make([]any, len(columns))
	var discard sql.RawBytes
	scanInto(val, cm, targets, &discard)

	// The embedded pointer should have been initialized.
	if item.InnerExported == nil {
		t.Fatal("expected embedded pointer to be initialized, got nil")
	}

	// Verify we can write through the target.
	*(targets[0].(*int)) = 99
	*(targets[1].(*string)) = "test"
	if item.Val != 99 {
		t.Errorf("expected Val=99, got %d", item.Val)
	}
	if item.Name != "test" {
		t.Errorf("expected Name='test', got %q", item.Name)
	}
}

// --- Mock DB for tests ---

type mockDB struct {
	queryErr error
	execRes  sql.Result
	execErr  error
}

func (m *mockDB) QueryContext(_ context.Context, _ string, _ ...any) (*sql.Rows, error) {
	return nil, m.queryErr
}

func (m *mockDB) ExecContext(_ context.Context, _ string, _ ...any) (sql.Result, error) {
	if m.execErr != nil {
		return nil, m.execErr
	}
	return m.execRes, nil
}

type mockExecResult struct {
	lastID   int64
	affected int64
}

func (r mockExecResult) LastInsertId() (int64, error) { return r.lastID, nil }
func (r mockExecResult) RowsAffected() (int64, error) { return r.affected, nil }

// setMockDefault sets a mock as the default and returns a cleanup function.
func setMockDefault(m *mockDB) func() {
	old := defaultDB
	SetDefault(m)
	return func() { defaultDB = old }
}

// --- conn.go tests ---

func TestConn_NoDefault(t *testing.T) {
	old := defaultDB
	defaultDB = nil
	defer func() { defaultDB = old }()

	_, err := Exec(context.Background(), "SELECT 1")
	if err == nil {
		t.Fatal("expected error when no default is set")
	}
	if code := errors.ErrorCode(err); code != errors.CodeInternal {
		t.Errorf("expected code %s, got %s", errors.CodeInternal, code)
	}
}

func TestConn_WithTxSetsContext(t *testing.T) {
	ctx := context.Background()

	// Before WithTx, context has no tx — conn() should fall back to default.
	cleanup := setMockDefault(&mockDB{execRes: mockExecResult{}})
	defer cleanup()

	db, err := conn(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := db.(*mockDB); !ok {
		t.Fatal("expected default mockDB before WithTx")
	}

	// We can't create a real *sql.Tx without a driver, so verify that
	// WithTx stores the value and conn() retrieves it via a nil-tx check:
	// a nil *sql.Tx should be ignored (fall back to default).
	ctx = WithTx(ctx, (*sql.Tx)(nil))
	db, err = conn(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := db.(*mockDB); !ok {
		t.Fatal("nil *sql.Tx should be ignored, expected default mockDB")
	}
}

func TestConn_DefaultIsUsed(t *testing.T) {
	cleanup := setMockDefault(&mockDB{execRes: mockExecResult{affected: 7}})
	defer cleanup()

	result, err := Exec(context.Background(), "DELETE FROM sessions")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	affected, _ := result.RowsAffected()
	if affected != 7 {
		t.Errorf("expected 7 rows affected, got %d", affected)
	}
}

// --- dbx.go tests ---

func TestQueryAll_QueryError(t *testing.T) {
	cleanup := setMockDefault(&mockDB{queryErr: fmt.Errorf("connection refused")})
	defer cleanup()

	_, err := QueryAll[basicStruct](context.Background(), "SELECT 1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code := errors.ErrorCode(err); code != errors.CodeDatabaseError {
		t.Errorf("expected code %s, got %s", errors.CodeDatabaseError, code)
	}
}

func TestQueryOne_QueryError(t *testing.T) {
	cleanup := setMockDefault(&mockDB{queryErr: fmt.Errorf("connection refused")})
	defer cleanup()

	_, err := QueryOne[basicStruct](context.Background(), "SELECT 1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code := errors.ErrorCode(err); code != errors.CodeDatabaseError {
		t.Errorf("expected code %s, got %s", errors.CodeDatabaseError, code)
	}
}

func TestExec_Success(t *testing.T) {
	cleanup := setMockDefault(&mockDB{execRes: mockExecResult{affected: 3}})
	defer cleanup()

	result, err := Exec(context.Background(), "DELETE FROM users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	affected, _ := result.RowsAffected()
	if affected != 3 {
		t.Errorf("expected 3 rows affected, got %d", affected)
	}
}

func TestExec_Error(t *testing.T) {
	cleanup := setMockDefault(&mockDB{execErr: fmt.Errorf("table not found")})
	defer cleanup()

	_, err := Exec(context.Background(), "DELETE FROM users")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code := errors.ErrorCode(err); code != errors.CodeDatabaseError {
		t.Errorf("expected code %s, got %s", errors.CodeDatabaseError, code)
	}
}

func TestExecQ_UnpacksQuery(t *testing.T) {
	cleanup := setMockDefault(&mockDB{execRes: mockExecResult{affected: 1}})
	defer cleanup()

	q := sqlbuilder.Query{SQL: "INSERT INTO users (name) VALUES ($1)", Args: []any{"alice"}}
	result, err := ExecQ(context.Background(), q)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	affected, _ := result.RowsAffected()
	if affected != 1 {
		t.Errorf("expected 1 row affected, got %d", affected)
	}
}

func TestQueryAllQ_QueryError(t *testing.T) {
	cleanup := setMockDefault(&mockDB{queryErr: fmt.Errorf("bad query")})
	defer cleanup()

	q := sqlbuilder.Query{SQL: "SELECT 1", Args: nil}
	_, err := QueryAllQ[basicStruct](context.Background(), q)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code := errors.ErrorCode(err); code != errors.CodeDatabaseError {
		t.Errorf("expected code %s, got %s", errors.CodeDatabaseError, code)
	}
}

func TestQueryOneQ_QueryError(t *testing.T) {
	cleanup := setMockDefault(&mockDB{queryErr: fmt.Errorf("bad query")})
	defer cleanup()

	q := sqlbuilder.Query{SQL: "SELECT 1", Args: nil}
	_, err := QueryOneQ[basicStruct](context.Background(), q)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code := errors.ErrorCode(err); code != errors.CodeDatabaseError {
		t.Errorf("expected code %s, got %s", errors.CodeDatabaseError, code)
	}
}
