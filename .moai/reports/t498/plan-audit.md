# SPEC Review Report: SPEC-CODEX-MIRROR-DOCTOR-001

Iteration: 1/2 (Tier M ceiling)
Verdict: **FAIL**
Overall Score: **0.775** (Tier M PASS threshold 0.80)

감사 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t498`, 브랜치 `WT-codex-mirror-doctor`, HEAD `26b77973e`.
저자 추론 맥락은 M1 Context Isolation 에 따라 무시했다 — 판정 근거는 SPEC 3종 아티팩트와 이 트리에서 직접 잰 측정뿐이다.
인용된 모든 `file:line` 은 이 트리에서 다시 읽어 확인했고, 인용을 신뢰하지 않았다.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -Eo '^REQ-CMD-[0-9]+' spec.md | sort | uniq -c` 결과 REQ-CMD-001…011 각 1회, 결번·중복 없음, 3자리 zero-padding 일관. Tier M REQ 상한 16 이내(11 사용).
- **[PASS] MP-2 GEARS format compliance (requirement layer)** — 판정 대상은 `spec.md` 의 `REQ-XXX` 요구 계층이다. `acceptance.md` 의 Given-When-Then 은 검증 계층의 정상 형식이므로 여기서 감점하지 않았다(Group 4 소관). 11건 전부 GEARS 패턴에 대응한다: Ubiquitous `REQ-CMD-001/009/011` (spec.md:47, :88, :99), Unwanted `REQ-CMD-002` ("shall not create, repair, remove…", spec.md:52), State-driven `REQ-CMD-003/006/007/008` (spec.md:56, :73, :79, :83), Where/When 복합 `REQ-CMD-004/005` (spec.md:62, :67), Event-driven `REQ-CMD-010` (spec.md:93). 폐기 대상 `If/then` 형태 0건, `should`/`may` 등 연성 조동사 0건 — 둘 다 grep 무출력으로 확인.
- **[PASS] MP-3 YAML frontmatter validity** — 정본 12필드 전부 존재(spec.md:2-13): `id`,`title`,`version`,`status`,`created`,`updated`,`author`,`priority`,`phase`,`module`,`lifecycle`,`tags` + 선택 필드 `tier: M`. 거부 대상 snake_case 별칭(`created_at`/`updated_at`/`labels`/`spec_id`) 0건. `phase: "v3.2.0 target"` 는 금지된 생명주기 토큰(plan/run/sync/mx)이 아니다.
  기계 검증: `moai spec lint .moai/specs/SPEC-CODEX-MIRROR-DOCTOR-001/spec.md` → `✓ No findings — all SPEC documents are valid`, `--json` → `[]`.
  **이 초록은 공허하지 않다** — 대조군으로 `module:` 한 줄을 지운 사본을 같은 명령에 넣어 RED 를 확인했다(아래 verbatim). `plan.md` / `acceptance.md` 는 `status:` 를 갖지 않아 artifact statelessness 규약도 만족한다.
