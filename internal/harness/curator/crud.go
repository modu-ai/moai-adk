package curator

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// bulletLine renders a single Bullet into its line representation.
func bulletLine(bullet Bullet) string {
	line := "- " + bullet.Text
	if bullet.LedgerKey != "" {
		line += " <!-- key: " + bullet.LedgerKey + " -->"
	}
	return line
}

// keyMarker returns the HTML-comment token used to identify a bullet by
// ledger_key within the block body.
func keyMarker(ledgerKey string) string {
	return "<!-- key: " + ledgerKey + " -->"
}

// readLines reads the file at path and returns its content split by newline.
// The trailing element after a final newline is preserved (it is "" for
// files ending with "\n"), so strings.Join(lines, "\n") round-trips exactly.
func readLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("curator: read %s: %w", path, err)
	}
	return strings.Split(string(data), "\n"), nil
}

// writeLines joins lines with newline and writes them to path.
func writeLines(path string, lines []string) error {
	updated := strings.Join(lines, "\n")
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("curator: write %s: %w", path, err)
	}
	return nil
}

// findEndMarkerLine returns the index of the first line containing the end
// marker for the given blockType, or -1 if not found.
func findEndMarkerLine(lines []string, blockType BlockType) int {
	spec, ok := markerRegistry[blockType]
	if !ok {
		return -1
	}
	for i, line := range lines {
		if strings.Contains(line, spec.endMarker) {
			return i
		}
	}
	return -1
}

// AddBullet appends a single bullet to the managed block of the given
// blockType. The bullet line is inserted immediately before the end marker.
// Existing bullets and all bytes outside the block are preserved unchanged.
//
// Returns ErrBlockNotFound when the file does not contain a marker block of
// the given type.
func AddBullet(path string, blockType BlockType, bullet Bullet) error {
	if path == "" {
		return errors.New("curator: empty path")
	}
	lines, err := readLines(path)
	if err != nil {
		return err
	}
	endIdx := findEndMarkerLine(lines, blockType)
	if endIdx < 0 {
		return fmt.Errorf("curator: %w (BlockType %d)", ErrBlockNotFound, blockType)
	}

	newLine := bulletLine(bullet)
	// Insert the bullet line before the end marker line.
	result := make([]string, 0, len(lines)+1)
	result = append(result, lines[:endIdx]...)
	result = append(result, newLine)
	result = append(result, lines[endIdx:]...)

	return writeLines(path, result)
}

// UpdateBullet locates the bullet identified by ledgerKey within the managed
// block and replaces its text with newText. The ledger_key marker is
// preserved; only the text portion of the line changes. All other bullets
// and all bytes outside the targeted line are preserved unchanged.
//
// Returns ErrBulletNotFound when no bullet with the given ledgerKey exists.
func UpdateBullet(path string, blockType BlockType, ledgerKey, newText string) error {
	if path == "" {
		return errors.New("curator: empty path")
	}
	if ledgerKey == "" {
		return errors.New("curator: UpdateBullet: empty ledgerKey")
	}
	lines, err := readLines(path)
	if err != nil {
		return err
	}
	marker := keyMarker(ledgerKey)
	found := false
	for i, line := range lines {
		if strings.Contains(line, marker) {
			lines[i] = "- " + newText + " " + marker
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("curator: %w: %s", ErrBulletNotFound, ledgerKey)
	}
	return writeLines(path, lines)
}

// DeleteBullet removes the bullet identified by ledgerKey from the managed
// block. The entire bullet line is removed; all other bullets and all bytes
// outside the deleted line are preserved unchanged.
//
// Returns ErrBulletNotFound when no bullet with the given ledgerKey exists.
func DeleteBullet(path string, blockType BlockType, ledgerKey string) error {
	if path == "" {
		return errors.New("curator: empty path")
	}
	if ledgerKey == "" {
		return errors.New("curator: DeleteBullet: empty ledgerKey")
	}
	lines, err := readLines(path)
	if err != nil {
		return err
	}
	marker := keyMarker(ledgerKey)
	found := false
	for i, line := range lines {
		if strings.Contains(line, marker) {
			lines = append(lines[:i], lines[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("curator: %w: %s", ErrBulletNotFound, ledgerKey)
	}
	return writeLines(path, lines)
}
