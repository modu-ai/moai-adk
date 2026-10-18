//go:build unix

package cli

import "golang.org/x/sys/unix"

// auditLogOpenNoFollow is the open(2) flag set the leader-anchor audit log adds
// to os.O_APPEND|os.O_CREATE|os.O_WRONLY (SPEC-LAUNCHER-ARGV-001 REQ-LA-009;
// card t1604, O28). O_NOFOLLOW refuses a symbolic link at the final path
// component, and O_NONBLOCK keeps a FIFO from blocking the open until a reader
// appears.
const auditLogOpenNoFollow = unix.O_NOFOLLOW | unix.O_NONBLOCK
