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

// @MX:ANCHOR: [AUTO] SHA-256 해시 계산 유틸리티 - 매니페스트 전반에서 6개 이상의 호출자가 사용
// @MX:REASON: fan_in=6, 반환 포맷("sha256:<hex>")이 FileEntry.TemplateHash/DeployedHash/CurrentHash와 직접 연동되므로 포맷 변경 금지
// HashBytes computes the SHA-256 hash of a byte slice.
// Returns the hash as "sha256:<hex>" format.
func HashBytes(data []byte) string {
	h := sha256.Sum256(data)
	return hashPrefix + hex.EncodeToString(h[:])
}
