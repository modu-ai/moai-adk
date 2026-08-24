# Progress — SPEC-ZONE-REGISTRY-RESYNC-001

## §E.1 Plan-phase Audit-Ready Signal

- Tier: M (근거: `plan.md` §A Tier 판정 — REQ 15/16, AC 14/16, 압축 없음)
- 산출물: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`
- 범위: 3축 전부 (clause 재동기화 + anchor 탐지·수리 + 기계적 가드)
- REQ: 15건 (REQ-ZRR-001..015) · AC: 14건 (AC-ZRR-001..014, BLOCKING 9)
- 마일스톤 순서 구속: M1 수리 → M2 가드 → M3 검증 (의존성, 권고 아님 — `plan.md` §F)
- 근거 보고서: `.moai/reports/t232/findings.md` + `validate-repro.txt` + `analysis-repro.json`
- RED 기준값이 모든 AC에 명시됨 (`acceptance.md` §D 매트릭스), 측정 트리 `294b4b6ab`
- plan-audit iter1: **FAIL 0.75** (Tier M 임계 0.80; must-pass 7/7 PASS, 4개 차원 전부 0.75). 보고서: `.moai/reports/t232/plan-audit-iter1.md`
- iter1 반영 (v0.3.0): blocking 6건(D-1 자기참조 `file:` / D-2 빈 clause / D-3 `|| true` / D-4+D-8 평가 엔트리 수·변이 대상 고정 / D-9 paths-filter / D-5 파일 수 오기) + optional 4건(D-6 HISTORY / D-7 REQ-015 분리·재배치 / D-10 slug 열거 5개 / D-11 §7 갱신) 전부 적용. 새 AC·티어 변경 없음 — 기존 AC 의 Then 절 강화로 흡수
- plan-audit iter2: **PASS-WITH-DEBT 0.925** (임계 0.80) — run-phase 진입 승인. blocking 6건 중 5건 완전 마감, AC 약화 0건, REQ 15 / AC 14 유지
- iter2 후속 (v0.4.0, 재감사 없음): 부채 #1 마감 — 빈 clause 금지를 **유일 적중**(정확히 1회) 요구로 강화(공백 한 칸·짧은 토큰 우회 차단; RED 1회 적중 24 / 2회 이상 1건 `CONST-V3R2-002`). N-1 §D8 인용 정정, N-2 REQ 전역 오름차순 재배치
- **잔여 부채 2건 — 수정 대상 아님, 판정자 지정됨** (`plan.md` §H): ① `CONST-V3R2-004` 근접 오답(`NOTICE.md` 로 이동 시 기계 통과 — sync 리뷰어가 이름으로 거부) ② 평가 엔트리 수 이중 카운트(clause 97 / anchor 101 (v0.5.0 C안) — sync-auditor 가 §E.2 인용에서 수가 둘인지 확인)
- status: draft — run-phase 판단 대기 (kanban lead)

## §F Phase 4 Mode Selection

- 입력: tier M · 스코프 ~20파일 (레지스트리 2 미러 편집 + 인용 대상 17 문서 읽기 + `registry_sync_test.go` 신규 + `ci.yml`) · 도메인 1 (constitution registry) · 언어 혼합 markdown+Go · 동시성 이득 LOW (coding-heavy, M1→M2→M3 의존성 고정)
- direct: 미선택 — 다중 파일·의미 있는 변경
- serial: **선택** — coding-heavy 순차 작업 (Anthropic coding-task parallelism caveat), 마일스톤 순서가 의존성으로 고정 (`plan.md` §F)
- fanout: 미선택 — 단일 도메인, 쓰기 경합 위험
- sweep: 미선택 — 비기계적 변환, 파일 수 미달
- 경계 사례: 없음
- **§1.2 충돌 판정: C안** (은퇴 엔트리 clause 면제 · anchor 검사 유지). 근거 3측정 @ 트리 `9ba1e308d`: ① 은퇴 4건(`CONST-V3R2-021..024`) anchor `#14-parallel-execution-safeguards` 가 로컬·템플릿 CLAUDE.md 양측 모두 해석됨(두 파일 바이트 동일, §14 = 153행) ② `[SUPERSEDED … see CLAUDE.md §14 …]` 마커가 가리키는 후계 교리(worktree-opt-in 포인터)가 §14 본문에 생존 — anchor 는 사라진 절이 아니라 **후계 절의 포인터**라 anchor 검사가 마커 안내의 유효성을 검증함 ③ 신규 분석: 총계(clause 실패 68 / anchor 실패 17)는 `analyze.py`, 은퇴·비은퇴 분리(clause 실패 중 은퇴 4 / anchor 실패 중 은퇴 0)는 `retired-vs-ac.py`(입력 `analysis-postmerge.json`) — plan-audit iter3 독립 재계산으로 재현 확인. B안은 측정상 불가능(인용 원문이 정의상 소멸), A안은 §14 개명 시 마커 포인터 사망을 무신호로 놓침. #1611 의 `--strict` 재검사 설계(`validator.go:214`)와 동형. 정본 반영: spec.md v0.5.0 (manager-spec 위임)

