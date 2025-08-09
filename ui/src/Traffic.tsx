import React from 'react';
import {
  PageSection,
  EmptyState,
  EmptyStateBody,
  EmptyStateHeader
} from '@patternfly/react-core';
import { useTranslation } from 'react-i18next';

const Traffic: React.FC = () => {
  const { t } = useTranslation();
  return (
    <PageSection>
      <EmptyState variant="sm">
        <EmptyStateHeader titleText={t('traffic.title')} headingLevel="h1" />
        <EmptyStateBody>{t('traffic.noData')}</EmptyStateBody>
      </EmptyState>
    </PageSection>
  );
};

export default Traffic;
