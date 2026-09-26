package escalation

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// invariantFrozenFiles is the contract invariant that enables class 2b.
const invariantFrozenFiles = "frozen-files"

// invariantConstitutionPrefix marks a constitution:<glob> invariant, which has
// no mechanical violation signal and is always not-observed (spec.md §F O2).
const invariantConstitutionPrefix = "constitution:"

// writeTarget is one write-capable call's target in the forms the path
// classes need.
type writeTarget struct {
	abs    string // absolute, as the tool named it (joined to CWD if relative)
	rel    string // slash path relative to the worktree root; "" when outside
	inRoot bool
}

// writeTarget resolves the event's write target against the worktree root,
// comparing canonical spellings so a symlinked parent does not move a path
// in or out of the root.
func (r *run) writeTarget() writeTarget {
	p := r.ev.FilePath
	if !filepath.IsAbs(p) {
		p = filepath.Join(r.ev.CWD, p)
	}
	w := writeTarget{abs: filepath.Clean(p)}
	rel, err := filepath.Rel(canonicalPath(r.root), canonicalPath(w.abs))
	if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel) {
		w.rel, w.inRoot = filepath.ToSlash(rel), true
	}
	return w
}

// classFrozenFile trips class 2b (invariant-violation, sub-kind frozen-file)
// for a write to a path matching the A1-derived frozen_files cached at
// arming, whoever the caller is (REQ-AE-007). The list is non-empty only when
// the contract's invariants carry frozen-files.
func (r *run) classFrozenFile(w writeTarget) {
	a := r.st.Armed
	if !w.inRoot || len(a.FrozenFiles) == 0 {
		return
	}
	glob := matchAny(a.FrozenFiles, w.rel)
	if glob == "" {
		return
	}
	cdata, _ := os.ReadFile(a.ContractPath)
	r.writeRecord(Record{
		Kind: KindContract, Class: ClassInvariantViolation, EscalateOn: ClassInvariantViolation,
		Fingerprint: Fingerprint(ClassInvariantViolation, "frozen-file", w.rel, glob),
		ContractRef: contractRef(ContractItemLine(cdata, invariantFrozenFiles, "invariants"), cdata, "invariants"),
		Observation: fmt.Sprintf("Sub-kind frozen-file: %s %s (tripped by frozen_files glob %s)", r.ev.ToolName, w.rel, glob),
		Options: []string{
			"Revert the write to the frozen file",
			"Stop and ask the operator to change the frozen file outside contract mode",
		},
	})
}

// classOwnershipMove trips class 3 (REQ-AE-008, REQ-AE-012, REQ-AE-013). The
// order is fixed: a write into the contract store trips first, ahead of every
// exemption and never listed as not-observed; then an in-root write is judged
// against the card's exemptions, effective_never, and ownership.write; then an
// outside-root write is exempt under a determined root, trips when every root
// was determined, and is listed not-observed otherwise.
func (r *run) classOwnershipMove(w writeTarget) {
	a := r.st.Armed
	cdata, _ := os.ReadFile(a.ContractPath)
	if store, err := StoreDir(r.root); err == nil && within(w.abs, store) {
		r.ownershipRecord(w.abs, "contract-store", fmt.Sprintf("%s %s (tripped by the contract store: judged before every exemption)",
			r.ev.ToolName, w.abs), contractRef(0, cdata, "ownership"))
		return
	}
	if !w.inRoot {
		r.outsideRoot(w, cdata)
		return
	}
	if strings.HasPrefix(w.rel, ".moai/reports/"+r.card+"/") || strings.HasPrefix(w.rel, ".moai/state/") ||
		matchAny(a.Scratch, w.rel) != "" {
		return
	}
	if len(a.Write) == 0 {
		// ownership globs unreadable from the arming snapshot (REQ-AE-022).
		r.notObserved("ownership-move", w.rel, []string{"ownership"})
		return
	}
	if glob := matchAny(a.EffectiveNever, w.rel); glob != "" {
		if line := ContractItemLine(cdata, glob, "ownership", "never"); line != 0 {
			r.ownershipRecord(w.rel, glob, fmt.Sprintf("%s %s (tripped by ownership.never glob %s)", r.ev.ToolName, w.rel, glob),
				contractRef(line, cdata, "ownership"))
		} else {
			r.ownershipRecord(w.rel, glob, fmt.Sprintf("%s %s (tripped by effective_never %s: immutable after signing)",
				r.ev.ToolName, w.rel, glob), contractRef(ContractLine(cdata, "signature"), cdata, "ownership"))
		}
		return
	}
	if matchAny(a.Write, w.rel) != "" {
		return
	}
	r.ownershipRecord(w.rel, "ownership.write", fmt.Sprintf("%s %s (tripped by ownership.write: no write glob covers it)",
		r.ev.ToolName, w.rel), contractRef(ContractLine(cdata, "ownership", "write"), cdata, "ownership"))
}

