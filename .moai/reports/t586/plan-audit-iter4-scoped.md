# SPEC 한정 재감사 보고: SPEC-INIT-TUX-I18N-001

> **운영자 한정 연장 승인(2026-09-11) — 범위: F1 해소 여부와 변경분의 신규 결함.**
> 3회 상한을 넘는 한정 연장이다. SPEC 전체를 다시 채점하지 않는다. 이 보고의 판정은 아래 두 질문에만 답한다.
> (1) 3회차 blocking 결함 F1 이 해소됐는가. (2) 3회차 감사 뒤 변경분이 새 결함을 들여왔는가.

- 카드: t586 · 감사 단계: plan · 회차: 4 (한정 연장, 전체 재채점 아님)
- 감사 대상 트리: 워크트리 `.claude/worktrees/t586`, 브랜치 `WT-init-tux-i18n`, HEAD `699bb8f602dc51d7a2c9690fc8b1e6bff4120384`. 착수 시와 보고서 작성 직후에 `git rev-parse HEAD` 로 같은 값을 확인했다. `git status --short -- .moai/specs/SPEC-INIT-TUX-I18N-001 internal/cli/profile_setup.go` 출력 없음(커밋본과 워킹 사본 동일).
- 감사한 변경분: `git diff 268cffe2c 699bb8f60 -- .moai/specs/SPEC-INIT-TUX-I18N-001/` (5파일, +50 −9). `git diff --stat 02227fa12 268cffe2c` 는 `.moai/reports/t586/plan-audit-iter3.md` 1파일(+347)뿐이라, 3회차 감사 트리 `02227fa12` 이후 SPEC 문서의 변경은 이 차분이 전부다. 차분의 커밋은 `699bb8f60` 하나다.
- 이전 판정: `.moai/reports/t586/plan-audit-iter3.md` (FAIL 0.90, blocking F1, optional F2~F5).
- 참고 입력: 레인 재현 기록 `.moai/reports/t586/f1-repro.md` 와 딸린 `f1-repro-*.txt`·`f1-anchor-*.txt`. 레인 기록은 참고만 했고, F1 판정은 아래 E-1~E-3 의 자체 재측정에 둔다.
- 판정 빌드: 이 트리(HEAD `699bb8f60`)에서 `go build -o <scratchpad>/iter4/moai ./cmd/moai` 로 만든 바이너리(종료 0, ldflags 없음 — `version` 출력 `v3.1.3 none built unknown`). 설치본 `moai` 는 쓰지 않았다.
- 작성자 추론 맥락은 M1 격리 원칙에 따라 배제했다. 이 파일에 앞선 시도가 남긴 부분 기록이 있었으나 통째로 대체했다.

## 판정

**한정 판정: PASS.** F1 은 증거 수준에서 해소됐다. 변경분에서 blocking 결함은 나오지 않았고 optional 관찰 두 건(O1, O2)만 있다.

점수는 매기지 않는다. 이 판정은 위 두 질문에 한정되며 SPEC 전체 품질 점수를 대신하지 않는다.

## F1 처분 — 해소

3회차 F1: S2 가드 `TestTUINestedConfigNoParallelWriter` 는 주석 아닌 줄에 `persistProjectConfig` 문자열이 있는지만 봤다(`profile_setup_nested_test.go:39-41`). 같은 파일에 정의 `func persistProjectConfig(`(`profile_setup.go:160`)가 있어, 호출(`:520`)만 지운 뮤턴트에서도 가드가 참이었다.

v0.2.3 의 수정: `design.md` §10 S2 행(`design.md:161`)이 양성 절 기준을 "`profile_setup.go` 의 주석 아닌 줄 가운데 `persistProjectConfig(` 를 담고 정의 줄 `func persistProjectConfig(` 가 아닌 줄이 1개 이상" 으로 바꿨고, `acceptance.md` AC-ITI-010 (2)(`:120`)와 (4)(`:122`)를 같은 기준으로 맞췄다. 셸 형태는 `research.md` §17 에 있다.

### E-1 원본과 호출 전용 뮤턴트에 앵커 적용

사본은 저장소 밖 스크래치 디렉터리(`<scratchpad>/iter4/`)에 두었고 제품 트리는 건드리지 않았다.

