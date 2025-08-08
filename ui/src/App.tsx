import React, { useEffect, useState } from 'react';
import { Page, PageSection, Title, Button, Spinner, Alert } from '@patternfly/react-core';
import backend from './backend';
import Peers from './Peers';
import InterfaceControls from './InterfaceControls';
import Diagnostics from './Diagnostics';

const App: React.FC = () => {
  const [loading, setLoading] = useState(true);
  const [needsSetup, setNeedsSetup] = useState(false);
  const [installing, setInstalling] = useState(false);
  const [summary, setSummary] = useState<string | null>(null);
  const [page, setPage] = useState<'main' | 'diagnostics'>('main');

  useEffect(() => {
    backend
      .checkPrereqs()
      .then((res) => {
        const ok = res.kernel && res.tools && res.systemd;
        setNeedsSetup(!ok);
        setLoading(false);
      })
      .catch(() => setLoading(false));
  }, []);

  const handleInstall = () => {
    setInstalling(true);
    setSummary(null);
    backend
      .installPackages()
      .then(() => {
        setSummary('Installation complete');
        setInstalling(false);
      })
      .catch((err) => {
        setSummary(`Installation failed: ${err}`);
        setInstalling(false);
      });
  };

  if (loading) {
    return (
      <Page>
        <PageSection>
          <Spinner />
        </PageSection>
      </Page>
    );
  }

  if (needsSetup) {
    return (
      <Page>
        <PageSection>
          <Title headingLevel="h1">WireGuard setup</Title>
          <p>WireGuard is not installed. Install WireGuard and enable systemd service templates.</p>
          <Button variant="primary" onClick={handleInstall} isDisabled={installing}>
            {installing ? 'Installing…' : 'Install WireGuard'}
          </Button>
          {summary && (
            <Alert isInline variant="info" title={summary} />
          )}
        </PageSection>
      </Page>
    );
  }

  return (
    <Page>
      <PageSection>
        <Title headingLevel="h1">Cockpit WireGuard</Title>
        <Button variant="link" onClick={() => setPage('main')}>Interfaces</Button>
        <Button variant="link" onClick={() => setPage('diagnostics')}>Diagnostics</Button>
        {page === 'main' ? (
          <>
            <InterfaceControls />
            <Peers />
          </>
        ) : (
          <Diagnostics />
        )}
      </PageSection>
    </Page>
  );
};

export default App;
