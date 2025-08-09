declare const cockpit: any;

class Backend {
  private channel = cockpit.channel({
    payload: "json",
    path: "wg-bridge",
    superuser: "require",
  });
  private seq = 0;

  private call(method: string, params: any = {}): Promise<any> {
    const id = String(this.seq++);
    return new Promise((resolve, reject) => {
      const handler = (_event: any, data: any) => {
        if (data.id === id) {
          this.channel.removeEventListener("message", handler);
          if (data.error) {
            reject(data.error);
          } else {
            resolve(data.result);
          }
        }
      };
      this.channel.addEventListener("message", handler);
      this.channel.send({ jsonrpc: "2.0", id, method, params });
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
