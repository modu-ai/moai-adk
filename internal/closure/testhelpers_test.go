package closure

import "os"

// writeFileForTest writes content to path with 0o644 (test helper shared by
// the records tests).
func writeFileForTest(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}
