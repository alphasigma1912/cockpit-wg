declare const cockpit: any;

class Backend {
  private channel = cockpit.channel({ payload: 'json', path: 'wg-bridge', superuser: 'require' });
  private seq = 0;

  private call(method: string, params: any = {}): Promise<any> {
    const id = String(this.seq++);
    return new Promise((resolve, reject) => {
      const handler = (_event: any, data: any) => {
        if (data.id === id) {
          this.channel.removeEventListener('message', handler);
          if (data.error) {
            reject(data.error);
          } else {
            resolve(data.result);
          }
        }
      };
      this.channel.addEventListener('message', handler);
      this.channel.send({ jsonrpc: '2.0', id, method, params });
    });
  }

  checkPrereqs(): Promise<any> {
    return this.call('CheckPrereqs');
  }

  installPackages(): Promise<any> {
    return this.call('InstallPackages');
  }
}

export default new Backend();
