import { useState } from 'react';

const features = [
  { title: 'Zero Dependencies', desc: 'Core uses only the Go standard library' },
  { title: 'Stdlib Compatible', desc: 'Works with http.Handler and any router' },
  { title: 'Type-Safe Generics', desc: 'Bind[T], QueryAll[T] for compile-time safety' },
  { title: '12 Modules', desc: 'Everything from routing to SQL building' },
  { title: 'Go 1.22+', desc: 'Leverages enhanced http.ServeMux routing' },
  { title: 'Production Ready', desc: 'Graceful shutdown, circuit breakers, rate limiting' },
];

export default function Hero() {
  const [copied, setCopied] = useState(false);
  const installCmd = 'go get github.com/KARTIKrocks/apikit';

  const handleCopy = () => {
    navigator.clipboard.writeText(installCmd);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <section id="top" className="py-16 border-b border-border">
      <h1 className="text-4xl md:text-5xl font-bold text-text-heading mb-4">
        Production-ready Go API toolkit
      </h1>
      <p className="text-lg text-text-muted max-w-2xl mb-8">
        A composable set of packages for building REST APIs in Go. Structured errors,
        request binding, JSON responses, middleware, routing, SQL building, and more —
        all with zero mandatory dependencies.
      </p>

      <div className="flex items-center gap-2 bg-bg-card border border-border rounded-lg px-4 py-3 max-w-lg mb-10">
        <span className="text-text-muted select-none">$</span>
        <code className="flex-1 text-sm font-mono text-accent">{installCmd}</code>
        <button
          onClick={handleCopy}
          className="text-xs text-text-muted hover:text-text px-2 py-1 rounded bg-white/5 hover:bg-white/10 transition-colors"
        >
          {copied ? 'Copied!' : 'Copy'}
        </button>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {features.map((f) => (
          <div key={f.title} className="bg-bg-card border border-border rounded-lg p-4">
            <h3 className="text-sm font-semibold text-text-heading mb-1">{f.title}</h3>
            <p className="text-xs text-text-muted">{f.desc}</p>
          </div>
        ))}
      </div>
    </section>
  );
}
