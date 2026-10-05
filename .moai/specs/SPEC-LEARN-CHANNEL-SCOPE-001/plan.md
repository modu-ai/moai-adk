---
id: SPEC-LEARN-CHANNEL-SCOPE-001
title: "plan.md — 학습 채널 범위의 정직화"
version: "0.1.1"
created: 2026-09-01
tier: S
---

# plan.md — SPEC-LEARN-CHANNEL-SCOPE-001

> 상태축 무관성 규칙에 따라 본 아티팩트는 `status:` 필드를 갖지 않는다 (spec-frontmatter-schema.md § Artifact Statelessness). 본 SPEC의 lifecycle 상태는 spec.md 단일 표면만 읽는다.

## §A. Context

### §A.1 작업 위치와 baseline

- 워크트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t260` · branch `WT-learn-channel-gap` · HEAD `d7ce6c6bd` (v0.1.1 라운드에서 재확인 — plan-phase 시작 시점 origin/develop과 동일, 유지) — 본 plan-phase `git rev-parse` 실행 관측값.
- 카드: t260 (운영자 지정, Tier S). 카드 전문이 범위 권위. 산출물 커밋은 오케스트레이터 소관 — 본 에이전트는 파일 작성만.
- 품질 모드: `quality.yaml` `constitution.development_mode: tdd` — 단, 본 SPEC은 Go 코드 0줄의 문서 전용 카드라 RED-GREEN-REFACTOR가 기계적으로 적용되지 않는다. run-phase 검증은 측정·grep 명령의 관측값으로 대체한다(§E). 이는 문서 전용 SPEC의 기존 선례와 같은 적응이며, "검증 없는 완료"와 다르다 — 모든 AC가 실행 가능한 명령을 가지고, 하중 AC는 spec.md §I에 RED-now 셀(실관측 4요소)을 갖는다.
- v0.1.0 → v0.1.1: plan-audit iter-1 FAIL 0.63 (`.moai/reports/t260/plan-audit-iter1.md`) — D1(writer 2패밀리 재프레임)·D2(패밀리-집합 판정식)·D3(AC-004 범위+스윕 목록)·D4(스테일 카운트)·D5(SHA 핀)·D6(카드 id 배제)·D7(RED-now 셀).

### §A.2 측정 baseline (귀속 요약 — 상세는 spec.md §B.1)

| 대상 | 내 값 (본 라운드 v0.1.1, 2026-09-01, worktree `d7ce6c6bd`) | 타 baseline (귀속 인용) |
|---|---|---|
| 인박스 구성 | 5,942/5,942행 `tool_failure` — `{tool_failure, test_fail}` 집합 밖 0 (jq 전수) | 5,916(오케스트레이터) → 5,919(v0.1.0) → 5,932(감사 iter-1) — append-only 성장, 구성 결론 4중 일치 |
| `test_fail` 행 | `grep -c 'test_fail:' <inbox>` → `0`, exit 1 — 능력은 배선, 이 파일에선 미관측 | 감사 iter-1 동일 결과 |
| writer 구조 | stub 생산자 2개 — `failure_observer.go:77`(`tool_failure:*`) + `:111`(`test_fail:<pkg>:`); 배선 `evidence_writer.go:583-591` `rec.IsTestFail`; 도입 `e70c77576` | 본 라운드 Read 직접 판독 (감사 판독과 일치) |
| 인간 채널 | feedback_*.md 164개, 08-25 이후 145개 (find) | 오케스트레이터·감사 측정과 3중 일치 |

run-phase에서 재검증할 때는 §C의 명령으로 다시 잰다 — 인용값을 재사용하지 않는다. **교훈(v0.1.0 D1): 측정 구성(100%)을 구조 주장(단일 패밀리 독점)으로 승격하면 안 된다** — `head` 절단된 grep 출력이 제2 패밀리를 가렸고, 문서 정직성을 파는 SPEC이 스스로 과소-기술을 저질렀다. §G anti-pattern으로 등록했다.

### §A.3 결정 축 해소 (권고 + 반전 비용) — 본 SPEC의 최고 변동 가능성 결정

> 카드가 준 3축에 대한 plan-phase 권고다. 최종 결정은 plan→run Implementation Kickoff Approval에서 **리드 경유 운영자**가 한다 — 본 에이전트가 아니다. 아래 표가 kickoff 질문의 판단 재료다.

| 축 | 권고 | 근거 (귀속) | 반전 비용 |
|---|---|---|---|
| **(a) 인간 루프를 학습 채널로 인정** | **YES — 문서 선언으로.** 채널 명명 + 자격은 "배선된 어느 패밀리로도 관측되지 않는 관측된 결함 계열, 반복 또는 비용의 증거 동반". 기록 대상은 기존 `feedback_*.md`+`MEMORY.md` 관례 그대로 (REQ-LCS-002) | 실측: 인박스 100% `tool_failure`(test_fail 0행 포함 전체 집합 밖 0) vs 인간 채널 1주 145건 (§A.2). "무엇이 실제로 학습되는지 정직하게 정한다" = 이미 일어나는 일을 문서로 고정하는 것 | 문장 2-3개 삭제 + 템플릿 미러 1회 — **낮음** |
| **(b) 카드/판정서에서 계열 추출** | **기전 아님 — 기존 흐름의 문서화만.** 리드가 카드 진행 기록·판정 보고서를 읽고 교훈을 쓰는 기존 경로를 anchor doc에 명명한다. 자동 추출 파이프라인·새 포맷·제3 패밀리는 REQ-LCS-005로 금지 | 카드 경고("기전 추가가 목적이 아니다") + 주간 145건 생산량이 무기전 루프로도 충분하다는 실측 — 기전은 지연·오탐·유지보수 비용을 더할 뿐이다. 제2 패밀리(`test_fail`)가 이미 배선돼 있다는 사실은 이 권고를 바꾸지 않는다 — 실패 이벤트 수집 능력과 도구-비가시 계열 포착은 별개 문제다 (D1 재프레임 반영) | 문단 1개 삭제 — **낮음**. 기전화 필요성이 실측되면 그때 후속 SPEC (anchor doc의 non-goal에 경계 기록) |
| **(c) 인박스 유용성 범위 문서화** | **YES — 능력+구성 정직형 경계 주장 + 측정 귀속.** "실패 이벤트 스텁 2패밀리만 기록 / 측정 구성 100% `tool_failure` / 도구·테스트 실패 어느 쪽도 아닌 계열은 인간 루프로 흐른다"를 모든 claim 표면에 정렬 (REQ-LCS-001/003/004) | §A.2 실측 + writer 구조 2패밀리 판독. 기존 표면은 능력은 말해도 구성과 경계·인간 루프를 말하지 않는다 (spec.md §B.2) | 되돌릴 것이 없다 — 주장은 오늘 참. 유일 비용은 baseline 신선도인데, AC-LCS-001이 재검증 가능성으로 판정해 부담이 없다 |

kickoff에서 각 축이 뒤집힐 경우의 영향: (a) NO → REQ-LCS-002/004/007과 AC-LCS-002/004/007 폐기, anchor doc 축소, SPEC은 (c) 단독으로 축소 가능. (b) 기전화 요구 → 본 SPEC 범위 밖 — 후속 SPEC으로 분리 권고 (범위 확장을 본 SPEC에 얹지 않는다). (c) NO → 본 SPEC의 존재 이유 소멸, 폐기 권고.

### §A.4 PRESERVE / EXTEND

**PRESERVE (무수정):**

- `internal/**` 전체 — 특히 `internal/hook/failure_observer.go`(kickoff 조건부 1행 주석 예외 제외), `internal/hook/evidence_writer.go`, `internal/graph/**` (리드 경고)
- `.claude/skills/hns-lsel-curator/{drain.sh,session_drain.sh,backlog_check.sh}` 및 그 테스트
- `.claude/rules/moai/core/moai-constitution.md` 본문 (항상 로드 — 토큰 다이어트)
- `.moai/lessons-inbox.jsonl` (append-only 불변) — 워크트리에 없고, 절대 복사해 오지 않는다
- `feedback_*.md` 포맷·`moai-memory.md` 택소노미, frozen applier (`internal/harness/applier.go`)
- 타 SPEC 디렉터리, `.moai/state/`, `.moai/cache/`
- navigator 표면(`moai-workflow-project/{references,scripts}` — 인박스 경로를 읽기-집합 배제 선언으로만 언급; §A.5 배제 사유)

**EXTEND (수정/생성 대상):**

- `.moai/docs/learning-channel-scope.md` — **신규** anchor doc (tracked, `.moai/docs`는 git 추적 + `moai update` 관리 뿌리 밖 확인: `git ls-files .moai/docs` 21파일, CLAUDE.local.md §2.3 삭제 뿌리 목록에 없음)
- `.claude/rules/moai/core/moai-constitution-detail.md` + `internal/template/templates/.claude/rules/moai/core/moai-constitution-detail.md` (미러 존재 확인: 본 plan-phase + 감사 iter-1 — 양쪽 L53 문구 바이트 동일 존재)
- `.claude/skills/hns-lsel-curator/SKILL.md` (dev-only, 미러 없음 — hns 네임스페이스) — :34 스테일 카운트 교체 포함 (D4)
- `CLAUDE.local.md` §28 (로컬 전용 — 템플릿 AGENTS.md에 LSEL 0회 확인: 본 plan-phase + 감사 재확인, 미러 불필요)

### §A.5 claim 표면 스윕 목록 (D3 — AC-LCS-004의 집행 대상 집합, 명시 열거)

REQ-LCS-004 트리거("describes the lessons-inbox's capture scope")에 적중하는 표면 — AC-LCS-004가 판정하는 전체 집합이다:

| # | 표면 | 성격 |
|---|---|---|
| 1 | `.claude/rules/moai/core/moai-constitution-detail.md` § Lessons Protocol | claim 표면 (템플릿 미러 쌍 수정) |
| 2 | `internal/template/templates/.claude/rules/moai/core/moai-constitution-detail.md` | 같은 내용의 미러 (패리티 판정 대상) |
| 3 | `.claude/skills/hns-lsel-curator/SKILL.md` | claim 표면 (:34 스테일 카운트 포함) |
| 4 | `CLAUDE.local.md` §28 (LSEL 운영 섹션) | claim 표면 (로컬 전용) |

**스윕 절차 (M3):** `grep -rln 'lessons-inbox' <worktree>` (`.git`, `.moai/specs`, `.moai/reports` 제외) 재실행 → 결과를 아래 배제표로 분류 → 목록 1-4 전부가 경계 주장 또는 anchor 포인터 보유 확인.

**배제 목록 (사유 명시 — 감사 iter-1 표면 전수 분석 계승):**

| 표면 | 배제 사유 |
|---|---|
| `moai-workflow-project/references/navigator.md:86,208` · `navigator-audit.md:215-218` · `scripts/navigator-audit.sh:27` (양쪽 모두 템플릿 미러 존재) | 인박스의 캡처 범위를 말하지 않는다 — Navigator 자신의 읽기-집합 배제 선언("never reads/writes"). REQ-LCS-004 트리거 비적중. 편집 시 Template-First 캐스케이드(미러 2쌍+make build)만 유발 |
| `.claude/skills/hns-lsel-applier/apply_test.sh:56-58` | 테스트 픽스처 열거(백업/복원 리스트) — 산문 claim 아님 |
| `.claude/lsel/frozen-allowlist.json` | 기계 설정 파일 — 산문 아님 |
| `CHANGELOG.md` | 이력 기록 |
| `internal/hook/{failure_observer,lessons_inbox_test}.go`, curator 스크립트 4종, navigator 테스트 4건, `.claude/workflows/lsel-drain-loop.js` | 코드·테스트·스크립트 — §G 무수정 범위 |
| `moai-constitution.md` 본문 | 리터럴 `lessons-inbox` 토큰 0히트(감사 확인) — 무수정 결정 방어 가능 |

## §B. Known Issues (Tier S 필터 — 관련 항목만)

- **B4 프론트매터**: `created:`/`updated:`/`tags:` 정본명 사용 (spec.md 작성 완료 — snake_case 별칭 0).
- **B6 spec-lint Out of Scope**: `### Out of Scope — <topic>` H3 + `-` 불릿 필수 (spec.md §G 준수).
- **§2.3 (CLAUDE.local.md)**: `.claude/rules/moai`는 `moai update` 때 통째 재배포 — 미러 없는 수정은 유실된다. 반대로 `.moai/docs`는 관리 뿌리 밖이라 anchor doc이 안전하다. 이 비대칭이 표면 선정의 근거다.
- **rule-authoring.md 비용 규율**: `moai-constitution-detail.md`는 항상 로드가 아니지만(detail companion), 성장분이 1,000바이트를 넘지 않게 문장 2-3개로 묶는다. 넘으면 커밋 본문에 크기+근거 명시.
- **B8/B10 작업위생**: 커밋은 오케스트레이터가 명시 pathspec으로 — 본 에이전트는 커밋하지 않는다.
- **B12**: sync는 manager-docs 소관 — 해당 없음.

## §C. Pre-flight (run-phase 시작 전)

```bash
# 1. baseline 재확인 — 스테일이면 정지·보고
git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t260 rev-parse --short HEAD
git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t260 branch --show-current

# 2. 인박스 재측정 (AC-LCS-001의 재검증 명령 — primary 절대경로, 복사 금지)
jq -r '.event_key // "NO_EVENT_KEY"' /Users/goos/MoAI/moai-adk-go/.moai/lessons-inbox.jsonl | cut -d: -f1-2 | sed 's/:$//' | sort -u
grep -c 'test_fail:' /Users/goos/MoAI/moai-adk-go/.moai/lessons-inbox.jsonl

# 3. 편집 대상 존재 확인
grep -c 'lessons-inbox' .claude/rules/moai/core/moai-constitution-detail.md
ls internal/template/templates/.claude/rules/moai/core/moai-constitution-detail.md
sed -n '34p' .claude/skills/hns-lsel-curator/SKILL.md
```

## §D. Constraints

- spec.md §D 전체를 그대로 따른다 (문서 전용+kickoff 조건부 1행 예외·기전 금지·미러 쌍·LSEL 무변경·항상 로드 무수정·포맷 불변·라이브 수치 미러 유입 금지+기존 스테일 수치 교체).
- 위임 제약: 커밋·push 금지(오케스트레이터 소관), AskUserQuestion 금지 — 필요 입력은 blocker 보고로 반환.
- 한국어 본문 + 영어 토큰(REQ/AC/path/command) — language.yaml `documentation: ko` 및 기존 SPEC 관례 준수.

## §E. Self-Verification (run-phase 완료 기준)

> 각 항목은 verification-claim-integrity 5절 형식(Claim/Evidence/Baseline-attribution/Gaps/Residual-risk)으로 보고한다. TDD의 RED-GREEN은 Go 코드 0줄로 부적용 — 아래 측정 명령의 관측값이 검증 층이고, 하중 AC의 RED는 spec.md §I에 이미 관측돼 있다(4요소 셀).

- **E1 — AC 매트릭스**: AC-LCS-001..007 각 행에 명령+실출력 (spec.md §I — RED-now 셀과 짝을 이룬다).
- **E2 — 문서 전용 diff 증명**: `git diff --stat` — markdown(+미러)만; kickoff에서 :80 주석 정정이 채택된 경우에만 `failure_observer.go` 1행 주석이 보인다.
- **E3 — 미러 패리티 + neutrality**: 양쪽 claim 문장 grep 일치 + 미러에 dev-local 값(날짜·카운트·SPEC ID·**card id(t260 포함 전반)**·머신 경로) 0회 (`internal_content_leak_test.go`·template-neutrality 가드 초록 — 변경 템플릿 경로 기준).
- **E4 — 임베드 재생성**: `make build` rc=0 (미러 편집 후 Template-First 사이클).
- **E5 — 일관성 스윕**: §A.5 절차 재실행 — 목록 1-4의 claim 표면 전부 경계 주장 또는 anchor 포인터, 배제 목록은 여전히 비-claim (AC-LCS-004).
- **E6 — 인박스 무손상**: 측정은 읽기 전용 jq/find/grep뿐 — 인박스 행수 변화는 외부 컬렉션의 append뿐임을 전후 카운트로 확인.

## §F. Milestones (의사결정 역전 가능성 순 — 결정이 먼저, 기계 작업이 나중)

- **M1 (Priority High) — 결정 축 착지 + 정준 주장 고정** (flips AC-LCS-001): `.moai/docs/learning-channel-scope.md` 작성. 내용: (1) dated baseline(인박스 tally 명령+카운트+날짜+tree SHA, `test_fail` 0행 관측, 인간 채널 카운트), (2) 능력+구성 정직형 경계 주장(REQ-LCS-001 문장 — 2패밀리·구성 100%·집합 밖 결함 계열 부재), (3) 인간 루프 인정+자격(REQ-LCS-002), (4) 상류 증거원 명명 — 카드 진행 기록·판정 보고서는 "이미 존재하는 증거원"으로 서술(REQ-LCS-005의 문서화 축), (5) 자동 추출 non-goal 경계, (6) §A.5 claim 표면 포인터 목록. 검증: 파일 존재 + §C-2 명령 재실행 일치.
- **M2 (Priority High) — 표면 전파** (flips AC-LCS-002/003/007): `moai-constitution-detail.md` § Lessons Protocol 정밀화(로컬+미러+`make build` — 능력은 유지하되 구성·경계·인간 루프 추가), `hns-lsel-curator/SKILL.md` 채널 범위 노트 + **:34 스테일 카운트("(624 stubs…)")를 anchor 포인터로 교체(D4)**, `CLAUDE.local.md` §28 문구. M1 문장과 발산 없는 표현 사용.
  - **kickoff 조건부 제안 (D1-iii — 본문 결정 아님, kickoff 판정 대상)**: `internal/hook/failure_observer.go:80`의 함수 헤더 주석 1행 정정("records … to usage-log.jsonl"에 :109-111의 인박스 스텁 append 병기). 운영자가 채택하면 제약 1의 단일 예외로 실행하고, 미채택 시 §B.1 관측 기록만으로 종결한다. 채택 여부와 근거는 progress.md §E.2에 기록.
- **M3 (Priority Medium) — 일관성 스윕 + 검증 배치** (flips AC-LCS-004/005/006): §A.5 절차 재실행, §E 전 항목 실행, AC 매트릭스 완성, docs-only diff 증명. 검증: E1-E6 전부 관측값 동봉.

## §G. Anti-Patterns

- **구성을 구조로 읽기 (v0.1.0 자기결함 — D1)**: 라이브 구성 100% `tool_failure`를 "writer가 그것만 배출한다"는 구조 주장으로 승격하지 마라. 배선된 패밀리 집합(능력)과 측정된 구성(경험)은 다른 진술이다. 절단된 grep 출력(`head`)이 제2 패밀리를 가려 이 결함이 생겼다 — 전수 판독 전에 구조 주장을 쓰지 않는다.
- **"테스트 실패도 담는다"는 느슨한 문구의 재생산**: 정확한 형태는 능력+구성 쌍이다 — "2패밀리 배선 / 측정 구성 100% `tool_failure`". 한쪽만 반복하면 v0.1.0의 반대 방향 오류(과소-기술) 또는 과대-기술로 돌아간다.
- **미러·로컬 갈라짐**: 한쪽만 고치면 update 시점에 되돌아가거나(로컬만), 배포판이 낡는다(미러만). 쌍 수정 + `make build`가 M2의 일부다.
- **미러에 라이브 수치·내부 토큰 유입**: 날짜·카운트·SPEC ID·카드 id는 neutrality C-클래스 위반이자 CI 가드 적색. baseline은 anchor doc에만.
- **AC를 영구 불변식으로 쓰기 (v0.1.0 자기결함 — D2)**: 살아있는 제2 패밀리가 있는 채널에 "영구 zero non-X" 판정식을 박으면 무관한 상류 이벤트(어느 세션의 실패한 테스트 1행)가 green을 뒤집는다 — 패밀리-집합형으로 쓴다.
- **문자적 범위 집행 (v0.1.0 자기결함 — D3)**: "referencing"으로 AC 범위를 열면 배제 선언·픽스처까지 편집 대상이 되고 Template-First 캐스케이드가 터진다 — "describing" 트리거 + 명시 목록(§A.5)으로 묶는다.
- **writer 손대기**: "그럼 `test_fail` 패밀리를 없애거나 확장하면 되지" — 카드가 부정한 기전 변경이다. REQ-LCS-005.
- **anchor doc을 인박스 데이터 복제본으로**: 지속 동기화 부채가 된다. dated baseline + 재검증 명령이 정답이고, 카운트 불일치는 AC가 아니라 baseline의 날짜로 설명된다.
- **항상 로드 본문에 문장 추가**: `moai-constitution.md` 무수정 — detail companion이 같은 주제의 상세 표면이다.

## §H. Cross-References

- spec.md: `.moai/specs/SPEC-LEARN-CHANNEL-SCOPE-001/spec.md` (§B 실측·§C REQ·§I AC+RED-now 셀)
- 감사: `.moai/reports/t260/plan-audit-iter1.md` (iter-1 FAIL 0.63 — D1-D7 수정 경로)
- 카드 t260 · 리드 워크트리 규율: CLAUDE.local.md §4.1, `.claude/rules/local/gitflow-lane-protocol.md`
- 선행 SPEC: SPEC-LSEL-LOCAL-EVOLUTION-001 · SPEC-LSEL-DRAIN-STALL-001 (파이프라인 불변식) · SPEC-HARNESS-RATCHET-REWIRE-001 (제2 패밀리 도입, `e70c77576`)
- 룰: `moai-memory.md` · `rule-authoring.md` · `verification-completeness.md` (RED-now 셀 형식) · `spec-frontmatter-schema.md` · `verification-claim-integrity.md`
- 가드: `.github/workflows/template-neutrality-check.yaml` · `internal/template/internal_content_leak_test.go`
