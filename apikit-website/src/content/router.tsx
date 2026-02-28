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
        'Custom error handler support',
      ]}
    >
      <h3 id="router-creating" className="text-lg font-semibold text-text-heading mt-8 mb-2">Creating a Router</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">New()</td><td className="py-2 text-text-muted">Create a new router (implements http.Handler)</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Use(mw...)</td><td className="py-2 text-text-muted">Add global middleware</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`r := router.New()
r.Use(middleware.RequestID(), middleware.Logger(slog.Default()))
srv := server.New(r, server.WithAddr(":8080"))
srv.Start()`} />

      <h3 id="router-methods" className="text-lg font-semibold text-text-heading mt-8 mb-2">Method Handlers</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Method</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Get(path, handler)</td><td className="py-2 text-text-muted">Register GET handler (returns error)</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Post(path, handler)</td><td className="py-2 text-text-muted">Register POST handler (returns error)</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Put(path, handler)</td><td className="py-2 text-text-muted">Register PUT handler (returns error)</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Patch(path, handler)</td><td className="py-2 text-text-muted">Register PATCH handler (returns error)</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Delete(path, handler)</td><td className="py-2 text-text-muted">Register DELETE handler (returns error)</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`r.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) error {
    id, err := request.PathParamInt(r, "id")
    if err != nil {
        return errors.BadRequest("Invalid ID")
    }
    user, err := db.FindUser(id)
    if err != nil {
        return err
    }
    response.OK(w, "Success", user)
    return nil
})

r.Post("/users", createUser)
r.Put("/users/{id}", updateUser)
r.Patch("/users/{id}", patchUser)
r.Delete("/users/{id}", deleteUser)`} />

      <h3 id="router-groups" className="text-lg font-semibold text-text-heading mt-8 mb-2">Route Groups</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Method</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Group(prefix, mw...)</td><td className="py-2 text-text-muted">Create a route group with prefix and optional middleware</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`api := r.Group("/api/v1", authMiddleware)
api.Get("/users", listUsers)
api.Post("/users", createUser)

// Nested groups accumulate prefix and middleware
admin := api.Group("/admin", adminOnly)
admin.Delete("/users/{id}", deleteUser)
// Registers "DELETE /api/v1/admin/users/{id}" with auth + adminOnly`} />

      <h3 id="router-stdlib" className="text-lg font-semibold text-text-heading mt-8 mb-2">Stdlib Handlers</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Method</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.GetFunc(path, handler)</td><td className="py-2 text-text-muted">Register standard http.HandlerFunc for GET</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.PostFunc / .PutFunc / .PatchFunc / .DeleteFunc</td><td className="py-2 text-text-muted">Same for other methods</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Handle(pattern, handler)</td><td className="py-2 text-text-muted">Register an http.Handler directly</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.HandleFunc(pattern, handler)</td><td className="py-2 text-text-muted">Register an http.HandlerFunc directly</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`r.GetFunc("/version", func(w http.ResponseWriter, r *http.Request) {
    response.OK(w, "OK", map[string]string{"version": "1.0.0"})
})

r.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./public"))))`} />

      <h3 id="router-errors" className="text-lg font-semibold text-text-heading mt-8 mb-2">Error Handling</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Method</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.WithErrorHandler(fn)</td><td className="py-2 text-text-muted">Set a custom error handler for all error-returning routes</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`r.WithErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) {
    slog.Error("handler error", "path", r.URL.Path, "err", err)
    response.Err(w, err)
})`} />
    </ModuleSection>
  );
}
