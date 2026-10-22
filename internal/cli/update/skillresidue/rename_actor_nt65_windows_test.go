//go:build windows

package skillresidue

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestWindowsNT65RenameInputUsesNativeLayout(t *testing.T) {
	for _, name := range []string{"b", "parent-moved", "이름", "𐐷"} {
		for _, flags := range []uint32{0, windows.FILE_RENAME_REPLACE_IF_EXISTS | windows.FILE_RENAME_POSIX_SEMANTICS} {
			const root = windows.Handle(0x1234)
			buf, err := renameInformationBuffer(root, name)
			if err != nil {
				t.Fatal(err)
			}
			info := (*renameInformation)(unsafe.Pointer(&buf[0]))
			info.ReplaceIfExists = flags
			encoded := utf16.Encode([]rune(name))
			if info.ReplaceIfExists != flags || info.RootDirectory != root || info.FileNameLength != uint32(len(encoded)*2) {
				t.Fatalf("NT65 flags/root/name length changed: %#v", info)
			}
			if len(buf) < int(unsafe.Sizeof(*info))+len(encoded)*2 {
				t.Fatalf("NT65 input lacks native size and filename bytes: %d", len(buf))
			}
			if got := utf16.Decode(unsafe.Slice(&info.FileName[0], len(encoded))); string(got) != name {
				t.Fatalf("NT65 input filename=%q, want %q", string(got), name)
			}
			t.Logf("nt65-layout name=%q flags=%#x sizeof=%d offset=%d filename-bytes=%d buffer=%d", name, flags, unsafe.Sizeof(*info), unsafe.Offsetof(info.FileName), info.FileNameLength, len(buf))
		}
	}
	if buf, err := renameInformationBuffer(0, "bad\x00name"); err == nil || buf != nil {
		t.Fatalf("NT65 input accepted an embedded NUL: buffer=%v error=%v", buf, err)
	}
}

