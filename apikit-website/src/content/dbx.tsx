import CodeBlock from '../components/CodeBlock';
import ModuleSection from '../components/ModuleSection';

export default function DbxDocs() {
  return (
    <ModuleSection
      id="dbx"
      title="dbx"
      description="Lightweight, generic row scanner for database/sql. Eliminates scan boilerplate while keeping full SQL control."
      importPath="github.com/KARTIKrocks/apikit/dbx"
      features={[
        'Generic QueryAll[T] and QueryOne[T] with automatic struct scanning',
        'Maps columns to struct fields via db tags',
        'Nullable columns via pointer fields',
        'Order-independent column matching',
        'Embedded struct support',
        'Transaction support via context',
        'Integrates with sqlbuilder (QueryAllQ, QueryOneQ)',
        'Type mappings cached via sync.Map',
      ]}
      apiTable={[
        { name: 'SetDefault(db)', description: 'Set default database connection' },
        { name: 'QueryAll[T](ctx, sql, args...)', description: 'Fetch multiple rows into a slice' },
        { name: 'QueryOne[T](ctx, sql, args...)', description: 'Fetch a single row' },
        { name: 'Exec(ctx, sql, args...)', description: 'Execute a statement' },
        { name: 'QueryAllQ[T](ctx, query)', description: 'Fetch rows from a sqlbuilder query' },
        { name: 'QueryOneQ[T](ctx, query)', description: 'Fetch one row from a sqlbuilder query' },
        { name: 'WithTx(ctx, tx)', description: 'Attach a transaction to a context' },
      ]}
    >
      <CodeBlock code={`// Set default connection once at startup
dbx.SetDefault(db)

// Define your struct with db tags
type User struct {
    ID    int     \`db:"id"\`
    Name  string  \`db:"name"\`
    Email *string \`db:"email"\` // nullable → pointer
}

// Fetch multiple rows
users, err := dbx.QueryAll[User](ctx,
    "SELECT id, name, email FROM users WHERE active = $1", true,
)

// Fetch one row (returns errors.CodeNotFound if no rows)
user, err := dbx.QueryOne[User](ctx,
    "SELECT id, name, email FROM users WHERE id = $1", 42,
)

// Execute statements
result, err := dbx.Exec(ctx, "DELETE FROM users WHERE id = $1", 42)`} />

      <h3 className="text-lg font-semibold text-text-heading mt-6 mb-2">sqlbuilder Integration</h3>
      <CodeBlock code={`q := sqlbuilder.Select("id", "name", "email").
    From("users").
    WhereEq("active", true).
    Build()
users, err := dbx.QueryAllQ[User](ctx, q)`} />

      <h3 className="text-lg font-semibold text-text-heading mt-6 mb-2">Transactions</h3>
      <CodeBlock code={`tx, _ := db.BeginTx(ctx, nil)
ctx = dbx.WithTx(ctx, tx)

dbx.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "Alice")
user, err := dbx.QueryOne[User](ctx,
    "SELECT id, name FROM users WHERE name = $1", "Alice",
)
tx.Commit()`} />
    </ModuleSection>
  );
}
