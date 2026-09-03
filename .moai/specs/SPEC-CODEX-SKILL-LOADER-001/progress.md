# SPEC-CODEX-SKILL-LOADER-001 — 진행 기록

카드: t452 · 브랜치: `WT-codex-skill-wiring` · 착수 base: `d592b0551`

## §E.1 Plan-phase Audit-Ready Signal

- Tier: M (요구사항 13 / 상한 16, 판정 13 / 상한 16)
- 산출물: spec.md · plan.md · acceptance.md · progress.md
- 착수 실측은 spec.md §A 에 있다. 카드 제목의 전제 반증(§A.1)과 미검증 전제의 분기 설계(§A.5)가 plan-audit 의 주요 판단 대상이다.
- 미해결 질문: 없음. §A.5 의 두 미검증 전제는 질문이 아니라 run-phase 의 측정 의무(REQ-CSL-001·004)로 묶여 있다.

## §E.2 Run-phase Evidence

### 기준 SHA 동결 (plan.md §C.0)

`BASELINE_SHA = c529b2e4aaf5148aee7e6c67649bf392837bbb06`

이 회차의 변경-집합 판정(AC-CSL-005 · AC-CSL-011 · AC-CSL-012)은 전부 이 값을 기준으로
`git diff "$BASELINE_SHA" -- <경로>` 로 결정했다. 맨 `git diff` 는 쓰지 않았다.

### 분기 기록 — **분기 A**

`<repo>/.agents/skills/<name>/SKILL.md` 가 **적재된다.** plan.md M0 분기 기록 1 번에 따라
**분기 A**. 판정 근거가 된 출력 줄 — M0 회차 B(`.moai/reports/t452/m0-runs/B-last.txt`),
R1 만 채운 회차의 최종 메시지에서:

```
moai-probe-agents
```

같은 셀렉터가 무표식 음성 대조 회차 A 에서 0 건이므로, 이 줄은 R1 의 점유에 귀속된다.
전문·회차표·양성/음성 대조는 `.moai/reports/t452/m0-verdict.md` 에 있다.

**`inconclusive` 는 발생하지 않았다.** 그 상태는 AC-CSL-002 에만 열려 있으며(완료 정의),
이 회차에서는 다섯 뿌리 중 셋(R1 · R2 · R3)이 발화해 양성 대조가 충족됐다. 신호 재선택
0 회.

**독립 재확인 (run-phase M2 에서).** M2 프로브의 `codex debug prompt-input` 출력이 스킬
뿌리 표를 그대로 찍었고, 거기에 프로젝트 `.agents/skills` 가 뿌리 `r2` 로,
심은 표식이 `(file: r2/moai-probe-m2skill/SKILL.md)` 로 나타났다
(`.moai/reports/t452/m2-runs/A-prompt-input.json`). M0 은 모델 자기보고로, M2 는 런타임이
조립한 입력 목록으로 같은 사실을 관측했다.

### 마일스톤 증거 경로

| 마일스톤 | 산출 | 증거 |
|---|---|---|
| M0 (상속) | 적재 뿌리 5 개 측정, 분기 A | `.moai/reports/t452/m0-verdict.md` · `m0-runs/` · `m0-preflight.md` · `m0-signal-choice.md` · `m0-fixtures/` |
| M1 | 해당 없음 (분기 B 정지 게이트) | — |
| M2 | `skills` 키 세 성질 = 전부 **확인되지 않음** | `.moai/reports/t452/m2-verdict.md` · `m2-runs/` · `m2-fixtures/` |
| M3 | `skill-loader` 행 → `documented-drop`, 재생성 4 명령 exit 0 | `.moai/reports/t452/m3-verdict.md` · `m3-runs/` |
| M4 | 변경 Go 소스 집합이 비어 판정 불성립 | `.moai/reports/t452/m4-verdict.md` |

### [HARD] 물려받은 드리프트 2 건 — 이 회차의 성과가 아니다

`make agents-emit` / `make build`(REQ-CSL-008 이 의무화한 명령)가 BASELINE_SHA 시점에
이미 존재하던 스테일 상태 두 건을 수리했다. 커밋에는 담되 이 회차의 변경으로 주장하지
않는다. 귀속 근거(baseline 블롭 대조 포함)는 `m3-verdict.md`.

- `internal/template/templates/.codex/agents/moai/sync-auditor.toml` — 중립 `.md` 원본의
  "Export mandate" 문단이 커밋된 `.toml` 에 없었다. 즉 `make agents-emit-check` 는
  baseline 에서 이미 붉었다.
- `internal/template/catalog.yaml` — `sync-auditor` 와 `moai` 두 항목의 해시가 스테일했다.
  두 원본 파일 모두 이 회차에 편집하지 않았다.

