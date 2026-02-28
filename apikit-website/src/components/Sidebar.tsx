import { useEffect, useState, useCallback } from 'react';

type SubSection = { id: string; label: string };
type Section = { id: string; label: string; children?: SubSection[] };

const sections: Section[] = [
  { id: 'getting-started', label: 'Getting Started' },
  {
    id: 'errors',
    label: 'errors',
    children: [
      { id: 'errors-factories', label: 'Error Factories' },
      { id: 'errors-wrapping', label: 'Wrapping & Details' },
      { id: 'errors-checking', label: 'Error Checking' },
      { id: 'errors-custom', label: 'Custom Codes' },
    ],
  },
  {
    id: 'router',
    label: 'router',
    children: [
      { id: 'router-creating', label: 'Creating a Router' },
      { id: 'router-methods', label: 'Method Handlers' },
      { id: 'router-groups', label: 'Route Groups' },
      { id: 'router-stdlib', label: 'Stdlib Handlers' },
      { id: 'router-errors', label: 'Error Handling' },
    ],
  },
  {
    id: 'middleware',
    label: 'middleware',
    children: [
      { id: 'middleware-core', label: 'Core Middleware' },
      { id: 'middleware-auth', label: 'Authentication' },
      { id: 'middleware-ratelimit', label: 'Rate Limiting' },
      { id: 'middleware-cors', label: 'CORS' },
      { id: 'middleware-composition', label: 'Composition' },
    ],
  },
  {
    id: 'request',
    label: 'request',
    children: [
      { id: 'request-binding', label: 'Body Binding' },
      { id: 'request-params', label: 'Path & Query Params' },
      { id: 'request-headers', label: 'Headers' },
      { id: 'request-pagination', label: 'Pagination' },
      { id: 'request-sorting', label: 'Sorting' },
      { id: 'request-filtering', label: 'Filtering' },
      { id: 'request-struct-validation', label: 'Struct Validation' },
      { id: 'request-programmatic-validation', label: 'Programmatic Validation' },
    ],
  },
  {
    id: 'response',
    label: 'response',
    children: [
      { id: 'response-success', label: 'Success Responses' },
      { id: 'response-error', label: 'Error Responses' },
      { id: 'response-builder', label: 'Builder Pattern' },
      { id: 'response-pagination', label: 'Pagination' },
      { id: 'response-streaming', label: 'Streaming & Formats' },
      { id: 'response-handler', label: 'Handler Wrapper' },
    ],
  },
  {
    id: 'server',
    label: 'server',
    children: [
      { id: 'server-creating', label: 'Creating a Server' },
      { id: 'server-options', label: 'Options' },
      { id: 'server-lifecycle', label: 'Lifecycle Hooks' },
      { id: 'server-tls', label: 'TLS' },
    ],
  },
  {
    id: 'config',
    label: 'config',
    children: [
      { id: 'config-loading', label: 'Loading Config' },
      { id: 'config-options', label: 'Options' },
      { id: 'config-tags', label: 'Struct Tags' },
      { id: 'config-nested', label: 'Nested Structs' },
      { id: 'config-types', label: 'Supported Types' },
    ],
  },
  {
    id: 'health',
    label: 'health',
    children: [
      { id: 'health-creating', label: 'Creating a Checker' },
      { id: 'health-checks', label: 'Adding Checks' },
      { id: 'health-handlers', label: 'HTTP Handlers' },
      { id: 'health-response', label: 'Response Format' },
    ],
  },
  {
    id: 'sqlbuilder',
    label: 'sqlbuilder',
    children: [
      { id: 'sqlbuilder-select', label: 'SELECT Basics' },
      { id: 'sqlbuilder-columns', label: 'Column Expressions' },
      { id: 'sqlbuilder-where', label: 'WHERE Helpers' },
      { id: 'sqlbuilder-joins', label: 'JOINs' },
      { id: 'sqlbuilder-join-subqueries', label: 'Join Subqueries' },
      { id: 'sqlbuilder-subqueries', label: 'Subqueries & CTEs' },
      { id: 'sqlbuilder-groupby', label: 'GROUP BY & HAVING' },
      { id: 'sqlbuilder-orderby', label: 'ORDER BY & Pagination' },
      { id: 'sqlbuilder-locking', label: 'Locking' },
      { id: 'sqlbuilder-set-ops', label: 'Set Operations' },
      { id: 'sqlbuilder-insert', label: 'INSERT' },
      { id: 'sqlbuilder-upsert', label: 'Upsert' },
      { id: 'sqlbuilder-update', label: 'UPDATE' },
      { id: 'sqlbuilder-delete', label: 'DELETE' },
      { id: 'sqlbuilder-expressions', label: 'Expressions' },
      { id: 'sqlbuilder-case', label: 'CASE / WHEN' },
      { id: 'sqlbuilder-window', label: 'Window Functions' },
      { id: 'sqlbuilder-returning', label: 'RETURNING' },
      { id: 'sqlbuilder-dialects', label: 'Dialects' },
      { id: 'sqlbuilder-request', label: 'Request Integration' },
      { id: 'sqlbuilder-clone', label: 'Clone & When' },
    ],
  },
  {
    id: 'dbx',
    label: 'dbx',
    children: [
      { id: 'dbx-setup', label: 'Setup' },
      { id: 'dbx-queryall', label: 'Querying Rows' },
      { id: 'dbx-queryone', label: 'Single Row' },
      { id: 'dbx-exec', label: 'Executing' },
      { id: 'dbx-sqlbuilder', label: 'sqlbuilder Integration' },
      { id: 'dbx-transactions', label: 'Transactions' },
      { id: 'dbx-tags', label: 'Struct Tags' },
    ],
  },
  {
    id: 'httpclient',
    label: 'httpclient',
    children: [
      { id: 'httpclient-creating', label: 'Creating a Client' },
      { id: 'httpclient-methods', label: 'HTTP Methods' },
      { id: 'httpclient-response', label: 'Response Handling' },
      { id: 'httpclient-builder', label: 'Request Builder' },
      { id: 'httpclient-headers', label: 'Default Headers' },
      { id: 'httpclient-circuit', label: 'Circuit Breaker' },
      { id: 'httpclient-mocking', label: 'Mocking' },
    ],
  },
  {
    id: 'apitest',
    label: 'apitest',
    children: [
      { id: 'apitest-requests', label: 'Building Requests' },
      { id: 'apitest-recording', label: 'Recording Responses' },
      { id: 'apitest-assertions', label: 'Assertions' },
      { id: 'apitest-decoding', label: 'Decoding' },
    ],
  },
];