- **[N/A] MP-4 Section 22 language neutrality** — 단일 언어(Go) 범위의 SPEC이다. `internal/cli/doctor_codex.go` 한 파일의 진단 로직을 다루며 16개 프로그래밍 언어 도구 중립성 축을 건드리지 않는다. N/A 는 자동 통과.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — `grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md | sort -u` → `SPEC-CODEX-MIRROR-DOCTOR-001` 자기 참조 1건뿐. 외부 SPEC 참조가 없으므로 retired/superseded/archived 대조 대상이 없고 BLOCKING 도 없다. 검증 동사는 실행 가능했다(N/A 아님).
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall'` → `spec.md:0`, `plan.md:0`, `acceptance.md:0`. 본문에 `syscall` 이 등장하지 않으므로 D8-4 에 따라 자동 PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' plan.md acceptance.md spec.md progress.md` → 무출력. `research.md` 는 Tier M 아티팩트 집합에 없어 부재가 정상이며, `plan.md` 가 존재하므로 N/A 가 아닌 실측 PASS.

must-pass 방화벽은 전부 통과했다. **이 FAIL 은 방화벽이 아니라 루브릭 점수가 임계 미만이어서 나온 것이다.**

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 | REQ-CMD-006/007(spec.md:73-81)에 배선 전제가 없는데 plan.md:63 은 `wired` 분기로만 한정 — 해석 여지가 남는다. spec.md:170 이 spec.md:113 자신의 조건부 서술과 모순(D3). 나머지 요구는 단일 해석이고 근거 줄이 전부 붙어 있다. |
| Completeness | 1.00 | 1.0 | HISTORY(:19), Context(:25), Requirements(:45), 분류 근거(:105), 지시문(:116), Exclusions(:132), Constraints(:163) 모두 존재. `### Out of Scope — <topic>` H3 4개(spec.md:134,141,148,155) 각각 구체 `-` 불릿 보유. frontmatter 12필드 완비. |
| Testability | 0.65 | 0.50-0.75 사이 | AC 12건에 weasel word 0건(grep 무출력)이고 대부분 이진 판정이 가능하다. 그러나 AC-CMD-008 은 **작성된 그대로 실행하면 실패한다**(D1, 실측). AC 8건은 존재하지 않는 테스트명을 판정자로 지목하면서 sweep 개수를 못박지 않아 공허한 초록에 노출돼 있다(D2). 루브릭 0.75("한 건이 minor interpretation 필요")보다 나쁘고 0.50("다수가 판단 개입 필요")보다는 낫다. |
| Traceability | 0.70 | 0.50-0.75 사이 | REQ-CMD-002→AC-010, 003→AC-001, 004→AC-003, 005→AC-004, 006→AC-005, 007→AC-006, 008→AC-007, 010→AC-008 은 깨끗하다. 그러나 REQ-CMD-001 은 대응 AC 가 없고(D6), REQ-CMD-011 은 acceptance.md:11-14 전문(preamble)으로만 덮이며, REQ-CMD-009 의 tail-drop 절은 어느 AC 도 검증하지 않는다(D5). 또한 어떤 AC 도 자기 REQ-ID 를 명시하지 않아 매핑이 전부 간접이다. |

산술: (0.75 + 1.00 + 0.65 + 0.70) / 4 = 3.10 / 4 = **0.775**. Tier M 임계 0.80 미달.

---

## Mechanical tooling output (verbatim)

### `moai spec lint` — 대상 SPEC

```
$ ~/go/bin/moai spec lint .moai/specs/SPEC-CODEX-MIRROR-DOCTOR-001/spec.md
✓ No findings — all SPEC documents are valid
```

```
$ ~/go/bin/moai spec lint .moai/specs/SPEC-CODEX-MIRROR-DOCTOR-001/spec.md --json
[]
```

```
$ ~/go/bin/moai spec lint .moai/specs/SPEC-CODEX-MIRROR-DOCTOR-001/spec.md --strict
✓ No findings — all SPEC documents are valid
```

### 위 초록의 비공허성 대조군 (mutation control)

`module:` 한 줄만 제거한 사본을 같은 명령에 통과시켰다. 린터가 이 파일 모양을 실제로 훑고 있음이 확인된다.

```
$ ~/go/bin/moai spec lint <scratch>/mutant.md
SEVERITY  CODE                FILE          LINE  MESSAGE
--------  ----                ----          ----  -------
ERROR     FrontmatterInvalid  .../mutant.md 1     Frontmatter required field missing: module

1 error(s), 0 warning(s)
```

### `mcp__moai__spec_audit` (project_root = 이 워크트리)

```json
{"audited_at":"2026-09-07T03:44:01.043144Z","total_specs":1,"grandfathered":0,"modern_era_clean":1,"drift_findings":[{"spec_id":"SPEC-CODEX-MIRROR-DOCTOR-001","era":"V3R6","finding_type":"EraAutoDetected","severity":"INFO","details":{"heuristic_matched":"H-5 (modern phase or created date)"}}]}
```

### `mcp__moai__spec_drift` (project_root = 이 워크트리)

루트 귀속과 sweep 규모를 먼저 확인한 뒤 해당 SPEC 행만 추출했다(전체 출력 93,353자).

```
$ jq -r '._root, .total_specs, .modern_clean' <tool-result>
{ "dir": "/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t498", "source": "param" }
776
511
```

```json
{"spec_id":"SPEC-CODEX-MIRROR-DOCTOR-001","era":"V3R6","finding_type":"EraAutoDetected","severity":"INFO","details":{"heuristic_matched":"H-5 (modern phase or created date)"}}
```

기계 도구가 낸 유일한 항목은 INFO 수준 `EraAutoDetected` 하나이며, 이는 결함이 아니라 `era:` 명시 override 가 없다는 알림이다. drift 0, ERROR 0, WARNING 0.

