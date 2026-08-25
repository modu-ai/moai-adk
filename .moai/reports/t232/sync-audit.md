# Sync-phase Audit Report — SPEC-ZONE-REGISTRY-RESYNC-001

- **감시자**: sync-auditor (독립 보상 감사 — progress.md §F 공개의 소유권 위반에 대한 대체 통제)
- **대상**: PR #1646 · branch `WT-zone-registry-drift` · head `ef93a9d1e` · base `origin/main`
- **측정 트리**: 워크트리 `.claude/worktrees/t232` @ `ef93a9d1e` (감사 전제: porcelain clean, 감사 헬퍼 3종 제외)
- **일시**: 2026-08-25

## Evaluation Report

SPEC: SPEC-ZONE-REGISTRY-RESYNC-001
Overall Verdict: **PASS-WITH-DEBT** (조화 평균 0.934, must-pass 방화벽 통과)

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 0.95 | PASS | 14/14 AC 독립 재검증 (아래 §AC 재검증 매트릭스). 내 도구로 재현: 가드 이중 카운트 `clause-checks=97 retired-skip=4 anchor-checks=101` 미러별 · 자체 스크립트 리터럴 버킷 `once=97 zero=0 multi=0 retired_exempt=4 self_reference=0` 양 트리(발생 횟수 의미론 — 가드의 라인 카운트보다 엄격) · span 기반 절 포함 97/97 · anchor 미해석 0/101(자체 slug 구현) · fresh-init `init rc=0 → validate exit 0 → doctor Pass 23/Warn 2/Fail 0`(내 스크래치 `/tmp/t232-audit-init`, 트리 빌드 바이너리) · 변이 주입 `CONST-V3R2-004` 1문자 → 가드 FAIL ID 지목 + parity 34957vs34956 → 복원 → 재초록 `ok 0.325s`. 공제: V3R5-010 clause가 `MUST`/`NOT` 경계에서 절단(§F1) — 기계 AC 위반 아님, 데이터 품질 흠 |
| Security (25%) | 1.00 | PASS | 신규 공격면 없음(레지스트리 데이터 + 읽기전용 가드 테스트 + CI 보조 스텝). 템플릿 중립성 3-grep `0/0/0`(레지스트리 + 재적중된 template `moai-constitution.md`). 파일 읽기 `#nosec G304` registry-controlled 명시. `MOAI_CONSTITUTION_SKIP_VALIDATE` 우회 시 가드가 실패로 처리(skip-env 양방향 서브테스트 통과 관측) |
| Craft (20%) | 0.85 | PASS | `go test -cover ./internal/constitution/` → `ok 0.657s coverage: 85.8% of statements`(≥85%) · `golangci-lint run ./internal/constitution/...` → `0 issues.` · `GOOS=windows GOARCH=amd64 go vet ./internal/constitution/...` → rc=0(_test.go 윈도우 컴파일) · 가드는 변이 관측(D10)으로 자기 방어력을 증명. 공제: clause span 일부가 볼드 문법 중간/구절 경계에서 절단(§F2), 가드가 엔트리 수만 pin(§F3) |
| Consistency (15%) | 0.95 | PASS | 미러 `cmp` 바이트 동일 34956B · matcher 불변 `git diff 1ae6e5c36..HEAD -- validator.go` = 0라인 · 규칙 문서 본문 0변경(D5 — zone-registry.md 제외 `.claude/rules/`·template 트리 diffstat 빈 출력, 인용 위조 경로 폐쇄) · 레지스트리 diff 구성 정확히 (file 4 + anchor 18 + clause 73)×2 = 190라인, `± id:` 0라인 · Conventional Commits 전 커밋 · CHANGELOG B12 준수 · `moai spec audit` 이 SPEC finding 0(3-phase close 인식) |

**조화 평균**: 4 / (1/0.95 + 1/1.00 + 1/0.85 + 1/0.95) = **0.934**
**must-pass 방화벽**: Functionality 0.95 ≥ 임계 · Security 1.00 ≥ 임계 → 통과

