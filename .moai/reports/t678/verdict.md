# t678 — 설치 바이너리 주체 불명 다운그레이드: 원인 규명 + 가드 판정서

- **Card**: t678 (Class B — 교체 주체 식별 → 재현 → 가드)
- **Branch**: `WT-binary-downgrade` (base: develop `c1baee210`)
- **Fix commit**: `a2572d0a5` (미푸시 1커밋)
- **Date**: 2026-09-13

## 결론 (한 문장)

교체 주체는 **moai 자신**이다 — `moai`의 SessionStart 훅에 배선된 자가-업데이트 핸들러
(`internal/hook/auto_update.go`, 등록: `internal/cli/deps.go:226`)가 모든 세션 시작마다
발화하고, 개발 빌드 판별식(`dirty`/`dev`/`none` 부분문자열)이 빌드 코드네임
`moai_cp/20260910_130400`을 개발 빌드로 분류하지 못해 GitHub 최신 릴리스 v3.1.2를
내려받아 `~/go/bin/moai`를 덮어썼다.

## 사건 연쇄 (관측 근거)

| 시각 | 사건 | 근거 |
|---|---|---|
| 19:26 | 리드 `make install` (93b0e8f04 계열, ~70MB) | `make-install.log` |
| 19:29:06.748 | 업데이터가 최신 릴리스 조회 → v3.1.2 available 판정 | `~/.moai/cache/update_check.json` `checked_at` 필드 (초 단위 일치) |
| 19:29:06 | 교체 전 바이너리(70,708,178B)를 백업 | `~/go/bin/moai.backup.1789295346` birth=mtime 19:29:06, 크기 일치 |
| 19:29:07 | `~/go/bin/moai`가 v3.1.2(4b2f203fe, 35,970,082B)로 교체 | 배차 지시문 실측 (35,970,082B = GoReleaser 스트립 산출물) |
| 19:34:01 | 리드 재설치 (e708884b0, 70,741,250B) | 현 바이너리 stat |

전례: `moai.backup.1787220531` (2026-08-20 19:08:51) — 동일 크기 35,970,082B(v3.1.2).
1회성 사고가 아니라 세션 시작마다 재발하는 경로다.

## 배제한 주체 후보 (관측 기반)

- **Claude 세션/서브에이전트**: 당시 활성 전사 12개 파일(`~/.moai/claude-profiles/moai-adk/projects/**`) 전수에서
  `"command":"…moai update"` Bash 도구 호출 0건 — 발견된 언급은 전부 문서 주입 텍스트(CLAUDE.local.md 등).
- **`go install @v3.1.2`**: `$(go env GOMODCACHE)/github.com/modu-ai/` 부재 — 모듈 캐시 흔적 없음.
- **대화형 셸**: `~/.zsh_history`에 사건 시각의 `moai update` 없음.
- **launchd/cron**: `kr.moai.t468.*`는 06:44-07:56 대만 실행. crontab 없음.
- **Codex 세션**: 당일 `~/.codex/sessions` 활성 없음.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

**Claim**: 개발 코드네임 빌드는 세션 시작 자가-업데이트로 릴리스로 덮여쓰일 수 있으며,
`pkg/version.IsDevBuild` 가드로 차단된다.

**Evidence** (이 run, 이 트리 `c1baee210` 기점에서 측정):
- RED: `go test ./pkg/version/` → `undefined: IsDevBuild` (build failed)
- GREEN: `go test ./pkg/version/` → `ok … 0.420s` — `TestIsDevBuild` 16 케이스
  (사건 문자열 `moai_cp/20260910_130400` → true 포함)
- 재배선 후: `go test ./internal/update/` → `ok 6.428s`;
  `go test ./internal/cli/ -run 'TestBuildAutoUpdateFunc|TestShouldSkipBinaryUpdate|TestRunUpdate_BinaryOnly|TestSkipBinaryUpdate_DeferredCheckSkipped|TestRunVersionBranch'`
  → `--- PASS` 26개, `ok`
- `go vet ./internal/cli/ ./internal/update/ ./pkg/version/` → exit 0
- `golangci-lint run pkg/version/... internal/update/...` → `0 issues`;
  `internal/cli/...` 37건은 전부 기존 파일(gateway*/migrate_cg/gpt_auth) — 내 편집 파일 0건

**Baseline-attribution**: 모든 수치는 이 커밋 작성 직전 워크트리
(`.claude/worktrees/t678`, `WT-binary-downgrade` @ `c1baee210`)에서 위 명령의 실출력.
포렌식 파일들(update_check.json, moai.backup.*)은 2026-09-13 실측 당시 birth/mtime.

