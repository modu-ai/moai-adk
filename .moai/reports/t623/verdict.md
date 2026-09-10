# t623 — plan-auditor D7·D8 예제 스크립트 판정 기록

card: t623 · class B · Tier S~M · lane-9
branch: WT-auditor-d7d8-scripts · base: 로컬 develop `1e207c3ff`
출처: `.moai/reports/instruction-audit-20260910-01a089f6/report.md` AC-02 · AC-03 (보고서 기준 main `2213871af`)

## 1. develop 재현 (전제 재측정)

### Claim
보고서의 두 결함은 develop `1e207c3ff` 에서도 그대로 재현된다. AC-02 는 거짓 양성(명시적 조정이 있는 정상 참조를 BLOCKING), AC-03 은 거짓 음성(다른 섹션의 `//go:build` 가 신규 섹션의 누락을 가림)이다. 로컬 사본과 템플릿 사본의 두 스크립트는 바이트 동일하다.

### Evidence
위치(보고서 줄번호는 main 기준이라 문구로 찾음): `git grep -n -E "D7|D8" 1e207c3ff -- .claude/agents/moai/plan-auditor.md internal/template/templates/.claude/agents/moai/plan-auditor.md` → 두 사본 모두 `### Group 7: Cross-SPEC Reconciliation (D7)` 336행, `### Group 8: Cross-Platform Discipline (D8)` 367행.

두 사본 차이: `git diff 1e207c3ff:.claude/agents/moai/plan-auditor.md 1e207c3ff:internal/template/templates/.claude/agents/moai/plan-auditor.md` → D7-1 의 예시 SPEC-ID 한 줄(템플릿은 중립화된 예시 ID)과 `## Retry Loop Contract` 본문만 다르다. D7·D8 스크립트 블록은 차이 없음.

스크립트 추출: 각 사본의 `### Group 7:`·`### Group 8:` 뒤 첫 ` ```bash ` 블록을 awk 로 그대로 뽑고, 자리표시자 `<new-spec.md>` 만 `new-spec.md` 로 바꿨다(`repro/d7-local.sh`, `d7-tmpl.sh`, `d8-local.sh`, `d8-tmpl.sh`). `cmp d7-local.sh d7-tmpl.sh` → `d7_cmp_exit=0`, `cmp d8-local.sh d8-tmpl.sh` → `d8_cmp_exit=0`.

픽스처(`repro/`, 각 픽스처 디렉터리를 작업 디렉터리로 두고 `sh <script>` 실행, 출력은 `out-<fixture>.log`):

| 픽스처 | 내용 | 현재 스크립트 출력 | 판정 |
|---|---|---|---|
| `d7-fp` | superseded SPEC 을 `## Reconciliation` 문단에서 "supersedes ... absorbed" 로 명시 조정 | `BLOCKING: SPEC-FIXTURE-OLD-001 has status=superseded but is referenced without reconciliation`, `exit=0` | **AC-02 거짓 양성 재현** |
| `d7-tp` | 같은 superseded SPEC 을 조정 문구 없이 참조 | `BLOCKING: ...`, `exit=0` | 옳은 BLOCKING (반대 방향 대조군) |
| `d7-live` | status=implemented SPEC 참조 | BLOCKING 없음, `exit=0` | 스크립트가 침묵할 수 있음을 보이는 대조군 |
| `d8-fn` | 1절에 `//go:build !windows`(syscall 없음), 2절에 `syscall.Kqueue`(태그 없음) | 출력 없음, `exit=0` | **AC-03 거짓 음성 재현** |
| `d8-tn` | 같은 절에 `syscall.Kqueue` + `//go:build darwin` | 출력 없음, `exit=0` | 옳은 통과 (반대 방향 대조군) |
| `d8-tp` | 태그 없이 `syscall.Kqueue` 만 | `BLOCKING: SPEC references syscall but no //go:build constraint or EXCL justification`, `exit=0` | 스크립트가 BLOCKING 을 낼 수 있음을 보이는 대조군 |

`d8-fn` 의 무출력이 "스크립트가 아무것도 못 냄" 이 아니라는 것은 같은 스크립트가 `d8-tp` 에서 BLOCKING 을 낸 것으로 선다. `d7-fp` 의 BLOCKING 이 "무엇이든 BLOCKING" 이 아니라는 것은 `d7-live` 의 침묵으로 선다.

### Baseline-attribution
워크트리 `WT-auditor-d7d8-scripts` HEAD `1e207c3ff`, 이 실행.

