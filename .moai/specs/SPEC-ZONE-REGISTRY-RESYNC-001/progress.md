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

### M2 — 가드 착지 — 2026-08-25

가드 커밋: `49630cba2` (`internal/constitution/registry_sync_test.go` 신규 + `.github/workflows/ci.yml` validate 스텝). 검증기 코드는 무변경(D1 — `SentinelAnchorNotFound` 배선은 설계표대로 후속 후보로 유지, 테스트 측 해석기로 동등 커버리지 확보).

가드 구성(`TestRegistrySyncGuard`, 미러별 서브테스트 local·template): ① 생산 `Validate` 호출 — `Skipped==true`면 실패(REQ-ZRR-010), `DriftCount!=0`면 실패 엔트리 ID 전량 출력 ② anchor 해석(6단계 slug 규칙을 코드로 선언 + REQ-ZRR-012 주석 — "이 규칙 아래 착지 시점 17건 실패", 트리 `e0afbb53c` 시대) 전 101엔트리 ③ 리터럴 체크(`literalHitCount` = `grep -F -c` 등가, 정확히 1회 적중, 비은퇴 97 — 검증기보다 엄격) ④ D13 자기참조 금지 ⑤ 평가 수 단언(clause 97 / retired-skip 4 / anchor 101, 미러별 분리 보고 — P4/AC-ZRR-007 부분 순회 mutant 방어) + `TestRegistrySyncMirrorsIdentical`(AC-ZRR-011 바이트 동일).

**slug 재구현 교차검증**: 동일 트리에서 Go 인터프리터 anchor-checks 101/101 통과 ↔ `analyze.py` `anchor_fail=0` — 두 독립 구현이 일치(plan §H "재구현이다" 잔여위험의 관측된 근거).

#### D10 — 붉은 것을 먼저 봤다 (변이 시나리오 `guard-failure-scenario.md` R1–R4 + SKIP, 각run 개별 복원·재초록)

전제 P1–P4 충족: 변이 직전 `git status --porcelain` 빈 출력 / M1 착지 완료 / 가드 통과 / 통과 출력의 평가 수 보고(아래 6번 인용). 모든 실행 `go test -count=1 ./internal/constitution/ -run RegistrySync`(캐시 무효화).

**1) 깨끗한 트리 통과 (P3·P4)**:
```
$ go test -count=1 -v ./internal/constitution/ -run RegistrySync ; echo rc=$?
    registry_sync_test.go:160: [local mirror] evaluated: clause-checks=97 retired-skip=4 anchor-checks=101 of 101 entries
    registry_sync_test.go:160: [template mirror] evaluated: clause-checks=97 retired-skip=4 anchor-checks=101 of 101 entries
    registry_sync_test.go:181: mirrors byte-identical: 34956 bytes
--- PASS: TestRegistrySyncGuard / TestRegistrySyncMirrorsIdentical
ok  	github.com/modu-ai/moai-adk/internal/constitution	0.548s
rc=0
```

**2) R1 — `CONST-V3R2-004` clause 1문자 변이**(`English:` → `Englishx:`, 인용값 내부):
```
rc=1
        registry_sync_test.go:89: validate [local mirror]: [DRIFT] CONST-V3R2-004 @ .claude/rules/moai/development/coding-standards.md #language-policy — clause "All instruction documents must be in Englishx:" not found in source ".claude/rules/moai/development/coding-standards.md"
FAIL	github.com/modu-ai/moai-adk/internal/constitution	0.455s
```
→ 복원 후 재초록 rc=0, `git status --porcelain` 빈 출력.

**R2 — 무작위 비은퇴 추첨 1건** (시나리오 R2, 추첨 시점 기록): **`CONST-V3R2-056`** (clause 길이 39, file design/constitution.md, `[FROZEN] GAN Loop contract (Section 11)`). 1문자 변이(`Section 11x`):
```
rc=1
        registry_sync_test.go:89: validate [local mirror]: [DRIFT] CONST-V3R2-056 @ .claude/rules/moai/design/constitution.md #2-frozen-vs-evolvable-zones — clause "[FROZEN] GAN Loop contract (Section 11x)" not found in source ".claude/rules/moai/design/constitution.md"
FAIL	github.com/modu-ai/moai-adk/internal/constitution	0.438s
```
→ 복원 후 재초록 rc=0, status 빈 출력.

