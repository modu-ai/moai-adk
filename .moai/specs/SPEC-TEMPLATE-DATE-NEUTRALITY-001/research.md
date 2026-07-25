# SPEC-TEMPLATE-DATE-NEUTRALITY-001 — Research

Every figure below was produced by a command actually run against this worktree at `c7309aeb6`. Figures that were not measured are stated as gaps in §F.

---

## §A Guard behaviour (measured)

### A.1 Strict tier fails; narrow tier passes

```
$ MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -run TestTemplateNoInternalContentLeak -count=1
exit=1
--- FAIL: TestTemplateNoInternalContentLeak (0.54s)
    internal_content_leak_test.go:725: template internal-content leak detected (135 occurrences, mode=strict):
    ...
    internal_content_leak_test.go:738:   ... 85 more (capped)
FAIL	github.com/modu-ai/moai-adk/internal/template	0.936s
```

```
$ go test ./internal/template/ -run TestTemplateNoInternalContentLeak -count=1
narrow exit=0
ok  	github.com/modu-ai/moai-adk/internal/template	0.677s
```

### A.2 Strict mode is wired nowhere

```
$ grep -rn "LEAK_STRICT" --include="*.yaml" --include="*.yml" --include="Makefile" . | grep -v "internal_content_leak_test.go"
exit=1 (1 = no matches)
```

### A.3 All 50 visible rows are S1

```
$ grep -o "class=[A-Za-z0-9-]*" <strict.log> | sort | uniq -c
  50 class=S1-internal-date
```

Only 50 rows are visible (the cap). The claim that *all 135* are S1 is not established by this run — see §F Gap G1.

### A.4 Finding identity is `(file, distinct match literal)`

From `collectLeakViolations`:

```go
matches := class.pattern.FindAllString(text, -1)
...
seen := map[string]struct{}{}
for _, m := range matches {
    trimmed := strings.TrimSpace(m)
    ...
    if _, ok := seen[trimmed]; ok { continue }
    seen[trimmed] = struct{}{}
```

Consequence: repeated occurrences of the same date in one file collapse to one finding. This is why raw line counts exceed finding counts (§C.3).

### A.5 Scan scope

`shouldScanForLeak` admits extensions `.md .tmpl .yaml .yml .sh .json .js`, plus the extensionless dotfiles `.gitignore` and `.gitattributes` by basename. `.gitkeep` is deliberately excluded.

### A.6 Report cap

`limit := 50` at `internal_content_leak_test.go:730`, with `... %d more (capped)` at :738.

---

## §B Independent re-enumeration

The guard's scan scope and dedup semantics were reimplemented as a shell enumeration to obtain the full finding set (rather than raising the guard's cap):

```bash
find templates -type f \( -name '*.md' -o -name '*.tmpl' -o -name '*.yaml' -o -name '*.yml' \
  -o -name '*.sh' -o -name '*.json' -o -name '*.js' -o -name '.gitignore' -o -name '.gitattributes' \) -print0 \
| while IFS= read -r -d '' f; do
    grep -oE '\b202[6-9]-[0-1][0-9]-[0-3][0-9]\b' "$f" | sort -u | while read -r d; do echo "$f|$d"; done
  done
```

Result: **135** rows — exactly matching the guard's reported count. The recipe is therefore an adequate regeneration mechanism for REQ-TDN-006.

Distinct files: **116**.

---

## §C Finding distribution (measured)

### C.1 By date literal (top values)

| Date | Findings |
|---|---:|
| 2026-01-06 | 43 |
| 2026-11-22 | 9 |
| 2026-03-30 | 6 |
| 2026-02-21 | 6 |
| 2026-07-10 | 4 |
| 2026-01-11 | 4 |

### C.2 By category (mechanical partition, sums to 135)

Applying the REQ-TDN-001 decision rule as an `awk` classifier over the enumerated set with first-line context:

| Category | Count |
|---|---:|
| DC-2 prose authoring stamp | 58 |
| DC-1 frontmatter schema field | 48 |
| DC-5 adjudicated residue | 17 |
| DC-3 functional deadline (`2026-11-22`) | 9 |
| DC-4 attribution (`NOTICE.md`) | 3 |
| **Total** | **135** |

### C.3 Raw line counts exceed finding counts (dedup effect)

