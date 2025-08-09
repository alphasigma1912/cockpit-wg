import React, { useEffect, useState } from 'react';
import {
  Button,
  Alert,
  FormGroup,
  Modal,
  PageSection,
  Title
} from '@patternfly/react-core';
import backend from './backend';
import MetricsGraph from './MetricsGraph';

const InterfaceControls: React.FC = () => {
  const [interfaces, setInterfaces] = useState<string[]>([]);
  const [selected, setSelected] = useState<string>('');
  const [status, setStatus] = useState('');
  const [lastChange, setLastChange] = useState('');
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');
  const [confirmDown, setConfirmDown] = useState(false);
  const [metrics, setMetrics] = useState<{ timestamps: number[]; rx: number[]; tx: number[] }>({
    timestamps: [],
    rx: [],
    tx: [],
  });

  useEffect(() => {
    backend
      .listInterfaces()
      .then((res) => {
        const list: string[] = res.interfaces || [];
        setInterfaces(list);
        if (list.length > 0) {
          setSelected(list[0]);
        }
      })
      .catch((e) => setError(String(e)));
  }, []);

  const refreshStatus = () => {
    if (!selected) return;
    backend
      .getInterfaceStatus(selected)
      .then((res) => {
        setStatus(res.status);
        setLastChange(res.last_change);
        setMessage(res.message || '');
      })
      .catch((e) => setError(String(e)));
  };

  useEffect(() => {
    refreshStatus();
    let cancelled = false;
    const fetchMetrics = () => {
      if (!selected) return;
      backend
        .getMetrics(selected)
        .then((res) => {
          if (!cancelled) setMetrics(res);
        })
        .catch((e) => setError(String(e)));
    };
    fetchMetrics();
    const id = setInterval(fetchMetrics, 2000);
    return () => {
      cancelled = true;
      clearInterval(id);
    };
  }, [selected]);

  const scheduleRefresh = () => {
    refreshStatus();
    const id = setInterval(refreshStatus, 2000);
    setTimeout(() => clearInterval(id), 10000);
  };

  const doAction = (action: 'up' | 'down' | 'restart') => () => {
    setError('');
    let p: Promise<any>;
    if (action === 'up') p = backend.upInterface(selected);
    else if (action === 'down') p = backend.downInterface(selected);
    else p = backend.restartInterface(selected);
    p.catch((e) => setError(String(e))).finally(scheduleRefresh);
  };

  const confirmDownAction = () => {
    setConfirmDown(false);
    doAction('down')();
  };

  return (
    <PageSection>
      <Title headingLevel="h1">Interfaces</Title>
      {error && <Alert isInline variant="danger" title={error} />}
      <FormGroup label="Interface" fieldId="iface-select" helperText="Select the interface to manage">
        <select
          id="iface-select"
          value={selected}
          onChange={(e) => setSelected(e.target.value)}
          aria-label="Interface selector"
        >
          {interfaces.map((n) => (
            <option key={n} value={n}>
              {n}
            </option>
          ))}
        </select>
      </FormGroup>
      <p>Status: {status}</p>
      <p>Last change: {lastChange}</p>
      {message && <pre>{message}</pre>}
      <MetricsGraph times={metrics.timestamps} rx={metrics.rx} tx={metrics.tx} />
      <Button variant="primary" onClick={doAction('up')} isDisabled={!selected}>
        Up
      </Button>{' '}
      <Button variant="secondary" onClick={() => setConfirmDown(true)} isDisabled={!selected}>
        Down
      </Button>{' '}
      <Button variant="secondary" onClick={doAction('restart')} isDisabled={!selected}>
        Restart
      </Button>
      <Modal
        title="Bring interface down?"
        isOpen={confirmDown}
        onClose={() => setConfirmDown(false)}
        actions={[
          <Button key="confirm" variant="danger" onClick={confirmDownAction}>
            Down
          </Button>,
          <Button key="cancel" variant="secondary" onClick={() => setConfirmDown(false)}>
            Cancel
          </Button>,
        ]}
      >
        Bringing the interface down will disconnect all peers.
      </Modal>
    </PageSection>
  );
};

export default InterfaceControls;