- **Phase 1 게이트 이력·에스컬레이션 판정 (post-iter3)**: iter3(v0.5.0 델타 감사) = **FAIL 0.825** (임계 0.80 초과·must-pass 7/7 PASS — FAIL 사유는 BLOCKING AC 3개의 문서 표면 간 모순 D1/D3/D4 + 신규 D2/D5/D6·D7; 보고서 `.moai/reports/t232/plan-audit-iter3.md`. 교차모델: claude+glm pass / codex fail — 감사가 codex 지적 6건 중 5건 독립 검증 후 수용·확정). 재시도 계약상 iter3 이후 감사 반복은 불가(계약상 에스컬레이션 3택) → **PASS-with-debt 경로로 run 진입을 판정**: 결함 D1-D9 는 전부 문서 동기화 결함이며 `5073c7fbd`(C안 확정)·`e0afbb53c`(잔여 시제·Then 절)·`ac5a69e9c`(D1-D8)·§G 패리티 커밋으로 마감됐고, 감사관 자신이 "범위 축소는 오진단 — 1회 동기화 수정 권장"으로 지목한 경로다. 잔여 부채: D10(`analysis-*.json` 3종 동일 blob 주석 — optional) + §H ①`CONST-V3R2-004` 근접 오답(sync 리뷰어 판정)·② 평가 수 이중 보고(sync-auditor 판정). 스킵 조건(아티팩트 해시 불변)이 커밋마다 재무효화되므로 본 판정이 최종 게이트 기록이다.
- **M1 수행 출처 공개 (프로세스 부채)**: M1 수리와 커밋(`2319df7ac`)은 manager-develop 위임 **전에** zrr-spec-amend(manager-spec 계열 subagent)가 자발 실행했다 — 최초 위임의 [HARD] "zone-registry.md 수정 금지(run-phase 소관)" 위반이며, 오케스트레이터의 "이 커밋 후 종료" 지시 이후에 커밋이 착지해 TaskStop 으로 정지했다. 오케스트레이터 독립 재검증(2026-08-25, 커밋 후 워킹트리 기준): analyzer clause_fail=4(은퇴 4건 전부)·anchor_fail=0 · 리터럴 once=97/zero=0/multi=0/retired=4/selfref=0 양 트리 각각 · 미러 바이트 동일 · `git diff HEAD -- internal/constitution/ internal/cli/ cmd/` 빈 출력(매처 무변경) · 엔트리 집합 불변(±`id:` 0건, 변경 라인은 file/anchor/clause 필드만) · `go run ./cmd/moai constitution validate` exit 0. 내용은 수용하되 소유권 위반은 부채로 남긴다 — sync 리뷰 판정 대상.

