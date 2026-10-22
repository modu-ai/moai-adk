package graph

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/privatefile"
)

func BuildPrivateGTDProjectionFromStore(ctx context.Context, store *factory.BacklogStore) (PrivateGTDProjection, error) {
	revision, relations, err := factory.GTDProjectionSource(ctx, store)
	if err != nil {
		return PrivateGTDProjection{}, err
	}
	edges := make([]PrivateGTDEdge, 0, len(relations))
	for _, r := range relations {
		edges = append(edges, PrivateGTDEdge{From: r.From, To: r.To, Kind: r.Kind, Sensitivity: r.Sensitivity})
	}
	return BuildPrivateGTDProjection(ctx, store.EnginePath(), revision, edges)
}

const (
	privateGTDDataName = "gtd-edges.jsonl"
	privateGTDMetaName = "gtd-edges.meta.json"
)

type PrivateGTDEdge struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Kind        string `json:"kind"`
	Sensitivity string `json:"sensitivity"`
}

type PrivateGTDProjection struct {
	DataPath       string
	MetaPath       string
	SourceRevision int64
	Generation     string
	EdgeCount      int
}

// PrivateGTDPermissionChecker makes account identity and kernel ownership an
// explicit, testable authorization boundary. Implementations must fail closed
// when an OS cannot report a stable owner identity. UID methods return the
// native account identity: a Unix UID or Windows SID. Mode is an observation;
// CheckPrivatePolicy is the mandatory native authorization oracle.
type PrivateGTDPermissionChecker interface {
	CurrentUID() (string, error)
	OwnerUID(path string) (string, error)
	Mode(path string) (os.FileMode, error)
	CheckPrivatePolicy(path string, expected os.FileMode) error
}

type systemPrivateGTDPermissionChecker struct{}

func (systemPrivateGTDPermissionChecker) CurrentUID() (string, error) {
	return privateGTDCurrentIdentity()
}

func (systemPrivateGTDPermissionChecker) OwnerUID(path string) (string, error) {
	return privateGTDOwnerIdentity(path)
}

func (systemPrivateGTDPermissionChecker) CheckPrivatePolicy(path string, expected os.FileMode) error {
	switch expected {
	case 0o700:
		return privatefile.CheckDirectory(path)
	case 0o600:
		return privatefile.Check(path)
	default:
		return errors.New("gtd projection: unsupported_private_policy")
	}
}

func (systemPrivateGTDPermissionChecker) Mode(path string) (os.FileMode, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Mode().Perm(), nil
}

func SystemPrivateGTDPermissionChecker() PrivateGTDPermissionChecker {
	return systemPrivateGTDPermissionChecker{}
}

// AuthorizePrivateGTDProjection permits only the current account, and only
// while the projection directory and both artifacts retain their private
// kernel owner and native private access-policy invariants. Windows approval
// requires the protected owner-only DACL, never its synthetic POSIX mode bits.
func AuthorizePrivateGTDProjection(projection PrivateGTDProjection, requesterUID string, checker PrivateGTDPermissionChecker) error {
	if checker == nil || requesterUID == "" {
		return errors.New("gtd projection: identity_required")
	}
	current, err := checker.CurrentUID()
	if err != nil || current == "" || requesterUID != current {
		return errors.New("gtd projection: account_denied")
	}
	for _, item := range []struct {
		path string
		mode os.FileMode
	}{{filepath.Dir(projection.DataPath), 0o700}, {projection.DataPath, 0o600}, {projection.MetaPath, 0o600}} {
		owner, ownerErr := checker.OwnerUID(item.path)
		policyErr := checker.CheckPrivatePolicy(item.path, item.mode)
		if ownerErr != nil || policyErr != nil || owner != current {
			return errors.New("gtd projection: permission_denied")
		}
	}
	return nil
}

type privateGTDMeta struct {
	Version        int    `json:"version"`
	SourceRevision int64  `json:"source_revision"`
	Generation     string `json:"generation"`
	DataSHA256     string `json:"data_sha256"`
	EdgeCount      int    `json:"edge_count"`
}

func validProjectionSensitivity(value string) bool {
	return value == "public" || value == "private" || value == "secret"
}

