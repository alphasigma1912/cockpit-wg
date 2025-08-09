package validator

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"wg-bridge/internal/config"
)

var (
	// Common validation patterns
	ifaceNameRx    = regexp.MustCompile(`^[a-zA-Z0-9._-]{1,15}$`)
	publicKeyRx    = regexp.MustCompile(`^[A-Za-z0-9+/]{42}[A-Ea-e048]==$`)
	privateKeyRx   = regexp.MustCompile(`^[A-Za-z0-9+/]{42}[A-Ea-e048]==$`)
	presharedKeyRx = regexp.MustCompile(`^[A-Za-z0-9+/]{42}[A-Ea-e048]==$`)
)

// ConfigValidator validates WireGuard configurations
type ConfigValidator struct {
	strictMode    bool
	maxPeers      int
	allowCatchAll bool
	requiredKeys  map[string][]string // section -> required keys
}

// NewValidator creates a new configuration validator
func NewValidator(strict bool) *ConfigValidator {
	return &ConfigValidator{
		strictMode:    strict,
		maxPeers:      100,
		allowCatchAll: !strict,
		requiredKeys: map[string][]string{
			"Interface": {"PrivateKey"},
			"Peer":      {"PublicKey", "AllowedIPs"},
		},
	}
}

// ValidateConfig performs comprehensive validation of a WireGuard config
func (v *ConfigValidator) ValidateConfig(summary *config.Summary) error {
	if err := v.validateInterface(summary.Interface); err != nil {
		return fmt.Errorf("interface validation failed: %w", err)
	}

	if err := v.validatePeers(summary.Peers); err != nil {
		return fmt.Errorf("peer validation failed: %w", err)
	}

	if err := config.DetectIPConflicts(summary.Peers); err != nil {
		return fmt.Errorf("IP conflict detected: %w", err)
	}

	return nil
}

// validateInterface validates the [Interface] section
func (v *ConfigValidator) validateInterface(iface map[string]string) error {
	// Check required keys
	for _, key := range v.requiredKeys["Interface"] {
		if _, ok := iface[key]; !ok {
			return fmt.Errorf("missing required key: %s", key)
		}
	}

	// Validate PrivateKey
	if privateKey, ok := iface["PrivateKey"]; ok {
		if !privateKeyRx.MatchString(privateKey) {
			return fmt.Errorf("invalid PrivateKey format")
		}
	}

	// Validate Address
	if address, ok := iface["Address"]; ok {
		if err := v.validateAddresses(address); err != nil {
			return fmt.Errorf("invalid Address: %w", err)
		}
	}

	// Validate ListenPort
	if port, ok := iface["ListenPort"]; ok {
		if err := v.validatePort(port); err != nil {
			return fmt.Errorf("invalid ListenPort: %w", err)
		}
	}

	// Validate DNS
	if dns, ok := iface["DNS"]; ok {
		if err := v.validateDNS(dns); err != nil {
			return fmt.Errorf("invalid DNS: %w", err)
		}
	}

	// Validate MTU
	if mtu, ok := iface["MTU"]; ok {
		if err := v.validateMTU(mtu); err != nil {
			return fmt.Errorf("invalid MTU: %w", err)
		}
	}

	return nil
}

// validatePeers validates all peer configurations
func (v *ConfigValidator) validatePeers(peers []map[string]string) error {
	if v.strictMode && len(peers) == 0 {
		return fmt.Errorf("no peers defined")
	}

	if len(peers) > v.maxPeers {
		return fmt.Errorf("too many peers: %d (max %d)", len(peers), v.maxPeers)
	}

	seen := make(map[string]struct{})

	for i, peer := range peers {
		if err := v.validatePeer(peer, i); err != nil {
			return err
		}

		// Check for duplicate public keys
		pk := peer["PublicKey"]
		if _, dup := seen[pk]; dup {
			return fmt.Errorf("duplicate peer PublicKey: %s", pk)
		}
		seen[pk] = struct{}{}
	}

	return nil
}

