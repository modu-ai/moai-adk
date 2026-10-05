# t408 — 저자에게 낡은 산출물 모양을 지시하던 문서

카드: t408 · 브랜치 `WT-progress-section-map` · 베이스 로컬 develop `2660bcd09`
(t389·t398는 `b7462203a` 기반. 그 사이 t364가 develop에 병합돼 베이스가 앞섰고, fast-forward이라 발산은 없다.)

## Claim

문서 두 곳(및 카드가 지목하지 않은 형제 한 곳)을 정본에 맞게 고쳤고, 문구와 산출물 양쪽이
되돌아가면 붉어지는 가드를 세웠다. **progress.md 오염 축에 대해서는 카드가 적은 발현 기제가
반증됐고**, 그 자리에서 카드가 말한 것과 다른 실제 결함을 측정해 기록한다.

## Evidence

측정 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t408`, HEAD `2660bcd09`(편집 전).
좌표는 세지 않고 grep 출력을 옮긴다.

### 표면 1 — 그리고 카드가 지목하지 않은 형제

```
.claude/skills/moai/workflows/plan.md:37:… /progress.md §F.1) + `Explore` …
.claude/skills/moai/workflows/sync.md:41:… docs-site + progress.md §F.3 + frontmatter …
```

카드는 `plan.md` 하나만 지목했다. 같은 모양이 `sync.md`에도 있다 — 같은 커밋이 만든 같은 문면이다.
`run.md:37`의 Phase Owners 줄은 progress.md 섹션을 아예 이름 붙이지 않아 이 결함이 없다.

정본은 `spec-frontmatter-schema.md` § progress.md Section Map:
`## §E.1 Plan-phase Audit-Ready Signal`(소유자 manager-spec)와
`## §E.4 Sync-phase Audit-Ready Signal`(소유자 manager-docs).
progress.md `§F`는 orchestrator의 Phase 4 Mode Selection 로그이고, 번호 붙은 단계-신호 하위 절을
갖지 않는다. 같은 파일 `:127`에 §F 혼동을 예견한 disambiguation 주석이 이미 달려 있다.

### 표면 2

```
.claude/skills/moai-workflow-spec/SKILL.md:219:[HARD] Every SPEC directory MUST contain all 3 files.
.claude/rules/moai/workflow/spec-workflow.md:140:| S (Simple) | … | **2 files**: spec.md + plan.md (AC inline in spec.md §3) | 0.75 |
```

Tier S를 규정대로 쓰면 SKILL.md를 위반하고, SKILL.md를 따르면 tier 규정을 위반한다.

### 어느 문서가 낡았는가 — 이력으로 판정 (문면 인상 아님)

| 문면 | 도입 커밋 | 날짜 |
|---|---|---|
| plan.md의 `progress.md §F.1` | `d9cce5427` (SPEC-V3R6-AGENT-TEAM-REBUILD-001 M2) | 2026-05-25 |
| 섹션 맵의 `§E.1 Plan-phase Audit-Ready Signal` | `44725c83c` | 2026-07-02 |
| SKILL.md의 `MUST contain all 3 files` | `9110b273c` (initial implementation) | 2026-02-03 |
| Tier 표의 `2 files: spec.md + plan.md` | `c0eb30da6` (SPEC-V3R5-WORKFLOW-LEAN-001) | 2026-05-20 |

양쪽 다 **문서 쪽이 낡았다.** 부수 관측: plan.md 문면을 만든 SPEC이 자기 progress.md도 §F.N으로
레터링했다 — 문서와 산출물이 같은 시점의 같은 관행이다.

### 카드의 발현 주장은 반증됐다

카드는 "`§F.1`로 저술된 progress.md는 구조를 못 찾아 정상 SPEC이 lifecycle drift로 오분류된다"고
적었다. 재측정 결과 **성립하지 않는다.**

era.go가 매칭하는 리터럴은 `§E.2`·`§E.4`·`§E.5`(`ClassifyEra` :137-139)와
`§E.2`·`§E.3`·`§E.4`·`§E.5`(`hasAnyProgressMarker` :216-219)뿐이다. `§E.1`은 **어느 쪽에도 없다** —
섹션 맵 :116이 그렇게 명시한다("only §E.2-§E.5 headings + the two SHA fields are matched").
즉 plan 신호를 `§E.1`로 쓰든 `§F.1`로 쓰든 era 분류에는 **차이가 없다.**
(카드가 준 좌표 136-138 / 204-209는 t382 착지로 137-139 / 216-219로 밀렸다. 결론은 불변.)

