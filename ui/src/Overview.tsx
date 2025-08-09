import React from 'react';
import { PageSection, Title, Skeleton } from '@patternfly/react-core';
import { useTranslation } from 'react-i18next';

const Overview: React.FC = () => {
  const { t } = useTranslation();
  return (
    <PageSection>
      <Title headingLevel="h1">{t('overview.title')}</Title>
      <Skeleton width="50%" />
      <Skeleton width="70%" />
    </PageSection>
  );
};

export default Overview;
