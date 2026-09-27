package escalation

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/spec"
)

// The remaining disarm reasons of REQ-AE-017 (contract-absent and
// signature-invalid are declared with the detector).
const (
	DisarmTerminalStatus = "terminal-status"
	DisarmCardMismatch   = "card-mismatch"
	DisarmStateTamper    = "state-tamper"
)

// tamperCheck runs design.md §C.6 steps 1-4 and returns the tamper findings
// and the evidence ids a judgment will account for. It never reads the state
// file as a disarm: a missing or emptied arming while the log shows the card
// armed is tamper, not a plain disarm.
func (r *run) tamperCheck(stExists bool, stBytes []byte) (details, accounted []string) {
	// Step 1: a chain break no earlier judgment accounted for.
	if line, ok := r.lg.unaccountedBreak(); ok {
		details = append(details, fmt.Sprintf("card log hash chain broken at line %d", line+1))
	}
	// Step 2: whenever the state file exists, its bytes must match the log's
	// latest state entry.
	if stExists {
		latest, ok := r.lg.latest(LineState)
		switch {
		case !ok:
			details = append(details, "card state file exists but the card log holds no state entry")
		case latest.StateSHA256 != SHA256Hex(stBytes):
			details = append(details, "card state file does not match the digest in the latest state entry")
		}
	}
	if r.lg.Armed() {
		// Step 4: an armed log with no arming in the state file.
		if !stExists || r.st.Armed == nil {
			details = append(details, "card log shows the card armed but the state file carries no arming")
		}
		return details, accounted
	}
	// Step 3: the log does not show the card armed; look for arming evidence
	// it does not account for.
	if a := r.st.Armed; a != nil && !r.lg.accounts("sha:"+a.ContractSHA256) {
		details = append(details, "card state names an arming of "+a.SpecID+" the card log does not record")
	}
	for _, ev := range r.recordEvidence() {
		details = append(details, ev.detail)
		accounted = append(accounted, ev.id)
	}
	return details, accounted
}

// evidence is one record that shows an arming the log does not account for.
type evidence struct{ id, detail string }

// recordEvidence lists unaccounted arming evidence among the card's records:
// a detection-disarmed record whose fingerprint no disarmed entry carries,
// and a contract-kind record while the log holds no armed entry.
func (r *run) recordEvidence() []evidence {
	paths, _ := filepath.Glob(filepath.Join(RecordDir(r.root, r.card), "*.md"))
	_, hasArmed := r.lg.latest(LineArmed)
	var out []evidence
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		rec, err := ParseRecord(data)
		if err != nil {
			continue
		}
		name := filepath.Base(p)
		switch {
		case rec.Class == ClassDetectionDisarmed && !r.lg.accounts("fp:"+rec.Fingerprint):
			out = append(out, evidence{"fp:" + rec.Fingerprint, "detection-disarmed record " + name + " has no disarmed entry"})
		case rec.Kind == KindContract && !hasArmed && !r.lg.accounts("record:"+name):
			out = append(out, evidence{"record:" + name, "contract record " + name + " exists while the card log holds no armed entry"})
		}
	}
	return out
}

// judgeTamper writes the state-tamper record and one disarmed entry (starting
// a new log if none exists), and retires any arming in the state file.
func (r *run) judgeTamper(details, accounted []string) {
	sha, specID := "", ""
	switch {
	case r.st.Armed != nil:
		sha, specID = r.st.Armed.ContractSHA256, r.st.Armed.SpecID
	default:
		if e, ok := r.lg.latest(LineArmed); ok {
			sha, specID = e.ContractSHA256, e.Spec
		} else if r.st.LastArming != nil {
			sha, specID = r.st.LastArming.ContractSHA256, r.st.LastArming.SpecID
		}
	}
	fp := Fingerprint(ClassDetectionDisarmed, sha)
	if sha == "" {
		fp = Fingerprint(ClassDetectionDisarmed, DisarmStateTamper, r.card)
	}
	r.writeRecordSpec(Record{
		Kind: KindOperational, Class: ClassDetectionDisarmed, Fingerprint: fp,
		ContractRef: "disarm:" + DisarmStateTamper,
		Observation: fmt.Sprintf("Card %s: the card state or its audit log changed outside the detector.\n\n- %s",
			r.card, strings.Join(details, "\n- ")),
		Options: []string{
			"Inspect who changed the card state or log, then re-sign so detection re-arms",
			"Continue without contract-mode detection and review the run by hand",
		},
	}, specID)
	r.appendLog(LogEntry{Kind: LineDisarmed, Card: r.card, Spec: specID, Reason: DisarmStateTamper,
		Fingerprint: fp, Accounted: accounted, Detail: strings.Join(details, "; ")})
	if a := r.st.Armed; a != nil {
		a.DisarmReasons = append(a.DisarmReasons, DisarmStateTamper)
		r.st.LastArming = a
	}
	r.st.Armed = nil
	r.st.VerifyCache = nil
	r.dirty = true
}

