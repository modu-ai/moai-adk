//go:build windows

// migrate_cwd_windows.go — the migration's process-cwd scan on Windows (cards
// t2075, t2094). There is no lsof and no supported Win32 call that lists every
// process's working directory, so the scan reads it out of each process's PEB:
// CreateToolhelp32Snapshot enumerates the processes, NtQueryInformationProcess
// locates the PEB, and ReadProcessMemory walks PEB → ProcessParameters →
// CurrentDirectory. Same-user processes open with
// PROCESS_QUERY_LIMITED_INFORMATION|PROCESS_VM_READ without elevation; every
// other process fails individually and is skipped.
//
// Contract (card t2094, F2): a scan that observed not a single working
// directory — ntdll resolution failed, the snapshot failed, every open was
// refused — reports errProcessCWDOSGuarded, so the plan keeps t2075's win of
// never skipping every tree: Windows refuses to rename a directory a process
// holds as its working directory, and the move step's own failure handling is
// the guard. A scan that observed at least one directory returns the partial
// list with a nil error; the processes it could not scan are caught at move
// time by that same OS guard. Never return an empty list with a nil error —
// an empty list reads as "no process inside".

package worktree

import (
	"encoding/binary"
	"fmt"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ntQueryInformationProcess is the ntdll export behind the PEB lookup (no
// Win32 wrapper exists for it).
var ntQueryInformationProcess = windows.NewLazySystemDLL("ntdll.dll").NewProc("NtQueryInformationProcess")

// processBasicInformationClass is NtQueryInformationProcess's
// ProcessBasicInformation query class.
const processBasicInformationClass = 0

// processBasicInformation is PROCESS_BASIC_INFORMATION; only PebBaseAddress is
// used.
type processBasicInformation struct {
	ExitStatus                   uintptr
	PebBaseAddress               uintptr
	AffinityMask                 uintptr
	BasePriority                 uintptr
	UniqueProcessId              uintptr
	InheritedFromUniqueProcessId uintptr
}

// unicodeString is UNICODE_STRING: a byte length and a pointer to UTF-16 data
// inside the target process.
type unicodeString struct {
	Length        uint16
	MaximumLength uint16
	Buffer        uintptr
}

// defaultActiveProcessCWDs scans the Win32 working directories of every
// process it can open (the conservative process-level guard for worktree
// moves).
func defaultActiveProcessCWDs() ([]string, error) {
	if err := ntQueryInformationProcess.Find(); err != nil {
		return nil, fmt.Errorf("%w: NtQueryInformationProcess: %v", errProcessCWDOSGuarded, err)
	}
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, fmt.Errorf("%w: CreateToolhelp32Snapshot: %v", errProcessCWDOSGuarded, err)
	}
	defer func() { _ = windows.CloseHandle(snap) }()
	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	var directories []string
	for err = windows.Process32First(snap, &entry); err == nil; err = windows.Process32Next(snap, &entry) {
		if dir, ok := windowsProcessCWD(entry.ProcessID); ok {
			directories = append(directories, dir)
		}
	}
	if len(directories) == 0 {
		// Not a single working directory was observable: total scan failure.
		return nil, errProcessCWDOSGuarded
	}
	return directories, nil
}

// windowsProcessCWD reads one process's Win32 working directory out of its
// PEB. Any failure — the open was refused (another user's process), the
// information query declined, the memory read failed (a different-bitness
// target answers the offsets with garbage) — reports not-ok and the caller
// skips that process.
func windowsProcessCWD(pid uint32) (string, bool) {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION|windows.PROCESS_VM_READ, false, pid)
	if err != nil {
		return "", false
	}
	defer func() { _ = windows.CloseHandle(h) }()
	var pbi processBasicInformation
	if status, _, _ := ntQueryInformationProcess.Call(
		uintptr(h), processBasicInformationClass,
		uintptr(unsafe.Pointer(&pbi)), unsafe.Sizeof(pbi), 0); status != 0 || pbi.PebBaseAddress == 0 {
		return "", false
	}
	params, ok := readProcessValue[uintptr](h, pbi.PebBaseAddress+pebProcessParametersOffset())
	if !ok || params == 0 {
		return "", false
	}
	current, ok := readProcessValue[unicodeString](h, params+rupCurrentDirectoryOffset())
	if !ok || current.Length == 0 || current.Buffer == 0 {
		return "", false
	}
	raw := make([]byte, current.Length)
	if !readProcessBytes(h, current.Buffer, raw) {
		return "", false
	}
	utf16 := make([]uint16, len(raw)/2)
	for i := range utf16 {
		utf16[i] = binary.LittleEndian.Uint16(raw[2*i:])
	}
	// The kernel keeps the directory with a trailing separator; drop it so the
	// entry compares like any other directory path.
	dir := strings.TrimRight(windows.UTF16ToString(utf16), `\`)
	if dir == "" {
		return "", false
	}
	return dir, true
}

// pebProcessParametersOffset and rupCurrentDirectoryOffset are the structure
// offsets the scan walks, per pointer width: PEB.ProcessParameters and
// RTL_USER_PROCESS_PARAMETERS.CurrentDirectory in the 64- and 32-bit layouts.
// A process of a different bitness than this one answers these offsets with a
// failed read and is skipped like any other unscannable process.
func pebProcessParametersOffset() uintptr {
	if unsafe.Sizeof(uintptr(0)) == 8 {
		return 0x20
	}
	return 0x10
}

func rupCurrentDirectoryOffset() uintptr {
	if unsafe.Sizeof(uintptr(0)) == 8 {
		return 0x38
	}
	return 0x24
}

// @MX:WARN: [AUTO] readProcessValue reads raw cross-process memory at offsets
// the target's own structure layout defines — a wrong offset or a
// foreign-bitness target yields garbage or a failed read, not a distinguishable
// error.
// @MX:REASON: the Win32 working directory is exposed by no supported Win32
// API; the PEB is the only source. A failed read is handled as "unobservable"
// (the process is skipped), which contains the failure mode.
func readProcessValue[T any](h windows.Handle, addr uintptr) (T, bool) {
	var v T
	if !readProcessBytes(h, addr, unsafe.Slice((*byte)(unsafe.Pointer(&v)), int(unsafe.Sizeof(v)))) {
		var zero T
		return zero, false
	}
	return v, true
}

// readProcessBytes fills buf with the target's bytes at addr; a short read
// counts as a failure.
func readProcessBytes(h windows.Handle, addr uintptr, buf []byte) bool {
	var read uintptr
	if err := windows.ReadProcessMemory(h, addr, &buf[0], uintptr(len(buf)), &read); err != nil {
		return false
	}
	return read == uintptr(len(buf))
}
