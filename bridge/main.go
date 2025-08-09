package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/coreos/go-systemd/v22/journal"
	"golang.zx2c4.com/wireguard/wgctrl"
)

var sensitiveRx = regexp.MustCompile(`(?i)(PrivateKey|PresharedKey)\s*=\s*[^\s]+`)

func sanitizeOutput(s string) string {
	return sensitiveRx.ReplaceAllString(s, "$1=[REDACTED]")
}

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *respError      `json:"error,omitempty"`
}

type respError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func main() {
	initMetricsCollector()
	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	for scanner.Scan() {
		line := scanner.Bytes()
		var req request
		if err := json.Unmarshal(line, &req); err != nil {
			continue
		}
		res := handleRequest(&req)
		b, _ := json.Marshal(res)
		writer.Write(b)
		writer.WriteByte('\n')
		writer.Flush()
	}
}

func handleRequest(req *request) *response {
	var result interface{}
	var err error
	switch req.Method {
	case "ListInterfaces":
		result, err = listInterfaces()
	case "ReadConfig":
		var p struct {
			Name string `json:"name"`
		}
		if err = json.Unmarshal(req.Params, &p); err == nil {
			result, err = readConfig(p.Name)
		}
	case "ValidateConfig":
		var p struct {
			Text string `json:"text"`
		}
		if err = json.Unmarshal(req.Params, &p); err == nil {
			result, err = validateConfig(p.Text)
		}
	case "ApplyChanges":
		var p struct {
			Name string `json:"name"`
			Text string `json:"text"`
		}
		if err = json.Unmarshal(req.Params, &p); err == nil {
			result, err = applyChanges(p.Name, p.Text)
		}
	case "WriteConfig":
		var p struct {
			Name string `json:"name"`
			Text string `json:"text"`
		}
		if err = json.Unmarshal(req.Params, &p); err == nil {
			result, err = writeConfig(p.Name, p.Text)
		}
	case "ReloadInterface":
		var p struct {
			Name string `json:"name"`
		}
		if err = json.Unmarshal(req.Params, &p); err == nil {
			result, err = reloadInterface(p.Name)
		}
	case "UpInterface":
		var p struct {
			Name string `json:"name"`
		}
		if err = json.Unmarshal(req.Params, &p); err == nil {
			result, err = upInterface(p.Name)
		}
	case "DownInterface":
		var p struct {
			Name string `json:"name"`
		}
		if err = json.Unmarshal(req.Params, &p); err == nil {
			result, err = downInterface(p.Name)
		}
	case "GetInterfaceStatus":
		var p struct {
			Name string `json:"name"`
		}
		if err = json.Unmarshal(req.Params, &p); err == nil {
			result, err = getInterfaceStatus(p.Name)
		}
	case "RestartInterface":
		var p struct {
			Name string `json:"name"`
		}
		if err = json.Unmarshal(req.Params, &p); err == nil {
			result, err = restartInterface(p.Name)
		}
	case "GetMetrics":
		var p struct {
			Name string `json:"name"`
		}
		if err = json.Unmarshal(req.Params, &p); err == nil {
			result, err = getMetrics(p.Name)
		}
	case "CheckPrereqs":
		result, err = checkPrereqs()
	case "InstallPackages":
		result, err = installPackages()
	case "RunSelfTest":
		result, err = runSelfTest()
	case "AddPeer":
		var p struct {
			Name string     `json:"name"`
			Peer peerParams `json:"peer"`
		}
		if err = json.Unmarshal(req.Params, &p); err == nil {
			result, err = addPeer(p.Name, p.Peer)
		}
	case "RemovePeer":
		var p struct {
			Name      string `json:"name"`
			PublicKey string `json:"publicKey"`
		}
		if err = json.Unmarshal(req.Params, &p); err == nil {
			result, err = removePeer(p.Name, p.PublicKey)
		}
	case "UpdatePeer":
		var p struct {
			Name      string     `json:"name"`
			PublicKey string     `json:"publicKey"`
			Peer      peerParams `json:"peer"`
		}
		if err = json.Unmarshal(req.Params, &p); err == nil {
			result, err = updatePeer(p.Name, p.PublicKey, p.Peer)
		}
	case "ListPeers":
		var p struct {
			Name string `json:"name"`
		}
		if err = json.Unmarshal(req.Params, &p); err == nil {
			result, err = listPeers(p.Name)
		}
	default:
		err = errors.New("unknown method")
	}

	if err != nil {
		auditLog(req.Method, req.Params, err)
		return &response{JSONRPC: "2.0", ID: req.ID, Error: &respError{Code: -1, Message: err.Error()}}
	}
	auditLog(req.Method, req.Params, nil)
	return &response{JSONRPC: "2.0", ID: req.ID, Result: result}
}