// This normal NT65 observation does not replace any existing rename oracle.
// Keep the source open identical to the Win32 matrix to isolate the setter.
// Never fall back to class 10 or use kernel-only bypass class 66.
// A held-child case may observe ACCESS_DENIED only after proving that names,
// identities and bytes stayed unchanged. Such an observation is not evidence
// that the original move/decoy attack or the production repair succeeded.
func TestWindowsAncestorNT65Matrix(t *testing.T) {
	for _, tc := range []struct {
		name, hold string
		flags      uint32
	}{
		{"zero-ex0", "none", 0},
		{"zero-ex3", "none", windows.FILE_RENAME_REPLACE_IF_EXISTS | windows.FILE_RENAME_POSIX_SEMANTICS},
		{"original-ex0", "original", 0},
		{"original-ex3", "original", windows.FILE_RENAME_REPLACE_IF_EXISTS | windows.FILE_RENAME_POSIX_SEMANTICS},
		{"metadata-ex0", "metadata", 0},
		{"metadata-ex3", "metadata", windows.FILE_RENAME_REPLACE_IF_EXISTS | windows.FILE_RENAME_POSIX_SEMANTICS},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const api = "NtSetInformationFile/FileRenameInformationEx65"
			base := t.TempDir()
			parent := filepath.Join(base, "parent")
			child := filepath.Join(parent, "child")
			payload := filepath.Join(child, "payload")
			moved := parent + "-moved"
			if err := os.MkdirAll(child, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(payload, []byte("owned"), 0o600); err != nil {
				t.Fatal(err)
			}
			root := openRootT(t, base)
			baseFile, err := root.Open(".")
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := baseFile.Close(); err != nil {
					t.Error(err)
				}
			}()
			p, err := holdDir(root, "parent")
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := p.file.Close(); err != nil {
					t.Error(err)
				}
				if err := p.root.Close(); err != nil {
					t.Error(err)
				}
			}()
			c, err := holdDir(root, "parent/child")
			if err != nil {
				t.Fatal(err)
			}
			childOpen := true
			defer func() {
				if childOpen {
					if err := c.file.Close(); err != nil {
						t.Error(err)
					}
					if err := c.root.Close(); err != nil {
						t.Error(err)
					}
				}
			}()
			statPath := func(path string) os.FileInfo {
				t.Helper()
				f, err := os.Open(path)
				if err != nil {
					t.Fatal(err)
				}
				info, statErr := f.Stat()
				closeErr := f.Close()
				if statErr != nil || closeErr != nil {
					t.Fatalf("capture handle-derived identity: stat=%v close=%v", statErr, closeErr)
				}
				return info
			}
			mustSame := func(path string, before os.FileInfo) {
				t.Helper()
				if !os.SameFile(before, statPath(path)) {
					t.Fatalf("object identity changed at %s", path)
				}
			}
			mustBytes := func(path, want string) {
				t.Helper()
				if data, err := os.ReadFile(path); err != nil || string(data) != want {
					t.Fatalf("payload at %s=%q, %v; want %q", path, data, err, want)
				}
			}
			mustAbsent := func(path string) {
				t.Helper()
				if _, err := os.Lstat(path); !errors.Is(err, fs.ErrNotExist) {
					t.Fatalf("unexpected namespace entry %s: %v", path, err)
				}
			}
			mustOnly := func(path, name string) {
				t.Helper()
				entries, err := os.ReadDir(path)
				if err != nil || len(entries) != 1 || entries[0].Name() != name {
					t.Fatalf("namespace at %s=%v, %v; want only %s", path, entries, err, name)
				}
			}
			parentInfo, err := p.file.Stat()
			if err != nil {
				t.Fatal(err)
			}
			childInfo, err := c.file.Stat()
			if err != nil {
				t.Fatal(err)
			}
			payloadInfo := statPath(payload)
			var childID windows.ByHandleFileInformation
			if err := windows.GetFileInformationByHandle(windows.Handle(c.file.Fd()), &childID); err != nil {
				t.Fatal(err)
			}
			pin := c.file
			if tc.hold == "metadata" {
				name, err := windows.NewNTUnicodeString("")
				if err != nil {
					t.Fatal(err)
				}
				attrs := windows.OBJECT_ATTRIBUTES{RootDirectory: windows.Handle(c.file.Fd()), ObjectName: name}
				attrs.Length = uint32(unsafe.Sizeof(attrs))
				var h windows.Handle
				if err := windows.NtCreateFile(&h, windows.FILE_READ_ATTRIBUTES|windows.SYNCHRONIZE, &attrs, &windows.IO_STATUS_BLOCK{}, nil, 0,
					windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, windows.FILE_OPEN,
					windows.FILE_DIRECTORY_FILE|windows.FILE_OPEN_REPARSE_POINT|windows.FILE_SYNCHRONOUS_IO_NONALERT, 0, 0); err != nil {
					t.Fatalf("same-object metadata pin open: %v", err)
				}
				pin = os.NewFile(uintptr(h), "nt65-metadata-child")
				defer func() {
					if err := pin.Close(); err != nil {
						t.Error(err)
					}
				}()
				if info, err := pin.Stat(); err != nil || !os.SameFile(childInfo, info) {
					t.Fatalf("metadata pin identity differs from original: %v", err)
				}
			}
			if tc.hold != "original" {
				if err := c.file.Close(); err != nil {
					t.Fatal(err)
				}
				if err := c.root.Close(); err != nil {
					t.Fatal(err)
				}
				childOpen = false
				if tc.hold == "none" {
					pin = nil
				}
			}
			mustAbsent(moved)
			err = func() error {
				name, err := windows.UTF16PtrFromString(parent)
				if err != nil {
					t.Fatal(err)
				}
				h, err := windows.CreateFile(name, windows.DELETE|windows.FILE_READ_ATTRIBUTES|windows.SYNCHRONIZE,
					windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING,
					windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
				if err != nil {
					t.Fatalf("open normal DELETE-access actor handle: %v", err)
				}
				actor := os.NewFile(uintptr(h), "nt65-actor-parent")
				defer func() {
					if err := actor.Close(); err != nil {
						t.Error(err)
					}
				}()
				if info, err := actor.Stat(); err != nil || !os.SameFile(parentInfo, info) {
					t.Fatalf("actor opened another parent: %v", err)
				}
				buf, err := renameInformationBuffer(windows.Handle(baseFile.Fd()), "parent-moved")
				if err != nil {
					t.Fatal(err)
				}
				info := (*renameInformation)(unsafe.Pointer(&buf[0]))
				info.ReplaceIfExists = tc.flags
				const fileRenameInformationEx = 65
				var iosb windows.IO_STATUS_BLOCK
				// Preserve the raw status before converting it for errors.Is.
				err = windows.NtSetInformationFile(h, &iosb, &buf[0], uint32(len(buf)), fileRenameInformationEx)
				runtime.KeepAlive(baseFile)
				runtime.KeepAlive(buf)
				mapped := moveError(err)
				t.Logf("nt65-input class=%d flags=%#x sizeof=%d offset=%d filename-bytes=%d buffer=%d iosb-status=%#x iosb-information=%d",
					fileRenameInformationEx, tc.flags, unsafe.Sizeof(*info), unsafe.Offsetof(info.FileName), info.FileNameLength, len(buf), uint32(iosb.Status), iosb.Information)
				if status, ok := err.(windows.NTStatus); ok {
					t.Logf("nt65-result ntstatus=%#x errno=%d error=%v", uint32(status), status.Errno(), mapped)
				} else {
					t.Logf("nt65-result raw=%v mapped=%v", err, mapped)
				}
				return mapped
			}()
			if err != nil {
				if tc.hold == "none" || !errors.Is(err, windows.ERROR_ACCESS_DENIED) {
					t.Fatalf("actor must move in zero-child controls; only held-child ACCESS_DENIED is an observation: %v", err)
				}
				mustOnly(base, "parent")
				mustOnly(parent, "child")
				mustOnly(child, "payload")
				mustAbsent(moved)
				mustSame(parent, parentInfo)
				mustSame(child, childInfo)
				mustSame(payload, payloadInfo)
				mustBytes(payload, "owned")
				if info, statErr := pin.Stat(); statErr != nil || !os.SameFile(childInfo, info) {
					t.Fatalf("denied actor changed pin identity: %v", statErr)
				}
				t.Logf("ancestor-actor hold=%s api=%s flags=%#x capability=denied namespace=unchanged identity=unchanged bytes=owned error=%v", tc.hold, api, tc.flags, err)
				return
			}
			mustAbsent(parent)
			mustOnly(base, "parent-moved")
			mustSame(moved, parentInfo)
			movedChild := filepath.Join(moved, "child")
			mustSame(movedChild, childInfo)
			mustSame(filepath.Join(movedChild, "payload"), payloadInfo)
			mustBytes(filepath.Join(movedChild, "payload"), "owned")
			if tc.hold == "none" {
				t.Logf("ancestor-actor hold=none api=%s flags=%#x capability=exercised move=real identity=unchanged bytes=owned", api, tc.flags)
				return
			}
			if err := os.MkdirAll(child, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(payload, []byte("foreign"), 0o600); err != nil {
				t.Fatal(err)
			}
			decoyInfo := statPath(payload)
			if os.SameFile(childInfo, statPath(child)) || os.SameFile(payloadInfo, decoyInfo) {
				t.Fatal("decoy unexpectedly names an original object")
			}
			if info, err := pin.Stat(); err != nil || !os.SameFile(childInfo, info) {
				t.Fatalf("successful actor changed held-child identity: %v", err)
			}
			if err := noReplaceRenameAt(pin, "payload", p.file, "restored"); err != nil {
				t.Fatalf("held-child no-replace after move and decoy: %v", err)
			}
			restored := filepath.Join(moved, "restored")
			mustSame(restored, payloadInfo)
			mustAbsent(filepath.Join(movedChild, "payload"))
			mustBytes(restored, "owned")
			mustSame(payload, decoyInfo)
			mustBytes(payload, "foreign")
			blocked := filepath.Join(movedChild, "blocked")
			if err := os.WriteFile(blocked, []byte("candidate"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := noReplaceRenameAt(pin, "blocked", p.file, "restored"); !errors.Is(err, fs.ErrExist) {
				t.Fatalf("existing original target was not refused: %v", err)
			}
			mustBytes(blocked, "candidate")
			mustSame(restored, payloadInfo)
			mustBytes(restored, "owned")
			mustSame(payload, decoyInfo)
			mustBytes(payload, "foreign")
			if info, err := pin.Stat(); err != nil || !os.SameFile(childInfo, info) {
				t.Fatalf("final held-child identity changed: %v", err)
			}
			t.Logf("ancestor-actor hold=%s api=%s flags=%#x capability=exercised move=real decoy=preserved no-replace=preserved volume=%x child-id=%x:%x bytes=owned", tc.hold, api, tc.flags, childID.VolumeSerialNumber, childID.FileIndexHigh, childID.FileIndexLow)
		})
	}
}
