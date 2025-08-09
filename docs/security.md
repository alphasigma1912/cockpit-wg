# Security Overview

## Threat model
- Unprivileged users attempting to escalate via the web UI
- Injection of shell arguments or malformed configs
- Theft of WireGuard private keys
- Tampering with configuration files on disk
- Exfiltration of `.wgx` bundles in transit or at rest

## Baseline controls
- Backend runs with **least privilege** and gates privileged calls with **Polkit**
- Only a small, explicit JSON-RPC API; **no arbitrary command execution**
- Strict input validation and canonicalization on all requests
- Frontend served with a **Content-Security-Policy** that forbids inline scripts
- Private keys never leave the backend and are zeroized from memory
- Sensitive files written with `umask 077` and **atomic writes** with rollback
- Audit logs recorded for all privileged operations
- Exchange inbox requires authenticated, **minisign-signed** bundles and verifies SHA-256 checksums