`grep -rn '§E\.1\|§F' internal --include='*.go' | grep -v _test.go` → 히트 전부가 SPEC 본문 절을
가리키는 산문 주석이고, 마커로 파싱하는 코드는 없다.

### 그러면 실제 해악은 무엇인가 — 다른 것이고, 측정했다

`moai spec audit --json`(이 트리에서 빌드한 바이너리) + `population_probe.py`:

```
progress.md 스캔                                   592
era-invisible (§E.{2,3,4,5} 헤딩 없음)              125
그중 §F.N으로 단계 신호를 레터링한 탓                  3
§F.N으로 단계 신호를 레터링한 파일 전체               10
```

125개 중 122개는 진짜 구세대 SPEC(V3R2~V3R5)이라 grandfather 보호가 옳게 적용된 것이고,
**실제 오분류는 3건**이다:

| SPEC | 판정 | 근거 |
|---|---|---|
| SPEC-EVIDENCE-CLAIM-INVARIANT-001 | V3R2-R4 | `H-2 (progress.md without §E.* markers)` |
| SPEC-V3R6-AGENT-TEAM-REBUILD-001 | V3R2-R4 | 같음 |
| SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001 | V3R2-R4 | 같음 |

원인은 `§F.1`이 아니라 **run·sync 신호까지 `§F.N`으로 레터링해 `§E.2`/`§E.4`가 아예 없는 것**이다.
(`era-classification.txt` / `spec-audit.json` 원문)

### "오염분만 §E.1로 정정"이 실행 불가능한 이유 — 10건 전수 분류

카드 지시대로 실행할 수 있는 인스턴스가 **하나도 없다**. 각각 다른 이유다:

| 파일 | §F.1 내용 | 정정 시 문제 |
|---|---|---|
| DEVPROT-REQUIRED-001 · FOURDIM-PHANTOM-001 · HARNESS-OUTCOME-CAPTURE-001 | `§F.1 Plan-phase…` | 같은 파일에 정본 `## §E.1`이 **이미 있다** → 개명하면 §E.1 중복 헤딩이 생겨 더 나빠진다 |
| EVIDENCE-CLAIM-INVARIANT-001 · PREPUSH-WIRING-001 · SEC-HARDEN-003 · V3R6-AGENT-TEAM-REBUILD-001 | `§F.1`~`§F.5` 전면 체계 | §F.1만 §E.1로 바꾸면 §F.2·§F.3·§F.4가 형제를 잃는다. 필요한 것은 파일 전체 재레터링이지 한 줄 개명이 아니다 |
| V3R6-V2-V3-CLEAN-REINSTALL-001 | `### §F.1 — Run-phase Commit Chain` (부모 `## §F — Sync-phase Audit-Ready Signal`의 하위) | plan 신호가 아니다. **다른 용법**이며 그 파일 안에서는 일관적이다 |
| STOP-EVIDENCE-WRITER-001 · WEB-CONSOLE-007 | `§F.1 Plan-phase…` | 두 파일 다 자체 §E.* 마커를 갖고 있어 era 판정은 옳게 난다 |

게다가 이 10개는 대부분 이미 닫힌 SPEC의 **기록**이다. 닫힌 SPEC의 진행 기록을 나중 규약으로
고쳐 쓰는 것은 이 카드가 살 수 있는 값이 아니다 — 문구를 고쳐 **더 만들어지지 않게** 하는 것이
이 카드가 사는 값이고, 그것은 했다.

리드 실측(9개 / 문자열 28)과 이 트리 실측(**10개 파일 / 11개 헤딩 / 문자열 29**)의 차이는
베이스가 앞선 것으로 설명된다.

### 수리

- `plan.md` `progress.md §F.1` → `§E.1` (루트 + 미러)
- `sync.md` `progress.md §F.3` → `§E.4` (루트 + 미러) — 카드 범위 밖에서 발견한 형제
- `moai-workflow-spec/SKILL.md`의 고정 개수 단언 → tier별 산출물 목록 + Tier 표 SSOT 지시 (루트 + 미러).
  카드 조언대로 **새 숫자를 세어 넣지 않고** 나열형으로 바꿨다 — 그래야 tier 규정이 바뀌어도
  이 문장이 다시 거짓이 되지 않는다.
- `catalog.yaml` 해시 1건 갱신(`moai-workflow-spec`).

### 회귀 — 대표 mutant 둘 다 걸리는가

`internal/spec/progress_section_letter_guard_test.go` (신규). 카드가 지정한 두 mutant를 각각 다른
테스트가 잡는다.

**뮤턴트 4종을 심어 RED를 관측하고 전부 복원했다:**

