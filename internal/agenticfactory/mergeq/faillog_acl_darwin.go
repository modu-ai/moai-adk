//go:build darwin

package mergeq

import (
	"fmt"
	"os/exec"
)

// clearExtendedACL removes the extended ACL of path (`chmod -N`). A directory with an inheritable ACL gives every
// file and directory made in it an ACL of its own, and the file mode does not remove that ACL: a 0600 file can still
// be readable by "group:everyone". The system chmod is called by its absolute path, so a PATH entry cannot stand in
// for it. A path without an ACL is not an error.
func clearExtendedACL(path string) error {
	if out, err := exec.Command("/bin/chmod", "-N", path).CombinedOutput(); err != nil {
		return fmt.Errorf("mergeq: clear the ACL of %s: %w (%.120s)", path, err, out)
	}
	return nil
}
