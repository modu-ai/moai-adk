# Progress — SPEC-ASTGREP-MULTILANG-001

## §E.1 Plan-phase Audit-Ready Signal

- **SPEC**: SPEC-ASTGREP-MULTILANG-001 (corrected from invalid `SPEC-ASTGREP-16LANG-001`; segment `16LANG` is digit-leading and fails `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`).
- **Tier**: M | **Era**: V3R6 | **Status**: draft.
- **Artifacts created (plan-phase)**: spec.md, plan.md, acceptance.md, progress.md.
- **Ground Truth**: GT-1..GT-5 recorded in spec.md from direct source inspection this session (`sg` = ast-grep 0.40.5 present). Task premise "gate OFF by default" corrected to the verified GT-4 statement (CLI path unconditional; commit-gate path off via a config key-path mismatch).
- **Scope decision**: curated production baseline (neutral `sgconfig.yml` + retained `go-hardcoding.yml` + English-vetted security set) + bounded cross-language security layer; exhaustive 16-language authoring, per-language domain rules, dogfood cleanup, and the gate config-wiring fix are all Out of Scope.
- **Plan-phase self-verification**: plan.md §E checklist complete; requirements in GEARS; Exclusions section carries `### Out of Scope — <topic>` H3 sub-headings with `-` bullets.
- **Ready for**: plan-audit gate → Implementation Kickoff Approval → run-phase (manager-develop).

## §E.2 Run-phase Evidence

### Shipped baseline (final layout)

The curated production baseline shipped to `internal/template/templates/.moai/config/astgrep-rules/` (5 files, exactly — verified by embed-FS walk against the deployment layout):

| File | Purpose |
|------|---------|
| `sgconfig.yml` | English, token-free; `ruleDirs: [go, security]` — only dirs shipping >=1 vetted rule |
| `go/hardcoding.yml` | 2 vetted Go hardcoding rules (ported to `rule:` form) |
| `security/credentials.yml` | Hardcoded-credential family x 4 languages |
| `security/injection.yml` | Command-injection family x 4 languages |
| `security/crypto.yml` | Weak-hash (MD5) x Go |

The prior root `go-hardcoding.yml` was removed: it is silently ignored by config-mode (`sgconfig.yml` present -> the scanner uses config-mode only, and a root file outside `ruleDirs` is not loaded — observed). Its two working rules were ported to `go/hardcoding.yml` in `rule:` form; its third rule (`go-no-raw-getenv`, `os.Getenv("$LITERAL")`) was already a no-op in the deployed form and is dropped (see run-phase finding R-1/R-2 below). Zero functional regression per NFR-AMR-003.

### Run-phase findings (evidence over assumption)

- **R-1 — flat `pattern:` form fails config-mode.** `sg scan --config` (0.40.5) uses the strict rule-config schema requiring a top-level `rule:` field; the dogfood flat-`pattern:` form errors with `missing field 'rule'`. All shipped rules therefore use the `rule:` wrapper. Observed: `Error: Cannot parse rule ... missing field 'rule' at line 8`.
- **R-2 — metavariable-inside-string-literal never matches.** `os.Getenv("$LITERAL")`, `"https://api.$$$REST"`, and `const $NAME = "sk-$$$REST"` all match nothing in 0.40.5 (a metavar embedded in a string literal is not a valid pattern position). The `sec-hardcoded-api-key` (`const $NAME = "sk-$$$REST"`) dogfood rule is broken and was NOT shipped; the credential family was re-authored as `kind: string` + anchored `regex` instead.
- **R-3 — type-conversion-call and metavariable-receiver method-call fail in config-mode.** `template.HTML($X)` (conversion-shaped call) and `$R.SignedString($A)` (metavar-receiver method call) match nothing in 0.40.5. The dogfood `sec-template-injection-html` and `sec-hardcoded-jwt-signing-key` rules were therefore NOT shipped (fail NFR-AMR-001 positive-fixture gate).
- **R-4 — `filepath.Join($BASE, $USER_INPUT)` is noisy.** It matches a legitimate `filepath.Join(home, ".config")` (matches its negative fixture -> EC-3 reject). The dogfood `sec-path-traversal-join-user-input` rule was NOT shipped.
- **R-5 — bare `exec($CMD)` (JS) is noisy** (fires on any user-defined `exec()`); the shipped JS/TS command-injection rule is member-scoped (`child_process.exec` / `cp.exec`) and does not fire on a user-defined `exec`. Verified.
- **R-6 — `os.Getenv($X)` (bare, correct-form) also fails in config-mode** independent of R-2 (a package-selector call with a single metavar arg does not match). `md5.New()` and `exec.Command("sh", "-c", $CMD)` DO match — confirmed the two that ship are reliable.