**3) R3 — `CONST-V3R2-004` anchor 변이**(`#language-policy` → `#language-policy-x`, 어느 heading에도 해석 안 됨):
```
rc=1
        registry_sync_test.go:133: [local mirror] CONST-V3R2-004: anchor "#language-policy-x" resolves to no heading in .claude/rules/moai/development/coding-standards.md (six-step slug rule, REQ-ZRR-002/012)
FAIL	github.com/modu-ai/moai-adk/internal/constitution	0.900s
```
주: 이 실패는 **검증기가 볼 수 없는 축**이다 — `Validate`는 anchor를 검사하지 않아 DRIFT 없이 통과했고, 테스트 측 해석기(line 133)가 잡았다. anchor 검사 배선의 존재 증명.

**4) R4 — 템플릿 미러만 변이** (동일 clause 1문자, local 무변):
```
rc=1
        registry_sync_test.go:89: validate [template mirror]: [DRIFT] CONST-V3R2-004 @ .claude/rules/moai/development/coding-standards.md #language-policy — clause "All instruction documents must be in Englishx:" not found in source ".claude/rules/moai/development/coding-standards.md"
        registry_sync_test.go:91: validate [template mirror]: drift/errors found (drift_count=1)
FAIL	github.com/modu-ai/moai-adk/internal/constitution	0.942s
```
실패 2행 전부 `[template mirror]` 지목, `[local mirror]` 실패 **0건**(`grep -c 'local mirror'` → 0) — 템플릿 면이 실패 표면임이 출력 자체로 구분됨(AC-ZRR-009).

**5) SKIP=1 + 주입 변이** (`MOAI_CONSTITUTION_SKIP_VALIDATE=1 go test …`, R1 변이 재주입 상태):
```
rc=1
WARN: validation skipped (MOAI_CONSTITUTION_SKIP_VALIDATE=1)          ← 서브테스트당 1회씩 2회 출력
        registry_sync_test.go:85: validation was bypassed (MOAI_CONSTITUTION_SKIP_VALIDATE=1): the guard must not report success when validation was skipped (REQ-ZRR-010 / D7)
FAIL	github.com/modu-ai/moai-adk/internal/constitution	0.371s
```
**정직한 의미론 관측**: `Validate` 진입부(validator.go:175)가 SKIP을 먼저 반환하므로, 가드의 실패 사유는 **변이 탐지가 아니라 '우회됨' 자체**다(line 85가 drift 판정보다 먼저 발화). 덧붙인 관측(시나리오 문서에 없는 것 — R5b로 기록): **깨끗한 트리 + SKIP=1만으로도 rc=1**(양 미러 동일 line 85) — REQ-ZRR-010의 계약("SKIP=1이어도 실패")을 변이 없이도 만족.

**6) 전 변이 복원 후 재초록 + 잔여물 검사**:
```
$ go test -count=1 -v ./internal/constitution/ -run RegistrySync ; echo rc=$?
    registry_sync_test.go:160: [local mirror] evaluated: clause-checks=97 retired-skip=4 anchor-checks=101 of 101 entries
    registry_sync_test.go:160: [template mirror] evaluated: clause-checks=97 retired-skip=4 anchor-checks=101 of 101 entries
    registry_sync_test.go:181: mirrors byte-identical: 34956 bytes
ok  	github.com/modu-ai/moai-adk/internal/constitution	0.426s
rc=0
$ git status --porcelain
(빈 출력 — 변이 잔여물 0)
```

시나리오 문서 정합: R1–R4 를 문서 그대로 실행(R2는 비은퇴 97에서 추첨해 `CONST-V3R2-056` 기록), 추가로 R5b(클린+SKIP) 관측을 덧붙였다. **CI job 결론 관측(시나리오 §3, scratch commit push + `gh pr checks`)은 M2 로컬 관측에含되지 않는다** — PR 시점(M3/PR 생성 후) 소관이며, 그 전까지 AC-ZRR-007 의 CI 축은 미충족으로 남는다.

#### CI 배선 (D6)

`constitution-check` 잡에 `moai constitution validate` 스텝 추가(빌드 후) + 스텝 주석: "this job is continue-on-error: true, so this step is a SECONDARY signal only — the blocking guard is TestRegistrySyncGuard … rides the ordinary go test ./... job". 차단 판정은 `go test ./...` 잡에서 도는 본 Go 테스트(AC-ZRR-008).

#### 부수 검증 (커밋 `49630cba2` 트리)

```
$ go build ./... → rc=0 · $ GOOS=windows GOARCH=amd64 go build ./... → rc=0
$ GOOS=windows GOARCH=amd64 go vet ./internal/constitution/... → rc=0   (테스트 파일 windows 컴파일 게이트)
$ go test -count=1 -cover ./internal/constitution/ → ok, coverage: 85.8% of statements
$ go vet ./internal/constitution/... → rc=0 · $ golangci-lint run ./internal/constitution/... → 0 issues
$ gofmt -l registry_sync_test.go → (빈, 정돈됨 — canary_test.go·validator.go의 기존 비정돈은 무변경 파일)
```