## §E.2 Run-phase Evidence

### M1 — 데이터 수리 (clause + anchor) — 2026-08-25

측정 트리: `WT-zone-registry-drift` @ `e0afbb53c`(수리 전 측정) / 동일 커밋 + M1 수리 워킹트리(수리 후 측정; M1 커밋 SHA 는 §E.3 `run_commit_sha` 에 런 종료 시 기록). 수리 스크립트: `.moai/reports/t232/m1-repair.py`(기존 분석기 체계와 동일 위치, 커밋 외 산출물).

#### 1. 사전 점검 (RED baseline) — 트리 `e0afbb53c`, §C 표 재측정

```
$ git branch --show-current && git rev-parse --short HEAD
WT-zone-registry-drift
e0afbb53c

$ go test -count=1 ./internal/constitution/...
ok  	github.com/modu-ai/moai-adk/internal/constitution	0.500s   (rc=0)

$ diff -q .claude/rules/moai/core/zone-registry.md internal/template/templates/.claude/rules/moai/core/zone-registry.md
(출력 없음 — 바이트 동일, rc=0)   grep -c '^- id: CONST-' → 양쪽 101

$ python3 .moai/reports/t232/analyze.py . .moai/reports/t232/analysis-m1-preflight.json
entries=101 clause_fail=68 anchor_fail=17
both_fail=8 clause_fail_anchor_ok=60 clause_ok_anchor_fail=9 missing_file=0

$ (retired-vs-ac 동일 로직, analysis-m1-preflight.json 대상)
retired entries: 4  (CONST-V3R2-021/022/023/024 — 전부 clause_ok=False, anchor_ok=True, file=CLAUDE.md)
clause failures total: 68 / of which retired: 4 / non-retired failures: 64
anchor failures total: 17 (retired among them: 0)
```

→ §C 기대치와 정확히 일치(clause 실패 68 = 은퇴 4 + live 64, anchor 실패 17 = 은퇴 0). **E8 RED 근거**: analyzer 사전 측정에서 live clause 실패 64 / anchor 실패 17 관측.

#### 2. 수리 범위 확정 — 리터럴 체크 기준 73건 산출 내역

AC-ZRR-002/003 의 GREEN 은 검증기가 아니라 **리터럴 체크**(`grep -F` 정확히 1회 적중) 97/97 이다. 이 기준의 live 실패 73건은 다음과 같이 환원된다(acceptance.md §AC-ZRR-002 각주의 8건·1건과 정확히 일치):

- 64건 — analyzer 기준 live clause 실패(패러프레이즈/요약 라벨)
- 8건 — `CONST-V3R5-004/005/006/007/008/010/011/013`: 검증기는 통과(정규화가 행 결합)하나 `grep -F` 실패(다중 행 clause) → 단일 행 구간으로 재선택
- 1건 — `CONST-V3R2-002`(`TRUST 5`, 3회 적중): 유일 적중 사다리 1단(더 긴 구간)으로 `All code changes must pass TRUST 5 validation` 해소

#### 3. 수리 실행 (단일 배치, 양 미러 동시 적용)

```
$ python3 .moai/reports/t232/m1-repair.py
OK: repaired entries=74 (clause edits=73, anchor edits=18, file re-points=4)
OK: all live clauses hit exactly once in BOTH trees; all 101 anchors resolve in BOTH trees
OK: retired 4 untouched; ids/zone/zone_class/canary_gate unchanged; mirrors byte-identical
```

스크립트는 기록 **전**에 전 엔트리에 대해 양 트리 검증(유일 적중·anchor 해석·자기참조 금지·D2 불변)을 통과해야 기록한다(1차 실행에서 전사(転写) 오류 1건 + em-dash slug 오산 2건이 이 게이트에서 저지돼 수정 후 재실행 — 기록되지 않은 실패 경로 없음).