### AC 재검증 매트릭스 (14/14 — §E.2 인용을 믿지 않고 재실행)

| AC | 판정 | 내 증거 (명령 → 관측) |
|----|------|----------------------|
| 001 | GREEN | `/tmp/t232-audit-init/proj`에서 `bin/moai constitution validate` → `OK — no drift or violations detected` exit=0 (4 retired skip 안내) |
| 002 | GREEN | `python3 .moai/reports/t232/sync-audit-literal.py` → LOCAL `{once:97 zero:0 multi:0 retired_exempt:4 selfref:0 empty:0}` |
| 003 | GREEN | 동일 스크립트 TMPL 면 → 동일 출력 (템플릿 트리 기준) |
| 004 | GREEN | `sync-audit-section.py` → `anchor_unresolved=0` (자체 6단계 slug 구현, 3차 독립 확인) |
| 005 | GREEN | `git diff 1ae6e5c36..HEAD -- internal/constitution/validator.go \| wc -l` → 0 |
| 006 | GREEN | 레지스트리 diff: `± id:` 0라인, 변경 95+95 = 필드 3종만. M1 §E.2 §7의 파싱 비교와 상호보완 |
| 007 | GREEN(양축) | 로컬축 1인칭: sed 변이 → FAIL `CONST-V3R2-004` 지목 → 복원 → `ok 0.325s`. CI축: §E.2 8항 `Test (ubuntu-latest)` FAILURE @ `a1f6622ee` (run 32773274873) → 최종 헤드 SUCCESS (오케스트레이터 관측 인용) |
| 008 | GREEN | `ci.yml:445-475` constitution-check `continue-on-error: true` + 신규 validate 스텝 주석이 "SECONDARY signal only… blocking guard is TestRegistrySyncGuard… rides the ordinary go test ./... job" 명시 — 스텝 억제 래핑 없음 |
| 009 | GREEN | 가드 구조: local/template 별도 서브테스트(내 -v 실행에서 분리 관측) + §E.2 R4 템플릿 단독 변이 관측 인용 |
| 010 | GREEN | 내 -v 실행: `skip-env_clean_tree_fails` / `skip-env_mutated_tree_still_fails` 양 PASS ("never passes" 메시지 관측) |
| 011 | GREEN | `cmp` rc=0 · 34956B · 가드 `TestRegistrySyncMirrorsIdentical` PASS. `make build` porcelain 무변경은 M3 §E.2 인용(재실행 않음 — Gap 기록) |
| 012 | GREEN | 3-grep `0/0/0` (레지스트리·template moai-constitution.md 각각, 엄격 패턴) |
| 013 | GREEN | 내 스크래치 `moai doctor` → `Constitution Registry ok — 101 entries (57 Frozen, 44 Evolvable)` · `Pass 23 Warn 2 Fail 0` |
| 014 | GREEN | 가드 출력에 REQ-ZRR-012 인용 anchor 실패 메시지 라인 + 6단계 slug 구현 관측 (M2 §E.2 상세와 상호보완) |

### 이름 지정 판정 — CONST-V3R2-004 (리드 명시 지시)

