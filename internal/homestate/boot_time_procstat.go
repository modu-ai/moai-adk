package homestate

import (
	"errors"
	"io"
	"time"
)

var (
	errProcStatNoBtime        = errors.New("procfs stat carries no btime record")
	errProcStatMalformedBtime = errors.New("procfs stat btime record is malformed")
)

// procStatBootTime is the build-tag-free procfs interpretation seam (REQ-006c).
func procStatBootTime(r io.Reader) (time.Time, error) {
	return time.Time{}, errProcStatNoBtime
}
