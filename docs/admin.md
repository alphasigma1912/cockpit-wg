# Admin Guide

## Installation

### Pre-built packages
```bash
# Download latest release for your architecture
wget https://github.com/alphasigma1912/cockpit-wg/releases/latest/download/cockpit-wg-linux-amd64.tar.gz

# Extract and install
sudo tar -xzf cockpit-wg-linux-amd64.tar.gz -C /usr/share/cockpit/

# Restart Cockpit to load the plugin
sudo systemctl restart cockpit
```

### Build from source
```bash
# Assemble the plugin for the current platform
make dist

# Install the build
sudo cp -r dist/cockpit-wg/ /usr/share/cockpit/
sudo systemctl restart cockpit
```

## First run
1. Log in to Cockpit at `https://<host>:9090`.
2. Open **Cockpit WireGuard Manager**.
3. Authorize the Polkit prompt to allow package installation and key generation.
4. The backend creates exchange/signing keys and installs WireGuard if missing.

## Basic operations
- **Create interface** – Use the *Add Interface* form, then click **Apply Changes**.
- **Add peer** – Select an interface, open *Peers*, and use **Add Peer**.
- **Start/stop interface** – Use the toggle in the interface list.
- **Export config** – Click **Export** to produce a `.wgx` bundle.
- **Import config** – Drop a `.wgx` bundle on the interface list or use *Import Bundle*.

### CLI examples
```bash
# List interfaces
sudo /usr/share/cockpit/cockpit-wg/wg-bridge <<'RPC'
{"jsonrpc":"2.0","id":1,"method":"ListInterfaces"}
RPC

# Bring an interface up
sudo /usr/share/cockpit/cockpit-wg/wg-bridge <<'RPC'
{"jsonrpc":"2.0","id":1,"method":"UpInterface","params":{"name":"wg0"}}
RPC
```
