# SPEC-WEB-CONSOLE-016 — Implementation Plan

> 이 계획은 **Implementation Kickoff Approval 게이트의 입력**이다. run-phase 진입은 승인되지 않았다. 축 (4)는 사용자에게 보이는 동작을 바꾸므로 운영자가 그 게이트에서 본다.

## §A. Context

카드 t1049 의 1차 측정(트리 `WT-partial-apply` @ `d8304b49a`)이 `handleSave` 의 실패 보고 표면에서 두 결함을 확립했다: 실패 페이지가 제출값을 디스크값인 양 되비추는 것(축 4), 그리고 배너가 아는 것보다 넓게 주장하는 것(축 5). 코드 변경은 0이고 계측기 한 개(`internal/web/partial_apply_repro_test.go`)가 커밋됐다.

`development_mode` 는 `quality.yaml` 을 따르되, 본 SPEC 의 성질상 **기존 거동을 바꾸는 작업**이므로 신규 거동은 RED→GREEN, 기존 경로는 characterization 보존으로 간다.

## §B. Tier 결정 + 정당화

**선택: Tier M (standard).**

**Tier S 가 아닌 이유** — 단일 파일·단일 라인 편집이 아니다:

1. **새 뷰-모델 상태가 필요하다.** "어느 단계가 들어갔는가"는 현재 `pageView` 에 존재하지 않는 정보다(`handlers.go:95-96` 은 `Banner`/`BannerKind` 스칼라 두 개뿐). 단계 결과를 축적해 뷰로 옮기는 경로를 신설해야 한다.
2. **호출 지점이 여섯 곳이다.** `renderErrorPage` 호출은 `:466`·`:486`·`:495`·`:504`·`:514`·`:524`·`:534`·`:547` 에 흩어져 있고, 각각 문구가 다르다. 한 곳을 고치고 끝나지 않는다.
3. **렌더 경로가 갈린다.** 성공 경로(`successProjectView` 선언 :564)는 디스크 재독을 하고 실패 경로(`renderErrorPage` 선언 :600)는 하지 않는다. 두 경로를 수렴시키면 템플릿 표면(`root.templ`)도 함께 움직인다.
4. **미측정 두 단계의 탐침 신설.** 7·9단계는 파일시스템 수준 탐침이 필요하다 — 기존 하네스의 확장이지 재사용이 아니다.
5. **AC 11개**, 신규 테스트 다수, 다중 파일(handlers.go + view 구조체 + templ + 테스트 2종).

**Tier L 이 과대인 이유**: 단일 패키지(`internal/web`)에 머물고, 새 config 섹션·validator·서브시스템이 0이며, 영속화 층(profile store / config manager / yamlpatch)을 건드리지 않는다. 마일스톤 3개로 충분하고 ≥3 마일스톤 AND ≥10 파일 조건을 채우지 않는다.

## §C. Pre-flight (run-phase 진입 전 재측정 의무)

plan-phase 좌표는 스냅샷이다. run-phase 착수 시 content-token 기준으로 다시 잰다.

```bash
git rev-parse --short HEAD && git branch --show-current
grep -n "func (a \*app) handleSave\|func (a \*app) renderErrorPage\|func (a \*app) successProjectView" internal/web/handlers.go
grep -c "a.renderErrorPage(" internal/web/handlers.go        # 현재 8 (실측)
go test ./internal/web/ -count=1                              # baseline GREEN (t1049 실측 EXIT 0, 25.2s)
go test ./internal/web/ -run TestPartialApplyOrder -v -count=1 # 계측기 GREEN + 단계 순서 로그
```

## §D. Constraints (HARD)

spec.md §4 HARD-1..6 전부. 특히:

- **HARD-1** 영속화 단계의 순서·개수·존재 무변경. 설계가 변경을 강제하면 → STOP + blocker(트랜잭션 축 침범).
- **HARD-3** 비밀값 비노출. 층 이름 + 파일 경로만.
- **HARD-4** `recordingSeams` 의 seam 목록/시그니처 보존(t1051 공유).
- **HARD-6** 배너는 Go 리터럴 관례 유지 — i18n 이전 금지.

## §E. Self-Verification (각 마일스톤 GREEN 기준)

- `go test ./internal/web/ -count=1` EXIT 0
- `go test ./internal/web/ -run TestSaveSyncFailureSurfacesReadableError -count=1` GREEN **무수정** (REQ-WC16-011)
- `go test ./internal/web/ -run TestPartialApplyOrderPositiveControl -count=1` GREEN (하네스 생존 — REQ-WC16-010)
- `templ generate && git diff --exit-code internal/web/*_templ.go` (템플릿을 만졌을 때만)
- **양성 대조 의무**: 각 신규 테스트는 "현재 코드에서 실패하는가"를 먼저 보인다. 현재 코드에서 통과하는 AC 는 AC 의 결함이다(acceptance.md §D.2).

## §F. Milestones (되돌리기 비싼 결정 우선; 시간 추정 없음)

가장 바뀌기 쉬운 결정(사용자 가시 동작 + 새 자료구조)을 앞에 두고, 기계적 정리를 뒤로 뺀다.

