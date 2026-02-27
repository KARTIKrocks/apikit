import { type ReactNode } from 'react';

interface ModuleSectionProps {
  id: string;
  title: string;
  description: string;
  importPath: string;
  features?: string[];
  children: ReactNode;
  apiTable?: { name: string; description: string }[];
}

export default function ModuleSection({
  id,
  title,
  description,
  importPath,
  features,
  children,
  apiTable,
}: ModuleSectionProps) {
  return (
    <section id={id} className="py-10 border-b border-border last:border-b-0">
      <h2 className="text-2xl font-bold text-text-heading mb-2">{title}</h2>
      <p className="text-text-muted mb-3">{description}</p>
      <code className="text-sm bg-bg-card px-2 py-1 rounded text-accent font-mono">
        import "{importPath}"
      </code>

      {features && features.length > 0 && (
        <ul className="mt-4 space-y-1">
          {features.map((f, i) => (
            <li key={i} className="text-sm text-text flex items-start gap-2">
              <span className="text-primary mt-0.5 shrink-0">&#x2022;</span>
              {f}
            </li>
          ))}
        </ul>
      )}

      <div className="mt-6">{children}</div>

      {apiTable && apiTable.length > 0 && (
        <div className="mt-6 overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border text-left">
                <th className="py-2 pr-4 text-text-heading font-semibold">Function</th>
                <th className="py-2 text-text-heading font-semibold">Description</th>
              </tr>
            </thead>
            <tbody>
              {apiTable.map((row, i) => (
                <tr key={i} className="border-b border-border/50">
                  <td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">{row.name}</td>
                  <td className="py-2 text-text-muted">{row.description}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}
