//go:build linux

// proc_info_linux.go — ancestry lookup via /proc/<pid>/stat.
package session

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// platformProcInfo parses /proc/<pid>/stat. The second field is the command
// name wrapped in parentheses and MAY itself contain spaces and parentheses,
// so the parse anchors on the LAST ')' rather than splitting the whole line;
// the state character and the parent PID follow it.
func platformProcInfo(pid int) (int, string, bool) {
	if pid <= 0 {
		return 0, "", false
	}
	raw, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0, "", false
	}
	line := string(raw)
	openIdx := strings.IndexByte(line, '(')
	closeIdx := strings.LastIndexByte(line, ')')
	if openIdx < 0 || closeIdx < openIdx {
		return 0, "", false
	}
	comm := line[openIdx+1 : closeIdx]
	rest := strings.Fields(line[closeIdx+1:])
	if len(rest) < 2 {
		return 0, "", false
	}
	ppid, err := strconv.Atoi(rest[1])
	if err != nil {
		return 0, "", false
	}
	return ppid, comm, true
}
