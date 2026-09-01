# SPEC Review Report: SPEC-EVIDENCE-CITATION-CANON-001

Iteration: 1/2 (Tier M ceiling)
Verdict: **FAIL**
Overall Score: **0.70** (Tier M PASS threshold 0.80)

측정 트리: 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t375`, 브랜치 `WT-state-evidence-canon`, HEAD `b64043481`. primary 체크아웃(`/Users/goos/MoAI/moai-adk-go`)에서 잰 값은 그렇게 표시한다. 모든 수치는 이 감사가 이 트리에서 직접 실행한 명령의 출력이다.

Reasoning context ignored per M1 Context Isolation — 오케스트레이터가 사전 검증했다고 전달한 6건은 재측정하지 않았고(지시대로), 나머지는 전부 이 감사가 스스로 쟀다.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -o 'REQ-ECC-[0-9]*' spec.md | sort -u` → `REQ-ECC-001` … `REQ-ECC-012`, 12건 연속, 결번·중복 없음, zero-padding 일관.
- **[PASS] MP-2 GEARS format compliance** — 요구 층(`REQ-XXX`, spec.md §2)에 대해 판정. 001/004/005/007/008/009/010/011 = ubiquitous(`shall` / `shall not`), 003 = event-driven(`When 어떤 산출물이 … 인용될 때`), 006 = Where(`Where 소비자가 기계인 경우`), 002·012 = ubiquitous + unwanted 복합. 12건 모두 다섯 패턴 중 하나에 대응. 검증 층(`AC-XXX`, acceptance.md)은 Given-When-Then이고 이는 올바른 형식이므로 여기서 감점하지 않았다(Group 4에서 별도 채점).
- **[PASS] MP-3 YAML frontmatter validity** — 12개 정본 필드 전부 존재하고 타입이 맞다: `id`/`title`/`version`("0.1.0", 인용)/`status`(draft)/`created`·`updated`(2026-08-31, ISO)/`author`/`priority`(P2)/`phase`/`module`/`lifecycle`(spec-anchored)/`tags`(쉼표 구분 문자열). 거부되는 snake_case 별칭(`created_at`/`updated_at`/`labels`/`spec_id`) 없음. `tier: M` 추가 필드 존재(허용).
- **[N/A] MP-4 language neutrality** — moai-adk-go 자체(Go 단일 언어) 저장소 규약을 고치는 SPEC이고, 16개 프로그래밍 언어 도구를 열거하는 문맥이 아니다. 템플릿 미러 편집은 경로 형태(`.moai/reports/<card-id>/`)뿐이라 언어 편향이 발생하지 않는다.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — `grep -hoE 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' *.md | sort -u` → 자기 자신 + `SPEC-EVIDENCE-CLAIM-INVARIANT-001` 1건. 후자는 `.moai/specs/SPEC-EVIDENCE-CLAIM-INVARIANT-001/spec.md` 존재, `status: completed` — retired/superseded/archived 아님. BLOCKING 없음.
- **[N/A] MP-6 D8 cross-platform discipline** — `grep -c 'syscall' *.md` → 4개 파일 전부 0. 자동 PASS.
- **[N/A] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-EVIDENCE-CITATION-CANON-001/` → rc=1, 매치 0. `research.md`는 부재(Tier M 입력 계약상 불필요).

**must-pass 실패 없음.** FAIL 판정은 총점(0.70 < 0.80)과 아래 blocking 결함 2건이 이 SPEC의 중심 기제(가드)를 무력화한다는 사실에서 나온다.

---

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | 산문은 이례적으로 명확하고 자기비판적이다(§3.1의 출처 등급 표기, AC-ECC-006의 한계 명시). 다만 세 지점이 해석을 요구한다: 허용목록의 **단위**가 미지정(D7), `.gitkeep` 생성 여부 미결(D8), Tier 파일 수 범위가 자기 §C의 최대 케이스와 불일치(D12). |
| Completeness | 0.80 | 0.75–1.0 경계 | HISTORY/배경/요구/판정/부채/Out of Scope 전부 존재. Out of Scope는 `### Out of Scope — <주제>` H3 3개 + 각 구체 불릿 ✓. frontmatter 12/12 ✓. 감점 사유: 에이전트 정의·출력 스타일 표면을 구속하는 REQ가 없다(D10) — spec.md §1.3이 "가장 강한 사례"라 부른 파일이 요구 층에 대응 항목을 갖지 못한다. |
| Testability | 0.55 | 0.50 | AC 14건 중 3건이 판정력을 갖지 못한다: AC-ECC-002의 명령이 **오늘 이미 통과**하고(D6), AC-ECC-010의 파일 수 하한이 실제 모집단보다 두 자릿수 작으며(D1), REQ-ECC-008(양쪽 트리)을 판정하는 AC가 없다(D2). 나머지 11건은 이진 판정 가능하고, AC-ECC-001/003/004/005의 grep이 오늘 트리에서 올바르게 RED임을 확인했다. |
| Traceability | 0.70 | 0.75 미달 | REQ 12건 전부 최소 1개 AC에 매핑됨 ✓(§D 매트릭스 대조 완료). 그러나 AC-ECC-007/008/009 3건이 **REQ가 아니라 plan.md §C.2/§C.3에 매핑**되어 요구 층에 근거가 없고(D10), REQ-ECC-008에 매핑된 유일한 AC(AC-ECC-010)가 그 요구를 판정하지 못한다(D2). |