```
$ sed -n '520p' internal/cli/profile_setup.go
			if err := persistProjectConfig(cwd, developmentMode, ""); err != nil {
$ sed -i '' '520s/persistProjectConfig(cwd, developmentMode, "")/error(nil)/' mut/profile_setup.go
$ diff orig/profile_setup.go mut/profile_setup.go
520c520
< 			if err := persistProjectConfig(cwd, developmentMode, ""); err != nil {
---
> 			if err := error(nil); err != nil {
diff_exit=1

== 원본
$ grep -n -F 'persistProjectConfig(' orig/profile_setup.go | grep -v -E '^[0-9]+:[[:space:]]*//' | grep -v -F 'func persistProjectConfig('
520:			if err := persistProjectConfig(cwd, developmentMode, ""); err != nil {
orig_anchor_exit=0

== 호출 전용 뮤턴트
$ grep -n -F 'persistProjectConfig(' mut/profile_setup.go | grep -v -E '^[0-9]+:[[:space:]]*//' | grep -v -F 'func persistProjectConfig('
(출력 없음)
mut_anchor_exit=1

== 뮤턴트 첫 단계(정의 줄 잔존 대조)
$ grep -n -F 'persistProjectConfig(' mut/profile_setup.go
160:func persistProjectConfig(projectRoot, devMode, convention string) error {
mut_first_stage_exit=0
```

판독: 원본에서 호출 줄 1개(`:520`), 뮤턴트에서 0줄이다. 뮤턴트의 첫 단계가 정의 줄 `:160` 을 읽으므로 빈 결과는 파일을 못 읽어서가 아니라 필터가 정의 줄을 걸러서 생긴 것이며, 이 상태가 바로 옛 기준에서 참이던 상태다. 뮤턴트는 컴파일 가능하다 — `developmentMode` 는 `:440` `Value(&developmentMode)` 에서도 쓰이고, `persistProjectConfig` 는 테스트 파일 3곳(`profile_setup_projectconfig_test.go`, `profile_setup_removed_questions_test.go`, `profile_setup_nested_test.go`)이 참조한다. 따라서 AC-ITI-010 (4) 가 요구하는 "`--- FAIL` 과 테스트 이름" 은 빌드 오류가 아닌 테스트 실패로 관측될 수 있다.

### E-2 앵커가 다른 방식으로 공허하지 않은지

| 변형 | 만든 방법 | 앵커 결과 | 판독 |
|---|---|---|---|
| 호출을 여러 줄로 나눈 재배치(M5 재조준에서 있을 법한 형태) | `perl` 로 `persistProjectConfig(\n cwd, developmentMode, "",\n)` 로 바꿈 | `520:			if err := persistProjectConfig(` · exit 0 | 참 유지. 재배치로 거짓 경보가 나지 않는다 |
| 호출 줄 전체를 주석 처리 | `520s/^([[:space:]]*)/\1\/\/ /` | 출력 없음 · exit 1 | 주석 필터가 막은 호출을 거짓으로 만든다(의도대로) |
| 원본의 주석 속 언급 | `grep -n 'persistProjectConfig' internal/cli/profile_setup.go` | `:152`·`:227`·`:243`·`:518` 모두 괄호 없음 | 첫 단계(`persistProjectConfig(`)에서 이미 빠진다. 주석 필터가 호출 줄 `:520` 을 떨어뜨리지 않는다(E-1 원본 결과) |
| 호출을 지우고 같은 줄 끝 주석에 `persistProjectConfig(` 를 남김 | `520s|$| // was persistProjectConfig(cwd)|` (뮤턴트 기준) | `520:			if err := error(nil); err != nil { // was persistProjectConfig(cwd)` · exit 0 | 참이 된다. 줄 끝 주석은 걸러지지 않는다 → O1 |

기존 가드의 주석 필터 `nonCommentLines`(`profile_setup_nested_test.go:46-56`)는 `strings.HasPrefix(strings.TrimSpace(line), "//")` 로 줄 전체 주석만 뺀다. SPEC 의 셸 필터 `^[0-9]+:[[:space:]]*//` 와 같은 동작이다. 구현할 때는 정의 줄 제외가 줄 단위라 지금의 `strings.Contains(codeLines, …)` 한 번으로는 표현되지 않고 줄마다 판정해야 하지만, 이는 run 단계 구현 방식이지 SPEC 결함이 아니다.

