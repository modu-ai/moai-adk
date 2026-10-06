package factory

// factory_relaunch_cmd.go — the one shared definition of the
// `moai factory relaunch` command grammar (SPEC-FACTORY-STALE-RUN-HEAL-001
// REQ-SRH-002, REQ-SRH-016). The hook's notices PRINT these lines and the
// verb in internal/cli PARSES the same flag names, so a printed line is
// accepted by the verb verbatim by construction.

import (
	"sort"
	"strings"
)

// Provider tokens of the verb's --provider flag.
const (
	RelaunchProviderCC    = "cc"
	RelaunchProviderGLM   = "glm"
	RelaunchProviderCodex = "codex"
)

// Flag names of the verb (without the leading dashes). The verb registers
// exactly these and the builder prints exactly these.
const (
	RelaunchFlagProvider = "provider"
	RelaunchFlagLane     = "lane"
	RelaunchFlagRun      = "run"
	RelaunchFlagFromRun  = "from-run"
	RelaunchFlagDryRun   = "dry-run"
)

const (
	relaunchVerbLine = "moai factory relaunch"
	// relaunchCandidateCap bounds the per-candidate lines of an ambiguity
	// notice (DP10): beyond it a count replaces the remaining lines.
	relaunchCandidateCap = 3
)

// RelaunchProviderForBackend maps the launcher's backend stamp
// (the backend stamp) to the verb's provider token (REQ-SRH-016): glm ->
// glm, gpt -> codex, and any other or empty value -> cc.
func RelaunchProviderForBackend(backend string) string {
	switch strings.TrimSpace(backend) {
	case BackendGLM:
		return RelaunchProviderGLM
	case BackendGPT:
		return RelaunchProviderCodex
	default:
		return RelaunchProviderCC
	}
}

// RelaunchCommand names the arguments of one `moai factory relaunch` line.
// An empty field is omitted from the printed line. For provider codex Lane
// and Run are never printed: the Codex relaunch entry is the bare `-l`.
type RelaunchCommand struct {
	Provider string
	Lane     string
	Run      string
	FromRun  string
}

func (c RelaunchCommand) isCodex() bool { return c.Provider == RelaunchProviderCodex }

// Line renders the command line the notices print and the verb accepts:
// `moai factory relaunch --provider P [--lane S] [--run R] [--from-run F]`.
//
// @MX:NOTE: [AUTO] The single definition of the relaunch command grammar
// (SPEC-FACTORY-STALE-RUN-HEAL-001 REQ-SRH-002) — hook notices print it, the
// verb in internal/cli parses the same flag names; a grammar change is made
// here and nowhere else.
func (c RelaunchCommand) Line() string {
	var b strings.Builder
	b.WriteString(relaunchVerbLine + " --" + RelaunchFlagProvider + " " + c.Provider)
	if !c.isCodex() {
		if c.Lane != "" {
			b.WriteString(" --" + RelaunchFlagLane + " " + c.Lane)
		}
		if c.Run != "" {
			b.WriteString(" --" + RelaunchFlagRun + " " + c.Run)
		}
	}
	if c.FromRun != "" {
		b.WriteString(" --" + RelaunchFlagFromRun + " " + c.FromRun)
	}
	return b.String()
}

// LaunchArgs renders the launcher argv that follows `moai`: for cc and glm
// `<provider> -l [--factory-run <id>]`, for codex the only line its entry
// accepts, `codex -l`. The lane entry `-l` takes no argument and claims the
// next free slot, so a pinned Lane label does not travel into the launch
// (SPEC-LAUNCHER-ENTRY-FLAGS-001: no entry selects one lane any more).
func (c RelaunchCommand) LaunchArgs() []string {
	args := []string{c.Provider, "-l"}
	if !c.isCodex() && c.Run != "" {
		args = append(args, "--factory-run", c.Run)
	}
	return args
}

// LaunchLine renders the launch line `--dry-run` prints.
func (c RelaunchCommand) LaunchLine() string {
	return "moai " + strings.Join(c.LaunchArgs(), " ")
}

// RelaunchNoticeState is the measured input of the notice-line table
// (spec.md §D.7): the session's vocabulary, the run its environment names and
// that run's measured state, and the set of runs measuring active.
type RelaunchNoticeState struct {
	Provider   string
	Legacy     bool
	Lane       string // the session's own lane label (current vocabulary)
	Run        string // the run the launch environment names
	RunActive  bool   // that run measures active
	ActiveRuns []string
	// SlotHeldByLiveOther marks row R3: the sole active run's slot is held by
	// a different live session.
	SlotHeldByLiveOther bool
}

// RelaunchNoticeLines is the command-line part of a notice: each element is
// one complete command line, More counts candidate runs the cap left out.
type RelaunchNoticeLines struct {
	Lines []string
	More  int
}

// RelaunchNoticeFor is the notice-line table of spec.md §D.7 as a pure
// function of (vocabulary, run state, active runs). It prints only complete
// commands: no row ever carries an angle-bracket placeholder.
func RelaunchNoticeFor(s RelaunchNoticeState) RelaunchNoticeLines {
	p := s.Provider
	if s.Legacy {
		if s.RunActive { // R6
			return relaunchOne(RelaunchCommand{Provider: p, FromRun: s.Run})
		}
		switch len(s.ActiveRuns) {
		case 0: // R7
			return RelaunchNoticeLines{}
		case 1: // R8
			return relaunchOne(RelaunchCommand{Provider: p})
		default: // R9
			return relaunchCandidates(p, "", s.ActiveRuns)
		}
	}
	if s.RunActive { // R1
		return RelaunchNoticeLines{}
	}
	switch len(s.ActiveRuns) {
	case 0: // R4
		return RelaunchNoticeLines{}
	case 1: // R2, R3
		if s.SlotHeldByLiveOther {
			return relaunchOne(RelaunchCommand{Provider: p, Run: s.ActiveRuns[0]})
		}
		return RelaunchNoticeLines{}
	default: // R5
		return relaunchCandidates(p, s.Lane, s.ActiveRuns)
	}
}

func relaunchOne(c RelaunchCommand) RelaunchNoticeLines {
	return RelaunchNoticeLines{Lines: []string{c.Line()}}
}

// relaunchCandidates prints one line per candidate run (sorted, at most
// relaunchCandidateCap, the rest as a count). Provider codex has no run flag,
// so its one line stands for every candidate.
func relaunchCandidates(provider, lane string, runs []string) RelaunchNoticeLines {
	if provider == RelaunchProviderCodex {
		return relaunchOne(RelaunchCommand{Provider: provider})
	}
	sorted := append([]string(nil), runs...)
	sort.Strings(sorted)
	out := RelaunchNoticeLines{}
	for i, run := range sorted {
		if i == relaunchCandidateCap {
			out.More = len(sorted) - relaunchCandidateCap
			break
		}
		out.Lines = append(out.Lines, RelaunchCommand{Provider: provider, Lane: lane, Run: run}.Line())
	}
	return out
}
