import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import type { Channel } from './logger';

const LOGGER_PATH = './logger';

beforeEach(() => {
  vi.resetModules();
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
  vi.unstubAllEnvs();
  vi.useRealTimers();
  if (globalThis.window?.localStorage) {
    window.localStorage.clear();
  }
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe('logger', () => {
  it('defaults to WARN in production mode', async () => {
    vi.stubEnv('PROD', 'true');
    const { logger } = await import(LOGGER_PATH);
    const group = vi.spyOn(console, 'groupCollapsed').mockImplementation(() => {});
    logger.debug('UI', 'dbg');
    expect(group).not.toHaveBeenCalled();
    logger.warn('UI', 'warn');
    expect(group).toHaveBeenCalledTimes(1);
  });

  it('defaults to DEBUG in development mode', async () => {
    const { logger } = await import(LOGGER_PATH);
    const group = vi.spyOn(console, 'groupCollapsed').mockImplementation(() => {});
    logger.debug('UI', 'dbg');
    expect(group).toHaveBeenCalledTimes(1);
    logger.trace('UI', 'trace');
    expect(group).toHaveBeenCalledTimes(1);
  });

  it('enables all log levels when ?debug=1 is present', async () => {
    vi.stubEnv('PROD', 'true');
    vi.stubGlobal('window', {
      location: { search: '?debug=1' },
      localStorage: { getItem: () => null, setItem: () => {}, clear: () => {} },
    });
    const { logger } = await import(LOGGER_PATH);
    const group = vi.spyOn(console, 'groupCollapsed').mockImplementation(() => {});
    logger.trace('UI', 'trace');
    expect(group).toHaveBeenCalledTimes(1);
  });

  it('enables all log levels when localStorage cwm.debug=1 is set', async () => {
    vi.stubEnv('PROD', 'true');
    window.localStorage.setItem('cwm.debug', '1');
    const { logger } = await import(LOGGER_PATH);
    const group = vi.spyOn(console, 'groupCollapsed').mockImplementation(() => {});
    logger.trace('UI', 'trace');
    expect(group).toHaveBeenCalledTimes(1);
  });

  it('supports channel filtering via enableChannels', async () => {
    const { logger, LogLevel } = await import(LOGGER_PATH);
    logger.setLevel(LogLevel.TRACE);
    const group = vi.spyOn(console, 'groupCollapsed').mockImplementation(() => {});
    logger.enableChannels('RPC');
    logger.info('RPC', 'yes');
    logger.info('UI', 'no');
    expect(group).toHaveBeenCalledTimes(1);
  });

  it('setLevel overrides current level', async () => {
    const { logger, LogLevel } = await import(LOGGER_PATH);
    logger.setLevel(LogLevel.ERROR);
    const group = vi.spyOn(console, 'groupCollapsed').mockImplementation(() => {});
    logger.warn('UI', 'no');
    logger.error('UI', 'err');
    expect(group).toHaveBeenCalledTimes(1);
  });

  it('formats messages with timestamp and grouping', async () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2024-01-02T03:04:05Z'));
    const { logger, LogLevel } = await import(LOGGER_PATH);
    logger.setLevel(LogLevel.TRACE);
    const group = vi.spyOn(console, 'groupCollapsed').mockImplementation(() => {});
    const log = vi.spyOn(console, 'log').mockImplementation(() => {});
    const end = vi.spyOn(console, 'groupEnd').mockImplementation(() => {});
    logger.info('UI', 'hello');
    expect(group).toHaveBeenCalledWith('[2024-01-02T03:04:05.000Z] [UI] INFO');
    expect(log).toHaveBeenCalledWith('[2024-01-02T03:04:05.000Z] [UI] hello');
    expect(end).toHaveBeenCalledTimes(1);
  });

  it('handles unknown or disabled channels', async () => {
    const { logger, LogLevel } = await import(LOGGER_PATH);
    logger.setLevel(LogLevel.TRACE);
    const group = vi.spyOn(console, 'groupCollapsed').mockImplementation(() => {});
    logger.enableChannels();
    logger.info('UI', 'no');
    logger.info('UNKNOWN' as unknown as Channel, 'no');
    expect(group).not.toHaveBeenCalled();
  });

  it('logs for error, warn, info, debug, and trace methods', async () => {
    const { logger, LogLevel } = await import(LOGGER_PATH);
    logger.setLevel(LogLevel.TRACE);
    const group = vi.spyOn(console, 'groupCollapsed').mockImplementation(() => {});
    logger.error('UI', 'e');
    logger.warn('UI', 'w');
    logger.info('UI', 'i');
    logger.debug('UI', 'd');
    logger.trace('UI', 't');
    expect(group).toHaveBeenCalledTimes(5);
  });

  it('stores redacted log events', async () => {
    const { logger, getLogEvents } = await import(LOGGER_PATH);
    const group = vi.spyOn(console, 'groupCollapsed').mockImplementation(() => {});
    logger.info('UI', 'x', { secret: 'y' });
    expect(group).toHaveBeenCalledTimes(1);
    const events = getLogEvents(1);
    expect(events[0].args).toEqual(['x', '[REDACTED]']);
  });

  it('redacts sensitive strings', async () => {
    const { logger, getLogEvents, clearLogEvents } = await import(LOGGER_PATH);
    vi.spyOn(console, 'groupCollapsed').mockImplementation(() => {});
    clearLogEvents();
    const msg = 'PrivateKey = ' + 'A'.repeat(43) + '=';
    logger.error('UI', msg);
    let events = getLogEvents(1);
    expect(events[0].args[0]).toBe('PrivateKey = [REDACTED]');
    clearLogEvents();
    const msg2 = 'PresharedKey = ' + 'A'.repeat(43) + '=';
    logger.error('UI', msg2);
    events = getLogEvents(1);
    expect(events[0].args[0]).toBe('PresharedKey = [REDACTED]');
  });
});

