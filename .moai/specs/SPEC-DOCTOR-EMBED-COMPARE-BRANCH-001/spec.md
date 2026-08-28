---
id: SPEC-DOCTOR-EMBED-COMPARE-BRANCH-001
title: "Agent Emit Embed 검사의 comparison-failure 분기(:146) 비회귀 테스트 — 약속된 4분기 중 무방비 1분기 봉합"
version: "0.1.0"
status: completed
created: 2026-08-28
updated: 2026-08-28
author: manager-spec
priority: P2
phase: "v3.1.4 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "doctor, embed-check, non-regression, test-only, mutation-red, changelog-promise"
tier: S
era: V3R6
related_specs: [SPEC-CI-DOCTOR-BIN-001, SPEC-AGENT-EMIT-LINEAGE-001]
---

# SPEC: doctor Agent Emit Embed — comparison-failure 분기의 전용 비회귀 테스트

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-08-28 | manager-spec | 최초 작성 — 카드 t356. 본 세션에서 실측한 `compareEmission` err 경로 + 프로브 관측(디렉터리형 커밋 항목 → "is a directory") 위에 구성 |

---

## 1. 배경 — 문서는 넷을 약속하고 테스트는 셋을 지킨다

`internal/cli/doctor_agentemit_embed.go`의 `checkAgentEmitEmbedAgainst`는 **판독 가능한 바이너리가
존재하는** 경로 아래에 `uikit.CheckFail` 분기를 네 개 갖는다.

| # | 위치 | 조건 | 메시지 접두 |
|---|---|---|---|
| 1 | `:139` | 추출 실패 (`extract` 오류) | `could not extract embedded artifacts from …` |
| 2 | `:146` | **비교 실패** (`compareEmission`이 err 반환) | `comparison failed: …` |
| 3 | `:155` | 기수 부족 (`compared < len(committed)`) | `compared N/M artifacts — …` |
| 4 | `:162` | 임베드 드리프트 (`len(differing) > 0`) | `… embeds stale agent-emit artifacts …` |

`internal/cli/doctor_agentemit_embed_test.go`에는 이 중 **셋만** 전용 비회귀 테스트를 갖는다 —
`TestAgentEmitEmbed_ExtractionErrorFails`(:139), `TestAgentEmitEmbed_PartialExtractionFails`(:155),
`TestAgentEmitEmbed_DriftFailsAndNamesPath`(:162). **:146 에는 어떤 전용 테스트도 없다.**

이 공백은 카드 t346보다 **앞선다** — t346의 diff는 바이너리 부재 분기만 건드렸다. 그러나 t346의 커밋
`c2b51293e`("name all four preserved fail branches")가 CHANGELOG를 개정해 *"all four fail branches
downstream of the binary-present check … still fail exactly as before"* 를 명시적으로 약속했다.
그 결과 **문서는 넷의 보존을 주장하고, 가드는 셋만 세운다**. 넷째 주장은 근거 없는 보존 주장이다
(`verification-claim-integrity.md` §1 — 관측하지 않은 검증 주장).

## 2. 본 세션에서 실측된 지반 (재도출 금지, 인용해 쓸 것)

`compareEmission`(`doctor_agentemit_embed.go:253`)이 non-nil err를 반환하는 경로는 **정확히 하나**다:
추출된 대응물 읽기는 **성공**했는데, 그 다음의 **커밋 산출물** `os.ReadFile(c)`가 실패하는 경우.

```go
got, readErr := os.ReadFile(filepath.Join(extractedDir, base))
if readErr != nil { uncompared = append(...); continue }   // ← err 아님, 기수 부족(:155)으로 흐름
want, readErr := os.ReadFile(c)
if readErr != nil { return ..., fmt.Errorf("read committed %s: %w", base, readErr) }  // ← 유일한 err 경로
```

`committedEmissionSet`은 `filepath.Glob(".../*.toml")`을 쓰는데, Glob은 **디렉터리도 매치한다**.
따라서 커밋 세트 항목이 `<name>.toml`이라는 **이름의 디렉터리**이면 `os.ReadFile`이 "is a directory"를
반환한다 — 이식성 있고(POSIX/Windows 공통), chmod 불필요, root 권한 민감성 없음.

프로브 실측 (명령: `go test ./internal/cli/ -run TestT356Probe -v -count=1`, 프로브 파일은 관측 후 삭제):

