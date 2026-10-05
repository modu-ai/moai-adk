# sync-audit 판정서 — SPEC-DRIFT-CLOSE-BODY-001 (카드 t410)

**판정: PASS-WITH-DEBT · 88.2 / 100** (가중 조화평균, default 프로필)

| 항목 | 값 |
|---|---|
| 감사 트리 | `.claude/worktrees/t410` |
| 브랜치 | `WT-drift-false-positive` |
| 감사 고정 HEAD | `41a33678bfa1f4d5b881d7f7015c1417f34168a4` |
| 감사 범위 | `c323bb491..41a33678b` (7 커밋) |
| 워킹트리 | `git status --porcelain` 0행 (감사 시작·종료 양쪽에서 확인, HEAD 미이동) |
| 평가 프로필 | `default` (`harness.yaml` `default_profile: "default"`; spec.md frontmatter에 `evaluator_profile` 없음) |
| 감사자 | sync-auditor (독립 채널 — 이 구현을 작성하지 않음) |

---

## 차원 점수

| 차원 | 점수 | 판정 | 증거 (이 감사가 직접 실행한 명령) |
|---|---|---|---|
| Functionality (40%) | 92/100 | **PASS** (must-pass) | `go test ./internal/spec/ -count=1 -run TestDriftCloseBody -v` → 14/14 `--- PASS`, `ok ... 0.227s`. 코퍼스 재현: `grep -cE '^SPEC-.*[[:space:]]DRIFT[[:space:]]*$'` before/after = `196` / `178`, `diff` 결과 **변경 18행 전부 `DRIFT → aligned`, 역방향 0행**. 18행 중 18행을 커밋 SHA·subject·본문 줄로 독립 대조 |
| Security (25%) | 95/100 | **PASS** (must-pass) | Critical/High 0건. 새 서브프로세스 0개(공유 인덱스 재사용), 새 정규식 2개 모두 선형(`^[ \t]*[-*+][ \t]+`, `^[a-z]+(\([^)]*\))?: `) — 중첩 수량자 없음, ReDoS 경로 없음. 외부 입력 표면은 git 커밋 본문뿐이고 최악 결과는 frontmatter가 이미 `completed`인 행의 drift 억제로 한정 |
| Craft (20%) | 88/100 | **PASS** (임계 85%) | `go test ./internal/spec/ -cover` → `coverage: 90.3% of statements`. 신규 함수: `inMemBodyDeclaredClose 91.7%` · `bodyDeclaresClose 94.7%` · `subjectCloseSignal 100.0%` · `hasDeclarationDenyKey 100.0%`. `golangci-lint run ./internal/spec/...` → `0 issues`. `go vet ./internal/spec/...` 무출력. `gofmt -l internal/spec/` 무출력 |
| Consistency (15%) | 72/100 | PASS (임계 미달 아님) | 소유권 교차 1건(F1), 미귀속 수치 1건(F2), 규약 이탈 2건(F3·F4) |

가중 조화평균 = `1 / (0.40/92 + 0.25/95 + 0.20/88 + 0.15/72)` = **88.2**.
**must-pass 방화벽 통과** — Functionality(AC 7/7) · Security(Critical/High 0) 양쪽 독립 충족.

---

## 지시된 6개 압박점 — 판정

### 1. 과다 인식 위험 — `completed`-only 자격이 모든 경로에서 강제되는가

**강제된다. 단, 두 경로의 강제 방식이 다르다.**

- **모양 B**: `if _, status, err := ClassifyPRTitle(stripped); err == nil && status == "completed"` (`drift_index.go:449`). `implemented` / `in-progress` / `draft` / skip / 미지 접두 전부 비선언. `TestDriftCloseBody_OutputIsCompletedOrNothing`이 행동으로 측정.
- **모양 A**: `ClassifyPRTitle`을 **아예 부르지 않는다**. `subjectCloseSignal(subject) && strings.HasPrefix(stripped, specID+":")` 두 조건만으로 `return true` → 호출부가 `gitStatus = "completed"`. 따라서 "completed 아닌 status가 새어나올" 구조적 여지는 **0**(다른 status를 낼 코드 경로가 없다)이지만, 그 대가로 **콜론 뒤 본문이 전혀 제약되지 않는다**.