### E-3 대상 파일이 흡수 뒤에도 맞는지

- `awk` 로 `:520` 을 감싼 함수: `240: func runProfileSetup(cmd *cobra.Command, args []string) (err error) {`.
- `design.md` §2.2(`:18-29`)의 변경 뒤 구조는 `runProfileSetup (cli)` 아래 `저장 이하 ← 그대로` 다. `plan.md` M4(`init.go`·`update.go` 의 확인창과 `runProfileSetup` 호출 삭제)는 `profile_setup.go` 를 건드리지 않고, M5 는 `profile_setup.go` 를 고치되 폼만 wizard 패키지로 옮긴다.
- 비테스트 코드에서 `persistProjectConfig` 를 담은 파일은 `internal/cli/profile_setup.go` 하나다(`grep -rn -l 'persistProjectConfig' internal/ --include='*.go'`, 나머지 셋은 `_test.go`).

판독: 호출 줄은 흡수 뒤에도 `profile_setup.go` 에 남는다는 설계 주장은 SPEC 문서 안에서 일관된다. 흡수 뒤 실제 파일로 실행 확인한 것은 아니다(Gaps).

### AC·REQ 정합

- AC-ITI-010 (2)(`acceptance.md:120`)의 S2 기준 서술과 `design.md:161` 의 기준 서술이 같은 세 조건(주석 아닌 줄, `persistProjectConfig(` 포함, `func persistProjectConfig(` 아님)을 쓴다. (2)에 "정의 줄만으로는 충족되지 않는다 … `research.md` §17 음성 대조" 가 붙어 있다.
- AC-ITI-010 (4)(`acceptance.md:122`)의 S2 뮤턴트는 "호출만 지우고 정의는 남긴 사본" 으로 정확히 지정됐고, 기대 결과가 `TestTUINestedConfigNoParallelWriter` 의 `--- FAIL` 이다. E-1 에서 이 사본의 앵커가 거짓이므로, 설계대로 구현한 가드는 이 뮤턴트에서 실패할 수 있다.
- 따라서 REQ-ITI-010 "every positive guard shall fail when its asserted construct is removed" 는 S2 에 대해 이제 AC 로 판정 가능하다. 3회차 Traceability 감점 사유(F1)는 사라졌다.

## 변경분 hunk 별 검토