func normalizePrivateGTDEdges(edges []PrivateGTDEdge) ([]PrivateGTDEdge, error) {
	copyEdges := append([]PrivateGTDEdge(nil), edges...)
	for _, edge := range copyEdges {
		if edge.From == "" || edge.To == "" || edge.Kind == "" || !validProjectionSensitivity(edge.Sensitivity) {
			return nil, errors.New("gtd projection: privacy_unclassified_or_invalid_edge")
		}
	}
	sort.Slice(copyEdges, func(i, j int) bool {
		a, b := copyEdges[i], copyEdges[j]
		if a.From != b.From {
			return a.From < b.From
		}
		if a.To != b.To {
			return a.To < b.To
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return a.Sensitivity < b.Sensitivity
	})
	out := copyEdges[:0]
	for _, edge := range copyEdges {
		if len(out) > 0 && out[len(out)-1] == edge {
			continue
		}
		out = append(out, edge)
	}
	return out, nil
}

func renderPrivateGTDEdges(edges []PrivateGTDEdge) ([]byte, error) {
	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	for _, edge := range edges {
		if err := enc.Encode(edge); err != nil {
			return nil, err
		}
	}
	return out.Bytes(), nil
}

func writePrivateTemp(dir, prefix string, data []byte) (string, error) {
	f, err := privatefile.CreateTemp(dir, prefix)
	if err != nil {
		return "", err
	}
	name := f.Name()
	ok := false
	defer func() {
		_ = f.Close()
		if !ok {
			_ = os.Remove(name)
		}
	}()
	if _, err := f.Write(data); err != nil {
		return "", err
	}
	if err := f.Sync(); err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	if err := privatefile.Check(name); err != nil {
		return "", err
	}
	ok = true
	return name, nil
}

// BuildPrivateGTDProjection publishes a deterministic, generation-paired
// projection beside backlog.db. Every input is classified before the output
// directory or a temporary file is created, so ambiguous privacy cannot leave
// a partial artifact.
func BuildPrivateGTDProjection(ctx context.Context, dbPath string, sourceRevision int64, edges []PrivateGTDEdge) (PrivateGTDProjection, error) {
	if err := ctx.Err(); err != nil {
		return PrivateGTDProjection{}, err
	}
	if !filepath.IsAbs(dbPath) || filepath.Base(dbPath) != "backlog.db" || sourceRevision < 0 {
		return PrivateGTDProjection{}, errors.New("gtd projection: invalid_source")
	}
	if info, err := os.Lstat(filepath.Dir(dbPath)); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return PrivateGTDProjection{}, errors.New("gtd projection: symlink_source_directory")
	}
	normalized, data, metaBytes, generation, err := preparePrivateGTDProjection(sourceRevision, edges)
	if err != nil {
		return PrivateGTDProjection{}, err
	}
	dataPath, metaPath, err := publishPrivateGTDProjection(filepath.Dir(dbPath), data, metaBytes)
	if err != nil {
		return PrivateGTDProjection{}, err
	}
	return PrivateGTDProjection{DataPath: dataPath, MetaPath: metaPath, SourceRevision: sourceRevision, Generation: generation, EdgeCount: len(normalized)}, nil
}

func preparePrivateGTDProjection(sourceRevision int64, edges []PrivateGTDEdge) ([]PrivateGTDEdge, []byte, []byte, string, error) {
	normalized, err := normalizePrivateGTDEdges(edges)
	if err != nil {
		return nil, nil, nil, "", err
	}
	data, err := renderPrivateGTDEdges(normalized)
	if err != nil {
		return nil, nil, nil, "", err
	}
	dataSum := sha256.Sum256(data)
	dataHash := hex.EncodeToString(dataSum[:])
	genSum := sha256.Sum256([]byte(strconv.FormatInt(sourceRevision, 10) + "\x00" + dataHash))
	generation := hex.EncodeToString(genSum[:])
	meta := privateGTDMeta{Version: 1, SourceRevision: sourceRevision, Generation: generation, DataSHA256: dataHash, EdgeCount: len(normalized)}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return nil, nil, nil, "", err
	}
	metaBytes = append(metaBytes, '\n')
	return normalized, data, metaBytes, generation, nil
}

type privateGTDPublisher interface {
	PrepareDir(string) error
	WriteTemp(string, string, []byte) (string, error)
	Install(string, string) error
	Remove(string)
}

type osPrivateGTDPublisher struct{}

