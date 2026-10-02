// session_start_record.go writes the factory state record for THIS session.
//
// The write lives here, and not in the launcher, because the launcher runs
// BEFORE the session it launches exists: the identifier it would key on
// belongs to a process that has not started, so it resolved one from the
// project-wide .moai/state/current-session-id.txt slot instead — a single slot
// the last SessionStart wins. The key was therefore not a function of the
// session the record described. Sometimes it coincided with the launched
// session, which is the worse case rather than the reassuring one: a consumer
// that found a record could not tell whether it described the session it asked
// about.
//
// SessionStart is the first actor holding the described session's own
// identifier, so keying on input.SessionID makes the record correct BY
// CONSTRUCTION rather than by a resolution step that could be got wrong again
// (SPEC-KANBAN-RECORD-SESSION-KEY-001 REQ-KRS-001/002/003).
//
// Fail-open throughout, matching the surrounding hook code and the guarantee
// the launcher path already carried: a record is an aid to the chain, never a
// precondition for starting one. WriteBestEffort discards every failure by
// design, and nothing here returns an error — a hook that failed loudly on an
// unwritable state directory would turn a silent degradation into a session
// that cannot boot (REQ-KRS-008).
package hook

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"gopkg.in/yaml.v3"
)

// worktreeRootDirName is the directory a card worktree's root must sit
// directly beneath for its basename to be a card identifier. The dispatch
// protocol fixes a card's worktree at <...>/worktrees/<card-id>, so the
// containment test is what separates a card id from a checkout name.
const worktreeRootDirName = "worktrees"

// writeFactorySessionRecord records this session's factory identity under its
// OWN identifier, or does nothing at all.
//
// Four gates precede the write, each of them a case where the correct answer
// is no record rather than a guessed one:
//
//   - No runtime session identifier, or no resolvable project root — there is
//     nothing to key on.
//   - A re-entry source (resume / clear / compact / fork). The record's three
//     orchestrator-written fields land independently as the chain progresses,
//     so rewriting on every SessionStart would discard them. An empty source
//     is treated as startup, matching the bootstrap notices.
//   - A session that is not a factory session. No launch
//     facts, no record — unchanged from today, where no launcher ran.
//   - A record already present for this identifier, for the same reason the
//     re-entry gate exists: the write is additive to the session's start, not
//     to its progress.
func writeFactorySessionRecord(input *HookInput) {
	if input == nil || input.SessionID == "" {
		return
	}
	if input.Source != "" && input.Source != "startup" {
		return
	}

	root := input.ProjectDir
	if root == "" {
		root = input.CWD
	}
	if root == "" {
		return
	}
	// Same guard as the write-side resolver (resolveProjectRoot): the record
	// write creates <root>/.moai/state/..., so a root that is not already a
	// MoAI root — CWD standing in for an absent ProjectDir from a subdirectory
	// — would grow a stray tree (card t1165).
	if _, err := os.Stat(filepath.Join(root, ".moai")); err != nil {
		return
	}

	// REQ-RNC-025: a session whose launch label carries a legacy role value is
	// never adopted into the new vocabulary — no record, and the stale-run
	// notice (emitted by the notice paths) tells the operator to relaunch.
	if legacy := legacyLaunchLabelValue(); legacy != "" {
		return
	}

	role, lane, ok := factoryRoleFromEnv()
	if !ok {
		return
	}

	if _, err := os.Stat(factory.RecordPath(root, input.SessionID)); err == nil {
		return
	}

	dir := input.CWD
	if dir == "" {
		dir = input.ProjectDir
	}

	backend := os.Getenv(config.EnvFactoryBackend)
	if isGatewaySession() {
		backend = os.Getenv(config.EnvMoaiLaunchProvider)
	}
	rec := factory.NewRecord(
		input.SessionID,
		"",
		backend,
	).WithRole(role).WithLane(lane).WithCard(resolveSessionCardID(dir))

	factory.WriteBestEffort(root, rec)
	pruneSessionRecordsBestEffort(root)
}

// pruneSessionRecordsBestEffort sweeps session records older than the
// configured retention window (card t1312) at the one point every factory
// session already touches this directory: its own record write. Fail-open in
// both directions — the sweep never gates the launch (the WriteBestEffort
// precedent), and an UNREADABLE config never enables a sweep either: deleting
// data on a guessed default is the one direction this path must not fail
// open in, so a config read or parse error disables the sweep for that
// launch instead of falling back to the shipped window.
func pruneSessionRecordsBestEffort(root string) {
	days, ok := sessionRecordRetentionDays(root)
	if !ok {
		return
	}
	_, _ = factory.PruneExpiredRecords(root, days, time.Now())
}