// Collect all observable IDs (module + subsection)
const allIds: string[] = [];
const childToParent = new Map<string, string>();
sections.forEach((s) => {
  allIds.push(s.id);
  s.children?.forEach((c) => {
    allIds.push(c.id);
    childToParent.set(c.id, s.id);
  });
});

interface SidebarProps {
  open: boolean;
  onClose: () => void;
}

export default function Sidebar({ open, onClose }: SidebarProps) {
  const [activeId, setActiveId] = useState('getting-started');
  const [expandedId, setExpandedId] = useState<string | null>(null);

  // Derive which module is active (for highlighting parent items)
  const activeModuleId = childToParent.get(activeId) ?? activeId;

  // Auto-expand parent module when a subsection becomes active
  useEffect(() => {
    const parent = childToParent.get(activeId);
    if (parent) {
      setExpandedId(parent);
    } else if (sections.find((s) => s.id === activeId && s.children)) {
      setExpandedId(activeId);
    }
  }, [activeId]);

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        const visible = entries.filter((e) => e.isIntersecting);
        if (visible.length > 0) {
          setActiveId(visible[0].target.id);
        }
      },
      { rootMargin: '-80px 0px -60% 0px', threshold: 0 }
    );

    allIds.forEach((id) => {
      const el = document.getElementById(id);
      if (el) observer.observe(el);
    });

    return () => observer.disconnect();
  }, []);

  const handleModuleClick = useCallback(
    (id: string, hasChildren: boolean) => {
      if (hasChildren) {
        setExpandedId((prev) => (prev === id ? null : id));
      }
    },
    []
  );

  const handleSubItemClick = useCallback(() => {
    onClose();
  }, [onClose]);

  return (
    <>
      {open && (
        <div
          className="fixed inset-0 z-30 bg-black/50 md:hidden"
          onClick={onClose}
        />
      )}
      <aside
        className={`fixed top-16 left-0 z-40 h-[calc(100vh-4rem)] w-64 bg-bg-sidebar border-r border-border overflow-y-auto transition-transform duration-200 ${
          open ? 'translate-x-0' : '-translate-x-full'
        } md:translate-x-0`}
      >
        <nav className="py-4 px-3">
          <p className="text-xs font-semibold text-text-muted uppercase tracking-wider px-3 mb-2">
            Introduction
          </p>
          <a
            href="#getting-started"
            onClick={onClose}
            className={`block px-3 py-1.5 rounded-md text-sm mb-1 transition-colors ${
              activeId === 'getting-started'
                ? 'bg-primary/15 text-primary font-medium'
                : 'text-text-muted hover:text-text hover:bg-bg-card'
            }`}
          >
            Getting Started
          </a>

          <p className="text-xs font-semibold text-text-muted uppercase tracking-wider px-3 mt-4 mb-2">
            Modules
          </p>
          {sections.slice(1).map((section) => {
            const isExpanded = expandedId === section.id;
            const isActiveModule = activeModuleId === section.id;
            const hasChildren = !!section.children?.length;

            return (
              <div key={section.id} className="mb-0.5">
                <a
                  href={`#${section.id}`}
                  onClick={(e) => {
                    if (hasChildren) {
                      e.preventDefault();
                      handleModuleClick(section.id, true);
                      document.getElementById(section.id)?.scrollIntoView({ behavior: 'smooth' });
                    } else {
                      onClose();
                    }
                  }}
                  className={`flex items-center justify-between px-3 py-1.5 rounded-md text-sm font-mono transition-colors ${
                    isActiveModule
                      ? 'bg-primary/15 text-primary font-medium'
                      : 'text-text-muted hover:text-text hover:bg-bg-card'
                  }`}
                >
                  <span>{section.label}</span>
                  {hasChildren && (
                    <svg
                      className={`w-3.5 h-3.5 shrink-0 transition-transform duration-200 ${
                        isExpanded ? 'rotate-90' : ''
                      }`}
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M9 5l7 7-7 7"
                      />
                    </svg>
                  )}
                </a>

                {hasChildren && isExpanded && (
                  <div className="ml-3 mt-0.5 border-l border-border/50">
                    {section.children!.map((child) => (
                      <a
                        key={child.id}
                        href={`#${child.id}`}
                        onClick={handleSubItemClick}
                        className={`block pl-4 pr-3 py-1 text-xs transition-colors ${
                          activeId === child.id
                            ? 'text-primary font-medium'
                            : 'text-text-muted hover:text-text'
                        }`}
                      >
                        {child.label}
                      </a>
                    ))}
                  </div>
                )}
              </div>
            );
          })}
        </nav>
      </aside>
    </>
  );
}