### Coverage matrix (pattern-family x language) — every cell `sg`-verified pos=1 / neg=0 in config-mode

| Pattern-family | go | python | javascript | typescript |
|----------------|:--:|:------:|:----------:|:----------:|
| hardcoded-credential (CWE-798) | YES | YES | YES | YES |
| command-injection (CWE-78) | YES | YES | YES | YES |
| weak-crypto MD5 (CWE-327) | YES | — | — | — |
| hardcoding: api-url + coverage-threshold | YES | — | — | — |

Legend: YES = shipped + `sg`-verified (positive fixture matches, negative fixture produces zero findings); — = equal-priority future addition (never "unsupported"). Every YES cell was verified with the exact rule YAML shipped, via `sg scan --config <sgconfig.yml>`. No language is marked PRIMARY — the alphabetical `ruleDirs` order and identical per-language pattern-families satisfy §15 / REQ-AMR-004.

**Deferred equal-priority future additions** (not shipped, not deprecated): rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, r, flutter, swift for both security families; per-language domain/idiom rules; XSS / template-injection / JWT / path-traversal families (blocked by the 0.40.5 pattern-expressibility limits R-3/R-4, revisit on a newer `sg`).

### Per-family verification evidence (representative commands + observed output)

```
# hardcoded-credential (go), config-mode:
$ sg scan --config <ship>/sgconfig.yml --json cred_pos.go   -> 1 finding  [sec-hardcoded-credential]
$ sg scan --config <ship>/sgconfig.yml --json cred_neg.go   -> 0 findings   ("hello sky" not flagged — provider-prefix anchored)
# command-injection (go): exec.Command("sh","-c",$CMD)
$ pos (userInput arg) -> 1;  neg (exec.Command("ls","-la")) -> 0
# weak-crypto (go): md5.New()
$ pos -> 1;  neg (sha256.New()) -> 0
# mixed multi-language E2E (config-mode):
$ mixed.go -> 5 [go-no-duplicate-coverage-threshold, go-no-hardcoded-api-url, sec-command-injection-shell, sec-hardcoded-credential, sec-weak-hash-md5]
$ mixed.py -> 2 [sec-command-injection-shell, sec-hardcoded-credential]
$ mixed.ts -> 2 [sec-command-injection-exec, sec-hardcoded-credential]
$ clean.go / clean.py / clean.ts -> 0 / 0 / 0
```

## §E.3 Run-phase Audit-Ready Signal

### AC matrix (AC-AMR-001..011)

| AC | Severity | Status | Verification | Observed |
|----|----------|--------|--------------|----------|
| AC-AMR-001 | MUST | PASS | every shipped rule has pos+neg fixture; no scaffold | 9 rules all pos=1/neg=0; `grep scaffold` -> 0 |
| AC-AMR-002 | MUST | PASS | Hangul/CJK grep on shipped tree | 0 matches |
| AC-AMR-003 | MUST | PASS | SPEC-ID/REQ/AC grep + leak-test + neutrality CI | tokens 0; `TestTemplateNoInternalContentLeak` ok; `TestTemplateNeutralityAudit` ok |
| AC-AMR-004 | MUST | PASS | coverage matrix identical per language; no PRIMARY | matrix above; alphabetical ruleDirs |
| AC-AMR-005 | MUST | PASS | `sg scan --config <sgconfig>` no parse/missing-ruleDir error | mixed.{go,py,ts} scans produced valid JSON, no config error |
| AC-AMR-006 | MUST | PASS | embed-FS walk of deployment layout | exactly 5 files embedded under `.moai/config/astgrep-rules` |
| AC-AMR-007 | MUST | PASS | `moai ast-grep vuln.go` -> vetted finding only | 1 finding `sec-weak-hash-md5`, exit 1; clean.go -> 0, exit 0 |
| AC-AMR-008 | MUST | PASS | coverage matrix recorded, equal treatment | §E.2 matrix; no PRIMARY marker |
| AC-AMR-009 | MUST | PASS | each shipped rule pos-match + neg-no-match | 9/9 verified; 0 `.gitkeep` embedded |
| AC-AMR-010 | MUST | PASS | go-hardcoding behavior preserved; clean -> 0 | 2 working rules ported; broken getenv rule (already no-op) dropped — no regression |
| AC-AMR-011 | SHOULD | PASS | dogfood tree unchanged | `git status --porcelain .moai/config/astgrep-rules/` -> 0 lines |