// sessionRecordRetentionDays reads state.session_record_retention_days from
// the project's state.yaml with a targeted unmarshal — the clean.go
// loadRetentionDays precedent, kept off the full config load so the hook's 5s
// budget carries one small file read. An ABSENT file is the shipped-default
// case (default-on out of the box). ok is false when a PRESENT file cannot be
// read or parsed — the caller treats that as "no sweep this launch", because
// deleting data on a guessed default is the one direction this path must not
// fail open in. An omitted key retains the config package's shipped default;
// an explicit 0 disables.
func sessionRecordRetentionDays(root string) (days int, ok bool) {
	data, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "state.yaml")) // #nosec G304 -- root is the hook's resolved project dir; the join is the fixed section path
	if err != nil {
		if os.IsNotExist(err) {
			return config.DefaultSessionRecordRetentionDays, true
		}
		return 0, false
	}
	var wrapper struct {
		State struct {
			SessionRecordRetentionDays *int `yaml:"session_record_retention_days"`
		} `yaml:"state"`
	}
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return 0, false
	}
	if wrapper.State.SessionRecordRetentionDays == nil {
		return config.DefaultSessionRecordRetentionDays, true
	}
	return *wrapper.State.SessionRecordRetentionDays, true
}

// factoryRoleFromEnv reports the role this session occupies in a factory run
// and, for a factory lane, its number. ok is false when the session is not part
// of a factory run at all, and when a label is present but does not parse — a
// malformed label yields no record rather than a guessed role. A LEGACY label
// also yields no record: the caller's legacy gate (above) stops it earlier, and
// this reader maps no legacy spelling to a role (SPEC-ROLE-NAMING-CODE-001
// REQ-RNC-009). The markers of the retired chain-session mode — the leader marker and
// the companion label a surviving session may still carry — map to no role.
//
// The discriminators and their ORDER mirror the factory bootstrap notice
// (session_start_factory.go): the lane label first, then the fan-out size.
//
// The lane number is returned as its own datum and is set through WithLane,
// never through WithRole — whose drop-unknown guard exists precisely so a
// consumer never has to defend against arbitrary launch-label text.
func factoryRoleFromEnv() (role string, lane int, ok bool) {
	if label := os.Getenv(config.EnvMoaiFactoryWorker); label != "" {
		n, parsed := factory.SplitFactoryLaneLabel(label)
		if !parsed {
			return "", 0, false
		}
		return factory.RoleLane, n, true
	}
	if os.Getenv(config.EnvMoaiFactoryWorkers) != "" {
		return factory.RoleLeader, 0, true
	}
	return "", 0, false
}

// resolveSessionCardID returns the queue card this session is working: the
// explicit override where one is supplied, and otherwise the card worktree
// this session stands in. An empty override is treated as unset so it never
// blanks a derivable value (REQ-KRS-005, acceptance.md §E).
func resolveSessionCardID(dir string) string {
	if override := strings.TrimSpace(os.Getenv(config.EnvFactoryCard)); override != "" {
		return override
	}
	return cardIDFromPath(dir)
}

// cardIDFromPath walks dir upward and returns the first ancestor whose PARENT
// directory is named "worktrees" — the card worktree this session stands in —
// or "" when there is none.
//
// The walk is what makes a session whose cwd sits DEEP inside a card worktree
// resolve the same card as one standing at its root, and it is deliberately
// pure path arithmetic: `git rev-parse --show-toplevel` names the same
// directory but costs a subprocess on a hook that runs under a 5s budget.
// Measured on this tree, resolving it that way pushed SessionStart's
// synchronous return from under 500ms to 650-890ms for every factory session.
//
// The containment test is what makes the "left empty rather than guessed"
// clause reachable. Without it the derivation always yields something inside
// any checkout: a session standing in a primary checkout named moai-adk-go
// would file "moai-adk-go" as its card identifier, and the console would
// render it as one. The test constrains where the value may come from, not
// whether the card exists — a card worktree whose directory name is not a real
// card id still records that name.
func cardIDFromPath(dir string) string {
	if strings.TrimSpace(dir) == "" {
		return ""
	}
	for dir = filepath.Clean(dir); ; {
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		if filepath.Base(parent) == worktreeRootDirName {
			base := filepath.Base(dir)
			if base == "." || base == string(filepath.Separator) {
				return ""
			}
			return base
		}
		dir = parent
	}
}
