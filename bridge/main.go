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
	case "InstallPackage":
		var p struct {
			Name string `json:"name"`
		}
		if err = json.Unmarshal(req.Params, &p); err == nil {
			result, err = installPackage(p.Name)
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
	if _, err := exec.LookPath("apt-get"); err == nil {
		return pkgManager{"apt-get", []string{"install", "-y"}}, nil
	}
	if _, err := exec.LookPath("dnf"); err == nil {
		return pkgManager{"dnf", []string{"install", "-y"}}, nil
	}
	if _, err := exec.LookPath("pacman"); err == nil {
		return pkgManager{"pacman", []string{"-S", "--noconfirm"}}, nil
	}
	return pkgManager{}, errors.New("no supported package manager found")
}

func installPackage(name string) (interface{}, error) {
	if name == "" {
		return nil, fmt.Errorf("empty package name")
	}
	pm, err := detectPackageManager()
	if err != nil {
		return nil, err
	}
	args := append(pm.installArgs, name)
	cmd := exec.Command(pm.cmd, args...)
	if err := cmd.Run(); err != nil {
		return nil, err
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