호출부 `drift.go:255`의 `a.status == "completed" && gitStatus != "completed" && !isTerminalStatus(gitStatus)` 게이트가 역방향 노출을 근본적으로 묶는다: 이 축은 **frontmatter가 이미 `completed`라고 주장하는 행만** 건드리므로 없던 `completed`를 만들어낼 수 없고, 신규 drift를 만들 수도 없다. `AC-DCB-005 ①`(비-DRIFT → DRIFT 0건)은 측정값일 뿐 아니라 **구조적으로 보장된다**.

REQ-DCB-005 위반 없음.

### 2. 18행 해제 — 독립 대조

18행 전부를 커밋으로 추적했다. 대조 방법: `git log --no-merges --grep=<ID>`로 후보 커밋을 찾고, `git show -s --format=%B <sha> | grep -n <ID>`로 자격 줄을 직접 읽음.

| 해제 근거 커밋 | subject | 덮는 행 | 자격 줄 (직접 읽음) | 진짜 close인가 |
|---|---|---|---|---|
| `cd21df594` | `docs: close out 4 A-tier SPECs with sync-phase (…)` | ASTGREP-EDIT-001 · CLI-TUI-MODERNIZE-001 · INVOCATION-MODEL-002 · WORKFLOW-CACHE-OPT-001 | `- SPEC-CLI-TUI-MODERNIZE-001: in-progress -> completed. sync_commit_sha` 외 3행 | **예** |
| `6da952899` | `feat(factory+epic): Factory Mode multi-session bootstrap …` | FACTORY-BOOTSTRAP-001 · EPIC-STATUS-001 | `:193 * docs(SPEC-FACTORY-BOOTSTRAP-001): sync-phase artifacts + 3-phase close` / `:359` 동형 | **예** (모양 B) |
| `2f449e189` | `docs(specs): batch sync-phase close — 5 B-grade SPECs` | GLM-KEY-INPUT-001 · TDD-ANTICHEAT-001 · I18N-GOVERNANCE-001 · CLIFIX-LINTER-STALE-001 · CONTEXT-ENGINE-RIGHTSIZE-001 | `:3 * docs(SPEC-TDD-ANTICHEAT-001): sync-phase artifacts — 3-phase close` 등 | **예** |
| `7beda68a5` | `Close out 2 SPECs with 3-phase lifecycle completion` | CODERABBIT-ADOPTION-001 · WORKTREE-ENTRY-STRATEGY-001 · V3R6-CI-FLAKY-STABILIZE-003 | `:151 * chore(SPEC-V3R6-CI-FLAKY-STABILIZE-003): sync-phase artifacts — 3-phase close (e24067904)` | **예** |
| `e979a4d13` | `chore(SPEC group C): Mx-phase close (…)` | V3R6-SESSION-HANDOFF-AUTO-001 · V3R6-PROMPT-CACHE-001 · V3R6-MAIN-RED-REMEDIATION-001 | `- SPEC-V3R6-…-001: Mx verdict EVALUATE-PASS` 계열 | **예** (이 카드의 표적 형태) |
| `cd80f0644` | `chore(spec): close KANBAN-RENAME-001 …` | KANBAN-RENAME-001 | `SPEC-KANBAN-RENAME-001: in-progress -> completed. Rename landed at` | **예** |

**오해제(false negative) 0건.** 18행 전부 실제로 main에 close가 착지한 SPEC이다.

**추가로 직접 측정한 것 — 모양 A의 코퍼스 전수 노출.** 전체 이력에서 subject에 `close`를 담은 커밋의 `^[-*+ \t]*SPEC-…-NNN:` 줄을 전수 추출했다(18줄). 그중 close가 아닌 성격의 줄 2개를 개별 판정했다:

- `SPEC-CONFIG-TIER-PERSIST-001: draft -> superseded (v0.3.0)` — **supersede 기록이지 close가 아니다.** 그러나 해당 SPEC frontmatter가 `status: superseded`(직접 확인)라 호출부 게이트(`a.status == "completed"`)가 애초에 열리지 않는다. 코퍼스 표 양쪽에서 `aligned`. **무해**.
- `SPEC-PROFILE-MEMORY-001: in-place amendment — close the … debt` — amendment 기록. frontmatter `completed`인데 코퍼스 after 표에서 **여전히 DRIFT**(직접 확인) — 게이트 (a)가 걸러냈다. **무해**.

