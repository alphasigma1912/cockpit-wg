package config

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseValidBasicConfig(t *testing.T) {
	parser := NewParser(true)
	configText := `[Interface]
PrivateKey = yAnz5TF+lXXJte14tji3zlMNq+hd2rYUIgJBgB3fBmk=
Address = 10.192.122.1/24

[Peer]
PublicKey = xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=
AllowedIPs = 10.192.122.2/32`

	summary, err := parser.Parse(configText)
	if err != nil {
		t.Fatalf("Expected valid config to parse, got error: %v", err)
	}

	if len(summary.Interface) == 0 {
		t.Fatal("Expected Interface section to be parsed")
	}

	if summary.Interface["PrivateKey"] != "yAnz5TF+lXXJte14tji3zlMNq+hd2rYUIgJBgB3fBmk=" {
		t.Errorf("Expected PrivateKey to be parsed correctly")
	}

	if len(summary.Peers) != 1 {
		t.Fatalf("Expected 1 peer, got %d", len(summary.Peers))
	}

	peer := summary.Peers[0]
	if peer["PublicKey"] != "xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=" {
		t.Errorf("Expected PublicKey to be parsed correctly")
	}
}

func TestParseEmptyConfig(t *testing.T) {
	parser := NewParser(true)
	_, err := parser.Parse("")
	if err == nil {
		t.Fatal("Expected empty config to fail parsing")
	}
	if !strings.Contains(err.Error(), "missing [Interface] section") {
		t.Errorf("Expected specific error about missing Interface section, got: %v", err)
	}
}

func TestParseConfigWithComments(t *testing.T) {
	parser := NewParser(false)
	configText := `# This is a comment
[Interface]
PrivateKey = yAnz5TF+lXXJte14tji3zlMNq+hd2rYUIgJBgB3fBmk=
; This is also a comment
Address = 10.192.122.1/24

# Peer section
[Peer]
PublicKey = xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=
AllowedIPs = 10.192.122.2/32`

	summary, err := parser.Parse(configText)
	if err != nil {
		t.Fatalf("Expected config with comments to parse, got error: %v", err)
	}

	if len(summary.Interface) == 0 {
		t.Fatal("Expected Interface section to be parsed")
	}
}

func TestParseConfigMissingInterface(t *testing.T) {
	parser := NewParser(true)
	configText := `[Peer]
PublicKey = xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=
AllowedIPs = 10.192.122.2/32`

	_, err := parser.Parse(configText)
	if err == nil {
		t.Fatal("Expected config without Interface section to fail")
	}
	if !strings.Contains(err.Error(), "missing [Interface] section") {
		t.Errorf("Expected error about missing Interface section, got: %v", err)
	}
}

func TestParseConfigNoPeersStrictMode(t *testing.T) {
	parser := NewParser(true)
	configText := `[Interface]
PrivateKey = yAnz5TF+lXXJte14tji3zlMNq+hd2rYUIgJBgB3fBmk=
Address = 10.192.122.1/24`

	_, err := parser.Parse(configText)
	if err == nil {
		t.Fatal("Expected config without peers to fail in strict mode")
	}
	if !strings.Contains(err.Error(), "no peers defined") {
		t.Errorf("Expected error about no peers, got: %v", err)
	}
}

func TestParseConfigNoPeersNonStrictMode(t *testing.T) {
	parser := NewParser(false)
	configText := `[Interface]
PrivateKey = yAnz5TF+lXXJte14tji3zlMNq+hd2rYUIgJBgB3fBmk=
Address = 10.192.122.1/24`

	summary, err := parser.Parse(configText)
	if err != nil {
		t.Fatalf("Expected config without peers to succeed in non-strict mode, got: %v", err)
	}
	if len(summary.Peers) != 0 {
		t.Errorf("Expected no peers, got %d", len(summary.Peers))
	}
}

func TestParseConfigInvalidSection(t *testing.T) {
	parser := NewParser(true)
	configText := `[Interface]
PrivateKey = yAnz5TF+lXXJte14tji3zlMNq+hd2rYUIgJBgB3fBmk=

[InvalidSection]
SomeKey = SomeValue`

	_, err := parser.Parse(configText)
	if err == nil {
		t.Fatal("Expected config with invalid section to fail")
	}
	if !strings.Contains(err.Error(), "unknown section") {
		t.Errorf("Expected error about unknown section, got: %v", err)
	}
}

func TestParseConfigKeyValueOutsideSection(t *testing.T) {
	parser := NewParser(true)
	configText := `SomeKey = SomeValue
[Interface]
PrivateKey = yAnz5TF+lXXJte14tji3zlMNq+hd2rYUIgJBgB3fBmk=`

	_, err := parser.Parse(configText)
	if err == nil {
		t.Fatal("Expected config with key-value outside section to fail")
	}
	if !strings.Contains(err.Error(), "key-value outside of section") {
		t.Errorf("Expected error about key-value outside section, got: %v", err)
	}
}

func TestParseConfigInvalidLine(t *testing.T) {
	parser := NewParser(true)
	configText := `[Interface]
PrivateKey = yAnz5TF+lXXJte14tji3zlMNq+hd2rYUIgJBgB3fBmk=
InvalidLineWithoutEquals`

	_, err := parser.Parse(configText)
	if err == nil {
		t.Fatal("Expected config with invalid line to fail")
	}
	if !strings.Contains(err.Error(), "invalid line") {
		t.Errorf("Expected error about invalid line, got: %v", err)
	}
}