#### 4. 수리 후 GREEN 측정

```
$ python3 .moai/reports/t232/analyze.py . .moai/reports/t232/analysis-m1-post.json
entries=101 clause_fail=4 anchor_fail=0
both_fail=0 clause_fail_anchor_ok=4 clause_ok_anchor_fail=0 missing_file=0
  (clause 실패 4건 = 은퇴 4건 전부 — anchor_ok=True 유지, clause 값 불변 확인)

$ (리터럴 체크 — grep -F -c 적중 횟수 집계, analysis-m1-post.json 기준)
LOCAL : hit_once=97 hit_zero=0 hit_multi=0 retired_clause_exempt=4 self_reference=0
TMPL  : hit_once=97 hit_zero=0 hit_multi=0 retired_clause_exempt=4 self_reference=0

$ go run ./cmd/moai constitution validate ; echo exit=$?
constitution validate: OK — no drift or violations detected (0 entries checked)

  4 retired entry/entries skipped ([SUPERSEDED …] marker); re-check them with --strict
exit=0
```

- **평가 수 분리 인용**(`plan.md` §H 잔여위험 ② 대응): **clause 검사 97**(live verbatim; 은퇴 4건은 retirement 마커 판정 후 skip) / **anchor 검사 101**(전 엔트리, 은퇴 포함) — 두 수를 별개로 인용한다. 리터럴 체크는 97 live × 2 트리(로컬·템플릿) 각각.
- `(0 entries checked)` 는 측정치가 아니다 — `internal/cli/constitution.go:386` 의 OK 경로가 상수 `0` 을 하드코딩한다(기존 결함, M1 불변 대상 밖). 실측 평가 수는 위 analyzer/리터럴 체크 출력이 담당한다. M2 가드는 이 출력이 아니라 자체 카운터를 단언한다(AC-ZRR-007 평가 엔트리 수 101×2).
- `--strict` 경로는 M2 소관이라 M1 에서 돌리지 않았다(은퇴 4건의 verbatim 은 정의상 실패 — spec.md §1.2).

#### 5. 재지정 목록 (sync 리뷰 판정 대상)

**`file:` 재지정 4건** (D4 — 교리 이사, CLAUDE.md §1 → moai-constitution.md):

| ID | 구 file → 신 file | anchor |
|---|---|---|
| CONST-V3R2-008 | CLAUDE.md → .claude/rules/moai/core/moai-constitution.md | #response-language |
| CONST-V3R2-009 | CLAUDE.md → .claude/rules/moai/core/moai-constitution.md | #parallel-execution |
| CONST-V3R2-010 | CLAUDE.md → .claude/rules/moai/core/moai-constitution.md | #output-format |
| CONST-V3R2-011 | CLAUDE.md → .claude/rules/moai/core/moai-constitution.md | #output-format |

신 file 은 전부 paths-filter 내(`.claude/rules/moai/**`, D12). **자기참조 `file:` 0건**(D13). `CONST-V3R2-004` 는 `coding-standards.md #language-policy` 에 잔류한다(`NOTICE.md` 근접 오답 미채택 — spec.md §H 잔여위험 그대로).

**`anchor:` 재지정 18건** (실패 17 + 하기 1):

