package hook

// Recurrence guard for the factory/kanban lane-gate environment axes
// (SPEC-TEST-ENV-HERMETIC-001, REQ-THE-005 / REQ-THE-009).
//
// Hazard: the hook handlers decide on several lane gate axes at once
// (stale_run_gate.go's run-state gate answers on the run id AND the fan-out
// marker conjunction). A test binary running inside a factory lane session
// inherits those axes in its ambient environment, and a test that pins only
// some of them flips verdict with the env of the session that runs it — the
// fake red this SPEC measured on the two StaleRunNotice tests. TestMain
// clears the declared scrub set (laneEnvScrubAxes below) once for the whole
// binary; these two tests keep that set honest:
//
//   - TestLaneEnvAxesCovered: every family axis the package's production
//     code references must be either scrubbed at start-up or carry a
//     reasoned, cited exemption row. The family is re-read from
//     internal/config/envkeys.go at test time, so a new axis joins it
//     without editing this guard.
//   - TestLaneEnvAxesScrubApplied: the declared set proves a list, not that
//     the binary applies it. This test re-executes the package's own test
//     binary with every family axis set to a sentinel value and asserts, in
//     the child (whose TestMain has run), that every referenced non-exempt
//     axis is gone.
//
// Test-only file: no production symbol is introduced here (REQ-THE-008).
// The guard never spells the family axis names as quoted literals — the one
// literal envkeys_factory_role_test.go forbids in this package — it reads
// the values from envkeys.go and reaches the one full name it needs through
// the config constants.

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// laneAxisFamilyFloor is the liveness floor for the family census read from
// internal/config/envkeys.go (16 axes at the c2 tree, spec.md §A.6). Lower it
// in the same change that deliberately removes a family constant; a develop
// absorption that moves the census re-derives it at pre-flight instead.
const laneAxisFamilyFloor = 16

// laneAxesProbeWitnessEnv marks the re-executed child of the applied-behaviour
// test; its value names the axes the parent set.
const laneAxesProbeWitnessEnv = "MOAI_HOOK_TEST_LANE_AXES_PROBE"

// laneAxesProbeSentinel is the value the parent sets on every family axis
// before re-executing the child.
const laneAxesProbeSentinel = "t1356-axes-sentinel"

// laneAxesProbeChildTimeout bounds the re-executed child, in seconds. Sized
// from the measured child runtime under load (recorded in the SPEC's
// progress record); never below the 20s precedent bound.
const laneAxesProbeChildTimeout = 300

// laneEnvAxisExemption is one reasoned, cited excuse for leaving a
// production-referenced family axis outside the start-up scrub set.
type laneEnvAxisExemption struct {
	Axis     string // the axis value as read from envkeys.go
	Reason   string // why a test needs the ambient value (non-empty)
	Citation string // a *_test.go file in this package directory that references the axis
}

// laneEnvAxisExemptions starts empty. A row is added only when a whole-
// package scrubbed-arm measurement shows a test going red because the axis
// was stripped; the row then cites that test file and the measurement is
// recorded in the SPEC's progress record.
var laneEnvAxisExemptions []laneEnvAxisExemption

// laneEnvScrubAxes is the factory/kanban axis set TestMain strips at
// start-up. It is declared here so the coverage test reads the declared set
// from the same file that names the family rule; TestMain (main_test.go)
// applies it before the first test runs. Filled at M3 (SPEC-TEST-ENV-
// HERMETIC-001) with the ten axes production code references (spec.md
// §A.6): the c1 whole-package scrubbed arm ran all sixteen family axes
// unset and exited 0 with zero failing rows, so no test in this package
// relies on any ambient family value (spec.md §D precondition (i)); the
// single-axis arms re-observed at M3 confirmed MOAI_KANBAN_ID and
// MOAI_FACTORY_WORKERS alone flip the two StaleRunNotice tests. A row in
// laneEnvAxisExemptions below, added only after a whole-package scrubbed
// arm shows a test going red because the axis was stripped, is the escape
// hatch.
var laneEnvScrubAxes = []string{
	config.EnvFactoryRunID,
	config.EnvMoaiFactoryWorkers,
	config.EnvFactorySettingsInjected,
	config.EnvFactoryLeadAddr,
	config.EnvFactoryBackend,
	config.EnvFactoryCard,
	config.EnvFactoryLeadName,
	config.EnvFactoryRole,
	config.EnvMoaiFactoryWorker,
	config.EnvFactoryAutoDispatch,
}

// scrubLaneEnvAxes clears the declared lane axis set for this process.
// TestMain calls it before any test runs, so tests that never compose a lane
// env of their own cannot see the ambient lane state of the session that
// launched go test.
func scrubLaneEnvAxes() {
	for _, key := range laneEnvScrubAxes {
		_ = os.Unsetenv(key)
	}
}

