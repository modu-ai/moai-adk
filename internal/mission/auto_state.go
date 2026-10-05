package mission

import (
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/modu-ai/moai-adk/internal/atomicfile"
	"github.com/modu-ai/moai-adk/internal/goal"
)

type MissionMode string

const ModeAuto MissionMode = "auto"

const autoMissionStateDir = ".moai/state/mission"

type AutoMission struct {
	SessionID       string               `json:"session_id"`
	Text            string               `json:"text"`
	MissionMode     MissionMode          `json:"mission_mode"`
	ProgressionMode goal.ProgressionMode `json:"progression_mode"`
	State           MissionState         `json:"state"`
	Contract        *MissionContract     `json:"contract,omitempty"`
	ContractHash    string               `json:"contract_hash,omitempty"`
	Snapshot        *MissionSnapshot     `json:"snapshot,omitempty"`
	LastBlocker     string               `json:"last_blocker,omitempty"`
	OperationIDs    []string             `json:"operation_ids,omitempty"`
}

func autoMissionPath(projectRoot, sessionID string) string {
	return filepath.Join(projectRoot, autoMissionStateDir, sessionID+".json")
}

func validateMissionSessionID(sessionID string) error {
	parsed, err := uuid.Parse(sessionID)
	if err != nil || parsed == uuid.Nil || parsed.String() != sessionID || strings.ToLower(sessionID) != sessionID {
		return errors.New("auto mission: invalid_session_id")
	}
	return nil
}

func ValidateMissionSessionID(sessionID string) error { return validateMissionSessionID(sessionID) }

func ValidateAutoMissionIntegrity(state *AutoMission) error {
	if state == nil || validateMissionSessionID(state.SessionID) != nil || state.MissionMode != ModeAuto {
		return errors.New("contract_integrity_mismatch")
	}
	if state.Contract == nil {
		if state.ContractHash != "" || state.Snapshot != nil || len(state.OperationIDs) != 0 {
			return errors.New("contract_integrity_mismatch")
		}
		return nil
	}
	sealed, err := SealMissionContract(*state.Contract)
	if err != nil || sealed.Version != state.Contract.PolicyVersion || len(sealed.Hash) != len(state.ContractHash) || subtle.ConstantTimeCompare([]byte(sealed.Hash), []byte(state.ContractHash)) != 1 {
		return errors.New("contract_integrity_mismatch")
	}
	if state.Snapshot != nil && (state.Snapshot.MissionID != state.SessionID || state.Snapshot.ContractHash != state.ContractHash || state.Snapshot.PolicyVersion != sealed.Version || state.Snapshot.SnapshotHash == "") {
		return errors.New("snapshot_lineage_mismatch")
	}
	for _, operationID := range state.OperationIDs {
		if len(operationID) != len("op-")+32 || operationID[:3] != "op-" {
			return errors.New("operation_lineage_mismatch")
		}
		if _, err := hex.DecodeString(operationID[3:]); err != nil {
			return errors.New("operation_lineage_mismatch")
		}
	}
	return nil
}

func secureMissionPath(projectRoot, sessionID string, createDir bool) (string, error) {
	if projectRoot == "" {
		return "", errors.New("auto mission: invalid_project_root")
	}
	if err := validateMissionSessionID(sessionID); err != nil {
		return "", err
	}
	root, err := filepath.Abs(projectRoot)
	if err != nil {
		return "", err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	dir := root
	for _, component := range []string{".moai", "state", "mission"} {
		dir = filepath.Join(dir, component)
		info, statErr := os.Lstat(dir)
		if statErr == nil {
			if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
				return "", errors.New("auto mission: unsafe_state_path")
			}
			continue
		}
		if !errors.Is(statErr, fs.ErrNotExist) {
			return "", statErr
		}
		if !createDir {
			return filepath.Join(dir, sessionID+".json"), nil
		}
		if err := os.Mkdir(dir, 0o700); err != nil && !errors.Is(err, fs.ErrExist) {
			return "", err
		}
	}
	resolvedDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(root, resolvedDir)
	if err != nil || rel == ".." || strings.HasPrefix(filepath.ToSlash(rel), "../") {
		return "", errors.New("auto mission: state_path_escape")
	}
	path := filepath.Join(resolvedDir, sessionID+".json")
	if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return "", errors.New("auto mission: unsafe_state_file")
	} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	return path, nil
}

func SaveAutoMission(projectRoot string, mission AutoMission) error {
	if strings.TrimSpace(mission.Text) == "" || mission.MissionMode != ModeAuto {
		return errors.New("auto mission: invalid_state")
	}
	path, err := secureMissionPath(projectRoot, mission.SessionID, true)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.Chmod(dir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(mission, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".mission-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(name)
		}
	}()
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := atomicfile.Replace(name, path); err != nil {
		return err
	}
	cleanup = false
	return os.Chmod(path, 0o600)
}

func LoadAutoMission(projectRoot, sessionID string) (*AutoMission, error) {
	path, err := secureMissionPath(projectRoot, sessionID, false)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var mission AutoMission
	if err := json.Unmarshal(data, &mission); err != nil {
		return nil, fmt.Errorf("auto mission parse: %w", err)
	}
	if err := ValidateAutoMissionIntegrity(&mission); err != nil {
		return &mission, err
	}
	return &mission, nil
}

func ClearAutoMission(projectRoot, sessionID string) error {
	path, err := secureMissionPath(projectRoot, sessionID, false)
	if err != nil {
		return err
	}
	if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return errors.New("auto mission: unsafe_state_file")
	}
	err = os.Remove(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}
