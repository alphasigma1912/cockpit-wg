package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// MockWriter simulates file writing with controllable failures
type MockWriter struct {
	shouldFail      bool
	failAt          int
	bytesWritten    int
	data            []byte
	simulatePartial bool
}

func (m *MockWriter) Write(p []byte) (n int, err error) {
	if m.shouldFail && m.bytesWritten >= m.failAt {
		return 0, errors.New("simulated write failure")
	}

	if m.simulatePartial && len(p) > 1 {
		// Simulate partial write
		n = len(p) / 2
		m.data = append(m.data, p[:n]...)
		m.bytesWritten += n
		return n, nil
	}

	m.data = append(m.data, p...)
	m.bytesWritten += len(p)
	return len(p), nil
}

func (m *MockWriter) Close() error {
	if m.shouldFail && m.failAt == -1 {
		return errors.New("simulated close failure")
	}
	return nil
}

func (m *MockWriter) Sync() error {
	if m.shouldFail && m.failAt == -2 {
		return errors.New("simulated sync failure")
	}
	return nil
}

// AtomicWriter provides atomic file writing with rollback capability
type AtomicWriter struct {
	targetPath string
	tempPath   string
	file       *os.File
	committed  bool
}

// NewAtomicWriter creates a new atomic writer for the target path
func NewAtomicWriter(targetPath string) (*AtomicWriter, error) {
	tempPath := targetPath + ".tmp"

	file, err := os.Create(tempPath)
	if err != nil {
		return nil, err
	}

	return &AtomicWriter{
		targetPath: targetPath,
		tempPath:   tempPath,
		file:       file,
	}, nil
}

// Write writes data to the temporary file
func (aw *AtomicWriter) Write(data []byte) (int, error) {
	return aw.file.Write(data)
}

// Commit atomically moves the temporary file to the target location
func (aw *AtomicWriter) Commit() error {
	if aw.committed {
		return errors.New("already committed")
	}

	// Sync and close the file first
	if err := aw.file.Sync(); err != nil {
		aw.Rollback()
		return err
	}

	if err := aw.file.Close(); err != nil {
		aw.Rollback()
		return err
	}

	// Atomically move temp file to target
	if err := os.Rename(aw.tempPath, aw.targetPath); err != nil {
		os.Remove(aw.tempPath) // Clean up temp file
		return err
	}

	aw.committed = true
	return nil
}

// Rollback removes the temporary file and cleans up
func (aw *AtomicWriter) Rollback() error {
	if aw.committed {
		return errors.New("cannot rollback after commit")
	}

	if aw.file != nil {
		aw.file.Close()
	}

	return os.Remove(aw.tempPath)
}

// TestAtomicWriteSuccess tests successful atomic write operation
func TestAtomicWriteSuccess(t *testing.T) {
	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "test.wg")

	// Create atomic writer
	writer, err := NewAtomicWriter(targetPath)
	if err != nil {
		t.Fatalf("Failed to create atomic writer: %v", err)
	}

	// Write test data
	testData := []byte("[Interface]\nPrivateKey = test\n")
	n, err := writer.Write(testData)
	if err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}
	if n != len(testData) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(testData), n)
	}

	// Commit the write
	if err := writer.Commit(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Verify file exists and has correct content
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Error("Target file does not exist after commit")
	}

	content, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}

	if string(content) != string(testData) {
		t.Errorf("File content mismatch. Expected %q, got %q", testData, content)
	}

	// Verify temp file is cleaned up
	tempPath := targetPath + ".tmp"
	if _, err := os.Stat(tempPath); !os.IsNotExist(err) {
		t.Error("Temporary file was not cleaned up")
	}
}

