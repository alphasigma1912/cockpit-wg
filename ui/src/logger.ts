/* eslint-disable no-console */
export enum LogLevel {
  ERROR = 0,
  WARN = 1,
  INFO = 2,
  DEBUG = 3,
  TRACE = 4,
}

export type Channel = 'UI' | 'RPC' | 'STATE' | 'METRICS' | 'EXCHANGE';

class Logger {
  private level: LogLevel;
  private channels: Set<Channel> = new Set(['UI', 'RPC', 'STATE', 'METRICS', 'EXCHANGE']);

  constructor() {
    const params = typeof window !== 'undefined' ? new globalThis.URLSearchParams(window.location.search) : new globalThis.URLSearchParams();
    const debugEnabled = (typeof window !== 'undefined' && window.localStorage.getItem('cwm.debug') === '1') || params.get('debug') === '1';
    if (debugEnabled) {
      this.level = LogLevel.TRACE;
    } else if (import.meta.env.PROD) {
      this.level = LogLevel.WARN; // only error and warn
    } else {
      this.level = LogLevel.DEBUG;
    }
  }

  setLevel(level: LogLevel) {
    this.level = level;
  }

  enableChannels(...channels: Channel[]) {
    this.channels = new Set(channels);
  }

  private shouldLog(level: LogLevel, channel: Channel) {
    return level <= this.level && this.channels.has(channel);
  }

  private log(level: LogLevel, channel: Channel, ...args: unknown[]) {
    if (!this.shouldLog(level, channel)) return;
    const timestamp = new Date().toISOString();
    const prefix = `[${timestamp}] [${channel}]`;
    console.groupCollapsed(`${prefix} ${LogLevel[level]}`);
    for (const arg of args) {
      if (typeof arg === 'string') {
        console.log(`${prefix} ${arg}`);
      } else {
        console.log(arg);
      }
    }
    console.groupEnd();
  }

  error(channel: Channel, ...args: unknown[]) {
    this.log(LogLevel.ERROR, channel, ...args);
  }

  warn(channel: Channel, ...args: unknown[]) {
    this.log(LogLevel.WARN, channel, ...args);
  }

  info(channel: Channel, ...args: unknown[]) {
    this.log(LogLevel.INFO, channel, ...args);
  }

  debug(channel: Channel, ...args: unknown[]) {
    this.log(LogLevel.DEBUG, channel, ...args);
  }

  trace(channel: Channel, ...args: unknown[]) {
    this.log(LogLevel.TRACE, channel, ...args);
  }
}

export const logger = new Logger();

export default logger;
