# SPEC-ALWAYS-LOADED-DIET-002 — 리서치

**출처**: 이 문서는 `.moai/reports/t1175/discovery.md` 에서 파생된다. 그 보고서는 이 실행, 이 트리(`.claude/worktrees/t1175`, base develop `a0b78213d`)에서 수행된 조사이며, 여기서 재실행하지 않고 인용한다. 원 보고서에 없는 측정은 §5 에만 있고 출처를 밝힌다.

---

## 1. 계량기의 동정 — 누가 경고를 내는가

경고를 내는 것은 **Claude Code 런타임**이다(`/memory` 처방이 런타임 어포던스). 단위는 문자, 한도는 합계 150,000 이며 둘 다 경고 문구 자체에 적혀 있다.

MoAI 는 이 경고를 내지 않는다는 것이 실측으로 확인됐다:

- `moai hook instructions-loaded` (`internal/hook/instructions_loaded.go:86-104`)는 **파일당 40,000자**만 보고 합계는 보지 않는다. 최대 always-loaded 파일이 39,077자이므로 현재 침묵 중이다.
- `internal/config/defaults.go:943`·`:1243` 의 `150000` 두 개는 `AutoClear.TokenThreshold` 와 `TokenBudget.SkipIfUsageAbove` — 세션 컨텍스트 사용량의 **토큰** 임계값이며, 지시 파일 크기와는 다른 축이다. 숫자가 같은 것은 우연이다.

공식 Claude Code 문서에는 지시문 합계에 대한 수치 경고 임계값이 기록돼 있지 않다. 기록된 것은 CLAUDE.md 당 200줄 가이드라인과 파일당 4 MiB 하드 컷오프뿐이다.

## 2. 경고의 표면 — 18개가 무엇인가

always-loaded 룰 14개 + `CLAUDE.md` + `AGENTS.md` + `@`-import 설정 2개(`user.yaml`, `language.yaml`). 출력 스타일은 포함되지 않는다.

이 트리 실측: **246,943자**. 한도 150,000. 요구 감축 **96,943자(39.3%)**.

## 3. Gap — 경고를 낸 트리는 식별되지 않았다 (이월)

`discovery.md` §3 이 명시한 미검증 항목을 **그대로 이월한다**(드롭하지 않는다).

경고의 파일별 수치가 `main` 에서도 `develop` 에서도 재현되지 않는다:

| 파일 | 경고 | main | develop |
|---|---:|---:|---:|
| `kanban-dispatch.md` | 36.1k | 32,540 | 39,077 |
| `agent-common-protocol.md` | 27.5k | 25,130 | 28,381 |
| `askuser-protocol.md` | 24.3k | 21,597 | 24,466 |

세 값 모두 두 브랜치 **사이**에 있어 제3의 트리(약 100개의 라이브 워크트리 중 하나, 또는 중간 develop 커밋)를 가리킨다. 바이트 수로도 재현되지 않는다(main: 32,800 / 25,279 / 21,822).

트리 식별은 **한정된 시도 후 포기됐다** — 워크트리 격리 가드가 git 루프를 거부해 20회의 개별 호출이 들고, 답이 목표를 움직이지 않기 때문이다. 150,000 은 선언된 한도이지 도출된 값이 아니며, 이 카드가 고쳐야 할 트리는 배포되는 트리다.

**이 Gap 의 잔여 위험**: 목표에 대해서는 없다 — 경고가 보고한 합계(244.7k)와 이 트리의 실측(246,943)은 0.9% 차이다. 다만 "경고가 사라졌는가"를 직접 확인할 수는 없고, 확인되는 것은 "이 트리의 18파일 합계가 150,000 미만인가"뿐이다. AC 는 후자만 판정한다.

## 4. 선행 작업 — 재발명하지 않는다

- **분리 규약**: `SPEC-ALWAYS-LOADED-DIET-001` REQ-ALD-003(companion `paths:` 는 domain-keyed; 유일 선례 `goal-directive-detail.md`), REQ-ALD-004(stub 포인터가 옮긴 절을 이름으로 호명 + companion 자기 소유 경계 선언 + stub 푸터 버전 줄), REQ-ALD-005(companion 은 원본에 없던 내용을 획득하지 않는다 — `session-handoff-examples.md` 40,891 B 가 부모 23,251 B 보다 크다).
- **always-loaded 표면 정의와 scope-first 의무**: `.claude/rules/moai/development/rule-authoring.md:14,32`.
- **구속 조항 재배치 금지**: `SPEC-AGENTS-MD-CANON-001` REQ-AMC-002, `spec.md:475`.
- **기존 예산 가드**: `TestAlwaysLoadedTokenBudget` (`internal/config/token_budget_guard.go:86`, `AlwaysLoadedTokenBudget = 77600`). 그 `@MX:DEBT` 가 76,000 → 77,600 상향 사슬이 **이 다이어트를 대신 서 있는 것**이라고 명시하고, `@MX:UPGRADE` 가 이미 대상을 실측치와 함께 호명한다. 표면과 단위가 달라 이 카드의 판정 대상은 아니다(spec.md §D).

