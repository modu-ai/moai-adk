package privatefile

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
	_ "modernc.org/sqlite"
)

type sqliteNativeTokenMetadata struct {
	UserSID      string   `json:"user_sid"`
	DefaultOwner string   `json:"default_owner_sid"`
	Errors       []string `json:"errors,omitempty"`
}

type sqliteNativeACE struct {
	Type  uint8  `json:"type"`
	Flags uint8  `json:"flags"`
	Size  uint16 `json:"size"`
	Mask  uint32 `json:"mask"`
	SID   string `json:"sid,omitempty"`
	Error string `json:"error,omitempty"`
}

type sqliteNativeObjectMetadata struct {
	Stage             string            `json:"stage"`
	Role              string            `json:"role"`
	Exists            bool              `json:"exists"`
	FileID            string            `json:"file_id,omitempty"`
	Size              int64             `json:"size"`
	Attributes        uint32            `json:"attributes"`
	OwnerSID          string            `json:"owner_sid,omitempty"`
	OwnerDefaulted    bool              `json:"owner_defaulted"`
	Control           uint16            `json:"control"`
	Revision          uint32            `json:"revision"`
	Protected         bool              `json:"protected"`
	DACLPresent       bool              `json:"dacl_present"`
	DACLNull          bool              `json:"dacl_null"`
	DACLDefaulted     bool              `json:"dacl_defaulted"`
	ACEs              []sqliteNativeACE `json:"aces"`
	StrictHandleError string            `json:"strict_handle_error,omitempty"`
	StrictPathError   string            `json:"strict_path_error,omitempty"`
	Errors            []string          `json:"errors,omitempty"`
}

// TestSQLiteNativeStrictPrivacyLifecycle retains the production policy as its
// oracle. A policy failure stays RED while later stages still emit metadata.
// All files and SQLite rows are synthetic and remain under t.TempDir().
func TestSQLiteNativeStrictPrivacyLifecycle(t *testing.T) {
	token := sqliteNativeReadToken()
	sqliteNativeLog(t, "token", token)
	for _, err := range token.Errors {
		t.Errorf("native token metadata: %s", err)
	}
	t.Run("null_attributes_before_bytes", sqliteNativeEmptyCreation)
	for _, prepared := range []bool{false, true} {
		name := "driver_created"
		if prepared {
			name = "prepared_then_last_close"
		}
		t.Run(name, func(t *testing.T) { sqliteNativeLifecycle(t, prepared) })
	}
}

func sqliteNativeReadToken() sqliteNativeTokenMetadata {
	var result sqliteNativeTokenMetadata
	token := windows.GetCurrentProcessToken() // Pseudo token; never close it.
	user, err := token.GetTokenUser()
	if err != nil || user == nil || user.User.Sid == nil || !user.User.Sid.IsValid() {
		result.Errors = append(result.Errors, fmt.Sprintf("TokenUser: invalid SID or error: %v", err))
	} else {
		result.UserSID = user.User.Sid.String()
		if result.UserSID == "" {
			result.Errors = append(result.Errors, "TokenUser: SID string conversion failed")
		}
	}
	var size uint32
	err = windows.GetTokenInformation(token, windows.TokenOwner, nil, 0, &size)
	if !errors.Is(err, windows.ERROR_INSUFFICIENT_BUFFER) || size < uint32(unsafe.Sizeof(uintptr(0))) {
		result.Errors = append(result.Errors, fmt.Sprintf("TokenOwner size: %v, bytes=%d", err, size))
		return result
	}
	buffer := make([]byte, size)
	if err := windows.GetTokenInformation(token, windows.TokenOwner, &buffer[0], size, &size); err != nil {
		result.Errors = append(result.Errors, "TokenOwner: "+err.Error())
		return result
	}
	owner := *(**windows.SID)(unsafe.Pointer(&buffer[0]))
	start, address := uintptr(unsafe.Pointer(&buffer[0])), uintptr(unsafe.Pointer(owner))
	if owner == nil || address < start || address > start+uintptr(len(buffer))-8 {
		result.Errors = append(result.Errors, "TokenOwner: SID outside returned buffer")
	} else if owner.Len() > int(start+uintptr(len(buffer))-address) || !owner.IsValid() {
		result.Errors = append(result.Errors, "TokenOwner: invalid SID")
	} else {
		result.DefaultOwner = owner.String()
		if result.DefaultOwner == "" {
			result.Errors = append(result.Errors, "TokenOwner: SID string conversion failed")
		}
	}
	runtime.KeepAlive(buffer)
	return result
}