**판정: 올바른 서식지 — 통과.** PR의 레지스트리 엔트리는 `file: .claude/rules/moai/development/coding-standards.md`, `anchor: "#language-policy"`, `clause: "All instruction documents must be in English:"`이며, 이 문장은 coding-standards.md `## Language Policy` 절 첫 행에 그대로 존재한다(직독 확인). 이 엔트리가 이름하는 교리 — 지시 문서(에이전트 정의·커맨드·스킬·훅·설정)의 영어 의무 — 의 서식지는 바로 그 절이고, zone registry는 그 교리의 정본을 향한다. 근접 오답 후보 NOTICE.md 31행에는 실제로 `16-language neutrality` 리터럴이 존재하지만(직독 확인), 그것은 라이선스 귀속 문장("adapted for MoAI-ADK terminology and 16-language neutrality while preserving…")의 일부로 16개 **프로그래밍 언어** 중립성을 말하는 연혁 서술이지 언어 정책 교리가 아니다 — `file:`을 NOTICE.md로 바꾸면 모든 기계 검사를 통과하면서 레지스트리를 라이선스 파일에 핀하게 된다. **PR은 그 이동을 하지 않았다**(엔트리 직독 + diff의 file 재지정 4건 목록에 V3R2-004 없음 — 재지정 4건은 전부 V3R2-008..011의 CLAUDE.md→moai-constitution.md 이사). M1의 근접 오답 기각 판정(plan §H, §E.2 §5)은 정당했다.

### 의미 정합성 spot-read (§H 잔여위험 ① 대응)

- **전수 기계 확인(가드가 안 하는 검사)**: `sync-audit-section.py` — span 기준(해당 heading부터 동수준 다음 heading까지)으로 97/97 clause가 자기 anchor 절 **안에** 존재, 이탈 0. anchor 미해석 0/101.
- **사람 표본 판정: 18건 표본 / 오인용 문장 0건 / 품질 흠 1건 + 경미 4건.** 표본: V3R2-001/002/003/004/007/008/009/010/011/028/029/056/153, V3R5-001/005/007/011/013, V3R6-001. 전수 정합 — 특히 V3R2-008(spec §2.1의 패러프레이즈 대표 사례)이 이사한 교리의 원문 `All user-facing responses MUST be in the user's conversation_language.`를 정확히 인용, ci-autofix 011/013의 구 `#protected-files` 공유 anchor를 Secrets 보호/CI 인프라 보존 두 절로 분리한 것은 의미상 정확, V3R5-005 `The auto-fix loop MUST attempt at most **3 iterations**`·V3R5-007 `new commit`·V3R2-003 mx-tag `## Scope` 첫 문장 모두 원문 직독 일치.
- **품질 흠(오인용 아님)**: V3R5-010의 clause `…test assertion failure) MUST` — 원문이 행갈이로 `NOT be automatically patched.`로 이어지는 지점에서 절단. 절·anchor·교리는 정확하나 저장 텍스트가 MUST-NOT 규칙을 `MUST`로 읽히게 한다(§F1).

### CHANGELOG 검증 (B12)

- **AC 수**: 항목 문구 "14 acceptance criteria, counted against `acceptance.md`" ↔ acceptance.md 고유 AC 토큰 AC-ZRR-001..014 = **14** — 일치.
- **파일 경로**: `internal/constitution/registry_sync_test.go` 존재(가드 실행됨) · 레지스트리 양 미러 존재(cmp) · `.github/workflows/ci.yml` constitution-check 잡 존재(445행) — 전부 실재.
- **중복**: `grep -c 'SPEC-ZONE-REGISTRY-RESYNC-001' CHANGELOG.md` → 1 (단일 항목, `[Unreleased]` → `### Fixed` 최상단).
- **내용 정확성**: "clause 73 + anchor 18 + file re-point 4; 67-error clause drift → 0" — diff 구성 측정(95 = 73+18+4)과 일치.

### 교차모델 감사 기록 (audit_multi, project_root=본 워크트리)

- **claude(required)**: pass · **codex(required)**: fail — 재현된 mutant 3건 · **glm(advisory)**: inconclusive(대상 전달 불일치 — 도구 target 파라미터 교훈, 아래 Gap) → `overall: fail, disagreement_flag: advisory`. 판정 권한은 본 감사관에게 있으며 codex 3건을 계약 대조로 개별 판정했다:

| codex 지적 | 사실 확인 | 계약 대조 판정 | 분류 |
|---|---|---|---|
| ① 다중 행 HARD 규칙의 연속 행 변이(NOT→MAY)가 가드를 통과 — V3R5-010 clause가 `MUST`에서 절단 | **사실** (엔트리·원문 직독 + wrap 전수 스캔: 행갈이 연속 18건 중 의미 경계 절단은 이 1건, 무해한 구절 절단 4-5건) | REQ-ZRR-007의 가드 검출면은 "clause/anchor를 깨는 편집"이고 D3는 단일 행 span을 **계약으로** 요구 — 연속 행 검증은 계약 밖 강화 요구. 단 저장 텍스트가 MUST-NOT을 `MUST`로 표시하는 것은 데이터 품질 흠 | **optional** — 후속 카드 (clause를 절 내 완결 문장으로 재선택 + "완전 논리 규칙 블록" 증인은 계약 변경 동반하므로 별도 설계) |
| ② 가드가 엔트리 **수**(101)만 pin, ID/zone/canary 집합은 pin 안 함 — ID 치환 변이 통과 | **사실** (`registry_sync_test.go:135` `len(reg.Entries) != wantRegistryEntries`, 주석도 "entry-set size"라고 정직) | AC-ZRR-006은 **수리 시점** 성질(§E.2 §7 파싱 비교 + 내 diff 분석으로 확인)이고 가드 설계표(plan §F M2 ⑤)는 평가 수 pin만 규정. 엔트리 **삭제**는 count로 잡힘; 개수 보존 ID 치환만 탈출 | **optional** — golden snapshot pinning 후속 카드 |
| ③ literalHitCount가 발생 횟수가 아니라 **적중 라인 수**를 센다(동일 행 2회 적중이 once로 분류) | **사실** (코드 직독, 실패 메시지 "hits %d lines") | acceptance.md AC-ZRR-002 판정 규격은 `grep -F -c`(=라인 수)이며 progress §E.2 §6이 이 선택을 명시 공개 — plan.md:108의 `strings.Count` 언급이 이상치. 현 데이터는 더 엄격한 발생 의미론에서도 once=97(내 독립 측정) | **optional** — plan/acceptance 문서 불일치 정리 + `strings.Count` 전환 후속 |

### 프로세스 부채 판정 (§F 공개에 대한)

§F는 M1 위임 전 자발 실행(TaskStop 이후 착지), M2 초안·M3·sync 3-phase close까지 정지-부활 에이전트 소산이었음을 공개하고, revert 사고(`83bd473ed` 오-revert → `74455ca1d`+`58178f9ec` forward-only 복구, force-push 없음)를 기록한다. 판정: **보상 통제로 병합 충분, 부채는 문서로 잔존.** 근거: (1) M2는 소유 manager-develop이 항목 대조·보완 후 독립 재관측으로 실질 소유했고(R2 추첨 상이·결론 동일), (2) M1·M3·sync는 오케스트레이터 전량 재검증 + **본 감사가 제3의 독립 손으로 거의 전 축을 재도출**했다(리터럴 버킷·anchor·절 포함·미러·matcher·중립성·fresh-init·변이 사이클·CHANGELOG·CI 배선 — 어느 것도 이상 실행자의 말에만 의존하지 않는다), (3) 최종 트리는 porcelain clean·가드 초록·`moai spec audit` finding 0, (4) 부활 메커니즘 교훈이 auto-memory로 기록됐다. 소유권 위반 자체는 프로세스 결함이므로 병합 판정에 반영하지 않되 §F 기록이 시정 산출물로 남는다.

### Findings (structured)

