# t152 — 허수 SPEC-ID 4건 판정서

베이스: `origin/release/v3.1.1` @ `ca7e15fa2` · 브랜치 `WT-phantom-spec-ids` · Class B (plan 생략)

판정 분기: (a) 유실 → 복원 · (b) 허수 → ID 제거/교정 · (c) 흡수 → 흡수처 ID 로 교정

결론: **(b) 3건 · 판정 정정 1건** — 하나의 결론이 4건에 일괄 적용되지 않았다.

| # | SPEC ID | 위치 | 판정 | 조치 |
|---|---------|------|------|------|
| A | SPEC-CLI-WORKTREE-ADVISORY-001 | `internal/core/project/initializer.go:56`, `internal/cli/wizard/types.go:53` | (b) 허수 | ID 제거 (2곳) |
| B | SPEC-MONITOR-001 | `internal/hook/post_tool.go:166` | (b) 허수 | ID 제거 |
| C | SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001 | `internal/hook/user_decision_capture.go:73` | **(나) 의도적 전방 참조 — 리드 분류 정정** | 무조치 |
| D | SPEC-V3R6-LINT-CLASSIFYPRTITLE-001 | `internal/spec/transitions.go:113` (@MX:REASON) | (b) 허수 | ID 제거 |

---

## A — SPEC-CLI-WORKTREE-ADVISORY-001 → (b) 허수

**근거**

- `.moai/specs/SPEC-CLI-WORKTREE-ADVISORY-001/` 는 현재도, 이력 어디에도 없다.
  `git log --all -- ".moai/specs/SPEC-CLI-WORKTREE-ADVISORY-001"` → 출력 0.
- 픽액스(`git log --all -S 'SPEC-CLI-WORKTREE-ADVISORY-001'`)가 잡는 유입 커밋은 단 하나:
  `7171880a9 feat: complete accumulated internal code (autonomy/navigator/mcp/audit) — build recovery (#1409)`.
  즉 대량 빌드 복구 커밋에서 코드와 함께 들어온 라벨이며, 선행 SPEC 문서가 존재한 적이 없다.
- `.moai/specs` / `.moai/reports` / `.claude` 범위 픽액스 결과는 t149 판정서 1건뿐 — 이 ID 를 정의한 문서는 어디에도 없다.

**흡수처 없음** — `WorktreeAutoCreate` 를 언급하는 실재 SPEC 3건
(`SPEC-CONFIG-KEY-HONESTY-001`, `SPEC-UPDATE-DOC-DRIFT-001`, `SPEC-WEB-CONSOLE-REDESIGN-001`)
은 전부 **감사(audit) 관점의 사후 인용**이다. `auto_create` 리더가 하나뿐임을 지적하는 쪽이지,
이 필드를 납품한 주체가 아니다. 흡수처로 교정할 대상이 없다.

**자립성 확인 (HARD)** — 주석 산문은 ID 없이 자립한다.
`Mirrors the wizard.WorktreeAutoCreate selection; persisted to workflow.worktree.auto_create at init.`
는 무엇을·어디에 쓰는지를 그 자체로 말한다. 따라서 SPEC 소급 작성이 아니라 ID 제거가 맞다.

## B — SPEC-MONITOR-001 → (b) 허수

**근거**

- SPEC 디렉터리는 존재한 적 없다. 이 ID 의 출처는 삭제된 기획 문서
  `.moai/project/plan-v2.5.0.md` (커밋 `fb7fd4379` 로 추가, `e7a11d864` 로 삭제) 의
  **예정 SPEC 목록 표**다: `| SPEC-MONITOR-001 | Task 메트릭 Hook 수집 (로깅만) | HIGH | ... |`.
- 코드는 `4a269de13 feat(hook,docs): implement Phase 2 of moai-adk v2.5.0` 에서
  SPEC 문서 없이 곧바로 구현됐다. 기획서의 ID 가 SPEC 이 되지 못한 채 주석에만 남은 형태다.
- 계획서가 함께 예고한 `internal/cli/insights.go` / `/moai insights` 는 끝내 만들어지지 않았다 —
  이 ID 로 묶인 범위가 통째로 SPEC 화되지 않았다는 방증.

**흡수처 없음** — `SPEC-OBSERVE-HYGIENE-001` 이 이 ID 를 언급하지만
(`spec.md:56`, `plan.md:38`) 어디까지나 **"그 주석이 있는 위치"를 가리키는 좌표**로 쓴다
(`SPEC-MONITOR-001 comment near line 166`). 범위를 흡수한 게 아니라 관측 대상으로 인용한 것이다.

**자립성 확인 (HARD)** — `Collect Agent (formerly Task) subagent metrics.` 는 자립한다.
바로 아래 두 줄(best-effort 성격, Task→Agent 개명 배경)이 나머지 맥락을 이미 담고 있다.

**부수 효과 기록** — `SPEC-OBSERVE-HYGIENE-001` 의 위 두 인용은 이제 스테일 좌표가 된다.
그러나 SPEC·리포트는 착지 시점의 기록이므로 현재 코드에 맞춰 고치지 않는다(고치면 기록이 거짓이 된다).

## C — SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001 → 판정 정정: (다) 아님, (나) 의도적 전방 참조

**리드의 (다) 분류를 정정한다. 무조치가 맞다.**

**근거**

- 주석 자체가 스스로를 미래형으로 선언한다:
  `The detection mechanism is deferred to future SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001.`
  "존재하는 SPEC 을 가리키는 척"이 아니라 "아직 없다"를 명시한 문장이다 — 허위 참조의 정의에 해당하지 않는다.