> 주: MCP 서버 빌드는 `v3.2.0-rc.0 / e79c010b8` 로 이 트리보다 앞선 커밋에서 기동돼 있다. 위 두 호출은 `project_root` 를 이 워크트리로 넘겼고 응답의 `_root.source` 가 `param` 임을 확인했으므로, 읽힌 트리는 이 워크트리가 맞다.

### 지시문 플래그 실재 확인 (`moai update --help`)

```
--force           Force update: bypass version-match skip, force backup+merge, ...
--templates-only  Skip binary update, sync templates only
--yes             Auto-Confirm all prompts (CI/CD mode)
```

3개 플래그 모두 현재 바이너리에 존재한다. 폭도 통과한다 — 지시문 자체는 46 runes, 실제 요약문에 실었을 때의 상한 후보는 104 runes(`codex wiring active but .agents/skills mirror is absent — run moai update --templates-only --force --yes`)로 `codexMessageWidthCeiling` 113 이내다.

### 베이스라인 게이트 (AC-CMD-011 / AC-CMD-012 도달 가능성)

```
$ golangci-lint run ./internal/cli/... --timeout=5m   → "0 issues.", EXIT=0
$ go vet ./internal/cli/...                           → EXIT=0
$ GOOS=windows GOARCH=amd64 go build ./...            → EXIT=0
```

세 게이트 모두 착수 시점에 초록이다. 즉 AC-CMD-011/012 는 "도착 시점부터 영원히 RED"인 impossible 유형이 아니다.

---

## 지목된 위험 축별 판정

### ① 검증 불가능한 AC — **결함 있음** (D1, D2)

명령·플래그·경로·심볼 인용은 전수로 트리에서 다시 읽었고, 아래는 전부 실재가 확인됐다: `checkCodexWiring`(doctor_codex.go:86), `codexMessageWidthCeiling`(:69), `codexWiringLookPath`(:38), `codexUserHomeDir`(테스트 seam, doctor_codex_test.go:76-81), `stubMoaiLookup`(:38) / `stubCodexLookup`(:53) / `stubCodexHome`(:76) / `wireProjectForDoctor`(:26), `uikit.CheckOK` / `uikit.CheckWarn`, `internal/cli/testdata/doctor-nocolor.golden`(존재, :43 에 Codex Wiring 행), `mirrorSkillsRelDir`(skill_mirror.go:52), `mirrorLinkTarget`(:151), `MirrorModeCopy`/`MirrorModeSkipped`(:38-42), copy-fallback 경고문(:249), `update.go` syncSkipped 조기 반환(:513-523 — SPEC 인용 509-523 과 일치).

**AC-CMD-009 가 "이미 파일에 있다"고 지목한 `TestCheckCodexWiring_RenderedPanelStaysInBand` 는 실제로 존재한다**(doctor_codex_test.go:437). 이 인용은 거짓이 아니다.
**AC-CMD-002 의 셀렉터 `TestDoctor.*Golden|TestDoctor_` 도 실제로 매칭한다** — `TestDoctorGolden_{Light,Dark,NoColor}`(doctor_golden_test.go:161,188,215)와 `TestDoctor_CheckCount` 등. 공허한 셀렉터가 아니다.

문제는 다른 곳이다 → D1, D2.

### ② un-nagging 불변식 — **주장 성립, 그러나 회귀 고정은 부분적**

REQ-CMD-003 이 "by construction" 이라고 말하는 근거를 함수를 읽어 확인했다. 조기 반환은 `doctor_codex.go:96-103` 에 정확히 있고, SPEC 이 인용한 줄 범위와 한 줄도 어긋나지 않는다:

```go
96  if !wired && !codexInstalled {
        // ... "This is the un-nagging invariant — it must survive every addition below."
100     check.Status = uikit.CheckOK
101     check.Message = "not wired (claude-only project) — skipped"
102     return check
103 }
```

plan.md:61 이 지시하는 배치("Call the inspector **after** the `!wired && !codexInstalled` early return")는 이 반환보다 뒤이므로, 주장대로 구성상 침묵이 유지된다. **이 축은 결함이 아니다.**

회귀 고정 여부도 실질적으로 확인했다: AC-CMD-001 이 `check.Message`/`check.Detail` 에 `.agents`·`mirror` 부재를 직접 단언하므로 명목상 회귀를 고정한다. 다만 그 판정 테스트가 아직 없어 D2 의 공허-초록 노출을 그대로 공유한다. 참고로 기존 `TestCheckCodexWiring_ClaudeOnlyMachineStaysSilent`(:547)가 이미 인접 표면을 덮고 있는데, SPEC/plan 어디에도 이 선행 테스트와의 관계가 적혀 있지 않다(신규 테스트가 그것을 대체하는지 보완하는지 불명 — D6 에 포함).

