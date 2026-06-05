// Package dbx provides lightweight, generic helpers for scanning database/sql
// rows into Go structs. It eliminates repetitive scan boilerplate while keeping
// full SQL control — no ORM, no query generation, just row mapping.
//
// Struct fields are mapped to result columns via the `db` struct tag.
// Fields without a `db` tag are ignored. Use `db:"-"` to explicitly skip a field.
//
// # Quick Start
//
// Set the default connection once at startup:
//
//	dbx.SetDefault(db) // *sql.DB, *sql.Tx, or *sql.Conn
//
// Then use the package-level functions — no connection argument needed:
//
//	type User struct {
//	    ID    int     `db:"id"`
//	    Name  string  `db:"name"`
//	    Email *string `db:"email"` // nullable → pointer
//	}
//
//	// Fetch all users
//	users, err := dbx.QueryAll[User](ctx, "SELECT id, name, email FROM users")
//
//	// Fetch one user
//	user, err := dbx.QueryOne[User](ctx, "SELECT id, name, email FROM users WHERE id = $1", 42)
//
//	// Execute a statement
//	result, err := dbx.Exec(ctx, "DELETE FROM users WHERE id = $1", 42)
//
// # Transactions
//
// Use [WithTx] to run queries inside a transaction. All dbx functions
// called with the returned context will use the transaction:
//
//	tx, _ := db.BeginTx(ctx, nil)
//	ctx := dbx.WithTx(ctx, tx)
//
//	dbx.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "Alice")
//	user, err := dbx.QueryOne[User](ctx, "SELECT id, name FROM users WHERE name = $1", "Alice")
//	tx.Commit()
//
// # sqlbuilder Integration
//
// The Q-suffixed variants accept a [sqlbuilder.Query] directly:
//
//	q := sqlbuilder.Select("id", "name", "email").From("users").Where("id = $1", 42).Build()
//	user, err := dbx.QueryOneQ[User](ctx, q)
//
// # NULL Handling
//
// Use pointer types for nullable columns. A NULL value results in nil;
// a non-NULL value is scanned into the pointed-to type:
//
//	type Row struct {
//	    Name  string  `db:"name"`  // NOT NULL column
//	    Bio   *string `db:"bio"`   // nullable column
//	}
//
// # Error Handling
//
// SQL and scan errors are wrapped with [errors.CodeDatabaseError].
// [QueryOne] returns [errors.CodeNotFound] when no rows match.
// [QueryAll] returns an empty slice (not an error) for no rows.
package dbx
