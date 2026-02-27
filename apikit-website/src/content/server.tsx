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
      apiTable={[
        { name: 'New(handler, opts...)', description: 'Create a new server' },
        { name: 'WithAddr(addr)', description: 'Set listen address' },
        { name: 'WithTLS(cert, key)', description: 'Enable HTTPS' },
        { name: 'WithReadTimeout(d)', description: 'Set read timeout' },
        { name: 'WithWriteTimeout(d)', description: 'Set write timeout' },
        { name: 'WithShutdownTimeout(d)', description: 'Set graceful shutdown timeout' },
        { name: 'OnStart(fn)', description: 'Register start lifecycle hook' },
        { name: 'OnShutdown(fn)', description: 'Register shutdown lifecycle hook' },
        { name: 'Start()', description: 'Start server (blocks until shutdown)' },
      ]}
    >
      <CodeBlock code={`srv := server.New(handler,
    server.WithAddr(":8080"),
    server.WithReadTimeout(15 * time.Second),
    server.WithWriteTimeout(60 * time.Second),
    server.WithIdleTimeout(120 * time.Second),
    server.WithShutdownTimeout(10 * time.Second),
    server.WithLogger(slog.Default()),
)

// HTTPS — just add WithTLS
srv := server.New(handler,
    server.WithAddr(":443"),
    server.WithTLS("cert.pem", "key.pem"),
)

// Lifecycle hooks
srv.OnStart(func() error {
    slog.Info("connecting to database...")
    return db.Connect()
})
srv.OnShutdown(func(ctx context.Context) error {
    return db.Close()
})

// Blocks until SIGINT/SIGTERM, then drains connections
if err := srv.Start(); err != nil {
    log.Fatal(err)
}`} />
    </ModuleSection>
  );
}