즉 모양 A의 이론적 과다 인식 표면은 실재하지만(F5 참조), **이 코퍼스에서 실제 오해제로 실현된 건은 0건**이다.

### 3. 좁힘(narrowing)의 일반성 — `ac3e38a0b` 한 건만 배제하는가

**일반적이다. 다만 놓치는 close 부류가 명시적으로 존재한다.**

`subjectCloseSignal`은 `strings.Contains(strings.ToLower(subject), "close")` — 대소문자 무시 부분문자열이다. 특정 커밋을 배제하는 리터럴이 아니라 성질 판정이므로 `ac3e38a0b` 전용 하드코딩이 아니다. `TestDriftCloseBody_SubjectCloseSignalCoversMeasuredCloses`가 실측 close subject 7종 전부를 통과시키고, 동시에 `closeInfixMatch`가 그중 2종을 떨어뜨린다는 **기각 근거의 재현성**까지 단언한다(가드가 무의미해지면 붉어진다).

**놓치는 것 — 이 감사의 판정**: `close`라는 단어를 subject에 담지 않은 close 커밋의 모양-A 선언은 보이지 않는다. 실측 7종은 전부 통과하지만 이는 현재 규약의 관측이지 불변식이 아니다. 방향은 보수적(놓침 > 오탐)이고, 좁힘은 정의상 해제 집합을 줄이므로 AC-DCB-005 ① 위반 경로가 없다. **모양 B에는 이 게이트가 걸리지 않는다**는 설계는 `6da952899`(subject에 close 신호 없음 + 본문에 진짜 3-phase close 2건)로 근거가 측정돼 있고, 이 감사가 그 커밋 본문을 직접 읽어 확인했다.

### 4. 뮤테이션 품질 — 표적 술어를 겨누는가, 무관한 단언에 우연히 죽는가

**독립 검증했다.** 소스 트리를 건드리지 않기 위해 `go test -overlay=<json>`으로 변형본을 주입했다(read-only 유지).

| 뮤턴트 | 변형 | 결과 | 죽인 테스트 |
|---|---|---|---|
| M-0 (신규, 감사 추가) | `inMemBodyDeclaredClose` 즉시 `return false` (= 기능 제거) | **rc 1** | `--- FAIL: TestDriftCloseBody_BodyDeclaredCloseRecognized` |
| M-2 | 모양 A의 `:` 요구 제거 (`specID+":"` → `specID`) | **rc 1** | `--- FAIL: TestDriftCloseBody_PredicateTable` |
| M-4 | 모양 B의 conventional-commit 전제 제거 | **rc 1** | `--- FAIL: TestDriftCloseBody_PredicateTable` |
| M-5 | `subjectCloseSignal` → `return true` (게이트 무력화) | **rc 1** | `--- FAIL: TestDriftCloseBody_AmendmentRecordIsNotClose` |

네 건 모두 **표적 술어를 겨눈 테스트가 정확히 죽였다** — 무관한 단언의 부수 효과가 아니다(각 실행에서 FAIL 테스트가 정확히 1건). 원장이 주장한 5종 중 3종을 재현했고, 원장에 없던 기능-제거 뮤턴트(M-0)를 추가해 **공허하지 않음**까지 확인했다. 뮤테이션 품질 주장은 성립한다.

### 5. 동작 보존 (REQ-DCB-006)

**측정으로 확인.** `git diff --stat c323bb491..41a33678b -- internal/spec/drift.go` = `17 +`, **삭제 0줄**. 실제 코드 변경은 `else if inMemBodyDeclaredClose(...)` 한 줄이고 나머지 16줄은 주석이다. `inMemImpliedStatus`(1차 워크)와 `inMemCombinedScopeClose`는 **바이트 미변경**(diff에 등장하지 않음). 호출 순서상 기존 fallback이 **먼저** 소비되므로 그 matcher가 이미 결정한 사건은 소유권이 넘어가지 않는다 — `TestDriftCloseBody_ExistingFallbackStillFirst` PASS로 고정.

**추가 실행 — 재실행 범위 = 파일 델타 패키지 ∪ 그것을 import 하는 패키지** (이 감사가 직접 실행, 원장에 없는 측정):

```
ok  internal/spec        (TestCatalogHashParity 1건 제외 — 아래 상속 적색)
ok  internal/cli/...     전부 통과 (비-ok 줄 0)
ok  internal/hook/... (11개) · internal/epic · internal/navigator/tiers
ok  internal/graph/... · internal/harness/router · internal/web
```