#### M2 보강 커밋 (동시 세션, 사후 검증됨) — 최종 가드 = `0b04f3412`

D10 증거 커밋(`adde4cfc9`) 이후 같은 워크트리의 병렬 세션이 가드를 2 커밋으로 보강했다(검증기·레지스트리 데이터 무변경, `registry_sync_test.go` 단일 파일):

- `ca7d966fd` — ① 리터럴 체크에 버킷 카운터(once/zero/multi/retired_exempt/self_reference) + 버킷 단언·통과 출력 보고(AC-ZRR-002/003 의 측정 형태를 통과 출력에 상시 탑재) ② 테스트 선두의 직접 환경검사(`os.Getenv(MOAI_CONSTITUTION_SKIP_VALIDATE)`) — `Validate` 의 Skipped 반환 상류 제거 시에도 우회 상태에서 실패 ③ skip-env 양방향 서브테스트 2종(`t.Setenv`, 클린 트리·변이 트리 모두 `Skipped=true`핀) — **위 R5/R5b 수동 관측이 상시 테스트로 기계화됨** ④ R1 변이 픽스처 헬퍼(`t.TempDir` 사본 — 레포 무변이)
- `0b04f3412` — R1 픽스처를 상수 고정에서 런타임 레지스트리 유도(regexp)로 변경 — 재워딩·스크래치 변이 상태와 무관하게 동작

주: 위 D10 인용의 `registry_sync_test.go:89/133/160/181` 줄번호는 `49630cba2` 시점 파일 기준이다(보강 이후 124/133→206/207/217/276 등으로 이동). 인용은 그 트리에서 실행된 관측의 기록 그대로다.

**추가 관측 (계획 외, 기록 가치)**: 보강 세션의 스크래치 R1 변이가 트리에 남아 있던 ~40초 창에 본 세션이 풀 스위트를 돌려 가드가 그 **실시간 변이**를 잡는 것을 관측했다(`[DRIFT] CONST-V3R2-004 … Englishx:` + 미러 불일치 34957 vs 34956 — 병렬 세션이 즉시 복원). 로컬 환경에서 의도치 않은 붉은 신호가 실제 붉은 트리를 정확히 지목한 사례.

**최종 상태 검증 (HEAD `0b04f3412`, 트리 클린·병렬 세션 정지 확인 후)**:
```
$ go test -count=1 ./internal/constitution/...            → ok, rc=0
$ go test -count=1 -v … -run RegistrySync                 → rc=0; [local]/[template] 각각
    evaluated: clause-checks=97 retired-skip=4 anchor-checks=101 of 101 entries
    clause literal buckets: once=97 zero=0 multi=0 retired_exempt=4 self_reference=0
    mirrors byte-identical: 34956 bytes
$ gofmt -l (신규 파일) → 빈 · golangci-lint → 0 issues · go vet → ok
$ GOOS=windows GOARCH=amd64 go vet ./internal/constitution/... → ok (테스트 windows 컴파일)
$ go build ./... · GOOS=windows GOARCH=amd64 go build ./...  → rc=0 / rc=0
```

### M3 — 미러·임베드·최종 검증 — 2026-08-25

기준: 병렬 세션 커밋 포함 브랜치 선두부(`9a1fbfdd2` — 조정자 M2 재검증 → `5e5cff235` — SplitSeq 채택). SplitSeq 리팩터링(`range strings.SplitSeq`, LSP 권고 2루프)은 본 세션이 편집했고 병렬 세션이 동일 내용을 `5e5cff235`로 커밋(편집 후 검증: `go test -count=1 ./internal/constitution/ -run RegistrySync` → `ok … 0.683s`, gofmt 빈).

#### 1. 미러 바이트 동일 + 임베드

```
$ diff -q .claude/rules/moai/core/zone-registry.md internal/template/templates/.claude/rules/moai/core/zone-registry.md ; echo rc=$?
(출력 없음) rc=0
$ cmp <local> <template> ; echo rc=$?
rc=0                                    # cmp 바이트 단위 동일
$ make build ; echo rc=$?
… catalog.yaml updated successfully (12899 bytes) …
rc=0                                    # //go:embed 재임베드 완료
$ git status --porcelain                # 빌드 직후
(출력 없음 — 빌드가 만든 미커밋 변경 0; bin/moai는 gitignored: git check-ignore bin/moai → rc=0)
# catalog.yaml 이 zone-registry를 추적하지 않음을 확인(grep -c 'zone-registry' internal/template/catalog.yaml → 0) — 레지스트리 수리는 카탈로그 재생성을 유발하지 않음(재생성 결과도 바이트 동일)
```

