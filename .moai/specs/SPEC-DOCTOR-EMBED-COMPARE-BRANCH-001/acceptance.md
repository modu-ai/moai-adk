# acceptance.md — SPEC-DOCTOR-EMBED-COMPARE-BRANCH-001

Tier S. 아래 AC는 전부 이진 판정 가능하며, 각 AC의 근거는 축자 명령 + 축자 출력이다.

## §D AC Matrix

| AC | REQ | 요구 | 심각도 | 판정 근거 |
|---|---|---|---|---|
| AC-DECB-001 | REQ-DECB-002 | 비교 실패 픽스처에서 `CheckFail` | blocking | 테스트 통과 |
| AC-DECB-002 | REQ-DECB-003 | 평결이 comparison-failure로 **식별** 가능 | blocking | 메시지 단언 + 형제 3분기 대조 |
| AC-DECB-003 | REQ-DECB-004 | mutant RED 관측 | blocking | 축자 FAIL 출력 |
| AC-DECB-004 | REQ-DECB-004 | 원복 GREEN 관측 | blocking | 테스트 이름을 담은 축자 `--- PASS:` 행 |
| AC-DECB-005 | REQ-DECB-005 | 프로덕션 diff 0 | blocking | `git diff --stat` |
| AC-DECB-006 | — (회귀 가드, REQ 없음: 의도적) | 형제 3분기 테스트 비회귀 | blocking | 패키지 테스트 통과 |
| AC-DECB-007 | REQ-DECB-007 | 증거가 primary 체크아웃에 존재 | blocking | 경로 존재 확인 |
| AC-DECB-008 | REQ-DECB-001, REQ-DECB-006 | 테스트 파일 diff의 선언 구성 | blocking | `git diff` 기계 판독 |

`AC-DECB-006`의 REQ 공란은 드리프트가 아니라 의도다 — 이 AC는 어떤 요구의 파생이 아니라
기존 세 형제 테스트를 깨뜨리지 않았음을 지키는 **회귀 가드**이며, 그 대상은 본 SPEC이 만든 것이 아니다.

---

### AC-DECB-001 — 비교 실패 분기가 fail을 낸다

**Given** `newEmbedFixtureRoot`로 만든 커밋 세트에서 한 항목(`manager-git.toml`)이 정상 파일이 아니라
**같은 이름의 디렉터리**로 존재하고, `writeFakeBinary`로 판정 대상 바이너리가 존재하며,
`newExtractedDir` + `staticExtractor`가 그 이름의 **정상 파일** 대응물을 공급할 때
**When** `checkAgentEmitEmbedAgainst(root, "", staticExtractor(extracted), false)`를 호출하면
**Then** 반환된 check의 `Status`가 `uikit.CheckFail`이다.

### AC-DECB-002 — 평결이 형제 세 분기와 구별된다

**Given** AC-DECB-001과 동일한 픽스처
**When** 같은 호출의 반환 메시지를 검사하면
**Then** 메시지가 `comparison failed` 문자열을 포함하고,
동시에 형제 분기의 고유 접두 3종(`could not extract`, `compared `, `embeds stale`)을
**포함하지 않는다**. 상태값만 보는 단언은 이 AC를 충족하지 않는다 — 형제 세 분기도 `CheckFail`이므로
상태 단언은 분기를 식별하지 못한다.

### AC-DECB-003 — mutant가 심어졌을 때 테스트가 RED

**Given** `doctor_agentemit_embed.go:146` 분기의 `check.Status = uikit.CheckFail`을
`uikit.CheckOK`로 **일시 변형**한 트리
**When** `go test ./internal/cli/ -run '<새 테스트 이름>' -count=1 -v`를 실행하면
**Then** 종료 코드가 0이 아니고 출력에 해당 테스트의 `--- FAIL`이 나타난다.
축자 명령과 축자 출력을 `verdict.md`에 기록한다.

### AC-DECB-004 — mutant 원복 후 GREEN

**Given** AC-DECB-003의 mutant를 원복해 `git diff --stat -- internal/cli/doctor_agentemit_embed.go`가
빈 출력인 트리
**When** 동일 명령을 `-v`와 함께 다시 실행하면
**Then** 종료 코드가 0이고, 출력에 테스트 이름을 **문자 그대로 담은** 행
`--- PASS: TestAgentEmitEmbed_ComparisonErrorFails`가 나타난다.
이 `--- PASS:` 행 자체가 증거 토큰이며, 축자 명령과 함께 `verdict.md`에 기록한다.