### ③ root-cause.md 로부터 물려받은 주장 — **한 건 과대 진술** (D3)

- spec.md:33-36 의 `moai update --yes` 미복구 주장은 root-cause Claim 2 표와 정확히 일치하며, 원인을 `update.go` 조기 반환이라는 **소스에서 읽은 기전**에 두므로 darwin 단일 프로브에서 전 플랫폼으로 일반화한 것이 아니다. 정당하다.
- spec.md:113 은 Windows copy-fallback 미실행 사실을 Gaps 로 정확히 인용하고, 그것을 **격상하지 않을 근거**로 보수적으로 쓴다. 모범적이다.
- spec.md:138-139 는 "이 저장소에 미러가 있었던 적이 있는가"라는 미결 간극을 확정된 것처럼 다루지 않는다 — 처방(deploy)만 말한다. 정당하다.
- **그러나 spec.md:170 은 다르다** → D3.

### ④ 범위 침범 — **없음**

읽기 전용 경계는 세 겹으로 지켜진다: REQ-CMD-002(spec.md:52)가 생성·수리·삭제·교체·변경을 전부 금지하고, `### Out of Scope — repairing a mirror from doctor`(spec.md:141)와 `### Out of Scope — the update short-circuit`(spec.md:148)이 각각 doctor 측 수리와 `runTemplateSyncWithProgress`/`update.go:509-523` 변경을 명시적으로 배제하며, plan.md:90 이 `os.Symlink`/`os.Remove`/`MkdirAll` 을 anti-pattern 으로 못박는다. 무엇보다 AC-CMD-010 이 호출 전후 `path + mode + symlink-target` 목록의 바이트 동일성으로 이를 **기계적으로** 고정한다 — 문장이 아니라 측정으로 막는 좋은 설계다. 미러를 만들거나 고치거나 지우는 요구는 11건 중 하나도 없다. **이 축은 결함이 아니다.**

### ⑤ 지시문 명령 — **정확함**

위 tooling 절 참조. 3개 플래그 전부 실재하고, 폭 상한도 실제 문자열로 계산해 통과를 확인했다. 루틴형 `moai update --yes` 를 지시문에서 의도적으로 배제한 판단(spec.md:126-127)도 측정과 일치한다. **이 축은 결함이 아니다.**

---

## Defects Found (structured defect-list)

D1. **AC-CMD-008 은 작성된 그대로 실행하면 실패한다** — `.moai/specs/SPEC-CODEX-MIRROR-DOCTOR-001/acceptance.md:85-93` — AC 는 `.agents/skills` 를 mode `0o000` 으로 두라고만 지시하고 모드 복원을 요구하지 않는다. `t.TempDir()` 의 자동 정리는 그 디렉터리를 열어야 하므로 실패하고, Go 는 이를 테스트 실패로 보고한다. 가설이 아니라 실측이다 — 이 트리의 Go 로 AC 문구를 그대로 옮긴 프로브를 돌렸다:
   ```
   === RUN   TestChmodZeroWithEntries
       ReadDir err (expected non-nil): open .../.agents/skills: permission denied
       testing.go:1464: TempDir RemoveAll cleanup: openfdat .../.agents/skills: permission denied
   --- FAIL: TestChmodZeroWithEntries (0.00s)
   === RUN   TestChmodZeroEmpty
   --- PASS: TestChmodZeroEmpty (0.00s)
   ```
   항목이 든 디렉터리에서는 FAIL, 빈 디렉터리에서는 PASS 다. AC 는 둘 중 어느 쪽인지 규정하지 않아 모호하며, 의미 있는 쪽(항목이 있는 실제 미러)이 곧 실패하는 쪽이다. 덧붙여 이 파일이 이미 확립한 indeterminate 픽스처 관용구는 chmod 가 아니라 **심링크 루프**다(`internal/cli/doctor_codex_test.go:335-352`, `TestCheckCodexWiring_IndeterminateStatNotMissing`) — 그 관용구는 정리 위험이 없고 이식성 판정(`t.Skipf`)까지 갖췄다. — Severity: **major** — Class: **blocking** — Required fix: AC-CMD-008 을 이웃 함수의 관용구로 다시 쓴다. 심링크 루프를 쓰거나, chmod 를 유지한다면 (a) 디렉터리에 항목이 있는지 명시하고 (b) `t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })` 로 모드 복원을 AC 본문에 못박는다.

