# progress.md — SPEC-REMOVAL-GUARD-EXTRAS-001

Card: t511 · Branch `WT-danger-guard-regex` · Base `0b1e27877` (`origin/develop`) · Tier M

> §F Phase 4 Mode Selection은 오케스트레이터 소관 — 첫 런 페이즈 `Agent()` 스폰 전에 이 파일에 기록된다 (orchestrator-owned placeholder).

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-07
plan_phase_artifacts: spec.md, plan.md, acceptance.md (Tier M set; research.md·progress.md 카드 관행 동반)
measured_baseline: 라이브 RED 3클래스(heredoc 데이터·unquoted 데이터·실행형 스크래치) 전부 extras 정규식 경로로 재현(verbatim 메시지 동일, tree 0b1e27877, research.md §R.2) · baseline 테스트 `ok ... 0.669s` (§R.1) · 결함 라인 사본 3곳 실측(디스패치 2곳 + testdata 신규 발견, §R.4) · plan 페이즈 코드·템플릿 변경 0건 (SPEC 아티팩트만 저작)

**사후 감사 개정 (post-audit-revision, 2026-09-07)** — plan-audit PASS 0.87 (S3 x4)에 대한 표적 응답. 판정 파일 `.moai/reports/t511/plan-audit-verdict.md`는 읽기 전용 보존. 네 해결:

- **F2 (AC-001 vs AC-008 자기 모순)**: 테스트 fixture는 R1의 실공간 예시 형태 행만 재현하고 패턴 리터럴 행은 모든 fixture에서 의도적 부재; 패턴 텍스트 필요 시 Go 프로그램적 구성 → plan.md:111 (M2.2 fixture 규율), acceptance.md:40 (AC-001 green path), acceptance.md:47 (AC-008 green path 도달 조건 명시)
- **F2c (스코프 grep 실측)**: `internal/` + `.moai/config/` 스코프 grep이 사전 수정 트리에서 정확히 3매치 실측 → acceptance.md:28 ([LEDGER-GREP3]), research.md §R.7.1
- **F1 (REQ-RGE-010 측정 AC 부재)**: 실측 AC 흡수 선택 — AC-010이 `git diff --numstat` deleted=1 added=0로 REQ-RGE-009+010을 함께 판정 (판별 뮤턴트 선택 실행) → acceptance.md:49
- **F3 (AC-009 RED 셀 4요소)**: 사후 감사 시점 라이브 재현 실행, verbatim 거부 기록 → acceptance.md:26 ([LEDGER-R4]), acceptance.md:48, research.md §R.7.2
- **F4 (docs-site 4로케일 낡은 예시)**: config-sections.md 131행 4로케일 실측 확인, sync가 같은 변경으로 갱신 의무화 → plan.md:128 (M3 §2), research.md §R.7.3

## §E.2 Run-phase Evidence

측정 좌표: 브랜치 `WT-danger-guard-regex`, base `0b1e27877`, 런 커밋 M1 `85dd4a718` · M2 `34794215f`. 사전 수정 RED는 plan 커밋 트리 `5629d9448` 위에 미커밋 테스트 파일만 얹은 상태에서 관측했다. 전체 증거 파일: `.moai/reports/t511/` (RED 3건 + GREEN 스윕 1건, M3 커밋으로 본 브랜치에 착지).

### RED (사전 수정 + 뮤턴트)