// outsideRoot judges a write outside the worktree root against the exemption
// roots of design.md §C.9.
func (r *run) outsideRoot(w writeTarget, cdata []byte) {
	if within(w.abs, tempRoot()) {
		return
	}
	var undetermined []string
	if r.ev.ScratchpadDir == "" {
		undetermined = append(undetermined, rootScratchpad)
	} else if within(w.abs, r.ev.ScratchpadDir) {
		return
	}
	mem, ok := memoryRoots(r.root)
	if !ok {
		undetermined = append(undetermined, rootMemory)
	}
	for _, m := range mem {
		if within(w.abs, m) {
			return
		}
	}
	if len(undetermined) > 0 {
		r.notObserved("ownership-move", w.abs, append([]string{"outside-root write " + w.abs}, undetermined...))
		return
	}
	r.ownershipRecord(w.abs, "outside-root", fmt.Sprintf("%s %s (tripped by outside-root: no exemption root covers it)",
		r.ev.ToolName, w.abs), contractRef(0, cdata, "ownership"))
}

// ownershipRecord writes one ownership-move record.
func (r *run) ownershipRecord(target, matched, observation, ref string) {
	r.writeRecord(Record{
		Kind: KindContract, Class: ClassOwnershipMove, EscalateOn: ClassOwnershipMove,
		Fingerprint: Fingerprint(ClassOwnershipMove, target, matched),
		ContractRef: ref, Observation: observation,
		Options: []string{
			"Revert the write and keep within the contract's ownership",
			"Amend the contract's ownership and re-sign it",
		},
	})
}

// classInvariantCommand observes a shell command against the command-kind
// invariants (REQ-AE-006): an equal command is recorded as executed, and one
// that failed trips invariant-violation (command). The command is never re-run.
func (r *run) classInvariantCommand() {
	a := r.st.Armed
	cmd := strings.TrimSpace(r.ev.Command)
	if cmd == "" || !slices.Contains(commandInvariants(a.Invariants), cmd) {
		return
	}
	if !slices.Contains(r.st.Counters.ExecutedInvariants, cmd) {
		r.st.Counters.ExecutedInvariants = append(r.st.Counters.ExecutedInvariants, cmd)
		r.dirty = true
	}
	if !r.ev.Failed {
		return
	}
	cdata, _ := os.ReadFile(a.ContractPath)
	r.writeRecord(Record{
		Kind: KindContract, Class: ClassInvariantViolation, EscalateOn: ClassInvariantViolation,
		Fingerprint: Fingerprint(ClassInvariantViolation, "command", cmd),
		ContractRef: contractRef(ContractItemLine(cdata, cmd, "invariants"), cdata, "invariants"),
		Observation: fmt.Sprintf("Sub-kind command: the invariant command `%s` completed with a failure", cmd),
		Options: []string{
			"Fix the failure and re-run the invariant command",
			"Stop and review whether the contract's invariant still holds",
		},
	})
}

// checkpointNotObserved lists, at a commit checkpoint, every detection the
// checkpoint could not complete (REQ-AE-022): constitution: invariants,
// command invariants no call executed since the previous checkpoint, and
// unreadable ownership. The executed set then starts over.
func (r *run) checkpointNotObserved() {
	a := r.st.Armed
	if a == nil {
		return
	}
	var items []string
	for _, inv := range a.Invariants {
		if strings.HasPrefix(inv, invariantConstitutionPrefix) {
			items = append(items, inv)
		}
	}
	for _, c := range commandInvariants(a.Invariants) {
		if !slices.Contains(r.st.Counters.ExecutedInvariants, c) {
			items = append(items, c)
		}
	}
	if len(a.Write) == 0 {
		items = append(items, "ownership")
	}
	if len(items) > 0 {
		r.notObserved("checkpoint", "commit checkpoint", items)
	}
	if len(r.st.Counters.ExecutedInvariants) > 0 {
		r.st.Counters.ExecutedInvariants = nil
		r.dirty = true
	}
}

// commandInvariants returns the command-kind invariants: neither
// frozen-files nor constitution:<glob>.
func commandInvariants(invariants []string) []string {
	var out []string
	for _, inv := range invariants {
		if inv != invariantFrozenFiles && !strings.HasPrefix(inv, invariantConstitutionPrefix) {
			out = append(out, inv)
		}
	}
	return out
}

// notObserved appends a not-observed line (REQ-AE-022).
func (r *run) notObserved(class, detail string, items []string) {
	r.appendLog(LogEntry{Kind: LineNotObserved, Card: r.card, Class: class, Detail: detail, NotObserved: items})
}
