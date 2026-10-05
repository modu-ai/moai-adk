# t1088 — TestCodexCommand_NeutralityScan FAIL 원인 판정서

- card: t1088 (class B — 원인 규명, 수리는 별도 Kickoff)
- branch: `WT-neutrality-scan` @ `d323f68fd` (= origin/develop, fetch 후 rev-parse 확인)
- date: 2026-09-23
- measured by: lane-30 session (t1088 worktree)

## Claim

`TestCodexCommand_NeutralityScan` FAIL 은 **스캐너 범위 문제가 아니라 실제 오염**이다.
t1078 의 커밋 `6af5dc233`("feat(t1078): load Claude local instructions in Codex")이
배포 바이너리의 사용자 대면 표면인 `codex.Long` 도움말에 로컬 파일명
(`"CLAUDE.local.md"`)을 상수 보간으로 넣었고, AC-CL-013 의 전체 금지 패턴 표가
설계된 대로 그것을 잡았다. 위반은 정확히 1건 — `codex.Long` × `"CLAUDE.local"`.

## Evidence

측정 명령: `go test ./internal/cli/ -run TestCodexCommand_NeutralityScan -count=1`

| baseline | 결과 |
|---|---|
| `f5fff2190` (t1020 GOAL-DIST 머지 — t1078 이전) | `ok github.com/modu-ai/moai-adk/internal/cli 0.814s` — **PASS** |
| `0b11bb137` (t1078 머지, 2026-09-22) | **FAIL** — `codex.Long = "..." matches forbidden pattern "CLAUDE.local"` |
| `d323f68fd` (HEAD = origin/develop tip) | **FAIL** — 동일 단일 위반 |

- 카드 [HARD] 요건 충족: t1078 이전 미발생이 측정으로 확정됐다 (f5fff2190 PASS).
- 리드의 CI 전제 확인: t1078 병합 시점 발현, 그 이전 미발생 — 둘 다 재측정으로 일치.

귀속:

- `git log -S 'is common local guidance shared with Claude' -- internal/cli/codex_launcher.go`
  → `6af5dc233` 단일 적중. (문장 전체 `-S` 는 0행이었다 — Long 은 상수 연접으로 조립돼
  소스에 통째로 존재하지 않는다. `codex_launcher.go:376-377` 가
  `codexClaudeLocalName + " is common local guidance shared with Claude;\n" + ...` 형태.)
- 상수 정의: `internal/cli/codex_contract.go:33-34` —
  `codexLocalInstructionName = "AGENTS.local.md"`, `codexClaudeLocalName = "CLAUDE.local.md"`.

스캐너 변경 교란 배제:

- `git diff --stat f5fff2190 0b11bb137 -- internal/cli/codex_launcher_guards_test.go`
  → 출력 없음 (바이트 동일). 두 시점 사이 테스트(패턴 표 포함) 불변 —
  원인은 내용 단독이다.

레인 측 커버리지 사실:

- `git merge-base --is-ancestor 6af5dc233 9bd521f5c` → exit 0.
  도입 커밋은 t1078 레인 자기 트리 팁(`9bd521f5c`, develop 흡수 완료 상태)의 조상이다.
  즉 위반은 병합 전 레인 트리에 이미 존재했고, 레인 로컬 검증은 이 패키지 테스트를
  실행하지 않았거나 표면하지 않았다. (sync-audit PASS 100/100 과 모순 아님 — 감사 축이 다르다.)

## 분류 — 스캐너 범위 문제 기각

1. **기록된 설계 결정** (`codex_launcher_guards_test.go:330`, plan.md M4.2):
   리터럴 스캔의 축소는 관측+기록과 함께 했고, "never by weakening
   TestCodexCommand_NeutralityScan" — 명령 표면은 전체 표를 유지하기로
   명시적으로 결정돼 있다. 스캐너는 의도대로 발화했다.
2. **t1078 카드 자신의 선행 판정과 모순**: 같은 카드가 배포 템플릿
   (`internal/template/templates/AGENTS.md.tmpl:271`)에는 파일명 대신
   "common local guidance / Codex-specific local guidance" 로 썼고, 그것을
   자기 테스트(`codex_local_instructions_test.go:394-395` —
   "template must describe common and Codex-specific inputs **without enumerating
   local filenames**")로 강제했다. 명령 표면(`codex.Long` — 배포 바이너리의
   `--help` 사용자 대면 문자열)은 같은 원리를 적용받았어야 했다.
3. 적중 대상이 "소스 내부 구현 참조"(리터럴 스캔이 축소한 축)가 아니라
   배포 표면 그 자체라는 점에서, 기각 근거(codex_contract.go 축소 사례)와
   사안이 다르다.

## 수리 방향 제안 (구현 아님 — Kickoff 별도)