| # | 파일:줄(HEAD) | 변경 | 검토 결과 |
|---|---|---|---|
| 1 | `spec.md:4` | `version` `0.2.2` → `0.2.3` | 문제 없음. 따옴표 semver 유지 |
| 2 | `spec.md:30` | HISTORY 0.2.3 행 추가 | F1~F5 반영 내용과 차분이 일치. 0.2.2 행의 "REQ-ITI-017 을 State-driven 으로 고침(N7)" 은 이력 기록이라 0.2.3 의 Where 변경과 모순이 아니다 |
| 3 | `spec.md:169` | §B 머리 주석에 `Where` 추가, 쓰임새 정의 | 문제 없음 |
| 4 | `spec.md:198` | REQ-ITI-017 `(State-driven) While` → `(Where) Where … shall` | 정식 Where 패턴(`Where [조건], the <subject> shall …`)과 맞는다. 조건 "the source tree carries the card t583 init question set" 은 정적 트리 내용이라 GEARS 의 Where(정적 구성·기능 게이트) 정의에 맞는다. 판정 빌드 lint: 이 트리에서 `0 error(s), 0 warning(s)`(exit 0). 양성 대조: 이 줄의 `shall` 을 모두 `should` 로 바꾼 사본에서 같은 빌드가 `WARNING ModalityMalformed … REQ REQ-ITI-017: EARS modality violation` 을 냈다. 현 문서에 REQ-ITI-017 을 State-driven 으로 부르는 곳은 이력 행 외에 없다 |
| 5 | `design.md:161` | S2 행 재작성 | F1 수정 본체. E-1~E-3 로 확인. O2(서술 보완 여지) |
| 6 | `acceptance.md:60` | 비교 제외 문단에 migrate-tx 체크포인트 제외 사유 추가 | 좌표 확인: `migrate_agency.go:188-194` `checkpointPath`(`MOAI_HOME` 이 비어 있지 않은 절대 경로면 그 아래, 아니면 `<homeDir>/.moai/`), `update_residue_cleanup.go:85` 가 `update.go:858` `runAgencyMigrationAdapter` 를 부름, `:251`·`:367` 에서 `cpPath` 를 받음 — 모두 원문과 같다. 도달성: 비테스트 호출자는 `update_residue_cleanup.go:85`(그 위 `runV3ResidueCleanup` 은 `update.go:480`·`:609` 에서만 호출)와 `moai migrate agency` 명령(`migrate_agency.go:725`)뿐이고, `:84` 에서 프로젝트 루트에 `.agency` 가 있어야 들어간다. pty 사례는 init 첫 화면(AC-ITI-003, 빈 임시 작업 디렉터리), 프로필 위저드, "실제 확인창 생성 헬퍼"(§C AC-ITI-015 (a)), 확인형 픽스처만 부른다. 이 경로에 닿지 않는다는 주장은 성립한다 |
| 7 | `acceptance.md:120` | AC-ITI-010 (2) S2 기준을 호출 줄로 | F1 정합. 위 "AC·REQ 정합" |
| 8 | `acceptance.md:122` | AC-ITI-010 (4) S2 뮤턴트를 호출 전용으로 명시 | F1 정합. 위 "AC·REQ 정합" |
| 9 | `research.md:9` | 자기 적중 문장 정정(F2) | HEAD 에서 재측정: 전체 `*.md` 형태 → `research.md:9` 한 줄, exit 0(자기 적중, 문장이 적은 대로). 다섯 문서 형태 → 출력 없음, exit 1. 대조군 `grep -c 'REQ-ITI-017'` → `plan.md:2`, `spec.md:6`, `design.md:0`, `acceptance.md:4`, `progress.md:1` 로 네 파일 ≥1. 문장의 "네 파일에서 1 이상" 과 일치하고, 0 결과 옆에 양성 대조군이 있다 |
| 10 | `research.md:143` | §6.1 대체 표식(F5) | 옛 `\.Group\b` 판독 문단 바로 앞에 붙었다. 표식 자체에는 `\b` 가 없다 |
| 11 | `research.md:329` | §14 끝에 F3 항목 추가 | hunk 6 과 같은 좌표·같은 결론. 문서 사이 불일치 없음 |
| 12 | `research.md:397-430` | §17 신설 | 측정 트리 `268cffe2c` 명시, 음성 결과에 정의 줄 잔존 대조와 첫 단계 두 줄 판독이 붙어 있다. ERE 에 `\b` 없음. 공백 절(테스트 미실행)을 스스로 밝혔다. 내 E-1 재측정이 같은 결과를 냈다 |
| 13 | `progress.md:14-16` | 3회차 결과 기록, v0.2.3 개정 기록 | 3회차 점수 0.90 은 3회차 보고서 이력표(`plan-audit-iter3.md:341`)와 같다. "재감사 결과는 아직 없다" 는 이 보고 전 시점 서술로 참이다 |

변경분 전체에서 새로 들어온 `\b`+ERE 조합, 대조군 없는 0 판정, 끊어진 추적성은 찾지 못했다. REQ 수 18·AC 수 22 는 바뀌지 않았고 판정 빌드 lint 의 커버리지·모달리티 경고는 0 이다(위 hunk 4 의 양성 대조로 lint 가 이 SPEC 의 REQ 를 실제로 읽음을 확인).

## 새 발견

O1. S2-ANCHOR-TRAILING-COMMENT — `design.md:161`, `acceptance.md:120`, `research.md` §17 — 새 기준은 줄 전체 주석만 거르므로, 호출을 지우고 같은 줄 끝 주석(또는 문자열 리터럴)에 `persistProjectConfig(` 를 남긴 사본에서도 참이다. 증거: E-2 마지막 행(`520:			if err := error(nil); err != nil { // was persistProjectConfig(cwd)`, exit 0). AC-ITI-010 (4) 가 지정한 뮤턴트(호출만 삭제)는 잡으므로 F1 해소에는 영향이 없고, 옛 기준에도 있던 계열이다. — Severity: minor — Class: optional — 수정 제안(선택): 구현 시 줄 끝 `//` 뒤를 잘라 판정하거나, 그대로 두고 잔여 위험으로 기록한다.

