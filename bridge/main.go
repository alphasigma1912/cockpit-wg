package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/coreos/go-systemd/v22/journal"
	"golang.zx2c4.com/wireguard/wgctrl"
)

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
	client, err := wgctrl.New()
	if err != nil {
		panic(err)
	}
	defer client.Close()

	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	for scanner.Scan() {
		line := scanner.Bytes()
		var req request
		if err := json.Unmarshal(line, &req); err != nil {
			continue
		}
		res := handleRequest(client, &req)
		b, _ := json.Marshal(res)
		writer.Write(b)
		writer.WriteByte('\n')
		writer.Flush()
	}
}

func handleRequest(client *wgctrl.Client, req *request) *response {
	var result interface{}
	var err error
	switch req.Method {
	case "ListInterfaces":
		result, err = listInterfaces(client)
	case "RestartInterface":
		var p struct {
			Name string `json:"name"`
		}
		if err = json.Unmarshal(req.Params, &p); err == nil {
			result, err = restartInterface(p.Name)
		}
	case "CheckPrereqs":
		result, err = checkPrereqs()
	case "InstallPackages":
		result, err = installPackages()
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

func listInterfaces(client *wgctrl.Client) (interface{}, error) {
	devices, err := client.Devices()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(devices))
	for _, d := range devices {
		names = append(names, d.Name)
	}
	return map[string]interface{}{"interfaces": names}, nil
}

var ifaceRx = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func restartInterface(name string) (interface{}, error) {
	if !ifaceRx.MatchString(name) {
		return nil, fmt.Errorf("invalid interface name")
	}
	cmd := exec.Command("systemctl", "restart", fmt.Sprintf("wg-quick@%s", name))
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return map[string]string{"status": "ok"}, nil
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
