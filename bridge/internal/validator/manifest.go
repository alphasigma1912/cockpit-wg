package validator

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// Manifest represents a .wgx exchange manifest
type Manifest struct {
	Interface string `json:"interface"`
	Version   int    `json:"version"`
	Checksum  string `json:"checksum"`
	Timestamp int64  `json:"timestamp,omitempty"`
	Source    string `json:"source,omitempty"`
}

// ManifestValidator validates .wgx manifest files
type ManifestValidator struct {
	strictMode      bool
	maxManifestSize int
	allowedVersions []int
}

// NewManifestValidator creates a new manifest validator
func NewManifestValidator(strict bool) *ManifestValidator {
	return &ManifestValidator{
		strictMode:      strict,
		maxManifestSize: 1024 * 1024, // 1MB
		allowedVersions: []int{1, 2},
	}
}

// ValidateManifest validates a manifest structure and content
func (v *ManifestValidator) ValidateManifest(manifest *Manifest, configData []byte) error {
	if err := v.validateManifestFields(manifest); err != nil {
		return fmt.Errorf("manifest validation failed: %w", err)
	}

	if err := v.validateChecksum(manifest, configData); err != nil {
		return fmt.Errorf("checksum validation failed: %w", err)
	}

	return nil
}

// ValidateManifestJSON validates raw JSON manifest data
func (v *ManifestValidator) ValidateManifestJSON(data []byte) (*Manifest, error) {
	if len(data) > v.maxManifestSize {
		return nil, fmt.Errorf("manifest too large: %d bytes (max %d)", len(data), v.maxManifestSize)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	if err := v.validateManifestFields(&manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// validateManifestFields validates manifest field values
func (v *ManifestValidator) validateManifestFields(manifest *Manifest) error {
	// Validate interface name
	if err := ValidateInterfaceName(manifest.Interface); err != nil {
		return fmt.Errorf("invalid interface name: %w", err)
	}

	// Validate version
	if !v.isVersionAllowed(manifest.Version) {
		return fmt.Errorf("unsupported version: %d (supported: %v)",
			manifest.Version, v.allowedVersions)
	}

	// Validate checksum format
	if err := v.validateChecksumFormat(manifest.Checksum); err != nil {
		return fmt.Errorf("invalid checksum: %w", err)
	}

	// Validate source (if present)
	if manifest.Source != "" {
		if err := v.validateSource(manifest.Source); err != nil {
			return fmt.Errorf("invalid source: %w", err)
		}
	}

	// Validate timestamp (if present and in strict mode)
	if v.strictMode && manifest.Timestamp != 0 {
		if err := v.validateTimestamp(manifest.Timestamp); err != nil {
			return fmt.Errorf("invalid timestamp: %w", err)
		}
	}

	return nil
}

// validateChecksum verifies the checksum against actual data
func (v *ManifestValidator) validateChecksum(manifest *Manifest, data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("empty config data")
	}

	hash := sha256.Sum256(data)
	expected := hex.EncodeToString(hash[:])

	if !strings.EqualFold(manifest.Checksum, expected) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, manifest.Checksum)
	}

	return nil
}

// validateChecksumFormat validates checksum string format
func (v *ManifestValidator) validateChecksumFormat(checksum string) error {
	if len(checksum) != 64 {
		return fmt.Errorf("invalid checksum length: %d (expected 64)", len(checksum))
	}

	checksumRx := regexp.MustCompile(`^[a-fA-F0-9]+$`)
	if !checksumRx.MatchString(checksum) {
		return fmt.Errorf("invalid checksum format: contains non-hex characters")
	}

	return nil
}

// isVersionAllowed checks if version is supported
func (v *ManifestValidator) isVersionAllowed(version int) bool {
	for _, allowed := range v.allowedVersions {
		if version == allowed {
			return true
		}
	}
	return false
}

// validateSource validates source field format
func (v *ManifestValidator) validateSource(source string) error {
	if len(source) == 0 || len(source) > 255 {
		return fmt.Errorf("source length out of range: %d (1-255)", len(source))
	}

	// Basic validation - no control characters
	for _, r := range source {
		if r < 32 || r > 126 {
			return fmt.Errorf("source contains invalid characters")
		}
	}

	return nil
}

// validateTimestamp validates timestamp values
func (v *ManifestValidator) validateTimestamp(timestamp int64) error {
	if timestamp <= 0 {
		return fmt.Errorf("invalid timestamp: %d", timestamp)
	}

	// Check if timestamp is reasonable (not too far in future)
	// This is a basic sanity check
	const maxFutureSeconds = 24 * 60 * 60 // 24 hours
	now := int64(1691587200)              // 2023-08-09 as baseline
	if timestamp > now+maxFutureSeconds {
		return fmt.Errorf("timestamp too far in future: %d", timestamp)
	}

	return nil
}

// GenerateChecksum generates SHA256 checksum for config data
func GenerateChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
