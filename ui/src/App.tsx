import React, { useEffect, useState } from 'react';
import {
  Page,
  PageSection,
  Title,
  Button,
  Spinner,
  Alert,
  Nav,
  NavList,
  NavItem,
  PageSidebar,
  PageSidebarBody
} from '@patternfly/react-core';
import backend from './backend';
import Peers from './Peers';
import InterfaceControls from './InterfaceControls';
import Diagnostics from './Diagnostics';
import Overview from './Overview';
import Traffic from './Traffic';
import Exchange from './Exchange';

const App: React.FC = () => {
  const [loading, setLoading] = useState(true);
  const [needsSetup, setNeedsSetup] = useState(false);
  const [installing, setInstalling] = useState(false);
  const [summary, setSummary] = useState<string | null>(null);
  const [page, setPage] = useState<
    'overview' | 'interfaces' | 'peers' | 'traffic' | 'diagnostics' | 'exchange'
  >('overview');

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

  const nav = (
    <Nav onSelect={(e, itemId) => setPage(itemId as any)} aria-label="Primary navigation">
      <NavList>
        <NavItem itemId="overview" isActive={page === 'overview'}>
          Overview
        </NavItem>
        <NavItem itemId="interfaces" isActive={page === 'interfaces'}>
          Interfaces
        </NavItem>
        <NavItem itemId="peers" isActive={page === 'peers'}>
          Peers
        </NavItem>
        <NavItem itemId="traffic" isActive={page === 'traffic'}>
          Traffic
        </NavItem>
        <NavItem itemId="diagnostics" isActive={page === 'diagnostics'}>
          Diagnostics
        </NavItem>
        <NavItem itemId="exchange" isActive={page === 'exchange'}>
          Exchange
        </NavItem>
      </NavList>
    </Nav>
  );

  const sidebar = (
    <PageSidebar>
      <PageSidebarBody>{nav}</PageSidebarBody>
    </PageSidebar>
  );

  return (
    <Page sidebar={sidebar} isManagedSidebar>
      {page === 'overview' && <Overview />}
      {page === 'interfaces' && <InterfaceControls />}
      {page === 'peers' && <Peers />}
      {page === 'traffic' && <Traffic />}
      {page === 'diagnostics' && <Diagnostics />}
      {page === 'exchange' && <Exchange />}
    </Page>
  );
};

export default App;
