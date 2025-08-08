import React, { useState } from 'react';
import { PageSection, Title, Button, Spinner } from '@patternfly/react-core';
import backend from './backend';

const Diagnostics: React.FC = () => {
  const [running, setRunning] = useState(false);
  const [report, setReport] = useState('');
  const [details, setDetails] = useState<any>(null);

  const handleRun = () => {
    setRunning(true);
    setReport('');
    setDetails(null);
    backend
      .runSelfTest()
      .then((res) => {
        setReport(res.report);
        setDetails(res.details);
      })
      .catch((err) => {
        setReport(`Error: ${err}`);
      })
      .finally(() => setRunning(false));
  };

  return (
    <PageSection>
      <Title headingLevel="h1">Diagnostics</Title>
      <Button onClick={handleRun} isDisabled={running}>
        {running ? 'Running…' : 'Run self-test'}
      </Button>
      {running && <Spinner />}
      {report && <pre>{report}</pre>}
      {details && <pre>{JSON.stringify(details, null, 2)}</pre>}
    </PageSection>
  );
};

export default Diagnostics;
