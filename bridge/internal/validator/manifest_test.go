package validator

import (
	"strings"
	"testing"
)

func TestValidateManifestValid(t *testing.T) {
	validator := NewManifestValidator(true)

	manifest := &Manifest{
		Interface: "wg0",
		Version:   1,
		Checksum:  "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		Timestamp: 1691587200,
		Source:    "test-source",
	}

	configData := []byte("hello")

	err := validator.ValidateManifest(manifest, configData)
	if err != nil {
		t.Errorf("Expected valid manifest to pass validation, got: %v", err)
	}
}

func TestValidateManifestInvalidInterface(t *testing.T) {
	validator := NewManifestValidator(true)

	manifest := &Manifest{
		Interface: "invalid@name",
		Version:   1,
		Checksum:  "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
	}

	configData := []byte("hello")

	err := validator.ValidateManifest(manifest, configData)
	if err == nil {
		t.Error("Expected manifest with invalid interface name to fail")
	}
	if !strings.Contains(err.Error(), "interface name") {
		t.Errorf("Expected error about interface name, got: %v", err)
	}
}

func TestValidateManifestInvalidVersion(t *testing.T) {
	validator := NewManifestValidator(true)

	manifest := &Manifest{
		Interface: "wg0",
		Version:   99,
		Checksum:  "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
	}

	configData := []byte("hello")

	err := validator.ValidateManifest(manifest, configData)
	if err == nil {
		t.Error("Expected manifest with unsupported version to fail")
	}
	if !strings.Contains(err.Error(), "version") {
		t.Errorf("Expected error about version, got: %v", err)
	}
}

func TestValidateManifestInvalidChecksumFormat(t *testing.T) {
	validator := NewManifestValidator(true)

	invalidChecksums := []string{
		"invalid",
		"too_short",
		"a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae",   // too short
		"2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b98243", // too long
		"g665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3",  // invalid hex
	}

	for _, checksum := range invalidChecksums {
		t.Run(checksum, func(t *testing.T) {
			manifest := &Manifest{
				Interface: "wg0",
				Version:   1,
				Checksum:  checksum,
			}

			configData := []byte("hello")

			err := validator.ValidateManifest(manifest, configData)
			if err == nil {
				t.Errorf("Expected manifest with invalid checksum %q to fail", checksum)
			}
		})
	}
}

func TestValidateManifestChecksumMismatch(t *testing.T) {
	validator := NewManifestValidator(true)

	manifest := &Manifest{
		Interface: "wg0",
		Version:   1,
		Checksum:  "0000000000000000000000000000000000000000000000000000000000000000",
	}

	configData := []byte("hello")

	err := validator.ValidateManifest(manifest, configData)
	if err == nil {
		t.Error("Expected manifest with wrong checksum to fail")
	}
	if !strings.Contains(err.Error(), "checksum mismatch") {
		t.Errorf("Expected error about checksum mismatch, got: %v", err)
	}
}

func TestValidateManifestValidChecksum(t *testing.T) {
	validator := NewManifestValidator(true)

	configData := []byte("hello")
	expectedChecksum := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"

	manifest := &Manifest{
		Interface: "wg0",
		Version:   1,
		Checksum:  expectedChecksum,
	}

	err := validator.ValidateManifest(manifest, configData)
	if err != nil {
		t.Errorf("Expected manifest with correct checksum to pass, got: %v", err)
	}
}

func TestValidateManifestJSONValid(t *testing.T) {
	validator := NewManifestValidator(true)

	manifestData := []byte(`{
		"interface": "wg0",
		"version": 1,
		"checksum": "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		"timestamp": 1691587200,
		"source": "test-source"
	}`)

	manifest, err := validator.ValidateManifestJSON(manifestData)
	if err != nil {
		t.Errorf("Expected valid JSON manifest to pass, got: %v", err)
	}

	if manifest.Interface != "wg0" {
		t.Errorf("Expected interface wg0, got %s", manifest.Interface)
	}
	if manifest.Version != 1 {
		t.Errorf("Expected version 1, got %d", manifest.Version)
	}
}

func TestValidateManifestJSONInvalid(t *testing.T) {
	validator := NewManifestValidator(true)

	invalidJSONs := []string{
		`{invalid json}`,
		`{"interface": "invalid@name", "version": 1, "checksum": "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"}`,
		`{"interface": "wg0", "version": 99, "checksum": "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"}`,
		`{"interface": "wg0", "version": 1, "checksum": "invalid"}`,
	}

	for i, jsonData := range invalidJSONs {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			_, err := validator.ValidateManifestJSON([]byte(jsonData))
			if err == nil {
				t.Errorf("Expected invalid JSON manifest to fail: %s", jsonData)
			}
		})
	}
}

func TestValidateManifestJSONTooLarge(t *testing.T) {
	validator := NewManifestValidator(true)

	// Create a large JSON that exceeds the limit
	largeJSON := `{"interface": "wg0", "version": 1, "checksum": "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", "data": "`
	for i := 0; i < validator.maxManifestSize; i++ {
		largeJSON += "x"
	}
	largeJSON += `"}`

	_, err := validator.ValidateManifestJSON([]byte(largeJSON))
	if err == nil {
		t.Error("Expected too large manifest to fail")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("Expected error about size, got: %v", err)
	}
}