#### 2. 템플릿 중립성 3-grep (엄격 패턴 — acceptance.md §AC-ZRR-012; 느슨한 `SPEC-…-\d`는 산문 토큰 `SPEC-ID` 오탐이므로 기각)

```
$ grep -cE 'SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}' internal/template/templates/.claude/rules/moai/core/zone-registry.md        → 0
$ grep -cE '20[0-9]{2}-[0-9]{2}-[0-9]{2}' internal/template/templates/.claude/rules/moai/core/zone-registry.md            → 0
$ grep -cE '\b[0-9a-f]{40}\b' internal/template/templates/.claude/rules/moai/core/zone-registry.md                        → 0
# M1 이 재지정 대상으로 새로 적중시키게 된 템플릿 moai-constitution.md 도 동일 3-grep → 0/0/0
```

#### 3. fresh-init 증명 (AC-ZRR-001/015) — 새 바이너리(`./bin/moai`, v3.1.2 / commit `9a1fbfdd2` 스탬프), 리포 밖 스크래치

```
$ mkdir -p /tmp/t232-m3-*/proj && ./bin/moai init --root /tmp/t232-m3-*/proj --non-interactive --language go
init-rc=0     # 주: 원본 재현(findings.md)과 동일 플래그. --root 는 기존 디렉터리 필요 — 미생성 시
              # "Initialization failed: validate project: invalid project root path"(detector.go validateRoot).
              # 1차 시도가 이것으로 실패해 mkdir 후 재시도했다(발견한 플래그 요구사항 — gap 항목에도 기록).

$ (cd /tmp/t232-m3-*/proj && <worktree>/bin/moai constitution validate) ; echo rc=$?
constitution validate: OK — no drift or violations detected (0 entries checked)

  4 retired entry/entries skipped ([SUPERSEDED …] marker); re-check them with --strict
rc=0                                   # AC-ZRR-001 GREEN: exit 0, DRIFT 0회(전문 위 — "(0 entries checked)"는 M1 §E.2 기록된 하드코딩 리터럴)

$ (cd /tmp/t232-m3-*/proj && <worktree>/bin/moai doctor) ; echo rc=$?
    ok      Constitution Registry  registry OK — 101 entries (57 Frozen, 44 Evolvable)
   …
   Pass 23    Warn 2    Fail 0
rc=0                                   # AC-ZRR-013 GREEN — RED(§1 재현)은 fail … Pass 22 Warn 2 Fail 1 이었음
```
경고 2건(`Telemetry Config` 기본값 부재, `Glamour Cache` 미통합)은 fresh 프로젝트의 원래 상태로 본 카드와 무관 — 수정 대상 아님(그대로 보고).

#### 4. 패키지 테스트 (영향 패키지 한정 — 전체 스위트는 CI)

```
$ go test -count=1 ./internal/constitution/... ./internal/template/...
ok  	github.com/modu-ai/moai-adk/internal/template	30.026s        (+ agentemit ok, scripts [no test files])
# internal/constitution: 아래 §5 상황 기록
```

#### 5. 최종 가드 판정과 스크래치 커밋 상황 (정직한 기록)

검증 도중 병렬 세션이 **CI 결론 관측용 스크래치 커밋** `a1f6622ee`(가드 변이 시나리오 §3 — 고의 R1 변이, "revert after observation" 계약)을 브랜치 선두에 올렸다. 이 커밋 아래에서 로컬 트리 가드는 **올바르게 붉다**:

```
--- FAIL: TestRegistrySyncGuard (0.20s)
    --- FAIL: TestRegistrySyncGuard/local (0.11s)
        registry_sync_test.go:125: validate [local mirror]: [DRIFT] CONST-V3R2-004 @ .claude/rules/moai/development/coding-standards.md #language-policy — clause "All instruction documents must be in Englishx:" not found in source "…"
--- FAIL: TestRegistrySyncMirrorsIdentical (0.00s)
        registry_sync_test.go:274: registry mirrors are not byte-identical (34957 vs 34956 bytes) — repair one mirror only and the parity is gone (AC-ZRR-011)
```