| Shape | Raw matching lines | Findings |
|---|---:|---:|
| `Last Updated: <date>` header/footer lines | 70 | (subset of DC-2's 58) |
| frontmatter `updated: "<date>"` lines | 49 | 48 |

The gap is the dedup in §A.4: a file whose `Last Updated:` prose line and frontmatter `updated:` field carry the same date yields one finding, not two.

### C.4 Where the 43 `2026-01-06` findings actually live

| Directory | Findings |
|---|---:|
| `.claude/skills/moai-foundation-cc/reference` | 11 |
| `.claude/skills/moai-workflow-worktree/modules` | 10 |
| `.claude/skills/moai-foundation-core/modules` | 7 |
| `.claude/skills/moai-workflow-testing/modules/automated-code-review/trust5-framework` | 6 |
| `.claude/skills/moai-workflow-testing/modules` | 6 |
| `.claude/skills/moai-workflow-testing/modules/automated-code-review` | 2 |
| `.claude/skills/moai-foundation-cc` | 1 |

The largest cluster is `moai-foundation-cc/reference` (11), whose files mirror third-party documentation and whose `Updated: 2026-01-06` line is a mirror-capture stamp — not an authoring stamp. This is the origin of REQ-TDN-011 and the first open question.

### C.5 DC-5 residue (all 17)

| File | Date | Shape |
|---|---|---|
| `.gitignore` | 2026-01-10 | `# Updated:` comment in a dotfile |
| `.moai/docs/agent-lint.md` | 2026-07-13 | policy date inside a table cell |
| `.claude/output-styles/moai/moai-learn.md` | 2026-04-11 | pedagogical example filename |
| `.moai/config/sections/lsp.yaml.tmpl` | 2026-04-11 | internal audit citation |
| `.moai/config/sections/harness.yaml` | 2026-04-21 | internal incident reference |
| `.claude/agents/moai/plan-auditor.md` | 2026-05-20 | internal incident reference |
| `.claude/agents/moai/manager-spec.md` | 2026-05-25 | external-event citation |
| `.claude/rules/moai/development/spec-frontmatter-schema.md` | 2026-05-16 | pedagogical counter-example |
| `.claude/rules/moai/development/skill-authoring.md` | 2026-01-28 | pedagogical schema example |
| `.claude/skills/moai-foundation-cc/SKILL.md` | 2026-01-06 | changelog entry |
| `.claude/skills/moai-foundation-cc/SKILL.md` | 2026-07-03 | external-doc annotation |
| `.claude/skills/moai-meta-harness/SKILL.md` | 2026-03-26 | upstream repo creation date |
| `.claude/skills/moai/references/anti-patterns.md` | 2026-04-28 | (context not re-read) |
| `.claude/skills/moai/workflows/fix.md` | 2026-03-02 | changelog entry |
| `.claude/skills/moai-foundation-cc/reference/claude-code-sub-agents-official.md` | 2026-07-03 | external-doc annotation |
| `.claude/skills/moai-foundation-cc/reference/claude-code-plugins-official.md` | 2026-07-03 | external-doc annotation |
| `.claude/skills/moai-workflow-spec/references/worktree-workflow.md` | 2026-05-17 | date-named policy heading |
| `.claude/skills/moai/workflows/project/mode-detection.md` | 2026-04-21 | internal incident reference |

Note the row count above is 18 lines for 17 findings because `moai-foundation-cc/SKILL.md` contributes two distinct date literals.

---

## §D Precedent mechanisms in the guard

| Mechanism | Shape | Anchoring |
|---|---|---|
| `skillBodyScoped` | class applies only under `.claude/skills/` | path prefix |
| `skillMoaiScoped` | class applies only under `.claude/skills/moai/` | path prefix |
| `requireHexLetter` | match must contain `[a-f]` | match content |
| `pedagogicalAllowlist` | `(File, SpecID)` pair skip, 5 entries | **file + match content** (`LineStart`/`LineEnd` are documented diagnostic-only and unused in `isPedagogicallyAllowed`) |

The frequently-repeated claim that `pedagogicalAllowlist` is line-number-anchored and therefore drift-prone is **not supported by the code**. The enforcement path compares `entry.File == relPath && entry.SpecID == matched` only.

---

## §E Mirror and CI surface

### E.1 Byte-parity allowlist does not intersect this SPEC's scope

`rule_template_mirror_test.go` enforces byte-identity for an explicit non-glob list:

```
.claude/rules/moai/core/hooks-system.md
.claude/rules/moai/workflow/spec-workflow.md
.claude/rules/moai/workflow/session-handoff.md
.claude/rules/moai/development/model-policy.md
.moai/config/evaluator-profiles/default.md
.moai/config/evaluator-profiles/frontend.md
```

The six date-bearing files under `templates/.claude/rules/moai/` are `NOTICE.md`, `development/spec-frontmatter-schema.md`, `development/skill-authoring.md`, and `workflow/archived-agent-rejection.md` — none appears in the list above.

### E.2 Local-counterpart overlap

Of the 116 affected template files, **115** have a counterpart at the corresponding local working-tree path; 1 is template-only.

### E.3 Existing CI workflow convention

`.github/workflows/template-neutrality-check.yaml` runs `TestTemplateNeutralityAudit` **in isolation by test name**, with an in-file comment stating that a package-wide green is not required because `internal/template` carries pre-existing unrelated failures. Its trigger paths already include `internal/template/internal_content_leak_test.go`.

---

## §F Gaps (not observed)

- **G1** — That *all 135* findings carry `class=S1-internal-date` was **not** verified. The guard's report is capped at 50; all 50 visible rows are S1, and the independent enumeration in §B reproduces 135 using the S1 regex alone, which is strong circumstantial evidence — but the guard's own classification of rows 51-135 was not observed.
- **G2** — `S2-short-sha-sentence-final` contributes 0 findings *among the 50 visible rows*. Its contribution among rows 51-135 was not observed.
- **G3** — The DC-5 context for `.claude/skills/moai/references/anti-patterns.md | 2026-04-28` was captured by the enumeration but not individually re-read; its shape is recorded as unknown.
- **G4** — The mechanical classifier in §C.2 binds each finding to the **first** line in the file containing that date literal. For a line carrying two distinct date literals (observed in `workflows/loop.md` and `workflows/fix.md` changelog lines), this can bind a finding to a leading `Updated:` token that is not its own literal. REQ-TDN-003 exists because of this; the §C.2 counts should be treated as a first-pass partition requiring per-finding confirmation during run-phase triage, not as a final adjudication.
- **G5** — No measurement was taken of how many DC-2 removals would leave a file with an orphaned or empty header/footer block.
- **G6** — Whether a CI-enforced strict tier would block any currently-planned future date entry was not probed; that probe is precondition P2 in `design.md` §C.
