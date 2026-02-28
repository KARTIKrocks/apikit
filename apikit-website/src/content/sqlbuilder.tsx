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
        'Aggregate functions: Count, Sum, Avg, Min, Max',
        'Expression helpers: Raw, RawExpr, Coalesce, Cast, NullIf via RawExpr',
        'CASE/WHEN and window functions via RawExpr',
        'Locking: FOR UPDATE, FOR SHARE, SKIP LOCKED, NOWAIT',
        'Set operations: Union, UnionAll, Intersect, Except',
        'Convenience Where helpers: WhereEq, WhereGt, WhereLike, etc.',
        'Conditional building with When() and Clone() for safe reuse',
        'Integration with request package (pagination, sorting, filtering)',
      ]}
    >
      <h3 id="sqlbuilder-select" className="text-lg font-semibold text-text-heading mt-8 mb-2">SELECT Basics</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Select(columns...)</td><td className="py-2 text-text-muted">Create SELECT with columns</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">SelectExpr(exprs...)</td><td className="py-2 text-text-muted">Create SELECT with expressions</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">SelectWith(d, columns...)</td><td className="py-2 text-text-muted">Create SELECT with dialect</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.From(table) / .FromAlias(table, alias)</td><td className="py-2 text-text-muted">FROM clause</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Distinct()</td><td className="py-2 text-text-muted">SELECT DISTINCT</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Limit(n) / .Offset(n)</td><td className="py-2 text-text-muted">LIMIT and OFFSET</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Build() / .MustBuild() / .Query() / .String()</td><td className="py-2 text-text-muted">Build output</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`sql, args := sqlbuilder.Select("id", "name", "email").
    From("users").
    Where("active = $1", true).
    OrderBy("name ASC").
    Limit(20).
    Build()
// "SELECT id, name, email FROM users WHERE active = $1 ORDER BY name ASC LIMIT 20"

// DISTINCT
sql, args := sqlbuilder.Select("department").From("employees").Distinct().Build()

// From with alias
sql, args := sqlbuilder.Select("u.id", "u.name").FromAlias("users", "u").Build()`} />

      <h3 id="sqlbuilder-columns" className="text-lg font-semibold text-text-heading mt-8 mb-2">Column Expressions</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Column(col) / .Columns(cols...) / .ColumnExpr(expr)</td><td className="py-2 text-text-muted">Add columns</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Count(col) / CountDistinct(col)</td><td className="py-2 text-text-muted">COUNT / COUNT(DISTINCT)</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Sum(col) / Avg(col) / Min(col) / Max(col)</td><td className="py-2 text-text-muted">Aggregate functions</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Raw(sql) / RawExpr(sql, args...)</td><td className="py-2 text-text-muted">Raw SQL expressions</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">expr.As(alias)</td><td className="py-2 text-text-muted">Alias an expression</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`sql, args := sqlbuilder.SelectExpr(
    sqlbuilder.Count("*").As("total"),
    sqlbuilder.Sum("amount").As("total_amount"),
    sqlbuilder.Avg("price").As("avg_price"),
    sqlbuilder.Min("created_at").As("earliest"),
    sqlbuilder.Max("created_at").As("latest"),
).From("orders").Build()

sql, args := sqlbuilder.SelectExpr(
    sqlbuilder.CountDistinct("user_id").As("unique_users"),
).From("orders").Build()`} />

      <h3 id="sqlbuilder-where" className="text-lg font-semibold text-text-heading mt-8 mb-2">WHERE Helpers</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Method</th><th className="py-2 text-text-heading font-semibold">SQL</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.WhereEq / .WhereNeq</td><td className="py-2 text-text-muted">= / !=</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.WhereGt / .WhereGte / .WhereLt / .WhereLte</td><td className="py-2 text-text-muted">&gt; / &gt;= / &lt; / &lt;=</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.WhereLike / .WhereILike</td><td className="py-2 text-text-muted">LIKE / ILIKE</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.WhereIn / .WhereNotIn</td><td className="py-2 text-text-muted">IN (...) / NOT IN (...)</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.WhereNull / .WhereNotNull</td><td className="py-2 text-text-muted">IS NULL / IS NOT NULL</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.WhereBetween(col, lo, hi)</td><td className="py-2 text-text-muted">BETWEEN ... AND ...</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.WhereOr(conditions...)</td><td className="py-2 text-text-muted">(cond1 OR cond2 OR ...)</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.WhereExists / .WhereNotExists</td><td className="py-2 text-text-muted">EXISTS / NOT EXISTS subquery</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.WhereInSubquery / .WhereNotInSubquery</td><td className="py-2 text-text-muted">IN / NOT IN subquery</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`sql, args := sqlbuilder.Select("*").From("users").
    WhereEq("status", "active").
    WhereGt("age", 18).
    WhereLike("name", "A%").
    WhereIn("role", "admin", "editor").
    WhereNotNull("email").
    WhereBetween("created_at", startDate, endDate).
    Build()

