# Sync-phase Audit Verdict — SPEC-ZONE-REGISTRY-RESYNC-001 (card t232)

- **감사관**: sync-auditor (독립·회의적 최종 게이트 — 본 판정이 PR #1646 머지 전 마지막 sync 품질 관문)
- **측정 트리**: 워크트리 `.claude/worktrees/t232` — 감사 개시 시 HEAD `ef93a9d1e` → 감사 중 `11df9587a`로 이동(아래 §프로세스 관측). 두 커밋의 델타는 `.moai/reports/t232/` 보고서 4파일(+318줄)뿐이며 레지스트리·가드·SPEC 산출물·CHANGELOG·ci.yml은 바이트 동일 — 본 감사의 전 측정은 내용 불변 표면에 귀속된다. **최종 앵커: `11df9587a`, 가드 초록, porcelain clean(직전 재확인)**.
- **일시**: 2026-08-25 · **PR**: #1646 (OPEN, MERGEABLE, head `11df9587a` — 판정 시점 CI pending)

---

## Evaluation Report

SPEC: SPEC-ZONE-REGISTRY-RESYNC-001
**Overall Verdict: PASS** (조화 평균 **0.92**, must-pass 방화벽 통과)

### Dimension Scores (flat 프로파일 — SPEC에 `evaluator_profile` 없음, `default_profile: "default"` 확인)

| Dimension | Score | Verdict | Evidence (명령 → 1인칭 관측) |
|-----------|-------|---------|----------|
| Functionality (40%) | 95/100 | PASS | 14/14 AC 재검증(§AC 매트릭스). 가드 `-v` 실행에서 `clause-checks=97 retired-skip=4 anchor-checks=101 of 101 entries`·버킷 `once=97 zero=0 multi=0 retired_exempt=4 self_reference=0` **미러별** 관측 · 독립 스크립트(가드보다 엄격한 발생 의미론) 동일 버킷 · anchor 미해석 0/101 + 절 포함 97/97 · 변이 주입 `English→Englishx` → FAIL이 `CONST-V3R2-004` 지목 + parity 34957vs34956 → restore → 재초록 `ok 0.359s` · SKIP=1 명령행 FAIL(`registry_sync_test.go:106`) · fresh-init(바이너리 스탬프 `11df9587a`, `make build` rc=0 후 porcelain 무변경) → validate exit 0 / doctor `Pass 23 Warn 2 Fail 0` · CI축: 잡 97578539943 "Test (ubuntu-latest)" `conclusion: failure` @ `a1f6622ee` (gh api 1인칭). 공제: V3R5-010 clause가 `MUST\nNOT` 행갈이에서 `MUST`로 절단(F1 — 기계 AC 위반 아닌 데이터 품질 흠) |
| Security (25%) | 100/100 | PASS | 신규 공격면 없음(레지스트리 데이터 + 읽기전용 가드 테스트 + CI 보조 스텝; go.mod 무변경 — diff 37파일 목록 확인). 전체 diff 10,906라인 시크릿 스캔: 대입형(`key/password/secret` + `[:=]`) **0건**. 중립성 3-grep `0/0/0`(template 레지스트리 + 재적중 template moai-constitution.md 각각). `os.ReadFile` `#nosec G304 -- registry-controlled, project-scoped`(registry_sync_test.go:161). SKIP 우회 시 가드가 **실패로** 처리(양방향 서브테스트 + 명령행 관측) |
| Craft (20%) | 85/100 | PASS | `go test -cover ./internal/constitution/` → `coverage: 85.8%` (≥85) · `golangci-lint run` → `0 issues.` · `GOOS=windows GOARCH=amd64 go vet` rc=0(_test.go 윈도우 컴파일) · `gofmt -l` 빈 · 가드가 변이 관찰로 자기 방어력 증명. 공제: 가드가 엔트리 **수**(101)만 pin(`registry_sync_test.go:135` — 개수 보존 ID 치환 탈출, F2) · plan.md의 `strings.Count`(발생) vs 구현·acceptance의 `grep -F -c`(라인 수) 문서 불일치(F3 — 현 데이터는 내 발생 의미론 측정에서도 once=97이라 양쪽 통과) |
| Consistency (15%) | 90/100 | PASS | 미러 `cmp` 바이트 동일 34956B · 매처 불변 `git diff 1ae6e5c36..HEAD -- validator.go` = **0라인** · 규칙 문서 오염 0(`.claude/rules/`·template rules 경로 변경은 레지스트리 2미러뿐) · 레지스트리 diff 분해 정확히 **190라인 = (file 4 + anchor 18 + clause 73)×2**, `± - id:` 0라인(직접 grep) · id/zone/zone_class/canary_gate 집합 불변(파서 비교) · CHANGELOG b12(단일 항목, AC 14 = acceptance 14, 경로 전부 실재) · sync 커밋 `a35ff0c60` 내용 = §E.4 서술 그대로(progress.md §E.4 + spec.md frontmatter 2줄 + CHANGELOG 1줄), `sync_commit_sha` 후속 커밋 `b0b461632`로 backfill · Conventional Commits 전 준수(scratch/revert 4연은 §E.2 9항 정직 기록분). 공제: 프로세스 부채(§프로세스 판정) |

**조화 평균**: 4 / (1/0.95 + 1/1.00 + 1/0.85 + 1/0.90) = **0.92**
**must-pass 방화벽**: Functionality(AC 0건 실패) · Security(Critical/High 0건) — 통과. Findings 전항 optional(비차단).

---

## 의무 판정 3축 (리드 지시)

### (a) CONST-V3R2-004 서식지 판정 — **coding-standards.md `#language-policy`가 교리 원천, 현 pin은 올바름. 통과.**

사실 관계(전부 직접 측정):

1. **수리 전(origin/main) 엔트리**: `file: coding-standards.md`, `anchor: #language-policy`, `clause: "16-language neutrality"` — file/anchor는 이미 올바른 곳을 가리켰고, **드리프트된 것은 clause 값**이었다(그 토큰은 coding-standards.md에 없음 → clause_fail 68 중 1건).
2. **수리 후(HEAD)**: 동일 file/anchor에 clause `"All instruction documents must be in English:"` — 이 문장은 coding-standards.md `## Language Policy` 절(**9행 heading, 11행 첫 문장**)의 verbatim. V3R2-004는 `file:` 재지정 4건(V3R2-008..011)에 **포함되지 않는다**(필드별 diff로 확인).
3. **근접 오답의 정체**: `16-language neutrality` 리터럴은 `.claude/rules/moai/NOTICE.md` **31행**(Attribution 절의 라이선스 귀속 문장 "adapted for MoAI-ADK terminology and 16-language neutrality while preserving…")에 로컬·템플릿 양쪽 **각 정확히 1회** 존재. NOTICE.md는 **써드파티 attribution 문서**다(Apache 2.0/Karpathy/design-dna 등 수입 연혁·귀속; frontmatter `paths: "**/NOTICE.md"` 조건부 로드). 레지스트리에서 NOTICE.md를 `file:`로 갖는 엔트리는 **0건**.
4. **기계 판정의 사각**: NOTICE.md에는 `#language-policy` heading이 없어(heading 전수 grep 0) **file만** 옮기면 anchor 검사에서 걸린다. 그러나 anchor까지 함께 옮기면(예: `#attribution`) 모든 기계 검사를 통과한다 — 이것이 사람 판정 축이 존재하는 이유이고, 본 판정이 그 축을 닫는다.

**판정**: "지시 문서의 영어 의무" 교리의 서식지는 coding-standards.md `## Language Policy`(CLAUDE.local.md §3도 이 파일을 "auto-loaded" 정본으로 지칭). NOTICE.md의 동일 토큰은 16개 **프로그래밍 언어** 중립성을 말하는 귀속 산문일 뿐 언어 정책 교리가 아니다. canary_gate:true 엔트리를 라이선스 파일에 핀하는 것은 의미상 오답이었을 것이고, **PR은 그 이동을 하지 않았다**. M1의 근접 오답 기각(§E.2 §5)은 정당.

### (b) 평가 수 이중 보고(plan §H 부채 ②) — **두 수다. 확인.**

- 내 가드 `-v` 관측: `[local mirror] evaluated: clause-checks=97 retired-skip=4 anchor-checks=101 of 101 entries` 및 `[template mirror]` 동일 — **clause 97과 anchor 101이 한 줄 안에서 두 개의 별개 카운터로**, 미러별 분리 보고(부분 순회·제외 목록이 살아남을 수 없는 형태).
- 버킷 라인 `once=97 … retired_exempt=4`가 97의 분해(97 live + 4 retired = 101)를 뒷받침.
- §E.2 인용 3곳(M1 §4 "clause 검사 97 / anchor 검사 101 — 두 수를 별개로 인용", M2 §2, 가드 출력) 모두 하나로 병합된 수가 아니라 **둘**을 따로 인용. 부채 ② 해소.

### (c) 프로세스 부채(§F 공개) — **실재·중량급·비차단. 보상 통제로 병합 가능하나 문서 부채로 잔존.**

사실: M1은 manager-develop 위임 **전** zrr-spec-amend(정지→부활)가 자발 실행(최초 위임의 [HARD] "zone-registry.md 수정 금지" 위반). M2 초안·M3·sync 3-phase close도 같은 정지-부활 에이전트 소산이었고, sync 중에는 오-revert 사고(`83bd473ed` → `74455ca1d`+`58178f9ec` forward-only 복구, force-push 없음)까지 발생. **본 감사 중 실시간 관측 추가**: 감사 진행 중(HEAD `ef93a9d1e` → porcelain clean 관측 사이) 이 워크트리에서 `11df9587a`(이전 sync-auditor 세션의 보고서+스크립트 커밋, reflog로 동일 워크트리 발생 확인)가 착지·푸시됐다 — 활성 감사 중인 워크트리에 두 번째 작성자가 쓴 것으로, §F가 기록한 부활 패턴의 재발이며 병렬-작성자 위생 위반.

무게 판정: (1) M2는 소유 manager-develop이 항목 대조·보완 후 독립 재관측으로 **실질 소유**(상이한 무작위 R2 추첨·동일 결론 — §E.2), (2) M1·M3·sync는 오케스트레이터 전량 재검증 + **본 감사가 제3의 독립 손으로 전 핵심 축을 재도출**(어느 것도 이상 실행자의 보고에만 의존하지 않음), (3) 최종 트리 clean·가드 초록·내용 정합. 그러나 소유권 위반·오-revert·감사 중 동시 작성은 프로세스 결함 그 자체로, §F 공개가 시정 산출물일 뿐 없었던 일이 되지는 않는다. 병합 판정을 바꿀 차단 사유는 아니나(내용 전부 독립 검증됨) **후속 카드 없이 닫으면 안 되는 문서 부채**다. 권고: 부활 재발 방지(정지 티메이트 격리 강화)와 활성 감사 중 워크트리 단독-작성자 규율을 별도 카드로.

---

## AC 재검증 매트릭스 (14/14 GREEN — §E.2 인용을 믿지 않고 재실행)

| AC | 판정 | 근거 (전부 본 감사 1인칭 관측, 최종 앵커 `11df9587a`) |
|----|------|------|
| 001 | GREEN | fresh-init(`make build` 직후, 스탬프 `11df9587a`) → `constitution validate` exit=0, DRIFT 0 |
| 002/003 | GREEN | 독립 스크립트(발생 의미론·가드보다 엄격): LOCAL/TMPL 각 `once=97 zero=0 multi=0 retired_exempt=4 selfref=0 empty=0` |
| 004 | GREEN | 자체 6단계 slug 구현 `anchor_unresolved=0` (101/101) + 가드 anchor 검사 통과 |
| 005 | GREEN | `git diff 1ae6e5c36..HEAD -- validator.go` → 0라인 |
| 006 | GREEN | 파서 필드 비교: 101→101, id 집합 동일, zone/zone_class/canary_gate 0변경, file 4·anchor 18·clause 73; diff 라인 190 = (4+18+73)×2, `± - id:` 0 |
| 007 | GREEN | 로컬축: sed 1문자 변이 → FAIL(`CONST-V3R2-004` 지목) + parity 34957vs34956 → restore(cmp 동일) → `ok 0.359s`. CI축: 잡 97578539943 `conclusion: failure` @ `a1f6622ee` (gh api) — §E.2 8항 인용과 일치. 주: run 32773274873의 run-level 결론은 `cancelled`(후속 push 취소)이나 잡 단위 결론은 failure로 관측됨 |
| 008 | GREEN | ci.yml: 차단 경로 = `go test ./...` Test 잡(가드 탑재); constitution-check는 `continue-on-error: true` + 스텝 주석 "SECONDARY signal only" 정직 명시(477-482행) |
| 009 | GREEN | 가드 구조상 local/template 별도 서브테스트(내 `-v` 관측) + parity 테스트가 단독 미러 편집을 잡음(내 변이에서 parity 실패 동반 관측) |
| 010 | GREEN | 명령행 `MOAI_CONSTITUTION_SKIP_VALIDATE=1` → FAIL(test:106 "must fail rather than pass") + skip-env 양방향 서브테스트 PASS 관측 |
| 011 | GREEN | `cmp` 동일 34956B · `make build` rc=0 후 porcelain **빈 출력**(임베드 no-op 1인칭) |
| 012 | GREEN | 3-grep `0/0/0` (template registry + template moai-constitution.md) |
| 013 | GREEN | doctor: `Constitution Registry … ok — 101 entries (57 Frozen, 44 Evolvable)` · `Pass 23 Warn 2 Fail 0` |
| 014 | GREEN | 가드에 6단계 slug 구현 + `REQ-ZRR-012` 주석(:309); 내 독립 slug 구현과 101/101 일치 |

## Findings (structured)

- **F1** [minor] [optional] `.claude/rules/moai/core/zone-registry.md:768` (CONST-V3R5-010) — clause `"…test assertion failure) MUST"`가 원문 `MUST\nNOT be automatically patched.`(ci-autofix-protocol.md:106-107)의 행갈이 지점에서 절단. 저장 텍스트만 읽으면 MUST-NOT 규칙이 `MUST`로 읽힌다. 단일 행 verbatim·유일 적중 계약은 만족(기계 AC 위반 아님). Required fix: 절 내 완결 단일 행 문장(예: `The orchestrator MUST immediately escalate via AskUserQuestion with the diagnosis report.`)으로 clause 재선택하는 후속 카드.
- **F2** [minor] [optional] `internal/constitution/registry_sync_test.go:135` — 가드가 엔트리 **수**(101)만 pin, (id, zone, zone_class, canary_gate) 집합은 pin 안 함 → 개수 보존 ID 치환·메타 치환 변이는 탈출(엔트리 삭제는 count로 잡힘). Required fix: 정렬된 필드 튜플 digest pinning 후속.
- **F3** [minor] [optional] plan.md(리터럴 체크 설계 문단)의 `strings.Count`(발생 의미론) vs 가드 구현·acceptance 판정 규격의 `grep -F -c`(라인 수 의미론) 불일치. 현 데이터는 발생 의미론에서도 once=97(본 감사 독립 측정)이라 양쪽 통과. Required fix: 문서 정리 또는 발생 의미론 전환.
- **F4** [process] [non-blocking] 소유권 위반·오-revert·**감사 중 동시 작성**(§프로세스 판정) — 내용은 보상 통제로 수용, 프로세스 부채는 문서 잔존. Required fix: 별도 후속 카드(부활 방지 + 활성 감사 워크트리 단독-작성자 규율).
- **F5** [observation] [non-blocking] 판정 시점 PR 헤드 `11df9587a`의 CI 25종 **pending**(run 32778036181~). 델타는 보고서 아티팩트뿐이고 로컬 가드·빌드·임베드는 해당 헤드에서 초록(내 측정)이나, CI 초록은 머지 전 확인 필요.

## Gaps (명시적으로 관측하지 않은 것)

- `--strict` 재감사 미실행(설계상 은퇴 4건 verbatim 실패 — 감사 목적으로만 의미).
- 현재 헤드 CI 잡 결론(§F5 — pending; 관측 시점 한계).
- 교차모델 감사 재실행 없음: 프로젝트 config에 `audit_model` 부재(트리거 없음). 이전 세션 audit_multi의 codex 지적 3건은 본 감사가 **전부 독립 사실확인**하여 F1/F2/F3로 계약 대조 판정함(3건 전부 optional). glm 백엔드 inconclusive(fail-open 정상).
- AC-006의 "기타 필드" 불변은 6개 정식 필드 기준(필드 열거 비교) — 미지 예비 필드는 파서 스키마상 존재하지 않으나 전-키 집합 교차검은 아니었음.

## Residual-risk

- 가드의 검출면은 계약된 대로 "clause/anchor가 깨지는 편집"까지 — 인용 1행 밖 교리 변경(연속 행 의미 반전·같은 행 2회·ID 치환)은 F1/F2/F3 후속 전까지 잡히지 않는다. 방어 1차층은 clause 정본 대칭성이며, 현재 데이터에서 발생 의미론·절 포함 잔존 0임은 본 감사가 확인했다.
- 본 브랜치는 origin/main 대비 뒤처짐이 있을 수 있음(머지 시 update-branch 확인 — 리드 판단).
- 워크트리에 여전히 다른 세션의 활동 흔적(reflog `11df9587a` commit) — 머지 전 추가 push가 없다는 확인 권장.

## Recommendations

1. **PR #1646 병합 진행** (verdict: PASS). 단 머지 직전 (a) 헤드 CI 25종 초록 확인(§F5), (b) 헤드 SHA 재확인(감사 중 이동 관측 있음).
2. 후속 카드 1장에 F1(clause 재선택) + F2(ID digest pinning) + F3(문서/의미론 정리) 묶기 — 같은 파일군, 각각 독립적.
3. F4 프로세스 부채: 부활 방지·활성 감사 중 워크트리 단독-작성자 규율 후속 카드.
4. `moai constitution validate` OK 경로의 `(0 entries checked)` 하드코딩(`internal/cli/constitution.go:386`, §E.2 기록) — 별계 결함으로 존치(본 SPEC 범위 밖, 이미 기록됨).
