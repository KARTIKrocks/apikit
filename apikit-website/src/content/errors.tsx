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
      apiTable={[
        { name: 'NotFound(msg)', description: 'Returns a 404 error' },
        { name: 'BadRequest(msg)', description: 'Returns a 400 error' },
        { name: 'Validation(msg, fields)', description: 'Returns a 422 error with field errors' },
        { name: 'Conflict(msg)', description: 'Returns a 409 error' },
        { name: 'Internal(msg)', description: 'Returns a 500 error' },
        { name: 'Unauthorized(msg)', description: 'Returns a 401 error' },
        { name: 'Forbidden(msg)', description: 'Returns a 403 error' },
        { name: 'New(code, msg)', description: 'Creates an error with a custom code' },
        { name: 'RegisterCode(code, status)', description: 'Register a custom error code and HTTP status' },
      ]}
    >
      <CodeBlock code={`// Create errors with factory functions
err := errors.NotFound("User")            // 404
err := errors.BadRequest("Invalid input") // 400
err := errors.Validation("Invalid", map[string]string{
    "email": "is required",
    "name":  "too short",
}) // 422

// Wrap underlying errors
err := errors.Internal("Database failed").Wrap(dbErr)

// Add details
err := errors.Conflict("Duplicate").
    WithField("email", "already exists").
    WithDetail("existing_id", "abc123")

// Check errors anywhere in your code
if errors.Is(err, errors.ErrNotFound) { /* ... */ }
if errors.Is(err, errors.ErrValidation) { /* ... */ }

// Extract error details
var apiErr *errors.Error
if errors.As(err, &apiErr) {
    log.Printf("Status: %d, Code: %s", apiErr.StatusCode, apiErr.Code)
}

// Custom error codes
errors.RegisterCode("SUBSCRIPTION_EXPIRED", 402)
err := errors.New("SUBSCRIPTION_EXPIRED", "Your subscription has expired")`} />
    </ModuleSection>
  );
}