**총점 산출**: (0.75 + 0.80 + 0.55 + 0.70) / 4 = **0.70**. Tier M PASS 임계 0.80 미달.

---

## Defects Found

### D1 — 가드의 반(反)공허 장치가 그 자체로 공허하다 [BLOCKING]

- **파일**: `acceptance.md` AC-ECC-010 / `plan.md` §D.4-1, §G 첫 불릿
- **주장**: "스캔한 파일 수가 **7 이상**이고 위반이 **0**"이면 범위 붕괴를 검출한다.
- **실측 (이 트리)**:
  - `find .claude/rules .claude/agents .claude/output-styles .claude/skills -name '*.md' -type f | wc -l` → **363**
  - `find internal/template/templates/.claude -name '*.md' -type f | wc -l` → **340**
  - 가드가 명시한 범위(D.1: 두 트리) 모집단 = **703**
- **실패**: 하한이 모집단보다 **두 자릿수 작다**. 누가 범위를 `.claude/rules/moai/core/` 하나로 좁혀도 스캔 파일 수는 7을 훨씬 넘으므로 하한이 걸리지 않고, 위반 0이 나와 초록이 된다. plan.md §G가 이름 붙인 "공허한 초록" 위험을 막으라고 세운 장치가 정확히 그 위험을 통과시킨다.
- **파생 오도출**: 하한 7의 근거로 제시된 "오늘 이 범위의 위반 파일이 정확히 7개"는 **저장소 루트 사본만** 센 값이다. `grep -rl 'state/verify' --include='*.md' .claude/` → 7, `grep -rl 'state/verify' internal/template/templates/` → **.md 7 + .toml 1**. 가드가 두 트리를 본다면 오늘의 위반 파일은 **14**이지 7이 아니다. 게다가 "위반 파일 수"와 "스캔 파일 수"는 서로 다른 모집단인데 전자로 후자의 하한을 정당화했다 — M1·M3 수리 후 위반은 0이 되므로 7은 어떤 관측량과도 연결되지 않는다.
- **Severity**: critical · **Class**: blocking
- **Required fix**: (a) 하한을 실제 모집단에서 도출한 값으로 바꾼다(예: 루트 ≥ 300, 미러 ≥ 300 — 두 트리 **각각**에 대해 단언). (b) "위반 파일 수"를 하한 근거로 쓰지 않는다. (c) 하한 도출 명령과 그 출력을 AC 본문에 적어, 다음 사람이 재측정할 수 있게 한다.

### D2 — REQ-ECC-008(양쪽 트리 스캔)을 판정하는 AC가 없다 [BLOCKING]

- **파일**: `spec.md` REQ-ECC-008 / `acceptance.md` §D 매트릭스 + AC-ECC-010 / `plan.md` §D.2
- **주장**: §D 매트릭스는 AC-ECC-010이 REQ-ECC-007·**008**을 판정한다고 적는다.
- **실패**: AC-ECC-010의 통과 조건은 "스캔 파일 ≥ 7, 위반 0" 하나뿐이다. **저장소 루트만 스캔한 가드도 이 조건을 만족한다**(363 ≥ 7, 위반 0). 즉 REQ-ECC-008을 어긴 구현이 그 요구에 매핑된 유일한 AC를 통과한다. AC-ECC-011/012는 픽스처 기반이라 트리를 보지 않고, AC-ECC-013은 허용목록 크기만 본다.
- **가중 사유**: plan.md §D.2가 바로 이 실패 형태를 이름 붙여 놓았다 — 기존 `internal/template/gitignore_agents_mirror_test.go`가 `templates/.gitignore`만 읽어 "저장소 루트에서 규칙을 지워도 초록"이라는 것. 파일을 열어 확인했다: 45행 `filepath.Join("templates", ".gitignore")`, 77행 임베드 사본 — 루트 `.gitignore`를 읽는 코드가 없다. t373 verdict도 같은 결함을 후속 후보 3번으로 적었다. **결함의 형태를 정확히 알면서 그것을 잡는 AC를 세우지 않았다.**
- **Severity**: critical · **Class**: blocking
- **Required fix**: "두 트리 루트가 각각 방문됐고 각 트리의 스캔 파일 수가 하한을 넘는다"를 단언하는 AC를 추가한다. 나아가 plan.md §D.2가 약속한 "두 사본이 이 항목에 관해 서로 일치함"의 단언도 AC로 승격한다 — 현재 어느 AC도 그것을 요구하지 않는다.

