import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import GlobalErrorBoundary from './GlobalErrorBoundary';
import { logger, clearLogEvents } from './logger';
import './i18n';

function Bomb() {
  throw new Error('boom');
}

describe('GlobalErrorBoundary', () => {
  it('shows overlay and copies details', () => {
    clearLogEvents();
    logger.info('UI', 'hello', { secret: 'x' });
    Object.assign(navigator, {
      clipboard: { writeText: vi.fn().mockResolvedValue(undefined) },
    });
    render(
      <GlobalErrorBoundary>
        <Bomb />
      </GlobalErrorBoundary>,
    );
    expect(screen.getByText('Reload plugin')).toBeInTheDocument();
    fireEvent.click(screen.getByText('Copy error details'));
    expect(navigator.clipboard.writeText).toHaveBeenCalled();
    const data = JSON.parse((navigator.clipboard.writeText as any).mock.calls[0][0]);
    expect(data.logs[0].args[1]).toBe('[REDACTED]');
  });
});
