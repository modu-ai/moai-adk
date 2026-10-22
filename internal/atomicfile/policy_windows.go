package atomicfile

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

const filePolicyInformation = windows.OWNER_SECURITY_INFORMATION | windows.GROUP_SECURITY_INFORMATION | windows.DACL_SECURITY_INFORMATION

// @MX:ANCHOR: Windows atomic rewrite policy shared by YAML, updates and wrapper migration.
// @MX:REASON: Three production callers require native owner/group/DACL checks before writing staged bytes.
// CreatePolicyTemp creates an exclusive, random stage beside target, with its
// native permission policy established and checked before the caller writes.
// Existing regular files keep their observed owner, group and DACL. A nonzero
// override replaces the DACL with the requested mode while retaining owner and
// group; new files use defaultMode and the current token's owner/group.
//
// A final reparse point is never followed or modified: its replacement gets a
// new-file policy. Unsupported security, readonly originals and insufficient
// ordinary rights fail closed. No privilege is enabled, source ACL changed or
// SACL preservation claimed. Callers own write/close/publication and cleanup.
func CreatePolicyTemp(target, pattern string, defaultMode, override os.FileMode) (*os.File, error) {
	sd, err := observedFilePolicy(target)
	if err != nil {
		return nil, err
	}
	if sd == nil || override != 0 {
		mode := defaultMode
		if override != 0 {
			mode = override
		}
		sd, err = modeFilePolicy(sd, mode)
		if err != nil {
			return nil, err
		}
	}
	want := sd.String()
	if want == "" {
		return nil, errors.New("stage security descriptor cannot be represented")
	}
	if strings.ContainsAny(pattern, `/\`) {
		return nil, errors.New("stage pattern contains a path separator")
	}
	prefix, suffix := pattern, ""
	if i := strings.LastIndexByte(pattern, '*'); i >= 0 {
		prefix, suffix = pattern[:i], pattern[i+1:]
	}
	// One cryptographic name and one CREATE_NEW: a collision is a refusal,
	// never an overwrite or a new retry budget.
	path := filepath.Join(filepath.Dir(target), prefix+rand.Text()+suffix)
	name, err := policyNativePath(path)
	if err != nil {
		return nil, err
	}
	sa := windows.SecurityAttributes{Length: uint32(unsafe.Sizeof(windows.SecurityAttributes{})), SecurityDescriptor: sd}
	h, err := windows.CreateFile(name, windows.GENERIC_WRITE|windows.READ_CONTROL|windows.WRITE_DAC,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, &sa,
		windows.CREATE_NEW, windows.FILE_ATTRIBUTE_NORMAL|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return nil, &os.PathError{Op: "create policy stage", Path: path, Err: err}
	}
	f := os.NewFile(uintptr(h), path)
	if err := checkStagedPolicy(h, want); err != nil {
		closeErr := f.Close()
		removeErr := os.Remove(path)
		return nil, errors.Join(err, closeErr, removeErr)
	}
	return f, nil
}

func policyNativePath(path string) (*uint16, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(abs, `\\?\`) {
		if strings.HasPrefix(abs, `\\`) {
			abs = `\\?\UNC\` + strings.TrimPrefix(abs, `\\`)
		} else {
			abs = `\\?\` + abs
		}
	}
	return windows.UTF16PtrFromString(abs)
}

func observedFilePolicy(path string) (sd *windows.SECURITY_DESCRIPTOR, result error) {
	name, err := policyNativePath(path)
	if err != nil {
		return nil, err
	}
	h, err := windows.CreateFile(name, windows.READ_CONTROL|windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil,
		windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if errors.Is(err, windows.ERROR_FILE_NOT_FOUND) || errors.Is(err, windows.ERROR_PATH_NOT_FOUND) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("observe destination security: %w", err)
	}
	defer func() { result = errors.Join(result, windows.CloseHandle(h)) }()
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &info); err != nil {
		return nil, err
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		return nil, nil
	}
	if info.FileAttributes&(windows.FILE_ATTRIBUTE_DIRECTORY|windows.FILE_ATTRIBUTE_READONLY) != 0 {
		return nil, errors.New("destination policy requires a writable regular file")
	}
	sd, err = windows.GetSecurityInfo(h, windows.SE_FILE_OBJECT, filePolicyInformation)
	if err != nil {
		return nil, err
	}
	if sd == nil || !sd.IsValid() {
		return nil, errors.New("destination security descriptor is unavailable")
	}
	owner, _, err := sd.Owner()
	if err != nil || owner == nil || !owner.IsValid() {
		return nil, errors.Join(errors.New("destination owner is unavailable"), err)
	}
	group, _, err := sd.Group()
	if err != nil || group == nil || !group.IsValid() {
		return nil, errors.Join(errors.New("destination group is unavailable"), err)
	}
	acl, _, err := sd.DACL()
	if err != nil || acl == nil {
		return nil, errors.Join(errors.New("destination DACL is unavailable or null"), err)
	}
	return sd, nil
}

func modeFilePolicy(existing *windows.SECURITY_DESCRIPTOR, mode os.FileMode) (*windows.SECURITY_DESCRIPTOR, error) {
	ownerRights := uint32(windows.FILE_GENERIC_READ | windows.FILE_GENERIC_WRITE | windows.DELETE | windows.WRITE_DAC)
	publicRights := uint32(0)
	switch mode {
	case 0o600:
	case 0o644:
		publicRights = windows.FILE_GENERIC_READ
	case 0o755:
		ownerRights |= windows.FILE_GENERIC_EXECUTE
		publicRights = windows.FILE_GENERIC_READ | windows.FILE_GENERIC_EXECUTE
	default:
		return nil, fmt.Errorf("unsupported native file permission %04o", mode)
	}
	var owner, group *windows.SID
	var err error
	if existing != nil {
		owner, _, err = existing.Owner()
		if err == nil {
			group, _, err = existing.Group()
		}
	} else {
		user, userErr := windows.GetCurrentProcessToken().GetTokenUser()
		if userErr != nil {
			return nil, userErr
		}
		primary, groupErr := windows.GetCurrentProcessToken().GetTokenPrimaryGroup()
		if groupErr != nil {
			return nil, groupErr
		}
		owner, group = user.User.Sid, primary.PrimaryGroup
	}
	if err != nil || owner == nil || group == nil || !owner.IsValid() || !group.IsValid() {
		return nil, errors.Join(errors.New("file policy owner/group is unavailable"), err)
	}
	sddl := fmt.Sprintf("O:%sG:%sD:P(A;;0x%x;;;%s)", owner, group, ownerRights, owner)
	if publicRights != 0 {
		sddl += fmt.Sprintf("(A;;0x%x;;;WD)", publicRights)
	}
	return windows.SecurityDescriptorFromString(sddl)
}

func checkStagedPolicy(h windows.Handle, want string) error {
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &info); err != nil {
		return err
	}
	if info.FileAttributes&(windows.FILE_ATTRIBUTE_DIRECTORY|windows.FILE_ATTRIBUTE_REPARSE_POINT|windows.FILE_ATTRIBUTE_READONLY) != 0 || info.NumberOfLinks != 1 {
		return errors.New("policy stage is not a writable regular single-link file")
	}
	actual, err := windows.GetSecurityInfo(h, windows.SE_FILE_OBJECT, filePolicyInformation)
	if err != nil {
		return err
	}
	if actual == nil {
		return errors.New("stage security descriptor is unavailable")
	}
	// Even an empty stage may already have been opened by another reader.
	// Tightening an unexpectedly broad DACL now cannot revoke that handle.
	// Require correct creation; never repair permissions after publication of
	// the stage name and then put data behind a previously acquired handle.
	if actual.String() != want {
		return errors.New("kernel did not preserve the requested stage owner/group/DACL")
	}
	return nil
}