### D3 — 131이라는 수의 출처가 그 수를 만들지 못한다 [BLOCKING]

- **파일**: `spec.md` §1.1 표 3행, §5(2회) / `plan.md` §D
- **주장**: "구체적 하위경로를 인용하는 문서 **131**", 출처 = "오케스트레이터 사전 측정, **증거표로 반출**".
- **실측 (이 트리)**: 지목된 증거표 `.moai/reports/t375/cited-path-resolution.txt`에서
  - `awk -F'\t' 'NR>1{print $2}' … | sort -u | wc -l` → **97** (distinct `first_citing_doc`)
  - 데이터 행 `awk -F'\t' 'NR>1' … | wc -l` → 228 ✓, 3열 `True` → 45 ✓, 4열 `True` → 0 ✓ (이 셋은 정확)
  - 재구성 시도: `git grep -lE '\.moai/state/verify/[A-Za-z0-9]' -- '*.md'` → **124**; `git grep -lE '\.moai/state/verify/[^ ]' -- '*.md'` → **174**
- **실패**: 131을 만드는 명령을 찾지 못했고, SPEC이 지목한 증거표는 97을 준다. 이 수는 장식이 아니라 **판정에 쓰인다** — spec.md §1.1의 정정 노트가 "판정에 쓰이는 수(184 / 131 / 228 / 45 / 0)"라고 스스로 지목하고, §5의 잔여 부채 규모와 plan.md §D의 "가드를 만드는 근거"가 전부 131 위에 선다. 귀속 무결성을 입법하는 SPEC 안의 미귀속 수치다.
- **Severity**: critical · **Class**: blocking
- **Required fix**: 131을 만든 명령을 찾아 표에 적거나, 찾지 못하면 증거표에서 재도출되는 수(97 또는 124)로 갈아끼우고 §5·plan.md §D의 파생 서술을 함께 고친다. 3열/4열 집계(228/45/0)는 재측정으로 확인됐으므로 손대지 않아도 된다.

### D4 — §1.1 표의 명령 2행이 자기 값을 만들지 못한다(§1.5와 같은 형태) [BLOCKING]

- **파일**: `spec.md` §1.1 표 1행·2행, 그리고 바로 아래 "총 출현 수에 관한 정정" 노트
- **실측 (이 트리, HEAD `b64043481`)**:

  | SPEC이 적은 것 | 적힌 명령의 실제 출력 | 값을 만드는 명령 |
  |---|---|---|
  | 184 파일 | `grep -rl '\.moai/state/verify' --include='*.md' .` → **187** | `git grep -l … -- '*.md'` → **184** |
  | 532 출현 | `grep -ro '\.moai/state/verify' --include='*.md' .` → **562** | `git grep -o … -- '*.md'` → **532** |

- **187 − 184의 정체**: 이 SPEC 자신의 3개 산출물(`spec.md`, `plan.md`, `acceptance.md`)이다(`comm`으로 확인). 작성 시점에는 184였겠으나, 적힌 명령은 **지금 이 트리에서 187을 낸다** — 그리고 SPEC은 그 사실도, 재측정 시점 의존성도 적지 않았다.
- **가중 사유**: 정정 노트가 "**위 표의 명령이 이 문서가 쓰는 값의 유일한 출처다**"라고 단언하는데, 그 단언이 두 행 모두에서 거짓이다. 두 값은 **추적 집합**(`git grep`)에 대해 옳고 **작업 트리 명령**(`grep -r`)에 대해 틀리다 — 오케스트레이터가 이미 찾은 §1.5 codemaps 결함(`git ls-files '.moai/project/codemaps/*.md'` → 6 vs 디렉터리 7)과 **정확히 같은 단위/라벨 불일치**이며, 이로써 이 형태의 결함은 1건이 아니라 **표 안에 3건**이다. (참고: §1.5의 6-vs-7은 t373 verdict의 서술에서 그대로 이어받은 것으로 보인다 — 원본 오류가 아니라 전파된 오류다.)
- **Severity**: critical · **Class**: blocking
- **Required fix**: 표의 "명령" 열을 실제로 값을 만든 명령(`git grep -l` / `git grep -o`)으로 바꾸고, "잰 트리" 열이 커밋(=추적 집합)을 뜻함을 명시한다. 정정 노트의 단언도 그에 맞춰 고친다.

