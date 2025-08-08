# Security Policy

## Threat Model

An attacker is assumed to have network access and possibly a low-privileged account. Potential attacks include:

- Injection of shell arguments through the web UI.
- Abuse of privileged operations.
- Theft of private keys.
- Tampering with configuration files.
- Exfiltration of configuration exchange bundles.

## Security Baseline Controls

To mitigate these threats, the project adopts the following baseline controls:

- Principle of least privilege for all components.
- Polkit-gated backend exposing a very small, explicit API.
- Strict input validation and canonicalization.
- No arbitrary command execution.
- Content-Security-Policy forbidding inline scripts.
- Zero private keys retained in frontend memory.
- `umask 077` for all sensitive file operations.
- Atomic writes with rollback on failure.
- Audit logs for privileged actions.
- Rate-limited, authenticated import of configuration exchange bundles.