// OR conditions
sql, args := sqlbuilder.Select("*").From("users").
    WhereOr(
        sqlbuilder.Or("email = $1", "alice@example.com"),
        sqlbuilder.Or("email = $1", "bob@example.com"),
    ).Build()

// Subquery conditions
sql, args := sqlbuilder.Select("*").From("users").
    WhereExists(sqlbuilder.Select("1").From("orders").Where("orders.user_id = users.id")).
    WhereInSubquery("id", sqlbuilder.Select("user_id").From("premium_members")).
    Build()`} />

      <h3 id="sqlbuilder-joins" className="text-lg font-semibold text-text-heading mt-8 mb-2">JOINs</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Method</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Join(table, on, args...)</td><td className="py-2 text-text-muted">INNER JOIN</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.LeftJoin / .RightJoin / .FullJoin</td><td className="py-2 text-text-muted">LEFT / RIGHT / FULL JOIN</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.CrossJoin(table)</td><td className="py-2 text-text-muted">CROSS JOIN</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`sql, args := sqlbuilder.Select("u.name", "o.total").
    From("users u").
    Join("orders o", "o.user_id = u.id").
    LeftJoin("profiles p", "p.user_id = u.id").
    WhereGt("o.total", 100).
    Build()`} />

      <h3 id="sqlbuilder-join-subqueries" className="text-lg font-semibold text-text-heading mt-8 mb-2">Join Subqueries</h3>
      <CodeBlock code={`sql, args := sqlbuilder.Select("sub.id", "sub.total").
    FromSubquery(
        sqlbuilder.Select("user_id AS id", "SUM(amount) AS total").
            From("orders").GroupBy("user_id"),
        "sub",
    ).WhereGt("sub.total", 1000).Build()`} />

      <h3 id="sqlbuilder-subqueries" className="text-lg font-semibold text-text-heading mt-8 mb-2">Subqueries & CTEs</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Method</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.With(name, query) / .WithSelect(name, sub)</td><td className="py-2 text-text-muted">Add CTE</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.WithRecursive(name, query)</td><td className="py-2 text-text-muted">Add recursive CTE</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.FromSubquery(sub, alias)</td><td className="py-2 text-text-muted">FROM (subquery) AS alias</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`activeUsers := sqlbuilder.Select("id", "name").From("users").WhereEq("active", true)

sql, args := sqlbuilder.Select("id", "name").
    WithSelect("active_users", activeUsers).
    From("active_users").
    OrderBy("name ASC").
    Build()