### D5 — §1.4의 "옳은 사용처가 하나뿐"이 거짓이고, carve-out 목록이 불완전하다

- **파일**: `spec.md` §1.4, REQ-ECC-006 / `acceptance.md` AC-ECC-005
- **주장**: "진짜로 옳은 `.moai/state/verify/` 사용처가 **하나** 있다" = `internal/verify/store.go:15` 스냅샷 저장소.
- **실측**: `grep -rn 'SnapshotDir\|state/verify' --include='*.go' internal/ pkg/ cmd/` (테스트 제외) 결과 기계 소비자가 **둘**이다.
  - `internal/verify/store.go:15` — `SnapshotDir = ".moai/state/verify/snapshots"` ✓ (SPEC이 서술한 대로)
  - **`internal/web/events.go:29`** — `"verify": {".moai/state/verify"}`, fsnotify 감시 맵. `snapshots`가 아니라 **`.moai/state/verify` 디렉터리 전체**를 SSE 이벤트 소스로 감시한다. REQ-ECC-006의 carve-out 열거(`.moai/state/verify/snapshots/`, `moai verify record|check`)에 없다.
  - 부수 발견: **`internal/cli/mcp_glm.go:110`**은 이 SPEC이 없애려는 결함의 **살아 있는 실례**다 — 코드 주석이 측정 근거로 `.moai/state/verify/t225/ac-amp-006-glm-differential-attempt1.md`를 인용한다. 가드 범위가 `.md` 한정이라 이런 재발을 잡지 못한다.
- **기계 파손 위험 판정 (지시 항목 2에 대한 답)**: **없다.** REQ-ECC-002는 "규칙 문서"를 구속하는 문서 규범이고, `moai verify record|check`도 attributable diff-check도 코드 경로다. `agent-common-protocol.md`에서 `state/verify` 출현은 268행 **1건뿐**이며 attributable diff-check 절은 경로가 아니라 CLI 동사를 인용한다(확인함). 베이스라인도 초록이다: `go test ./internal/verify/...` → `ok … 1.864s`. **AC-ECC-006이 스스로 적은 한계("문서 변경이 코드를 깨지 않았음만 보인다")는 정직하다** — 다만 그 한계가 가리는 것은 파손이 아니라 **carve-out 열거의 누락**이고, 그건 AC-ECC-005의 grep(존재 확인)으로도 잡히지 않는다.
- **Severity**: major · **Class**: blocking
- **Required fix**: §1.4의 "하나"를 고치고, REQ-ECC-006의 carve-out에 `internal/web/events.go`의 디렉터리 감시를 포함하거나 왜 문서 규범의 대상이 아닌지 한 줄로 적는다. `mcp_glm.go:110`은 별건 후속 후보로 §5 부채에 넣는다.

### D6 — AC-ECC-002의 판정 명령이 오늘 이미 통과한다

- **파일**: `acceptance.md` AC-ECC-002
- **주장**: `grep -n 'scratch\|export' .claude/rules/moai/core/agent-common-protocol.md`로 (a)스크래치 명명 (b)추적 경로 인용 (c)인용 전 반출 세 요소를 판정한다.
- **실측 (수리 이전 트리)**: 이 grep이 **오늘 매치한다** — `agent-common-protocol.md:337`, `| Location | CWD is the primary checkout. Exempt: an already-isolated worktree, /tmp, or a session-private scratch dir |`. Pre-Edit Sync Check 표의 무관한 문장이다.
- **실패**: 명령이 편집 전후를 구별하지 못한다(rc=0 → rc=0). AC 본문은 "세 요소가 모두 있어야 통과다 — 둘만 있으면 실패다"라고 엄격히 적지만, 적힌 명령은 **한 요소도** 증명하지 않는다. 이 AC가 덮는 것은 REQ-ECC-001과 REQ-ECC-003 — 이 SPEC의 중심 요구 둘이다.
- **대조군(정상)**: AC-ECC-001 두 grep 모두 오늘 1행 매치(수리 후 0행이 되므로 판정력 있음) ✓ / AC-ECC-003 grep rc=1 ✓ / AC-ECC-004 grep rc=1 ✓ / AC-ECC-005 `grep -c 'state/verify/snapshots'` → 0 ✓. 즉 **AC-ECC-002만** 이 성질을 잃었다.
- **Severity**: major · **Class**: blocking
- **Required fix**: 세 요소 각각에 대해 오늘 RED인 고정 문자열 grep을 하나씩 세운다(예: 스크래치 명명은 `machine-local scratch` 같은 신설 문구, 반출 의무는 `export before citing` 같은 신설 문구). 넓은 단어 대안(`scratch\|export`)은 버린다.

