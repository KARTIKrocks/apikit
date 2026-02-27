import { useState } from 'react';
import Navbar from './components/Navbar';
import Sidebar from './components/Sidebar';
import Hero from './components/Hero';
import GettingStarted from './content/getting-started';
import ErrorsDocs from './content/errors';
import RouterDocs from './content/router';
import MiddlewareDocs from './content/middleware';
import RequestDocs from './content/request';
import ResponseDocs from './content/response';
import ServerDocs from './content/server';
import ConfigDocs from './content/config';
import HealthDocs from './content/health';
import SqlbuilderDocs from './content/sqlbuilder';
import DbxDocs from './content/dbx';
import HttpclientDocs from './content/httpclient';
import ApitestDocs from './content/apitest';

export default function App() {
  const [menuOpen, setMenuOpen] = useState(false);

  return (
    <div className="min-h-screen">
      <Navbar onMenuToggle={() => setMenuOpen((o) => !o)} menuOpen={menuOpen} />
      <Sidebar open={menuOpen} onClose={() => setMenuOpen(false)} />

      <main className="pt-16 md:pl-64">
        <div className="max-w-4xl mx-auto px-4 md:px-8 pb-20">
          <Hero />
          <GettingStarted />
          <ErrorsDocs />
          <RouterDocs />
          <MiddlewareDocs />
          <RequestDocs />
          <ResponseDocs />
          <ServerDocs />
          <ConfigDocs />
          <HealthDocs />
          <SqlbuilderDocs />
          <DbxDocs />
          <HttpclientDocs />
          <ApitestDocs />

          <footer className="py-10 text-center text-sm text-text-muted border-t border-border mt-10">
            <p>
              apikit is open source under the{' '}
              <a
                href="https://github.com/KARTIKrocks/apikit/blob/main/LICENSE"
                className="text-primary hover:underline"
                target="_blank"
                rel="noopener noreferrer"
              >
                MIT License
              </a>
            </p>
          </footer>
        </div>
      </main>
    </div>
  );
}
