import React from 'react';
import { Button } from '@patternfly/react-core';
import { useTranslation } from 'react-i18next';
import logger, { getLogEvents } from './logger';

interface ErrorBoundaryState {
  error: Error | null;
  info: React.ErrorInfo | null;
}

function ErrorOverlay({ error, info }: { error: Error; info: React.ErrorInfo | null }) {
  const { t } = useTranslation();
  const reload = () => window.location.reload();
  const safeMode = () => {
    const url = new URL(window.location.href);
    url.searchParams.set('safeMode', '1');
    window.location.href = url.toString();
  };
  const copy = () => {
    const details = {
      message: error.message,
      stack: info?.componentStack ?? '',
      logs: getLogEvents(20),
    };
    void navigator.clipboard.writeText(JSON.stringify(details, null, 2));
  };
  return (
    <div
      style={{
        position: 'fixed',
        top: 0,
        left: 0,
        width: '100%',
        height: '100%',
        background: 'rgba(0,0,0,0.6)',
        zIndex: 1000,
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      <div style={{ background: 'white', padding: '1rem', borderRadius: '4px' }}>
        <p>{t('unexpectedError')}</p>
        <Button variant="primary" onClick={reload}>
          {t('reloadPlugin')}
        </Button>{' '}
        <Button variant="secondary" onClick={safeMode}>
          {t('enterSafeMode')}
        </Button>{' '}
        <Button variant="link" onClick={copy}>
          {t('copyErrorDetails')}
        </Button>
      </div>
    </div>
  );
}

class GlobalErrorBoundary extends React.Component<React.PropsWithChildren, ErrorBoundaryState> {
  state: ErrorBoundaryState = { error: null, info: null };

  static getDerivedStateFromError(error: Error): Partial<ErrorBoundaryState> {
    return { error };
  }

  componentDidCatch(error: Error, info: React.ErrorInfo) {
    this.setState({ info });
    logger.error('UI', error.message, info.componentStack);
  }

  render() {
    if (this.state.error) {
      return (
        <>
          <ErrorOverlay error={this.state.error} info={this.state.info} />
        </>
      );
    }
    return this.props.children as React.ReactElement;
  }
}

export default GlobalErrorBoundary;
