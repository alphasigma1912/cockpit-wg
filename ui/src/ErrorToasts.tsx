import React, { useEffect, useState } from 'react';
import { Alert, AlertGroup } from '@patternfly/react-core';
import { useTranslation } from 'react-i18next';
import { LoggedError, onError } from './errorBuffer';

interface Toast extends LoggedError {
  count: number;
}

const ErrorToasts: React.FC = () => {
  const { t } = useTranslation();
  const [errors, setErrors] = useState<Toast[]>([]);

  useEffect(() => {
    return onError((err) => {
      setErrors((prev) => {
        const existing = prev.find((e) => e.message === err.message);
        if (existing) {
          return prev.map((e) =>
            e.message === err.message ? { ...e, count: e.count + 1 } : e,
          );
        }
        return [...prev, { ...err, count: 1 }];
      });
    });
  }, []);

  return (
    <AlertGroup isToast isLiveRegion>
      {errors.map((e) => (
        <Alert
          key={e.timestamp}
          variant="danger"
          title={`${t('unexpectedError')} (${e.code})${e.count > 1 ? ` (x${e.count})` : ''}`}
        >
          {e.trace && <p>{t('traceId', { trace: e.trace })}</p>}
        </Alert>
      ))}
    </AlertGroup>
  );
};

export default ErrorToasts;
