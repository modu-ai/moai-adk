package core

import (
	"encoding/json"
	"strings"
	"testing"
)

// Merge-gate (G) reasons are gate records on a stage, never PARKED overlays.
func TestGateReasonsAreTheMergeGateSetAndNeverParkable(t *testing.T) {
	want := []BlockReason{ReasonCheckNotPassed, ReasonDebtPresent, ReasonVerdictNotPass, ReasonProducerNotAllowed}
	got := GateReasons()
	if len(got) != len(want) {
		t.Fatalf("gate reasons = %v", got)
	}
	for _, r := range want {
		if !IsGateReason(r) {
			t.Errorf("%s must be a gate reason", r)
		}
		if err := (Park{Reason: r, Evidence: "e", Resume: "r"}).Validate(); err == nil {
			t.Errorf("%s is a gate reason and must not park", r)
		}
	}
	for _, r := range []BlockReason{ReasonRateLimited, ReasonDependency, ReasonForeignClosing, ReasonOperatorHold} {
		if IsGateReason(r) {
			t.Errorf("%s is not a gate reason", r)
		}
	}
}

func TestGateAndRefusalValidate(t *testing.T) {
	g := Gate{Reason: ReasonDebtPresent, Detail: "debt in audit", Resume: "same-HEAD rerun passes without debt", Head: "a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"}
	if err := g.Validate(); err != nil {
		t.Errorf("good gate: %v", err)
	}
	for name, bad := range map[string]Gate{
		"not a gate reason": {Reason: ReasonRateLimited, Detail: "d", Resume: "r"},
		"no detail":         {Reason: ReasonDebtPresent, Resume: "r"},
		"no resume":         {Reason: ReasonDebtPresent, Detail: "d"},
	} {
		if err := bad.Validate(); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
	if err := (Refusal{Reason: ReasonForeignClosing, Detail: "PR text would close #3000"}).Validate(); err != nil {
		t.Errorf("good refusal: %v", err)
	}
	for name, bad := range map[string]Refusal{
		"unknown reason": {Reason: "made-up", Detail: "d"},
		"no detail":      {Reason: ReasonForeignClosing},
	} {
		if err := bad.Validate(); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}

func TestParkGateAndRefusalJSONKeys(t *testing.T) {
	b, _ := json.Marshal(Park{Reason: ReasonRateLimited, Evidence: "429", Resume: "after 15:00"})
	for _, k := range []string{`"reason":"rate-limited"`, `"evidence":"429"`, `"resume_condition":"after 15:00"`} {
		if !strings.Contains(string(b), k) {
			t.Errorf("park json %s lacks %s", b, k)
		}
	}
	b, _ = json.Marshal(Gate{Reason: ReasonCheckNotPassed, Detail: "d", Resume: "r", Head: "h"})
	for _, k := range []string{`"reason":"check-not-passed"`, `"detail":"d"`, `"resume_condition":"r"`, `"head":"h"`} {
		if !strings.Contains(string(b), k) {
			t.Errorf("gate json %s lacks %s", b, k)
		}
	}
}

func TestIsParkableSeparatesOverlayGateAndRefusalReasons(t *testing.T) {
	for _, r := range append(CanonParkReasons(), ReasonRateLimited, ReasonOperatorHold) {
		if !IsParkable(r) {
			t.Errorf("%s must be parkable", r)
		}
	}
	for _, r := range append(GateReasons(), ReasonForeignClosing, ReasonNotIssueAuthor, "made-up") {
		if IsParkable(r) {
			t.Errorf("%s must not be parkable", r)
		}
	}
}