즉 M3 시점의 "브랜치 선두 가드 붉음"은 결함이 아니라 스크래치 변이에 대한 가드의 정상 반응이며, 스크래치 revert 후 초록으로 돌아온다(시나리오 문서 계약대로). 스크래치 커밋의 push·CI 관측·revert는 PR 시점 소관(AC-ZRR-007 CI 축) — 본 M3 커밋은 스크래치를 포함하지 않고 위에 쌓인다. 스크래치 직전 커밋(`5e5cff235` — 본 카드 최종 데이터 상태)에서의 가드 통과는 M2 §E.2 최종 검증 블록 및 SplitSeq 검증(`ok … 0.683s`)이 이미 인용한다.

#### 6. AC 최종 상태 (M3 종료 시점)

| AC | 상태 |
|---|---|
| AC-ZRR-001 (fresh-init validate exit 0) | **GREEN** (§3) |
| AC-ZRR-002/003 (양 트리 리터럴 97/97) | **GREEN** (M1 §E.2 + M2 가드 상시화) |
| AC-ZRR-004 (anchor 101/101) | **GREEN** (M1 §E.2 + M2 가드) |
| AC-ZRR-005 (매처 불변) | **GREEN** (validator.go 무변경 — 커밋 전수 확인) |
| AC-ZRR-006 (엔트리 보존) | **GREEN** (M1 §E.2 §7) |
| AC-ZRR-007 (변이 관측) | **GREEN(로컬 축)** — CI 축(`gh pr checks` 결론)은 PR 시점 잔여: 스크래치 `a1f6622ee` 가 그 관측을 위해 대기 중 |
| AC-ZRR-008 (차단 경로) | **GREEN** (M2 — go test ./... 잡 + D6 주석) |
| AC-ZRR-009 (템플릿 커버) | **GREEN** (M2 R4) |
| AC-ZRR-010 (skip 불허) | **GREEN** (M2 R5/R5b + 상시 서브테스트) |
| AC-ZRR-011 (미러 동일 + 임베드) | **GREEN** (§1 — 가드 상시 테스트 포함) |
| AC-ZRR-012 (중립성 비회귀) | **GREEN** (§2 — 0/0/0) |
| AC-ZRR-013 (doctor Fail 0) | **GREEN** (§3) |
| AC-ZRR-014 (slug 규칙 코드 선언) | **GREEN** (M2 headingSlug + REQ-ZRR-012 주석) |

잔여: AC-ZRR-007 CI 축(PR 시점), `--strict` 재감사(설계상 은퇴 4 verbatim 실패 — 감사 목록록으로만 의미).

### M2 — 가드 완성 트리에서의 독립 재관측 — 2026-08-25

측정 트리: `WT-zone-registry-drift` @ `0b04f3412`(변이 관측·E항목 전부). 사전 점검(§C)만 `2cd846377`에서 측정.

**수행 출처 공개 (프로세스 부채, M1 과 동일 형태)**: 위 섹션(가드 착지, 트리 `49630cba2`)과 이어지는 `adde4cfc9`의 D10 관측은 **동시 작성자**(같은 워크트리의 병렬 세션)가 수행했다 — manager-develop 위임 프롬프트 수신 시점엔 해당 커밋이 존재하지 않았고, SPEC 산출물 읽기 도중 디스크에 나타났다. 위임받은 manager-develop 은 착지물을 plan.md §F M2 고정 설계와 항목별 대조해 버킷 카운팅(설계 5항)과 SKIP 가드 조항·t.Setenv 양방향 서브테스트(설계 7항)의 누락을 확인, 보완 커밋 `ca7d966fd`·`0b04f3412`로 가드를 완성했다. 본 섹션은 **완성된 가드**(버킷 라인·skip 가드 조항 포함)에 대해 R1-R4·SKIP 양방향을 독립적으로 재실행한 관측이다(R2 추첨도 독립: `CONST-V3R2-061`, 상대 관측은 `CONST-V3R2-056` — 서로 다른 무작위 대상에서 동일 결론은 가드가 특정 엔트리를 특수취급하지 않는다는 부가 근거). R1-CI 관측(시나리오 §3)은 push·PR 생성 후 별도 항으로 추가 기록된다. 병렬 세션의 자발 실행은 소유권 위반 부채로 남긴다 — sync 리뷰 판정 대상(M1 §F 공개와 동일 취급).

#### 1. 사전 점검 (Section C) — 트리 `2cd846377`

