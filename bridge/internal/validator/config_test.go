package validator

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"wg-bridge/internal/config"
)

func TestValidateConfigValid(t *testing.T) {
	validator := NewValidator(false)
	parser := config.NewParser(false)

	configText := `[Interface]
PrivateKey = yAnz5TF+lXXJte14tji3zlMNq+hd2rYUIgJBgB3fBmk=
Address = 10.192.122.1/24
ListenPort = 51820

[Peer]
PublicKey = xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=
AllowedIPs = 10.192.122.2/32`

	summary, err := parser.Parse(configText)
	if err != nil {
		t.Fatalf("Failed to parse valid config: %v", err)
	}

	err = validator.ValidateConfig(summary)
	if err != nil {
		t.Errorf("Expected valid config to pass validation, got: %v", err)
	}
}

func TestValidateInterfaceNameValid(t *testing.T) {
	validNames := []string{
		"wg0",
		"wg-vpn",
		"wireguard_tunnel",
		"wg.test",
		"123",
		"a",
		"very-long-name1",
	}

	for _, name := range validNames {
		t.Run(name, func(t *testing.T) {
			err := ValidateInterfaceName(name)
			if err != nil {
				t.Errorf("Expected valid interface name %q to pass, got: %v", name, err)
			}
		})
	}
}

func TestValidateInterfaceNameInvalid(t *testing.T) {
	invalidNames := []string{
		"",
		"wg@invalid",
		"toolonginterfacename",
		"wg space",
		"wg/invalid",
		"wg:invalid",
	}

	for _, name := range invalidNames {
		t.Run(name, func(t *testing.T) {
			err := ValidateInterfaceName(name)
			if err == nil {
				t.Errorf("Expected invalid interface name %q to fail", name)
			}
		})
	}
}

func TestValidateInterfaceMissingPrivateKey(t *testing.T) {
	validator := NewValidator(true)
	iface := map[string]string{
		"Address": "10.0.0.1/24",
	}

	err := validator.validateInterface(iface)
	if err == nil {
		t.Error("Expected interface without PrivateKey to fail")
	}
	if !strings.Contains(err.Error(), "PrivateKey") {
		t.Errorf("Expected error about missing PrivateKey, got: %v", err)
	}
}

func TestValidateInterfaceInvalidPrivateKey(t *testing.T) {
	validator := NewValidator(true)
	iface := map[string]string{
		"PrivateKey": "invalid_key_format",
	}

	err := validator.validateInterface(iface)
	if err == nil {
		t.Error("Expected interface with invalid PrivateKey to fail")
	}
	if !strings.Contains(err.Error(), "PrivateKey format") {
		t.Errorf("Expected error about PrivateKey format, got: %v", err)
	}
}

func TestValidateInterfaceValidPrivateKey(t *testing.T) {
	validator := NewValidator(true)
	iface := map[string]string{
		"PrivateKey": "yAnz5TF+lXXJte14tji3zlMNq+hd2rYUIgJBgB3fBmk=",
	}

	err := validator.validateInterface(iface)
	if err != nil {
		t.Errorf("Expected interface with valid PrivateKey to pass, got: %v", err)
	}
}

func TestValidateInterfaceInvalidAddress(t *testing.T) {
	validator := NewValidator(true)
	iface := map[string]string{
		"PrivateKey": "yAnz5TF+lXXJte14tji3zlMNq+hd2rYUIgJBgB3fBmk=",
		"Address":    "invalid.address",
	}

	err := validator.validateInterface(iface)
	if err == nil {
		t.Error("Expected interface with invalid Address to fail")
	}
}

func TestValidateInterfaceValidAddress(t *testing.T) {
	validator := NewValidator(true)
	validAddresses := []string{
		"10.0.0.1/24",
		"192.168.1.1/32",
		"10.0.0.1/24, 2001:db8::1/64",
		"192.168.1.1",
	}

	for _, addr := range validAddresses {
		t.Run(addr, func(t *testing.T) {
			iface := map[string]string{
				"PrivateKey": "yAnz5TF+lXXJte14tji3zlMNq+hd2rYUIgJBgB3fBmk=",
				"Address":    addr,
			}

			err := validator.validateInterface(iface)
			if err != nil {
				t.Errorf("Expected interface with valid Address %q to pass, got: %v", addr, err)
			}
		})
	}
}

func TestValidateInterfaceInvalidPort(t *testing.T) {
	validator := NewValidator(true)
	invalidPorts := []string{
		"0",
		"65536",
		"99999",
		"invalid",
		"-1",
	}

	for _, port := range invalidPorts {
		t.Run(port, func(t *testing.T) {
			iface := map[string]string{
				"PrivateKey": "yAnz5TF+lXXJte14tji3zlMNq+hd2rYUIgJBgB3fBmk=",
				"ListenPort": port,
			}

			err := validator.validateInterface(iface)
			if err == nil {
				t.Errorf("Expected interface with invalid port %q to fail", port)
			}
		})
	}
}