### D7 — 허용목록의 **단위**가 미지정이고, 파일 단위라면 대상 7건 중 3건이 영구 면제된다

- **파일**: `plan.md` §D.3 / `spec.md` REQ-ECC-009 / `acceptance.md` AC-ECC-013
- **관측**: REQ-ECC-009와 AC-ECC-013은 허용목록의 **개수**를 못 박지만, 항목이 무엇인지(파일 단위인지 / 파일+행 단위인지 / 문자열 문맥 단위인지)는 어느 문서도 정하지 않는다.
- **실패**: §C.3의 carve-out 후보는 `gate.md:122`, `loop.md:115`, `run.md:199` 세 파일이다. 허용목록이 **파일 단위**면 이 셋은 통째로 면제되고, 나중에 누가 이 파일들에 진짜 위반 인용을 추가해도 가드가 보지 못한다. 저장소 루트 사본 기준 위반 대상 7건 중 **3건(43%)**이 가드 사각지대가 된다. AC-ECC-013의 개수 단언은 "목록이 조용히 길어지는 것"은 막지만 "항목 하나가 파일 전체를 삼키는 것"은 막지 못한다.
- **Severity**: major · **Class**: blocking
- **Required fix**: 허용목록 항목을 **파일 + 행 문맥**(또는 파일 + 정확한 리터럴) 단위로 규정하고, AC-ECC-013에 "항목이 파일 전체를 면제하지 않는다"는 단언을 더한다.

### D8 — §4.2 판정의 미명시 귀결: `.gitkeep`을 만들면 opt-in이 사라지고, 안 만들면 예외 줄이 무효다

- **파일**: `spec.md` §4.2, REQ-ECC-011 / `acceptance.md` AC-ECC-014
- **검증한 근거 (전부 확인됨)**: `.gitignore:229-230`에 텔레메트리 형제 2줄 존재 ✓ / `internal/hook/post_tool_duration.go`의 `hookMetricsRelPath` 상수와 관문 둘(임계값 조기 반환, `os.Stat(obsDir)` 부재 시 REQ-CC2122-HOOK-001-003 인용하며 조용히 반환) ✓ / 두 트리 모두 `.moai/observability` 디렉터리 부재 ✓ / 어느 `.gitignore`에도 항목 없음 ✓.
- **§4.2의 층위 분리 논증 판정 (지시 항목 3에 대한 답)**: **논증 자체는 타당하다.** 디렉터리 존재가 여는 것은 코드상 *기록*이고(`os.Stat` 관문), *공유*를 여는 코드 결합은 존재하지 않는다. t373이 갈리지 않는다고 본 두 읽기가 다른 층에 있다는 주장은 소스로 뒷받침된다.
- **그러나 귀결이 적히지 않았다**: `git ls-files`로 확인한 결과 형제 선례 `.moai/evolution/telemetry/.gitkeep`은 **실제로 추적된다**. §4.2가 "텔레메트리 형제와 **같은 두 줄 형태**"를 택하고 "디렉터리(=opt-in)는 커밋 가능하게 남고"라고 적으므로, 같은 형태를 끝까지 따르면 `.moai/observability/.gitkeep`이 추적되고 — **모든 clone에서 디렉터리가 존재하게 되어 `os.Stat(obsDir)` 관문이 항상 통과한다.** 코드가 일부러 세운 opt-in(REQ-CC2122-HOOK-001-003)이 opt-out으로 뒤집힌다. 이것은 §4.2의 논증이 딛고 선 바로 그 opt-in을 제거하는 귀결인데, 어디에도 적혀 있지 않다.
- **반대 뿔**: `.gitkeep`을 만들지 않으면 `!.moai/observability/.gitkeep`은 **존재한 적 없고 앞으로도 없을 파일에 대한 예외**가 되어, §4.3이 "아직 만들어진 적 없는 파일을 예측으로 무시하는 것"이라며 반대한 바로 그 수를 두게 된다. AC-ECC-014는 ignore 두 줄의 존재만 단언하므로 **어느 뿔도 해소하지 않는다**.
- **Severity**: major · **Class**: blocking
- **Required fix**: `.gitkeep`을 만들 것인지 명시하고, 만든다면 "모든 clone에서 훅 기록이 기본 활성화된다"는 귀결을 §4.2에 적은 뒤 그것이 의도인지 판정한다. 만들지 않는다면 `!` 예외 줄을 빼거나 왜 남기는지 적는다.

