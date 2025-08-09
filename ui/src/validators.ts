export function validatePeerForm(
  endpoint: string,
  allowedIPs: string,
  keepalive: string
): string | null {
  if (!endpoint.trim()) return 'peers.endpointRequired';
  if (!allowedIPs.trim()) return 'peers.allowedRequired';
  if (keepalive && isNaN(Number(keepalive))) return 'peers.keepaliveNumber';
  return null;
}
