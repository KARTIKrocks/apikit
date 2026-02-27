package dbx

import (
	"context"
	"database/sql"

	"github.com/KARTIKrocks/apikit/errors"
)

// DB combines query and exec capabilities.
// It is satisfied by *sql.DB, *sql.Tx, and *sql.Conn.
type DB interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// defaultDB is the package-level connection set by SetDefault.
var defaultDB DB

// SetDefault sets the default database connection used by all
// package-level functions (QueryAll, QueryOne, Exec, etc.).
// It must be called once during application startup, before any
// queries are executed. It is not safe for concurrent use.
//
//	dbx.SetDefault(db) // *sql.DB
func SetDefault(db DB) {
	defaultDB = db
}

// txKey is the context key for transaction override.
type txKey struct{}

// WithTx returns a child context that makes all dbx functions in that
// context use tx instead of the default connection. This is how you
// run queries inside a transaction:
//
//	tx, _ := db.BeginTx(ctx, nil)
//	ctx := dbx.WithTx(ctx, tx)
//	dbx.Exec(ctx, "INSERT INTO users ...") // uses tx
//	dbx.QueryOne[User](ctx, "SELECT ...")  // uses tx
//	tx.Commit()
func WithTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// conn returns the DB to use for the given context: the transaction
// from WithTx if present, otherwise the default set by SetDefault.
func conn(ctx context.Context) (DB, error) {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok && tx != nil {
		return tx, nil
	}
	if defaultDB != nil {
		return defaultDB, nil
	}
	return nil, errors.Internal("dbx: no default connection set; call dbx.SetDefault(db) at startup")
}
