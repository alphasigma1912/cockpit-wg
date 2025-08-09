export interface PeerState {
  publicKey: string;
  enabled: boolean;
}

export type PeerAction =
  | { type: 'add'; peer: PeerState }
  | { type: 'toggle'; publicKey: string };

export function peerReducer(state: PeerState[], action: PeerAction): PeerState[] {
  switch (action.type) {
    case 'add':
      return [...state, action.peer];
    case 'toggle':
      return state.map((p) =>
        p.publicKey === action.publicKey ? { ...p, enabled: !p.enabled } : p
      );
    default:
      return state;
  }
}
