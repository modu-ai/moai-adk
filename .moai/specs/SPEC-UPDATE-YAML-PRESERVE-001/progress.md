# Progress — SPEC-UPDATE-YAML-PRESERVE-001

## §E.1 Plan-phase Audit-Ready Signal

- **Artifacts**: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` created 2026-07-31 by `manager-spec`.
- **Tier**: M (justified in `plan.md` §A).
- **SPEC ID regex check**: executed as Bash, verbatim output `PASS` for `SPEC-UPDATE-YAML-PRESERVE-001` against `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$`.
- **ID uniqueness**: `ls .moai/specs/ | grep -c "SPEC-UPDATE-YAML-PRESERVE-001"` → `0` prior occurrences.
- **Baseline verification**: every file:line reference in `spec.md` §A was re-read this session and confirmed against the current tree (table in `plan.md` §A). The round-trip loss was reproduced by executing a scratch test against `internal/template/templates/.moai/config/sections/cache.yaml` with the pinned `gopkg.in/yaml.v3 v3.0.1` — observed `comments in=16 out=0`, keys alphabetized, `"1h"` → `1h`, 2-space → 4-space indent.
- **Edge-case survey**: executed over the 30 shipped section templates — 0 anchors, 0 merge keys, 0 multi-document; 8 files carry block sequences (so REQ-UYP-014 is a live decision, the other three are defensive contracts). 1 `.tmpl` (`quality.yaml.tmpl`) is unparseable by the map decoder due to unquoted `{{...}}` placeholders — see `plan.md` Decision D6 and `spec.md` §A blast radius.
- **Open clarifications**: none. Decision D5 (`SaveTemplateDefaults` base derivation) is explicitly DEFERRED with rationale in `plan.md` §E, per the SPEC's fourth stated requirement.
- **Plan-audit iter-1 (FAIL 0.71) revision (2026-08-03)**: D1–D6 + SHOULD-FIX D9–D13 applied. `version: "0.2.0"`, `updated: 2026-08-03`. Re-audit scoped to the 6 blocking items. Coverage baseline captured at plan time: **88.9%** (acceptance.md AC-UYP-020). Misnamed subtest at `update_yaml_test.go:591-603` was rewritten by an interim commit on `main` (the destructive `t.Errorf("expected user_added to be dropped…")` line is gone, subtest renamed to `"user added key not in base preserved"`); M4 therefore becomes a partial no-op + add the missing stderr-advisory sibling assertion (REQ-UYP-007). See `plan.md` M4 pre-flight note.
- **Status**: `draft` — awaiting re-audit and Implementation Kickoff Approval.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
