## CON-002 Amendment Evidence

SPC-001 amends the FROZEN-zone clause CONST-V3R2-001 (`.claude/rules/moai/workflow/spec-workflow.md` SPEC+EARS format) to introduce hierarchical Acceptance Criteria schema. Per CON-002 §5, all 5 safety layers must furnish evidence before amendment lands.

### Layer 1 — Frozen Guard

- **Mechanism**: Pre-write check that no FROZEN-zone clause is modified without explicit amendment authority.
- **Evidence**:
  - Amendment vehicle: SPEC-V3R2-SPC-001 (this SPEC), declared as `amends CONST-V3R2-001` in spec.md §1.
  - Cross-link: `.claude/rules/moai/core/zone-registry.md` CONST-V3R2-001 `clause` field reads "SPEC+EARS format (amended by SPEC-V3R2-SPC-001 to add hierarchical AC schema)" (origin/main e47b5e20f).
  - Audit-trail tag: `@MX:WARN reason="FROZEN-zone amendment per CON-002 §5 Layer 1 (Frozen Guard); modifications require full Canary + HumanOversight cycle"` placed at `.claude/rules/moai/workflow/spec-workflow.md:139` (T-11 commit 7978f40d4).
- **Verdict**: PASS — amendment is explicit, not silent.

### Layer 2 — Canary Check

- **Mechanism**: Re-parse the last 10 landed v2 SPECs through the amended parser; no project may regress by >0.10 nor introduce SPC-001-attributable warnings.
- **Evidence**:
  - Canary log: `.moai/specs/SPEC-V3R2-SPC-001/canary-v2-reparse.txt` (commit e59c40535, 2026-05-13).
  - Coverage: 10 SPECs (WF-002, WF-003, WF-004, WF-005, CLI-TUI-001, CI-AUTONOMY-001, CON-002, MIG-001, RT-004, CATALOG-002).
  - Result: 9/10 `moai spec view` exit 0; 1/10 (CLI-TUI-001) exit 1 due to pre-existing missing `## Acceptance` header in the source SPEC — parser-unrelated, predates SPC-001.
  - Auto-wrap synthesis verification: CON-002 yielded 13 synthesised `.a` leaves; CATALOG-002 yielded 9 leaves. REQ-coverage parity confirmed: every top-level AC REQ tail migrated to the synthesised leaf without loss.
  - SPC-001-attributable regressions: 0.
- **Verdict**: PASS — no parser-attributable regression detected.

### Layer 3 — Contradiction Detector

- **Mechanism**: Scan for conflicting rules between the amendment and existing FROZEN/EVOLVABLE clauses.
- **Evidence**:
  - SPC-001 §11 self-audit confirms: hierarchical schema is purely additive to the flat schema; the 185-SPEC corpus remains 100% flat-parseable via auto-wrap (`internal/spec/parser.go:200-227`).
  - No FROZEN clause inverted: SPEC+EARS format remains required; only the AC sub-structure is extended.
  - MIG-001 §11.1 cross-link explicitly partitions responsibility: SPC-001 owns runtime auto-wrap; MIG-001 owns optional cosmetic rewrite. No behavioural overlap.
- **Verdict**: PASS — no contradictions surfaced.

### Layer 4 — Rate Limiter

- **Mechanism**: Cap of ≤3 FROZEN amendments per v3.x release cycle.
- **Evidence**:
  - v3.x cycle FROZEN amendments inventory (as of 2026-05-13):
    1. SPC-001 (this amendment) — hierarchical AC schema; lands now.
    2. HRN-002 (Sprint Contract) — references SPC-001 leaf AC IDs; planned for Sprint 12+.
    3. (slot reserved for emergency amendment; currently unused.)
  - Count: 1 of 3 used. Rate limit honoured.
- **Verdict**: PASS — within rate cap.

### Layer 5 — Human Oversight

- **Mechanism**: Maintainer approval recorded with timestamp + reviewer identity in the landing PR description before merge.
- **Evidence**:
  - Plan-auditor verdict: FAIL → BYPASSED (2026-05-11T00:00:00Z, reviewer: Goos Kim, reason: "MP-1/MP-2 findings are project-convention mismatches — intentional EARS modality block numbering; AC=Gherkin is project standard"). Recorded in `.moai/specs/SPEC-V3R2-SPC-001/progress.md` Wave A Complete block.
  - Run-phase HumanOversight: maintainer (Goos Kim, bobby@afamily.kr) approval required at PR open for SPC-001 final landing. To be recorded in the run-phase PR description with timestamp upon approval.
  - This evidence document itself serves as the maintainer's pre-approval brief.
- **Verdict**: PENDING-FINAL — maintainer approval required at PR open. All other 4 layers PASS.

## Overall Verdict

4/5 layers PASS, 1/5 (Human Oversight) PENDING maintainer approval at PR creation. Recommended action: open the run-phase PR with this evidence block in the PR description body, request maintainer (@goos / Goos Kim) approval comment with timestamp, then admin-merge.
