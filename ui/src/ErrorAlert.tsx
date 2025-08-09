import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Alert, AlertActionLink } from '@patternfly/react-core';
import { BackendError, errorMessages } from './errorCodes';

interface Props {
  error: BackendError;
}

const ErrorAlert: React.FC<Props> = ({ error }) => {
  const { t } = useTranslation();
  const [show, setShow] = useState(false);
  const title = errorMessages[error.code]
    ? t(errorMessages[error.code])
    : error.message;
  return (
    <Alert
      isInline
      variant="danger"
      title={title}
      actionLinks=
        {error.details && (
          <AlertActionLink onClick={() => setShow(!show)}>
            {show ? t('hideDetails') : t('showDetails')}
          </AlertActionLink>
        )}
    >
      {show && error.details && <pre>{error.details}</pre>}
    </Alert>
  );
};

export default ErrorAlert;
