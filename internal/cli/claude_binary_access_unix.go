//go:build !windows

package cli

import (
	"fmt"
	"os"
)

func checkClaudeBinaryExecuteAccess(_ string, info os.FileInfo) error {
	if info.Mode()&0o111 == 0 {
		return fmt.Errorf("chmod +x it or point the pin at an executable file")
	}
	return nil
}