**Gaps** (관측하지 못한 것):
- 19:29:06에 SessionStart를 발화한 **구체 세션 id**는 특정하지 못했다
  (같은 창에 86aede58 trace와 ef5f5dea custom-title이 기록됐지만 어느 쪽 발화인지는 단정 불가).
  기제 규명에는 무관 — 핸들러는 "모든 세션 시작"에 발화가 설계돼 있음.
- 실바이너리 덮어쓰기 재연출은 수행하지 않았다(조사 중 `~/go/bin/moai` 덮어쓰기 금지 [HARD]).
  재현은 판별식 단위(RED→GREEN)와 포렌식 연쇄로 갈음.
- deps.go:430의 채널 선택축(rc/alpha/beta → 브랜치 채널)은 의도가 다른 코드로 확인하고
  손대지 않았다 — 이 축의 별도 검증은 하지 않았다.

**Residual-risk**:
- SessionStart 자가-업데이트 설계 자체(사용자 바이너리를 조용히 교체)는 유지된다 —
  릴리스 바이너리 사용자에게는 이전과 동일하게 작동한다. 설계 변경은 별도 논의 사항.
- `moai.backup.*` 파일은 계속 축적된다(청소 경로 별도 카드 후보).
- 릴리스 버전 문자열을 3-숫자-컴포넌트로 검증하므로, 향후 4-컴포넌트 버전 관례가
  도입되면 그 버전은 dev로 분류된다(보수적 방향으로 안전측 실패).

## 변경 파일

| 파일 | 변경 |
|---|---|
| `pkg/version/version.go` | `IsDevBuild(v string) bool` 신설 — SSOT 판별식 |
| `pkg/version/version_test.go` | `TestIsDevBuild` 16 케이스 (사건 문자열 포함) |
| `internal/cli/deps.go` | `buildAutoUpdateFunc` → `version.IsDevBuild` |
| `internal/cli/update.go` | `shouldSkipBinaryUpdate` → `version.IsDevBuild` |
| `internal/update/local.go` | `localChecker.isDevVersion` → `version.IsDevBuild` 위임 |

## 통합 창 재측정 (리드 요구 — 행동 RED 대조, 2026-09-13 19:5x KST)

리드 지적: 초기 RED(`undefined: IsDevBuild` 빌드 실패)는 결함 자체의 증명이 아니다.
요구대로 **수리 전 판별식이 사건 입력을 release로 분류함을 보이는 행동 RED**를 확보했다.

- **수리 전 판별식 출처**: `develop e7b93c120:internal/cli/deps.go:500-502` 발췌(그대로 복사),
  임시 테스트로 같은 입력 `"moai_cp/20260910_130400"`에 단언.
- **행동 RED** (병합 트리 `9a0692fa7`에서 실행):
  `go test ./pkg/version/ -run TestT678BehavioralRed` →
  `--- FAIL: … 1` — 관측값 `IsDevBuild(입력) = false, want true (dev) — defect reproduced:
  input read as RELEASE, auto-update proceeds`. **1실행 · 1 FAIL · 0 PASS.**
  (측정 후 임시 테스트 파일은 삭제 — 커밋에 미포함)
- **행동 GREEN** (같은 병합 트리, 같은 입력이 표에 포함된 커밋된 테스트):
  `go test ./pkg/version/ -run TestIsDevBuild -v` →
  `--- PASS: TestIsDevBuild` 1 + 서브테스트 `--- PASS` 16 = **17 PASS · 0 FAIL.**
- **재배선부 재측정** (병합 트리): cli 선택자 테스트 `--- PASS` **26**, `ok`;
  `go vet` 3 패키지 exit 0.
- **트리 동일성**: develop 병합 커밋 `cd1e5298c`의 `HEAD^{tree}` = `219826baa57e…` =
  사전 재측정 병합 트리(`9a0692fa7`의 트리) — **양변 동일**, 위 근거가 develop 병합 결과에 그대로 귀속.

## 통합 기록

- 창: `moai integration acquire --name lane-1 --card t678` (보유) → `release` (반납 완료)
- 흡수: `git merge --no-edit origin/develop` (= `e7b93c120`) → 병합 커밋 `9a0692fa7` (충돌 0)
- develop 병합: `EnterWorktree(.claude/worktrees/develop)`에서
  `git merge --no-ff WT-binary-downgrade` → **로컬 develop 병합 SHA `cd1e5298c`**
- 미푸시 3커밋: `cd1e5298c`(병합) · `9a0692fa7`(흡수 병합) · `a2572d0a5`(수리)
- push는 하지 않음 — 리드 일괄 소관. 워크트리 `.claude/worktrees/t678` 유지.