func (osPrivateGTDPublisher) PrepareDir(dir string) error {
	if err := os.MkdirAll(filepath.Dir(dir), 0o700); err != nil {
		return err
	}
	if err := privatefile.Mkdir(dir); err != nil {
		if !os.IsExist(err) {
			return err
		}
		current, identityErr := privateGTDCurrentIdentity()
		owner, ownerErr := privateGTDOwnerIdentity(dir)
		if identityErr != nil || ownerErr != nil || current == "" || owner != current {
			return errors.New("gtd projection: directory_owner_denied")
		}
		return privatefile.SecureDirectory(dir)
	}
	return nil
}
func (osPrivateGTDPublisher) WriteTemp(dir, prefix string, data []byte) (string, error) {
	return writePrivateTemp(dir, prefix, data)
}
func (osPrivateGTDPublisher) Install(source, target string) error {
	// A same-directory move retains the private staging object. In particular,
	// Windows MoveFileEx must not copy a payload into an existing broad DACL.
	if filepath.Clean(filepath.Dir(source)) != filepath.Clean(filepath.Dir(target)) {
		return errors.New("gtd projection: staging_directory_mismatch")
	}
	before, err := os.Lstat(source)
	if err != nil {
		return err
	}
	if err := privatefile.Check(source); err != nil {
		return err
	}
	if err := atomicfile.Replace(source, target); err != nil {
		return err
	}
	after, err := os.Lstat(target)
	if err == nil && !os.SameFile(before, after) {
		err = errors.New("gtd projection: publication_identity_changed")
	}
	if err == nil {
		err = privatefile.Check(target)
	}
	if err != nil {
		_ = os.Remove(target)
		return err
	}
	return nil
}
func (osPrivateGTDPublisher) Remove(path string) { _ = os.Remove(path) }

func publishPrivateGTDProjection(dir string, data, metaBytes []byte) (string, string, error) {
	return publishPrivateGTDProjectionWith(osPrivateGTDPublisher{}, dir, data, metaBytes)
}

func publishPrivateGTDProjectionWith(publisher privateGTDPublisher, dir string, data, metaBytes []byte) (string, string, error) {
	if err := publisher.PrepareDir(dir); err != nil {
		return "", "", err
	}
	dataTmp, err := publisher.WriteTemp(dir, ".gtd-edges-", data)
	if err != nil {
		return "", "", err
	}
	defer publisher.Remove(dataTmp)
	metaTmp, err := publisher.WriteTemp(dir, ".gtd-meta-", metaBytes)
	if err != nil {
		return "", "", err
	}
	defer publisher.Remove(metaTmp)
	dataPath := filepath.Join(dir, privateGTDDataName)
	metaPath := filepath.Join(dir, privateGTDMetaName)
	if err := publisher.Install(dataTmp, dataPath); err != nil {
		return "", "", err
	}
	if err := publisher.Install(metaTmp, metaPath); err != nil {
		publisher.Remove(dataPath)
		return "", "", err
	}
	return dataPath, metaPath, nil
}

func PrivateGTDProjectionFresh(metaPath string, sourceRevision int64) bool {
	metaBytes, err := os.ReadFile(metaPath)
	if err != nil {
		return false
	}
	var meta privateGTDMeta
	if json.Unmarshal(metaBytes, &meta) != nil || meta.Version != 1 || meta.SourceRevision != sourceRevision {
		return false
	}
	dataPath := filepath.Join(filepath.Dir(metaPath), privateGTDDataName)
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return false
	}
	sum := sha256.Sum256(data)
	hash := hex.EncodeToString(sum[:])
	gen := sha256.Sum256([]byte(strconv.FormatInt(sourceRevision, 10) + "\x00" + hash))
	return hash == meta.DataSHA256 && hex.EncodeToString(gen[:]) == meta.Generation
}

func PrivateGTDBackupFiles(projection PrivateGTDProjection, explicitOptIn bool) []string {
	if !explicitOptIn {
		return nil
	}
	return []string{projection.DataPath, projection.MetaPath}
}

func RevokePrivateGTDProjection(projection PrivateGTDProjection) error {
	var errs []error
	for _, path := range []string{projection.DataPath, projection.MetaPath} {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("remove %s: %w", path, err))
		}
	}
	return errors.Join(errs...)
}
