package manifest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

const hashPrefix = "sha256:"

// HashFile computes the SHA-256 hash of a file using streaming I/O.
// It never loads the entire file into memory, making it safe for large files.
// Returns the hash as "sha256:<hex>" format.
func HashFile(path string) (hashResult string, hashErr error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("hash file: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && hashErr == nil {
			hashResult = ""
			hashErr = fmt.Errorf("hash file close: %w", closeErr)
		}
	}()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("hash file: %w", err)
	}

	return hashPrefix + hex.EncodeToString(h.Sum(nil)), nil
}

// @MX:ANCHOR: [AUTO] SHA-256 hash computation utility - used by 6 or more callers throughout the manifest
// @MX:REASON: [AUTO] fan_in=6, the return format ("sha256:<hex>") is directly coupled with FileEntry.TemplateHash/DeployedHash/CurrentHash; do not change the format
// HashBytes computes the SHA-256 hash of a byte slice.
// Returns the hash as "sha256:<hex>" format.
func HashBytes(data []byte) string {
	h := sha256.Sum256(data)
	return hashPrefix + hex.EncodeToString(h[:])
}