// WITH active_users AS (SELECT ...) SELECT id, name FROM active_users ...`} />

      <h3 id="sqlbuilder-groupby" className="text-lg font-semibold text-text-heading mt-8 mb-2">GROUP BY & HAVING</h3>
      <CodeBlock code={`sql, args := sqlbuilder.Select("department", "COUNT(*) as cnt").
    From("employees").
    GroupBy("department").
    Having("COUNT(*) > $1", 3).
    OrderByDesc("cnt").
    Build()`} />

      <h3 id="sqlbuilder-orderby" className="text-lg font-semibold text-text-heading mt-8 mb-2">ORDER BY & Pagination</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Method</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.OrderBy(clauses...) / .OrderByAsc / .OrderByDesc</td><td className="py-2 text-text-muted">ORDER BY</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.OrderByExpr(expr)</td><td className="py-2 text-text-muted">ORDER BY expression</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`sql, args := sqlbuilder.Select("*").From("users").
    OrderByDesc("created_at").OrderByAsc("name").
    Limit(20).Offset(40).Build()

sql, args := sqlbuilder.Select("*").From("products").
    OrderByExpr(sqlbuilder.Raw("CASE WHEN featured THEN 0 ELSE 1 END")).Build()`} />

      <h3 id="sqlbuilder-locking" className="text-lg font-semibold text-text-heading mt-8 mb-2">Locking</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Method</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.ForUpdate() / .ForShare()</td><td className="py-2 text-text-muted">Row-level locks</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.SkipLocked() / .NoWait()</td><td className="py-2 text-text-muted">Lock modifiers</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`// Queue processing pattern
sql, args := sqlbuilder.Select("*").From("jobs").
    WhereEq("status", "pending").
    OrderByAsc("created_at").Limit(1).
    ForUpdate().SkipLocked().Build()

// Fail immediately if locked
sql, args := sqlbuilder.Select("*").From("inventory").
    WhereEq("product_id", 42).ForUpdate().NoWait().Build()`} />

      <h3 id="sqlbuilder-set-ops" className="text-lg font-semibold text-text-heading mt-8 mb-2">Set Operations</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Method</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Union(other) / .UnionAll(other)</td><td className="py-2 text-text-muted">UNION / UNION ALL</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Intersect(other) / .Except(other)</td><td className="py-2 text-text-muted">INTERSECT / EXCEPT</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`admins := sqlbuilder.Select("id", "name").From("admins")
editors := sqlbuilder.Select("id", "name").From("editors")
sql, args := admins.Union(editors).Build()

sql, args := sqlbuilder.Select("user_id").From("all_users").
    Except(sqlbuilder.Select("user_id").From("banned_users")).Build()`} />

      <h3 id="sqlbuilder-insert" className="text-lg font-semibold text-text-heading mt-8 mb-2">INSERT</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Method</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Insert(table)</td><td className="py-2 text-text-muted">Create INSERT builder</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Columns(cols...) / .Values(vals...)</td><td className="py-2 text-text-muted">Set columns and values</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.ValueMap(m) / .BatchValues(rows)</td><td className="py-2 text-text-muted">Map or batch insert</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.FromSelect(sel)</td><td className="py-2 text-text-muted">INSERT FROM SELECT</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Returning(cols...)</td><td className="py-2 text-text-muted">RETURNING clause</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`sql, args := sqlbuilder.Insert("users").
    Columns("name", "email").
    Values("Alice", "alice@example.com").
    Returning("id").Build()

// Batch insert
sql, args := sqlbuilder.Insert("users").
    Columns("name", "email").
    BatchValues([][]any{
        {"Alice", "alice@example.com"},
        {"Bob", "bob@example.com"},
    }).Build()

// INSERT FROM SELECT
sql, args := sqlbuilder.Insert("archive_users").
    Columns("id", "name").
    FromSelect(sqlbuilder.Select("id", "name").From("users").WhereLt("last_login", cutoff)).
    Build()`} />

      <h3 id="sqlbuilder-upsert" className="text-lg font-semibold text-text-heading mt-8 mb-2">Upsert</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Method</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.OnConflictDoNothing(target...)</td><td className="py-2 text-text-muted">ON CONFLICT DO NOTHING</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.OnConflictUpdate(target, updates)</td><td className="py-2 text-text-muted">ON CONFLICT DO UPDATE with values</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.OnConflictUpdateExpr(target, updates)</td><td className="py-2 text-text-muted">ON CONFLICT DO UPDATE with expressions</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`sql, args := sqlbuilder.Insert("users").
    Columns("email", "name").Values("alice@example.com", "Alice").
    OnConflictUpdate([]string{"email"}, map[string]any{"name": "Alice Updated"}).Build()

sql, args := sqlbuilder.Insert("counters").
    Columns("key", "count").Values("views", 1).
    OnConflictUpdateExpr([]string{"key"}, map[string]sqlbuilder.Expr{
        "count": sqlbuilder.Raw("counters.count + EXCLUDED.count"),
    }).Build()`} />

      <h3 id="sqlbuilder-update" className="text-lg font-semibold text-text-heading mt-8 mb-2">UPDATE</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Method</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Set(col, val) / .SetExpr(col, expr) / .SetMap(m)</td><td className="py-2 text-text-muted">SET clauses</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Increment(col, n) / .Decrement(col, n)</td><td className="py-2 text-text-muted">col = col +/- n</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.From(tables...)</td><td className="py-2 text-text-muted">FROM clause (PostgreSQL)</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Returning(cols...)</td><td className="py-2 text-text-muted">RETURNING clause</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`sql, args := sqlbuilder.Update("users").
    Set("name", "Bob").
    SetExpr("updated_at", sqlbuilder.Raw("NOW()")).
    WhereEq("id", 1).Returning("id", "name").Build()

sql, args := sqlbuilder.Update("products").
    Increment("view_count", 1).Decrement("stock", 1).
    WhereEq("id", 42).Build()

// Multi-table update (PostgreSQL)
sql, args := sqlbuilder.Update("orders").
    Set("status", "cancelled").From("users").
    Where("orders.user_id = users.id").WhereEq("users.banned", true).Build()`} />

      <h3 id="sqlbuilder-delete" className="text-lg font-semibold text-text-heading mt-8 mb-2">DELETE</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Method</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Delete(table)</td><td className="py-2 text-text-muted">Create DELETE builder</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Using(tables...)</td><td className="py-2 text-text-muted">USING clause (PostgreSQL)</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Returning(cols...)</td><td className="py-2 text-text-muted">RETURNING clause</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`sql, args := sqlbuilder.Delete("users").WhereEq("id", 1).Returning("id", "name").Build()

sql, args := sqlbuilder.Delete("orders").
    Using("users").Where("orders.user_id = users.id").
    WhereEq("users.banned", true).Returning("orders.id").Build()`} />

      <h3 id="sqlbuilder-expressions" className="text-lg font-semibold text-text-heading mt-8 mb-2">Expressions</h3>
      <p className="text-text-muted mb-3">Use <code className="text-accent">Raw()</code> and <code className="text-accent">RawExpr()</code> for Coalesce, NullIf, Cast, and other SQL expressions.</p>
      <CodeBlock code={`// COALESCE
sql, args := sqlbuilder.Select("id").
    ColumnExpr(sqlbuilder.RawExpr("COALESCE(nickname, name, $1)", "Anonymous").As("display_name")).
    From("users").Build()

// NULLIF
sql, args := sqlbuilder.Select("id").
    ColumnExpr(sqlbuilder.Raw("NULLIF(score, 0)").As("score")).From("results").Build()

// CAST
sql, args := sqlbuilder.Select("id").
    ColumnExpr(sqlbuilder.Raw("CAST(price AS INTEGER)").As("price_int")).From("products").Build()`} />

      <h3 id="sqlbuilder-case" className="text-lg font-semibold text-text-heading mt-8 mb-2">CASE / WHEN</h3>
      <CodeBlock code={`sql, args := sqlbuilder.Select("id", "name").
    ColumnExpr(sqlbuilder.RawExpr(
        "CASE WHEN status = $1 THEN 'Active' WHEN status = $2 THEN 'Inactive' ELSE 'Unknown' END",
        "active", "inactive",
    ).As("status_label")).From("users").Build()

// CASE in ORDER BY
sql, args := sqlbuilder.Select("*").From("tickets").
    OrderByExpr(sqlbuilder.RawExpr(
        "CASE priority WHEN $1 THEN 1 WHEN $2 THEN 2 WHEN $3 THEN 3 END",
        "critical", "high", "low",
    )).Build()`} />

      <h3 id="sqlbuilder-window" className="text-lg font-semibold text-text-heading mt-8 mb-2">Window Functions</h3>
      <CodeBlock code={`// ROW_NUMBER
sql, args := sqlbuilder.Select("id", "name", "department").
    ColumnExpr(sqlbuilder.Raw("ROW_NUMBER() OVER (PARTITION BY department ORDER BY salary DESC)").As("rank")).
    From("employees").Build()

// Running total
sql, args := sqlbuilder.Select("id", "amount").
    ColumnExpr(sqlbuilder.Raw("SUM(amount) OVER (ORDER BY created_at ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW)").As("running_total")).
    From("transactions").Build()

// LAG / LEAD
sql, args := sqlbuilder.Select("id", "amount").
    ColumnExpr(sqlbuilder.Raw("LAG(amount) OVER (ORDER BY created_at)").As("prev_amount")).
    From("transactions").Build()`} />

      <h3 id="sqlbuilder-returning" className="text-lg font-semibold text-text-heading mt-8 mb-2">RETURNING</h3>
      <p className="text-text-muted mb-3"><code className="text-accent">Returning()</code> is available on INSERT, UPDATE, and DELETE builders (PostgreSQL).</p>
      <CodeBlock code={`sql, args := sqlbuilder.Insert("users").Columns("name").Values("Alice").Returning("id", "created_at").Build()
sql, args := sqlbuilder.Update("users").Set("name", "Bob").WhereEq("id", 1).Returning("id", "updated_at").Build()
sql, args := sqlbuilder.Delete("sessions").WhereLt("expires_at", now).Returning("user_id").Build()`} />

      <h3 id="sqlbuilder-dialects" className="text-lg font-semibold text-text-heading mt-8 mb-2">Dialects</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Constant</th><th className="py-2 text-text-heading font-semibold">Placeholders</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent">Postgres</td><td className="py-2 text-text-muted">$1, $2, $3 (default)</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent">MySQL</td><td className="py-2 text-text-muted">?, ?, ?</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent">SQLite</td><td className="py-2 text-text-muted">?, ?, ?</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`sql, args := sqlbuilder.Select("id", "name").From("users").
    WhereEq("active", true).SetDialect(sqlbuilder.MySQL).Build()
