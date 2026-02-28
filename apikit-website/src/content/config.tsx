import CodeBlock from '../components/CodeBlock';
import ModuleSection from '../components/ModuleSection';

export default function ConfigDocs() {
  return (
    <ModuleSection
      id="config"
      title="config"
      description="Load application configuration from environment variables, .env files, and JSON config files into typed Go structs."
      importPath="github.com/KARTIKrocks/apikit/config"
      features={[
        'Load from env vars, .env files, and JSON files',
        'Typed struct mapping with env tags',
        'Default values via default tags',
        'Validation via validate tags',
        'Nested struct support with automatic flattening',
        'Duration and slice types supported',
        'Priority: env vars > .env file > JSON file > defaults',
      ]}
    >
      <h3 id="config-loading" className="text-lg font-semibold text-text-heading mt-8 mb-2">Loading Config</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Function</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Load(cfg, opts...)</td><td className="py-2 text-text-muted">Load config, return error on failure</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">MustLoad(cfg, opts...)</td><td className="py-2 text-text-muted">Load config or panic</td></tr>
        </tbody></table>
      </div>
      <CodeBlock code={`type AppConfig struct {
    Host     string        \`env:"HOST"      default:"localhost"  validate:"required"\`
    Port     int           \`env:"PORT"      default:"8080"       validate:"required,min=1,max=65535"\`
    Debug    bool          \`env:"DEBUG"     default:"false"\`
    DBUrl    string        \`env:"DB_URL"    validate:"required,url"\`
    LogLevel string        \`env:"LOG_LEVEL" default:"info"       validate:"oneof=debug info warn error"\`
    Timeout  time.Duration \`env:"TIMEOUT"   default:"30s"\`
    Tags     []string      \`env:"TAGS"\`
}

var cfg AppConfig
config.MustLoad(&cfg,
    config.WithPrefix("APP"),
    config.WithEnvFile(".env"),
    config.WithJSONFile("config.json"),
)`} />

      <h3 id="config-options" className="text-lg font-semibold text-text-heading mt-8 mb-2">Options</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Option</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">WithPrefix(prefix)</td><td className="py-2 text-text-muted">Read env vars with prefix (APP_HOST)</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">WithEnvFile(path)</td><td className="py-2 text-text-muted">Load a .env file</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">WithJSONFile(path)</td><td className="py-2 text-text-muted">Load JSON config as base layer</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">WithRequired(fields...)</td><td className="py-2 text-text-muted">Require specific env vars</td></tr>
        </tbody></table>
      </div>

      <h3 id="config-tags" className="text-lg font-semibold text-text-heading mt-8 mb-2">Struct Tags</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Tag</th><th className="py-2 text-text-heading font-semibold">Description</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent">env:"VAR_NAME"</td><td className="py-2 text-text-muted">Map field to environment variable</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent">default:"value"</td><td className="py-2 text-text-muted">Default value if not set</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent">validate:"rules"</td><td className="py-2 text-text-muted">Validation rules</td></tr>
        </tbody></table>
      </div>

      <h3 id="config-nested" className="text-lg font-semibold text-text-heading mt-8 mb-2">Nested Structs</h3>
      <CodeBlock code={`type Config struct {
    DB struct {
        Host     string \`env:"HOST"     default:"localhost"\`
        Port     int    \`env:"PORT"     default:"5432"\`
        Password string \`env:"PASSWORD" validate:"required"\`
    }
    Redis struct {
        Addr string \`env:"ADDR" default:"localhost:6379"\`
        DB   int    \`env:"DB"   default:"0"\`
    }
}
// Reads: DB_HOST, DB_PORT, DB_PASSWORD, REDIS_ADDR, REDIS_DB
// With prefix "APP": APP_DB_HOST, APP_REDIS_ADDR, etc.`} />

      <h3 id="config-types" className="text-lg font-semibold text-text-heading mt-8 mb-2">Supported Types</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm"><thead><tr className="border-b border-border text-left"><th className="py-2 pr-4 text-text-heading font-semibold">Go Type</th><th className="py-2 text-text-heading font-semibold">Env Format</th></tr></thead><tbody>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent">string</td><td className="py-2 text-text-muted">Plain string</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent">int, int64</td><td className="py-2 text-text-muted">Integer ("8080")</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent">float64</td><td className="py-2 text-text-muted">Float ("3.14")</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent">bool</td><td className="py-2 text-text-muted">"true", "false", "1", "0"</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent">time.Duration</td><td className="py-2 text-text-muted">"30s", "5m", "1h"</td></tr>
          <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent">[]string</td><td className="py-2 text-text-muted">Comma-separated ("a,b,c")</td></tr>
        </tbody></table>
      </div>
    </ModuleSection>
  );
}
