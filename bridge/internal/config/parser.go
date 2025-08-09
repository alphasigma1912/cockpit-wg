package config

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// Summary represents a parsed WireGuard configuration
type Summary struct {
	Interface map[string]string   `json:"interface"`
	Peers     []map[string]string `json:"peers"`
}

// Parser handles WireGuard configuration parsing
type Parser struct {
	// strictMode enables additional validation rules
	strictMode bool
}

// NewParser creates a new configuration parser
func NewParser(strict bool) *Parser {
	return &Parser{strictMode: strict}
}

// Parse parses a WireGuard configuration from text
func (p *Parser) Parse(text string) (*Summary, error) {
	summary := &Summary{Interface: make(map[string]string)}
	var curr map[string]string
	scanner := bufio.NewScanner(strings.NewReader(text))
	line := 0

	for scanner.Scan() {
		line++
		l := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if l == "" || strings.HasPrefix(l, "#") || strings.HasPrefix(l, ";") {
			continue
		}

		// Handle section headers
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

		// Handle key-value pairs
		if curr == nil {
			return nil, fmt.Errorf("key-value outside of section at line %d", line)
		}

		parts := strings.SplitN(l, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid line %d: %s", line, l)
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		if key == "" {
			return nil, fmt.Errorf("empty key at line %d", line)
		}

		curr[key] = val
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	// Basic validation
	if len(summary.Interface) == 0 {
		return nil, fmt.Errorf("missing [Interface] section")
	}

	if p.strictMode && len(summary.Peers) == 0 {
		return nil, fmt.Errorf("no peers defined")
	}

	return summary, nil
}

// ValidateIPs validates AllowedIPs format and constraints
func ValidateIPs(allowedIPs string, strictMode bool) error {
	if allowedIPs == "" {
		return fmt.Errorf("empty AllowedIPs")
	}

	ips := strings.Split(allowedIPs, ",")
	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}

		// Check for disallowed catch-all routes
		if strictMode && (ip == "0.0.0.0/0" || ip == "::/0") {
			return fmt.Errorf("disallowed AllowedIPs %s", ip)
		}

		// Try CIDR first, then single IP
		if _, _, err := net.ParseCIDR(ip); err != nil {
			if net.ParseIP(ip) == nil {
				return fmt.Errorf("invalid AllowedIPs %s", ip)
			}
		}
	}

	return nil
}

// DetectIPConflicts checks for overlapping CIDR ranges
func DetectIPConflicts(peers []map[string]string) error {
	var ranges []*net.IPNet
	var ips []net.IP

	for peerIdx, peer := range peers {
		pk, ok := peer["PublicKey"]
		if !ok || pk == "" {
			return fmt.Errorf("peer %d missing PublicKey", peerIdx)
		}

		allowed, ok := peer["AllowedIPs"]
		if !ok {
			return fmt.Errorf("peer %s missing AllowedIPs", pk)
		}

		ipList := strings.Split(allowed, ",")
		for _, ip := range ipList {
			ip = strings.TrimSpace(ip)
			if ip == "" {
				continue
			}

			// Parse CIDR
			if _, network, err := net.ParseCIDR(ip); err == nil {
				// Check for conflicts with existing ranges
				for _, existing := range ranges {
					if networksOverlap(network, existing) {
						return fmt.Errorf("overlapping AllowedIPs: %s conflicts with existing range", ip)
					}
				}
				ranges = append(ranges, network)
			} else if parsedIP := net.ParseIP(ip); parsedIP != nil {
				// Check single IP against ranges and other IPs
				for _, existing := range ranges {
					if existing.Contains(parsedIP) {
						return fmt.Errorf("IP %s conflicts with existing range %s", ip, existing.String())
					}
				}
				for _, existingIP := range ips {
					if parsedIP.Equal(existingIP) {
						return fmt.Errorf("duplicate IP %s", ip)
					}
				}
				ips = append(ips, parsedIP)
			}
		}
	}

	return nil
}

// networksOverlap checks if two CIDR networks overlap
func networksOverlap(a, b *net.IPNet) bool {
	return a.Contains(b.IP) || b.Contains(a.IP)
}
