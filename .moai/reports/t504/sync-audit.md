# t504 sync-audit — SPEC-CODEX-SKILLCONFIG-SHAPE-001

- **Audit scope**: measurement-integrity audit (NFC-1 — no product code exists to audit). Review lens `--deep`. Audit target = the lab archive's internal consistency, the admissibility logic, AC coverage, and the evidence report's honesty — not an implementation.
- **Audit date**: 2026-09-13:13 KST (same day as the run; live `~/.codex/config.toml` re-hashed and still identical to the archived hash — the measured state has not drifted since the experiment).
- **Auditor**: sync-auditor (independent; lab archive read-only; no cell re-run per dispatch).

## Verdict

**PASS — 0.92 / 1.00**

4-dimension (harmonic mean): Functionality 0.95 · Security 0.95 · Craft 0.90 · Consistency 0.88. Must-pass firewall (Functionality + Security) satisfied independently. 4 defects found, all `[Low][optional]` — none affects any measured verdict.

## Per-check findings

| # | Check (what I recomputed) | Expected (report claim) | Observed (my measurement) | Source | Result |
|---|---------------------------|-------------------------|---------------------------|--------|--------|
| 1 | Round-1 cell rows: IV/N/V2/V1/V3/V4 — out/err bytes, rc, marker | 24957/0/0/IV=1 · 24910/0/0/1 · 0/151/1/0 · 0/151/1/0 · 0/151/1/0 · 24752/0/0/0 | identical, all 24 figures | `lab/cells/*.{out,err,rc}` | ✓ |
| 2 | Round-2 rows: V5/V6/V7 | 24753B / err 0 / rc 0 / marker 0 ×3 | identical | `lab/cells/V{5,6,7}.*` | ✓ |
| 3 | D-series rows: D1f/D2f/D1d/D2d | 0(24754)/1(24914)/1(24914)/1(24914), rc 0, err 0 | identical | `lab/cells/D*.*`, `lab/summary3.txt` | ✓ |
| 4 | D1f byte-delta claim ("정확히 1개분 160B") | 24914 − 24754 = 160 | 160 | `wc -c` on D2f.out/D1f.out | ✓ |
| 5 | `t504probe` absent from every negative cell | 0 in V4/V5/V6/V7/R.out | 0 in all five | grep on out files | ✓ |
| 6 | No-touch proof (AC-06) | before == after == `c45741c1…8638e01a` | identical; D-series re-check (`summary3.txt` notouch-recheck) identical; **live re-hash today identical** | `lab/hash-{before,after}.txt`, `summary3.txt`, `shasum` | ✓ |
| 7 | Version stamp (NFC-4) | `codex-cli 0.153.4` run-time measured | `lab/version.txt` = same; not a session carry-over | `lab/version.txt` | ✓ |
| 8 | F1 verbatim error | `missing field \`enabled\` in skills.config`, rc=1, out 0 | verbatim match, 151B ×3 identical across V1/V2/V3 | `lab/cells/V{1,2,3}.err` | ✓ |
| 9 | F1 corroboration: real config 49/49 `enabled` | 49/49 | 49/49 (independent recount, twice: `summary2.txt` census + my grep today; config 629 lines, hash unchanged so census still valid) | `summary2.txt`, live config | ✓ |
| 10 | RQ3: all 49 target paths dead | existing=0 | existence walk over all 49 path values: 0 exist; `~/.codex/skills/` holds only `.system`+`hatch-pet` | live config + `ls` | ✓ |
| 11 | RQ3: R-cell moai hits are plugin roots | only moai-cowork plugin roots r34–r51 | roots table r34–r51 = `~/.codex/plugins/cache/moai-cowork/moai-*`; moai-* skill entries in the list derive from those roots — attribution honest, RQ3 expected-absence maintained for the 49 entries | `lab/cells/R.out` | ✓ |
| 12 | Secret redaction (dispatch-mandated) | lab must not contain `backup-config.toml`, `Z_AI_API_KEY` | `Z_AI` 0 hits, `api_key` 0 hits, `sk-[A-Za-z0-9]{20,}` 0 hits; no `backup-config*` file present | grep/find over `lab/` | ✓ |
| 13 | Fixture-shape fidelity (13 scratch homes) | each config matches its claimed cell shape | all 13 match exactly (V5=file+true, V6=absent+true, V7=dir+true, D-series file/dir × true/false); SKILL.md fixtures carry T504MARKER in frontmatter description AND body; IV AGENTS.md carries T504IVMARKER | `lab/home-*/config.toml`, `lab/neutral/`, `lab/home-IV/AGENTS.md` | ✓ |
| 14 | Harness discipline (NFC-3, plan §B) | serial, 0 model calls, counts from disk files not pipes, no background | `harness{,2,3}.sh`: serial `run_cell` calls only; `codex debug prompt-input "hi"` only; `grep -c TOKEN "$out"` with `\|\| true` (exit-code gate absent); stdout/stderr captured separately; hash-before before any cell / hash-after after R / re-hash after D-series | `harness.sh:74-99`, `harness2.sh:49-58`, `harness3.sh:62-94` | ✓ |
| 15 | Scope discipline (NFC-1) | no file outside `.moai/{specs,reports}` | plan commit `028d75d2c`: spec.md+plan.md only (2 artifacts, Tier S); run commit `27b01eec1`: 74 files, all under `.moai/reports/t504/` + SPEC dir; zero product paths | `git show --stat` | ✓ |
| 16 | progress.md §E.2/§E.3/§F/§E.4 | claims match evidence | all headline figures match; §F `Decision: direct` recorded (grep AC = 2 ≥ 1); §E.4 carries sanctioned `sync_commit_sha: pending-backfill-sync` placeholder (D3 backfill window) | `progress.md` | ✓ |
| 17 | Admissibility / VOID conditions (plan §D) | none triggered | IV=1 (isolation valid — no fallback), N=1 (detector live), N positive so "N neg AND V2 pos" VOID unreachable; V6=0 clean negative; round-1 V1/V2/V3 correctly REFUSED as error-caused (rc=1) and re-anchored to V5/V6/V7 — the gates did exactly what AC-03 designed | report 채택성 게이트 절 + raw cells | ✓ |
| 18 | Vacuous-green attack (marker would appear without loading) | detector discriminative | D2f.out context shows the marker ONLY inside the loaded-skill list entry (`t504probe: T504MARKER … (file: r0/…)`); identical fixture left unloaded in V5/V6/V7 → 0. A blind detector cannot produce this pattern | `summary3.txt` marker-context, cells | ✓ (attack fails) |
| 19 | Inverted-control attack (convenient zero accepted) | zeros interrogated | the round-1 0s were NOT accepted — rc=1 + 151B stderr exposed them as schema errors, cells voided, matrix redesigned with `enabled = true`. The exact opposite of the inverted-control failure | report round-1 table + F1 section | ✓ (attack fails) |
| 20 | Control-inversion robustness ("would any verdict survive inverted controls?") | verdicts control-anchored | V5=0 stands only because N=1 proves detector sensitivity and D2f=1 proves sensitivity persists in config-present environments; V6=0 rules out path-existence confound. If any control were inverted, the report's own gate section would force VOID — the gates are load-bearing, not decoration | report 채택성 게이트 절 | ✓ |
| 21 | AC-CSS-001-01..08 coverage | each AC's observable present | AC-01 IV count+bytes ✓ (fallback correctly unneeded); AC-02 N=1 + convention named ✓; AC-03 V2 captures preserved verbatim + round-2 V6 clean negative ✓; AC-04 RQ1 DEAD attributed to V5 command + verbatim output + version stamp ✓ (re-anchored from literal V1 — transparently documented, judged in-substance compliant since V5 is the faithful realization of REQ-CSS-003's design once F1 revealed the schema requirement); AC-05 directory (V7 does-not-load) + enabled=false (V4-vs-V5 creation-axis no-difference; D-series gate-axis suppresses, file-shape only) verdicts recorded per axis ✓; AC-06 hash pair verbatim ✓; AC-07 marker 0 + complaint-absence recorded as an observation (0-byte stderr) + "no cell other than R touched the real home" stated ✓; AC-08 5 sections present, all cells have rows, Gaps substantive ✓ | evidence report §셀 매트릭스/§판정/§5-section | ✓ |

