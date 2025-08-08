import React, { useState } from 'react';
import { Form, FormGroup, TextInput, Checkbox, Button, Alert } from '@patternfly/react-core';
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

  return (
    <div>
      <Form>
        <FormGroup label="Endpoint" fieldId="endpoint">
          <TextInput id="endpoint" value={endpoint} onChange={(_, v) => setEndpoint(v)} />
        </FormGroup>
        <FormGroup label="AllowedIPs" fieldId="allowed">
          <TextInput id="allowed" value={allowedIPs} onChange={(_, v) => setAllowedIPs(v)} />
        </FormGroup>
        <FormGroup label="PersistentKeepalive" fieldId="keepalive">
          <TextInput id="keepalive" value={keepalive} onChange={(_, v) => setKeepalive(v)} />
        </FormGroup>
        <FormGroup label="PresharedKey" fieldId="psk">
          <Checkbox id="psk" isChecked={preshared} onChange={(_, v) => setPreshared(v)} label="Generate preshared key" />
        </FormGroup>
        <FormGroup label="Enabled" fieldId="enabled">
          <Checkbox id="enabled" isChecked={enabled} onChange={(_, v) => setEnabled(v)} label="Peer enabled" />
        </FormGroup>
        <Button variant="primary" onClick={handleAdd}>Add peer</Button>
      </Form>
      {publicKey && <Alert isInline variant="info" title={`Public key: ${publicKey}`} />}
      {qr && <img src={qr} alt="peer qr" />}
    </div>
  );
};

export default Peers;