### [HARD] 귀속 불가 변경 1 건 — 커밋하지 않았다

`.claude/settings.json` 이 세션 도중 21 행 → 215 행으로 바뀌었다. 세션 첫
`git status --porcelain` 에서는 깨끗했고, mtime(18:11:48)이 M3 세 명령의 산출물
(18:15:11 / 18:15:28 / 18:15:47)보다 **앞서므로** 그 명령들이 쓴 것이 아니다. 무엇이
썼는지 확정하지 못했다. 스테이징하지 않고 워킹 트리에 남겨 리드 판단에 맡긴다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-03
run_commit_sha: 9a09452ee
run_status: complete-with-one-ac-fail
baseline_sha: c529b2e4aaf5148aee7e6c67649bf392837bbb06
codex_version_observed: "codex-cli 0.152.1"
branch_recorded: A
inconclusive_occurred: false
ac_pass_count: 7
ac_fail_count: 1
ac_not_applicable_count: 5
ac_inconclusive_count: 0
preserve_list_post_run_count: 0
new_warnings_or_lints_introduced: 0
cross_platform_build:
  applicable: false
  reason: "no Go source changed in this run (git diff --name-only $BASELINE_SHA -- '*.go' → 0)"
total_run_phase_files: 4
m1_to_mN_commit_strategy: single-run-commit
```

### 판정 행렬 — 13 개 전부

| AC | 판정 | 근거 |
|---|---|---|
| AC-CSL-001 | PASS | M0 여섯 회차 각각에 명령·출력 전문·exit code·`codex --version`·트리 SHA 가 남아 있고 뿌리마다 적재 여부 판정이 있다 (`m0-verdict.md`, `m0-runs/`) |
| AC-CSL-002 | PASS | **분기 A** 를 §E.2 에 명시하고 근거 줄(`moai-probe-agents`)을 인용했다. 양성 대조 충족(R1·R2·R3 발화) |
| AC-CSL-003 | **해당 없음** | 분기 A 이므로 이 판정의 `Given`(AC-CSL-002 가 분기 B 를 기록했을 때)이 성립하지 않는다. AC-CSL-003 의 적용 조건이 이 사실을 **§E.3 에 명시**하도록 요구하므로, 조용히 생략하지 않고 여기 적는다 |
| AC-CSL-004 | PASS | 존재·값 형태·오값 거동 세 관측 각각에 명령과 출력이 남아 있고, 각각 **확인되지 않음**으로 판정되어 있다(`m2-verdict.md` 판정표). 판정 없이 넘어간 관측 0 건 |
| AC-CSL-005 | **해당 없음** | `git diff "$BASELINE_SHA" -- internal/template/templates/.codex/agents/moai/ \| grep '^+[a-z0-9_]* *=' \| sed 's/^+//;s/ *=.*//' \| sort -u` → 빈 집합(count=0). 이 SPEC 은 새 필드를 하나도 도입하지 않았다. 빈 집합은 통과가 아니므로 해당 없음 |
| AC-CSL-006 | **해당 없음** | AC-CSL-004 가 `skills` 키를 확인하지 못했으므로 `Given` 불성립. AC-CSL-007 과 상호 배타이며 그쪽이 PASS |
| AC-CSL-007 | PASS | `disposition: documented-drop` 이고 rationale 이 관측 명령(`codex debug prompt-input`, spawn_agent 스키마 출력, 두 오값 대조군)·출력 요지(exit 0 침묵)·codex 버전(`codex-cli 0.152.1`) 셋을 담는다 |
| AC-CSL-008 | PASS | `make agents-emit` → `make build` → `make agents-emit-check` → `go test ./internal/template/agentemit/... -count=1` 넷 모두 exit 0, 마지막 출력에 `ok` (`m3-runs/`) |
| AC-CSL-009 | PASS | `skill-loader` 행 rationale 첫 줄에 `codex-cli 0.152.1`. 최상단 `codex_measured_version` 은 `"0.147.0"` 으로 **의도적으로 유지**했고 그 사실과 이유(재측정한 필드는 이 한 축뿐이라 선행 7 개까지 덮으면 갖지 않은 커버리지를 주장하게 됨)를 `m3-verdict.md` 에 기록했다 — 조용한 미갱신과 구분된다 |
| AC-CSL-010 | PASS | `shasum -a 256 ~/.codex/config.toml` 이 프로브 전후 동일(`8adec56f…4ae7`), `find ~/.codex/skills -maxdepth 1 -mindepth 1` 도 동일(`.system`, `hatch-pet`). 프로브가 쓴 `CODEX_HOME` 은 `/tmp/t452-m2/codexhome` 으로 명령 문면에서 확인된다 |
| AC-CSL-011 | **해당 없음** | `git diff --name-only "$BASELINE_SHA" -- '*.go'` → 0 건. 스윕 대상이 비어 교집합의 0 이 아무것도 주장하지 않는다. 셀렉터 생존 대조군은 기록했다(write-family 4771 매치, `skills.config` 44 매치) — `m4-verdict.md` |
| AC-CSL-012 | **FAIL** | 스윕 대상 1 건(`…/.codex/agents/moai/sync-auditor.toml`)으로 1 이상은 충족하나, `SPEC-` grep 이 **1 매치**를 낸다 → 문면상 FAIL. 매치는 60 행 `SPEC: {SPEC-ID}` 로, 보고 서식의 **자리표시자**이며 baseline 의 `.md` 원본·미러에 그대로 있었고 이 diff 가 더한 줄이 아니다. 프로젝트 자체 가드 두 개는 초록이다(아래 주석) |
| AC-CSL-013 | **해당 없음** | 두 대상 집합이 모두 빈다 — 커밋한 프로브 **스크립트** 0 건(이 회차가 커밋한 `m2-fixtures/*.toml`·`SKILL.md` 는 실행 가능한 스크립트가 아니라 데이터다), 새로 추가된 `os.Stat` 사용처 0 건(Go 변경 0). 주어 없는 PASS 는 빈 스윕의 초록이므로 적지 않는다 |

### 완료 정의 대조 (분기 A 경로)

- AC-CSL-003 = 해당 없음 ✔
- AC-CSL-006 과 AC-CSL-007 중 정확히 하나가 PASS ✔ (007 PASS / 006 해당 없음)
- `inconclusive` 는 어느 판정에도 적지 않았다 ✔ (AC-CSL-002 에만 열려 있고 발생하지 않음)

### AC-CSL-012 FAIL 에 대한 판정 주석 — **범위 밖 수정이 필요하다**

FAIL 을 완화하지 않고 그대로 적되, 원인 귀속을 함께 남긴다.

- 실패시킨 토큰은 `SPEC: {SPEC-ID}` 자리표시자다. `git show $BASELINE_SHA:…/sync-auditor.md`
  와 그 템플릿 미러 양쪽에 이미 있었고, 이 회차의 diff 가 더한 줄("Export mandate" 문단)
  에는 없다.
- 프로젝트 자체의 중립성 기구는 **초록**이다:
  `go test ./internal/template/... -run 'TestTemplateNeutralityAudit'` → exit 0,
  `… -run 'TestTemplateNoInternalContentLeak'` → exit 0. 둘 다 `internal/template`
  패키지에서 실제로 실행됐다(`ok  github.com/modu-ai/moai-adk/internal/template`);
  같은 출력의 `agentemit` 줄에 붙은 `[no tests to run]` 은 그 패키지에 대상 테스트가
  없다는 뜻이므로 판정 근거로 쓰지 않았다.
- 즉 이것은 이 작업이 만든 중립성 위반이 아니라, 판정 셀렉터(`SPEC-` 문자열)가
  자리표시자까지 잡는 **과다 매칭**이다. 셀렉터를 좁히는 것은 `acceptance.md` 본문
  수정이므로 이 에이전트의 소관이 아니다 — manager-spec 재위임 대상이며, 리드 판단을
  기다린다.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §E.0 Plan-phase closure — residual risk

Plan-audit ran five rounds: iter-1 FAIL 0.79, iter-2 PASS-WITH-DEBT 0.86, a scoped delta-audit
(debt NOT closed), a confirm-only pass, and a final confirm. The final confirm recorded the
iter-2 debt as debt-closed (N1, N2, N3, X1-X5, Y1 all closed) and allowed plan-phase to close.

Two clauses landed AFTER that final audit and were NOT re-audited:

- AC-CSL-009 gained its `적용 조건` clause, applied verbatim from the auditor's prescribed text.
- The `inconclusive` row's justifying sub-bullet dropped a qualifier that was false for AC-CSL-009
  (finding Y2).

The auditor stated a sixth iteration was not warranted because both are single clauses with
prescribed text, each verifiable by reading the line it lands on. The lane verified both that way
(clause byte-identical to the prescription at `acceptance.md:77`; false qualifier absent, with
`성립` → 5 as a live control) and re-confirmed REQ 13 / AC 13 and `moai spec lint` 0 error.

**Residual risk, stated rather than left implicit**: across this plan-phase, five repair rounds
each planted a new defect, every one of them in a sentence that READ a newly-added rule rather
than in the rule itself. These two clauses had no independent audit. Run-phase should re-read
AC-CSL-009's body and the `inconclusive` row together before M0 — not because a defect is known,
but because the base rate here is 5 for 5 and the lane's own sweep only covers axes the lane
anticipated.
