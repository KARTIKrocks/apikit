import CodeBlock from '../components/CodeBlock';
import ModuleSection from '../components/ModuleSection';

export default function SqlbuilderDocs() {
  return (
    <ModuleSection
      id="sqlbuilder"
      title="sqlbuilder"
      description="Fluent SQL query builder for PostgreSQL, MySQL, and SQLite. Produces (string, []any) pairs — no database/sql dependency."
      importPath="github.com/KARTIKrocks/apikit/sqlbuilder"
      features={[
        'SELECT, INSERT, UPDATE, DELETE builders',
        'PostgreSQL ($1), MySQL (?), and SQLite dialects',
        'JOINs, GROUP BY, HAVING, subqueries, CTEs, UNION',
        'Upsert (ON CONFLICT) and batch insert',
        'Convenience Where helpers: WhereEq, WhereGt, WhereLike, etc.',
        'Automatic placeholder rebasing across multiple Where calls',
        'Conditional building with When()',
        'Clone() for safe query reuse',
        'Integration with request package (pagination, sorting, filtering)',
      ]}
    >
      <h3 className="text-lg font-semibold text-text-heading mb-2">SELECT</h3>
      <CodeBlock code={`sql, args := sqlbuilder.Select("id", "name", "email").
    From("users").
    Where("active = $1", true).
    OrderBy("name ASC").
    Limit(20).
    Build()
// sql:  "SELECT id, name, email FROM users WHERE active = $1 ORDER BY ..."
// args: [true]

// Convenience Where helpers — no placeholder syntax needed
sql, args := sqlbuilder.Select("id").From("users").
    WhereEq("status", "active").
    WhereGt("age", 18).
    WhereLike("name", "A%").
    Build()

// JOINs, GROUP BY, HAVING
sql, args := sqlbuilder.Select("u.id", "COUNT(o.id) as orders").
    From("users u").
    LeftJoin("orders o", "o.user_id = u.id").
    Where("u.active = $1", true).
    GroupBy("u.id").
    Having("COUNT(o.id) > $1", 5).
    Build()`} />

      <h3 className="text-lg font-semibold text-text-heading mt-6 mb-2">INSERT & Upsert</h3>
      <CodeBlock code={`sql, args := sqlbuilder.Insert("users").
    Columns("name", "email").
    Values("Alice", "alice@example.com").
    Returning("id").
    Build()

// Upsert (ON CONFLICT)
sql, args := sqlbuilder.Insert("users").
    Columns("email", "name").
    Values("alice@example.com", "Alice").
    OnConflictUpdate([]string{"email"}, map[string]any{
        "name": "Alice Updated",
    }).
    Build()`} />

      <h3 className="text-lg font-semibold text-text-heading mt-6 mb-2">UPDATE & DELETE</h3>
      <CodeBlock code={`sql, args := sqlbuilder.Update("users").
    Set("name", "Bob").
    SetExpr("updated_at", sqlbuilder.Raw("NOW()")).
    WhereEq("id", 1).
    Build()

sql, args = sqlbuilder.Update("products").
    Increment("view_count", 1).
    WhereEq("id", 42).
    Build()

sql, args := sqlbuilder.Delete("users").
    WhereEq("id", 1).
    Returning("id", "name").
    Build()`} />

      <h3 className="text-lg font-semibold text-text-heading mt-6 mb-2">MySQL / SQLite Dialect</h3>
      <CodeBlock code={`sql, args = sqlbuilder.Select("id", "name").
    From("users").
    WhereEq("active", true).
    SetDialect(sqlbuilder.MySQL).
    Build()
// sql:  "SELECT id, name FROM users WHERE active = ?"
// args: [true]`} />

      <h3 className="text-lg font-semibold text-text-heading mt-6 mb-2">Request Integration</h3>
      <CodeBlock code={`pg, _ := request.Paginate(r)
sorts, _ := request.ParseSort(r, sortCfg)
filters, _ := request.ParseFilters(r, filterCfg)

cols := map[string]string{
    "name": "u.name", "created_at": "u.created_at",
}

sql, args := sqlbuilder.Select("u.id", "u.name", "u.email").
    From("users u").
    Where("u.active = $1", true).
    ApplyFilters(filters, cols).
    ApplySort(sorts, cols).
    ApplyPagination(pg).
    Build()`} />
    </ModuleSection>
  );
}
