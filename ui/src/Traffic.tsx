import React from 'react';
import {
  PageSection,
  EmptyState,
  EmptyStateBody,
  EmptyStateHeader
} from '@patternfly/react-core';

const Traffic: React.FC = () => (
  <PageSection>
    <EmptyState variant="sm">
      <EmptyStateHeader titleText="Traffic" headingLevel="h1" />
      <EmptyStateBody>No traffic data available.</EmptyStateBody>
    </EmptyState>
  </PageSection>
);

export default Traffic;
