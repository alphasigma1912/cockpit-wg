import { logError } from './errorBuffer';
import logger, { redactString } from './logger';
import { BackendError, CodeUnhandled } from './errorCodes';

let mounted = false;
export function markMounted() {
  mounted = true;
}

const recent = new Map<string, { count: number; start: number }>();
function shouldLog(message: string): boolean {
  const now = Date.now();
  const entry = recent.get(message) || { count: 0, start: now };
  if (now - entry.start > 10000) {
    entry.count = 0;
    entry.start = now;
  }
  entry.count++;
  recent.set(message, entry);
  return entry.count <= 5;
}

function handle(message: string) {
  const clean = redactString(message.replace(/^Uncaught (in promise )?/, ''));
  if (!shouldLog(clean)) return;
  const trace = globalThis.crypto?.randomUUID
    ? globalThis.crypto.randomUUID()
    : Math.random().toString(36).slice(2);
  const err: BackendError = { code: CodeUnhandled, message: clean, trace };
  logger.error('UI', clean, trace);
  logError(err);
  if (!mounted && typeof (window as any).__renderSafeMode === 'function') {
    (window as any).__renderSafeMode('RUNTIME_ERR', clean);
  }
}

export function registerGlobalErrorHandlers() {
  window.addEventListener('error', (ev) => {
    const msg = ev.message || ev.error?.message || String(ev.error || 'Unknown error');
    handle(msg);
  });
  window.addEventListener('unhandledrejection', (ev) => {
    const reason = ev.reason instanceof Error ? ev.reason.message : String(ev.reason);
    handle(reason);
  });
}