## Defects found

All four are `[Low][optional]` — none is blocking; none touches a measured verdict.

- **D1** `[Low]` `[optional]` `.moai/reports/t504/skills-config-path-shape.md:107` — the no-touch paragraph cites the pre-R backup copy as `lab/backup-config.toml`, a path that does not exist in the archive (deliberately excluded per the report's own §Gaps; the copy lived at `$EXP_ROOT/backup-config.toml`, `harness.sh:96`, and died with the tmp root). The no-touch proof itself rests on the hash pair, which is intact. Required fix: reword the parenthetical to `$EXP_ROOT/backup-config.toml (세션 종료 후 OS 폐기 — 아카이브 의도적 제외, §Gaps 참조)`.
- **D2** `[Low]` `[optional]` `.moai/reports/t504/skills-config-path-shape.md:112` — the t506 implication grounds "삭제가 스킬 로딩에 미치는 영향 = 0" in "(로드 무시 표면이므로)", a reason the same report's D-series overturns (file-shaped entries DO gate live skills). The conclusion still holds for these 49 entries — every path is nonexistent, so no live gate is removed — but the stated reason is stale and would mislead if generalized to entries pointing at live files. Required fix: re-ground the reason on the measured facts ("49개 경로가 전부 부재 — 제거되는 살아있는 게이트가 없다").
- **D3** `[Low]` `[optional]` `.moai/reports/t504/skills-config-path-shape.md:77` — the RQ1 heading parenthetical `DEAD (load-ignored)` is a stale round-1 label; the summary (line 12) and the RQ1 body (line 81, "완전히 무시는 아니다") carry the post-D-series correction, but the heading still reads as a flat load-ignored verdict. Required fix: `DEAD (생성 축)` or `DEAD (creation axis)`.
- **D4** `[Low]` `[optional]` `.moai/specs/SPEC-CODEX-SKILLCONFIG-SHAPE-001/spec.md:46` (carried at `plan.md:24`) — the plan-phase premise "every one `path = "…/moai-*/SKILL.md"`" is off by exactly one entry: 48 are `moai-*`, 1 is `moai/SKILL.md` (bare `moai` directory). Measured today on the byte-unchanged config. The evidence report does not repeat the claim, so no verdict is affected. Required fix: follow-up doc touch by the SPEC owner (SPEC body is post-run here; this audit writes only this verdict file).

## Gaps (what this audit did NOT verify)

- **No cell re-run** (per dispatch): the binary state is the operator's live environment; all verdicts rest on the archived captures. I verified the archive's internal consistency and the still-valid live hash, not fresh executions.
- **Codex internals**: the mechanism claims (gate keyed on SKILL.md file path; directory-shaped entries never match) are the report's observed-level claims. I verified the observations reproduce from raw files; codex-cli's parser source is not consulted.
- **R.out full-text complaint sweep**: I verified stderr = 0 bytes, rc = 0, marker = 0, and the moai-hit attribution via the roots table, but did not line-scan all 36,717 bytes of R.out for complaint-shaped text in stdout.
- **Chronology of rounds**: round-1 → round-2 → D-series ordering is inferred from harness argument shape (`harness2/3.sh` take `$EXP_ROOT`) and summary contents, not from timestamps (the tmp root is gone).
- **One stray file**: `lab/.moai/state/context-usage/3d48a876-….json` is a gitignored runtime session-state snapshot (`.gitignore:224` `**/.moai/state/`), NOT in the run commit, and carries no experiment data — noise, not evidence; not chased further.
- **Provenance of the 49 entries** — explicitly out of scope per spec.md §F.
- **Desktop-app surface, cross-version behavior, `enabled`-adjacent fields** — remain unmeasured, exactly as the report's own §Gaps discloses (the disclosure is honest and complete).

## Close recommendation

**Close the card as PASS** — the measurement is internally consistent to the byte across 40+ recomputed figures, the controls genuinely discriminate (both adversarial attack shapes fail), the no-touch proof holds against today's live hash, secret redaction is complete, and NFC-1 scope discipline is clean; land D1–D3 (one-line wording repairs, all in the evidence file) as an optional evidence-file polish commit if the lead wants the record self-consistent before t502 consumes it, and carry D4 to the next SPEC-body touch.
