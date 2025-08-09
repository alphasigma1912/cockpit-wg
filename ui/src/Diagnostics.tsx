import React, { useState } from 'react';
import { PageSection, Title, Button, Spinner } from '@patternfly/react-core';
import { useTranslation } from 'react-i18next';
import backend from './backend';

const Diagnostics: React.FC = () => {
  const { t } = useTranslation();
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
        setReport(t('diagnostics.error', { error: err }));
      })
      .finally(() => setRunning(false));
  };

  return (
    <PageSection>
      <Title headingLevel="h1">{t('diagnostics.title')}</Title>
      <Button onClick={handleRun} isDisabled={running}>
        {running ? t('diagnostics.running') : t('diagnostics.runSelfTest')}
      </Button>
      {running && <Spinner />}
      {report && <pre>{report}</pre>}
      {details && <pre>{JSON.stringify(details, null, 2)}</pre>}
    </PageSection>
  );
};

export default Diagnostics;