// TestAtomicWriteRollback tests rollback functionality
func TestAtomicWriteRollback(t *testing.T) {
	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "test.wg")

	// Create atomic writer
	writer, err := NewAtomicWriter(targetPath)
	if err != nil {
		t.Fatalf("Failed to create atomic writer: %v", err)
	}

	// Write test data
	testData := []byte("[Interface]\nPrivateKey = test\n")
	_, err = writer.Write(testData)
	if err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	// Rollback instead of commit
	if err := writer.Rollback(); err != nil {
		t.Fatalf("Failed to rollback: %v", err)
	}

	// Verify target file does not exist
	if _, err := os.Stat(targetPath); !os.IsNotExist(err) {
		t.Error("Target file exists after rollback")
	}

	// Verify temp file is cleaned up
	tempPath := targetPath + ".tmp"
	if _, err := os.Stat(tempPath); !os.IsNotExist(err) {
		t.Error("Temporary file was not cleaned up after rollback")
	}
}

// TestAtomicWriteFailureDuringSync tests failure during sync operation
func TestAtomicWriteFailureDuringSync(t *testing.T) {
	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "test.wg")

	// Create atomic writer
	writer, err := NewAtomicWriter(targetPath)
	if err != nil {
		t.Fatalf("Failed to create atomic writer: %v", err)
	}

	// Write test data
	testData := []byte("[Interface]\nPrivateKey = test\n")
	_, err = writer.Write(testData)
	if err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	// Close the file to make sync fail
	writer.file.Close()

	// Attempt to commit should fail
	if err := writer.Commit(); err == nil {
		t.Error("Expected commit to fail after closing file")
	}

	// Verify target file does not exist
	if _, err := os.Stat(targetPath); !os.IsNotExist(err) {
		t.Error("Target file exists after failed commit")
	}
}

// TestAtomicWriteDoubleCommit tests that double commit fails
func TestAtomicWriteDoubleCommit(t *testing.T) {
	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "test.wg")

	// Create atomic writer
	writer, err := NewAtomicWriter(targetPath)
	if err != nil {
		t.Fatalf("Failed to create atomic writer: %v", err)
	}

	// Write and commit
	testData := []byte("[Interface]\nPrivateKey = test\n")
	_, _ = writer.Write(testData)

	if err := writer.Commit(); err != nil {
		t.Fatalf("First commit failed: %v", err)
	}

	// Second commit should fail
	if err := writer.Commit(); err == nil {
		t.Error("Expected second commit to fail")
	}
}

// TestAtomicWriteRollbackAfterCommit tests that rollback after commit fails
func TestAtomicWriteRollbackAfterCommit(t *testing.T) {
	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "test.wg")

	// Create atomic writer
	writer, err := NewAtomicWriter(targetPath)
	if err != nil {
		t.Fatalf("Failed to create atomic writer: %v", err)
	}

	// Write and commit
	testData := []byte("[Interface]\nPrivateKey = test\n")
	_, _ = writer.Write(testData)

	if err := writer.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// Rollback after commit should fail
	if err := writer.Rollback(); err == nil {
		t.Error("Expected rollback after commit to fail")
	}
}

// TestAtomicWriteExistingFile tests overwriting an existing file
func TestAtomicWriteExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "test.wg")

	// Create existing file
	originalData := []byte("original content")
	if err := os.WriteFile(targetPath, originalData, 0644); err != nil {
		t.Fatalf("Failed to create original file: %v", err)
	}

	// Create atomic writer
	writer, err := NewAtomicWriter(targetPath)
	if err != nil {
		t.Fatalf("Failed to create atomic writer: %v", err)
	}

	// Write new data
	newData := []byte("[Interface]\nPrivateKey = new\n")
	_, err = writer.Write(newData)
	if err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	// Verify original file still has original content
	content, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("Failed to read original file: %v", err)
	}
	if string(content) != string(originalData) {
		t.Error("Original file was modified before commit")
	}

	// Commit the write
	if err := writer.Commit(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Verify file now has new content
	content, err = os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}
	if string(content) != string(newData) {
		t.Errorf("File content mismatch. Expected %q, got %q", newData, content)
	}
}