func listInterfaces() (interface{}, error) {
	dir := "/etc/wireguard"
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]interface{}{"interfaces": []string{}}, nil
		}
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".conf") {
			base := strings.TrimSuffix(name, ".conf")
			if ifaceRx.MatchString(base) {
				names = append(names, base)
			}
		}
	}
	sort.Strings(names)
	return map[string]interface{}{"interfaces": names}, nil
}

var ifaceRx = regexp.MustCompile(`^[a-zA-Z0-9_.-]{1,15}$`)

type configSummary struct {
	Interface map[string]string   `json:"interface"`
	Peers     []map[string]string `json:"peers"`
}

func readConfig(name string) (interface{}, error) {
	if !ifaceRx.MatchString(name) {
		return nil, fmt.Errorf("invalid interface name")
	}
	path := filepath.Join("/etc/wireguard", name+".conf")
	if !strings.HasPrefix(filepath.Clean(path), "/etc/wireguard/") {
		return nil, fmt.Errorf("invalid path")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	summary, err := parseConfig(string(data))
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"raw": string(data), "summary": summary}, nil
}

func validateConfig(text string) (interface{}, error) {
	summary, err := parseConfig(text)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"summary": summary}, nil
}

func applyChanges(name, text string) (interface{}, error) {
	auditApply("start", name, "", nil)

	if !ifaceRx.MatchString(name) {
		err := fmt.Errorf("invalid interface name")
		auditApply("failure", name, "validate", err)
		return nil, err
	}
	summary, err := parseConfig(text)
	if err != nil {
		auditApply("failure", name, "validate", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	lockDir := "/run/cockpit-wg/locks"
	if err := os.MkdirAll(lockDir, 0755); err != nil {
		auditApply("failure", name, "lock", err)
		return nil, err
	}
	lockFile, err := os.OpenFile(filepath.Join(lockDir, name+".lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		auditApply("failure", name, "lock", err)
		return nil, err
	}
	defer lockFile.Close()
	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX); err != nil {
		auditApply("failure", name, "lock", err)
		return nil, err
	}
	defer syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)

	dir := "/etc/wireguard"
	cfgPath := filepath.Join(dir, name+".conf")
	if !strings.HasPrefix(filepath.Clean(cfgPath), "/etc/wireguard/") {
		err := fmt.Errorf("invalid path")
		auditApply("failure", name, "validate", err)
		return nil, err
	}

	tmp, err := os.CreateTemp(dir, name+".tmp")
	if err != nil {
		auditApply("failure", name, "write", err)
		return nil, err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		auditApply("failure", name, "write", err)
		return nil, err
	}
	if _, err := tmp.WriteString(text); err != nil {
		tmp.Close()
		auditApply("failure", name, "write", err)
		return nil, err
	}
	tmp.Close()

	backupPath := cfgPath + ".bak"
	if err := os.Rename(cfgPath, backupPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			auditApply("failure", name, "backup", err)
			return nil, err
		}
	}
	if err := os.Rename(tmpName, cfgPath); err != nil {
		os.Rename(backupPath, cfgPath)
		auditApply("failure", name, "write", err)
		return nil, err
	}

	if err := exec.Command("wg", "syncconf", name, cfgPath).Run(); err != nil {
		auditApply("failure", name, "syncconf", err)
		os.Remove(cfgPath)
		os.Rename(backupPath, cfgPath)
		exec.Command("wg", "syncconf", name, cfgPath).Run()
		auditApply("rollback", name, "syncconf", err)
		return nil, fmt.Errorf("wg syncconf failed: %w", err)
	}

	if err := verifyAppliedConfig(name, summary); err != nil {
		auditApply("failure", name, "verify", err)
		os.Remove(cfgPath)
		os.Rename(backupPath, cfgPath)
		exec.Command("wg", "syncconf", name, cfgPath).Run()
		auditApply("rollback", name, "verify", err)
		return nil, fmt.Errorf("verification failed: %w", err)
	}

	os.Remove(backupPath)
	auditApply("success", name, "", nil)
	return map[string]string{"status": "ok"}, nil
}

