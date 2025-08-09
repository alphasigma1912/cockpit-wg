import React from 'react';
import ReactDOM from 'react-dom/client';
import '@patternfly/react-core/dist/styles/base.css';
import './preloadFonts';
import './theme';
import './i18n';
import './tokens.css';
import './accessibility.css';
import App from './App';
import GlobalErrorBoundary from './GlobalErrorBoundary';
import { markMounted } from './globalErrorHandlers';

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <React.StrictMode>
    <GlobalErrorBoundary>
      <App />
    </GlobalErrorBoundary>
  </React.StrictMode>
);
markMounted();
