import CodeBlock from '../components/CodeBlock';
import ModuleSection from '../components/ModuleSection';

export default function RouterDocs() {
  return (
    <ModuleSection
      id="router"
      title="router"
      description="Route grouping and method helpers on top of http.ServeMux."
      importPath="github.com/KARTIKrocks/apikit/router"
      features={[
        'Method helpers: Get, Post, Put, Patch, Delete',
        'Route groups with prefix and per-group middleware',
        'Nested groups accumulate prefix and middleware',
        'Both error-returning and standard http.HandlerFunc handlers',
        'Middleware resolved at registration time (chi/echo/gin behavior)',
      ]}
      apiTable={[
        { name: 'New()', description: 'Create a new router (implements http.Handler)' },
        { name: 'Use(mw...)', description: 'Add global middleware' },
        { name: 'Get/Post/Put/Patch/Delete(path, handler)', description: 'Register error-returning handler' },
        { name: 'GetFunc/PostFunc/etc(path, handler)', description: 'Register standard http.HandlerFunc' },
        { name: 'Group(prefix, mw...)', description: 'Create a route group with prefix and middleware' },
        { name: 'Handle(pattern, handler)', description: 'Register an http.Handler' },
      ]}
    >
      <CodeBlock code={`r := router.New()

// Global middleware
r.Use(middleware.RequestID(), middleware.Logger(slog.Default()))

// Error-returning handlers
r.Get("/health", func(w http.ResponseWriter, r *http.Request) error {
    response.OK(w, "OK", nil)
    return nil
})

// Standard http.HandlerFunc — use GetFunc/PostFunc/etc.
r.GetFunc("/version", func(w http.ResponseWriter, r *http.Request) {
    response.OK(w, "OK", map[string]string{"version": "1.0.0"})
})

// Route groups with prefix and per-group middleware
api := r.Group("/api/v1", authMiddleware)
api.Get("/users", listUsers)
api.Post("/users", createUser)

// Nested groups accumulate prefix and middleware
admin := api.Group("/admin", adminOnly)
admin.Delete("/users/{id}", deleteUser)
// Registers "DELETE /api/v1/admin/users/{id}" with auth + adminOnly

// Use with server package
srv := server.New(r, server.WithAddr(":8080"))
srv.Start()`} />
    </ModuleSection>
  );
}