// validatePeer validates a single peer configuration
func (v *ConfigValidator) validatePeer(peer map[string]string, index int) error {
	// Check required keys
	for _, key := range v.requiredKeys["Peer"] {
		if _, ok := peer[key]; !ok {
			return fmt.Errorf("peer %d missing required key: %s", index, key)
		}
	}

	// Validate PublicKey
	if publicKey, ok := peer["PublicKey"]; ok {
		if !publicKeyRx.MatchString(publicKey) {
			return fmt.Errorf("peer %d: invalid PublicKey format", index)
		}
	}

	// Validate PresharedKey
	if presharedKey, ok := peer["PresharedKey"]; ok && presharedKey != "" {
		if !presharedKeyRx.MatchString(presharedKey) {
			return fmt.Errorf("peer %d: invalid PresharedKey format", index)
		}
	}

	// Validate AllowedIPs
	if allowedIPs, ok := peer["AllowedIPs"]; ok {
		if err := config.ValidateIPs(allowedIPs, v.strictMode); err != nil {
			return fmt.Errorf("peer %d: %w", index, err)
		}
	}

	// Validate Endpoint
	if endpoint, ok := peer["Endpoint"]; ok && endpoint != "" {
		if err := v.validateEndpoint(endpoint); err != nil {
			return fmt.Errorf("peer %d: invalid Endpoint: %w", index, err)
		}
	}

	// Validate PersistentKeepalive
	if keepalive, ok := peer["PersistentKeepalive"]; ok && keepalive != "" {
		if err := v.validateKeepalive(keepalive); err != nil {
			return fmt.Errorf("peer %d: invalid PersistentKeepalive: %w", index, err)
		}
	}

	return nil
}

// ValidateInterfaceName validates WireGuard interface names
func ValidateInterfaceName(name string) error {
	if !ifaceNameRx.MatchString(name) {
		return fmt.Errorf("invalid interface name: %s", name)
	}
	return nil
}

// validateAddresses validates Address field (comma-separated IPs/CIDRs)
func (v *ConfigValidator) validateAddresses(addresses string) error {
	addrs := strings.Split(addresses, ",")
	for _, addr := range addrs {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			continue
		}

		// Must be valid CIDR or IP
		if _, _, err := net.ParseCIDR(addr); err != nil {
			if net.ParseIP(addr) == nil {
				return fmt.Errorf("invalid address: %s", addr)
			}
		}
	}
	return nil
}

// validatePort validates port numbers
func (v *ConfigValidator) validatePort(port string) error {
	p, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("invalid port number: %s", port)
	}
	if p < 1 || p > 65535 {
		return fmt.Errorf("port out of range: %d", p)
	}
	return nil
}

// validateDNS validates DNS server addresses
func (v *ConfigValidator) validateDNS(dns string) error {
	servers := strings.Split(dns, ",")
	for _, server := range servers {
		server = strings.TrimSpace(server)
		if server == "" {
			continue
		}
		if net.ParseIP(server) == nil {
			return fmt.Errorf("invalid DNS server: %s", server)
		}
	}
	return nil
}

// validateMTU validates MTU values
func (v *ConfigValidator) validateMTU(mtu string) error {
	m, err := strconv.Atoi(mtu)
	if err != nil {
		return fmt.Errorf("invalid MTU: %s", mtu)
	}
	if m < 576 || m > 65535 {
		return fmt.Errorf("MTU out of range: %d (576-65535)", m)
	}
	return nil
}

// validateEndpoint validates peer endpoints
func (v *ConfigValidator) validateEndpoint(endpoint string) error {
	// Format: host:port or [host]:port for IPv6
	parts := strings.Split(endpoint, ":")
	if len(parts) < 2 {
		return fmt.Errorf("endpoint missing port: %s", endpoint)
	}

	// Extract port (last part)
	port := parts[len(parts)-1]
	if err := v.validatePort(port); err != nil {
		return err
	}

	// Extract host
	host := strings.Join(parts[:len(parts)-1], ":")
	host = strings.Trim(host, "[]") // Remove IPv6 brackets

	// Validate host (IP or hostname)
	if net.ParseIP(host) == nil {
		// If not an IP, should be a valid hostname
		if len(host) == 0 || len(host) > 253 {
			return fmt.Errorf("invalid hostname length: %s", host)
		}
	}

	return nil
}

// validateKeepalive validates PersistentKeepalive values
func (v *ConfigValidator) validateKeepalive(keepalive string) error {
	k, err := strconv.Atoi(keepalive)
	if err != nil {
		return fmt.Errorf("invalid keepalive: %s", keepalive)
	}
	if k < 0 || k > 65535 {
		return fmt.Errorf("keepalive out of range: %d (0-65535)", k)
	}
	return nil
}
