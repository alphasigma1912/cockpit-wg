# Cockpit WireGuard Manager

Cockpit WireGuard Manager is a [Cockpit](https://cockpit-project.org/) plugin that simplifies WireGuard setup and administration on Linux servers.

## Scope

- Automatically installs WireGuard on first run using least-privilege escalation.
- Manages WireGuard interfaces and peers.
- Safely edits configuration files with validation and atomic writes.
- Displays live traffic graphs.
- Supports a secure configuration exchange mechanism between nodes without opening additional network services.

## Repository Layout

- **ui/** – React + PatternFly frontend built with Vite
- **bridge/** – Go backend (`wg-bridge`)
- **packaging/** – Distribution packaging files
- **ansible/** – Ansible role for mass deployment
- **schemas/** – JSON schemas for API and `.wgx` format
- **docs/** – Project documentation
- **scripts/** – Development helpers

## Build

Run `make dist` to build the frontend and backend and assemble the Cockpit plugin under `dist/cockpit-wg/`.

## Security

See [SECURITY.md](SECURITY.md) for the threat model and security baseline.