## 5. 이 SPEC 이 추가로 잰 것

`discovery.md` 는 재배치 가능 풀을 **집계**로만 냈다(173,286). 파일별 목표를 세우려면 파일 단위 분해가 필요해 같은 방법(문단 단위 분류)을 라이브 트리의 18파일에 적용했다. 결과는 `design.md` §1 에 있고 합계는 **172,363** — discovery 의 173,286 과 0.5% 차이이며, 차이의 원인은 discovery 가 룰 14개를 템플릿 트리에서 잰 반면 여기서는 라이브 트리에서 쟀기 때문이다.

추가로 잰 것 하나 더 — **구속 조항 줄 기준선**. 16개 마크다운 파일에서 `[HARD]` / `MUST` / `shall ` 중 하나를 담은 줄을 뽑아 정규화·정렬한 결과:

```
170 lines
sha256 d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3
```

토큰 출현: `[HARD]` 128, `MUST` 57, `MUST NOT` 23, `shall` 1. `shall` 이 사실상 부재한 것은 이 룰 트리가 `[HARD]`/`MUST` 어휘를 쓰고 `shall` 은 SPEC 본문 어휘이기 때문이다 — REQ-AMC-002 의 세 토큰 중 하나가 이 표면에서 거의 쓰이지 않는다는 사실 자체가 관측이며, 그래서 검사는 세 토큰 모두를 계속 훑는다.

## 5.1 두 번째 축 — 파일당 40,000자 (v0.2.0 에서 편입)

`discovery.md` §1 은 이 한도를 **다른 계량기의 부재를 확인하는 맥락**에서만 언급했다 — "MoAI 는 합계 경고를 내지 않는다, 파일당 40,000자만 본다, 최대 always-loaded 파일이 39,077 이라 현재 침묵 중"이라는 negative result 였다. always-loaded 표면만 보면 그 진술은 참이다.

리드가 이 축을 범위에 넣은 뒤 소스를 직접 읽어 범위를 다시 쟀다(`internal/hook/instructions_loaded.go:86-104`). `checkCharacterBudget` 는 `utf8.RuneCount` 로 **로드되는 파일마다** 재고 40,000 초과 시 위반을 낸다 — 디스크 상태가 아니라 **로드 시점**에 발화하므로, path-scoped companion 도 로드되면 똑같이 걸린다. 따라서 이 축의 대상은 always-loaded 14개가 아니라 **룰 86개 전부**다.

이 재측정이 discovery 의 negative result 를 뒤집지는 않는다 — always-loaded 파일 중에는 여전히 초과가 없다. 바뀐 것은 **어느 집합을 봐야 하는가**이고, path-scoped 를 포함하면 두 트리 모두 4개가 초과다(2026-09-25 실측):

```
61435  workflow/worktree-integration.md
41616  workflow/session-handoff-examples.md
41036  workflow/kanban-dispatch-detail.md
40799  workflow/spec-workflow.md
```

`spec-workflow.md` 는 이번 세션에서 실제로 위반 메시지를 냈다 — 이 한도는 이론이 아니라 발화 중인 가드다.

**이 측정이 설계를 뒤집은 지점**은 `design.md §2` 에 기록했다. 요지: 계획된 유입량을 목적지에 더하면 셋이 한도를 넘기고, 그중 `agent-common-protocol-reference.md`(31,207 + 15,000 = 46,207)는 **이 카드가 새로 만드는 초과**다. 이미 초과한 파일만 훑는 방식으로는 보이지 않으므로, 목적지 14개를 전부 실측해야 드러난다.

## 6. 기각된 접근

`kanban-dispatch.md`(39,077자, `[HARD]` 38개)를 always-loaded 표면에서 떼어 `-k`/`-f` 세션에만 SessionStart `AdditionalContext` 채널로 주입하는 방안(`internal/hook/handoff_inject.go:126`, 모드 감지는 `internal/hook/session_start_kanban.go:95` 에 이미 존재)은 기계적으로 가능하다. **기각** — REQ-AMC-002 가 금지하고, 재배치 가능 풀이 요구치의 1.8배이므로 필요하지도 않다. 재제안되지 않도록 기록한다.

🗿 MoAI