| 셀 | 트리/상태 | verbatim 요지 | 증거 파일 |
|----|-----------|---------------|-----------|
| 사전 수정 allow RED | `5629d9448` + 미커밋 테스트 | `TestDangerousRemovalDeployed_Allows*` 3종 FAIL — 배포 policy가 extras 정규식으로 heredoc 데이터 언급·unquoted 데이터 언급·실행형 스크래치 정리를 전부 `deny` 판정 (TDD RED) | `red-allow-direction-prefix.txt` |
| 뮤턴트 A [LEDGER-MUT-A] | `34794215f` + C1 라인 템플릿 재추가 + `make build` exit 0 | 허용 방향 4테스트 스윕 중 정확히 3종 FAIL로 복귀 (QuotedDataMention은 quote folding으로 GREEN 유지 — 예상대로) → 테스트가 extras 병합을 실제 로드한다는 판별 증거 | `red-mutant-a.txt` |
| 뮤턴트 B [LEDGER-MUT-B] | `34794215f` + `checkBashCommand` 구조 체크 호출 블록 임시 주석 | 차단 방향 2테스트 모두 FAIL로 복귀 → t286 구조 체크의 회귀 방어가 살아 있음 | `red-mutant-b.txt` |

두 뮤턴트 모두 실행 후 즉시 원복 — 원복 후 `git status --porcelain`은 의도된 파일만 표시했고 `pre_tool.go`는 base 대비 diff 0 (`git diff 0b1e27877..HEAD -- internal/hook/pre_tool.go` 무출력, 바이트 동일 복원). 뮤턴트 상태로 커밋한 것 없음.

### GREEN (사후, 트리 `34794215f` 클린)

- **AC 통합 스윕** (acceptance.md 채택 절차 1의 env-scrub 단일 호출): `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook/ -run 'TestDangerousRemoval' -count=1 -v` → exit 0, **스윕 10테스트 PASS / 0 FAIL** (기존 4 + 신규 6), 패키지 `ok ... 0.773s` — 전문 `green-ac-matrix-sweep.txt`
- **AC-008 스코프 grep**: 동일 명령 `grep -rn 'rm\\s+-rf' internal/ .moai/config/` — 사전 수정 실측 3매치(템플릿:12·dogfood:7·testdata:7, pre-flight 대조군) → 사후 **0매치** (exit 1). 같은 명령의 3→0 대조이므로 0이 스캔 실패가 아님. M2.2 fixture 규율 준수 — 신규 테스트 소스에 패턴 리터럴 착지 0 (실공간 예시 형태만 사용)
- **AC-010 numstat 실측** (`git show --numstat 85dd4a718`): yaml 3사본 각각 `0 1` (added=0 deleted=1), spec.md `1 1` (frontmatter status 행), catalog.yaml `4 4` (해시 미러 — 같은 커밋). 템플릿 추가 행 0 = 중립성 위반 주석 부재 + 전체 변경이 1행 삭제 = REQ-RGE-010 실측. 판별 뮤턴트(타 라인 변경 시 numstat 이탈)는 미실행 — 선택 항목

### AC 이진 매트릭스 (11건)

| AC | 판정 | 판정 근거 |
|----|------|----------|
| AC-RGE-001 | PASS | GREEN 스윕에 `TestDangerousRemovalDeployed_AllowsDeepPathHeredocDataMention` 포함 PASS + 사전 수정 RED 셀 (red-allow-direction-prefix.txt) |
| AC-RGE-002 | PASS | `..._AllowsUnquotedDataMention` PASS + 사전 수정 RED |
| AC-RGE-003 | PASS | `..._AllowsScratchCleanupExecution` PASS + 사전 수정 RED |
| AC-RGE-004 | PASS | `..._AllowsQuotedDataMention` PASS (green-now 회귀 가드 — RED 셀 불요, 배포 policy 재실행으로 고정) |
| AC-RGE-005 | PASS | `..._DeniesProtectedTargetsAllOrders` PASS (11형태 전부 deny + reason이 구조 체크 prefix) + 뮤턴트 B RED 셀 |
| AC-RGE-006 | PASS | `..._DeniesHeredocBodyProtectedTarget` PASS + 뮤턴트 B RED 셀 |
| AC-RGE-007 | PASS | 뮤턴트 A RED 셀 — 재추가 시 허용 3테스트 RED 복귀 관측 (GREEN이었다면 REQ-RGE-006 위반으로 재작업 대상이었으나 해당 없음) |
| AC-RGE-008 | PASS | 스코프 grep 3매치(사전) → 0매치(사후) 대조 실측 |
| AC-RGE-009 | PASS | M1 `make build` exit 0 (2회: M1 적용·뮤턴트 A 원복, 각각 로그 확인) + `deployedPolicy` 헬퍼가 임베디드 FS에서 편집된 yaml을 읽어 판정 — 임베디드 반영 자체 증명 |
| AC-RGE-010 | PASS | numstat `0 1` × 3사본 실측 (위) |
| AC-RGE-011 | PASS | 모든 GREEN 판정에 스윕 수 명시 (10) — `-run` 셀렉터별 스윕: 전체 10 / Allows 4 / Denies 2, 0매치 스윕 없음 |