```
$ git branch --show-current && git rev-parse --short HEAD
WT-zone-registry-drift
2cd846377
(git status --porcelain → 빈 출력)

$ go test -count=1 ./internal/constitution/...
ok  	github.com/modu-ai/moai-adk/internal/constitution	0.498s   (rc=0)

$ go run ./cmd/moai constitution validate; echo exit=$?
constitution validate: OK — no drift or violations detected (0 entries checked)

  4 retired entry/entries skipped ([SUPERSEDED …] marker); re-check them with --strict
exit=0

$ diff -q .claude/rules/moai/core/zone-registry.md internal/template/templates/.claude/rules/moai/core/zone-registry.md
(출력 없음, rc=0)   grep -c '^- id: CONST-' → 양쪽 101

$ golangci-lint run --timeout=2m ./internal/constitution/...
0 issues.  (rc=0)
```

#### 2. 깨끗한 트리 통과 — 이중 카운트 + 버킷 라인 (종료조건 1)

```
$ go test -count=1 -v -run 'TestRegistrySync' ./internal/constitution/   # 트리 0b04f3412
=== RUN   TestRegistrySyncGuard/local
    registry_sync_test.go:206: [local mirror] evaluated: clause-checks=97 retired-skip=4 anchor-checks=101 of 101 entries
    registry_sync_test.go:216: [local mirror] clause literal buckets: once=97 zero=0 multi=0 retired_exempt=4 self_reference=0
=== RUN   TestRegistrySyncGuard/template
    registry_sync_test.go:206: [template mirror] evaluated: clause-checks=97 retired-skip=4 anchor-checks=101 of 101 entries
    registry_sync_test.go:216: [template mirror] clause literal buckets: once=97 zero=0 multi=0 retired_exempt=4 self_reference=0
=== RUN   TestRegistrySyncGuard/skip-env_clean_tree_fails
    (PASS — SKIP 방향 관측은 아래 4항)
=== RUN   TestRegistrySyncGuard/skip-env_mutated_tree_still_fails
    (PASS — SKIP 방향 관측은 아래 4항)
--- PASS: TestRegistrySyncGuard (0.21s)
=== RUN   TestRegistrySyncMirrorsIdentical
    registry_sync_test.go:275: mirrors byte-identical: 34956 bytes
--- PASS: TestRegistrySyncMirrorsIdentical (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/constitution	0.665s  (rc=0)
```

평가 수 분리 인용(§H 잔여위험 ②): **clause 검사 97 / anchor 검사 101, 미러별 분리** — 두 수가 다른 것이 option-C 계약이며 버킷 라인(`once=97 … retired_exempt=4`)이 이를 뒷받침한다.

#### 3. 변이 관측 R1-R4 (종료조건 2-5; 정본 `guard-failure-scenario.md` §1-§3)

각 실행은 **단일 변이만**在工作 트리에 둔 채 `go test -count=1 -run 'TestRegistrySync' ./internal/constitution/` 을 돌리고, 관측 직후 `git restore` 로 되돌린 뒤 재초록과 `git status --porcelain` 빈 출력을 확인했다(§5 순서 규율: R1→R2→R3→R4, 각자 revert+re-green).

**R1 — `CONST-V3R2-004` clause 1글자 삽입(로컬)**: `All instruction documents must be in English:` → `…Englishx:`(인용부 안, mid-span)

```
$ sed -i '' 's/…English:/…Englishx:/' .claude/rules/moai/core/zone-registry.md
$ go test -count=1 -run 'TestRegistrySync' ./internal/constitution/ ; echo exit=$?
WARN: validation skipped (MOAI_CONSTITUTION_SKIP_VALIDATE=1)      ← skip 서브테스트의 Validate WARN (정상)
WARN: validation skipped (MOAI_CONSTITUTION_SKIP_VALIDATE=1)
--- FAIL: TestRegistrySyncGuard (0.22s)
    --- FAIL: TestRegistrySyncGuard/local (0.10s)
        registry_sync_test.go:125: validate [local mirror]: [DRIFT] CONST-V3R2-004 @ .claude/rules/moai/development/coding-standards.md #language-policy — clause "All instruction documents must be in Englishx:" not found in source ".claude/rules/moai/development/coding-standards.md"
        registry_sync_test.go:127: validate [local mirror]: drift/errors found (drift_count=1)
--- FAIL: TestRegistrySyncMirrorsIdentical (0.00s)
    registry_sync_test.go:274: registry mirrors are not byte-identical (34957 vs 34956 bytes) — repair one mirror only and the parity is gone (AC-ZRR-011)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/constitution	0.506s
FAIL
exit=1
```
→ exit 1 + `CONST-V3R2-004` ID 지목 ✓. revert 후 재초록 `ok 0.409s` + porcelain 빈 ✓.

