import React, { useEffect, useState } from 'react';
import { Button, Alert, FormGroup } from '@patternfly/react-core';
import backend from './backend';

const InterfaceControls: React.FC = () => {
  const [interfaces, setInterfaces] = useState<string[]>([]);
  const [selected, setSelected] = useState<string>('');
  const [status, setStatus] = useState('');
  const [lastChange, setLastChange] = useState('');
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');

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

  return (
    <div>
      {error && <Alert isInline variant="danger" title={error} />}
      <FormGroup label="Interface" fieldId="iface-select">
        <select id="iface-select" value={selected} onChange={(e) => setSelected(e.target.value)}>
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
      <Button variant="primary" onClick={doAction('up')} isDisabled={!selected}>
        Up
      </Button>{' '}
      <Button variant="secondary" onClick={doAction('down')} isDisabled={!selected}>
        Down
      </Button>{' '}
      <Button variant="secondary" onClick={doAction('restart')} isDisabled={!selected}>
        Restart
      </Button>
    </div>
  );
};

export default InterfaceControls;