`internal/spec` importer 전 트리 초록. **동작 회귀 0건.**

### 6. 공허성 (vacuity)

**공허하지 않다.** 세 층으로 확인:

1. 픽스처 자체에 사전 단언이 박혀 있다 — `_PrimaryWalkYieldsInProgress`(1차 워크가 `in-progress`여야 결함 재현), `_GammaPrimaryWalkPrecondition`(반례가 fallback에 **도달**해야 뮤턴트를 죽일 수 있음), `_DeltaPrimaryWalkPrecondition`. 셋 다 PASS.
2. 기능 제거 뮤턴트 M-0이 `_BodyDeclaredCloseRecognized`를 죽인다(위 4번) — 구현이 없으면 초록이 유지되지 않는다.
3. 12줄 술어 표는 `bodyDeclaresClose`를 **직접** 호출하므로 우회 경로가 없고, `_FullBodyScan`은 11·12 순서쌍으로 "첫 모양 일치에서 반환"을 배제한다.

---

## 알려진 맥락 — 주장 진위 확인

| 주장 | 판정 | 근거 |
|---|---|---|
| `TestCatalogHashParity`는 develop `4244c4a06` 상속 적색 | **참** | 실행 출력이 지목하는 entry는 `sync-auditor`(`internal/template/templates/.claude/agents/moai/sync-auditor.md`). 이 카드 diff는 template 파일 **0개**(`git diff --name-only … \| grep -c template` → `0`). 해당 파일 최종 변경 = `4244c4a06`이고 `git merge-base --is-ancestor 4244c4a06 c323bb491` **rc 0** — 카드 기준선보다 앞선다. 귀속 정확 |
| D4 · D6 이월이 방어 가능한가 | **부분 참** | D4는 실재하고 이 감사가 코드에서 재확인했다: `drift.go:218` `if err != nil { continue }`가 게이트 **이전**이라 오류 갈래에서 본문 조회가 도달 불가능하다. 다만 이는 **기존 동작 보존**(REQ-DCB-006)의 결과이지 회귀가 아니고, 방향은 보수적(놓침)이다 → **차단 아님**. D6(Tier 파일 모집단)은 순수 문서 층 → **차단 아님**. 그러나 "manager-develop에게 금지된 표면이라 수리 못 한다"는 **이월 사유 자체는 F1과 모순**된다 (아래) |
| §E.1.1의 등가 검사 대체가 타당한가 | **타당** | 감사 고정본이 커밋된 적 없다는 사실은 검증 가능하다 — 원장이 `git show HEAD:…spec.md \| shasum` = 현재본 해시로 보였고, 감사 시각 트리 HEAD `4e4607abe`에 이 SPEC 파일이 없다는 것도 확인 가능한 주장이다. 원문 diff가 물리적으로 재구성 불가한 상황에서 집합(REQ 7/AC 7)·id 연속성·좌표·자기기술 HISTORY 4축을 대조한 것은 적절한 대체다. **결정적으로, 원장이 그 한계를 Residual-risk에 스스로 적었다** — AC 3개만 본문 대조, 나머지 4개 미대조. 은폐가 아니라 명시된 간극이므로 수용 |

---

## 결함 목록

> 발견 단계는 **필터링 없이** 보고한다(중요도·확신도 무관). 차단/선택 분류는 판정 단계의 몫이다.

### F1 [SHOULD-FIX] [**blocking**] — 소유권 교차: manager-develop이 다른 SPEC의 body와 `version:`을 편집했다