- **A (권장)**: `codex_launcher.go` Long 의 파일명 보간 2줄(376-377)을
  템플릿 §8 이 쓰는 일반 어휘로 교체 — 예: "Common local guidance shared with
  Claude; Codex-specific local guidance. Both non-empty project-root files are
  injected as developer instructions, in that order." 스캐너 무약화. 영향 1파일.
  수리 시 확인 대상: Long 문구를 exact-match 로 단언하는 테스트가 있는지
  (`codex_launcher.go:362-364` 주석이 exact-match 셀 언급), 형제
  `codex_local_instructions_*` 계열 재측정.
- **B (운영자 판정 사항)**: 도움말이 파일명을 명시하는 것이 사용자 문서로서 옳다는
  방향이면, AC-CL-013 명령 표면 표에서 `CLAUDE.local` 클래스를 재고하는
  **독트린 변경**이 필요하다 — 기록된 M4.2 결정을 뒤집는 일이라 레인 단독 불가,
  리드/운영자 상신.

## Gaps

- 전체 스위트 판정은 CI 몫 — 로컬은 영향 패키지의 단일 테스트만 측정했다.
- t1078 레인이 당시 실행한 검증 명령 기록은 이 카드에서 읽지 않았다 —
  "레인 트리에 위반이 존재했다"(ancestor 측정)까지만 관측했다.
- B 방향 선택 시 영향 범위(AC-CL-013 표의 다른 소비자)는 미측정.
- 루트 `AGENTS.md:265` 는 파일명을 유지 중이고 템플릿(`:271`)은 아니라, 두 사본이
  이 축에서 갈라져 있다 — 본 카드 범위 밖이나 A 수리의 어휘 선택 시 참조된다.

## Residual-risk

- `codex.Long` 의 exact-match 검사가 존재하면 A 수리 때 그 테스트들도 함께
  수정돼야 한다 (수리 Kickoff 후 발견될 것).
- 배차 전제 외의 다른 패키지에 동종 오염(t1078 이 넣은 다른 사용자 대면 문자열)
  가능성은 본 측정이 전수하지 않았다 — `go test ./internal/cli/` 전체 또는 CI 가
  담당한다.

---

## 수리 기록 (A안 이행 — 운영자 Kickoff 승인, 2026-09-23)

- base: develop `17f71a13d` 흡수(fast-forward) 후 착수
- 변경 2파일: `internal/cli/codex_launcher.go:376-378`(Long 파일명 보간 → 일반 어휘),
  `internal/cli/codex_local_instructions_test.go:144-156`

### 확인 항목

1. **exact-match 단언 여부**: `codex_launcher.go:362-364` 주석의 exact-match 셀은 진단 상수
   (`codexUsageDiag` 등) 대상이며 Long 을 통째로 단언하는 테스트는 없음. 단, Long 부분 문자열
   단언 중 **`TestCodexLocalInstructions_DocumentedInLauncherHelp`** 가 `"CLAUDE.local.md"` /
   `"AGENTS.local.md"` 포함을 **요구**하고 있었다 — 중립성 스캔과 정면 모순(잔여 위험 예고 적중).
   A안과 같은 원리로 개정: 일반 어휘(`Common local guidance`, `shared with Claude`,
   `Codex-specific local`, `developer instructions`, `non-empty`) 포함 + 두 파일명 **비포함** 단언
   추가(템플릿 테스트 `:394-395` 와 동일 원리).
2. **형제 재측정**: `go test ./internal/cli/ -run TestCodex -count=1 -v` → exit 0,
   `ok github.com/modu-ai/moai-adk/internal/cli 13.169s`, `--- PASS` 248 / `--- FAIL` 0,
   `TestCodexLocalInstructions*` 19건 PASS. 수정본 복원 후 재실행 → exit 0 `ok ... 11.073s`.
3. **스캐너 무변이 PASS + 복원 변이 FAIL**: `codex_launcher_guards_test.go` 무변경.
   수정본에서 `--- PASS: TestCodexCommand_NeutralityScan`. Long 을 원문(파일명 보간)으로 되돌린
   변이에서 exit 1:
   - `codex_launcher_guards_test.go:158: codex.Long = "...CLAUDE.local.md is common local guidance..." matches forbidden pattern "CLAUDE.local"` → `--- FAIL: TestCodexCommand_NeutralityScan`
   - `codex_local_instructions_test.go:154: launcher help enumerates local filename "CLAUDE.local.md"` / `"AGENTS.local.md"` → `--- FAIL: TestCodexLocalInstructions_DocumentedInLauncherHelp`

### 정적 검사

- `gofmt -l` 두 파일 → 출력 없음
- `go vet ./internal/cli/` → exit 0
- `golangci-lint run ./internal/cli/` → exit 0, `0 issues.`

### Gaps

- `internal/cli` 전체 스위트와 타 패키지는 미실행(CI 몫, 전체 스위트 금지 지시).
- 루트 `AGENTS.md:265` 의 파일명 서술은 범위 밖으로 손대지 않음(판정서 Gaps 와 동일).