### Gaps
- 부수 관측: 세 D7 픽스처 모두 새 SPEC 자신의 ID(`SPEC-FIXTURE-NEW-001`)에 대해 `SHOULD: referenced SPEC ... not found` 를 냈다. 픽스처의 새 SPEC 이 `.moai/specs/` 아래에 없기 때문에 생긴 것이며, 실제 감사에서는 새 SPEC 이 자기 경로에 있어 해소되는지는 관측하지 않았다. 이 카드 범위 밖.
- 워크트리 생성 직후 첫 `git merge --ff-only 1e207c3ff` 가 `exit=1` 로 실패했다(출력을 버렸기 때문에 원인 미관측). 곧이은 재시도는 `ff_exit=0` 으로 성공했고 `git merge-base --is-ancestor d060e0d13 1e207c3ff` 는 `exit=0` 이었다.

### Residual-risk
픽스처는 결함의 최소 형태다. 실제 SPEC 본문의 다른 형태(코드 펜스 안의 `syscall`, 제목 없는 문서 등)는 여기서 재지 않았다.

## 2. 수리 초안 — 픽스처 사전 실행 (문서 반영 전)

### Claim
초안 스크립트(`repro/d7-fixed.sh`, `repro/d8-fixed.sh`)는 두 결함을 닫고, 반대 방향 대조군을 망가뜨리지 않는다.

### Evidence
초안의 요지:
- D7: 스크립트가 BLOCKING 을 직접 내지 않는다. retired·superseded·archived 참조마다 `REVIEW:` 한 줄을 내고, 그 SID 가 조정 키워드(revers·supersed·absorb·carve-out)와 같은 문단에 있으면 그 문단을 `reconciliation candidate` 로 함께 출력한다. BLOCKING 결정은 감사자가 그 문단을 읽고 내린다.
- D8: 제목(`^#+ `) 단위 섹션마다 `syscall` 과 `//go:build`·`cross-platform exemption`·`EXCL.*syscall` 을 따로 본다. 같은 섹션에 태그가 없으면 섹션 제목을 붙여 BLOCKING. 입력을 읽을 수 없으면 `GAP:`.

실행 방식은 §1 과 같다(픽스처 디렉터리에서 `sh <script>`, 출력은 `repro/fixed-<fixture>.log`).

| 픽스처 | 초안 출력 | 판정 |
|---|---|---|
| `d7-fp` | `REVIEW: SPEC-FIXTURE-OLD-001 has status=superseded — confirm explicit reconciliation ...` + `reconciliation candidate (paragraph 4): This SPEC supersedes SPEC-FIXTURE-OLD-001: its eviction requirement is absorbed here, and the old SPEC stays retired.`, `exit=0` | BLOCKING 없음 — **AC-02 거짓 양성 닫힘** |
| `d7-tp` | `REVIEW: ...` 만, 후보 문단 없음, `exit=0` | 침묵하지 않음 — 반대 방향(조정 없는 참조) 유지 |
| `d7-live` | REVIEW 없음, `exit=0` | 대조군 유지 |
| `d8-fn` | `BLOCKING: section "## 2. New file watcher" references syscall but carries no //go:build constraint or EXCL justification`, `exit=0` | **AC-03 거짓 음성 닫힘** |
| `d8-tn` | 출력 없음, `exit=0` | 반대 방향(같은 섹션 태그) 통과 유지 |
| `d8-tp` | `BLOCKING: section "## 1. New file watcher" ...`, `exit=0` | 대조군 유지 |
| `d8-gap` (`new-spec.md` 없음) | `GAP: new-spec.md is not readable — D8 was not observed`, `exit=0` | 미관측이 통과로 읽히지 않음 |

세 D7 픽스처의 `SHOULD: ... SPEC-FIXTURE-NEW-001 not found` 는 §1 과 같은 픽스처 부산물이다.

### Baseline-attribution
워크트리 `WT-auditor-d7d8-scripts` HEAD `1e207c3ff`, 문서 미수정 상태, 이 실행.

### Gaps
- 초안은 아직 문서에 반영하지 않았다. 반영 후 두 사본에서 다시 추출해 같은 픽스처로 재실행해야 한다.
- 리드 판정 대기 두 건: (1) D7 스크립트가 BLOCKING 을 내지 않게 되면서 MP-5 산문을 "REVIEW 줄을 읽은 감사자가 BLOCKING 을 낸다" 로 맞추는 방향, (2) C3 사본 `internal/template/templates/.codex/agents/moai/plan-auditor.toml` 에도 같은 스크립트가 들어 있어(D8 제목 363행, grep 381행) `make agents-emit` 재생성이 필요하다는 점.
- D7 키워드 목록이 놓치는 조정 표현(예: "replaces")은 후보 문단으로 뜨지 않는다. 그 경우에도 REVIEW 줄은 남으므로 감사자가 읽어서 판정한다 — 스크립트가 자동 BLOCKING 을 내지 않는 이유다.

### Residual-risk
- D8 섹션 경계는 마크다운 제목만 본다. 코드 펜스 안의 `#` 로 시작하는 줄은 제목으로 오인될 수 있다.
- awk 의 `RS = ""` 문단 모드와 `tolower` 는 POSIX awk 기능이며 macOS 기본 awk 에서만 실행했다.
