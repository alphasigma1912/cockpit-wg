#!/bin/sh
set -e
install -d -m 0755 /usr/share/cockpit/cockpit-wg
chmod 0755 /usr/share/cockpit/cockpit-wg/wg-bridge 2>/dev/null || true
if command -v pkcheck >/dev/null 2>&1; then
  pkcheck --version >/dev/null 2>&1 || true
fi
if command -v systemctl >/dev/null 2>&1; then
  systemctl daemon-reload >/dev/null 2>&1 || true
fi
exit 0
