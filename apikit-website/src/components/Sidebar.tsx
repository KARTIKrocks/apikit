import { useEffect, useState } from 'react';

const sections = [
  { id: 'getting-started', label: 'Getting Started' },
  { id: 'errors', label: 'errors' },
  { id: 'router', label: 'router' },
  { id: 'middleware', label: 'middleware' },
  { id: 'request', label: 'request' },
  { id: 'response', label: 'response' },
  { id: 'server', label: 'server' },
  { id: 'config', label: 'config' },
  { id: 'health', label: 'health' },
  { id: 'sqlbuilder', label: 'sqlbuilder' },
  { id: 'dbx', label: 'dbx' },
  { id: 'httpclient', label: 'httpclient' },
  { id: 'apitest', label: 'apitest' },
];

interface SidebarProps {
  open: boolean;
  onClose: () => void;
}

export default function Sidebar({ open, onClose }: SidebarProps) {
  const [activeId, setActiveId] = useState('getting-started');

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

    sections.forEach(({ id }) => {
      const el = document.getElementById(id);
      if (el) observer.observe(el);
    });

    return () => observer.disconnect();
  }, []);

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
          {sections.slice(1).map(({ id, label }) => (
            <a
              key={id}
              href={`#${id}`}
              onClick={onClose}
              className={`block px-3 py-1.5 rounded-md text-sm mb-0.5 font-mono transition-colors ${
                activeId === id
                  ? 'bg-primary/15 text-primary font-medium'
                  : 'text-text-muted hover:text-text hover:bg-bg-card'
              }`}
            >
              {label}
            </a>
          ))}
        </nav>
      </aside>
    </>
  );
}