D2. **AC 8건이 아직 존재하지 않는 테스트명을 판정자로 지목하면서 sweep 개수를 못박지 않아, 구현 전에도 초록으로 읽힌다** — `acceptance.md:24, 43, 53, 63, 73, 83, 93, 112` — `TestCheckCodexWiring_ClaudeOnlyMachineNoMirrorRow`, `_MirrorAbsentAdvisesRedeploy`, `_DanglingMirrorEntriesCounted`, `_CopyModeDetailOnly`, `_UnmirroredSkillsDetailOnly`, `_UnwiredNoMirrorNag`, `_MirrorUnreadableIndeterminate`, `_MirrorCheckIsReadOnly` 및 AC-CMD-009 의 `_MirrorSummaryWidth` 는 현재 트리에 없다(`grep 'func TestCheckCodexWiring' internal/cli/doctor_codex_test.go` 로 전수 확인 — 19개 기존 테스트 중 어느 것도 이 이름들이 아니다). `go test -run <없는이름>` 은 exit 0 에 `ok ... [no tests to run]` 를 찍으므로, 이 AC 들은 구현 전에도 "통과"한다. `.claude/rules/moai/development/verification-completeness.md` §1.1 이 이름 붙인 결함 형태 그대로이며(“A pass whose swept set is empty asserts nothing”), 같은 파일 §2 가 요구하는 RED-now 셀도 어느 AC 에도 없다. 새 이름이라 미존재 자체는 정상이지만, **판정 절차가 그 미존재를 초록과 구분하지 못하는 것**이 결함이다. — Severity: **major** — Class: **blocking** — Required fix: 각 AC 의 `Decided by` 에 sweep 개수 확인을 함께 못박는다(예: `go test ./internal/cli/ -run <name> -v` 출력에 `--- PASS: <name>` 행이 정확히 1건 존재할 것, 또는 `-run` 결과에 `[no tests to run]` 이 없을 것). AC-CMD-001 처럼 회귀를 고정하는 항목에는 §2 의 RED-now 셀(명령·verbatim 출력·exit code·tree SHA)을 붙인다.

D3. **spec.md §6 이 측정되지 않은 조건을 사실로 진술하며, 같은 문서 §3 과 모순된다** — `.moai/specs/SPEC-CODEX-MIRROR-DOCTOR-001/spec.md:169-170` — "the check must build and behave on linux/darwin/windows, **where a copy-mode mirror is the normal case**". root-cause.md Gaps 는 "Windows copy-fallback behaviour was not exercised" 로 명시하고, 생산자 코드 `internal/template/skill_mirror.go:234-236` 은 copy 를 "Windows **without the privilege**, 이색 파일시스템, 샌드박스" 조건에서의 fallback 으로 규정한다 — 권한이 있는 Windows 는 심링크를 만든다. 같은 SPEC 의 spec.md:113 은 이 조건을 정확히 "wherever symlink creation is unavailable (Windows without the privilege…)" 로 달아 두었으므로, §6 은 자기 문서의 더 정확한 서술과 어긋난다. 어떤 AC 도 이 문장에 의존하지 않아 실행 위험은 낮지만, 미측정 전제를 사실로 적은 것은 `verification-claim-integrity.md` §1 위반이다. — Severity: **minor** — Class: **blocking** — Required fix: spec.md:170 을 §3 의 조건부 표현에 맞춘다 — 예: "…on linux/darwin/windows; where symlink creation is unavailable a copy-mode mirror is the expected materialization (frequency unmeasured — root-cause.md Gaps)".

D4. **REQ-CMD-006/007 은 배선 전제 없이 쓰였는데 plan 은 `wired` 분기로만 한정한다** — `spec.md:73-81` ↔ `plan.md:61-64` — 두 요구는 "While the mirror directory exists" 로만 조건 지어져 있어 `!wired && codexInstalled` 분기에도 적용되는 것으로 읽힌다. 반면 plan.md:63 은 "Call it only in the `wired` branch (REQ-CMD-008)" 로 검사기 호출을 배선된 분기에 한정한다. REQ-CMD-008 이 금지하는 것은 mirror **finding** 이고 006/007 이 요구하는 것은 Detail **카운트**이므로 정면 충돌은 아니지만, 미배선 + 미러 존재 상태에서 Detail 카운트를 내야 하는지가 어느 문서에서도 결정되지 않는다. AC-CMD-007 은 미배선 + 미러 **부재** 만 덮어 이 조합을 검증하지 않는다. — Severity: **minor** — Class: **blocking** — Required fix: REQ-CMD-006/007 앞에 `Where the project declares Codex wiring,` 전제를 추가해 plan.md:63 과 일치시키거나, 반대로 plan.md 를 요구에 맞추고 해당 조합을 덮는 AC 를 1건 추가한다.