- `CONST-V3R2-001`: `#phase-overview` → `#plan-phase` (anchor-only — "SPEC+EARS format" 교리의 실제 서식지는 `## Plan Phase` 의 "Create comprehensive specification using EARS format."; file 불변)
- `CONST-V3R2-003`: `#mx-tag-types` → `#scope`
- `CONST-V3R2-008..011`: `#1-hard-rules` → 상기 표 (file 이사 수반)
- `CONST-V3R2-028/029`: `#opus-47-prompt-philosophy` → `#opus-5-48-prompt-philosophy` (heading 개명 "Opus 5 / 4.8 Prompt Philosophy" — slug 규칙상 `/`·`.` 제거)
- `CONST-V3R5-004`: `#ci-auto-fix-loop-entry-condition` → `#entry-condition`
- `CONST-V3R5-005/006`: `#iteration-limit` → `#iteration-cap`
- `CONST-V3R5-007`: `#commit-strategy` → `#patch-commit-rule-no-force-push`
- `CONST-V3R5-008/009`: `#user-interaction-channel` → `#askuserquestion-boundary`
- `CONST-V3R5-010`: `#semantic-failure-handling` → `#semantic-failure-no-auto-patch`
- `CONST-V3R5-011`: `#protected-files` → `#secrets-and-credentials-protection`
- `CONST-V3R5-012`: `#audit-log` → `#audit-log-requirement`
- `CONST-V3R5-013`: `#protected-files` → `#ci-infrastructure-preservation`

참고: ci-autofix-protocol.md 의 `<!-- anchor: … -->` HTML 주석(구 anchor명)은 slug 규칙이 읽지 않으므로 전부 실제 heading slug 로 재지정했다. 011/013 이 같은 구 anchor(`#protected-files`)를 공유하고 있었으나 실제로는 서로 다른 절(Secrets 보호 / CI 인프라 보존)을 인용하므로 각 절로 분리했다.

#### 6. 유일 적중 사다리 기록

- 1단(더 긴 구간): `CONST-V3R2-002`(`TRUST 5` 3회 → 전체 문장 span) 외, 전 clause 가 단일 행 전체/연속 구간으로 선택되어 기본 적용. 2단(최소 20자): 전 clause 20자 이상 충족(최소 길이 clause 도 43자+). **3단 잔여(불가능 엔트리 기록 대상): 0건.**
- 우회 금지 확인: 빈 clause 0건, clause 내 이중따옴표 0건(analyzer 파서가 escape 를 풀지 못해 `\"` 가 문자 그대로 남는 우회 경로를 원천 차단), 자기참조 `file:` 0건, 길이 임계·제외 목록·fuzzy 없음(D7).

#### 7. 엔트리 집합 불변 증명 (AC-ZRR-006, M1 시점)

```
before=101 after=101 / id sets identical: True
zone changed: [] / zone_class changed: [] / canary_gate changed: []
file changed (4): CONST-V3R2-008/009/010/011
anchor changed (18): (상기 §5 목록)
clause changed: 73
other-field changes: []
git diff 라인 분류: 변경 190 라인 = (file 4 + anchor 18 + clause 73) × 2 — 이 외 라인 변경 0
```

#### 8. 부수 검증

```
$ go test -count=1 ./internal/constitution/...        → ok  3.457s  (rc=0)
$ go build ./...                                      → rc=0
$ GOOS=windows GOARCH=amd64 go build ./...            → rc=0
$ go vet ./internal/constitution/...                  → rc=0
$ 템플릿 중립성 grep(SPEC-ID/ISO-날짜/40-hex-SHA)      → 0 / 0 / 0  (AC-ZRR-012 비회귀 유지)
$ diff -q <local registry> <template registry>        → 출력 없음(바이트 동일) + 양쪽 엔트리 101
```

#### 9. 관찰 사항 (M1 판정 외)

- **동시 작성자 관찰**: M1 수행 중(04:51–04:52) 같은 워크트리의 `.moai/reports/t232/guard-failure-scenario.md`·`watch-review.sh` 가 plan-audit iter3 응답(미커밋, `plan-audit-iter3.md` 미추적 파일 수반)으로 편집됐다. M1 기록면(레지스트리 2 미러·progress.md §E.2)과 분리돼 M1 커밋에 포함하지 않았다(명시적 pathspec 스테이징). 내용은 option-C 와 정합(변이 시나리오 R2 를 비은퇴 97 로 한정).
- `moai constitution validate --json` 플래그는 이 트리에 없음(unknown flag) — 판정은 텍스트 출력 + exit 코드로 인용.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