func TestValidateInterfaceValidPort(t *testing.T) {
	validator := NewValidator(true)
	validPorts := []string{
		"1",
		"51820",
		"65535",
		"1234",
	}

	for _, port := range validPorts {
		t.Run(port, func(t *testing.T) {
			iface := map[string]string{
				"PrivateKey": "yAnz5TF+lXXJte14tji3zlMNq+hd2rYUIgJBgB3fBmk=",
				"ListenPort": port,
			}

			err := validator.validateInterface(iface)
			if err != nil {
				t.Errorf("Expected interface with valid port %q to pass, got: %v", port, err)
			}
		})
	}
}

func TestValidatePeerMissingPublicKey(t *testing.T) {
	validator := NewValidator(true)
	peer := map[string]string{
		"AllowedIPs": "10.0.0.1/32",
	}

	err := validator.validatePeer(peer, 0)
	if err == nil {
		t.Error("Expected peer without PublicKey to fail")
	}
	if !strings.Contains(err.Error(), "PublicKey") {
		t.Errorf("Expected error about missing PublicKey, got: %v", err)
	}
}

func TestValidatePeerInvalidPublicKey(t *testing.T) {
	validator := NewValidator(true)
	peer := map[string]string{
		"PublicKey":  "invalid_key",
		"AllowedIPs": "10.0.0.1/32",
	}

	err := validator.validatePeer(peer, 0)
	if err == nil {
		t.Error("Expected peer with invalid PublicKey to fail")
	}
}

func TestValidatePeerValidPublicKey(t *testing.T) {
	validator := NewValidator(true)
	peer := map[string]string{
		"PublicKey":  "xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=",
		"AllowedIPs": "10.0.0.1/32",
	}

	err := validator.validatePeer(peer, 0)
	if err != nil {
		t.Errorf("Expected peer with valid PublicKey to pass, got: %v", err)
	}
}

func TestValidatePeerInvalidPresharedKey(t *testing.T) {
	validator := NewValidator(true)
	peer := map[string]string{
		"PublicKey":    "xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=",
		"PresharedKey": "invalid_key",
		"AllowedIPs":   "10.0.0.1/32",
	}

	err := validator.validatePeer(peer, 0)
	if err == nil {
		t.Error("Expected peer with invalid PresharedKey to fail")
	}
}

func TestValidatePeerValidPresharedKey(t *testing.T) {
	validator := NewValidator(true)
	peer := map[string]string{
		"PublicKey":    "xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=",
		"PresharedKey": "FpCyhws9cxwWoV4xELtfJvjJN+zQVRPISllRWgeopVE=",
		"AllowedIPs":   "10.0.0.1/32",
	}

	err := validator.validatePeer(peer, 0)
	if err != nil {
		t.Errorf("Expected peer with valid PresharedKey to pass, got: %v", err)
	}
}

func TestValidatePeerInvalidEndpoint(t *testing.T) {
	validator := NewValidator(true)
	invalidEndpoints := []string{
		"invalid",
		"example.com",
		"192.168.1.1",
		"192.168.1.1:99999",
		"invalid:port:format",
	}

	for _, endpoint := range invalidEndpoints {
		t.Run(endpoint, func(t *testing.T) {
			peer := map[string]string{
				"PublicKey":  "xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=",
				"AllowedIPs": "10.0.0.1/32",
				"Endpoint":   endpoint,
			}

			err := validator.validatePeer(peer, 0)
			if err == nil {
				t.Errorf("Expected peer with invalid endpoint %q to fail", endpoint)
			}
		})
	}
}

func TestValidatePeerValidEndpoint(t *testing.T) {
	validator := NewValidator(true)
	validEndpoints := []string{
		"example.com:51820",
		"192.168.1.1:51820",
		"[2001:db8::1]:51820",
		"test-host.example.org:1234",
	}

	for _, endpoint := range validEndpoints {
		t.Run(endpoint, func(t *testing.T) {
			peer := map[string]string{
				"PublicKey":  "xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=",
				"AllowedIPs": "10.0.0.1/32",
				"Endpoint":   endpoint,
			}

			err := validator.validatePeer(peer, 0)
			if err != nil {
				t.Errorf("Expected peer with valid endpoint %q to pass, got: %v", endpoint, err)
			}
		})
	}
}

