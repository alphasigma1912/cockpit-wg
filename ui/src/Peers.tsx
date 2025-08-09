import React, { useState } from 'react';
import {
  Form,
  FormGroup,
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

const Peers: React.FC = () => {
  const [endpoint, setEndpoint] = useState('');
  const [allowedIPs, setAllowedIPs] = useState('');
  const [keepalive, setKeepalive] = useState('');
  const [preshared, setPreshared] = useState(false);
  const [enabled, setEnabled] = useState(true);
  const [publicKey, setPublicKey] = useState<string | null>(null);
  const [qr, setQr] = useState<string | null>(null);
  const [search, setSearch] = useState('');

  const handleAdd = async () => {
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
      <Title headingLevel="h1">Peers</Title>
      <Form>
        <FormGroup label="Endpoint" fieldId="endpoint" helperText="Peer's endpoint in host:port format">
          <TextInput
            id="endpoint"
            value={endpoint}
            onChange={(_, v) => setEndpoint(v)}
            aria-label="Endpoint"
          />
        </FormGroup>
        <FormGroup label="Allowed IPs" fieldId="allowed" helperText="Comma-separated list of allowed IPs">
          <TextInput
            id="allowed"
            value={allowedIPs}
            onChange={(_, v) => setAllowedIPs(v)}
            aria-label="Allowed IPs"
          />
        </FormGroup>
        <FormGroup
          label="Persistent keepalive"
          fieldId="keepalive"
          helperText="Seconds between keepalive packets"
          labelIcon={
            <Tooltip content="Advanced: helps maintain NAT mappings">
              <InfoCircleIcon />
            </Tooltip>
          }
        >
          <TextInput
            id="keepalive"
            value={keepalive}
            onChange={(_, v) => setKeepalive(v)}
            aria-label="Persistent keepalive"
          />
        </FormGroup>
        <FormGroup
          label="Preshared key"
          fieldId="psk"
          helperText="Generate an optional preshared key"
          labelIcon={
            <Tooltip content="Advanced: adds symmetric encryption">
              <InfoCircleIcon />
            </Tooltip>
          }
        >
          <Checkbox
            id="psk"
            isChecked={preshared}
            onChange={(_, v) => setPreshared(v)}
            label="Generate preshared key"
            aria-label="Preshared key"
          />
        </FormGroup>
        <FormGroup label="Enabled" fieldId="enabled" helperText="Peer is enabled">
          <Checkbox
            id="enabled"
            isChecked={enabled}
            onChange={(_, v) => setEnabled(v)}
            label="Peer enabled"
            aria-label="Peer enabled"
          />
        </FormGroup>
        <Button variant="primary" onClick={handleAdd}>Add peer</Button>{' '}
        <Button variant="secondary" onClick={handleCancel}>Cancel</Button>
      </Form>
      {publicKey && <Alert isInline variant="info" title={`Public key: ${publicKey}`} />}
      {qr && <img src={qr} alt="peer qr" />}
      <Toolbar>
        <ToolbarContent>
          <ToolbarItem>
            <SearchInput
              aria-label="Search peers"
              value={search}
              onChange={(_, v) => setSearch(v)}
            />
          </ToolbarItem>
        </ToolbarContent>
      </Toolbar>
      <EmptyState variant="sm">
        <EmptyStateHeader titleText="No peers" headingLevel="h2" />
        <EmptyStateBody>Add a peer to get started.</EmptyStateBody>
      </EmptyState>
    </PageSection>
  );
};

export default Peers;