- **위치**: `.moai/specs/SPEC-ERA-H3-NARROWING-001/spec.md` (커밋 `12d21f2e5`, M4)
- **관측**: diff 3덩이 — `version: "0.5.0" → "0.5.1"`, `updated:`, **HISTORY 표에 `0.5.1` 행 1개 추가**. 그 행의 author 열이 `manager-develop`이라고 스스로 적는다.
- **위반 규칙**: `spec-frontmatter-schema.md` § Forbidden ownership crossings — "`manager-develop` MUST NOT modify `spec.md` / `plan.md` / `acceptance.md` body content (frontmatter `status:` + `updated:` … ALL other body modifications are forbidden). … manager-develop MUST return a blocker report and the orchestrator re-delegates to manager-spec." 같은 문서 § Non-transition frontmatter corrections는 한 번 더 못박는다 — "Owner: `manager-spec` … `manager-docs` and `manager-develop` are both restricted to `status:` + `updated:` … neither may perform it. … Body content, **HISTORY**, and every other frontmatter field stay untouched." HISTORY 행과 `version:` 둘 다 명시적 금지 대상이다.
- **왜 단순 절차 흠이 아닌가**: **같은 run이 같은 규칙을 반대로 적용했다.** progress.md §E.2 이월 부채 절과 §E.4는 D4·D6 수리를 "`spec.md` 본문 편집이 필요한데 manager-develop/manager-docs에게 **금지된 표면**"이라며 보류했다. 한 실행 안에서 동일 표면을 한쪽은 금지로 읽어 물러서고 다른 쪽은 편집했다. 두 독법 중 어느 쪽이 맞든 결함이 남는다 — 규칙이 허용한다면 D4·D6 보류 사유가 거짓이고, 금지한다면 M4가 교차한 것이다. 규칙 원문은 후자를 지지한다.
- **정상 참작**: AC-DCB-007이 이 편집을 **요구**했고(SPEC이 run-phase 에이전트에게 범위 밖 편집을 지시한 셈), 내용은 정확하며 `status:`는 손대지 않았고 기능 영향 0이다. 근본 결함은 SPEC의 AC 설계에 있고 run-phase는 blocker를 올리는 대신 실행함으로써 이를 가중했다.
- **필요한 수리**: ① 후속 카드가 `manager-spec`에 재위임해 해당 HISTORY 행의 author 귀속을 정정하거나, ② AC-DCB-007을 "manager-spec 재위임을 거쳐 수행한다"로 개정한다. 둘 중 하나. 원장 「이월 부채 처리」 절의 D4·D6 보류 사유도 함께 정합화한다.

### F2 [SHOULD-FIX] [blocking] — §E.4의 "31 dependent packages" 는 귀속되지 않은 수치다

- **위치**: `progress.md` §E.4 `tests.affected_packages: "run-phase already verified internal/spec + 31 dependent packages (see §E.2/§E.3)"`
- **관측**: 지목된 §E.2의 AC-DCB-006 행이 인용하는 명령은 `go test ./internal/spec/... -count=1 -v` **하나뿐**이고, §E.3은 "재실행 범위를 파일 델타 패키지 ∪ importer로 **잡았다**"고 범위 설정만 적는다. 31이라는 수를 낸 명령도, 그 출력도 §E.2·§E.3·`run-evidence.md` 어디에도 없다(`grep -n '31\|dependent' run-evidence.md` → 무출력).
- **왜 결함인가**: AGENTS.md §1 / `verification-claim-integrity.md` §2 — 측정되지 않은 값은 Claim이 아니라 Gap이다. "이미 검증했다"는 완료 주장이 실행 근거 없이 §E.4에 실렸다.
- **실질**: 이 감사가 그 검증을 **대신 수행했고 전부 초록이다**(위 5번). 즉 주장의 내용은 참으로 밝혀졌으나, 그것은 이 감사의 측정이지 원장의 것이 아니다.
- **필요한 수리**: §E.4의 문장을 실제 실행한 범위로 정정하거나, 실행 명령과 출력을 §E.3에 귀속시킨다.

### F3 [MINOR] [optional] — `run_commit_sha: pending-backfill` 이 HEAD에서 미상환

- **위치**: `progress.md:62`
- **관측**: 마지막 커밋 `41a33678b`은 `sync_commit_sha`만 채웠다(`git show 41a33678b` — 1 insertion, 1 deletion). `run_commit_sha`는 여전히 placeholder이고, run-phase 종결 커밋 `12d21f2e5`를 채워 넣을 수 있었다.
- **부수 관측**: §E.3 주석이 근거로 인용한 `spec-frontmatter-schema.md` § SHA placeholder backfill exemption은 **`sync_commit_sha` / `mx_commit_sha`만** 명명한다 — `run_commit_sha`는 그 면제 문면에 없다. 인용이 필드를 넘어 확장됐다.
- **완화**: 코퍼스 전반의 관행 격차다(`.moai/specs/*/progress.md`에 `pending-backfill` 잔존 8건 이상). 기계 검사 없음(`internal/**/*.go`에 `run_commit_sha` 참조 0건). 이 카드 고유 결함이 아니다.