O2. DESIGN-S2-TARGET-RATIONALE-OMITS-M5 — `design.md:161` — "M4 는 `init.go`·`update.go` 만 고치므로 대상 파일은 `profile_setup.go` 그대로다" 는 `profile_setup.go` 를 실제로 고치는 M5 를 언급하지 않는다. 같은 문장이 §2.2 "저장 이하 그대로" 를 근거로 들고 `research.md` §17 이 M5 를 명시적으로 다루므로 결론은 성립한다. — Severity: minor — Class: optional — 수정 제안(선택): "M4 는 … 고치지 않고 M5 는 저장 이하를 그대로 두므로" 로 다듬는다.

참고(판정 대상 아님): 레인 재현 기록 `f1-repro.md` Residual-risk 는 "호출을 함수 값으로 부르는 형태로 바뀌면 놓친다" 고 적었으나, 그 경우 앵커는 0줄을 내 가드가 **실패**한다(거짓 경보 쪽). 놓치는 방향이 아니다. SPEC 문서가 아니라 보고 기록이므로 결함으로 올리지 않는다.

## Defects Found

blocking 결함 없음. optional 2건(O1, O2)은 위 "새 발견" 참조.

## Regression Check

| 이전 결함 | 처분 | 근거 |
|---|---|---|
| F1 (blocking) | 해소 | E-1~E-3, AC·REQ 정합 |
| F2 (optional) | 해소 | hunk 9 재측정 |
| F3 (optional) | 해소 | hunk 6·11 좌표와 도달성 확인 |
| F4 (optional) | 해소 | hunk 4 패턴 확인과 lint 양성 대조 |
| F5 (optional) | 해소 | hunk 10 |

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

- **Claim**: F1 은 SPEC 문서 수준에서 해소됐고, 변경분은 blocking 결함을 들여오지 않았다.
- **Evidence**: 위 E-1~E-3 명령과 출력, hunk 표의 재측정, lint 출력(`0 error(s), 0 warning(s)` / 양성 대조 `ModalityMalformed`).
- **Baseline-attribution**: 모든 측정은 이번 실행에서 HEAD `699bb8f60` 트리와, 같은 트리로 빌드한 `<scratchpad>/iter4/moai` 로 했다. 변경분은 `.moai/specs/` 문서만 바꾸므로 `internal/cli/profile_setup.go` 는 3회차 트리와 같다(레인 기록의 `268cffe2c` 측정과 내 측정 결과가 일치).
- **Gaps**: 가드 테스트는 아직 새 기준으로 구현되지 않았으므로, 테스트 바이너리로 S2 가 `--- FAIL` 하는 모습은 관측하지 않았다(run 단계 M5, AC-ITI-010 (4) 소관). 흡수 뒤(t583 병합 이후) `profile_setup.go` 로는 실행하지 않았다. 변경분 밖의 SPEC 본문은 재채점하지 않았다.
- **Residual-risk**: O1 계열(줄 끝 주석·문자열 리터럴 안의 호출 토큰)은 앵커를 속일 수 있다. M5 재조준이 호출 형태를 함수 값 호출로 바꾸면 앵커가 0줄을 내 가드가 거짓 경보로 실패한다 — 놓치는 방향은 아니지만 run 단계에서 기준 재조정이 필요할 수 있다.

## 출처

- 감사 HEAD: `699bb8f602dc51d7a2c9690fc8b1e6bff4120384` (착수 시·작성 직후 동일)
- 감사한 차분: `268cffe2c..699bb8f60 -- .moai/specs/SPEC-INIT-TUX-I18N-001/`
- 판정 빌드: HEAD `699bb8f60` 에서 `go build -o <scratchpad>/iter4/moai ./cmd/moai`(ldflags 없음, `v3.1.3 none built unknown`)
- 감사자: plan-auditor (한정 연장 4회차, 2026-09-11)