```
status="fail" msg="comparison failed: read committed manager-git.toml: read <tmp>/internal/template/templates/.codex/agents/moai/manager-git.toml: is a directory"
```

## 3. 요구사항 (GEARS)

- **REQ-DECB-001 (Ubiquitous)** — `internal/cli/doctor_agentemit_embed_test.go`는
  `compareEmission`의 err 반환 경로(`doctor_agentemit_embed.go:146` 분기)를 단독으로 겨냥하는
  테스트 함수를 정확히 하나 보유해야 한다(shall).
- **REQ-DECB-002 (Event-driven)** — **When** 커밋 세트의 한 항목이 `<name>.toml`이라는 이름의
  디렉터리이고 추출 대응물은 정상 파일로 존재하는 픽스처 위에서 `checkAgentEmitEmbedAgainst`가
  호출되면, 그 테스트는 반환된 check의 상태가 `uikit.CheckFail`임을 단언해야 한다(shall).
- **REQ-DECB-003 (Ubiquitous)** — 그 테스트는 **comparison-failure 평결임을 형제 세 분기와
  구별 가능하게** 단언해야 한다(shall). 상태만 보는 단언은 불충분하다 — 형제 세 분기도 동일하게
  `CheckFail`을 낸다.
- **REQ-DECB-004 (Event-driven)** — **When** run-phase가 `:146` 분기를 `uikit.CheckOK` 반환으로
  일시 변형(mutant)하면, 그 테스트는 **실패해야** 한다(shall); mutant 원복 후에는 통과해야 한다.
  두 관측(축자 명령 + 축자 출력)이 완료의 유일한 근거다.
- **REQ-DECB-005 (Unwanted)** — 본 SPEC은 프로덕션 코드를 변경하지 않는다(shall not);
  변경면은 `internal/cli/doctor_agentemit_embed_test.go` 단일 파일의 함수 추가 하나뿐이다.
- **REQ-DECB-006 (Where)** — **Where** 기존 픽스처(`newEmbedFixtureRoot`, `writeFakeBinary`,
  `newExtractedDir`, `staticExtractor`)로 조건을 구성할 수 있는 한, 새 헬퍼를 도입하지 않는다(shall not).
- **REQ-DECB-007 (Ubiquitous)** — 판정 증거는 **primary 체크아웃**의
  `/Users/goos/MoAI/moai-adk-go/.moai/reports/t356/verdict.md`에 기록해야 한다(shall) —
  워크트리 내부의 gitignored 산출물은 폐기 시 유실된다.

## 4. Out of Scope

본 SPEC이 **만들지 않는 것**:

### Out of Scope — 프로덕션 코드 변경
- `doctor_agentemit_embed.go`의 어떤 분기·메시지·시그니처도 수정하지 않는다.
- `committedEmissionSet`의 Glob이 디렉터리를 매치하는 성질은 **결함으로 다루지 않는다** — 본 SPEC은
  그 성질을 테스트 지렛대로 **이용**할 뿐이며, 그 수리 여부는 별도 카드 소관이다.

### Out of Scope — 다른 분기·다른 검사
- `:139`/`:155`/`:162` 세 분기의 기존 테스트는 무수정 보존 대상이며 강화 대상이 아니다.
- 바이너리 부재 skip 분기(SPEC-CI-DOCTOR-BIN-001 소관), not-applicable 분기, `doctor.go`의 레지스트리·
  exit-status 계산은 비접촉.

### Out of Scope — 문서·CI
- CHANGELOG의 t346 항목 문안은 수정하지 않는다 — 본 SPEC은 그 약속을 **참으로 만드는** 쪽이지
  약속을 축소하는 쪽이 아니다.
- `.github/workflows/**`, 커버리지 임계, doctor 검사 신규 등록 일체 비접촉.

### Out of Scope — 검증 범위
- 로컬 전체 스위트(`go test ./...`) 실행. 판정면은 `./internal/cli/...` 한정 + 원격 CI.

## 5. 성공 기준

- 새 테스트 함수 1개, 신규 파일 0개, 프로덕션 diff 0줄.
- mutant RED / 원복 GREEN 두 관측이 `verdict.md`에 축자로 남는다.
- `go test ./internal/cli/... -count=1` 통과, `go vet ./internal/cli/` 무경고.
