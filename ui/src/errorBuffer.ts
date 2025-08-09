import { BackendError } from './errorCodes';

export interface LoggedError extends BackendError {
  timestamp: number;
}

const buffer: LoggedError[] = [];
const listeners: ((err: LoggedError) => void)[] = [];

export function logError(err: BackendError) {
  const logged = { ...err, timestamp: Date.now() };
  buffer.push(logged);
  if (buffer.length > 50) buffer.shift();
  for (const l of listeners) l(logged);
}

export function getErrorLog(): LoggedError[] {
  return buffer;
}

export function clearErrorLog() {
  buffer.length = 0;
}

export function onError(cb: (err: LoggedError) => void) {
  listeners.push(cb);
  return () => {
    const idx = listeners.indexOf(cb);
    if (idx >= 0) listeners.splice(idx, 1);
  };
}