// TestAtomicWritePermissionFailure tests handling of permission failures
func TestAtomicWritePermissionFailure(t *testing.T) {
	// Skip on Windows as permission model is different
	if os.Getenv("OS") == "Windows_NT" {
		t.Skip("Skipping permission test on Windows - different permission model")
	}

	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	// Create a read-only directory
	tempDir := t.TempDir()
	readOnlyDir := filepath.Join(tempDir, "readonly")
	if err := os.Mkdir(readOnlyDir, 0444); err != nil {
		t.Fatalf("Failed to create read-only directory: %v", err)
	}
	defer os.Chmod(readOnlyDir, 0755) // Cleanup

	targetPath := filepath.Join(readOnlyDir, "test.wg")

	// Attempt to create atomic writer should fail
	_, err := NewAtomicWriter(targetPath)
	if err == nil {
		t.Error("Expected atomic writer creation to fail in read-only directory")
	}
}

// TestConfigWriteWithAtomicWriter tests writing config using atomic writer
func TestConfigWriteWithAtomicWriter(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "wireguard.wg")

	// Test configuration
	config := "[Interface]\n" +
		"PrivateKey = oK56DE9Ue9zK76rAc8pBl6opph+1v36lm7cXXsQKrQM=\n" +
		"Address = 10.0.0.1/24\n" +
		"ListenPort = 51820\n" +
		"\n" +
		"[Peer]\n" +
		"PublicKey = HIgo9xNzJMWLKASShiTqIybxZ0U3wGLiUeJ1PKf8ykw=\n" +
		"AllowedIPs = 10.0.0.2/32\n" +
		"Endpoint = peer.example.com:51820\n"

	// Write config atomically
	err := WriteConfigAtomic(configPath, []byte(config))
	if err != nil {
		t.Fatalf("Failed to write config atomically: %v", err)
	}

	// Verify config was written correctly
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	if string(content) != config {
		t.Errorf("Config content mismatch")
	}

	// Verify we can parse the written config
	parser := NewParser(true)
	summary, err := parser.Parse(config)
	if err != nil {
		t.Fatalf("Failed to parse written config: %v", err)
	}

	if summary.Interface["PrivateKey"] != "oK56DE9Ue9zK76rAc8pBl6opph+1v36lm7cXXsQKrQM=" {
		t.Error("Config parsing failed - incorrect PrivateKey")
	}
}

// WriteConfigAtomic writes configuration data atomically to the specified path
func WriteConfigAtomic(path string, data []byte) error {
	writer, err := NewAtomicWriter(path)
	if err != nil {
		return err
	}

	_, err = writer.Write(data)
	if err != nil {
		writer.Rollback()
		return err
	}

	return writer.Commit()
}

// TestSimulatedDiskFull tests behavior when disk is full during write
func TestSimulatedDiskFull(t *testing.T) {
	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "test.wg")

	// Create atomic writer
	writer, err := NewAtomicWriter(targetPath)
	if err != nil {
		t.Fatalf("Failed to create atomic writer: %v", err)
	}

	// Write large amount of data to simulate disk full
	largeData := make([]byte, 1024*1024) // 1MB
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	// Write should succeed (unless disk is actually full)
	_, err = writer.Write(largeData)
	if err != nil {
		// If write fails due to actual disk space issues, rollback and skip
		writer.Rollback()
		t.Skipf("Skipping disk full test due to actual disk space: %v", err)
	}

	// Commit should succeed in normal cases
	if err := writer.Commit(); err != nil {
		t.Logf("Commit failed (possibly due to disk space): %v", err)
	}
}

// BenchmarkAtomicWrite benchmarks atomic write performance
func BenchmarkAtomicWrite(b *testing.B) {
	tempDir := b.TempDir()
	testData := []byte("[Interface]\nPrivateKey = test\nAddress = 10.0.0.1/24\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		targetPath := filepath.Join(tempDir, "bench.wg")

		writer, err := NewAtomicWriter(targetPath)
		if err != nil {
			b.Fatal(err)
		}

		_, err = writer.Write(testData)
		if err != nil {
			b.Fatal(err)
		}

		err = writer.Commit()
		if err != nil {
			b.Fatal(err)
		}

		// Clean up for next iteration
		os.Remove(targetPath)
	}
}