func verifyAppliedConfig(name string, summary *configSummary) error {
	client, err := wgctrl.New()
	if err != nil {
		return err
	}
	defer client.Close()

	dev, err := client.Device(name)
	if err != nil {
		return err
	}

	if lp, ok := summary.Interface["ListenPort"]; ok && lp != "" {
		exp, err := strconv.Atoi(lp)
		if err != nil {
			return fmt.Errorf("invalid listen port %q", lp)
		}
		if dev.ListenPort != exp {
			return fmt.Errorf("listen port %d != expected %d", dev.ListenPort, exp)
		}
	}

	expected := make(map[string]struct{})
	for _, p := range summary.Peers {
		if pk, ok := p["PublicKey"]; ok {
			expected[pk] = struct{}{}
		}
	}
	if len(dev.Peers) != len(expected) {
		return fmt.Errorf("peer count mismatch")
	}
	for _, p := range dev.Peers {
		if _, ok := expected[p.PublicKey.String()]; !ok {
			return fmt.Errorf("unexpected peer %s", p.PublicKey.String())
		}
	}
	return nil
}

func writeConfig(name, text string) (interface{}, error) {
	if !ifaceRx.MatchString(name) {
		return nil, fmt.Errorf("invalid interface name")
	}
	if _, err := validateConfig(text); err != nil {
		return nil, err
	}
	dir := "/etc/wireguard"
	tmp, err := os.CreateTemp(dir, name+".tmp")
	if err != nil {
		return nil, err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		return nil, err
	}
	if _, err := tmp.WriteString(text); err != nil {
		tmp.Close()
		return nil, err
	}
	tmp.Close()

	if _, err := validateConfig(text); err != nil {
		return nil, err
	}

	cfgPath := filepath.Join(dir, name+".conf")
	if !strings.HasPrefix(filepath.Clean(cfgPath), "/etc/wireguard/") {
		return nil, fmt.Errorf("invalid path")
	}
	backupPath := cfgPath + ".bak"
	if err := os.Rename(cfgPath, backupPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}
	if err := os.Rename(tmpName, cfgPath); err != nil {
		os.Rename(backupPath, cfgPath)
		return nil, err
	}
	if _, err := reloadInterface(name); err != nil {
		os.Rename(cfgPath, tmpName)
		os.Rename(backupPath, cfgPath)
		reloadInterface(name)
		return nil, err
	}
	os.Remove(backupPath)
	return map[string]string{"status": "ok"}, nil
}

func reloadInterface(name string) (interface{}, error) {
	if !ifaceRx.MatchString(name) {
		return nil, fmt.Errorf("invalid interface name")
	}
	path := filepath.Join("/etc/wireguard", name+".conf")
	if !strings.HasPrefix(filepath.Clean(path), "/etc/wireguard/") {
		return nil, fmt.Errorf("invalid path")
	}
	if err := exec.Command("wg", "syncconf", name, path).Run(); err != nil {
		if err2 := exec.Command("systemctl", "reload", fmt.Sprintf("wg-quick@%s", name)).Run(); err2 != nil {
			return nil, err2
		}
	}
	return map[string]string{"status": "ok"}, nil
}

