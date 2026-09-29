// Package discovery implements the lane-join leader discovery for factory
// runs (SPEC-FACTORY-LANE-JOIN-SOCKET-001): when a lane join's run resolution
// fails with NO_ACTIVE_FACTORY, the launcher probes for a live leader session
// of this project before refusing (REQ-001), and — through the caller's
// resume writer — restores that leader's run record so the join re-enters the
// ordinary gate (REQ-004).
//
// The proof standard is the retirement machinery's own (REQ-002): a candidate
// is verified live only by pid liveness TOGETHER with a process-start
// fingerprint match, exactly the ClassifyOwnerWith predicate family that
// retires run rows. Every other surface — a socket file in the runtime socket
// directory, a broker peer row, a registry row — feeds the candidate LIST and
// never the verdict (AC-005).
//
// Discovery never dials the runtime socket: the directory is a candidate
// enumeration source only. The lane↔lead exchange rides the run-scoped broker
// as it always has.
//
// Discovery is read-then-write: everything in this package is read-only; the
// single write of the path (the resume) lives in homestate.ResumeRun and is
// performed by the caller.
package discovery

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// runtimeSocketDir is the Claude Code runtime's per-pid socket directory. Its
// entries are CANDIDACY ONLY (REQ-002): many are stale, existence proves
// nothing, and the directory's protocol is an undetermined runtime question
// this package deliberately never depends on (spec.md §A).
const runtimeSocketDir = "/tmp/cc-socks"

// discoveryDeadline bounds one DiscoverLeader pass. The live machine measured
// ~97 candidate sockets, most of them dead pids that cost one cheap
// liveness probe each; only live candidates pay the argv/env/cwd readers.
const discoveryDeadline = 15 * time.Second

// VerifiedLeader is a candidate that passed the full proof (REQ-002 + REQ-003):
// live by pid + process-start fingerprint, carrying the targeted leader label
// in its argv, a member of this canonical project, and readable run identity.
type VerifiedLeader struct {
	// RunID is the candidate's OWN MOAI_KANBAN_ID — the address of the broker
	// the lead already speaks on. It is never guessed, defaulted, or minted.
	RunID string
	// PID and ProcessStart are the probe-measured owner identity the resume
	// stamps (retire-grade correct at write time).
	PID          int
	ProcessStart string
	// Name is the session name the candidate's argv carries (== the targeted
	// label); the caller exports it as MOAI_KANBAN_LEAD_NAME.
	Name string
	// Basis records the probe facts that verified the candidate; it rides the
	// resume event's payload for audit.
	Basis string
}

// The seams. Each is a package variable so tests inject fakes and the
// classifier matrix (AC-004/005/011/017) runs on any host with zero platform
// dependence — the RUN-RETIRE REQ-013 seam pattern. The default implementations
// are the platform readers behind this build-tag-free interface.
var (
	// candidatePIDs enumerates the pid candidate list (candidacy, not liveness).
	candidatePIDs = defaultCandidatePIDs
	// probeIdentity is the liveness + process-start fingerprint probe.
	probeIdentity = homestate.ProbeProcessIdentity
	// processArgv / processEnv / processCwd read a live process per platform.
	// Each declines (false) where the platform or privilege cannot answer —
	// a decline is an honest NO_ACTIVE_FACTORY, never a wrong join.
	processArgv = platformArgv
	processEnv  = platformEnv
	processCwd  = platformCwd
	// socketScanDir is the runtime socket directory candidates enumerate
	// from; a var so tests point it at a fixture directory.
	socketScanDir = runtimeSocketDir
)

