import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';
import Peers from './Peers';

vi.mock('./backend', () => ({
  default: {
    addPeer: vi.fn().mockResolvedValue({ publicKey: 'pub', privateKey: 'priv' }),
  },
}));

const backend = (await import('./backend')).default as { addPeer: any };

describe('Peers form', () => {
  it('validates required fields', async () => {
    render(<Peers />);
    await userEvent.click(screen.getByRole('button', { name: /add peer/i }));
    expect(await screen.findByRole('alert')).toHaveTextContent(/endpoint is required/i);
    expect(backend.addPeer).not.toHaveBeenCalled();
  });

  it('submits valid data', async () => {
    render(<Peers />);
    await userEvent.type(screen.getByLabelText(/endpoint/i), '1.2.3.4');
    await userEvent.type(screen.getByLabelText(/allowed ips/i), '10.0.0.0/24');
    await userEvent.click(screen.getByRole('button', { name: /add peer/i }));
    expect(backend.addPeer).toHaveBeenCalled();
  });
});