### D9 — §4.3이 §4.1에서 뽑은 판별식을 정작 자기가 무시하기로 한 경로에 적용하지 않았다

- **파일**: `spec.md` §4.1, §4.3, REQ-ECC-012
- **t373 선례 충실도 판정 (지시 항목 4에 대한 답)**: **충실하다.** `/Users/goos/MoAI/moai-adk-go/.moai/reports/t373/verdict.md` 131-147행을 직접 읽었다. 원문은 "`.gitignore`는 '이 파일은 무시해도 된다'는 선언인데, 여기 남은 잔여는 정확히 **무시하면 안 되는 것**이다 — 처분되지 않았다는 뜻이기 때문이다 … `git status`에 뜨는 것이 **처분이 안 됐다는 유일한 신호**"라고 적는다. §4.1의 판별식("그 자리에 파일이 남아 있다는 사실이 무언가 잘못됐다는 신호인가?")은 편의적 재구성이 아니라 정확한 추출이다.
- **실패**: 판별식은 §4.2(observability)에는 적용되고 §4.3에는 **한 번도 적용되지 않는다**. §4.3의 논거는 전부 codemaps 대비 종류 구분(추려진 지도 vs 1회분 요청 상태)뿐이다. 그런데 판별식을 실제로 돌리면 답이 뒤집힐 여지가 크다: `internal/cli/navigator_fix.go`(18-22행, 58-62행)는 `fix-drafts/<draft-id>/request.json`을 **핸드오프 계약**으로 방출하고 오케스트레이터가 그것을 소비해 draft 위임을 spawn한다. 남아 있는 `request.json`은 **요청됐으나 완료되지 않은 위임**을 뜻하며, 이는 t373이 무시하지 않기로 한 `chain/`·`migrate-tx`와 같은 모양이다.
- **부가 — 논거의 비대칭 적용**: §4.3은 나머지 navigator 산출물을 무시하지 않는 근거로 "아직 만들어진 적 없는 파일을 예측으로 무시하는 것"을 든다. 그러나 `fix-drafts/`도 똑같이 존재하지 않는다 — `find .moai/project/navigator -maxdepth 2` → 추적 템플릿 `symbols/narrative.template.md` **1개뿐**(primary도 동일). 존재하지 않는다는 반론이 A는 막고 B는 막지 않는 이유가 적혀 있지 않다.
- **정확했던 부분**: `capability-map.md`가 입력으로 읽힌다는 주장은 확인됨(`navigator_enrich.go:75`, `filepath.Join(root, ".moai", "project", "navigator", "capability-map.md")`, 출력은 `.moai/project/codemaps`). 디렉터리 통째 ignore를 막는 판정은 옳다.
- **Severity**: major · **Class**: blocking
- **Required fix**: §4.3에서 `fix-drafts/`에 §4.1 판별식을 명시적으로 적용하고, "미완 위임의 잔여 신호"라는 반대 독해를 다루거나 그 독해가 왜 성립하지 않는지 적는다. 존재-부재 논거의 비대칭도 해소한다.

### D10 — AC 3건에 요구 층 근거가 없고, SPEC이 "가장 강한 사례"라 부른 파일이 그중 하나다

- **파일**: `acceptance.md` §D 매트릭스(AC-ECC-007/008/009 행), `spec.md` §1.3, REQ-ECC-002
- **관측**: AC-ECC-007(manager-lead.md 단정 삭제)·AC-ECC-008(출력 스타일 3지점)·AC-ECC-009(경계 사례 3건 기록)의 "요구" 열은 각각 `§C.2`, `§C.2`, `§C.3` — **plan.md의 절 번호**이지 REQ가 아니다.
- **실패**: REQ-ECC-002는 "**규칙 문서**는 …"으로 시작한다. `.claude/agents/moai/manager-lead.md`는 에이전트 정의, `.claude/output-styles/moai/moai.md`는 출력 스타일 — 둘 다 규칙 문서가 아니다. 즉 spec.md §1.3이 "거짓 전제를 **지시로 바꾸는** 지점"이자 "가장 강한 형태"라 지목한 파일을 고치라고 명령하는 **요구가 존재하지 않는다**. 현재 커버리지는 REQ-ECC-007(가드 범위에 `.claude/agents/`·`.claude/output-styles/` 포함) + AC-ECC-010(위반 0)을 경유한 **간접**뿐이다 — 그리고 그 경로는 D1·D2로 이미 약하다.
- **Severity**: major · **Class**: blocking
- **Required fix**: REQ-ECC-002의 주어를 "규칙 문서"에서 "doctrine 표면 문서(`.claude/rules/`, `.claude/agents/`, `.claude/output-styles/`, `.claude/skills/`)"로 넓히거나, 에이전트 정의·출력 스타일을 구속하는 REQ를 신설하고 AC-ECC-007/008을 거기에 매핑한다.