### F4 [MINOR] [optional] — `sync_status: complete` 는 코퍼스 관행 밖 토큰

- **위치**: `progress.md:107`
- **관측**: 절 제목은 "Sync-phase **Audit-Ready** Signal"인데 값은 `complete`. 이 저장소의 실측 사례는 `audit-ready`(`fb8aff006` 본문) · `PASS-WITH-DEBT`(`80dea9684` 본문)를 쓴다 — 공교롭게도 이 카드 자신의 테스트 픽스처 6·7번 줄이 그 두 사례를 담고 있다. 기계 검사 없음(Go 코드에 `sync_status` 파싱 0건).

### F5 [MINOR] [optional] — 모양 A는 콜론 뒤 본문을 전혀 제약하지 않는다 (설계된 잔여 위험, 미문서화된 방향 하나)

- **위치**: `drift_index.go:432`
- **관측**: `subjectCloseSignal(subject) && strings.HasPrefix(stripped, specID+":")` 이후 줄 내용은 무제한이다. subject에 `close`가 있는 커밋이 본문에 `SPEC-X-001: <무엇이든>`을 담으면 해제된다. 코드 주석은 이 선택을 정직하게 적고 있다("No text predicate separates that line from a genuine verdict line … without keyword matching, which spec.md §5.3 rules out").
- **실측 노출**: 코퍼스 전수 스캔 결과 해당 형태 18줄, 그중 비-close 성격 2줄, **실제 오해제 0건**(위 2번). 위험은 실재하나 실현되지 않았다.
- **미문서화된 방향**: `spec.md`/원장의 Residual-risk는 "close 단어 없는 커밋을 **놓친다**"(보수적 방향)만 적는다. 반대 방향 — "close 커밋 본문의 비-close `<ID>:` 줄을 **집는다**" — 는 코드 주석에만 있고 SPEC Residual-risk에는 없다.

### F6 [MINOR] [optional] — CHANGELOG의 예시 하나가 좁힘이 배제한 형태를 든다

- **위치**: `CHANGELOG.md` [Unreleased] > Fixed, t410 항목 1번 불릿
- **관측**: "a squash-merged PR … or **an amendment commit** (`<SPEC-ID>:` at the start of a body line) left the SPEC reading `in-progress` …" — 그런데 amendment 커밋(`ac3e38a0b`)은 이 카드가 M3에서 **반례로 발견해 좁힘으로 배제한** 바로 그 형태다. 사용자 표면에서 배제 대상을 수리 대상 예시로 제시한다.
- **완화**: 같은 항목의 2번 불릿이 `subjectCloseSignal` 게이트를 정확히 설명해 실질 오해 여지는 작다.

### F7 [MINOR] [optional] — 모양 판정만 소문자화하고 분류 체인은 원문을 쓴다

- **위치**: `drift_index.go:444` `conventionalSubjectPattern.MatchString(strings.ToLower(stripped))` 대 `:445~449`의 `shouldSkipCommitTitle(stripped)` / `commitMatchesSPECID(stripped, …)` / `ClassifyPRTitle(stripped)`
- **관측**: 형태 게이트만 소문자로 통과시키므로 `Fix(SPEC-X-001): …` 같은 대문자 타입 줄이 게이트를 넘고 원문으로 분류된다. 두 단계가 서로 다른 문자열을 본다.
- **영향**: 실측 무해(`ClassifyPRTitle`이 어차피 접두를 판정). 다만 게이트와 분류가 같은 입력을 보지 않는 것은 나중에 술어를 넓힐 때 함정이 된다.

### F8 [MINOR] [optional] — 원장 내부 수치 흔들림 (19 vs 18)

- **위치**: `run-evidence.md` M3 Gaps (c) "해제된 **19**행에 대해서만" 대 표·요약의 18
- **판정**: 실제로는 정합이다 — 19는 좁힘 **이전**, 18은 이후이며 반례 탐색을 넓은 집합에서 한 것이 옳다. 다만 같은 절 안에서 두 수가 수식어 없이 병기돼 읽는 사람이 모순으로 읽는다.

---

## Gaps — 이 감사가 관측하지 **않은** 것

