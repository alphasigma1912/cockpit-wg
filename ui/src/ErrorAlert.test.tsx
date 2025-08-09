import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import ErrorAlert from './ErrorAlert';
import { BackendError, CodeValidationFailed } from './errorCodes';
import './i18n';

describe('ErrorAlert', () => {
  it('shows message and toggles details', () => {
    const err: BackendError = {
      code: CodeValidationFailed,
      message: 'x',
      details: 'stack',
    };
    render(<ErrorAlert error={err} />);
    expect(screen.getByText('Configuration validation failed')).toBeInTheDocument();
    fireEvent.click(screen.getByText('Show details'));
    expect(screen.getByText('stack')).toBeInTheDocument();
  });
});