func parseConfig(text string) (*configSummary, error) {
	summary := &configSummary{Interface: make(map[string]string)}
	var curr map[string]string
	scanner := bufio.NewScanner(strings.NewReader(text))
	line := 0
	for scanner.Scan() {
		line++
		l := strings.TrimSpace(scanner.Text())
		if l == "" || strings.HasPrefix(l, "#") || strings.HasPrefix(l, ";") {
			continue
		}
		if strings.HasPrefix(l, "[") && strings.HasSuffix(l, "]") {
			sect := strings.TrimSpace(l[1 : len(l)-1])
			switch sect {
			case "Interface":
				curr = summary.Interface
			case "Peer":
				m := make(map[string]string)
				summary.Peers = append(summary.Peers, m)
				curr = m
			default:
				return nil, fmt.Errorf("unknown section %q at line %d", sect, line)
			}
			continue
		}
		if curr == nil {
			return nil, fmt.Errorf("key-value outside of section at line %d", line)
		}
		parts := strings.SplitN(l, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid line %d", line)
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		curr[key] = val
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(summary.Interface) == 0 {
		return nil, fmt.Errorf("missing [Interface] section")
	}
	if len(summary.Peers) == 0 {
		return nil, fmt.Errorf("no peers defined")
	}
	seen := make(map[string]struct{})
	for _, p := range summary.Peers {
		pk, ok := p["PublicKey"]
		if !ok || pk == "" {
			return nil, fmt.Errorf("peer missing PublicKey")
		}
		if _, dup := seen[pk]; dup {
			return nil, fmt.Errorf("duplicate peer %s", pk)
		}
		seen[pk] = struct{}{}
		allowed, ok := p["AllowedIPs"]
		if !ok {
			return nil, fmt.Errorf("peer %s missing AllowedIPs", pk)
		}
		ips := strings.Split(allowed, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if ip == "" {
				continue
			}
			if ip == "0.0.0.0/0" || ip == "::/0" {
				return nil, fmt.Errorf("disallowed AllowedIPs %s", ip)
			}
			if _, _, err := net.ParseCIDR(ip); err != nil {
				if net.ParseIP(ip) == nil {
					return nil, fmt.Errorf("invalid AllowedIPs %s", ip)
				}
			}
		}
	}
	return summary, nil
}

func restartInterface(name string) (interface{}, error) {
	if !ifaceRx.MatchString(name) {
		return nil, fmt.Errorf("invalid interface name")
	}
	cmd := exec.Command("systemctl", "restart", fmt.Sprintf("wg-quick@%s", name))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s", sanitizeOutput(string(out)))
	}
	return map[string]string{"status": "ok"}, nil
}

func upInterface(name string) (interface{}, error) {
	if !ifaceRx.MatchString(name) {
		return nil, fmt.Errorf("invalid interface name")
	}
	cmd := exec.Command("systemctl", "start", fmt.Sprintf("wg-quick@%s", name))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s", sanitizeOutput(string(out)))
	}
	return map[string]string{"status": "ok"}, nil
}

func downInterface(name string) (interface{}, error) {
	if !ifaceRx.MatchString(name) {
		return nil, fmt.Errorf("invalid interface name")
	}
	cmd := exec.Command("systemctl", "stop", fmt.Sprintf("wg-quick@%s", name))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s", sanitizeOutput(string(out)))
	}
	return map[string]string{"status": "ok"}, nil
}

func getInterfaceStatus(name string) (interface{}, error) {
	if !ifaceRx.MatchString(name) {
		return nil, fmt.Errorf("invalid interface name")
	}
	unit := fmt.Sprintf("wg-quick@%s", name)
	out, err := exec.Command("systemctl", "show", unit, "--no-page", "--property=ActiveState,SubState,ActiveEnterTimestamp,InactiveEnterTimestamp").Output()
	if err != nil {
		return nil, err
	}
	info := make(map[string]string)
	for _, l := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		parts := strings.SplitN(l, "=", 2)
		if len(parts) == 2 {
			info[parts[0]] = parts[1]
		}
	}
	status := info["ActiveState"]
	if sub, ok := info["SubState"]; ok && sub != "" {
		status = fmt.Sprintf("%s (%s)", status, sub)
	}
	ts := info["InactiveEnterTimestamp"]
	if info["ActiveState"] == "active" {
		ts = info["ActiveEnterTimestamp"]
	}
	jOut, _ := exec.Command("journalctl", "-u", unit, "-n", "20", "--no-pager", "--output=cat").Output()
	msg := sanitizeOutput(string(jOut))
	return map[string]string{"status": status, "last_change": ts, "message": msg}, nil
}