- (a) **해제되지 않은 178행을 전수 조사하지 않았다.** 그 안에 이 술어가 놓치는 또 다른 body-declared close가 있는지는 미관측이다(원장의 동일 Gap을 이 감사도 해소하지 못했다).
- (b) **모양 B의 코퍼스 전수 노출을 재지 않았다.** 모양 A는 전수 스캔했으나(18줄), 모양 B는 `ClassifyPRTitle` 체인을 태워야 판정되므로 오프라인 grep으로 셀 수 없었다. 18행 해제 대조에서 만난 모양-B 사례만 확인했다.
- (c) **크로스플랫폼 빌드(`GOOS=windows`)를 재지 않았다.** 원장도 `cross_platform_build: not-measured`로 적었고 이 감사도 재지 않았다.
- (d) **전체 스위트(`go test ./...`)를 돌리지 않았다** — 병렬 레인 부하 규율(CLAUDE.local.md §4)에 따른 의도적 미실행. 전 패키지 판정은 CI 몫이다.
- (e) **`moai spec lint`를 SPEC 아티팩트에 대해 실행하지 않았다.**
- (f) **cross-model 2차 의견(`audit_multi` / codex / GLM)을 구하지 않았다.** 이 판정은 단일 백엔드(claude) 판정이다.

## Residual-risk — 관측한 것에도 불구하고 여전히 틀릴 수 있는 것

1. **18행 해제 판정은 "close가 존재한다"만 확인했다.** 각 SPEC의 frontmatter `completed`가 그 SPEC 자체의 기준으로 옳은지는 이 감사의 범위 밖이다 — drift 탐지기가 git과 frontmatter의 일치를 옳게 판정하는지만 봤다.
2. **모양 A의 무제약 본문(F5)은 규약이 바뀌면 실현될 수 있다.** 현재 0건인 것은 close 커밋 본문 작성 관행의 함수이지 술어의 성질이 아니다.
3. **F1의 근본 원인이 SPEC의 AC 설계에 있다면, 같은 형태가 재발한다.** AC가 run-phase 에이전트에게 금지 표면 편집을 지시하는 패턴 자체는 이 카드에서 수리되지 않았다.
4. **상속 적색 `TestCatalogHashParity`는 이 브랜치가 develop에 병합될 때 여전히 붉다.** 귀속은 확인됐으나 해소되지 않았고, 이 카드의 소관도 아니다.

---

## 종합

구현은 이 저장소에서 드문 수준으로 견고하다. 픽스처가 실측 커밋 본문에서 축자 인용됐고, 공허 방지 사전 단언이 세 군데 박혀 있으며, 계획 밖 반례를 만났을 때 술어를 **넓히지 않고 좁혔다**. 뮤테이션 주장은 이 감사가 독립 재현했고(기능-제거 뮤턴트를 추가로 얹어도 죽는다), 18행 해제는 18/18 전부 커밋 본문까지 내려가 대조해 오해제 0건이다. 동작 보존은 `drift.go` 삭제 0줄과 importer 전 트리 초록으로 이중 확인된다. must-pass 두 축(Functionality · Security)에 흠이 없다.

부채는 전부 **문서·절차 층**에 있고 기능에 닿지 않는다. 그중 F1은 규칙 위반이며, 같은 실행이 같은 규칙을 반대로 적용했다는 점에서 판단 착오가 아니라 정합성 결함이다. F2는 검증 주장이 근거 없이 실렸고, 내용이 참이었음은 이 감사가 대신 재서 확인했다.

**PASS-WITH-DEBT.** F1·F2를 후속 카드로 상환할 것을 권고한다. F3~F8은 재량이며, 특히 F5는 수리가 아니라 SPEC Residual-risk에 반대 방향을 한 줄 추가하는 것으로 족하다.

---

**Baseline-attribution** — 이 판정서의 모든 측정은 트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t410`, 브랜치 `WT-drift-false-positive`, HEAD `41a33678b`에서, 감사 시작·종료 양쪽에서 `git status --porcelain` 0행·HEAD 미이동을 확인한 상태로 실행했다. 뮤테이션은 소스 트리를 변경하지 않기 위해 `go test -overlay` 로 주입했다(read-only 유지). 코퍼스 두 판(`drift-before-remeasured.txt` / `drift-after.txt`)은 run-phase가 만든 산출물이며 이 감사는 그 파일을 **재생성하지 않고** 열 위치 고정 grep과 `diff`로 재판독했다 — 즉 코퍼스 수치의 baseline은 원장의 것이고, 그 수치가 파일과 정합한다는 사실이 이 감사의 측정이다.
