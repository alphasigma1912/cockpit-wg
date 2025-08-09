import React from 'react';
import {
  PageSection,
  EmptyState,
  EmptyStateBody,
  EmptyStateHeader
} from '@patternfly/react-core';

const Exchange: React.FC = () => (
  <PageSection>
    <EmptyState variant="sm">
      <EmptyStateHeader titleText="Exchange" headingLevel="h1" />
      <EmptyStateBody>No exchange data available.</EmptyStateBody>
    </EmptyState>
  </PageSection>
);

export default Exchange;