// readLaneAxisFamily parses internal/config/envkeys.go at test time and
// returns the family: constants whose value starts with the factory or
// kanban prefix, plus the autonomy-tier axis. Key = axis value, value = the
// Go identifier carrying it.
func readLaneAxisFamily(t *testing.T) map[string]string {
	t.Helper()
	src, err := os.ReadFile(filepath.Join("..", "config", "envkeys.go"))
	if err != nil {
		t.Errorf("read envkeys.go for the axis family: %v", err)
		return nil
	}
	re := regexp.MustCompile(`(?m)^\s*(?:const\s+)?(\w+)\s*=\s*"([A-Z_0-9]+)"`)
	family := map[string]string{}
	for _, m := range re.FindAllStringSubmatch(string(src), -1) {
		v := m[2]
		if strings.HasPrefix(v, "MOAI_FACTORY_") || strings.HasPrefix(v, "MOAI_KANBAN") || v == config.EnvAutonomyTier {
			family[v] = m[1]
		}
	}
	return family
}

// productionLaneAxisReferences scans the package directory's non-test Go
// files and returns, per axis value, the files referencing it by identifier
// (config.<Name>) or by quoted literal — either form counts (spec.md §A.6).
func productionLaneAxisReferences(t *testing.T, family map[string]string) map[string][]string {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Errorf("read package directory for the reference scan: %v", err)
		return nil
	}
	refs := map[string][]string{}
	for _, e := range entries {
		if e.IsDir() || strings.HasSuffix(e.Name(), "_test.go") || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		src, err := os.ReadFile(e.Name())
		if err != nil {
			t.Errorf("read %s for the reference scan: %v", e.Name(), err)
			continue
		}
		for value, ident := range family {
			if strings.Contains(string(src), "config."+ident) || strings.Contains(string(src), `"`+value+`"`) {
				refs[value] = append(refs[value], e.Name())
			}
		}
	}
	return refs
}

// declaredLaneScrubSet intersects the TestMain scrub list with the family.
func declaredLaneScrubSet(family map[string]string) map[string]bool {
	scrub := map[string]bool{}
	for _, k := range laneEnvScrubAxes {
		if _, ok := family[k]; ok {
			scrub[k] = true
		}
	}
	return scrub
}

// sortedLaneAxisValues returns the family's axis values in sorted order so
// the guard's report is deterministic.
func sortedLaneAxisValues(family map[string]string) []string {
	values := make([]string, 0, len(family))
	for v := range family {
		values = append(values, v)
	}
	sort.Strings(values)
	return values
}

// stripLaneFamilyFromEnv removes every family axis from an env list by
// case-insensitive key comparison (Windows env names are case-insensitive
// and CI runs this binary there), plus the named extra keys, and returns
// the trimmed list.
func stripLaneFamilyFromEnv(t *testing.T, environ []string, family map[string]string, extra ...string) []string {
	t.Helper()
	drop := map[string]bool{}
	for v := range family {
		drop[strings.ToLower(v)] = true
	}
	for _, k := range extra {
		drop[strings.ToLower(k)] = true
	}
	kept := make([]string, 0, len(environ))
	for _, kv := range environ {
		name := kv
		if i := strings.IndexByte(kv, '='); i >= 0 {
			name = kv[:i]
		}
		if drop[strings.ToLower(name)] {
			continue
		}
		kept = append(kept, kv)
	}
	return kept
}

