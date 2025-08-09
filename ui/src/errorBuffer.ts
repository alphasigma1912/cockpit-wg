import { BackendError } from './errorCodes';

export interface LoggedError extends BackendError {
  timestamp: number;
}

const buffer: LoggedError[] = [];

export function logError(err: BackendError) {
  buffer.push({ ...err, timestamp: Date.now() });
  if (buffer.length > 50) buffer.shift();
}

export function getErrorLog(): LoggedError[] {
  return buffer;
}

export function clearErrorLog() {
  buffer.length = 0;
}
