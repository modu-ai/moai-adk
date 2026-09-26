package homestate

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

var (
	errProcStatNoBtime        = errors.New("procfs stat carries no btime record")
	errProcStatMalformedBtime = errors.New("procfs stat btime record is malformed")
)

// procStatBootTime interprets procfs stat content and returns the host boot
// time from its `btime` record (REQ-006c). It carries no build constraint so
// every host can exercise it; the linux reader is only its caller.
//
// Lines are read without a length cap: the `intr` line ahead of `btime` exceeds
// the 64 KiB bufio.Scanner default on hosts with many interrupt sources, and a
// scanner stopping there would lose the record. A read error before the record
// is found is returned as the cause, so it stays distinguishable from a source
// read to its end without one (errProcStatNoBtime). Once `btime` is parsed the
// rest of the source is not read.
func procStatBootTime(r io.Reader) (time.Time, error) {
	br := bufio.NewReader(r)
	for {
		line, err := br.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return time.Time{}, fmt.Errorf("read procfs stat: %w", err)
		}
		if fields := strings.Fields(line); len(fields) > 0 && fields[0] == "btime" {
			return parseProcStatBtime(fields)
		}
		if err != nil { // io.EOF: the whole source was read.
			return time.Time{}, errProcStatNoBtime
		}
	}
}

func parseProcStatBtime(fields []string) (time.Time, error) {
	if len(fields) != 2 {
		return time.Time{}, fmt.Errorf("%w: %d fields", errProcStatMalformedBtime, len(fields))
	}
	sec, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %v", errProcStatMalformedBtime, err)
	}
	if sec <= 0 {
		return time.Time{}, fmt.Errorf("%w: non-positive value %d", errProcStatMalformedBtime, sec)
	}
	return time.Unix(sec, 0), nil
}