func sqliteNativeLog(t *testing.T, kind string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Errorf("encode native metadata: %v", err)
		return
	}
	t.Logf("sqlite-native-%s %s", kind, data)
}

// @MX:WARN: [AUTO] Collect every independent native readback failure without guessing an access policy.
// @MX:REASON: This diagnostic must keep later lifecycle observations available while strict policy assertions remain RED.
func sqliteNativeReadObject(path string, directory bool) (result sqliteNativeObjectMetadata) {
	if _, err := os.Lstat(path); err != nil {
		if !os.IsNotExist(err) {
			result.Errors = append(result.Errors, "Lstat: "+err.Error())
		}
		return result
	}
	result.Exists = true
	f, err := openExisting(path, windows.READ_CONTROL|windows.FILE_READ_ATTRIBUTES, directory)
	if err != nil {
		result.Errors = append(result.Errors, "open metadata: "+err.Error())
		return result
	}
	// openExisting shares DELETE as well as READ/WRITE. Close before returning,
	// so observing an artifact cannot pin it during SQLite's next close phase.
	defer func() {
		if err := f.Close(); err != nil {
			result.Errors = append(result.Errors, "close metadata: "+err.Error())
		}
	}()
	h := windows.Handle(f.Fd())
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &info); err != nil {
		result.Errors = append(result.Errors, "file identity: "+err.Error())
	} else {
		result.FileID = fmt.Sprintf("%x:%x:%x", info.VolumeSerialNumber, info.FileIndexHigh, info.FileIndexLow)
		result.Attributes = info.FileAttributes
	}
	if info, err := f.Stat(); err != nil {
		result.Errors = append(result.Errors, "handle stat: "+err.Error())
	} else {
		result.Size = info.Size()
	}
	if err := checkObjectHandle(h, directory); err != nil {
		result.StrictHandleError = err.Error()
	}
	sd, err := windows.GetSecurityInfo(h, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		result.Errors = append(result.Errors, "security info: "+err.Error())
		return result
	}
	owner, defaulted, err := sd.Owner()
	if err != nil || owner == nil || !owner.IsValid() {
		result.Errors = append(result.Errors, fmt.Sprintf("owner: %v", err))
	} else {
		result.OwnerSID, result.OwnerDefaulted = owner.String(), defaulted
		if result.OwnerSID == "" {
			result.Errors = append(result.Errors, "owner SID string conversion failed")
		}
	}
	control, revision, err := sd.Control()
	if err != nil {
		result.Errors = append(result.Errors, "control: "+err.Error())
	} else {
		result.Control, result.Revision = uint16(control), revision
		result.Protected = control&windows.SE_DACL_PROTECTED != 0
		result.DACLPresent = control&windows.SE_DACL_PRESENT != 0
	}
	acl, defaulted, err := sd.DACL()
	if err != nil {
		result.Errors = append(result.Errors, "DACL: "+err.Error())
		return result
	}
	result.DACLNull, result.DACLDefaulted = acl == nil, defaulted
	if acl == nil {
		return result
	}
	for index := uint32(0); index < uint32(acl.AceCount); index++ {
		result.ACEs = append(result.ACEs, sqliteNativeReadACE(acl, index))
	}
	return result
}