D5. **REQ-CMD-009 의 tail-drop 절을 검증하는 AC 가 없다** — `spec.md:88-91` — 이 요구는 두 가지를 말한다: (a) 요약문이 단독으로 113 runes 이내일 것, (b) `problems` 에 append 되어 기존 `joinCodexSummaries` tail-drop truncation 에 **변경 없이** 참여할 것. AC-CMD-009 는 (a) 만 덮는다. (b) 는 plan.md:35-38(§C.2)이 스스로 "공유 자원이라 새 요약이 꼬리에서 잘릴 수 있다"고 위험으로 지목한 절인데, 그 위험을 확인하는 판정자가 없다. `joinCodexSummaries`(doctor_codex.go:228-248)는 선두 요약 1건이 단독으로 상한을 넘으면 통째로 내보내는 예외까지 갖고 있어, 상호작용이 자명하지 않다. — Severity: **minor** — Class: **blocking** — Required fix: 미러 finding 이 기존 finding 과 함께 올라간 상태에서 Message 가 상한을 넘지 않고 overflow marker 가 정확히 드러나는지 확인하는 AC 를 1건 추가한다.

D6. **REQ-CMD-001 에 대응 AC 가 없고, REQ-CMD-011 은 전문(preamble)으로만 덮이며, 어떤 AC 도 자기 REQ-ID 를 명시하지 않는다** — `spec.md:47-50`, `spec.md:99-103`, `acceptance.md:11-14` — REQ-CMD-001(호스트 함수와 2-register finding 형태)은 AC 12건 중 어느 것도 직접 판정하지 않는다. REQ-CMD-011(테스트 seam 사용)은 acceptance.md 전문에 서술될 뿐 판정 명령이 붙은 AC 가 없다. 또한 AC 본문에 `REQ-CMD-XXX` 참조가 하나도 없어(전문의 REQ-CMD-011 1회 제외) 매핑이 전부 spec.md §3 표를 통한 간접 추론이다. 아울러 AC-CMD-001 이 기존 `TestCheckCodexWiring_ClaudeOnlyMachineStaysSilent`(doctor_codex_test.go:547)와 어떤 관계인지(대체/보완) 어디에도 적혀 있지 않다. — Severity: **minor** — Class: **blocking** — Required fix: 각 AC 표제 아래 대응 REQ-ID 를 1줄로 명시하고, REQ-CMD-001 과 REQ-CMD-011 을 덮는 AC 를 추가하거나 두 요구를 판정 가능한 형태로 재작성한다. AC-CMD-001 이 기존 테스트를 대체하지 않음을 명시한다.

D7. **REQ-CMD-001/011 이 행동이 아니라 구현 배치와 심볼명을 요구한다** — `spec.md:47-50`, `spec.md:99-103` — "shall live inside `checkCodexWiring` (`internal/cli/doctor_codex.go:86`)", "shall be verifiable through the existing `codexWiringLookPath` and `codexUserHomeDir` test seams" 는 WHAT/WHY 가 아니라 HOW 다. GEARS 형식 자체는 충족하므로 MP-2 는 통과하지만(요구 계층 판정), 파일이 리팩터될 때 요구가 먼저 낡는다. 다만 이 SPEC 은 기존 함수 안에 진단을 접어 넣는 것 자체가 un-nagging 불변식을 지키는 수단이므로 배치를 못박을 **정당한 이유가 있다** — 그래서 blocking 이 아니라 optional 로 분류한다. — Severity: **minor** — Class: **optional** — Required fix: (선택) 요구를 결과로 다시 쓰고 배치 근거는 §6 Constraints 로 옮긴다 — 예: "The mirror observation shall be reported through the existing 'Codex Wiring' row rather than a new `DiagnosticCheck` row."