**패키지 요약 `ok …`는 이 AC의 근거가 아니다.** `go test -run '<name>'`에서 셀렉터가 오타이거나
테스트가 개명되면 **매치 0건**으로 아무 테스트도 돌지 않은 채 `ok`와 종료 코드 0을 낸다 —
본 SPEC이 막으려는 바로 그 공허한 초록이, 하필 mutation 쌍의 증거로 인용되는 쪽에서 성립한다.
따라서 근거는 매치된 테스트가 실제로 실행되었음을 이름으로 보이는 `--- PASS:` 행에 한정한다.

### AC-DECB-005 — 프로덕션 코드 무변경

**Given** 카드 브랜치의 최종 커밋
**When** `git diff --stat c6aa61346..HEAD -- internal/cli/doctor_agentemit_embed.go`를 실행하면
**Then** 출력이 비어 있다. 그리고 SPEC 산출물을 제외한 코드 diff 대상은
`internal/cli/doctor_agentemit_embed_test.go` 단일 파일이다.

### AC-DECB-008 — 테스트 파일 diff는 함수 하나만 늘린다

**Given** 카드 브랜치의 최종 커밋
**When** `git diff c6aa61346..HEAD -- internal/cli/doctor_agentemit_embed_test.go | grep '^+func '`를
실행하면
**Then** 출력이 정확히 한 행이고, 그 행이 `+func TestAgentEmitEmbed_ComparisonErrorFails(t *testing.T) {`이다.
즉 추가된 최상위 선언은 테스트 함수 하나뿐이며, 새 픽스처 헬퍼 선언
(`func new…` 류, 또는 `newEmbedFixtureRoot` / `newExtractedDir` / `writeFakeBinary` / `staticExtractor`의
변형·대체)은 **0건**이다 (REQ-DECB-006). 헬퍼가 하나라도 추가되면 이 AC는 FAIL이다.

### AC-DECB-006 — 형제 세 분기 비회귀

**Given** 새 테스트가 추가된 트리
**When** `go test ./internal/cli/... -count=1`을 실행하면
**Then** 종료 코드 0이고, `TestAgentEmitEmbed_ExtractionErrorFails` ·
`TestAgentEmitEmbed_PartialExtractionFails` · `TestAgentEmitEmbed_DriftFailsAndNamesPath`가 모두 통과한다.

### AC-DECB-007 — 증거의 소재

**Given** run-phase 종료 시점
**When** `ls /Users/goos/MoAI/moai-adk-go/.moai/reports/t356/verdict.md`를 실행하면
**Then** 파일이 존재하고, AC-DECB-003·004의 축자 명령과 출력을 담고 있다.
워크트리 내부(`.claude/worktrees/t356/.moai/reports/`)에만 존재하는 사본은 이 AC를 충족하지 않는다.

---

## §D.1 엣지 케이스

- **디렉터리형 항목이 `uncompared`로 새지 않는가**: 추출 대응물이 **반드시 정상 파일**이어야 한다.
  대응물이 없으면 `os.ReadFile(extracted)`가 먼저 실패해 `continue` → 기수 부족(:155) 경로로 흘러
  잘못된 분기를 검증하게 된다. 픽스처 구성 시 이 순서를 지킬 것.
- **커밋 세트 크기**: 항목 1개로 최소화한다. 추가 항목은 `uncompared`로 쌓일 뿐 err 반환이 선행하므로
  분기 판정에 영향이 없다 — `compareEmission`은 커밋 측 읽기 실패에서 즉시 err를 반환하고,
  `:155`/`:162` 게이트는 `err == nil`일 때만 돌기 때문이다(`doctor_agentemit_embed.go:144-162`).
  순회 순서가 형제 분기를 이기게 하는 일은 일어날 수 없다.
- **플랫폼**: "is a directory" 문안은 OS별로 다를 수 있으므로, 단언 대상은 check 메시지의
  `comparison failed` 접두이지 OS의 errno 문안이 아니다.

## §D.2 Definition of Done

- [ ] 새 테스트 함수 1개 추가, 신규 파일 0, **새 픽스처 헬퍼 0** (AC-DECB-008 / REQ-DECB-006)
- [ ] AC-DECB-001..008 전부 PASS
- [ ] `go vet ./internal/cli/` 무경고
- [ ] 커밋 메시지에 카드 id `t356` 병기 (Conventional Commits)
- [ ] `verdict.md`가 primary 체크아웃에 존재