**R2 — 무작위 비은퇴 1건 clause 1글자 삽입(로컬)**: 추첨 `CONST-V3R2-061`(비은퇴 97 중 run-time 추첨, clause 길이 77자), `load brand context` → `load brandx context`

```
$ go test -count=1 -run 'TestRegistrySync' ./internal/constitution/ ; echo exit=$?
--- FAIL: TestRegistrySyncGuard (0.17s)
    --- FAIL: TestRegistrySyncGuard/local (0.08s)
        registry_sync_test.go:125: validate [local mirror]: [DRIFT] CONST-V3R2-061 @ .claude/rules/moai/design/constitution.md #31-brand-context-constitutional-parent — clause "[HARD] manager-spec MUST load brandx context before generating BRIEF documents" not found in source ".claude/rules/moai/design/constitution.md"
        registry_sync_test.go:127: validate [local mirror]: drift/errors found (drift_count=1)
--- FAIL: TestRegistrySyncMirrorsIdentical (0.00s)
    (동일 byte-parity 실패 — 34957 vs 34956)
FAIL	…	0.429s
exit=1
```
→ exit 1 + `CONST-V3R2-061` ID 지목 ✓. revert 후 재초록 `ok 0.420s` ✓.

**R3 — `CONST-V3R2-004` anchor 미해석 slug(로컬)**: `#language-policy` → `#language-policy-zzz-no-such-heading`

```
$ go test -count=1 -run 'TestRegistrySync' ./internal/constitution/ ; echo exit=$?
--- FAIL: TestRegistrySyncGuard (0.12s)
    --- FAIL: TestRegistrySyncGuard/local (0.06s)
        registry_sync_test.go:171: [local mirror] CONST-V3R2-004: anchor "#language-policy-zzz-no-such-heading" resolves to no heading in .claude/rules/moai/development/coding-standards.md (six-step slug rule, REQ-ZRR-002/012)
        registry_sync_test.go:207: [local mirror] evaluated: clause-checks=97 retired-skip=4 anchor-checks=101 of 101 entries
        registry_sync_test.go:217: [local mirror] clause literal buckets: once=97 zero=0 multi=0 retired_exempt=4 self_reference=0
--- FAIL: TestRegistrySyncMirrorsIdentical (0.00s)
    (동일 byte-parity 실패 — 34976 vs 34956)
FAIL	…	0.333s
exit=1
```
→ exit 1 + anchor 미해석 지목 ✓. anchor 실패는 Fatalf 가 아니라 Errorf 로 기록되므로 전 순회 카운터(97/101)가 실패 출력 안에 함께 찍힌다 — 부분 순회가 아님의 증거. revert 후 재초록 `ok 0.515s` ✓.

**R4 — `CONST-V3R2-004` clause 1글자 삽입(템플릿 미러만)**: 로컬은 무변경

```
$ sed -i '' 's/…English:/…Englishx:/' internal/template/templates/.claude/rules/moai/core/zone-registry.md
$ go test -count=1 -run 'TestRegistrySync' ./internal/constitution/ ; echo exit=$?
--- FAIL: TestRegistrySyncGuard (0.14s)
    --- FAIL: TestRegistrySyncGuard/template (0.07s)
        registry_sync_test.go:125: validate [template mirror]: [DRIFT] CONST-V3R2-004 @ .claude/rules/moai/development/coding-standards.md #language-policy — clause "All instruction documents must be in Englishx:" not found in source ".claude/rules/moai/development/coding-standards.md"
        registry_sync_test.go:127: validate [template mirror]: drift/errors found (drift_count=1)
--- FAIL: TestRegistrySyncMirrorsIdentical (0.00s)
    registry_sync_test.go:274: registry mirrors are not byte-identical (34956 vs 34957 bytes) — repair one mirror only and the parity is gone (AC-ZRR-011)
FAIL	…	0.379s
exit=1
```
→ exit 1 + **`[template mirror]` 표면이 지목**(local 서브테스트는 통과 — 템플릿 검증이 독립 표면임을 증명) ✓. revert 후 재초록 `ok 0.452s` ✓.

#### 4. SKIP 환경변수 양방향 (종료조건 6; AC-ZRR-010 / plan-audit iter3 C5)

**4a. 변이 트리 + SKIP=1** (R1 변이를 다시 적용한 뒤):

