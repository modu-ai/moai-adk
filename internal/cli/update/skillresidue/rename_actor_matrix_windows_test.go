//go:build windows

package skillresidue

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

// This independent observation does not replace any existing rename oracle.
// A held-child case may observe ACCESS_DENIED only after proving that names,
// identities and bytes stayed unchanged. Such an observation is not evidence
// that the original move/decoy attack or the production repair succeeded.
func TestWindowsAncestorActorMatrix(t *testing.T) {
	for _, tc := range []struct {
		name, hold string
		classic    bool
		flags      uint32
	}{
		{"zero-classic", "none", true, windows.MOVEFILE_REPLACE_EXISTING},
		{"zero-ex0", "none", false, 0},
		{"zero-ex3", "none", false, windows.FILE_RENAME_REPLACE_IF_EXISTS | windows.FILE_RENAME_POSIX_SEMANTICS},
		{"original-ex0", "original", false, 0},
		{"original-ex3", "original", false, windows.FILE_RENAME_REPLACE_IF_EXISTS | windows.FILE_RENAME_POSIX_SEMANTICS},
		{"metadata-ex0", "metadata", false, 0},
		{"metadata-ex3", "metadata", false, windows.FILE_RENAME_REPLACE_IF_EXISTS | windows.FILE_RENAME_POSIX_SEMANTICS},
	} {
		t.Run(tc.name, func(t *testing.T) {
			api := "FileRenameInfoEx"
			if tc.classic {
				api = "MoveFileExW"
			}
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
				pin = os.NewFile(uintptr(h), "matrix-metadata-child")
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
				if tc.classic {
					return os.Rename(parent, moved)
				}
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
				actor := os.NewFile(uintptr(h), "matrix-actor-parent")
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
				// This is Win32 FILE_INFO_BY_HANDLE_CLASS 22, not NT class 22
				// or the kernel-only access-check bypass class 66.
				(*renameInformation)(unsafe.Pointer(&buf[0])).ReplaceIfExists = tc.flags
				err = windows.SetFileInformationByHandle(h, windows.FileRenameInfoEx, &buf[0], uint32(len(buf)))
				runtime.KeepAlive(baseFile)
				return err
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
