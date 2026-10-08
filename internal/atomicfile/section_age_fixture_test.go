package atomicfile

import (
	"encoding/json"
	"os"
)

// writeLiveFixture seeds a lock file with an explicit owner label (the
// age-exceeded fixture).
func writeLiveFixture(path string, owner LockOwner) error {
	raw, err := json.Marshal(owner)
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o600)
}