type pkgManager struct {
	cmd         string
	installArgs []string
}

func detectPackageManager() (pkgManager, error) {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return pkgManager{}, err
	}
	info := make(map[string]string)
	lines := strings.Split(string(data), "\n")
	for _, l := range lines {
		parts := strings.SplitN(l, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			val := strings.Trim(parts[1], "\"")
			info[key] = val
		}
	}
	ids := []string{info["ID"], info["ID_LIKE"]}
	for _, id := range ids {
		if strings.Contains(id, "debian") || strings.Contains(id, "ubuntu") {
			return pkgManager{"apt-get", []string{"install", "-y"}}, nil
		}
		if strings.Contains(id, "fedora") || strings.Contains(id, "centos") || strings.Contains(id, "rhel") {
			return pkgManager{"dnf", []string{"install", "-y"}}, nil
		}
		if strings.Contains(id, "arch") {
			return pkgManager{"pacman", []string{"-S", "--noconfirm"}}, nil
		}
	}
	return pkgManager{}, errors.New("unsupported OS")
}

func checkPrereqs() (interface{}, error) {
	kernel := exec.Command("modprobe", "-n", "wireguard").Run() == nil
	_, err := exec.LookPath("wg")
	tools := err == nil
	systemd := exec.Command("systemctl", "list-unit-files", "wg-quick@.service").Run() == nil
	return map[string]bool{"kernel": kernel, "tools": tools, "systemd": systemd}, nil
}

func installPackages() (interface{}, error) {
	pm, err := detectPackageManager()
	if err != nil {
		return nil, err
	}
	pkgs := []string{"wireguard", "wireguard-tools"}
	args := append(pm.installArgs, pkgs...)
	cmd := exec.Command(pm.cmd, args...)
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	exec.Command("systemctl", "daemon-reload").Run()
	if _, err := exec.LookPath("wg"); err != nil {
		return nil, fmt.Errorf("wg binary not found after installation")
	}
	return map[string]string{"status": "ok"}, nil
}

func runSelfTest() (interface{}, error) {
	details := make(map[string]interface{})
	lines := []string{}

	kernel := exec.Command("modprobe", "-n", "wireguard").Run() == nil
	details["kernel"] = kernel
	lines = append(lines, fmt.Sprintf("Kernel module present: %v", kernel))

	_, err := exec.LookPath("wg")
	tools := err == nil
	details["tools"] = tools
	lines = append(lines, fmt.Sprintf("wg binary available: %v", tools))

	systemd := exec.Command("systemctl", "list-unit-files", "wg-quick@.service").Run() == nil
	details["systemd"] = systemd
	lines = append(lines, fmt.Sprintf("systemd units present: %v", systemd))

	ipv4, ipv6 := checkIPForwarding()
	details["ipForwarding"] = map[string]bool{"ipv4": ipv4, "ipv6": ipv6}
	lines = append(lines, fmt.Sprintf("IP forwarding (IPv4/IPv6): %v/%v", ipv4, ipv6))

	fw := checkFirewall()
	details["firewall"] = fw
	lines = append(lines, fmt.Sprintf("Firewall: %s", fw))

	conflicts, err := findRouteConflicts()
	if err != nil {
		details["routeConflictsError"] = err.Error()
		lines = append(lines, "Route conflicts: error checking")
	} else {
		details["routeConflicts"] = conflicts
		if len(conflicts) == 0 {
			lines = append(lines, "Route conflicts: none")
		} else {
			lines = append(lines, "Route conflicts detected: "+strings.Join(conflicts, ", "))
		}
	}

	clock := checkClockSync()
	details["clock"] = clock
	lines = append(lines, fmt.Sprintf("Clock sync: %s", clock))

	dummy := `[Interface]
Address = 10.0.0.1/32
PrivateKey = AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=

[Peer]
PublicKey = AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
AllowedIPs = 10.0.0.2/32
`
	if _, err := validateConfig(dummy); err != nil {
		details["validateConfig"] = err.Error()
		lines = append(lines, "Config validation: "+err.Error())
	} else {
		details["validateConfig"] = "ok"
		lines = append(lines, "Config validation: ok")
	}

	cmd := exec.Command("wg", "syncconf", "wgselftest0", "-")
	cmd.Stdin = strings.NewReader(dummy)
	out, err := cmd.CombinedOutput()
	if err != nil {
		details["syncconf"] = sanitizeOutput(string(out))
		lines = append(lines, "Syncconf: "+strings.TrimSpace(sanitizeOutput(string(out))))
	} else {
		details["syncconf"] = "ok"
		lines = append(lines, "Syncconf: ok")
	}

	return map[string]interface{}{"report": strings.Join(lines, "\n"), "details": details}, nil
}