// TestLaneEnvAxesCovered is the coverage half of the guard pair
// (SPEC-TEST-ENV-HERMETIC-001 REQ-THE-005). Liveness assertions report with
// t.Errorf so the comparison always runs to completion and one red carries
// every message that applies.
func TestLaneEnvAxesCovered(t *testing.T) {
	family := readLaneAxisFamily(t)
	if len(family) < laneAxisFamilyFloor {
		t.Errorf("axis family from envkeys.go has %d members, below the floor %d: the family read has silently stopped matching", len(family), laneAxisFamilyFloor)
	}
	refs := productionLaneAxisReferences(t, family)
	if len(refs) == 0 {
		t.Errorf("reference scan found zero production references: the scan has silently stopped matching (moved files, renamed constants)")
	}
	scrub := declaredLaneScrubSet(family)
	if len(scrub) == 0 {
		t.Errorf("declared scrub set is empty: TestMain would strip no family axis")
	}

	exempt := map[string]bool{}
	for _, row := range laneEnvAxisExemptions {
		if strings.TrimSpace(row.Reason) == "" {
			t.Errorf("exemption row for %s has an empty reason", row.Axis)
		}
		if row.Citation == "" {
			t.Errorf("exemption row for %s has an empty citation", row.Axis)
			continue
		}
		if !strings.HasSuffix(row.Citation, "_test.go") {
			t.Errorf("exemption row for %s cites %s, which is not a _test.go file in this package directory", row.Axis, row.Citation)
			continue
		}
		src, err := os.ReadFile(row.Citation)
		if err != nil {
			t.Errorf("exemption row for %s cites %s, which cannot be read: %v", row.Axis, row.Citation, err)
			continue
		}
		ident, ok := family[row.Axis]
		cited := ok && (strings.Contains(string(src), "config."+ident) || strings.Contains(string(src), `"`+row.Axis+`"`))
		if !cited {
			t.Errorf("exemption row for %s cites %s, which does not reference the axis", row.Axis, row.Citation)
			continue
		}
		exempt[row.Axis] = true
	}

	for _, axis := range sortedLaneAxisValues(family) {
		if len(refs[axis]) == 0 || scrub[axis] || exempt[axis] {
			continue
		}
		t.Errorf("family axis %s is referenced by production (%s) but is in neither the TestMain scrub set nor the exemption table: a lane session's ambient value reaches every test in this binary", axis, strings.Join(refs[axis], ", "))
	}

	// Unasked stale-guard signal (spec.md §E): this package's coverage test
	// asserts the sibling package's guard file still exists and declares its
	// guard tests, so deleting or renaming one guard turns the other
	// package's run red without anyone asking.
	sibling, err := os.ReadFile(filepath.Join("..", "cli", "factory_env_axes_test.go"))
	if err != nil {
		t.Errorf("sibling guard file ../cli/factory_env_axes_test.go is missing: %v", err)
	} else {
		for _, name := range []string{"TestFactoryEnvAxesCovered", "TestFactoryEnvAxesScrubApplied"} {
			if !strings.Contains(string(sibling), "func "+name+"(") {
				t.Errorf("sibling guard file ../cli/factory_env_axes_test.go no longer declares %s", name)
			}
		}
	}
}

// TestLaneEnvAxesScrubApplied is the applied-behaviour half of the guard
// pair (SPEC-TEST-ENV-HERMETIC-001 REQ-THE-009). The parent re-executes this
// package's test binary with every family axis set to a sentinel value and
// the witness variable naming them; in the child, whose TestMain has already
// run the start-up scrub, the test asserts every referenced non-exempt axis
// is absent. The probe never calls the scrub function itself — only TestMain
// may, or the check proves nothing.
func TestLaneEnvAxesScrubApplied(t *testing.T) {
	family := readLaneAxisFamily(t)
	refs := productionLaneAxisReferences(t, family)
	exempt := map[string]bool{}
	for _, row := range laneEnvAxisExemptions {
		exempt[row.Axis] = true
	}

	if witness := os.Getenv(laneAxesProbeWitnessEnv); witness != "" {
		// Child mode: this process was re-executed by the parent below and
		// its TestMain has run. Assert absence of every referenced axis.
		for _, axis := range sortedLaneAxisValues(family) {
			if len(refs[axis]) == 0 || exempt[axis] {
				continue
			}
			if _, ok := os.LookupEnv(axis); ok {
				t.Errorf("family axis %s is still present in the child after TestMain: the start-up scrub is declared but not applied", axis)
			}
		}
		return
	}

	// Parent mode: compose the child env explicitly — parent env minus every
	// family axis, plus every family axis at the sentinel value, plus the
	// witness. This package composes no pin marker for its helper children
	// (the pre-flight census found none carrying a family axis as payload).
	childEnv := stripLaneFamilyFromEnv(t, os.Environ(), family, laneAxesProbeWitnessEnv)
	values := sortedLaneAxisValues(family)
	for _, axis := range values {
		childEnv = append(childEnv, axis+"="+laneAxesProbeSentinel)
	}
	childEnv = append(childEnv, laneAxesProbeWitnessEnv+"="+strings.Join(values, ","))

	ctx, cancel := context.WithTimeout(context.Background(), laneAxesProbeChildTimeout*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^"+t.Name()+"$", "-test.v")
	cmd.Env = childEnv
	out, err := cmd.CombinedOutput()
	t.Logf("child re-execution output (success and failure both logged):\n%s", out)
	if ctx.Err() != nil {
		t.Errorf("child re-execution exceeded its %ds bound (child runtime under load is recorded in the SPEC progress record)", laneAxesProbeChildTimeout)
		return
	}
	if err != nil {
		t.Errorf("child re-execution failed: %v", err)
	}
	if !strings.Contains(string(out), "--- PASS: "+t.Name()) {
		t.Errorf("child re-execution produced no --- PASS line for %s: the child ran nothing or died early", t.Name())
	}
}