### GT-4 verified gate blast-radius statement

The deployed curated baseline is consumed by two independent paths, verified this run:

1. **`moai ast-grep` CLI (`internal/cli/astgrep.go`) — unconditionally active.** With `sg` in PATH (ast-grep 0.40.5, present) the CLI runs `Scanner.Scan` with `WarnOnlyMode: false` and `os.Exit(1)` on error-severity findings. Verified: `moai ast-grep vuln.go` -> 1 error-severity finding (`sec-weak-hash-md5`), exit 1; `moai ast-grep clean.go` -> 0 findings, exit 0. This is the always-on consumer — the deployed ruleset's quality matters regardless of the commit-gate default, because error-severity rules block CLI usage.
2. **Commit-time quality gate (`pre_tool.go` -> `QualityGate.Run` -> `RunAstGrepGateV2`) — effectively OFF via config path.** The Go struct reads a top-level `gate:` key (`config/types.go`); the template's `ast_grep_gate:` block is nested under `constitution:` in `quality.yaml.tmpl`, so a config-loaded session resolves `config.Gate.AstGrepGate.Enabled` to the zero value `false`. This key-path mismatch (`constitution.ast_grep_gate` vs `gate.ast_grep_gate`) is a separate config-wiring defect recorded Out of Scope in spec.md; this SPEC ships ruleset content, not gate activation.

Net: the baseline's quality is load-bearing through the CLI path even though the commit-gate path is off. The curated set is `error`-severity for the security rules (credential x4, command-injection x4, weak-crypto MD5 x1) and `warning`-severity for the 2 Go hardcoding rules. No demonstrative/empty rule can produce a CLI-blocking finding because none ship.

### §D.3 Quality Gate outputs

- `go test ./internal/astgrep/... ./internal/template/...` -> **ok / ok** (both packages PASS)
- `go test ./internal/template/... -run 'TestTemplateNeutralityAudit'` -> **ok** (CI neutrality guard)
- `go test ./internal/template/... -run 'TestTemplateNoInternalContentLeak'` -> **ok** (CI leak guard, narrow)
- `go build ./...` (host darwin) -> exit 0; `GOOS=windows GOARCH=amd64 go build` -> exit 0 (cross-platform)
- Every shipped rule paired positive + negative fixture passes NFR-AMR-001 (9/9)
- Strict-mode leak test (`MOAI_TEMPLATE_LEAK_STRICT=1`) fails on pre-existing `skills/*/workflows/*.md` dates — NOT introduced by this SPEC; the astgrep-rules tree has zero strict-mode violations (verified). Strict mode is opt-in and NOT the CI gate.

## §E.4 Sync-phase Audit-Ready Signal

- 3-phase close (Tier M, orchestrator-direct sync): status `draft → completed`, era `V3R6` (frontmatter H-override), updated `2026-07-02`.
- Run-phase은 manager-develop이 L1 격리 worktree(`agent-a5e3ed044af847d4d`)에서 수행 → 오케스트레이터가 5개 큐레이션 파일(`sgconfig.yml` + `go/hardcoding.yml` + `security/{credentials,injection,crypto}.yml`)을 main 트리로 reconcile + 구 `go-hardcoding.yml` 삭제 적용(L1 격리 이탈 reconcile 패턴, worktree 정리 예정).
- 오케스트레이터 독립 재검증(main 트리): neutrality Hangul/CJK 0 + SPEC/REQ/AC 토큰 0; `sg scan --config sgconfig.yml` 파싱 정상(credential 룰 영어 메시지 발화); `go test ./internal/astgrep/... ./internal/template/...` GREEN(embed/leak/neutrality 회귀 없음); `go build ./...` exit 0.
- CHANGELOG `### Added` 엔트리 추가.
- sync_commit_sha: 82b49d82e94663cc662f2da5c81fc4c249179dc0
- MX Tag: Tier M — MX는 sync sub-step(3-phase close, 별도 Mx 커밋 없음).
