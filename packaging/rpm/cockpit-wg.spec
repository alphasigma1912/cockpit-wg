Name:           cockpit-wg
Version:        0.0.0
Release:        1%{?dist}
Summary:        Cockpit WireGuard Manager plugin
License:        MIT
URL:            https://github.com/cockpit-wg
BuildArch:      x86_64

%description
Cockpit plugin for managing WireGuard.

%files
%dir /usr/share/cockpit/cockpit-wg
%attr(0755,root,root) /usr/share/cockpit/cockpit-wg/wg-bridge
/usr/share/cockpit/cockpit-wg/*
/usr/share/polkit-1/actions/org.cockpit-project.cockpit-wg.policy

%post
/usr/bin/pkcheck --version >/dev/null 2>&1 || true
/usr/bin/systemctl daemon-reload >/dev/null 2>&1 || true

%postun
rm -rf /usr/share/cockpit/cockpit-wg
rm -f /usr/share/polkit-1/actions/org.cockpit-project.cockpit-wg.policy
/usr/bin/pkcheck --version >/dev/null 2>&1 || true
/usr/bin/systemctl daemon-reload >/dev/null 2>&1 || true
