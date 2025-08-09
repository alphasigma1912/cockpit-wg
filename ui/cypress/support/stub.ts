export function mockBackend(
  win: any,
  overrides: Record<string, any> = {}
) {
  const defaultResponses: Record<string, any> = {
    CheckPrereqs: { kernel: true, tools: true, systemd: true },
    InstallPackages: {},
    ListInterfaces: { interfaces: ['wg0'] },
    GetInterfaceStatus: { status: 'up', last_change: 'now', message: '' },
    GetMetrics: { timestamps: [0, 1], rx: [0, 1], tx: [0, 1] },
    AddPeer: { publicKey: 'pub', privateKey: 'priv' },
    ImportBundle: {},
    GetExchangeKey: 'abc',
    RotateKeys: {},
    DownInterface: {},
  };
  const responses = { ...defaultResponses, ...overrides };
  const listeners = new Set<(e: any, data: any) => void>();
  win.cockpit = {
    channel: () => ({
      addEventListener: (_: any, cb: any) => listeners.add(cb),
      removeEventListener: (_: any, cb: any) => listeners.delete(cb),
      send: (req: any) => {
        const { id, method } = req;
        const res = responses[method];
        const result = typeof res === 'function' ? res(req.params) : res;
        listeners.forEach((cb) => cb(null, { id, result }));
      },
    }),
  };
}
