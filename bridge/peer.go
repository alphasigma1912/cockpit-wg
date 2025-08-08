package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
)

type peerParams struct {
	Endpoint            string   `json:"endpoint"`
	AllowedIPs          []string `json:"allowed_ips"`
	PersistentKeepalive int      `json:"persistent_keepalive"`
	Preshared           bool     `json:"preshared"`
	Enabled             bool     `json:"enabled"`
}

func addPeer(name string, p peerParams) (interface{}, error) {
	allowed, err := normalizeCIDRs(p.AllowedIPs)
	if err != nil {
		return nil, err
	}

	priv, pub, err := genKeyPair()
	if err != nil {
		return nil, err
	}
	psk := ""
	if p.Preshared {
		if psk, err = genPSK(); err != nil {
			return nil, err
		}
	}
	lines := []string{"[Peer]", fmt.Sprintf("PublicKey = %s", pub)}
	if psk != "" {
		lines = append(lines, fmt.Sprintf("PresharedKey = %s", psk))
	}
	if p.Endpoint != "" {
		lines = append(lines, fmt.Sprintf("Endpoint = %s", strings.TrimSpace(p.Endpoint)))
	}
	if len(allowed) > 0 {
		lines = append(lines, fmt.Sprintf("AllowedIPs = %s", strings.Join(allowed, ", ")))
	}
	if p.PersistentKeepalive > 0 {
		lines = append(lines, fmt.Sprintf("PersistentKeepalive = %d", p.PersistentKeepalive))
	}
	block := strings.Join(lines, "\n") + "\n"
	if !p.Enabled {
		block = commentBlock(block)
	}

	path := fmt.Sprintf("/etc/wireguard/%s.conf", name)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if _, err := f.WriteString("\n" + block); err != nil {
		return nil, err
	}
	return map[string]string{"publicKey": pub, "privateKey": priv, "presharedKey": psk}, nil
}

func normalizeCIDRs(list []string) ([]string, error) {
	out := []string{}
	for _, ip := range list {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}
		_, nw, err := net.ParseCIDR(ip)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR %s", ip)
		}
		out = append(out, nw.String())
	}
	return out, nil
}

func genKeyPair() (string, string, error) {
	privBytes, err := exec.Command("wg", "genkey").Output()
	if err != nil {
		return "", "", err
	}
	priv := strings.TrimSpace(string(privBytes))
	cmd := exec.Command("wg", "pubkey")
	cmd.Stdin = strings.NewReader(priv)
	pubBytes, err := cmd.Output()
	if err != nil {
		return "", "", err
	}
	return priv, strings.TrimSpace(string(pubBytes)), nil
}

func genPSK() (string, error) {
	b, err := exec.Command("wg", "genpsk").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func commentBlock(block string) string {
	lines := strings.Split(block, "\n")
	for i, l := range lines {
		if strings.TrimSpace(l) != "" {
			lines[i] = "# " + l
		}
	}
	return strings.Join(lines, "\n")
}

func removePeer(name, pub string) (interface{}, error) {
	path := fmt.Sprintf("/etc/wireguard/%s.conf", name)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	sections := strings.Split(string(data), "\n[Peer]")
	result := sections[0]
	for i := 1; i < len(sections); i++ {
		blk := sections[i]
		if strings.Contains(blk, fmt.Sprintf("PublicKey = %s", pub)) || strings.Contains(blk, fmt.Sprintf("PublicKey=%s", pub)) {
			continue
		}
		result += "\n[Peer]" + blk
	}
	if err := os.WriteFile(path, []byte(result), 0600); err != nil {
		return nil, err
	}
	return map[string]string{"status": "ok"}, nil
}

func updatePeer(name, pub string, p peerParams) (interface{}, error) {
	if _, err := removePeer(name, pub); err != nil {
		return nil, err
	}
	// use existing public key
	p.Preshared = false
	res, err := addPeerWithKey(name, pub, p)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func addPeerWithKey(name, pub string, p peerParams) (interface{}, error) {
	allowed, err := normalizeCIDRs(p.AllowedIPs)
	if err != nil {
		return nil, err
	}
	psk := ""
	lines := []string{"[Peer]", fmt.Sprintf("PublicKey = %s", pub)}
	if psk != "" {
		lines = append(lines, fmt.Sprintf("PresharedKey = %s", psk))
	}
	if p.Endpoint != "" {
		lines = append(lines, fmt.Sprintf("Endpoint = %s", strings.TrimSpace(p.Endpoint)))
	}
	if len(allowed) > 0 {
		lines = append(lines, fmt.Sprintf("AllowedIPs = %s", strings.Join(allowed, ", ")))
	}
	if p.PersistentKeepalive > 0 {
		lines = append(lines, fmt.Sprintf("PersistentKeepalive = %d", p.PersistentKeepalive))
	}
	block := strings.Join(lines, "\n") + "\n"
	if !p.Enabled {
		block = commentBlock(block)
	}
	path := fmt.Sprintf("/etc/wireguard/%s.conf", name)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if _, err := f.WriteString("\n" + block); err != nil {
		return nil, err
	}
	return map[string]string{"publicKey": pub}, nil
}

func listPeers(name string) (interface{}, error) {
	path := fmt.Sprintf("/etc/wireguard/%s.conf", name)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []map[string]interface{}{}, nil
		}
		return nil, err
	}
	sections := strings.Split(string(data), "\n[Peer]")
	peers := []map[string]interface{}{}
	for i := 1; i < len(sections); i++ {
		blk := sections[i]
		enabled := true
		if strings.HasPrefix(strings.TrimSpace(blk), "#") {
			enabled = false
		}
		lines := strings.Split(blk, "\n")
		info := map[string]interface{}{"enabled": enabled}
		for _, l := range lines {
			l = strings.TrimSpace(strings.TrimPrefix(l, "#"))
			if l == "" {
				continue
			}
			parts := strings.SplitN(l, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			switch key {
			case "PublicKey":
				info["publicKey"] = val
			case "Endpoint":
				info["endpoint"] = val
			case "AllowedIPs":
				info["allowedIPs"] = strings.Split(val, ",")
			case "PersistentKeepalive":
				info["persistentKeepalive"] = val
			case "PresharedKey":
				// never expose actual PSK
				info["presharedKey"] = true
			}
		}
		if _, ok := info["publicKey"]; ok {
			peers = append(peers, info)
		}
	}
	return peers, nil
}
