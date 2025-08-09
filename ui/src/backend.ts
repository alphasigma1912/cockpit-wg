import { logError } from './errorBuffer';
import logger from './logger';
declare const cockpit: any;

class Backend {
  private channel = cockpit.channel({
    payload: "json",
    path: "wg-bridge",
    superuser: "require",
  });
  private seq = 0;

  private redact(obj: any): any {
    if (obj && typeof obj === 'object') {
      const res: any = Array.isArray(obj) ? [] : {};
      for (const k of Object.keys(obj)) {
        res[k] = '***';
      }
      return res;
    }
    return obj;
  }

  private call(method: string, params: any = {}): Promise<any> {
    const id = String(this.seq++);
    const trace = crypto.randomUUID();
    const start = performance.now();
    const timeoutMs = method === 'InstallPackages' ? 30000 : 10000;
    const controller = new AbortController();
    return new Promise((resolve, reject) => {
      const cleanup = (handler: any) => {
        window.clearTimeout(timer);
        this.channel.removeEventListener('message', handler);
      };
      const timer = window.setTimeout(() => {
        controller.abort();
      }, timeoutMs);

      const handler = (_event: any, data: any) => {
        if (data.id === id && data.trace === trace) {
          cleanup(handler);
          const duration = performance.now() - start;
          if (data.error) {
            const err = { ...data.error, trace };
            logger.error('RPC', 'error', { trace, method, duration, code: err.code, message: err.message });
            logError(err);
            reject(err);
          } else {
            const size = JSON.stringify(data.result ?? {}).length;
            logger.info('RPC', 'response', { trace, method, duration, size });
            resolve(data.result);
          }
        }
      };

      controller.signal.addEventListener('abort', () => {
        cleanup(handler);
        const duration = performance.now() - start;
        const err = { code: -1, message: 'timeout', trace };
        logger.error('RPC', 'timeout', { trace, method, duration });
        logError(err);
        reject(err);
      });

      logger.info('RPC', 'request', { trace, method, params: this.redact(params) });
      this.channel.addEventListener('message', handler);
      this.channel.send({ jsonrpc: '2.0', id, method, params, trace });
    });
  }

  checkPrereqs(): Promise<any> {
    return this.call("CheckPrereqs");
  }

  installPackages(): Promise<any> {
    return this.call("InstallPackages");
  }

  runSelfTest(): Promise<any> {
    return this.call("RunSelfTest");
  }

  listInterfaces(): Promise<any> {
    return this.call("ListInterfaces");
  }

  getInterfaceStatus(name: string): Promise<any> {
    return this.call("GetInterfaceStatus", { name });
  }

  upInterface(name: string): Promise<any> {
    return this.call("UpInterface", { name });
  }

  downInterface(name: string): Promise<any> {
    return this.call("DownInterface", { name });
  }

  restartInterface(name: string): Promise<any> {
    return this.call("RestartInterface", { name });
  }

  getMetrics(name: string): Promise<any> {
    return this.call("GetMetrics", { name });
  }

  addPeer(name: string, peer: any): Promise<any> {
    return this.call("AddPeer", { name, peer });
  }

  listPeers(name: string): Promise<any> {
    return this.call("ListPeers", { name });
  }

  removePeer(name: string, publicKey: string): Promise<any> {
    return this.call("RemovePeer", { name, publicKey });
  }

  updatePeer(name: string, publicKey: string, peer: any): Promise<any> {
    return this.call("UpdatePeer", { name, publicKey, peer });
  }

  importBundle(bundle: string): Promise<any> {
    return this.call("ImportBundle", { bundle });
  }

  getExchangeKey(): Promise<any> {
    return this.call("GetExchangeKey");
  }

  rotateKeys(): Promise<any> {
    return this.call("RotateKeys");
  }
}

export default new Backend();