```
$ MOAI_CONSTITUTION_SKIP_VALIDATE=1 go test -count=1 -run 'TestRegistrySync' ./internal/constitution/ ; echo exit=$?
--- FAIL: TestRegistrySyncGuard (0.00s)
    registry_sync_test.go:106: validation skipped: MOAI_CONSTITUTION_SKIP_VALIDATE=1 present in the test environment — the registry-sync guard must fail rather than pass (REQ-ZRR-010 / AC-ZRR-010)
--- FAIL: TestRegistrySyncMirrorsIdentical (0.00s)
    (byte-parity 실패 — 변이가 여전히 있으므로 정상)
FAIL	…	0.240s
exit=1
```

**4b. 깨끗한 트리 + SKIP=1** (변이 없음):

```
$ MOAI_CONSTITUTION_SKIP_VALIDATE=1 go test -count=1 -run 'TestRegistrySync' ./internal/constitution/ ; echo exit=$?
--- FAIL: TestRegistrySyncGuard (0.00s)
    registry_sync_test.go:106: validation skipped: MOAI_CONSTITUTION_SKIP_VALIDATE=1 present in the test environment — the registry-sync guard must fail rather than pass (REQ-ZRR-010 / AC-ZRR-010)
FAIL	…	0.245s
exit=1
```

→ 양방향 모두 exit 1 + 명시적 "validation skipped" fatal(테 시작 가드 조항, `os.Getenv` 직접 검사 — `registry_sync_test.go:106`) ✓. 검증 건너뜀을 이유로 **실패**하며 결코 통과하지 않는다. 변이 제거 후 재초록 `ok 0.386s` + porcelain 빈 ✓(종료조건 7).

#### 5. 자가 검증 E1-E5 — 트리 `0b04f3412`

```
E1a  $ git diff 1ae6e5c36..HEAD -- internal/constitution/validator.go | wc -l
     0                                                    ← 매처 불변 (AC-ZRR-005)
E1b  $ diff -q <local registry> <template registry>; echo $?
     (출력 없음) diff-rc=0                                ← 미러 바이트 동일 (AC-ZRR-011 유지;
                                                            make build 재임베드는 M3 소관 — 명시적 deferral)
E2a  $ go build ./...                                      → rc=0
E2b  $ GOOS=windows GOARCH=amd64 go build ./...            → rc=0
E2c  $ GOOS=windows GOARCH=amd64 go vet ./internal/constitution/...  → rc=0   ← _test.go 윈도우 컴파일 (B1)
E3   $ go test -cover ./internal/constitution/...
     ok  …  0.764s  coverage: 85.8% of statements
E4   $ grep -rn 'AskUserQuestion' internal/constitution/ | grep -v _test.go | wc -l
     1   ← 전부 주석: amendment.go:183 "(AskUserQuestion is orchestrator-only, …)" — #851 시절 문서화 주석,
          본 SPEC diff 0라인 (git diff 1ae6e5c36..HEAD -- amendment.go → 0). 코드 위반 0.
E5   $ golangci-lint run --timeout=2m ./internal/constitution/...
     0 issues.                                            ← 신규 이슈 0 (baseline 0 유지)
```

#### 6. 가드 구현 관찰 (판정 외)

- REQ-ZRR-012 주석의 측정 트리 표기는 `294b4b6ab`(acceptance RED 기준)가 아니라 "tree e0afbb53c era"로 적혀 있다 — e0afbb53c 에서도 anchor 실패 17이 실측됐으므로(§E.2 M1 1항) 사실 관계는 정확하고, 취지("이 규칙 아래에서 착지 시점 17건이 실패했다" + 규칙 6단계 명시)는 AC-ZRR-014 를 만족한다. 동시 작성자 커밋의 표기로 그대로 둔다.
- 리터럴 체크는 발생 횟수가 아니라 **적중 라인 수**(`grep -F -c` 동등)를 센다 — M1 측정과 동일 체계이며 acceptance AC-ZRR-002 판정 규격(`grep -F -c`)과 일치한다.
- R1/R4 같은 clause 변이에서 DRIFT Fatalf 가 literal 체크보다 먼저 서브테스트를 종료시키므로 버킷 라인은 R3(에러가 Errorf 로 기록되는 anchor 변이) 출력에서 관측된다 — 두 층(validator / 독립 literal)이 각각 단독으로 변이를 잡을 수 있는 구조는 유지된다.

#### 7. 커밋

- `49630cba2` — 가드 본체 + CI 보조 스텝 (동시 작성자; 위 공개 참조)
- `ca7d966fd` — 버킷 카운팅·버킷 라인 + SKIP 가드 조항·t.Setenv 양방향 서브테스트 + go.mod 워크업 루트 해석 (manager-develop)
- `0b04f3412` — R1 fixture 런타임 파생화 (스크래치 변이 상태에서 fixture 실패 잡음 제거)


_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
