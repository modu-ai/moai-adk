package codexbridge

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
	"github.com/modu-ai/moai-adk/internal/codexapp"
	"github.com/modu-ai/moai-adk/internal/codextools"
)

// FileStore records recovery barriers, never credentials, reasoning or tool
// results. Its directory must be protected by the owner's App Server profile
// lease for the Engine lifetime; separate processes must not share a store.
// Records deliberately cannot be resumed by a new Engine in AS3.
type FileStore struct {
	mu  sync.Mutex
	dir string
}
type record struct {
	Owner                         codextools.Binding
	Phase, TurnID, CallID, Prefix string
}

func OpenStore(dir string) (*FileStore, error) {
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return nil, err
	}
	if err = codexapp.ValidatePrivatePath(resolved, true); err != nil {
		return nil, err
	}
	return &FileStore{dir: resolved}, nil
}
func storeKey(owner codextools.Binding) string {
	sum := sha256.Sum256([]byte(owner.AccountScope + "\x00" + owner.ConversationID))
	return hex.EncodeToString(sum[:])
}
func (s *FileStore) exists(owner codextools.Binding) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := codexapp.ValidatePrivatePath(s.dir, true); err != nil {
		return false, err
	}
	path := filepath.Join(s.dir, storeKey(owner)+".json")
	_, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err = codexapp.ValidatePrivatePath(path, false); err != nil {
		return false, err
	}
	return true, nil
}
func (s *FileStore) save(r record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := codexapp.ValidatePrivatePath(s.dir, true); err != nil {
		return err
	}
	raw, err := json.Marshal(r)
	if err != nil {
		return err
	}
	path := filepath.Join(s.dir, storeKey(r.Owner)+".json")
	if _, err := os.Lstat(path); err == nil {
		if err = codexapp.ValidatePrivatePath(path, false); err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	file, err := os.CreateTemp(s.dir, ".phase-*")
	if err != nil {
		return err
	}
	tmp := file.Name()
	defer os.Remove(tmp)
	if _, err = file.Write(raw); err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if err = codexapp.ValidatePrivatePath(tmp, false); err != nil {
		return err
	}
	if err = atomicfile.Replace(tmp, path); err != nil {
		return err
	}
	if err = codexapp.ValidatePrivatePath(path, false); err != nil {
		return err
	}
	got, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !bytes.Equal(got, raw) {
		return errors.New("bridge state readback mismatch")
	}
	if runtime.GOOS != "windows" {
		dir, err := os.Open(s.dir)
		if err != nil {
			return err
		}
		err = dir.Sync()
		closeErr = dir.Close()
		if err != nil {
			return err
		}
		return closeErr
	}
	return nil
}