D8. **doctor 가 `.agents/skills` 경로 리터럴을 복제해야 하는데, 생산자와의 결속을 못박는 요구가 없다** — `spec.md:65` ↔ `internal/template/skill_mirror.go:52`, `spec.md:70-71` ↔ `skill_mirror.go:151` — REQ-CMD-004 는 `mirrorSkillsRelDir` 를, REQ-CMD-005 는 `mirrorLinkTarget` 를 "Source read" 로 인용하지만, 둘 다 package `template` 의 **미노출(unexported)** 식별자다. package `cli` 의 doctor 는 이를 참조할 수 없으므로 plan.md:52-54 대로 경로와 링크 형태를 하드코딩하게 된다. 생산자가 미러 경로나 링크 본문을 바꾸면 doctor 는 조용히 틀린 곳을 보게 되고, 그 상태에서 "미러 없음"이라는 **잘못된 finding** 을 낸다 — 읽기 전용 검사에서 나올 수 있는 가장 나쁜 오류다. — Severity: **minor** — Class: **optional** — Required fix: (선택) `mirrorSkillsRelDir` / `mirrorLinkTarget` 를 export 하여 doctor 가 참조하게 하거나, 최소한 두 리터럴이 생산자와 일치함을 고정하는 AC 를 1건 둔다. 이 SPEC 범위를 넘는다고 판단되면 후속 카드로 남긴다.

D9. **AC-CMD-002 의 두 번째 판정 절이 tree SHA 에 고정되지 않아 사실상 아무것도 단언하지 않는다** — `acceptance.md:32-33` — `git diff --name-only -- internal/cli/testdata/` 가 무출력이면 통과라고 하는데, golden 테스트는 실패할 뿐 파일을 자동 갱신하지 않으므로(갱신 플래그를 쓰지 않는 한) 이 절은 거의 항상 무출력이다. 또 이동하는 참조(작업 트리) 기준이라 `verification-completeness.md` §4 의 고정 요구를 만족하지 않는다. 첫 번째 절(golden 테스트 실행)은 유효하며 셀렉터도 실제로 매칭함을 확인했으므로, 이 AC 가 공허하지는 않다 — 두 번째 절만 장식이다. — Severity: **minor** — Class: **optional** — Required fix: (선택) 두 번째 절을 삭제하거나, 비교 기준을 base SHA(`ace1c5440`)로 고정한다.

D10. **Definition of Done 의 변경 파일 목록이 `progress.md` 를 빠뜨린다** — `acceptance.md:132-134` — "`git diff --name-only` shows only `internal/cli/doctor_codex.go` and `internal/cli/doctor_codex_test.go` plus this SPEC directory" 라고 하는데, run-phase 는 `progress.md` §E.2/§E.3 를 반드시 쓴다. "plus this SPEC directory" 가 그것을 포함한다고 읽을 수 있어 치명적이지는 않으나, 문자 그대로의 판정에서는 흔들린다. — Severity: **minor** — Class: **optional** — Required fix: (선택) "plus `.moai/specs/SPEC-CODEX-MIRROR-DOCTOR-001/` (progress.md 갱신 포함)" 로 명시한다.

집계: 총 10건 — blocking 6건(major 2, minor 4), optional 4건(전부 minor). critical 0건(must-pass 위반 없음).

---

## Regression Check

해당 없음 — iteration 1. `.moai/reports/t498/` 에 선행 plan-audit 보고서가 없음을 확인했다(`root-cause.md` 만 존재).

---

## Recommendation

이 SPEC 은 근거의 질이 평균을 크게 웃돈다. 인용된 `file:line` 을 전수 재확인했는데 어긋난 것이 하나도 없었고, un-nagging 불변식의 "by construction" 주장은 함수를 읽어 성립함을 확인했으며, 읽기 전용 경계는 문장이 아니라 AC-CMD-010 의 바이트 동일성 측정으로 막혀 있다. §3 의 finding/detail 분류는 미측정 사실(Windows copy-fallback, 미러 분모)을 **격상하지 않을 근거**로 보수적으로 사용한다 — 드문 규율이다. FAIL 은 이 골격을 부정하지 않는다.

임계 미달의 실질은 **검증 계층 6건(D1-D6)** 이며, 전부 acceptance.md 국소 수정으로 닫힌다. 순서대로:

1. **D1 을 먼저 고친다.** AC-CMD-008 을 `doctor_codex_test.go:335-352` 의 심링크-루프 관용구로 다시 쓴다. chmod 를 고집한다면 항목 유무를 명시하고 `t.Cleanup` 모드 복원을 AC 본문에 넣는다. 이것만이 "실행하면 실패한다"가 실측으로 확정된 항목이다.
2. **D2 를 닫는다.** 9개 AC 의 `Decided by` 에 sweep 개수 확인을 더한다(`--- PASS: <name>` 1행 존재, 또는 `[no tests to run]` 부재). AC-CMD-001 에는 RED-now 셀을 붙인다 — 이 AC 가 고정하려는 회귀가 이 변경의 최대 위험이므로, 그 판정자만은 공허할 수 없어야 한다.
3. **D5, D6, D4 로 추적성과 범위를 닫는다.** tail-drop 상호작용 AC 1건 추가(D5), REQ-CMD-006/007 에 배선 전제 추가(D4), 각 AC 에 REQ-ID 1줄 명시 + REQ-CMD-001/011 커버(D6).
4. **D3 을 한 줄 고친다.** spec.md:170 을 spec.md:113 의 조건부 표현에 맞춘다.

D7-D10 은 optional 이다. **이것들을 이유로 재수정을 돌리지 말 것을 권한다** — D8 은 실제 위험(경로 드리프트 시 잘못된 finding)을 담고 있으나 export 변경은 이 SPEC 이 스스로 그은 읽기 전용·2파일 범위를 넘고, 나머지 셋은 문구 취향에 가깝다. D8 은 후속 카드 후보로 기록하는 편이 낫다.

위 1-4 를 반영하면 Testability 와 Traceability 가 0.90 대로 올라가 총점은 0.85 이상이 되어 Tier M 임계를 여유 있게 넘는다. iteration 2 는 이 열거된 결함 델타에만 범위를 두어 재감사한다(Tier M ceiling 2 — 이번이 1회차이므로 1회가 남는다).

---

## Gaps (이 감사가 관측하지 **않은** 것)

- `go test ./internal/cli/... -count=1` 베이스라인은 **측정하지 않았다.** AC-CMD-011 의 세 다리 중 `golangci-lint`(EXIT=0)와 `go vet`(EXIT=0)만 실측했다. `internal/cli` 전체 테스트는 CLAUDE.local.md §4 가 경고하는 고부하 대상이고 다른 레인과 경합할 수 있어 의도적으로 돌리지 않았다. 따라서 AC-CMD-011 이 착수 시점에 완전히 도달 가능한지는 **미확정**이다.
- 미러 검사기가 실제로 어떤 요약 문자열을 낼지는 구현 전이라 알 수 없다. 폭 판정은 SPEC 문구에서 재구성한 **후보** 문자열(최장 104 runes)로 했으며, 구현된 실제 문자열의 폭은 미측정이다.
- `TestDoctorGolden_*` 가 어떤 환경 스텁 아래 도는지는 확인하지 않았다. golden 이 claude-only 행을 담고 있다는 사실(`doctor-nocolor.golden:43`)만 읽었고, 그 테스트가 `codex` 설치된 머신에서도 안정적인지는 이 SPEC 의 범위 밖이며 관측하지 않았다.
- MCP 서버 빌드(`e79c010b8`)는 이 워크트리 HEAD(`26b77973e`)보다 앞선 커밋에서 기동돼 있다. `_root.source: param` 으로 읽힌 트리는 확인했으나, 서버 측 감사 로직 자체는 이 트리의 소스가 아니라 그 빌드의 것이다.

## Residual risk (관측한 것에도 불구하고 여전히 틀릴 수 있는 것)

- D2 를 닫지 않은 채 run-phase 로 넘어가면, 구현 에이전트가 존재하지 않는 테스트에 대한 `ok` 출력을 AC 통과 증거로 인용할 수 있다. 이 위험은 점수와 무관하게 남으며, 감사가 아니라 §E 증거 귀속에서만 잡힌다.
- 읽기 전용 경계는 AC-CMD-010 이 기계적으로 막지만, 그 AC 역시 아직 없는 테스트에 의존한다(D2 대상 8건 중 하나). 즉 **범위 침범 방지 장치 자체가 D2 의 공허-초록 노출을 공유한다** — 이것이 D2 를 major 로 둔 두 번째 이유다.
- D8 의 경로 드리프트는 이 SPEC 을 완벽히 구현해도 남는다. 생산자가 `mirrorSkillsRelDir` 를 바꾸는 순간 doctor 는 조용히 잘못된 "미러 없음" finding 을 내며, 어떤 AC 도 그것을 잡지 못한다.
