package homestate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var (
	shaPattern          = regexp.MustCompile(`^[0-9a-f]{7,64}$`)
	cardIDPattern       = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)
	verdictLinePattern  = regexp.MustCompile(`^verdict: (PASS|PASS-WITH-DEBT|FAIL)$`)
	auditedSHAPattern   = regexp.MustCompile(`^audited_sha: ([0-9a-f]{12,40})$`)
	verdictIterPattern  = regexp.MustCompile(`-iter(\d+)`)
	maxEvidenceFileSize = int64(4 << 20)
)

// ValidCardID reports whether id is safe to use as a card identifier (and as
// a path segment under .moai/reports/).
func ValidCardID(id string) bool { return cardIDPattern.MatchString(id) && !strings.Contains(id, "..") }

func evidenceErr(format string, args ...any) error {
	return fmt.Errorf("%w: "+format, append([]any{ErrEvidence}, args...)...)
}

// resolveCommit verifies that sha names a commit in dir and returns its full
// object name.
func resolveCommit(ctx context.Context, dir, sha string) (string, error) {
	sha = strings.ToLower(strings.TrimSpace(sha))
	if !shaPattern.MatchString(sha) {
		return "", evidenceErr("%q is not a commit SHA", sha)
	}
	full, err := gitRead(ctx, dir, "rev-parse", "--verify", "--quiet", sha+"^{commit}")
	if err != nil || full == "" {
		return "", evidenceErr("commit %s does not resolve in %s", sha, dir)
	}
	return full, nil
}

func isAncestor(ctx context.Context, dir, ancestor, descendant string) bool {
	_, err := gitRead(ctx, dir, "merge-base", "--is-ancestor", ancestor, descendant)
	return err == nil
}

// verifyCommitAtHead is the T10 reader: the SHA resolves in the card's
// worktree and is an ancestor of, or equal to, that worktree's HEAD.
func verifyCommitAtHead(ctx context.Context, dir, sha string) (string, error) {
	full, err := resolveCommit(ctx, dir, sha)
	if err != nil {
		return "", err
	}
	if !isAncestor(ctx, dir, full, "HEAD") {
		return "", evidenceErr("commit %s is not an ancestor of the worktree HEAD", full)
	}
	return full, nil
}

// verifyAuditEntry is E-ENTRY (T5, T11): the SHA resolves, is an ancestor of
// or equal to the worktree HEAD, and contains the artifact at that commit.
func verifyAuditEntry(ctx context.Context, dir, sha, artifact string) (string, error) {
	artifact = strings.TrimSpace(artifact)
	if artifact == "" || strings.HasPrefix(artifact, "-") || strings.ContainsRune(artifact, 0) {
		return "", evidenceErr("artifact path %q is not usable", artifact)
	}
	full, err := verifyCommitAtHead(ctx, dir, sha)
	if err != nil {
		return "", err
	}
	if _, err := gitRead(ctx, dir, "cat-file", "-e", full+":"+filepath.ToSlash(artifact)); err != nil {
		return "", evidenceErr("artifact %s is absent at commit %s", artifact, full)
	}
	return full, nil
}

// auditVerdict is what E-VERDICT extracted from a verdict file.
type auditVerdict struct {
	Path, Verdict, AuditedSHA string
}

// readAuditVerdict is E-VERDICT (T6, T7, T12, T13): it reads the newest
// `<phase>*.md` under <worktree>/.moai/reports/<card-id>/ and extracts the
// machine-readable `verdict:` and `audited_sha:` lines, each of which must
// carry exactly one value. The audited SHA must name the commit recorded at
// audit entry (a prefix of at least 12 hex characters is accepted).
func readAuditVerdict(dir, cardID, phase, evidenceSHA string) (auditVerdict, error) {
	if !ValidCardID(cardID) {
		return auditVerdict{}, evidenceErr("card id %q is not a safe path segment", cardID)
	}
	if strings.TrimSpace(dir) == "" {
		return auditVerdict{}, evidenceErr("card has no worktree path")
	}
	pattern := filepath.Join(dir, ".moai", "reports", cardID, phase+"*.md")
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return auditVerdict{}, evidenceErr("no %s verdict file under %s", phase, filepath.Dir(pattern))
	}
	newest, err := newestVerdictFile(matches)
	if err != nil {
		return auditVerdict{}, err
	}
	raw, err := readBoundedFile(newest)
	if err != nil {
		return auditVerdict{}, evidenceErr("verdict file %s: %v", newest, err)
	}
	return parseAuditVerdict(newest, raw, evidenceSHA)
}

// ParseAuditVerdictFile applies the E-VERDICT line rules to one file's bytes
// against the recorded evidence commit. It is the reader the transition API
// uses, exposed so producers of verdict files can be checked against it.
func ParseAuditVerdictFile(path string, raw []byte, evidenceSHA string) (verdict, auditedSHA string, err error) {
	v, err := parseAuditVerdict(path, raw, evidenceSHA)
	return v.Verdict, v.AuditedSHA, err
}