func TestValidatePeerInvalidKeepalive(t *testing.T) {
	validator := NewValidator(true)
	invalidKeepalives := []string{
		"-1",
		"99999",
		"invalid",
	}

	for _, keepalive := range invalidKeepalives {
		t.Run(keepalive, func(t *testing.T) {
			peer := map[string]string{
				"PublicKey":           "xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=",
				"AllowedIPs":          "10.0.0.1/32",
				"PersistentKeepalive": keepalive,
			}

			err := validator.validatePeer(peer, 0)
			if err == nil {
				t.Errorf("Expected peer with invalid keepalive %q to fail", keepalive)
			}
		})
	}
}

func TestValidatePeerValidKeepalive(t *testing.T) {
	validator := NewValidator(true)
	validKeepalives := []string{
		"0",
		"25",
		"60",
		"65535",
	}

	for _, keepalive := range validKeepalives {
		t.Run(keepalive, func(t *testing.T) {
			peer := map[string]string{
				"PublicKey":           "xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=",
				"AllowedIPs":          "10.0.0.1/32",
				"PersistentKeepalive": keepalive,
			}

			err := validator.validatePeer(peer, 0)
			if err != nil {
				t.Errorf("Expected peer with valid keepalive %q to pass, got: %v", keepalive, err)
			}
		})
	}
}

func TestValidateConfigDuplicatePublicKeys(t *testing.T) {
	validator := NewValidator(true)
	parser := config.NewParser(true)

	configText := `[Interface]
PrivateKey = yAnz5TF+lXXJte14tji3zlMNq+hd2rYUIgJBgB3fBmk=

[Peer]
PublicKey = xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=
AllowedIPs = 10.0.0.1/32

[Peer]
PublicKey = xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=
AllowedIPs = 10.0.0.2/32`

	summary, err := parser.Parse(configText)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	err = validator.ValidateConfig(summary)
	if err == nil {
		t.Error("Expected config with duplicate PublicKeys to fail")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("Expected error about duplicate keys, got: %v", err)
	}
}

func TestValidateConfigFromTestData(t *testing.T) {
	validator := NewValidator(false)
	parser := config.NewParser(false)

	// Test valid configurations
	validConfigs := []string{
		"valid_basic.conf",
		"valid_complex.conf",
	}

	for _, filename := range validConfigs {
		t.Run(filename, func(t *testing.T) {
			content, err := ioutil.ReadFile(filepath.Join("../../testdata", filename))
			if err != nil {
				t.Skipf("Could not read test file %s: %v", filename, err)
				return
			}

			summary, err := parser.Parse(string(content))
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", filename, err)
			}

			err = validator.ValidateConfig(summary)
			if err != nil {
				t.Errorf("Expected valid config %s to pass validation, got: %v", filename, err)
			}
		})
	}

	// Test invalid configurations
	invalidConfigs := []string{
		"invalid_duplicate_keys.conf",
		"invalid_overlapping_ips.conf",
		"invalid_malformed.conf",
	}

	for _, filename := range invalidConfigs {
		t.Run(filename, func(t *testing.T) {
			content, err := ioutil.ReadFile(filepath.Join("../../testdata", filename))
			if err != nil {
				t.Skipf("Could not read test file %s: %v", filename, err)
				return
			}

			summary, err := parser.Parse(string(content))
			if err != nil {
				// Some files might fail parsing, which is expected
				return
			}

			err = validator.ValidateConfig(summary)
			if err == nil {
				t.Errorf("Expected invalid config %s to fail validation", filename)
			}
		})
	}
}

func TestValidateConfigStrictMode(t *testing.T) {
	strictValidator := NewValidator(true)
	parser := config.NewParser(true)

	// Test catch-all routes in strict mode
	content, err := ioutil.ReadFile(filepath.Join("../../testdata", "catchall_routes.conf"))
	if err != nil {
		t.Skip("Could not read catch-all test file")
		return
	}

	summary, err := parser.Parse(string(content))
	if err != nil {
		t.Fatalf("Failed to parse catch-all config: %v", err)
	}

	err = strictValidator.ValidateConfig(summary)
	if err == nil {
		t.Error("Expected catch-all routes to fail in strict mode")
	}
}

// Benchmark tests
func BenchmarkValidateConfig(b *testing.B) {
	validator := NewValidator(false)
	parser := config.NewParser(false)

	configText := `[Interface]
PrivateKey = yAnz5TF+lXXJte14tji3zlMNq+hd2rYUIgJBgB3fBmk=
Address = 10.192.122.1/24

[Peer]
PublicKey = xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=
AllowedIPs = 10.192.122.2/32`

	summary, err := parser.Parse(configText)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := validator.ValidateConfig(summary)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidateInterfaceName(b *testing.B) {
	name := "wg0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := ValidateInterfaceName(name)
		if err != nil {
			b.Fatal(err)
		}
	}
}