func sqliteNativeReadACE(acl *windows.ACL, index uint32) sqliteNativeACE {
	var ace *windows.ACCESS_ALLOWED_ACE
	if err := windows.GetAce(acl, index, &ace); err != nil {
		return sqliteNativeACE{Error: err.Error()}
	}
	if ace == nil {
		return sqliteNativeACE{Error: "nil ACE"}
	}
	result := sqliteNativeACE{Type: ace.Header.AceType, Flags: ace.Header.AceFlags, Size: ace.Header.AceSize}
	offset := int(unsafe.Offsetof(ace.SidStart))
	if (result.Type != windows.ACCESS_ALLOWED_ACE_TYPE && result.Type != windows.ACCESS_DENIED_ACE_TYPE) || int(result.Size) < offset+8 {
		result.Error = "unsupported or short ACE layout"
		return result
	}
	result.Mask = uint32(ace.Mask)
	sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
	if sid.Len() > int(result.Size)-offset || !sid.IsValid() {
		result.Error = "invalid or out-of-bounds ACE SID"
		return result
	}
	result.SID = sid.String()
	if result.SID == "" {
		result.Error = "ACE SID string conversion failed"
	}
	return result
}

// @MX:ANCHOR: [AUTO] Emit native metadata before enforcing the unchanged strict privacy oracle.
// @MX:REASON: Private-root setup, empty NULL-attributes creation and SQLite artifact stages share this observation boundary.
func sqliteNativeObserve(t *testing.T, stage, role, path string, directory, required bool) {
	t.Helper()
	metadata := sqliteNativeReadObject(path, directory)
	metadata.Stage, metadata.Role = stage, role
	if metadata.Exists {
		check := Check
		if directory {
			check = CheckDirectory
		}
		if err := check(path); err != nil {
			metadata.StrictPathError = err.Error()
		}
	}
	sqliteNativeLog(t, "object", metadata)
	if len(metadata.Errors) != 0 || metadata.StrictHandleError != "" || metadata.StrictPathError != "" {
		t.Errorf("strict native policy stage=%s role=%s: handle=%q path=%q metadata=%v", stage, role, metadata.StrictHandleError, metadata.StrictPathError, metadata.Errors)
	}
	for _, ace := range metadata.ACEs {
		if ace.Error != "" {
			t.Errorf("native ACE stage=%s role=%s: %s", stage, role, ace.Error)
		}
	}
	if required && !metadata.Exists {
		t.Errorf("required artifact absent: stage=%s role=%s", stage, role)
	}
}

func sqliteNativeRoot(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "private-parent")
	if err := Mkdir(path); err != nil {
		t.Errorf("private parent creation: %v", err)
		sqliteNativeObserve(t, "parent-creation-failed", "parent", path, true, true)
		return ""
	}
	sqliteNativeObserve(t, "before-any-bytes", "parent", path, true, true)
	return path
}

func sqliteNativeEmptyCreation(t *testing.T) {
	dir := sqliteNativeRoot(t)
	if dir == "" {
		return
	}
	path := filepath.Join(dir, "empty-native-file")
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	// Match the driver's writable NULL-attributes OPEN_ALWAYS call. This
	// empty creation witness receives no payload and changes no owner/policy.
	h, err := windows.CreateFile(p, windows.GENERIC_READ|windows.GENERIC_WRITE, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_ALWAYS, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		t.Errorf("NULL-attributes creation: %v", err)
		return
	}
	f := os.NewFile(uintptr(h), path)
	t.Cleanup(func() { sqliteNativeCloseFile(t, f) })
	sqliteNativeObserve(t, "null-attributes-before-bytes", "empty-file", path, false, true)
	sqliteNativeCloseFile(t, f)
}

func sqliteNativeCloseFile(t *testing.T, f *os.File) {
	t.Helper()
	if err := f.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
		t.Errorf("close native fixture: %v", err)
	}
}