func checkIPForwarding() (bool, bool) {
	ipv4 := readSysctl("/proc/sys/net/ipv4/ip_forward") == "1"
	ipv6 := readSysctl("/proc/sys/net/ipv6/conf/all/forwarding") == "1"
	return ipv4, ipv6
}

func readSysctl(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return "0"
	}
	return strings.TrimSpace(string(data))
}

func checkFirewall() string {
	if _, err := exec.LookPath("nft"); err == nil {
		if out, err := exec.Command("nft", "list", "ruleset").Output(); err == nil {
			if bytes.Contains(out, []byte("dport 51820")) && (bytes.Contains(out, []byte("drop")) || bytes.Contains(out, []byte("reject"))) {
				return "blocked"
			}
		}
	}
	if _, err := exec.LookPath("iptables"); err == nil {
		if out, err := exec.Command("iptables", "-S").Output(); err == nil {
			if bytes.Contains(out, []byte("--dport 51820")) && (bytes.Contains(out, []byte("DROP")) || bytes.Contains(out, []byte("REJECT"))) {
				return "blocked"
			}
		}
	}
	return "ok"
}

func findRouteConflicts() ([]string, error) {
	out, err := exec.Command("wg", "show", "all", "allowed-ips").Output()
	if err != nil {
		return nil, err
	}
	seen := make(map[string]string)
	conflicts := []string{}
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		iface := strings.TrimSuffix(parts[0], ":")
		for _, ip := range strings.Split(parts[1], ",") {
			ip = strings.TrimSpace(ip)
			if other, ok := seen[ip]; ok && other != iface {
				conflicts = append(conflicts, ip)
			} else {
				seen[ip] = iface
			}
		}
	}
	return conflicts, nil
}

func checkClockSync() string {
	if out, err := exec.Command("timedatectl", "show", "--property=NTPSynchronized", "--value").Output(); err == nil {
		if strings.TrimSpace(string(out)) == "yes" {
			return "synchronized"
		}
		return "unsynchronized"
	}
	return "unknown"
}

func auditApply(action, iface, step string, err error) {
	fields := map[string]interface{}{"action": action, "iface": iface}
	if step != "" {
		fields["step"] = step
	}
	if err != nil {
		fields["error"] = err.Error()
	}
	msgBytes, _ := json.Marshal(fields)
	journal.Send(string(msgBytes), journal.PriInfo, nil)
}

func auditLog(method string, params json.RawMessage, err error) {
	fields := make(map[string]interface{})
	json.Unmarshal(params, &fields)
	redact(fields)
	fields["method"] = method
	if err != nil {
		fields["error"] = err.Error()
	}
	msgBytes, _ := json.Marshal(fields)
	journal.Send(string(msgBytes), journal.PriInfo, nil)
}

func redact(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		for k, vv := range val {
			if isSecret(k) {
				val[k] = "<redacted>"
			} else {
				val[k] = redact(vv)
			}
		}
	case []interface{}:
		for i, vv := range val {
			val[i] = redact(vv)
		}
	}
	return v
}

func isSecret(key string) bool {
	k := strings.ToLower(key)
	return strings.Contains(k, "key") || strings.Contains(k, "password") || strings.Contains(k, "secret") || strings.Contains(k, "psk")
}