- **F1** [minor] [optional] `.claude/rules/moai/core/zone-registry.md:768`(CONST-V3R5-010) — clause `…assertion failure) MUST`가 `MUST\nNOT be automatically patched.`의 행갈이 지점에서 절단 — 저장 감사 텍스트가 MUST-NOT 규칙을 `MUST`로 표시, 가드는 연속 행 변이에 불가시. Required fix: 절 내 완결 단일 행 문장(예: 3행째 `The orchestrator MUST immediately escalate via AskUserQuestion with the diagnosis report.`)으로 clause 재선택하는 후속 카드.
- **F2** [minor] [optional] 레지스트리 전반 — clause span 다수가 행 끝이 아닌 구절 중간에서 끝남(예: V3R5-013 `…scripts or`, V3R2-028/029 볼드 문법 중간). 계약(단일 행 연속 verbatim·유일 적중) 위반 아님 — 표본 직독에서 의미 오인용 0건. Required fix: 없음(후속 미용 패스 선택).
- **F3** [minor] [optional] `internal/constitution/registry_sync_test.go:135` — 엔트리 수 pin만 존재, ID/zone/zone_class/canary 집합 pin 부재(개수 보존 치환 탈출 가능). Required fix: 정렬된 (id, zone, zone_class, canary_gate) 튜플 digest pinning 후속 카드.
- **F4** [minor] [optional] `internal/constitution/registry_sync_test.go:363` — literalHitCount 라인 수 의미론(grep -F -c 동등, acceptance 규격과 일치·progress 공개)이나 plan.md:108 `strings.Count`와 불일치. Required fix: 문서 정리 또는 발생 의미론 전환(현 데이터는 양쪽 다 통과 — 내 측정).

### Gaps (미검증)

- `make build` porcelain 무변경(AC-011 임베드 축)은 M3 §E.2 인용만으로 대체 — 본 감사가 재실행하지 않음(바이너리 05:47 빌드 존재를 fresh-init에 사용했으나 임베드 no-op 자체는 미관측).
- CI 잡 결론의 1인칭 관측 아님 — 변이 헤드 FAILURE(run 32773274873)는 §E.2 8항 인용. 로컬 변이 사이클과 CI 배선 독해로 대리 확인.
- `--strict` 경로 미실행(설계상 은퇴 4건 verbatim 실패 — 감사 목적으로만 의미).
- audit_multi의 glm 백엔드가 target 파라미터를 해석 못 해 inconclusive — 도구 target은 enum(`uncommittedChanges`/`baseBranch`)으로만 전달해야 한다(fail-open 정상 처리).
- 문서 2건의 trailing whitespace(codex 관측 — `git diff --check` 소음 수준).

### Residual-risk

- 가드의 검출면은 계약된 대로 "clause/anchor가 깨지는 편집"까지다 — 인용 1행 밖의 교리 변경(연속 행·같은 행 2회·ID 치환)은 F1/F3/F4 후속 전까지 잡히지 않는다. 방어의 1차층은 그대로 clause 정본 대칭성이며, 본 감사의 발생 의미론·span 검사가 현재 데이터에서 잔존 0임은 확인했다.
- codex가 지적한 대로 본 브랜치는 origin/main 대비 3 커밋 뒤짐(중첩 CHANGELOG.md) — 병합 시 update-branch 필요 가능. 리드 판단 사항.

### Recommendations

1. PR #1646 병합 진행 (verdict: PASS-WITH-DEBT).
2. 후속 카드 1장으로 F1(clause 재선택) + F4(문서/의미론 정리) + F3(ID digest pinning) 묶기 — 각각 독립적이나 같은 파일군.
3. 병합 전 origin/main 3커밋 동기화(update-branch) 확인.
4. "완전 논리 규칙 블록 증인"(codex 권고)은 D3 계약 변경을 동반하므로 후속 설계 논의로 — 이 SPEC 범위 아님.

## 증거 산출물 (본 감사 도구 — 재현용)

- `.moai/reports/t232/sync-audit-literal.py` — 독립 리터럴 버킷(발생 의미론, YAML 인용부 strip)
- `.moai/reports/t232/sync-audit-section.py` — anchor 해석 + span 기반 절 포함 전수 검사
- `.moai/reports/t232/sync-audit-wrapscan.py` — 행갈이/구절 절단 전수 스캔(F1 규모 측정)