### D11 — REQ-ECC-004/005는 "인용 자체가 넓은" 경우의 통째 반출을 막지 못한다

- **파일**: `spec.md` §3.1, REQ-ECC-004, REQ-ECC-005
- **주장 (지시 항목 5)**: 선택 기준과 통째 반출 금지가 **둘 다** 있어야 13 MB가 목적지만 바꿔 이동하는 것을 막는다.
- **실패**: 두 요구가 서로 **다른 상한**을 건다. REQ-ECC-004의 상한은 "인용문이 실제로 이름 붙인 것", REQ-ECC-005의 기준은 "판정을 결정한 명령과 그 판정 결정선". 이 둘이 화해되지 않아, **인용문을 넓게 쓰면 양쪽을 동시에 만족하면서 전부 반출할 수 있다**. 구체적으로: 오늘 `moai.md:384`가 쓰는 형태 그대로 `evidence: .moai/state/verify/<session>/`라고 인용하면 그 인용문은 디렉터리를 "이름 붙인" 것이고, REQ-ECC-004는 문자 그대로 만족된다. **인용의 넓이 자체를 구속하는 요구가 없다.**
- **Severity**: major · **Class**: blocking
- **Required fix**: REQ-ECC-004에 "인용은 디렉터리가 아니라 파일을 이름 붙인다" 또는 그에 준하는 인용-넓이 상한을 넣거나, REQ-ECC-005의 기준을 REQ-ECC-004의 상한으로 삼도록 두 요구를 하나의 상한으로 묶는다.

### D12 — Tier M 판정이 자기 §C의 최대 케이스를 세지 않았다

- **파일**: `plan.md` §B(영향 파일 11–13), §C
- **재계산 (plan.md 자신의 §C에서)**: C.1 규칙 2 + C.2 지시 2 + C.3 스킬 **최대 3** + C.4 gitignore 2 + C.5 가드 1 + C.6 미러 **최대 7**(미러 grep으로 확인: 두 protocol 파일, moai.md, manager-lead.md, gate/loop/run.md) + 방출 `.toml` 1 = **최대 18**.
- **실패**: 11–13은 C.3 세 파일이 **전부 carve-out으로 결정될 때만** 성립하는데, 그 결정은 §C.3이 명시적으로 M2로 미룬다. 최대 케이스 18은 Tier M 구간(5–15) 밖이다. Tier 판정이 아직 내리지 않은 결정의 유리한 쪽을 가정하고 있다.
- **Severity**: minor · **Class**: optional (Tier 상향이 필요하다고 단정하지 않는다 — 최소 케이스는 M에 들어맞는다)
- **Required fix**: §B에 범위를 조건부로 적는다("C.3이 전부 carve-out이면 11, 전부 교체면 18") 또는 M2 결정 후 Tier를 재확인한다는 문장을 넣는다.

### D13 — §1.1의 "세션 디렉터리 124개"가 스냅샷 저장소를 세션으로 센다

- **파일**: `spec.md` §1.1 (primary 실체 문단)
- **실측 (primary)**: `find /Users/goos/MoAI/moai-adk-go/.moai/state/verify -maxdepth 1 -mindepth 1 -type d | wc -l` → **124** ✓ (수 자체는 맞다). 그러나 그 124개 중 하나가 `.moai/state/verify/snapshots` — §1.4가 세션 증거와 명시적으로 구분한 기계 저장소다. 실제 세션 디렉터리는 **123**.
- **함께 확인된 정확한 값**: 파일 905 ✓, `du -sh` 13M ✓, `.moai/reports` 2,421 파일 ✓ / 29M ✓ / `git ls-files .moai/reports | wc -l` → 974 ✓, `.moai/reports/t341` 13 파일 ✓ / 52K ✓ / 하위 `verify/` 1개 ✓.
- **Severity**: minor · **Class**: optional
- **Required fix**: "세션 디렉터리 124개"를 "최상위 디렉터리 124개(그중 1개는 `snapshots` 기계 저장소, 세션 123개)"로 고친다.

### D14 — 미러 정합성 (지시 항목 8에 대한 답): **누락 없음**