func sqliteNativePrepare(t *testing.T, path string) {
	t.Helper()
	for _, suffix := range []string{"", "-wal", "-shm"} {
		f, err := OpenAppend(path + suffix)
		if err != nil {
			t.Errorf("no-write preparation %q: %v", suffix, err)
			continue
		}
		sqliteNativeCloseFile(t, f)
	}
}

func sqliteNativeArtifacts(t *testing.T, stage, path string, sidecarsRequired bool) {
	t.Helper()
	for _, artifact := range []struct{ role, suffix string }{{"db", ""}, {"wal", "-wal"}, {"shm", "-shm"}} {
		sqliteNativeObserve(t, stage, artifact.role, path+artifact.suffix, false, artifact.suffix == "" || sidecarsRequired)
	}
}

func sqliteNativeOpen(t *testing.T, path string) *sql.DB {
	t.Helper()
	p := filepath.ToSlash(path)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	values := url.Values{"_pragma": {"journal_mode(WAL)", "busy_timeout(5000)"}}
	dsn := (&url.URL{Scheme: "file", Path: p, RawQuery: values.Encode()}).String()
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Errorf("open synthetic SQLite: %v", err)
		return nil
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("cleanup synthetic SQLite: %v", err)
		}
	})
	return db
}

// @MX:WARN: [AUTO] Preserve ordered SQLite lifecycle observations after privacy assertions fail.
// @MX:REASON: The before-close preparation gap and later recreation must remain observable without retries or policy weakening.
func sqliteNativeLifecycle(t *testing.T, prepared bool) {
	dir := sqliteNativeRoot(t)
	if dir == "" {
		return
	}
	path := filepath.Join(dir, "probe.db")
	if prepared {
		sqliteNativePrepare(t, path)
		sqliteNativeArtifacts(t, "prepared-before-SQL", path, true)
	}
	for phase := 1; phase <= 3; phase++ {
		db := sqliteNativeOpen(t, path)
		if db == nil {
			sqliteNativeArtifacts(t, fmt.Sprintf("open-failed-%d", phase), path, false)
			return
		}
		_, writeErr := db.ExecContext(t.Context(), `CREATE TABLE IF NOT EXISTS probe(n INTEGER NOT NULL); INSERT INTO probe(n) VALUES(1)`)
		if writeErr != nil {
			t.Errorf("synthetic SQLite write phase=%d: %v", phase, writeErr)
		}
		sqliteNativeArtifacts(t, fmt.Sprintf("writer-open-%d", phase), path, true)
		var rows int
		if err := db.QueryRowContext(t.Context(), `SELECT count(*) FROM probe`).Scan(&rows); err != nil || rows != phase {
			t.Errorf("synthetic row continuity phase=%d count=%d err=%v", phase, rows, err)
		}
		if prepared && phase == 1 {
			// Model the next opener's preparation before the last live
			// connection closes. No timing retries or concurrent load needed.
			sqliteNativePrepare(t, path)
			sqliteNativeArtifacts(t, "prepared-before-last-close", path, true)
		}
		closeErr := db.Close()
		if closeErr != nil {
			t.Errorf("close synthetic SQLite phase=%d: %v", phase, closeErr)
		}
		sqliteNativeArtifacts(t, fmt.Sprintf("after-close-attempt-%d", phase), path, false)
		if closeErr != nil {
			// Do not remove sidecars or claim a last-connection boundary
			// when the native close itself failed. Cleanup will retry Close.
			return
		}
		if phase == 2 {
			// Only closed, test-owned sidecars are removed. This forces a
			// recreation witness even if this SQLite build retained them.
			for _, suffix := range []string{"-wal", "-shm"} {
				if err := os.Remove(path + suffix); err != nil && !os.IsNotExist(err) {
					t.Errorf("remove closed synthetic sidecar %s: %v", suffix, err)
				}
			}
			sqliteNativeArtifacts(t, "closed-before-forced-recreation", path, false)
		}
	}
}
