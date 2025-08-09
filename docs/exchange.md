# Configuration Exchange

## `.wgx` bundle format
```
bundle.wgx  (age-encrypted and minisign-signed)
└── tar archive
    ├── manifest.json   # {"interface","version","checksum"}
    ├── config.conf     # WireGuard config
    └── meta/           # optional metadata files
```
- `checksum` is the SHA-256 of `config.conf`
- `version` starts at `1` and increments on updates

## Key provisioning
Exchange and signing keys live in `/etc/cockpit-wg/keys` and are created on first run.
```bash
# Retrieve this node's exchange public key
echo '{"jsonrpc":"2.0","id":1,"method":"GetExchangeKey"}' \
  | sudo /usr/share/cockpit/cockpit-wg/wg-bridge

# Rotate all exchange and signing keys
echo '{"jsonrpc":"2.0","id":1,"method":"RotateKeys"}' \
  | sudo /usr/share/cockpit/cockpit-wg/wg-bridge
```

## Exporting
```bash
# Produce a bundle for interface wg0 encrypted to recipient's public key
RECIPIENT="age1example..."
echo '{"jsonrpc":"2.0","id":1,"method":"ExportConfig","params":{"iface":"wg0","recipient":"'"$RECIPIENT"'"}}' \
  | sudo /usr/share/cockpit/cockpit-wg/wg-bridge
# Output: path to generated .wgx and .wgx.minisig files
```

## Importing
1. Place `<file>.wgx` and `<file>.wgx.minisig` into `/var/lib/cockpit-wg/inbox/`
2. The daemon verifies signature, decrypts, and stages the config in `pending/`
3. Apply the pending config through the UI or JSON-RPC
```bash
# List bundles detected in the inbox
echo '{"jsonrpc":"2.0","id":1,"method":"ListInbox"}' \
  | sudo /usr/share/cockpit/cockpit-wg/wg-bridge
```
