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
      apiTable={[
        { name: 'MustLoad(cfg, opts...)', description: 'Load config or panic on error' },
        { name: 'WithPrefix(prefix)', description: 'Read env vars with a prefix (e.g., APP_)' },
        { name: 'WithEnvFile(path)', description: 'Load a .env file' },
        { name: 'WithJSONFile(path)', description: 'Load a JSON config file' },
      ]}
    >
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
    config.WithPrefix("APP"),           // reads APP_HOST, APP_PORT, etc.
    config.WithEnvFile(".env"),          // load .env file
    config.WithJSONFile("config.json"), // JSON as base layer
)`} />

      <h3 className="text-lg font-semibold text-text-heading mt-6 mb-2">Nested Structs</h3>
      <CodeBlock code={`type Config struct {
    DB struct {
        Host string \`env:"HOST" default:"localhost"\`
        Port int    \`env:"PORT" default:"5432"\`
    }
}

// Reads DB_HOST, DB_PORT (or APP_DB_HOST with WithPrefix("APP"))`} />
    </ModuleSection>
  );
}
