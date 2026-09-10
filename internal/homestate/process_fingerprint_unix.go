//go:build !windows

package homestate

import (
	"os/exec"
	"strconv"
	"strings"
)

func platformProcessFingerprint(pid int) (string, bool) {
	out, err := exec.Command("ps", "-o", "lstart=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return "", false
	}
	value := strings.TrimSpace(string(out))
	return value, value != ""
}