### E2 빌드 / E3 커버리지 / E4 경계 / E5 lint

- **E2**: `go build ./...` exit 0 · `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (사전·사후 모두 실측)
- **E3**: `go test -cover ./internal/hook/... ./internal/settings/...` → `internal/hook` **85.3%** (패키지 목표 85% 충족), `internal/settings` 90.3%, 나머지 서브패키지 전부 ok (전문은 커버리지 배치 출력). **Gap**: 사전 수정 시점의 -cover 기준선을 측정하지 않아 Δ수치는 미제공 — 절대값만 근거로 제시한다. `internal/hook/mx/complexity` 83.6%는 본 SPEC이 건드리지 않은 사전 결함 영역
- **E4**: 디스패치 B3 형태(주석 필터 없음)는 사전 수정 트리에서도 0이 될 수 없는 형태다 — 사전 존재 22매치 전부 경계 규율을 서술하는 주석과 관측 파이프라인 코드. 정준 형태(`grep -v _test.go | grep -v '// '`) → 1매치 (`pre_tool.go:647` 도구명 비교 관측 분기, 사전 존재). 호출 형태 스캔(`AskUserQuestion(`) → 0매치. **본 러인 diff가 도입한 매치 0건** — 유일한 non-test 변경 후보인 `pre_tool.go`는 diff 0
- **E5**: `golangci-lint run ./internal/hook/... ./internal/settings/...` → baseline `0 issues.` = 사후 `0 issues.` — **NEW 결함 0건**. `go vet` 3패키지 exit 0

### 커밋 / push 상태

- M1 `85dd4a718` (feat) — yaml 3사본 + catalog.yaml + spec.md frontmatter 전환 · M2 `34794215f` (test) — 배포 policy 회귀 표면 · M3 (docs) — 본 문서 + 증거 파일 (+ run_commit_sha backfill 1건 예정)
- push는 **레인 제외** — 본 브랜치 push는 리드 일괄 소관 (gitflow 레인 프로토콜 §4). 수행한 push 없음

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-09-07
run_commit_sha: 6871218aa (M3 — run phase spans 85dd4a718 / 34794215f / 6871218aa; backfilled by the following commit per the D3 self-reference exemption)
evidence_path: .moai/reports/t511/ (RED 3건·GREEN 스윕 1건)
sweep_count: 10 PASS / 0 FAIL (`-run 'TestDangerousRemoval'`)
new_defects: lint 0 · vet 0 · 경계 grep 신규 도입 0

sync-phase 인도 항목 (manager-docs):

1. **docs-site 4로케일 갱신 (F4)** — `docs-site/content/{ko,en,ja,zh}/advanced/config-sections.md` 131행이 제거된 정규식을 배포 설정 예시로 문서화 중. 4-locale same-PR 규칙으로 같은 변경에서 갱신 필요. 판정: `grep -rn "rm\\\\s+-rf" docs-site/content/` → 0매치 (plan.md M3 §2)
2. **acceptance.md [LEDGER-MUT-A]/[LEDGER-MUT-B] 셀 backfill** — 소유권 경계로 manager-develop이 acceptance.md 본문을 수정하지 않았다. RED verbatim은 `progress.md §E.2` 표 + `.moai/reports/t511/red-mutant-{a,b}.txt` 로 인도됐으므로, acceptance.md 본문 셀 채움이 필요하면 manager-spec 재위임(또는 sync 페이즈에서 소유자 경유)으로 수행한다. 셀 내용 요지: 뮤턴트 A = C1 라인 템플릿 재추가 + `make build` → Allows 3종 RED / 뮤턴트 B = `dangerousRemovalTarget` 호출 블록 주석 → Denies 2종 RED, 양쪽 모두 원복·클린 확인 완료
3. **CHANGELOG** — `#1658`/`#1686` 참조 + "배포 템플릿 extras 오탐 제거" 프레이밍 (plan.md §B.12)

## §E.4 Sync-phase Audit-Ready Signal

sync_status: audit-ready
sync_complete_at: 2026-09-07
sync_commit_sha: 1c0ddf2fb (sync 커밋 본체 — D3 자기참조 면제에 따라 직후 커밋에서 backfill한 실측값)
evidence_path: .moai/reports/t511/

sync 페이즈 인도 실측 (manager-docs):

1. **docs-site 4로케일 갱신 (F4)** — `docs-site/content/{ko,en,ja,zh}/advanced/config-sections.md` 각 131행의 제거된 정규식 예시 1행을 같은 변경에서 제거 (주변 산문은 로케일별 자연 문장 유지, 예시와 배포 주장 문장만 제거). 게이트 실측: `grep -rnF 'rm\s+-rf' docs-site/content/` → 무출력, exit 1 (0매치). hugo 재빌드·내비게이션 설정 변경 없음
2. **CHANGELOG** — `[Unreleased]` § Fixed에 발행. 발행 전 중복 카운트 `grep -c 'SPEC-REMOVAL-GUARD-EXTRAS-001' CHANGELOG.md` → 0 실측 (신규 발행). 프레이밍: "배포 보안 템플릿 extras에서 대체된 위험 제거 정규식 제거" + `(#1658)` `(#1686)` + 구조 체크(dangerousRemovalTarget, develop 기착 — 본 브랜치 pre_tool.go diff 0)가 단독으로 서는 점 명시
3. **spec.md frontmatter** — `status: in-progress → completed` (status 행 1행만; 본문·HISTORY 무변경)
4. **reporter 답변 초안** — `.moai/reports/t511/reporter-replies.md` 2건 (jjjh7401 #1658, hansooha #1686). GitHub 게시 없음 — 리드 배치 push 착지 후 별도 소관이며, 커밋 순서상 본 close 커밋 이후에 착지한다

MX 태그 판정: 본 sync 변경(마크다운 문서·yaml 예시 1행 제거)에 Go 소스·내보내기 함수 변화가 없어 추가 태그 0건 — "no tags required" 판정.

## §F Phase 4 Mode Selection

- Input parameters: tier=M · scope(files)=4(보안 yaml 3사본 + 테스트 1파일) · domains=2(hook 테스트, template/config) · language mix=Go 테스트 + YAML · concurrency benefit=LOW(코딩 중심) · agent-team prereqs=미요청
- Mode evaluation: direct=미선정(의미 변경 + 테스트 수반) · serial=**선정** · fanout=미선정(코딩 중심 — Anthropic coding-task caveat) · sweep=미선정(기계적 대량 변형 아님)
- Decision: serial
- Justification: 3 yaml 1행 제거 + 배포 정책 회귀 테스트 + 뮤턴트 2회의 코딩 중심 소규모 변경이라 단일 manager-develop 순차 스폰이 유일한 합리적 축이다. 병렬화 이익이 없어 fanout은 부적합하고, 균일 변형 대량 작업이 아니어서 sweep도 아니다. 카드 t511을 전 구간(plan→run→sync) 책지는 Factory 레인 구조와도 직렬 스폰이 정합이다. plan-audit iter2 PASS 0.93(아티팩트 해시 현재) 이후 첫 run 페이즈 스폰이다.
