import { describe, expect, it } from 'vitest';
import { peerReducer, PeerState } from './reducers';

describe('peerReducer', () => {
  it('adds a peer', () => {
    const peer: PeerState = { publicKey: 'abc', enabled: true };
    const state: PeerState[] = [];
    const next = peerReducer(state, { type: 'add', peer });
    expect(next).toHaveLength(1);
    expect(next[0]).toEqual(peer);
  });

  it('toggles a peer', () => {
    const state: PeerState[] = [{ publicKey: 'abc', enabled: true }];
    const next = peerReducer(state, { type: 'toggle', publicKey: 'abc' });
    expect(next[0].enabled).toBe(false);
  });
});
