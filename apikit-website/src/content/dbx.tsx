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
        'Integrates with sqlbuilder (QueryAllQ, QueryOneQ, ExecQ)',
        'Type mappings cached via sync.Map',
      ]}
    >
      <h3 id="dbx-setup" className="text-lg font-semibold text-text-heading mt-8 mb-2">Setup</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">SetDefault(db)</td><td className="py-2 text-text-muted">Set the default *sql.DB connection</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`db, _ := sql.Open("postgres", connStr)
dbx.SetDefault(db)

type User struct {
    ID    int     \`db:"id"\`
    Name  string  \`db:"name"\`
    Email *string \`db:"email"\` // nullable → pointer
}`} />

      <h3 id="dbx-queryall" className="text-lg font-semibold text-text-heading mt-8 mb-2">Querying Rows</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">QueryAll[T](ctx, sql, args...)</td><td className="py-2 text-text-muted">Fetch multiple rows into typed slice</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`users, err := dbx.QueryAll[User](ctx,
    "SELECT id, name, email FROM users WHERE active = $1", true,
)`} />

      <h3 id="dbx-queryone" className="text-lg font-semibold text-text-heading mt-8 mb-2">Single Row</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">QueryOne[T](ctx, sql, args...)</td><td className="py-2 text-text-muted">Fetch one row (ErrNotFound if none)</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`user, err := dbx.QueryOne[User](ctx,
    "SELECT id, name, email FROM users WHERE id = $1", 42,
)
if errors.Is(err, errors.ErrNotFound) {
    return errors.NotFound("User")
}`} />

      <h3 id="dbx-exec" className="text-lg font-semibold text-text-heading mt-8 mb-2">Executing</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Exec(ctx, sql, args...)</td><td className="py-2 text-text-muted">Execute a statement</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`result, err := dbx.Exec(ctx, "DELETE FROM users WHERE id = $1", 42)
rowsAffected, _ := result.RowsAffected()`} />

      <h3 id="dbx-sqlbuilder" className="text-lg font-semibold text-text-heading mt-8 mb-2">sqlbuilder Integration</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">QueryAllQ[T](ctx, query)</td><td className="py-2 text-text-muted">Fetch rows from sqlbuilder.Query</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">QueryOneQ[T](ctx, query)</td><td className="py-2 text-text-muted">Fetch one row from sqlbuilder.Query</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">ExecQ(ctx, query)</td><td className="py-2 text-text-muted">Execute sqlbuilder.Query</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`q := sqlbuilder.Select("id", "name", "email").
    From("users").WhereEq("active", true).
    OrderByDesc("created_at").Limit(20).Query()

users, err := dbx.QueryAllQ[User](ctx, q)

// Single row
q := sqlbuilder.Select("id", "name").From("users").WhereEq("id", 42).Query()
user, err := dbx.QueryOneQ[User](ctx, q)

// Execute
q := sqlbuilder.Update("users").Set("name", "Alice").WhereEq("id", 42).Query()
result, err := dbx.ExecQ(ctx, q)`} />

      <h3 id="dbx-transactions" className="text-lg font-semibold text-text-heading mt-8 mb-2">Transactions</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">WithTx(ctx, tx)</td><td className="py-2 text-text-muted">Attach transaction — all dbx calls use it</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`tx, _ := db.BeginTx(ctx, nil)
ctx = dbx.WithTx(ctx, tx)

dbx.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "Alice")
user, err := dbx.QueryOne[User](ctx, "SELECT id, name FROM users WHERE name = $1", "Alice")

if err != nil { tx.Rollback(); return err }
tx.Commit()`} />

      <h3 id="dbx-tags" className="text-lg font-semibold text-text-heading mt-8 mb-2">Struct Tags</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Tag</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent">db:"column_name"</td><td className="py-2 text-text-muted">Map field to SQL column</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent">db:"-"</td><td className="py-2 text-text-muted">Skip this field</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`type User struct {
    ID       int       \`db:"id"\`
    Name     string    \`db:"name"\`
    Email    *string   \`db:"email"\`      // nullable → pointer
    Password string    \`db:"-"\`           // skip
}

// Embedded structs are flattened
type Timestamps struct {
    CreatedAt time.Time \`db:"created_at"\`
    UpdatedAt time.Time \`db:"updated_at"\`
}
type Post struct {
    ID    int    \`db:"id"\`
    Title string \`db:"title"\`
    Timestamps
}`} />
    </ModuleSection>
  );
}