- 완료된 실재 SPEC 이 이 ID 를 정식 전방 링크로 등록해 두었다.
  `SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001` 의 `acceptance.md:222` **FL-2**,
  `spec.md:149` Forward-link, `plan.md:91`, REQ-RR-006/007 범위 규정이 모두
  "기계적 강제는 이 후속 SPEC 으로 이연"이라고 적는다.
- `SPEC-V3R6-ASKUSER-DECISION-MEMORY-001` 은 한 걸음 더 나아가, 이 주석의 존재 자체를
  **AC-ADM-010 의 관측 증거**로 삼는다(`progress.md:341` — doctrine-honest 마커 4개 존재 확인,
  over-claim 마커 0건). ID 를 지우면 그 AC 의 증거가 사라진다.
- `SPEC-V3R6-ORCH-IGGDA-001/research.md:257` 도 candidate 로 등재.

즉 이 ID 는 **여러 완료 SPEC 이 의도적으로 참조하는 미착수 후속 SPEC** 이다.
제거하면 AP-RR-006(탐지 불가를 정직하게 문서화하라)이 요구한 정직성 자체가 훼손된다.

**전례 기록** — `SPEC-V3R6-ZONE-REGISTRY-PACKAGING-001` 은 이 ID 를 **템플릿 배포본의**
`zone-registry.md:1017` 에서만 제거했다(`plan.md:28`: "drop the ID, keep the rule meaning intact").
사유는 허수 판정이 아니라 **템플릿 중립성 누출 검사(C1)** 다. `internal/hook/*.go` 는 템플릿 배포 대상이
아니므로 그 사유가 이 현장에는 적용되지 않는다.

**카드 지시에 따른 기록** — 이 SPEC 이 끝내 작성되지 않으면, 그때는 (다)로 내려온다.
그 시점의 조치는 ID 제거가 아니라 **전방 링크를 건 3개 SPEC 의 FL 항목과 함께 일괄 정리**여야 한다.

## D — SPEC-V3R6-LINT-CLASSIFYPRTITLE-001 → (b) 허수

**근거**

- 유입 커밋이 정확히 하나: `97ba19d1a fix(spec-lint): classify plan-phase feat as draft (LINT-CLASSIFYPRTITLE)`
  (2026-06-19, 19줄 추가, `internal/spec/transitions.go` 단독 수정).
  커밋 본문이 스스로 밝히듯 "LINT-CLASSIFYPRTITLE **도구 부채**" 를 고친 즉응 수정이며,
  SPEC 문서를 낸 적이 없다. 부채 라벨이 `SPEC-V3R6-` 접두사를 얻어 SPEC ID 처럼 보이게 된 사례다.
- 현재 이 ID 를 참조하는 곳은 이 주석 한 줄뿐(`internal/spec/transitions.go:113`) — 흡수처도, 하류 인용도 없다.

**대조군** — 바로 위 6줄에 있는 형제 주석
`@MX:REASON: SPEC-V3R6-DRIFT-CONVENTION-ALIGN-001 — ...` 의 SPEC 은 **실재한다**
(`.moai/specs/SPEC-V3R6-DRIFT-CONVENTION-ALIGN-001/`). 두 줄이 같은 모양이라고 같은 판정이 아니다.

**형식 근거** — `mx-tag-protocol.md:27,144` 상 `@MX:REASON` 은 WARN/ANCHOR 에 필수인 **사유** 서브라인이고,
SPEC ID 는 별도의 `@MX:SPEC` 서브라인 소관이다. 실재 SPEC 이 없는 이상 `@MX:SPEC` 으로 옮길 수도 없으므로,
사유만 남기는 편집이 프로토콜에 부합한다.

**자립성 확인 (HARD)** — `plan-phase commit must not imply run completion` 은 자립한다.
바로 위 `@MX:NOTE` 와 그 위 4줄 주석이 오분류의 구체적 결과(StatusGitConsistencyRule false-positive drift)까지 이미 설명한다.

---

## 검증

| 검사 | 명령 | 결과 |
|---|---|---|
| 회귀 테스트 의존 없음 | `grep -rn --include='*_test.go' -e 'CLI-WORKTREE-ADVISORY' -e 'SPEC-MONITOR-001' -e 'LINT-CLASSIFYPRTITLE' internal/ pkg/ cmd/` | 출력 0 — 어떤 테스트도 이 ID 문자열을 단언하지 않는다 |
| 빌드 | `go build ./...` | 통과 |
| 정적 검사 | `go vet ./internal/core/project/... ./internal/cli/wizard/... ./internal/hook/... ./internal/spec/...` | 통과 |
| 영향 패키지 테스트 | `go test ./internal/spec/... ./internal/hook/... ./internal/cli/wizard/... ./internal/core/project/... -timeout 300s` | 14개 패키지 전부 `ok` |

전체 스위트는 돌리지 않았다(카드 지시 + 로컬 부하 규율). 전 패키지 판정은 CI 몫.

**템플릿 미러 불필요** — 수정한 4개 파일은 전부 `internal/**/*.go` 로 `internal/template/templates/` 배포 대상이 아니다. `make build` 재생성 사유 없음.

## 미검증 / 잔여

- 남은 판정은 주석 문구 수준의 변경이며 동작 경로를 건드리지 않았다 — 런타임 회귀 위험은 없다고 본다.
- C 의 무조치는 "지금 시점의 판정"이다. 후속 SPEC 이 착수되지 않은 채 오래 남으면 재평가 대상.
- B 의 부수 효과(`SPEC-OBSERVE-HYGIENE-001` 좌표 스테일)는 의도적 존치이며 별도 카드 소관도 아니다.