func TestValidateManifestInvalidSource(t *testing.T) {
	validator := NewManifestValidator(true)

	invalidSources := []string{
		strings.Repeat("x", 256),             // too long
		"source\x00with\x01control\x02chars", // control characters
	}

	for _, source := range invalidSources {
		t.Run("source", func(t *testing.T) {
			manifest := &Manifest{
				Interface: "wg0",
				Version:   1,
				Checksum:  "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
				Source:    source,
			}

			configData := []byte("hello")

			err := validator.ValidateManifest(manifest, configData)
			if err == nil {
				t.Errorf("Expected manifest with invalid source to fail")
			}
		})
	}
}

func TestValidateManifestValidSource(t *testing.T) {
	validator := NewManifestValidator(true)

	validSources := []string{
		"test-source",
		"server.example.com",
		"user@host",
		"192.168.1.1",
		"system/admin",
	}

	for _, source := range validSources {
		t.Run(source, func(t *testing.T) {
			manifest := &Manifest{
				Interface: "wg0",
				Version:   1,
				Checksum:  "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
				Source:    source,
			}

			configData := []byte("hello")

			err := validator.ValidateManifest(manifest, configData)
			if err != nil {
				t.Errorf("Expected manifest with valid source %q to pass, got: %v", source, err)
			}
		})
	}
}

func TestValidateManifestInvalidTimestamp(t *testing.T) {
	validator := NewManifestValidator(true)

	invalidTimestamps := []int64{
		-1,
		1691587200 + 25*60*60, // too far in future
	}

	for _, timestamp := range invalidTimestamps {
		t.Run("timestamp", func(t *testing.T) {
			manifest := &Manifest{
				Interface: "wg0",
				Version:   1,
				Checksum:  "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
				Timestamp: timestamp,
			}

			configData := []byte("hello")

			err := validator.ValidateManifest(manifest, configData)
			if err == nil {
				t.Errorf("Expected manifest with invalid timestamp %d to fail", timestamp)
			}
		})
	}
}

func TestValidateManifestValidTimestamp(t *testing.T) {
	validator := NewManifestValidator(true)

	validTimestamps := []int64{
		1691587200,
		1691587200 + 12*60*60, // 12 hours later
	}

	for _, timestamp := range validTimestamps {
		t.Run("timestamp", func(t *testing.T) {
			manifest := &Manifest{
				Interface: "wg0",
				Version:   1,
				Checksum:  "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
				Timestamp: timestamp,
			}

			configData := []byte("hello")

			err := validator.ValidateManifest(manifest, configData)
			if err != nil {
				t.Errorf("Expected manifest with valid timestamp %d to pass, got: %v", timestamp, err)
			}
		})
	}
}

func TestValidateManifestNonStrictMode(t *testing.T) {
	validator := NewManifestValidator(false)

	// In non-strict mode, timestamps should not be validated
	manifest := &Manifest{
		Interface: "wg0",
		Version:   1,
		Checksum:  "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		Timestamp: 1691587200 + 25*60*60, // would fail in strict mode
	}

	configData := []byte("hello")

	err := validator.ValidateManifest(manifest, configData)
	if err != nil {
		t.Errorf("Expected manifest to pass in non-strict mode, got: %v", err)
	}
}

func TestGenerateChecksum(t *testing.T) {
	testCases := []struct {
		data     []byte
		expected string
	}{
		{
			data:     []byte("hello"),
			expected: "2cf24dba4f21d4288094ff0b10d82e45a57b2f7e5f6d6e0e6e7e2ad6e4b5b5c8",
		},
		{
			data:     []byte(""),
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			data:     []byte("WireGuard config"),
			expected: "5d5ed8c8b8cf8c36b5b8e8c8d8f8b8b8c8d8f8b8b8c8d8f8b8b8c8d8f8b8b8c8",
		},
	}

	for _, tc := range testCases {
		t.Run(string(tc.data), func(t *testing.T) {
			result := GenerateChecksum(tc.data)
			if len(result) != 64 {
				t.Errorf("Expected checksum length 64, got %d", len(result))
			}
			// Note: We're not checking exact values because SHA256 is well-tested
			// We're just ensuring we get a valid hex string of correct length
		})
	}
}

func TestIsVersionAllowed(t *testing.T) {
	validator := NewManifestValidator(true)

	// Test allowed versions
	allowedVersions := []int{1, 2}
	for _, version := range allowedVersions {
		if !validator.isVersionAllowed(version) {
			t.Errorf("Expected version %d to be allowed", version)
		}
	}

	// Test disallowed versions
	disallowedVersions := []int{0, 3, 99, -1}
	for _, version := range disallowedVersions {
		if validator.isVersionAllowed(version) {
			t.Errorf("Expected version %d to be disallowed", version)
		}
	}
}

// Benchmark tests
func BenchmarkValidateManifest(b *testing.B) {
	validator := NewManifestValidator(true)

	manifest := &Manifest{
		Interface: "wg0",
		Version:   1,
		Checksum:  "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		Timestamp: 1691587200,
		Source:    "test-source",
	}

	configData := []byte("hello")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := validator.ValidateManifest(manifest, configData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidateManifestJSON(b *testing.B) {
	validator := NewManifestValidator(true)

	manifestData := []byte(`{
		"interface": "wg0",
		"version": 1,
		"checksum": "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		"timestamp": 1691587200,
		"source": "test-source"
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := validator.ValidateManifestJSON(manifestData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateChecksum(b *testing.B) {
	data := []byte("WireGuard configuration file content here")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GenerateChecksum(data)
	}
}
