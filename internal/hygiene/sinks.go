package hygiene

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// PassLockName is the rotator's pass-lock artifact under the logs directory
// (REQ-HYG-003, D34): the cross-process lockfile whose Windows LockFileEx
// sidecar form is this same file. Exclusion is advisory-lock-based, so an
// orphaned file after a crash is harmless.
const PassLockName = ".hygiene-rotation.lock"

// sinkRegistry is the closed, named list of append-only audit sinks the
// rotator owns (REQ-HYG-003). The hygiene audit sink self-registers
// (REQ-HYG-004); its own rotation is apply-mode-only like every other sink.
// The registry was extended at run phase (v0.4.1) with the eleven
// append-only writers the registry-completeness guard verified in the real
// tree — the plan-phase list carried only the eight measured-largest sinks
// and missed these (each entry's writer is named in the spec §G list).
var sinkRegistry = []string{
	"rule-load-audit.jsonl",
	"agent-model-audit.jsonl",
	"preedit-session-guard.log",
	"codex-adapter.jsonl",
	"status-transition-audit.log",
	"subagent-write-guard.log",
	"agent-stop-audit.jsonl",
	"hook-runtime.log",
	"hygiene-audit.jsonl",
	// v0.4.1 run-phase additions — verified append-only writers under
	// .moai/logs that the plan-phase registry missed:
	"anchor-relocation-audit.jsonl", // internal/session/anchor_relocate_audit.go
	"anchor-trace.jsonl",            // internal/session/anchor_trace.go
	"askuser-observations.jsonl",    // internal/hook/askuser_observer.go
	"autonomy-downgrade.log",        // internal/config AppendDowngradeAdvisory
	"hook-skip.log",                 // internal/hook/security/guardian.go
	"lifecycle-close.log",           // internal/spec/closer.go
	"migrations.log",                // internal/migration/log.go
	"navigator-sync.log",            // internal/navigator/fix/request.go
	"permission.log",                // internal/permission/conflict.go
	"slot-lease-audit.jsonl",        // internal/factory/slot_lease.go
	"task-metrics.jsonl",            // internal/hook/post_tool_metrics.go
}

// sinkSet is the lookup form of the registry.
var sinkSet = func() map[string]bool {
	m := make(map[string]bool, len(sinkRegistry))
	for _, name := range sinkRegistry {
		m[name] = true
	}
	return m
}()

// SinkRegistry returns a copy of the closed sink registry.
//
// @MX:ANCHOR: [AUTO] SinkRegistry — the closed-registry accessor
// @MX:REASON: the rotator's registry-completeness guard test and the CLI
// report both enumerate this list; a divergent copy would let the guard
// pass while the rotator rotates an unregistered name (REQ-HYG-003).
func SinkRegistry() []string {
	out := make([]string, len(sinkRegistry))
	copy(out, sinkRegistry)
	return out
}

// IsRegisteredSink reports whether name is in the closed sink registry.
func IsRegisteredSink(name string) bool {
	return sinkSet[name]
}

// chunkNames returns the rotator-owned chunk artifacts of a sink primary
// (REQ-HYG-002, REQ-HYG-003): the displaced chunk and its mid-placement
// staging form.
func chunkNames(primary string) (chunk, staging string) {
	return primary + ".1", primary + ".1.staging"
}

// ownerExceptions name sink-shaped files whose lifecycle is owned by
// another mechanism (SPEC-MOAI-HYGIENE-001 §D) and are therefore not
// required to be in the rotator's registry. The guard skips them; every
// entry carries its owner.
var ownerExceptions = map[string]string{
	"lessons-inbox.jsonl": "internal/hook/inbox_lifecycle.go + the moai inbox verbs",
	"usage-log.jsonl":     "internal/harness/retention.go (30-day prune to monthly gz)",
}

// sinkLiteral matches a quoted file-name literal that looks like an
// append-only sink name.
var sinkLiteral = regexp.MustCompile(`"[A-Za-z0-9][A-Za-z0-9._-]*\.(jsonl|log)"`)

// writeCallSignal matches the write-shape calls that distinguish a sink
// writer from a reader (a file that only reads the sink — a doctor check,
// for instance — is not an unregistered writer).
var writeCallSignal = regexp.MustCompile(`O_APPEND|WriteFile|os\.Create|OpenFile`)

// ScanSourceSinks scans a source tree for append-only sink writers under
// the logs directory that have no registry entry, returning the
// unregistered names sorted (REQ-HYG-003's registry-completeness guard).
// A source file counts as a logs writer when it references the "logs"
// directory, carries at least one sink-shaped file-name literal, and
// performs a write-shaped call (O_APPEND / WriteFile / Create / OpenFile) —
// a reader-only reference is not a writer. Test files are excluded: tests
// legitimately quote registered and fixture sink names in assertions; the
// guard hunts production writers.
//
// @MX:ANCHOR: [AUTO] ScanSourceSinks — the registry-completeness guard
// @MX:REASON: one new append-only sink under the logs directory without a
// registry entry would grow unbounded precisely because no rotator owns it;
// this scan is the mechanism that keeps the registry closed (REQ-HYG-003).
func ScanSourceSinks(sourceRoot string) ([]string, error) {
	var unregistered []string
	seen := map[string]bool{}
	err := filepath.WalkDir(sourceRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == "vendor" || name == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(content)
		if !strings.Contains(text, `"logs"`) {
			return nil
		}
		if !writeCallSignal.MatchString(text) {
			return nil
		}
		for _, lit := range sinkLiteral.FindAllString(text, -1) {
			name := strings.Trim(lit, `"`)
			if sinkSet[name] {
				continue
			}
			if _, owned := ownerExceptions[name]; owned {
				continue
			}
			if !seen[name] {
				seen[name] = true
				unregistered = append(unregistered, name)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(unregistered)
	return unregistered, nil
}

// appendText appends one line to a log file, creating the parent directory
// and the file when missing (O_APPEND — the sink appenders take no lock).
func appendText(path string, line string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(line); err != nil {
		return err
	}
	return nil
}

// fileExists reports whether path exists as a regular file (a dangling
// symlink or a directory reads as absent for rotation purposes).
func fileExists(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode().IsRegular()
}

// readLinesFile reads a text file line-by-line, tolerating a missing file.
func readLinesFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}
