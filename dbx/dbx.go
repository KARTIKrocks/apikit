package dbx

import (
	"context"
	"database/sql"
	"reflect"

	"github.com/KARTIKrocks/apikit/errors"
	"github.com/KARTIKrocks/apikit/sqlbuilder"
)

// QueryAll executes a query and returns all rows scanned into []T.
// T must be a struct with `db` tags on its fields.
// Returns an empty slice (not an error) when the query yields no rows.
func QueryAll[T any](ctx context.Context, query string, args ...any) ([]T, error) {
	db, err := conn(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.New(errors.CodeDatabaseError, "query failed").Wrap(err)
	}
	defer rows.Close() //nolint:errcheck // closing query rows

	return scanRows[T](rows)
}

// QueryOne executes a query and scans the first row into T.
// T must be a struct with `db` tags on its fields.
// Returns errors.CodeNotFound if the query yields no rows.
func QueryOne[T any](ctx context.Context, query string, args ...any) (T, error) {
	var zero T
	db, err := conn(ctx)
	if err != nil {
		return zero, err
	}
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return zero, errors.New(errors.CodeDatabaseError, "query failed").Wrap(err)
	}
	defer rows.Close() //nolint:errcheck // closing query rows

	return scanRow[T](rows)
}

// Exec executes a non-returning query (INSERT/UPDATE/DELETE without RETURNING).
func Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	db, err := conn(ctx)
	if err != nil {
		return nil, err
	}
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errors.New(errors.CodeDatabaseError, "exec failed").Wrap(err)
	}
	return result, nil
}

// QueryAllQ is like QueryAll but accepts a sqlbuilder.Query.
func QueryAllQ[T any](ctx context.Context, q sqlbuilder.Query) ([]T, error) {
	return QueryAll[T](ctx, q.SQL, q.Args...)
}

// QueryOneQ is like QueryOne but accepts a sqlbuilder.Query.
func QueryOneQ[T any](ctx context.Context, q sqlbuilder.Query) (T, error) {
	return QueryOne[T](ctx, q.SQL, q.Args...)
}

// ExecQ is like Exec but accepts a sqlbuilder.Query.
func ExecQ(ctx context.Context, q sqlbuilder.Query) (sql.Result, error) {
	return Exec(ctx, q.SQL, q.Args...)
}

// scanRows scans all remaining rows into a slice of T.
func scanRows[T any](rows *sql.Rows) ([]T, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, errors.New(errors.CodeDatabaseError, "failed to get columns").Wrap(err)
	}

	m := getMapping[T]()
	cm := buildColumnMapping(columns, m)
	targets := make([]any, len(columns))
	var discard sql.RawBytes
	var result []T

	for rows.Next() {
		var item T
		val := reflect.ValueOf(&item).Elem()
		scanInto(val, cm, targets, &discard)

		if err := rows.Scan(targets...); err != nil {
			return nil, errors.New(errors.CodeDatabaseError, "scan failed").Wrap(err)
		}
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.New(errors.CodeDatabaseError, "row iteration failed").Wrap(err)
	}

	if result == nil {
		result = []T{}
	}

	return result, nil
}

// scanRow scans the first row into T. Returns NotFound if there are no rows.
func scanRow[T any](rows *sql.Rows) (T, error) {
	var item T

	columns, err := rows.Columns()
	if err != nil {
		return item, errors.New(errors.CodeDatabaseError, "failed to get columns").Wrap(err)
	}

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return item, errors.New(errors.CodeDatabaseError, "row iteration failed").Wrap(err)
		}
		return item, errors.NotFound("record")
	}

	m := getMapping[T]()
	cm := buildColumnMapping(columns, m)
	targets := make([]any, len(columns))
	var discard sql.RawBytes
	val := reflect.ValueOf(&item).Elem()
	scanInto(val, cm, targets, &discard)

	if err := rows.Scan(targets...); err != nil {
		return item, errors.New(errors.CodeDatabaseError, "scan failed").Wrap(err)
	}

	return item, nil
}
