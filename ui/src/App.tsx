import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
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
import ErrorToasts from './ErrorToasts';

const App: React.FC = () => {
  const { t } = useTranslation();
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
        setSummary(t('installationComplete'));
        return backend.checkPrereqs();
      })
      .then((res) => {
        const ok = res.kernel && res.tools && res.systemd;
        setNeedsSetup(!ok);
        setInstalling(false);
      })
      .catch((err) => {
        setSummary(t('installationFailed', { error: err }));
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
          <Title headingLevel="h1">{t('wireguardSetup')}</Title>
          <p>{t('wireguardNotInstalled')}</p>
          <Button variant="primary" onClick={handleInstall} isDisabled={installing}>
            {installing ? t('installing') : t('installWireGuard')}
          </Button>
          {summary && (
            <Alert isInline variant="info" title={summary} isLiveRegion />
          )}
        </PageSection>
      </Page>
    );
  }

  const nav = (
    <Nav onSelect={(e, itemId) => setPage(itemId as any)} aria-label={t('nav.primary')}>
      <NavList>
        <NavItem itemId="overview" isActive={page === 'overview'}>
          {t('nav.overview')}
        </NavItem>
        <NavItem itemId="interfaces" isActive={page === 'interfaces'}>
          {t('nav.interfaces')}
        </NavItem>
        <NavItem itemId="peers" isActive={page === 'peers'}>
          {t('nav.peers')}
        </NavItem>
        <NavItem itemId="traffic" isActive={page === 'traffic'}>
          {t('nav.traffic')}
        </NavItem>
        <NavItem itemId="diagnostics" isActive={page === 'diagnostics'}>
          {t('nav.diagnostics')}
        </NavItem>
        <NavItem itemId="exchange" isActive={page === 'exchange'}>
          {t('nav.exchange')}
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
    <>
      <ErrorToasts />
      <Page sidebar={sidebar} isManagedSidebar>
        {page === 'overview' && <Overview />}
        {page === 'interfaces' && <InterfaceControls />}
        {page === 'peers' && <Peers />}
        {page === 'traffic' && <Traffic />}
        {page === 'diagnostics' && <Diagnostics />}
        {page === 'exchange' && <Exchange />}
      </Page>
    </>
  );
};

export default App;