// armedDisarmReason runs step 5 for the armed card and returns the first
// disarm reason found, "" when none. Class 1 runs inside the verify step.
func (r *run) armedDisarmReason() (string, string) {
	return r.disarmReason(r.st.Armed, true)
}

// disarmReason evaluates the disarm reasons against an arming. Verify runs
// fresh at a commit or on-demand checkpoint, and otherwise only when the
// contract bytes differ from the cached digest (spec.md C4). With live set,
// verify results update the cache and class 1 is evaluated.
func (r *run) disarmReason(a *Arming, live bool) (string, string) {
	data, err := os.ReadFile(a.ContractPath)
	if errors.Is(err, fs.ErrNotExist) {
		return DisarmContractAbsent, "the armed contract " + a.ContractPath + " is gone"
	}
	if err != nil {
		r.notChecked("disarm-check", err.Error())
		return "", ""
	}
	if c, ok := cardField(data); ok && c != r.card {
		return DisarmCardMismatch, "the contract's card field now reads " + c
	}
	claimants, _, err := findClaimants(r.root, r.card)
	if err != nil {
		r.notChecked("disarm-check", err.Error())
		return "", ""
	}
	dir := filepath.Dir(a.ContractPath)
	for _, c := range claimants {
		if c.dir != dir && !isTerminal(c.status) {
			return DisarmCardMismatch, "another contract claims the card: " + c.specID
		}
	}
	status, _ := spec.ParseStatus(dir)
	if isTerminal(status) {
		return DisarmTerminalStatus, a.SpecID + " status is " + status
	}
	digest := SHA256Hex(data)
	cache := r.st.VerifyCache
	fresh := r.isCommitCheckpoint() || r.ev.Hook == HookCheckpoint
	if !live {
		if digest == a.ContractDigest && !fresh {
			return "", ""
		}
	} else if !fresh && cache != nil && cache.ContractDigest == digest {
		if cache.State != contract.StateSignedValid {
			return DisarmSignatureInvalid, "cached verify: " + strings.Join(cache.Reasons, ", ")
		}
		return "", ""
	}
	rep, err := verifySpec(dir, status, r.verifyEnv())
	if err != nil {
		r.notChecked("verify", err.Error())
		return "", ""
	}
	if live {
		r.st.VerifyCache = &VerifyCache{ContractDigest: digest, State: rep.State, Reasons: slices.Clone(rep.Reasons)}
		r.dirty = true
		r.classAcceptanceChange(rep, data)
	}
	if rep.State != contract.StateSignedValid {
		return DisarmSignatureInvalid, "verify " + rep.State + ": " + strings.Join(rep.Reasons, ", ")
	}
	return "", ""
}

// lastArmingReasons handles a disarm reason observed after the first
// (REQ-AE-017): a reason the ended arming has not seen increments that
// arming's record and writes nothing new.
func (r *run) lastArmingReasons() {
	a := r.st.LastArming
	reason, detail := r.disarmReason(a, false)
	if reason == "" || slices.Contains(a.DisarmReasons, reason) {
		return
	}
	a.DisarmReasons = append(a.DisarmReasons, reason)
	r.dirty = true
	r.writeRecordSpec(Record{
		Kind: KindOperational, Class: ClassDetectionDisarmed,
		Fingerprint: Fingerprint(ClassDetectionDisarmed, a.ContractSHA256),
		ContractRef: "disarm:" + reason,
		Observation: detail,
		Options:     []string{"Restore the contract state and re-sign", "Review the run by hand"},
	}, a.SpecID)
}