// runIDPattern is the run-id shape the run resolution and the broker validate
// (factorymsg.ValidateActiveRun's rule); a candidate env value must parse to
// it before it can name a broker address.
var runIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$`)

// DiscoverLeader returns every live leader of projectRoot carrying targetLabel,
// verified under the retirement-grade proof. The caller decides: zero → the
// original refusal stands; one → resume; two or more → fail closed (REQ-005)
// — selection among live leaders is the same prohibition as selection among
// active records.
//
// The pass is read-only and deadline-bounded (discoveryDeadline).
//
// @MX:ANCHOR: [AUTO] DiscoverLeader — the verified-leader discovery seam at the shared lane-join gate (REQ-001/REQ-010)
// @MX:REASON: the public boundary every lane join (cc/glm/codex twin) reaches through one shared join point; a launcher-private copy of this path is the AC-013 defect
func DiscoverLeader(ctx context.Context, projectRoot, targetLabel string) ([]VerifiedLeader, error) {
	ctx, cancel := context.WithTimeout(ctx, discoveryDeadline)
	defer cancel()

	pids, err := candidatePIDs(ctx, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("enumerate leader candidates: %w", err)
	}
	canonical := homestate.CanonicalProjectRoot(projectRoot)
	knownRuns := projectRuns(ctx, projectRoot)

	var verified []VerifiedLeader
	// A pid that surfaced from two enumeration sources is ONE candidate —
	// counting it twice would masquerade as the multi-leader ambiguity.
	seen := map[int]bool{}
	for _, pid := range pids {
		if ctx.Err() != nil {
			break
		}
		if seen[pid] {
			continue
		}
		seen[pid] = true
		if v, ok := verifyCandidate(ctx, pid, canonical, targetLabel, knownRuns); ok {
			verified = append(verified, v)
		}
	}
	return verified, ctx.Err()
}

// verifyCandidate applies the four independently decline-able facts of the
// proof (design.md §B): liveness, leader-label match, project membership, and
// readable run identity. Every decline is silent by design — a declined
// candidate contributes nothing, and the caller's zero-leader refusal is the
// honest answer.
//
// @MX:WARN: [AUTO] liveness verdict joins pid + process-start fingerprint; weakening either fact resumes against a dead or recycled process
// @MX:REASON: the inverse defect (joining a dead lead) is the same fault this SPEC repairs, in the other direction (REQ-002/REQ-007); AC-005's mutant exists for exactly this check
func verifyCandidate(ctx context.Context, pid int, canonicalRoot, targetLabel string, knownRuns map[string]int) (VerifiedLeader, bool) {
	// Fact 1 — liveness (REQ-002): live pid AND a non-empty process-start
	// fingerprint, the ClassifyOwnerWith predicate family. Record presence
	// never reaches here.
	fingerprint, state := probeIdentity(pid)
	if state != homestate.ProcessIdentityLive || strings.TrimSpace(fingerprint) == "" {
		return VerifiedLeader{}, false
	}

	// Fact 2 — leader-label match (REQ-008): the candidate's argv names the
	// targeted leader label. A lane-<n> match is not a leader match, and a
	// name that has no leader shape maps to no leader at all.
	argv, ok := processArgv(ctx, pid)
	if !ok {
		return VerifiedLeader{}, false
	}
	name := namedFlagValue(argv)
	if name != targetLabel {
		return VerifiedLeader{}, false
	}
	if _, leaderShaped := kanban.SplitLeaderLabel(name); !leaderShaped {
		return VerifiedLeader{}, false
	}

	// Fact 4 — run identity (REQ-003): the candidate's own MOAI_KANBAN_ID,
	// read from its process environment. Unreadable env declines (AC-017
	// leg b); so does an env value that cannot name a broker address.
	env, ok := processEnv(ctx, pid)
	if !ok {
		return VerifiedLeader{}, false
	}
	runID := strings.TrimSpace(env[config.EnvMoaiKanbanID])
	if runID == "" || !runIDPattern.MatchString(runID) {
		return VerifiedLeader{}, false
	}

	// Fact 3 — project membership (REQ-003): the candidate's cwd must
	// canonicalize to this project's canonical root. Where the platform
	// cannot read the cwd, the candidate's run id appearing in THIS
	// project's run records (any status) is the membership evidence; with
	// neither, the candidate is declined. A READABLE cwd that canonicalizes
	// elsewhere is a hard decline (AC-017 leg a) — membership never
	// overrides contradicting cwd evidence.
	membership := "run-record"
	cwd, ok := processCwd(ctx, pid)
	if ok {
		if homestate.CanonicalProjectRoot(cwd) != canonicalRoot {
			return VerifiedLeader{}, false
		}
		membership = "cwd"
	} else if _, recorded := knownRuns[runID]; !recorded {
		return VerifiedLeader{}, false
	}

	basis := fmt.Sprintf(`{"liveness":"pid+fingerprint","label":%q,"membership":%q,"run_id_source":"env"}`, name, membership)
	return VerifiedLeader{RunID: runID, PID: pid, ProcessStart: fingerprint, Name: name, Basis: basis}, true
}

// namedFlagValue returns the value of the first `--name` / `-n` flag in argv,
// in the four forms parseNamedLabel accepts (flag space value, flag=value).
func namedFlagValue(argv []string) string {
	for i := 0; i < len(argv); i++ {
		arg := argv[i]
		switch {
		case arg == "--name" || arg == "-n":
			if i+1 < len(argv) {
				return argv[i+1]
			}
		case strings.HasPrefix(arg, "--name="):
			return strings.TrimPrefix(arg, "--name=")
		case strings.HasPrefix(arg, "-n="):
			return strings.TrimPrefix(arg, "-n=")
		}
	}
	return ""
}

// defaultCandidatePIDs composes the candidate list from the three enumeration
// sources (design.md §B, in order): pid-named sockets in the runtime socket
// directory, lead pids of the project's run rows in ANY status, and broker
// peer pids of those runs. Deduplicated; sources that cannot be read are
// skipped — enumeration is candidacy only, and a skipped source can only miss
// a candidate, never invent one.
func defaultCandidatePIDs(ctx context.Context, projectRoot string) ([]int, error) {
	seen := map[int]bool{}
	var out []int
	add := func(pid int) {
		if pid >= 1 && !seen[pid] {
			seen[pid] = true
			out = append(out, pid)
		}
	}

	for _, name := range socketCandidateNames(socketScanDir) {
		if pid, err := strconv.Atoi(strings.TrimSuffix(name, ".sock")); err == nil {
			add(pid)
		}
	}

	runLeadPIDs := projectRuns(ctx, projectRoot)
	for runID, leadPID := range runLeadPIDs {
		if ctx.Err() != nil {
			return out, nil
		}
		add(leadPID)
		for _, pid := range brokerPeerPIDs(ctx, projectRoot, runID) {
			add(pid)
		}
	}
	return out, nil
}

// socketCandidateNames lists *.sock entries of dir; a missing directory is a
// normal non-candidate (the runtime may not be installed or may use another
// scheme), not an error.
func socketCandidateNames(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sock") {
			names = append(names, e.Name())
		}
	}
	return names
}

// projectRuns returns the run rows this project's factory state records, in
// any status, mapped to their lead pid (a retired row's lead pid is as much a
// candidate as an active one — the probe decides). Read-only: a project with
// no factory state yet yields nothing and creates nothing, so discovery stays
// read-then-write (plan.md §D). The same map is the membership-evidence
// source for candidates whose cwd the platform cannot read.
func projectRuns(ctx context.Context, projectRoot string) map[string]int {
	out := map[string]int{}
	db, ok := openExistingFactory(projectRoot)
	if !ok {
		return out
	}
	defer func() { _ = db.Close() }()
	rows, err := db.DB.QueryContext(ctx, `SELECT run_id,lead_pid FROM runs`)
	if err != nil {
		return out
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var id string
		var pid int
		if err := rows.Scan(&id, &pid); err == nil {
			out[id] = pid
		}
	}
	return out
}

// brokerPeerPIDs returns the distinct peer pids of one run's broker. The
// broker database is opened the way the established read paths open it; a
// broker that cannot be read contributes no candidates.
func brokerPeerPIDs(ctx context.Context, projectRoot, runID string) []int {
	path, err := factorymsg.BrokerPath(projectRoot, runID)
	if err != nil {
		return nil
	}
	db, err := openSQLiteReadOnly(path)
	if err != nil {
		return nil
	}
	defer func() { _ = db.Close() }()
	rows, err := db.QueryContext(ctx, `SELECT DISTINCT pid FROM peers`)
	if err != nil {
		return nil
	}
	defer func() { _ = rows.Close() }()
	var pids []int
	for rows.Next() {
		var pid int
		if err := rows.Scan(&pid); err == nil {
			pids = append(pids, pid)
		}
	}
	return pids
}

// openExistingFactory opens the project's factory database WITHOUT creating
// anything: a missing file reports not-ok. This is what keeps discovery
// read-then-write — homestate.OpenFactory would ensure (create) the layout.
func openExistingFactory(projectRoot string) (*homestate.FactoryDB, bool) {
	path, err := homestate.FactoryDBPath(projectRoot)
	if err != nil {
		return nil, false
	}
	if _, err := os.Stat(path); err != nil {
		return nil, false
	}
	db, err := homestate.OpenFactoryPath(path)
	if err != nil {
		return nil, false
	}
	return db, true
}

// openSQLiteReadOnly opens one sqlite database file read-only (broker peer
// enumeration). A WAL database whose recovery would need the write lock fails
// here and contributes no candidates — enumeration fail-open.
func openSQLiteReadOnly(path string) (*sql.DB, error) {
	v := url.Values{}
	v.Add("_pragma", "busy_timeout(100)")
	v.Add("mode", "ro")
	db, err := sql.Open("sqlite", (&url.URL{Scheme: "file", Path: filepath.ToSlash(path), RawQuery: v.Encode()}).String())
	if err != nil {
		return nil, err
	}
	return db, nil
}

// DescribeVerifiedLeaders renders verified candidates for the multi-leader
// fail-closed refusal (REQ-005): each candidate with its run id and pid, no
// selection among them.
func DescribeVerifiedLeaders(verified []VerifiedLeader) string {
	parts := make([]string, 0, len(verified))
	for _, v := range verified {
		parts = append(parts, fmt.Sprintf("%s (leader pid %d, %s)", v.RunID, v.PID, v.Name))
	}
	return strings.Join(parts, ", ")
}
