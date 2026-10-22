//go:build windows

package skillresidue

// Observe whether a metadata-only pin permits an ancestor move while keeping
// descriptor-relative no-replace operations on the original child. This probe
// does not change the production handle policy or the existing rename oracles.
import (
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestWindowsAncestorMetadataPinProbe(t *testing.T) {
	base := t.TempDir()
	parent := filepath.Join(base, "parent")
	child := filepath.Join(parent, "child")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(child, "payload"), []byte("owned"), 0o600); err != nil {
		t.Fatal(err)
	}
	root := openRootT(t, base)
	p, err := holdDir(root, "parent")
	if err != nil {
		t.Fatal(err)
	}
	defer p.close()
	c, err := holdDir(root, "parent/child")
	if err != nil {
		t.Fatal(err)
	}
	defer c.close()
	originalStat, err := c.file.Stat()
	if err != nil {
		t.Fatal(err)
	}
	var original windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(windows.Handle(c.file.Fd()), &original); err != nil {
		t.Fatal(err)
	}
	moved := parent + "-moved"
	baseline := os.Rename(parent, moved)
	t.Logf("ancestor rename while child os.Root/File held: %v", baseline)
	if baseline == nil {
		if err := os.Rename(moved, parent); err != nil {
			t.Fatal(err)
		}
	}
	const access = windows.FILE_READ_ATTRIBUTES | windows.SYNCHRONIZE
	const share = windows.FILE_SHARE_READ | windows.FILE_SHARE_WRITE | windows.FILE_SHARE_DELETE
	// Go's Windows OBJECT_ATTRIBUTES.init maps "." to an empty name relative
	// to the held directory. Obtain a new FileObject for that same directory
	// before closing either original read/list handle.
	name, err := windows.NewNTUnicodeString("")
	if err != nil {
		t.Fatal(err)
	}
	attrs := windows.OBJECT_ATTRIBUTES{RootDirectory: windows.Handle(c.file.Fd()), ObjectName: name}
	attrs.Length = uint32(unsafe.Sizeof(attrs))
	var h windows.Handle
	if err := windows.NtCreateFile(&h, access, &attrs, &windows.IO_STATUS_BLOCK{}, nil, 0, share, windows.FILE_OPEN,
		windows.FILE_DIRECTORY_FILE|windows.FILE_OPEN_REPARSE_POINT|windows.FILE_SYNCHRONOUS_IO_NONALERT, 0, 0); err != nil {
		t.Fatalf("open same child object before closing original handles: %v", err)
	}
	pin := os.NewFile(uintptr(h), "held-child-metadata")
	defer func() {
		if err := pin.Close(); err != nil {
			t.Error(err)
		}
	}()
	pinnedStat, err := pin.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if !originalStat.IsDir() || !pinnedStat.IsDir() || !os.SameFile(originalStat, pinnedStat) {
		t.Fatal("metadata pin did not preserve the opened directory identity")
	}
	var before windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &before); err != nil {
		t.Fatal(err)
	}
	if original.VolumeSerialNumber != before.VolumeSerialNumber || original.FileIndexHigh != before.FileIndexHigh || original.FileIndexLow != before.FileIndexLow {
		t.Fatal("reopen did not preserve the original child identity")
	}
	if err := c.file.Close(); err != nil {
		t.Fatal(err)
	}
	if err := c.root.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(parent, moved); err != nil {
		t.Fatalf("ancestor rename with same-object metadata child pin: %v", err)
	}
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(child, "payload"), []byte("foreign"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := noReplaceRenameAt(pin, "payload", p.file, "restored"); err != nil {
		t.Fatalf("rename relative to metadata-pinned child after ancestor replacement: %v", err)
	}
	if b, err := os.ReadFile(filepath.Join(moved, "restored")); err != nil || string(b) != "owned" {
		t.Fatalf("held destination lost original bytes: %q %v", b, err)
	}
	if b, err := os.ReadFile(filepath.Join(child, "payload")); err != nil || string(b) != "foreign" {
		t.Fatalf("replacement child changed: %q %v", b, err)
	}
	var after windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &after); err != nil {
		t.Fatal(err)
	}
	if before.VolumeSerialNumber != after.VolumeSerialNumber || before.FileIndexHigh != after.FileIndexHigh || before.FileIndexLow != after.FileIndexLow {
		t.Fatal("metadata pin identity changed during ancestor move")
	}
	t.Logf("metadata pin stayed bound: volume=%x id=%x:%x access=%d share=%d", after.VolumeSerialNumber, after.FileIndexHigh, after.FileIndexLow, access, share)
}
