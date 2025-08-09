import { describe, it, expect } from 'vitest';
import { logError, getErrorLog, clearErrorLog, onError } from './errorBuffer';

import { BackendError } from './errorCodes';

describe('error buffer', () => {
  it('stores errors with timestamp', () => {
    clearErrorLog();
    const err: BackendError = { code: 1, message: 'boom' };
    logError(err);
    const log = getErrorLog();
    expect(log.length).toBe(1);
    expect(log[0].message).toBe('boom');
    expect(typeof log[0].timestamp).toBe('number');
  });

  it('notifies listeners', () => {
    clearErrorLog();
    const err: BackendError = { code: 2, message: 'nope' };
    let notified = false;
    const off = onError(() => {
      notified = true;
    });
    logError(err);
    off();
    expect(notified).toBe(true);
  });
});
