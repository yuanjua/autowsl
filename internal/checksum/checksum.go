package checksum

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

// VerifyFile checks if a file matches the expected SHA256 checksum
func VerifyFile(path, expectedSHA256 string) error {
	if expectedSHA256 == "" {
		return fmt.Errorf("no checksum provided for verification")
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file for checksum: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("failed to compute checksum: %w", err)
	}

	actual := hex.EncodeToString(h.Sum(nil))
	expected := strings.ToLower(strings.TrimSpace(expectedSHA256))

	if actual != expected {
		return fmt.Errorf("checksum mismatch: got %s, expected %s", actual, expected)
	}

	return nil
}

// ComputeFile computes the SHA256 checksum of a file
func ComputeFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("failed to compute checksum: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