| Mn | Objective | 왜 이 순서인가 | Key files |
|----|-----------|----------------|-----------|
| **M1** | **사용자 가시 표면 설계 확정** — 실패 페이지가 "무엇이 들어갔고 무엇이 안 들어갔는가"를 어떤 형태로 보이는가(단계 목록? 층 단위 요약?), 그리고 "unverified" 를 어떻게 표시하는가. 문구 초안 + 마크업 스케치. | 되돌리기가 가장 비싸다. 자료구조와 문구가 여기서 갈리며, 운영자가 Kickoff 게이트에서 실제로 보는 것이 이것이다. 코드를 먼저 쓰면 이 결정이 구현에 끌려간다. | (설계 산출물; `root.templ` 스케치) |
| **M2** | **단계-결과 자료구조 신설 + 디스크 재독 수렴** — 영속화 진행 상태를 축적하는 타입, `pageView` 확장, `renderErrorPage` 가 `successProjectView` 와 같은 재독 seam 을 쓰도록 수렴. 재독 실패 시 unverified 저하(REQ-WC16-004). RED→GREEN. | 새 타입 인터페이스는 이후 전부가 의존한다. 늦게 바꾸면 테스트까지 함께 무너진다. | handlers.go, root.templ, handlers_test.go |
| **M3** | **배너 정밀화** — 여덟 개 호출 지점의 문구를 관측 범위에 맞춘다. 1단계 실패 시 "아무것도 저장되지 않음"(REQ-WC16-008), `settings saved` 과잉 주장 제거(REQ-WC16-007). Advisory D1 가드 테스트 무수정 GREEN 확인. | 문구는 M2 의 자료구조가 무엇을 아는지에 종속된다. | handlers.go, handlers_test.go |
| **M4** | **미측정 2단계 탐침** — 7·9단계를 파일시스템 수준(대상 경로 쓰기 불가)으로 실패시키는 탐침 추가, `recordingSeams` 는 확장하되 시그니처 보존. 비밀값 비노출 단언 포함. RED→GREEN. | 기존 하네스 확장이며 앞 결정에 의존한다. | partial_apply_repro_test.go |
| **M5** | **회귀 가드 + 정리** — 패키지 전량 GREEN, 계측기 양성 대조 GREEN, 중복 문구 정리(REFACTOR). | 순수 기계적. | 전체 |

> 마일스톤 5개, 단일 패키지 — Round 분할 불요.

## §F.1 Open questions — Kickoff 게이트 이전에 해소해야 할 것

세 항목은 plan-phase 에서 확정하지 않았다. 운영자 결정이 필요하며, plan-auditor 는 이 마커들이 남아 있는 한 Kickoff 승인 전 해소를 권고한다.

- **[NEEDS CLARIFICATION: 실패 페이지의 "무엇이 들어갔는가" 표현 형태]** — 아홉 단계를 그대로 나열할 것인가(내부 구현 누설), 세 영속화 층 단위로 요약할 것인가(프로필 / 프로젝트 설정 / 에이전트·자격증명), 아니면 필드 단위 "저장됨 / 저장 안 됨" 배지인가. M1 의 산출물이며 사용자 가시 표면이다.
- **[NEEDS CLARIFICATION: 디스크 재독의 표면 범위]** — 실측상 읽기 seam 세 개가 이미 존재한다(`readPreferences` app.go:39, `readProjectConfig` :61, `readProjectNestedConfig` :68), 그리고 스키마 10섹션은 실패 경로에서도 이미 디스크를 읽는다(`applySchemaCurrentBestEffort` handlers.go:593). 남는 결정은 (a) 프로필 preferences + 두 스칼라만 디스크 재독으로 전환할 것인가, (b) nested 표면까지 성공 경로와 수렴시킬 것인가, (c) 에이전트 override / GLM 자격증명 표면은 재독 대상에서 어떻게 처리할 것인가(전자는 대응 read seam 미확인, 후자는 HARD-3 상 값 표시 자체가 금지).
- **[NEEDS CLARIFICATION: "unverified" 의 사용자 표기]** — 재독 실패 시 값을 비울 것인가, 마지막 알려진 값에 경고 배지를 붙일 것인가, 섹션을 통째로 접을 것인가. REQ-WC16-004 는 "제출값 폴백 금지"만 정하고 표기는 정하지 않는다.

## §G. Anti-Patterns (회피 목록)

- **트랜잭션을 "김에" 도입** — HARD-1/HARD-2 위반. 크기가 작아 보이는 롤백도 3층을 가로지르는 순간 별건이다.
- **제출값을 남겨 두고 배너로만 경고** — 축 (4)를 축 (5)로 대체하는 것. 화면이 여전히 거짓을 말한다.
- **디스크 재독 실패를 제출값으로 폴백** — REQ-WC16-004 위반. 저하를 진실로 위장한다.
- **`recordingSeams` 를 새 하네스로 교체** — t1051 을 깬다(REQ-WC16-010).
- **배너에 실패 원인의 원문 에러를 그대로 실어 비밀값 유출** — HARD-3. 현재 문구는 `err.Error()` 를 직접 잇는다(`:534`, `:547`), 여기가 위험 지점이다.
- **7·9단계를 "재지 못했으므로 안전"으로 처리** — 미측정과 부재의 혼동(spec.md §6).
- **현재 코드에서 이미 통과하는 AC 작성** — 반증력 0.

## §H. Cross-References

- spec.md §1(측정) / §4(HARD) / §5(인벤토리) / §6(Gaps) / §F(Exclusions)
- acceptance.md — AC-WC16-001..011 + 각 AC 의 "현재 무엇 때문에 실패하는가"
- `.moai/reports/t1049/findings.md` — 5절 증거 기록
- `internal/web/partial_apply_repro_test.go` — 재사용 대상 하네스