func parseAuditVerdict(path string, raw []byte, evidenceSHA string) (auditVerdict, error) {
	verdicts, shas := map[string]bool{}, map[string]bool{}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimRight(line, " \t\r")
		if m := verdictLinePattern.FindStringSubmatch(line); m != nil {
			verdicts[m[1]] = true
		}
		if m := auditedSHAPattern.FindStringSubmatch(line); m != nil {
			shas[m[1]] = true
		}
	}
	out := auditVerdict{Path: path}
	switch len(verdicts) {
	case 0:
		return out, evidenceErr("verdict file %s carries no `verdict:` line", path)
	case 1:
		for v := range verdicts {
			out.Verdict = v
		}
	default:
		return out, evidenceErr("verdict file %s carries conflicting `verdict:` lines", path)
	}
	switch len(shas) {
	case 0:
		return out, evidenceErr("verdict file %s carries no `audited_sha:` line", path)
	case 1:
		for s := range shas {
			out.AuditedSHA = s
		}
	default:
		return out, evidenceErr("verdict file %s carries conflicting `audited_sha:` lines", path)
	}
	recorded := strings.ToLower(strings.TrimSpace(evidenceSHA))
	if recorded == "" || !strings.HasPrefix(recorded, out.AuditedSHA) {
		return out, evidenceErr("verdict file %s audited %s, not the recorded evidence commit %q", path, out.AuditedSHA, recorded)
	}
	return out, nil
}

func readBoundedFile(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > maxEvidenceFileSize {
		return nil, fmt.Errorf("not a regular file of at most %d bytes", maxEvidenceFileSize)
	}
	return os.ReadFile(path)
}

// newestVerdictFile picks the highest iteration number in the file name (a
// name without `-iter<N>` is iteration 1), then the latest modification time.
func newestVerdictFile(paths []string) (string, error) {
	type cand struct {
		path string
		iter int
		mod  int64
	}
	var best *cand
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		c := cand{path: p, iter: 1, mod: info.ModTime().UnixNano()}
		if m := verdictIterPattern.FindStringSubmatch(filepath.Base(p)); m != nil {
			if n, err := strconv.Atoi(m[1]); err == nil {
				c.iter = n
			}
		}
		if best == nil || c.iter > best.iter || (c.iter == best.iter && c.mod > best.mod) {
			cc := c
			best = &cc
		}
	}
	if best == nil {
		return "", evidenceErr("no readable verdict file among %v", paths)
	}
	return best.path, nil
}

func validBranchName(name string) bool {
	name = strings.TrimSpace(name)
	return name != "" && !strings.HasPrefix(name, "-") && !strings.Contains(name, "..") && !strings.ContainsAny(name, " \t\n~^:?*[\\")
}

// mergeEvidence is what E-MERGE verified.
type mergeEvidence struct {
	SHA, Tree, RemeasurePath string
}

// verifyMerge is E-MERGE (T16): a two-parent merge commit whose tree equals
// its second parent's tree, reachable from the local integration branch, and
// a re-measure file that names it.
func verifyMerge(ctx context.Context, dir, sha, remeasure, integration string) (mergeEvidence, error) {
	if !validBranchName(integration) {
		return mergeEvidence{}, evidenceErr("integration branch %q is not usable", integration)
	}
	full, err := resolveCommit(ctx, dir, sha)
	if err != nil {
		return mergeEvidence{}, err
	}
	parents, err := gitRead(ctx, dir, "rev-list", "--parents", "-n", "1", full)
	if err != nil {
		return mergeEvidence{}, evidenceErr("read parents of %s: %v", full, err)
	}
	if len(strings.Fields(parents)) != 3 {
		return mergeEvidence{}, evidenceErr("commit %s is not a two-parent merge", full)
	}
	tree, err := gitRead(ctx, dir, "rev-parse", full+"^{tree}")
	if err != nil {
		return mergeEvidence{}, evidenceErr("read tree of %s: %v", full, err)
	}
	second, err := gitRead(ctx, dir, "rev-parse", full+"^2^{tree}")
	if err != nil {
		return mergeEvidence{}, evidenceErr("read second-parent tree of %s: %v", full, err)
	}
	if tree != second {
		return mergeEvidence{}, evidenceErr("merge %s tree %s differs from its second parent's tree %s", full, tree, second)
	}
	if !isAncestor(ctx, dir, full, "refs/heads/"+integration) {
		return mergeEvidence{}, evidenceErr("merge %s is not reachable from %s", full, integration)
	}
	path := strings.TrimSpace(remeasure)
	if path == "" {
		return mergeEvidence{}, evidenceErr("a re-measure evidence file is required")
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(dir, path)
	}
	raw, err := readBoundedFile(path)
	if err != nil {
		return mergeEvidence{}, evidenceErr("re-measure file %s: %v", path, err)
	}
	if !strings.Contains(strings.ToLower(string(raw)), full[:12]) {
		return mergeEvidence{}, evidenceErr("re-measure file %s does not name merge %s", path, full)
	}
	return mergeEvidence{SHA: full, Tree: tree, RemeasurePath: path}, nil
}

// verifyPushed is the T17 reader: the merge commit is an ancestor of the
// remote-tracking ref of the integration branch as it stands. It never
// fetches; the lead fetches before deciding.
func verifyPushed(ctx context.Context, dir, mergeSHA, integration string) (string, error) {
	if !validBranchName(integration) {
		return "", evidenceErr("integration branch %q is not usable", integration)
	}
	remotes, err := gitRemotes(ctx, dir)
	if err != nil || len(remotes) == 0 {
		return "", evidenceErr("no remote is configured")
	}
	remote := remotes[0]
	if slices.Contains(remotes, "origin") {
		remote = "origin"
	}
	ref := "refs/remotes/" + remote + "/" + integration
	if _, err := gitRead(ctx, dir, "rev-parse", "--verify", "--quiet", ref); err != nil {
		return "", evidenceErr("remote-tracking ref %s does not exist", ref)
	}
	full, err := resolveCommit(ctx, dir, mergeSHA)
	if err != nil {
		return "", err
	}
	if !isAncestor(ctx, dir, full, ref) {
		return "", evidenceErr("merge %s is not contained in %s", full, ref)
	}
	return ref, nil
}
