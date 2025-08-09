import React from 'react';
import {
  PageSection,
  EmptyState,
  EmptyStateBody,
  EmptyStateHeader
} from '@patternfly/react-core';
import { useTranslation } from 'react-i18next';

const Exchange: React.FC = () => {
  const { t } = useTranslation();
  return (
    <PageSection>
      <EmptyState variant="sm">
        <EmptyStateHeader titleText={t('exchange.title')} headingLevel="h1" />
        <EmptyStateBody>{t('exchange.noData')}</EmptyStateBody>
      </EmptyState>
    </PageSection>
  );
};

export default Exchange;