func TestParseConfigEmptyKey(t *testing.T) {
	parser := NewParser(true)
	configText := `[Interface]
= yAnz5TF+lXXJte14tji3zlMNq+hd2rYUIgJBgB3fBmk=`

	_, err := parser.Parse(configText)
	if err == nil {
		t.Fatal("Expected config with empty key to fail")
	}
	if !strings.Contains(err.Error(), "empty key") {
		t.Errorf("Expected error about empty key, got: %v", err)
	}
}

func TestValidateIPsValid(t *testing.T) {
	testCases := []string{
		"10.0.0.1/32",
		"192.168.1.0/24",
		"10.0.0.1",
		"192.168.1.1",
		"2001:db8::1/128",
		"2001:db8::/64",
		"10.0.0.1/32, 192.168.1.0/24",
		"192.168.1.1, 2001:db8::1",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			err := ValidateIPs(tc, false)
			if err != nil {
				t.Errorf("Expected valid IPs %q to pass validation, got: %v", tc, err)
			}
		})
	}
}

func TestValidateIPsInvalid(t *testing.T) {
	testCases := []string{
		"",
		"invalid",
		"999.999.999.999",
		"10.0.0.1/99",
		"192.168.1.0/64", // Invalid for IPv4
		"::1/129",        // Invalid prefix length
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			err := ValidateIPs(tc, false)
			if err == nil {
				t.Errorf("Expected invalid IPs %q to fail validation", tc)
			}
		})
	}
}

func TestValidateIPsCatchAllStrictMode(t *testing.T) {
	testCases := []string{
		"0.0.0.0/0",
		"::/0",
		"10.0.0.1/32, 0.0.0.0/0",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			err := ValidateIPs(tc, true)
			if err == nil {
				t.Errorf("Expected catch-all route %q to fail in strict mode", tc)
			}
		})
	}
}

func TestValidateIPsCatchAllNonStrictMode(t *testing.T) {
	testCases := []string{
		"0.0.0.0/0",
		"::/0",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			err := ValidateIPs(tc, false)
			if err != nil {
				t.Errorf("Expected catch-all route %q to pass in non-strict mode, got: %v", tc, err)
			}
		})
	}
}

func TestDetectIPConflictsNone(t *testing.T) {
	peers := []map[string]string{
		{
			"PublicKey":  "xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=",
			"AllowedIPs": "10.0.0.1/32",
		},
		{
			"PublicKey":  "TrMvSoP4jYQlY6RIzBgbssQqY3vxI2Pi+y71lOWWXX0=",
			"AllowedIPs": "10.0.0.2/32",
		},
	}

	err := DetectIPConflicts(peers)
	if err != nil {
		t.Errorf("Expected no conflicts, got: %v", err)
	}
}

func TestDetectIPConflictsOverlapping(t *testing.T) {
	peers := []map[string]string{
		{
			"PublicKey":  "xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=",
			"AllowedIPs": "10.0.0.0/24",
		},
		{
			"PublicKey":  "TrMvSoP4jYQlY6RIzBgbssQqY3vxI2Pi+y71lOWWXX0=",
			"AllowedIPs": "10.0.0.1/32",
		},
	}

	err := DetectIPConflicts(peers)
	if err == nil {
		t.Error("Expected IP conflict to be detected")
	}
	if !strings.Contains(err.Error(), "conflicts") {
		t.Errorf("Expected conflict error message, got: %v", err)
	}
}

func TestDetectIPConflictsDuplicate(t *testing.T) {
	peers := []map[string]string{
		{
			"PublicKey":  "xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=",
			"AllowedIPs": "10.0.0.1",
		},
		{
			"PublicKey":  "TrMvSoP4jYQlY6RIzBgbssQqY3vxI2Pi+y71lOWWXX0=",
			"AllowedIPs": "10.0.0.1",
		},
	}

	err := DetectIPConflicts(peers)
	if err == nil {
		t.Error("Expected duplicate IP to be detected")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("Expected duplicate error message, got: %v", err)
	}
}

func TestParseConfigFromTestData(t *testing.T) {
	parser := NewParser(false)

	// Test valid configurations
	validConfigs := []string{
		"valid_basic.conf",
		"valid_complex.conf",
	}

	for _, filename := range validConfigs {
		t.Run(filename, func(t *testing.T) {
			content, err := ioutil.ReadFile(filepath.Join("../testdata", filename))
			if err != nil {
				t.Skipf("Could not read test file %s: %v", filename, err)
				return
			}

			_, err = parser.Parse(string(content))
			if err != nil {
				t.Errorf("Expected valid config %s to parse, got error: %v", filename, err)
			}
		})
	}

	// Test invalid configurations
	invalidConfigs := []string{
		"invalid_no_interface.conf",
		"invalid_duplicate_keys.conf",
		"invalid_malformed.conf",
	}

	for _, filename := range invalidConfigs {
		t.Run(filename, func(t *testing.T) {
			content, err := ioutil.ReadFile(filepath.Join("../testdata", filename))
			if err != nil {
				t.Skipf("Could not read test file %s: %v", filename, err)
				return
			}

			_, err = parser.Parse(string(content))
			if err == nil {
				t.Errorf("Expected invalid config %s to fail parsing", filename)
			}
		})
	}
}

// Benchmark tests
func BenchmarkParseBasicConfig(b *testing.B) {
	parser := NewParser(true)
	configText := `[Interface]
PrivateKey = yAnz5TF+lXXJte14tji3zlMNq+hd2rYUIgJBgB3fBmk=
Address = 10.192.122.1/24

[Peer]
PublicKey = xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=
AllowedIPs = 10.192.122.2/32`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(configText)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidateIPs(b *testing.B) {
	ips := "10.0.0.1/32, 192.168.1.0/24, 2001:db8::1/128"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := ValidateIPs(ips, false)
		if err != nil {
			b.Fatal(err)
		}
	}
}
