import React from 'react';
import { PageSection, Title, Skeleton } from '@patternfly/react-core';

const Overview: React.FC = () => (
  <PageSection>
    <Title headingLevel="h1">Overview</Title>
    <Skeleton width="50%" />
    <Skeleton width="70%" />
  </PageSection>
);

export default Overview;