// "SELECT id, name FROM users WHERE active = ?"`} />

      <h3 id="sqlbuilder-request" className="text-lg font-semibold text-text-heading mt-8 mb-2">Request Integration</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Method</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.ApplyPagination(p)</td><td className="py-2 text-text-muted">Apply LIMIT/OFFSET from request</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.ApplySort(sorts, cols)</td><td className="py-2 text-text-muted">Apply ORDER BY from request</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.ApplyFilters(filters, cols)</td><td className="py-2 text-text-muted">Apply WHERE from request</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`pg, _ := request.Paginate(r)
sorts, _ := request.ParseSort(r, sortCfg)
filters, _ := request.ParseFilters(r, filterCfg)

cols := map[string]string{"name": "u.name", "created_at": "u.created_at"}

sql, args := sqlbuilder.Select("u.id", "u.name", "u.email").
    From("users u").WhereEq("u.active", true).
    ApplyFilters(filters, cols).
    ApplySort(sorts, cols).
    ApplyPagination(pg).Build()`} />

      <h3 id="sqlbuilder-clone" className="text-lg font-semibold text-text-heading mt-8 mb-2">Clone & When</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Method</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Clone()</td><td className="py-2 text-text-muted">Deep copy for safe reuse</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.When(cond, fn)</td><td className="py-2 text-text-muted">Conditionally apply a function</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`// Clone — reuse a base query safely
base := sqlbuilder.Select("id", "name").From("users").WhereEq("active", true)
countQ := base.Clone().Columns("COUNT(*) as total")
listQ := base.Clone().OrderByDesc("created_at").Limit(20)

// When — conditional building
sql, args := sqlbuilder.Select("*").From("users").
    When(search != "", func(q *sqlbuilder.SelectBuilder) {
        q.WhereLike("name", "%"+search+"%")
    }).
    When(minAge > 0, func(q *sqlbuilder.SelectBuilder) {
        q.WhereGte("age", minAge)
    }).Build()`} />
    </ModuleSection>
  );
}
