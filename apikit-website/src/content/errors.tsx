import CodeBlock from '../components/CodeBlock';
import ModuleSection from '../components/ModuleSection';

export default function ErrorsDocs() {
  return (
    <ModuleSection
      id="errors"
      title="errors"
      description="Structured API errors that integrate with Go's standard errors package."
      importPath="github.com/KARTIKrocks/apikit/errors"
      features={[
        'HTTP-aware errors with status codes and error codes',
        'Full errors.Is/errors.As support for error matching',
        'Sentinel errors (ErrNotFound, ErrValidation, etc.)',
        'Field-level details and error wrapping',
        'Custom error code registration',
      ]}
    >
      {/* ── Error Factories ── */}
      <h3 id="errors-factories" className="text-lg font-semibold text-text-heading mt-8 mb-2">Error Factories</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border text-left">
              <th className="py-2 pr-4 text-text-heading font-semibold">Function</th>
              <th className="py-2 text-text-heading font-semibold">Description</th>
            </tr>
          </thead>
          <tbody>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">NotFound(msg)</td><td className="py-2 text-text-muted">Returns a 404 NOT_FOUND error</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">BadRequest(msg)</td><td className="py-2 text-text-muted">Returns a 400 BAD_REQUEST error</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Validation(msg, fields)</td><td className="py-2 text-text-muted">Returns a 422 VALIDATION error with field-level details</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Conflict(msg)</td><td className="py-2 text-text-muted">Returns a 409 CONFLICT error</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Internal(msg)</td><td className="py-2 text-text-muted">Returns a 500 INTERNAL error</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Unauthorized(msg)</td><td className="py-2 text-text-muted">Returns a 401 UNAUTHORIZED error</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Forbidden(msg)</td><td className="py-2 text-text-muted">Returns a 403 FORBIDDEN error</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">TooManyRequests(msg)</td><td className="py-2 text-text-muted">Returns a 429 TOO_MANY_REQUESTS error</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">New(code, msg)</td><td className="py-2 text-text-muted">Creates an error with a custom code</td></tr>
          </tbody>
        </table>
      </div>
      <CodeBlock code={`// Create errors with factory functions
err := errors.NotFound("User")            // 404
err := errors.BadRequest("Invalid input") // 400
err := errors.Unauthorized("Bad token")   // 401
err := errors.Forbidden("Not allowed")    // 403
err := errors.Conflict("Duplicate email") // 409
err := errors.Internal("Database failed") // 500
err := errors.TooManyRequests("Slow down") // 429

// Validation errors with per-field details
err := errors.Validation("Invalid input", map[string]string{
    "email": "is required",
    "name":  "too short",
}) // 422`} />

      {/* ── Wrapping & Details ── */}
      <h3 id="errors-wrapping" className="text-lg font-semibold text-text-heading mt-8 mb-2">Wrapping & Details</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border text-left">
              <th className="py-2 pr-4 text-text-heading font-semibold">Method</th>
              <th className="py-2 text-text-heading font-semibold">Description</th>
            </tr>
          </thead>
          <tbody>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Wrap(err)</td><td className="py-2 text-text-muted">Wrap an underlying error (for errors.Is/As chain)</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.WithField(key, msg)</td><td className="py-2 text-text-muted">Add a field-level error detail</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.WithDetail(key, val)</td><td className="py-2 text-text-muted">Add arbitrary metadata to the error</td></tr>
          </tbody>
        </table>
      </div>
      <CodeBlock code={`// Wrap underlying errors — preserves the chain for errors.Is/As
err := errors.Internal("Database failed").Wrap(dbErr)

// Add field-level details
err := errors.Conflict("Duplicate").
    WithField("email", "already exists").
    WithField("username", "already taken")

// Add arbitrary metadata
err := errors.NotFound("User").
    WithDetail("requested_id", "abc123").
    WithDetail("suggestion", "try /users/search")`} />

      {/* ── Error Checking ── */}
      <h3 id="errors-checking" className="text-lg font-semibold text-text-heading mt-8 mb-2">Error Checking</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border text-left">
              <th className="py-2 pr-4 text-text-heading font-semibold">Function / Sentinel</th>
              <th className="py-2 text-text-heading font-semibold">Description</th>
            </tr>
          </thead>
          <tbody>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Is(err, target)</td><td className="py-2 text-text-muted">Check if err matches a sentinel error</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">As(err, target)</td><td className="py-2 text-text-muted">Extract *errors.Error from the chain</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">ErrNotFound</td><td className="py-2 text-text-muted">Sentinel for 404 errors</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">ErrBadRequest</td><td className="py-2 text-text-muted">Sentinel for 400 errors</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">ErrValidation</td><td className="py-2 text-text-muted">Sentinel for 422 errors</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">ErrConflict</td><td className="py-2 text-text-muted">Sentinel for 409 errors</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">ErrInternal</td><td className="py-2 text-text-muted">Sentinel for 500 errors</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">ErrUnauthorized</td><td className="py-2 text-text-muted">Sentinel for 401 errors</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">ErrForbidden</td><td className="py-2 text-text-muted">Sentinel for 403 errors</td></tr>
          </tbody>
        </table>
      </div>
      <CodeBlock code={`// Check errors with sentinels
if errors.Is(err, errors.ErrNotFound) {
    // handle 404
}
if errors.Is(err, errors.ErrValidation) {
    // handle 422
}

// Extract structured error details
var apiErr *errors.Error
if errors.As(err, &apiErr) {
    log.Printf("Status: %d, Code: %s", apiErr.StatusCode, apiErr.Code)
    log.Printf("Message: %s", apiErr.Message)
    log.Printf("Fields: %v", apiErr.Fields)
}`} />

      {/* ── Custom Codes ── */}
      <h3 id="errors-custom" className="text-lg font-semibold text-text-heading mt-8 mb-2">Custom Codes</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border text-left">
              <th className="py-2 pr-4 text-text-heading font-semibold">Function</th>
              <th className="py-2 text-text-heading font-semibold">Description</th>
            </tr>
          </thead>
          <tbody>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">RegisterCode(code, status)</td><td className="py-2 text-text-muted">Register a custom error code with its HTTP status</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">New(code, msg)</td><td className="py-2 text-text-muted">Create an error using a registered (or any) code</td></tr>
          </tbody>
        </table>
      </div>
      <CodeBlock code={`// Register custom error codes at init time
errors.RegisterCode("SUBSCRIPTION_EXPIRED", 402)
errors.RegisterCode("FEATURE_DISABLED", 403)
errors.RegisterCode("QUOTA_EXCEEDED", 429)

// Use custom codes
err := errors.New("SUBSCRIPTION_EXPIRED", "Your subscription has expired")
err := errors.New("QUOTA_EXCEEDED", "API quota exceeded").
    WithDetail("limit", "1000").
    WithDetail("reset_at", "2024-01-01T00:00:00Z")`} />
    </ModuleSection>
  );
}