결함이 아니라 확인 결과다. `plan.md` §C.6 + M6이 요구하는 미러가 전부 존재함을 확인했다: `grep -rl 'state/verify' internal/template/templates/` → `.claude/{agents/moai/manager-lead.md, output-styles/moai/moai.md, rules/moai/core/agent-common-protocol{,-reference}.md, skills/moai/workflows/{gate,loop,run}.md}` 7건 + `.codex/agents/moai/manager-lead.toml` 1건. `.gitignore` 미러는 C.4/M5가 별도 행으로 처리 ✓. 방출 체인 서술도 정확하다 — C.6은 `make agents-emit`이 "`.claude/agents/moai/*.md` **미러에서**" 방출한다고 적는데, CLAUDE.local.md §2.0의 C2→C3 관계와 일치한다.

**다만 잔여 위험 하나**: `.codex/agents/moai/manager-lead.toml`은 `.toml`이라 가드(`.md` 한정) 밖이다. 오늘 그 파일은 `canonical persistence location` 문장을 담고 있다(`grep -rn` 확인). M6의 `make agents-emit` → `make agents-emit-check`가 전이적으로 고치므로 이 SPEC 내에서는 닫히지만, **가드가 지키는 것은 아니다**. acceptance.md 경계 사례 3번째 항목이 이 점을 이미 적고 있으므로 결함으로는 올리지 않는다.

### D15 — 잔여 부채 서술 (지시 항목 6에 대한 답): 행동 가능하되 근거 수치가 D3에 오염됨

`spec.md` §5는 (a)반출 후 인용 교체 / (b)미귀속 표시라는 **두 갈래 행동**을 적고 있어 카드가 되기에 충분히 구체적이다. 부채 규모 수치(183/228, 228/228)는 산술적으로 정합하고(228 − 45 = 183) 증거표에서 재도출됨을 확인했다. 다만 대상 문서 수 131이 D3의 미귀속 수치이므로, 카드로 올릴 때 그 수를 함께 고쳐야 한다. DoD 3번이 리드 보고를 요구하는 것도 적절하다. **Severity: minor · Class: optional.**

---

## Regression Check

해당 없음 — iteration 1.

---

## Recommendation

**FAIL.** 총점 0.70이 Tier M 임계 0.80에 미달하고, blocking 11건 중 3건(D1·D2·D3)이 이 SPEC의 중심 주장을 직접 무너뜨린다. must-pass 위반은 없으므로 구조적 재작성이 아니라 **열거된 결함 델타에 대한 수리**로 충분하다.

수리 순서 제안 — 앞의 셋이 나머지보다 훨씬 무겁다:

1. **D1 + D2 (가드)** — 이 SPEC의 존재 이유가 가드인데, 현재 AC 집합으로는 "저장소 루트만 스캔하고 범위가 무너진 가드"가 전부 초록으로 통과한다. 실제 모집단(루트 363 / 미러 340)에서 도출한 하한을 **두 트리 각각**에 걸고, 두 트리 방문을 단언하는 AC를 추가한다. plan.md §D.2가 이미 실패 형태를 정확히 서술해 두었으므로 AC로 승격만 하면 된다.
2. **D3 + D4 (귀속)** — 귀속 무결성을 입법하는 문서 안에 출처가 자기 값을 만들지 못하는 수치가 **3건**(184 / 532 / 131) 있다. 앞의 둘은 명령을 `git grep`으로 바로잡으면 닫히고, 131은 재도출하거나 증거표의 97로 갈아끼워야 한다. §1.5의 codemaps 6-vs-7과 합치면 §1.1–§1.5 안에서만 같은 형태가 4건이다.
3. **D6 (판정력 없는 AC)** — AC-ECC-002의 grep을 오늘 RED인 고정 문자열로 교체한다. 나머지 AC의 grep은 오늘 트리에서 올바르게 RED임을 확인했으므로 이 하나만 고치면 된다.
4. **D5 · D7 · D8 · D9 · D10 · D11** — 각각 한 문단 이하의 문서 수리로 닫힌다. 새 측정이 필요한 것은 D5뿐이며(`internal/web/events.go:29`를 carve-out에 넣을지 판정), 나머지는 이미 적힌 논거의 구멍을 메우는 작업이다.
5. **D12 · D13 · D15** — optional. 수리해도 좋고 부채로 넘겨도 판정에 영향이 없다.

**재감사 범위**: 위 델타에 한정한다(Tier M 재감사 상한 iteration 2). 전면 재감사는 필요하지 않다 — must-pass 7항목과 §1.1의 나머지 측정치(228/45/0, 905/13M/124, 974/2421/29M, t341 13파일/52K), t373 선례 충실도, 미러 정합성, 베이스라인 초록(`go test ./internal/verify/...`, `go vet` + `go test ./internal/template/...`)은 이 감사가 이미 재측정해 통과 확인했다.
