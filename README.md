# Cockpit WireGuard Manager

Cockpit WireGuard Manager is a [Cockpit](https://cockpit-project.org/) plugin that simplifies WireGuard setup and administration on Linux servers.

## Scope

- Automatically installs WireGuard on first run using least-privilege escalation.
- Manages WireGuard interfaces and peers.
- Safely edits configuration files with validation and atomic writes.
- Displays live traffic graphs.
- Supports a secure configuration exchange mechanism between nodes without opening additional network services.

## Acceptance Criteria

The project is considered successful when all of the following measurable conditions are met:

1. **Zero critical/high vulnerabilities** – Linters and security scanners report no critical or high severity issues.
2. **No private key leakage** – Private keys are never sent to the browser or recorded in logs.
3. **Validated, atomic writes** – Changes to `/etc/wireguard` are validated and written atomically.
4. **Reproducible builds** – Builds produce identical artifacts when run with the same inputs.
5. **Full test coverage for critical paths** – Parsing, apply, and configuration exchange code paths have 100% test coverage.
6. **Distribution packages** – Packages are produced for Debian/Ubuntu and RHEL/Fedora.

