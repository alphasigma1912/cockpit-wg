package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func FuzzParseConfig(f *testing.F) {
	// Add seed corpus from test data
	testDataDir := "../../testdata"

	// Add valid configurations as seeds
	seedFiles := []string{
		"valid_config.wg",
		"valid_config_minimal.wg",
		"valid_config_complex.wg",
	}

	for _, filename := range seedFiles {
		path := filepath.Join(testDataDir, filename)
		if data, err := os.ReadFile(path); err == nil {
			f.Add(string(data))
		}
	}

	// Add some basic seeds
	f.Add("[Interface]\nPrivateKey = oK56DE9Ue9zK76rAc8pBl6opph+1v36lm7cXXsQKrQM=\n")
	f.Add("[Peer]\nPublicKey = HIgo9xNzJMWLKASShiTqIybxZ0U3wGLiUeJ1PKf8ykw=\n")
	f.Add("")
	f.Add("invalid")
	f.Add("[Interface]\n")
	f.Add("[Peer]\n")

	f.Fuzz(func(t *testing.T, text string) {
		parser := NewParser(true)

		// Parsing should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Parser panicked with input %q: %v", text, r)
			}
		}()

		// Try to parse the data
		config, err := parser.Parse(text)

		// If parsing succeeds, the config should be valid
		if err == nil && config != nil {
			// Basic sanity checks
			if config.Interface == nil {
				t.Error("Parsed config has nil interface")
			}

			// Validate that parsed peers have required fields
			for _, peer := range config.Peers {
				if allowedIPs, ok := peer["AllowedIPs"]; ok && allowedIPs != "" {
					// AllowedIPs should be parseable
					if validateErr := ValidateIPs(allowedIPs, false); validateErr != nil {
						// This is not necessarily an error in fuzzing - just log it
						t.Logf("Parsed AllowedIPs validation failed: %v", validateErr)
					}
				}
			}
		}

		// Errors should be informative
		if err != nil && err.Error() == "" {
			t.Error("Error with empty message")
		}
	})
}

func FuzzValidateIPs(f *testing.F) {
	// Add seed corpus
	f.Add("192.168.1.0/24")
	f.Add("10.0.0.0/8, 172.16.0.0/12")
	f.Add("0.0.0.0/0")
	f.Add("::/0")
	f.Add("2001:db8::/32")
	f.Add("192.168.1.1/32, 192.168.1.2/32")
	f.Add("")
	f.Add("invalid")
	f.Add("192.168.1.256/24")
	f.Add("192.168.1.1/33")

	f.Fuzz(func(t *testing.T, ips string) {
		// ValidateIPs should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ValidateIPs panicked with input %q: %v", ips, r)
			}
		}()

		// Test both strict and non-strict modes
		err1 := ValidateIPs(ips, true)
		err2 := ValidateIPs(ips, false)

		// Errors should be informative
		if err1 != nil && err1.Error() == "" {
			t.Error("Strict mode error with empty message")
		}
		if err2 != nil && err2.Error() == "" {
			t.Error("Non-strict mode error with empty message")
		}

		// Strict mode should be more restrictive than non-strict
		if err1 == nil && err2 != nil {
			t.Error("Non-strict mode failed while strict mode passed")
		}
	})
}

func FuzzDetectIPConflicts(f *testing.F) {
	// Add seed corpus as JSON strings
	f.Add(`[{"PublicKey":"HIgo9xNzJMWLKASShiTqIybxZ0U3wGLiUeJ1PKf8ykw=","AllowedIPs":"192.168.1.0/24"},{"PublicKey":"xTIBA5rboUvnH4htodjb6e2QK5AzPVjCyno8rUzsVs=","AllowedIPs":"192.168.1.1/32"}]`)
	f.Add(`[{"PublicKey":"key1","AllowedIPs":"10.0.0.0/8"},{"PublicKey":"key2","AllowedIPs":"172.16.0.0/12"}]`)
	f.Add(`[]`)
	f.Add(`[{"InvalidKey":"value"}]`)
	f.Add(`[{"PublicKey":"test","AllowedIPs":"invalid"}]`)

	f.Fuzz(func(t *testing.T, peersJSON string) {
		// DetectIPConflicts should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("DetectIPConflicts panicked with input %q: %v", peersJSON, r)
			}
		}()

		// Try to parse JSON
		var peers []map[string]string
		if err := json.Unmarshal([]byte(peersJSON), &peers); err != nil {
			// Invalid JSON is fine - just skip
			return
		}

		err := DetectIPConflicts(peers)

		// Errors should be informative
		if err != nil && err.Error() == "" {
			t.Error("Error with empty message")
		}
	})
}
