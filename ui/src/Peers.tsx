import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Form,
  FormGroup,
  FormHelperText,
  TextInput,
  Checkbox,
  Button,
  Alert,
  Tooltip,
  PageSection,
  Title,
  SearchInput,
  Toolbar,
  ToolbarContent,
  ToolbarItem,
  EmptyState,
  EmptyStateBody,
  EmptyStateHeader
} from '@patternfly/react-core';
import { InfoCircleIcon } from '@patternfly/react-icons';
import QRCode from 'qrcode';
import backend from './backend';
import { validatePeerForm } from './validators';

const Peers: React.FC = () => {
  const { t } = useTranslation();
  const [endpoint, setEndpoint] = useState('');
  const [allowedIPs, setAllowedIPs] = useState('');
  const [keepalive, setKeepalive] = useState('');
  const [preshared, setPreshared] = useState(false);
  const [enabled, setEnabled] = useState(true);
  const [publicKey, setPublicKey] = useState<string | null>(null);
  const [qr, setQr] = useState<string | null>(null);
  const [search, setSearch] = useState('');
  const [error, setError] = useState('');

  const handleAdd = async () => {
    const v = validatePeerForm(endpoint, allowedIPs, keepalive);
    if (v) {
      setError(t(v));
      return;
    }
    setError('');
    const peer = {
      endpoint,
      allowed_ips: allowedIPs.split(',').map((s) => s.trim()).filter(Boolean),
      persistent_keepalive: keepalive ? parseInt(keepalive, 10) : 0,
      preshared,
      enabled,
    };
    try {
      const res = await backend.addPeer('wg0', peer);
      setPublicKey(res.publicKey);
      // Trigger private key download
      const blob = new Blob([res.privateKey], { type: 'text/plain' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${res.publicKey}.key`;
      a.click();
      URL.revokeObjectURL(url);
      // Generate QR code
      const qrData = await QRCode.toDataURL(res.privateKey);
      setQr(qrData);
    } catch (e) {
      console.error(e);
    }
  };

  const handleCancel = () => {
    setEndpoint('');
    setAllowedIPs('');
    setKeepalive('');
    setPreshared(false);
    setEnabled(true);
    setPublicKey(null);
    setQr(null);
  };

  return (
    <PageSection>
      <Title headingLevel="h1">{t('peers.title')}</Title>
      <Form>
        {error && <Alert isInline variant="danger" title={error} role="alert" />}
        <FormGroup label={t('peers.endpoint')}>
          <TextInput
            id="endpoint"
            value={endpoint}
            onChange={(_, v) => setEndpoint(v)}
            aria-label={t('peers.endpoint')}
          />
          <FormHelperText>
            {t('peers.endpointHelp')}
          </FormHelperText>
        </FormGroup>
        <FormGroup label={t('peers.allowedIps')}>
          <TextInput
            id="allowed"
            value={allowedIPs}
            onChange={(_, v) => setAllowedIPs(v)}
            aria-label={t('peers.allowedIps')}
          />
          <FormHelperText>
            {t('peers.allowedIpsHelp')}
          </FormHelperText>
        </FormGroup>
        <FormGroup
          label={
            <>
              {t('peers.persistentKeepalive')}{' '}
              <Tooltip content={t('peers.keepaliveAdvanced')}>
                <InfoCircleIcon aria-label={t('peers.moreInfo')} />
              </Tooltip>
            </>
          }
        >
          <TextInput
            id="keepalive"
            value={keepalive}
            onChange={(_, v) => setKeepalive(v)}
            aria-label={t('peers.persistentKeepalive')}
          />
          <FormHelperText>
            {t('peers.keepaliveHelp')}
          </FormHelperText>
        </FormGroup>
        <FormGroup
          label={
            <>
              {t('peers.presharedKey')}{' '}
              <Tooltip content={t('peers.presharedAdvanced')}>
                <InfoCircleIcon aria-label={t('peers.moreInfo')} />
              </Tooltip>
            </>
          }
        >
          <Checkbox
            id="psk"
            isChecked={preshared}
            onChange={(_, v) => setPreshared(v)}
            label={t('peers.generatePreshared')}
            aria-label={t('peers.presharedKey')}
          />
          <FormHelperText>
            {t('peers.presharedHelp')}
          </FormHelperText>
        </FormGroup>
        <FormGroup label={t('peers.enabled')}>
          <Checkbox
            id="enabled"
            isChecked={enabled}
            onChange={(_, v) => setEnabled(v)}
            label={t('peers.peerEnabled')}
            aria-label={t('peers.peerEnabled')}
          />
          <FormHelperText>
            {t('peers.peerEnabled')}
          </FormHelperText>
        </FormGroup>
        <Button variant="primary" onClick={handleAdd}>{t('peers.addPeer')}</Button>{' '}
        <Button variant="secondary" onClick={handleCancel}>{t('peers.cancel')}</Button>
      </Form>
      {publicKey && <Alert isInline variant="info" title={t('peers.publicKey', { publicKey })} isLiveRegion />}
      {qr && <img src={qr} alt={t('peers.peerQrAlt')} />}
      <Toolbar>
        <ToolbarContent>
          <ToolbarItem>
            <SearchInput
              aria-label={t('peers.searchPeers')}
              value={search}
              onChange={(_, v) => setSearch(v)}
            />
          </ToolbarItem>
        </ToolbarContent>
      </Toolbar>
      <EmptyState variant="sm">
        <EmptyStateHeader titleText={t('peers.noPeers')} headingLevel="h2" />
        <EmptyStateBody>{t('peers.addPeerToGetStarted')}</EmptyStateBody>
      </EmptyState>
    </PageSection>
  );
};

export default Peers;
