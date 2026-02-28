import CodeBlock from '../components/CodeBlock';
import ModuleSection from '../components/ModuleSection';

export default function ServerDocs() {
  return (
    <ModuleSection
      id="server"
      title="server"
      description="Production-ready HTTP server with graceful shutdown, signal handling, and lifecycle hooks."
      importPath="github.com/KARTIKrocks/apikit/server"
      features={[
        'Graceful shutdown with SIGINT/SIGTERM handling',
        'Configurable timeouts (read, write, idle, shutdown)',
        'TLS support with WithTLS option',
        'OnStart and OnShutdown lifecycle hooks',
        'Structured logging with slog',
      ]}
    >
      <h3 id="server-creating" className="text-lg font-semibold text-text-heading mt-8 mb-2">Creating a Server</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">New(handler, opts...)</td><td className="py-2 text-text-muted">Create a new server with options</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Start()</td><td className="py-2 text-text-muted">Start server (blocks until shutdown)</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.Addr()</td><td className="py-2 text-text-muted">Get the listen address</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`srv := server.New(r, server.WithAddr(":8080"))
if err := srv.Start(); err != nil {
    log.Fatal(err)
}`} />

      <h3 id="server-options" className="text-lg font-semibold text-text-heading mt-8 mb-2">Options</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Option</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">WithAddr(addr)</td><td className="py-2 text-text-muted">Set listen address (default ":8080")</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">WithReadTimeout(d)</td><td className="py-2 text-text-muted">Set read timeout</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">WithWriteTimeout(d)</td><td className="py-2 text-text-muted">Set write timeout</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">WithIdleTimeout(d)</td><td className="py-2 text-text-muted">Set idle connection timeout</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">WithShutdownTimeout(d)</td><td className="py-2 text-text-muted">Set graceful shutdown timeout</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">WithLogger(logger)</td><td className="py-2 text-text-muted">Set structured logger</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">WithTLS(cert, key)</td><td className="py-2 text-text-muted">Enable HTTPS</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`srv := server.New(handler,
    server.WithAddr(":8080"),
    server.WithReadTimeout(15 * time.Second),
    server.WithWriteTimeout(60 * time.Second),
    server.WithIdleTimeout(120 * time.Second),
    server.WithShutdownTimeout(10 * time.Second),
    server.WithLogger(slog.Default()),
)`} />

      <h3 id="server-lifecycle" className="text-lg font-semibold text-text-heading mt-8 mb-2">Lifecycle Hooks</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Method</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.OnStart(fn)</td><td className="py-2 text-text-muted">Hook called before listening</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">.OnShutdown(fn)</td><td className="py-2 text-text-muted">Hook called during graceful shutdown</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`srv.OnStart(func() error {
    slog.Info("connecting to database...")
    return db.Connect()
})
srv.OnShutdown(func(ctx context.Context) error {
    return db.Close()
})`} />

      <h3 id="server-tls" className="text-lg font-semibold text-text-heading mt-8 mb-2">TLS</h3>
      <CodeBlock code={`srv := server.New(handler,
    server.WithAddr(":443"),
    server.WithTLS("cert.pem", "key.pem"),
)
srv.Start()`} />
    </ModuleSection>
  );
}