| 뮤턴트 | 잡은 테스트 | 관측 |
|---|---|---|
| M1 `progress.md §F.1` 문면 재도입 | 문구 가드 | FAIL — 인용 좌표와 정본을 함께 보고 |
| M2 문면을 고치는 대신 **삭제** | 문구 가드(공허 방지 leg) | FAIL — "correct citation, not a deleted one" |
| M3 고정 개수 단언 재도입 | 산출물 가드 | FAIL — `MUST contain all 3 files` 적발 |
| M4 새 progress.md가 `§F.N` 레터링을 따름 | 코퍼스 래칫 | FAIL — 집합이 10 → 11로 이동, 신규 id를 이름으로 보고 |

M4가 카드의 "문구만 고치고 산출물을 안 세는" mutant에 대응한다. 다만 래칫이 고정하는 것은
"§F.1로 쓰인 것"이 아니라 **"단계 신호를 §F.N으로 레터링한 것"**이다 — 위에서 반증한 대로
전자는 무해하고 후자가 실제 해악의 모양이기 때문이다.

공허 방지: 문구 가드는 파일마다 `progress.md §E.` 인용이 **남아 있음**을 함께 단언하고,
산출물 가드는 Tier 표 SSOT를 이름 부름을 단언하며, 래칫은 스캔 하한 400(실측 592)을 걸어
잘못된 디렉터리를 읽어 조용히 0건을 보고하는 것을 막는다.

### 검증 범위

```
go test ./internal/spec/...      → ok  65.982s
go test ./internal/template/...  → ok  25.779s (template) + ok 0.817s (agentemit)
go vet ./internal/spec/          → 출력 없음
gofmt -l <신규 테스트>            → 출력 없음
```

`go test ./...` 로컬 전면 실행은 하지 않았다. 전 패키지 판정은 CI 몫이다.

## Baseline-attribution

592 / 125 / 3 / 10 / 11 / 29는 전부 이 워크트리, HEAD `2660bcd09` 트리에서 이 세션이 잰 값이다.
era 판정은 `go build -o /tmp/t408-moai ./cmd/moai`로 이 트리에서 세운 바이너리의
`moai spec audit --json` 출력이며, 원문을 `spec-audit.json`으로 반출했다.
리드가 전한 9 / 28은 다른 트리·다른 시점의 값이라 재측정값으로 대체했다.

## Gaps

- **CI 판정 없음** — 브랜치가 미푸시라 이 변경에 대한 CI 판정이 존재하지 않는다. 위 초록은 로컬 darwin이다.
- **3건의 era 오분류는 고치지 않았다.** 별도 결함이고 별도 수리(파일 전체 재레터링)를 요구하며,
  닫힌 SPEC의 기록을 다시 쓰는 일이라 이 Tier S 문서 카드가 단독으로 결정할 사안이 아니다.
  **후속 카드 후보**로 남긴다 — 위 표가 그 카드의 착수 근거가 된다.
- **카드가 부수 관측으로 적은 §E.5 축은 확인만 했다.** `hasAnyProgressMarker`(:219)가 여전히
  `§E.5`를 찾고 스키마 :106은 §E.5 은퇴를 적는데, era.go :173-178이 이를 H-4-legacy 마이그레이션
  창으로 **의도적으로 유지**한다고 주석에 명시한다. 어긋남이 아니라 문서화된 하위호환이다.
- 문구 가드는 이 4개 문서만 본다. 다른 문서가 같은 오귀속을 새로 만들면 잡지 못한다.
  전 저장소 스캔으로 넓히지 않은 것은, 정당한 SPEC 본문 §F 인용과 구별할 판별식을 아직
  측정하지 않았기 때문이다.

## Residual-risk

- 코퍼스 래칫은 집합 상등이라 다른 레인이 `.moai/specs/`에 SPEC을 추가하면 이론상 붉어질 수 있다.
  붉어지는 조건은 그 SPEC이 §F.N으로 단계 신호를 레터링했을 때뿐이라 오탐이 아니라 적발이지만,
  병합 시점에 다른 카드의 산출물이 이 테스트를 건드릴 수 있다는 사실은 남긴다.
- 스캔 하한 400은 오늘 592에서 나온 값이다. 아카이브 정책으로 코퍼스가 정당하게 줄면 재측정해
  갱신해야 하며, 지우면 안 된다.
- SKILL.md 재작성은 Tier 표를 SSOT로 지목한다. 표가 옮겨 가면 이 문장이 다시 스테일해지고,
  가드는 `SPEC Complexity Tier`라는 이름이 남아 있는지만 보므로 그 이동을 잡지 못한다.
