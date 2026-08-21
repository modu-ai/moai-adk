# SPEC 감사 보고서 (독립 iteration 2): SPEC-CODEX-SKILLS-CANONICAL-001

Verdict: **FAIL**
Overall Score: **0.7625** (산술평균, iter-1 과 같은 산출 방식). 조화평균으로 계산하면 **0.747**. Tier M PASS 임계 **0.80** — 어느 방식이든 미달이며, 그와 별개로 아래 blocking 6건이 독립적으로 FAIL 을 강제한다.

감사 대상 스냅샷: `spec.md` v0.3.0 (`02:54:59`) / `acceptance.md` (`02:56:51`) / `plan.md` (`02:58:09`). 감사 시작부터 종료까지 세 파일 mtime 무변경 — 개정 중간 상태를 읽지 않았다.

Reasoning context ignored per M1 Context Isolation — 리드 지시문에 담긴 작성자 주장("무엇을 고쳤다")은 판정 근거로 쓰지 않고 전부 아티팩트에서 재확인했다. 모든 수치는 이 worktree(`WT-skills-canonical`, go1.26.x darwin/arm64)에서 직접 실행한 명령의 출력이다. iter-1 보고서와 `m1-preflight-measurements.md` 의 숫자는 인용하지 않고 **재측정**했다.

동시 감사 기록: 이 감사 진행 중 다른 감사자가 `.moai/reports/t81/plan-audit.md` 를 `03:02:46` 에 갱신했다(iteration 2 표기, FAIL 0.78). 그 파일은 지시대로 **건드리지 않았다**. 그 보고서와 겹치는 항목(N4)·갈리는 항목(그 보고서의 N10/N11)은 해당 위치에 명시했고, **N4 는 독립 발견이 아니라 독립 검증**임을 §N4 에 그대로 적었다(아래 §"감사 범위" 참조).

---

## 감사 범위 — 완료한 검사와 도달하지 못한 검사

리드 지시로 **조기 종료**했다(병렬 감사 완료 + iteration 3 개정 착수 임박 → 이 시점 이후의 측정은 곧 바뀔 트리를 읽게 된다). 아래 판정은 **완료한 검사에 한해** 유효하다.

**완료한 검사**

- Must-pass 7건 전부(MP-1 ~ MP-7) — 각 항목에 실행 명령과 출력 인용.
- Group 1 frontmatter 12필드 / Group 2 절 구성 / Group 3 REQ 16개 GEARS·번호 / Group 4 AC 16개 형식 / Group 5 언어 중립성(N/A) / Group 7 D7 / Group 8 D8.
- 추적성 양방향 전수 대조(REQ 16 → AC, AC 16 → REQ) — §D 매트릭스 16행과 §D.3 을 각각 손으로 맞춰봤다.
- iteration 1 결함 D1~D13 **전건** 회귀 확인.
- spec §A.1 / §A.3 / §A.4 / §A.6 / §A.7 / §A.8 / §A.9 의 실측 주장 **전건** 재측정(아래 "재현한 실측" 표).
- `.claude/skills/` 34개 이름 접두 전수 계수, 템플릿 트리 링크 카운트, `catalog.yaml` tier 계수.
- Go 프로브 1회 실행: 디렉터리 링크에 대한 `os.Open`+`io.Copy`, `WalkDir` 의 `ModeSymlink` 보고, `os.Stat`/`os.Lstat` 차이, dangling 링크의 `os.Stat`·`os.Lstat`·`filepath.Glob` 거동. 프로브 디렉터리는 삭제 완료.
- 코드 seam 확인: `ManagedCleanTargets` / `CleanMoaiManagedPaths` / `backupThenRemove` / `templateManagedPaths` / `manifest.Track` / `HashFile` / `deployer` 구조체·`Deploy` 서명 / `IsUserOwnedNamespace` / `EmbeddedMoaiSkillNames` / `update_template_sync.go` 스텝 순서.

**도달하지 못한 검사 (미검증 — PASS 로 읽지 말 것)**

- **Group 6 일관성(CN-1~CN-3) 전수 점검**. 발견한 모순 3건(N1·N2·N3)은 다른 검사 중에 걸린 것이고, REQ 16개 × 16개 쌍대 대조를 **끝까지 돌리지 못했다**. 남은 모순이 없다는 뜻이 아니다.
- **`plan.md` §F 마일스톤 ↔ AC 닫힘 조건의 전수 대조**. M2·M3·M6 만 확인했고 M1·M4·M5 의 닫힘 조건 목록이 실제 AC 범위와 맞는지는 보지 않았다.
- **`plan.md` §B 위험표 R1~R12 의 완화 매핑 검증**. 표에 적힌 REQ·AC 번호가 실제 내용과 맞는지 확인하지 않았다.
- **`progress.md` 검토**. Tier M 입력 계약 밖이라 mtime·`syscall`·`[NEEDS CLARIFICATION]` 스캔에만 포함했고 본문은 읽지 않았다.
- **슬림/전량 배포 경로의 실제 실행 확인**. `shouldDistributeAll` 분기와 `NewSlimDeployerWithRenderer` 존재만 확인했고, AC-CSC-003 의 3번 단언(슬림 < 전량)이 실제 배포에서 성립하는지는 실행하지 않았다(정적으로 21 < 34 이므로 성립할 것으로 보이나 **미실행**).
- **`.gitignore` 항목 형태가 링크/복사 양쪽에서 git 이 실제로 무시하는지**의 실동작 확인. 항목 부재만 확인했다.
- **Windows 경로 실동작**. `GOOS=windows` 컴파일조차 돌리지 않았다 — AC-CSC-007 판정은 코드 독해 기반이다.
- 점수는 위 완료 범위에서 매겼다. 미도달 검사에서 결함이 더 나오면 **점수는 내려갈 수 있고 올라가지 않는다**.

---

## Must-Pass 결과

- **[PASS] MP-1 REQ 번호 일관성** — `grep -o 'REQ-CSC-[0-9]\{3\}' | sort -u` → `REQ-CSC-001 … REQ-CSC-016` 연속 16개, 결번·중복 0, zero-pad 일관. 정의 라인 `grep -c '^- \*\*REQ-CSC-'` = **16** 로 일치. Tier M 상한 16(`spec-workflow.md:152`)에 정확히 도달, 초과 아님.
- **[PASS] MP-2 GEARS 준수 (요구사항 계층 한정)** — 16개 전부 GEARS 5패턴에 대응한다: ubiquitous(001·007·010·015·016), unwanted `shall not`(002), event-driven `When`(003·005·008·011·012·013·014), where-gate `Where`(004), state-driven `While`(009), 파생조건+`shall not`(006). 비형식 서술("should" 류) 0건. `acceptance.md` 의 Given-When-Then 은 **검증 계층(AC-XXX)** 이므로 이 판정 대상이 아니며 Group 4 에서 채점했다 — 이 MP-2 판정은 **요구사항 계층(spec.md §C)에 대해서만** 내렸다.
- **[PASS] MP-3 YAML frontmatter 유효성** — canonical 12필드 전부 존재·타입 적합(`id` / `title` / `version: "0.3.0"` quoted semver / `status: draft` ∈ enum / `created`·`updated` ISO / `author` / `priority: P2` / `phase` / `module` / `lifecycle: spec-anchored` / `tags` CSV 문자열). 거부 alias(`created_at`·`updated_at`·`labels`·`spec_id`) 0건. 추가 필드 `tier: M` 은 허용 범위.
- **[N/A] MP-4 §22 언어 중립성** — 본 SPEC 은 moai-adk 자체의 Go 배포기·템플릿 트리를 다루며 16개 지원 언어의 도구 체인을 열거·차등하지 않는다. 단일 언어 범위 → 자동 PASS.
- **[PASS] MP-5 D7 교차-SPEC 정합** — 아티팩트 4종에서 추출된 외부 SPEC 참조는 `SPEC-CODEX-PHASE2-001` 1건(`progress.md` 의 중복 검사 기록). `.moai/specs/SPEC-CODEX-PHASE2-001/spec.md` 존재, `status: completed` — retired/superseded/archived 아님. BLOCKING 없음.
- **[PASS] MP-6 D8 크로스 플랫폼 규율** — `grep -c syscall` = spec 0 / plan 0 / acceptance 0 / progress 0. 자동 PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-CODEX-SKILLS-CANONICAL-001/` → 매치 **0**. Tier M 이라 `research.md` 부재는 정상.

Must-pass 7개 전부 통과. **FAIL 은 must-pass 위반이 아니라 집계 점수 미달 + blocking 6건에서 나온다.**

---

## 회귀 점검 — iteration 1 결함 D1~D13

| ID | 등급(iter-1) | 판정 | 근거 |
|---|---|---|---|
| D1 | blocking/critical | **CLOSED** | 아래 상세 |
| D2 | blocking/critical | **CLOSED** (잔여 N8) | AC-CSC-010 이 "동일 프로세스 seam 토글 불변식"으로 재작성됨(`acceptance.md:104-112`), `[HARD]` 로 커밋 기준선 형태를 금지. AC-CSC-013(2) 이 그 불변식을 참조(`:146`). 커밋 SHA 대조는 §D.4 비-AC 로 강등(`:188`). AP-9 도 신설. |
| D3 | blocking/major | **CLOSED** (전제 오류 N3) | REQ-CSC-015 신설(`spec.md:165`) + AC-CSC-016(`acceptance.md:165-171`) + plan R11·M6. |
| D4 | blocking/major | **CLOSED** (부작용 N2) | REQ-CSC-010 방향 반전(`spec.md:160`) + AC-CSC-012 2번 팔(`acceptance.md:135`) + plan R8·M3. |
| D5 | blocking/major | **CLOSED** (범위 과다 N7) | REQ-CSC-016 신설(`spec.md:166`) + AC-CSC-015(`acceptance.md:155-163`) + plan M5 Priority Low→High. |
| D6 | blocking/major | **CLOSED** | REQ-CSC-013·014 신설(`spec.md:163-164`) + AC-CSC-011 3-상태(`acceptance.md:114-126`) + AP-10. |
| D7 | blocking/major | **CLOSED** | §A.6 실측 추가(`spec.md:75-88`), REQ-CSC-010 반전, plan M3 재작성. |
| D8 | blocking/minor | **CLOSED** | AC-CSC-007 이 `filepath.ToSlash(t.DisplayPath) == ".agents/skills/moai*"` 로 바뀌고 `[HARD]` 로 리터럴 비교 금지(`acceptance.md:77-83`). |
| D9 | optional/minor | **CLOSED** | §A.3 이 "`optional-pack:*` 13, `harness_generated` 스킬 0" 으로 정정(`spec.md:52`). 재측정: `grep -c 'tier: optional-pack' catalog.yaml` = **13**, `harness_generated.skills:` = **[]**(`catalog.yaml:261-262`), 해당 tier 유일 항목은 에이전트 `builder-harness`. 정정이 정확하다. |
| D10 | optional/minor | **CLOSED** | §D.4 Codex 수동 확인 항목이 "상대 디렉터리 링크 형태"를 명시(`acceptance.md:190`). |
| D11 | optional/minor | **CLOSED** | §B.D5 에 `[HARD]` 문단 추가(`spec.md:147`). 재측정으로 사실 확인: `IsUserOwnedNamespace`(`internal/cli/update/plan/plan.go:152-177`)의 판정 4개가 전부 `.claude/` 접두(`.claude/skills/hns-`, `.claude/skills/harness-`, `.claude/skills/my-harness-`, `.claude/agents/harness`)이며 `.agents/` 는 시야에 없다. 추가된 서술이 정확하다. |
| D12 | optional/minor | **기각(justified)** | 아래 "기각 판단" 절. |
| D13 | optional/minor | **CLOSED** | AC-CSC-003 에 3번 단언("슬림 집합 크기 < 전량 집합 크기") 추가(`acceptance.md:57`). |

**D1 상세 — 실제로 닫혔는지 재유도.** 새 AC-CSC-001(`acceptance.md:30-41`)의 1번 단언은 `filepath.WalkDir` + `d.Type()&fs.ModeSymlink != 0` 로 **트리 전체** 링크를 센다. 이 형태가 실패 케이스를 실제로 잡는지 직접 실행했다:

```
AC001 symlink seen by WalkDir: lnk IsDir()= false
AC001 symlink count via WalkDir: 1
AC001 os.Stat.IsDir= true  os.Lstat.IsDir= false
```

`WalkDir` 은 링크를 따라가지 않되 **항목 자체는 보고**하며 `Type()` 에 `ModeSymlink` 가 서 있다. 따라서 템플릿 트리에 디렉터리 링크를 넣으면 카운트가 1 이 되어 **FAIL 한다** — iter-1 이 지적한 "자기가 막으려는 실패에서 통과" 형태가 제거됐다. 2번 단언도 `d.IsDir()` 사용을 명시 금지하고 `os.Stat` 기반 수집을 요구하며, 그 판단의 근거(`Lstat` 이라 링크가 양쪽 집합에서 동시에 빠진다)가 본문과 AP-8 에 적혀 있다. 범위도 `.claude/skills/` → 트리 전체로 넓어져 REQ-CSC-002 의 실제 범위와 일치한다. **CLOSED.**

**참고 — 다른 감사자의 N10/N11 은 현재 디스크 상태에서 성립하지 않는다.** `plan-audit.md`(03:02) 는 `plan.md` §F M3 이 "M3 — manifest 기록 (Priority Medium)" 이라 REQ-CSC-010 과 모순된다고 적었으나, 내가 읽은 `plan.md`(mtime 02:58:09, 감사 전 구간 무변경)의 `:64` 는 **"### M3 — 미러를 기록·백업 대상에서 제외 (Priority High)"** 이고 본문이 반전 사유(§A.6·§A.7)를 담고 있다. 그 지적은 개정 전 스냅샷을 인용한 것으로 보이며, 나는 모순을 관측하지 못했다.

---

## 범주별 점수 (rubric 앵커)

| 차원 | 점수 | 밴드 | 근거 |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | REQ 16개 대부분이 단일 해석을 갖고 iter-1 의 모호성(대상 선점·manifest 방향)이 해소됐다. 감점 둘: REQ-CSC-001 이 **무조건 `shall`** 인데 REQ-CSC-011·014 가 예외를 만들어 요구사항끼리 충돌한다(N1). §A.9 가 근거로 든 코드 불변식이 실제 상수와 다르다(N3). |
| Completeness | 0.70 | 0.75 밴드 미달 | 절 구성·frontmatter·Out of Scope(H3 4개, 각 `-` 불릿 보유) 완비. 그러나 부하를 지는 두 항목이 REQ·AC 어디에도 없다: **미러의 실제 제거 의미론**(dangling 링크 — N4)과 **모드 보고가 나갈 출력 표면**(N6). REQ-CSC-008 은 등록만 규정하고 "무엇이 실제로 지워지는가"를 규정하지 않는다. |
| Testability | 0.65 | 0.50–0.75 | AC 계층은 크게 좋아졌다(AC-001 양팔 / AC-010 동일-프로세스 불변식 / AC-011 3-상태 / AC-012 양팔 / AC-007 구분자 중립). 그럼에도 **MUST 인 AC-008·009 가 이 SPEC 최대 실패 형태를 구조적으로 검출하지 못하고**(N5), AC-006·011(3)·013(3) 세 개가 **존재하지 않는 출력 seam** 을 전제한다(N6). MUST 13개 중 5개의 판정력에 구멍이 있다. |
| Traceability | 0.95 | 1.0 근접 | §D 매트릭스 16행 + §D.3 역방향 매핑이 REQ 16개를 전부 덮는다(직접 대조: 001→AC-002/003, 002→001, 003→004, 004→005, 005→006, 006→003/014, 007→010, 008→007/008, 009→008/009, 010→012, 011→013, 012·013·014→011, 015→016, 016→015). 고아 AC 0, 미커버 REQ 0. 의도적 다대일 두 건(REQ-009, REQ-012/013/014)은 §D.3 에 사유가 적혀 있다. |

- 산술평균 = (0.75 + 0.70 + 0.65 + 0.95) / 4 = **0.7625**
- 조화평균 = 4 / (1/0.75 + 1/0.70 + 1/0.65 + 1/0.95) = **0.747**

**점수 추이 — LEAN STOP 조건에 해당한다.** iter-1 **0.775** → iter-2 **0.7625**(같은 산술 방식). 개정이 blocking 8건을 닫았음에도 총점이 **내려갔다** — 닫은 만큼의 새 blocking 이 개정 자체에서 생겼기 때문이다(N1·N2·N3). 다만 두 점수는 감사자가 다르므로 채점 편차가 섞여 있다는 점을 명시한다. Tier M 반복 상한 2 에 이미 도달했으므로, STOP 여부와 무관하게 다음 행동은 사용자 선택 게이트다(아래 Recommendation).

---

## Defects Found

### N1 — REQ-CSC-001 은 무조건 `shall` 인데 REQ-CSC-011·014 가 예외를 만든다 (blocking)

`spec.md:151`(REQ-CSC-001) ↔ `spec.md:161`(REQ-CSC-011) ↔ `spec.md:164`(REQ-CSC-014) — Severity: **major** — Class: **blocking**

REQ-CSC-001: "배포기는 이번 실행에서 `.claude/skills/<name>/` 에 배포한 **모든** 스킬에 대해 `.agents/skills/<name>/SKILL.md` 로 읽을 수 있는 접근 경로를 제공해야 한다(shall)." 조건절이 없다.

그런데 이번 개정이 추가·유지한 두 요구사항이 정확히 그 반례를 **의무화**한다.

- REQ-CSC-014 — 대상에 사용자 실 항목이 있으면 "건너뛰고 경고해야 한다(shall)". 건너뛰면 그 스킬의 접근 경로는 **없다**.
- REQ-CSC-011 — 링크·복사 양쪽 실패 시 "경고를 남기고 계속 진행해야 한다". 이때도 접근 경로는 없다.

AC 계층이 이 충돌을 그대로 물려받는다. AC-CSC-002 는 "배포된 **모든** 스킬"에 대해 양쪽 경로가 읽힌다고 단언하고(`acceptance.md:47`), AC-CSC-011(3) 은 같은 배포 실행에서 (iii) 스킬의 미러가 **사용자 파일 그대로**라고 단언한다(`acceptance.md:122`). 두 AC 를 동시에 만족하는 구현은 없다 — 서로 다른 fixture 라서 실행 시 충돌이 드러나지 않을 뿐, 요구사항 계층은 이미 모순이다.

이것은 이번 개정이 **새로 만든** 결함이다. iter-1 의 REQ 12개에는 REQ-CSC-014 가 없었고 D6 의 수정 지시가 그것을 추가하게 했는데, 상위 REQ-CSC-001 의 무조건성을 함께 손보지 않았다.

Required fix: REQ-CSC-001 에 예외 절을 명시한다 — "…제공해야 한다(shall). 단 REQ-CSC-014 가 규정하는 대상 선점 상태와 REQ-CSC-011 이 규정하는 미러 생성 실패는 예외이며, 그 경우 경고가 그 자리를 대신한다." 번호를 새로 쓰지 않는 절 수정이므로 예산 상한(16)에 걸리지 않는다. AC-CSC-002 의 "모든 스킬"에도 같은 단서를 단다.

### N2 — REQ-CSC-010(백업 금지)이 REQ-CSC-008(청소 글롭)과 만나 무백업 데이터 손실 경로를 만든다 (blocking)

`spec.md:160`(REQ-CSC-010) ↔ `spec.md:158`(REQ-CSC-008) ↔ `acceptance.md:135`(AC-CSC-012 2번) ↔ `acceptance.md:124`(AC-CSC-011 `[HARD]`) — Severity: **major** — Class: **blocking**

REQ-CSC-010 의 반전은 §A.6·§A.7 실측 위에서 옳다(둘 다 아래 "재현한 실측"에서 독립 확인). 그런데 반전이 **잃어버린 것**이 규정되지 않았다.

청소 대상은 `.agents/skills/moai*` 이고, `moai` 접두는 사용자도 쓸 수 있다 — 이 사실은 SPEC 자신이 AC-CSC-011 의 `[HARD]` 주석에 적어 두었고, 그것이 REQ-CSC-014(배포기는 덮어쓰지 않는다)의 존재 이유다. 그런데 같은 항목이 **청소 단계**에 오면 글롭이 매치해 제거되고, AC-CSC-012 2번 단언은 `.moai-backups/**/pre-clean/.agents/` 아래 파일 수가 **0** 이어야 한다고 **절대 형태로** 요구한다. 즉 SPEC 은 한쪽(배포)에서는 사용자 실 항목을 지키라고 하고, 다른 쪽(청소)에서는 지우면서 **백업조차 금지**한다.

`backupThenRemove` 의 현재 동작으로 확인하면 이 충돌은 구현 시점에 곧바로 드러난다(`internal/cli/update/deploy/deploy.go:390-398`): 실 디렉터리는 `templateManagedPaths` 가 공집합이므로 전부 unmanaged → **백업된다** → AC-CSC-012(2) 가 FAIL 한다. 구현자는 `.agents/` 를 백업에서 통째로 제외해 AC 를 통과시킬 것이고, 그 순간 사용자의 `moai-` 이름 실 항목은 **경고도 백업도 없이** 사라진다.

Required fix: 둘 중 하나를 SPEC 이 선택해 명시한다. (a) REQ-CSC-010 의 백업 금지를 "**배포기가 만든** 미러 항목"으로 한정하고, 그 밖의 항목은 기존 백업 규칙을 따른다 — 이 경우 AC-CSC-012(2) 의 단언을 "미러 산출물에 대해 0" 으로 좁힌다. (b) REQ-CSC-009 를 확장해 청소가 **링크가 아닌 실 항목은 `moai*` 이라도 제거하지 않는다**로 규정한다. 어느 쪽이든 §D 의 "파괴하지 않는다는 보장 두 가지" 서술도 함께 갱신해야 한다.

### N3 — §A.9·plan §F M6·plan §H 가 근거로 든 코드 불변식은 `moai-` 이고, 카탈로그는 그것을 이미 위반한다 (blocking)

`spec.md:117`("`moaiSkillPrefix` 필터") + `plan.md:99` + `plan.md:124` ↔ `internal/template/skills_manifest.go:15,42` — Severity: **major** — Class: **blocking**

SPEC 은 REQ-CSC-015 의 정당화로 "이 불변식은 코드에 이미 암묵적으로 있다 — `internal/template/skills_manifest.go` 의 `moaiSkillPrefix` 필터"라고 적는다. 실측:

```
const moaiSkillPrefix = "moai-"          # skills_manifest.go:15
if e.IsDir() && strings.HasPrefix(e.Name(), moaiSkillPrefix)   # :42

$ find internal/template/templates/.claude/skills -mindepth 1 -maxdepth 1 -type d -exec basename {} \; > names
$ wc -l < names            → 34
$ grep -cv '^moai'  names  → 0      # REQ-CSC-015 의 불변식(접두 `moai`)은 성립
$ grep -cv '^moai-' names  → 1      # 코드 상수의 불변식(접두 `moai-`)은 성립하지 않는다
$ grep -v '^moai-' names   → moai
```

배포 카탈로그에는 이름이 **정확히 `moai`** 인 스킬(통합 오케스트레이터 스킬)이 있고, `EmbeddedMoaiSkillNames()` 는 그것을 **뺀 33개**를 돌려준다. 즉 SPEC 이 "이미 있다"고 말한 불변식은 REQ-CSC-015 가 세우려는 것과 **다른 불변식**이며, 그 다른 불변식은 지금 이미 깨져 있다.

파급이 문서 정확성에 그치지 않는다.

1. `plan.md:99`(M6) 이 같은 상수를 가리키므로, 구현자가 그 seam 을 재사용해 미러 집합을 만들면 **`moai` 스킬 하나가 조용히 미러에서 빠진다** — REQ-CSC-006(미러 집합 == 배포 집합) 위반이고, AC-CSC-002 는 34개 전부를 요구하므로 그때 FAIL 한다. 발견은 되지만 원인 표지가 SPEC 안에 잘못 심겨 있다.
2. `plan.md:124`(§H 정리 판별 기준)은 `template.EmbeddedMoaiSkillNames()` 를 "현재 배포 카탈로그"의 정의로 제시한다. 그 집합에 `moai` 가 없으므로, 이 기준을 따르면 사용자 홈의 `~/.codex/skills/moai` 가 **삭제 후보로 분류된다** — 현역 스킬을 은퇴 스킬로 오분류하는 형태다. §H 가 삭제 실행을 승인 뒤로 미루므로 즉시 손실은 아니지만, 판별 기준 자체가 틀렸다.

Required fix: §A.9 에서 `moaiSkillPrefix` 인용을 정정한다 — 코드 상수는 `moai-` 이고 카탈로그의 `moai` 항목이 그 필터 밖이므로, REQ-CSC-015 의 접두 `moai` 불변식은 **코드에 없다**(그래서 더욱 필요하다). `plan.md` §F M6 에서도 같은 문장을 고치고, §H 판별 기준의 `EmbeddedMoaiSkillNames()` 를 `internal/template/templates/.claude/skills/*/` 실측 또는 `moai` 접두(하이픈 없음) 기준으로 바꾼다. AC-CSC-016 문구(접두 `moai`)는 그대로 두면 된다 — 그쪽은 옳다.

### N4 — 실제 `moai update` 실행 순서에서 청소는 미러를 **한 개도 지우지 못한다** (blocking)

`spec.md:158`(REQ-CSC-008) + `internal/cli/update_template_sync.go:297,323` + `internal/cli/update/deploy/deploy.go:63-68, 372-378` — Severity: **critical** — Class: **blocking**

세 사실이 겹친다. 전부 직접 측정했다.

1. **clean 이 deploy 보다 먼저다.** 스텝 배열에서 `deploy.CleanMoaiManagedPaths`(`update_template_sync.go:297`)가 `deployer.Deploy`(`:323`)보다 앞 항목이다.
2. **clean 내부에서 `.claude/skills/moai*` 가 먼저 지워진다.** `CleanMoaiManagedPaths` 는 `ManagedCleanTargets` 슬라이스를 **순서대로** 돈다(`deploy.go:112`). 그 4번째 항목이 `.claude/skills/moai*` 글롭이고(`deploy.go:63-68`), D5/M4 가 지시하는 `.agents/skills/moai*` 항목은 그 뒤에 붙는다. 그러면 `.agents/skills/<name>` 링크를 처리할 시점에 정본은 **같은 실행에서 이미 삭제**돼 모든 미러 링크가 dangling 이다.
3. **dangling 링크는 무음으로 건너뛰어진다.** `backupThenRemove` 의 첫 동작이 `os.Stat`(링크를 따라감)이고 `os.IsNotExist` 면 `return 0, nil` — 제거 없이 성공 반환이다(`deploy.go:372-378`). 직접 실행:

```
DANGLING os.Stat err: ... no such file or directory  IsNotExist= true
DANGLING os.Lstat err: <nil>
DANGLING glob matches: 1
```

`filepath.Glob` 은 dangling 링크를 **매치한다**. 청소는 대상을 찾아내고도 지우지 않는다.

귀결: 은퇴·개명된 스킬의 미러 링크가 사용자 프로젝트에 영구 잔존한다 — §A.4 가 `~/.codex/skills/` 오염으로 관측하고 REQ-CSC-008 이 막으려는 **바로 그 실패 형태**다. 규정된 대로 구현하면 목적이 실행 시점에 무효화된다.

경계 확인: 링크가 **살아 있을 때**는 §A.7 의 "링크 모드 — 안전" 서술이 옳다(`os.Stat` 성공 → `WalkDir` 이 정규 파일 0개 → `RemoveAll` 이 링크만 제거). 문제는 살아 있지 않은 순간이고, 실제 실행 순서가 그 순간을 보장한다.

**출처 고지 — 독립 발견이 아니라 독립 검증이다.** 이 결함에 내가 스스로 도달했는지 리드가 물었으므로 정확히 적는다. 감사 도중 런타임이 `plan-audit.md` 파일 변경 알림을 띄우면서 그 보고서의 **앞부분(N1 포함)이 내 컨텍스트에 들어왔고**, 나는 그것을 읽은 **뒤에** 위 세 측정을 수행했다. 따라서 내가 제공하는 것은 발견의 독립성이 아니라 **메커니즘의 독립 재현**이다 — 스텝 배열 순서(`update_template_sync.go:297` vs `:323`), `ManagedCleanTargets` 슬라이스 내 `.claude/skills/moai*` 의 위치(`deploy.go:63-68`, 4번째), `backupThenRemove` 의 `os.Stat`→`IsNotExist`→`return 0, nil` 조기 반환(`deploy.go:372-378`), 그리고 dangling 링크에 대한 `os.Stat`/`os.Lstat`/`filepath.Glob` 거동(위 프로브 출력)을 각각 내가 직접 확인했고, **어느 것도 반증되지 않았다**. 수렴을 독립 증거로 셈하지 말고, "두 번째 감사자가 같은 코드에서 같은 사실을 확인했다"까지만 셈할 것.

한편 이 결함과 **모순되는 측정은 얻지 못했다.** 반대 방향으로 확인한 것은 하나뿐이다 — 링크가 살아 있을 때는 §A.7 의 "링크 모드 — 안전" 서술이 옳다(위 경계 확인). 그 사실은 N4 를 약화시키지 않는다. N4 가 성립하려면 청소 시점에 링크가 dangling 이어야 하는데, 실행 순서가 정확히 그 상태를 만든다.

Required fix: REQ-CSC-008 에 절을 하나 더한다 — "청소는 `.agents/skills/moai*` 대상의 존재 여부를 **`os.Lstat` 기준으로 판정해야 하며(shall)**, 대상이 존재하지 않는 링크(dangling)도 제거해야 한다(shall)." 새 번호가 필요 없다. 청소 항목 순서를 `.claude/skills/moai*` 앞으로 두는 우회는 권하지 않는다 — 순서 의존은 나중에 조용히 깨진다.

### N5 — AC-CSC-008 / AC-CSC-009 는 **실 디렉터리** fixture 라 N4 를 검출할 수 없다 (blocking)

`acceptance.md:87-100` — Severity: **major** — Class: **blocking**

두 AC 모두 `.agents/skills/<name>/SKILL.md` 를 **실 파일로** 심는다. 이 SPEC 이 배포하는 산출물은 D1·REQ-CSC-003 에 따라 **심볼릭 링크**다. 실 디렉터리는 `os.Stat` 이 성공하므로 청소가 정상 동작하고, 두 AC 는 N4 가 살아 있는 상태에서도 **통과한다**. MUST 로 분류된 AC 두 개가 이 SPEC 최대 실패 형태의 정확히 반대편만 검사한다.

AC-CSC-008 의 `[HARD]` 주석은 "제거만 단언하는 테스트"를 경계하는데, 실제로 놓치는 축은 제거/생존이 아니라 **실 항목 / 링크 항목**이다.

부가로 AC-CSC-009 의 생존 단언은 반증하기 매우 어렵다 — `filepath.Glob(".agents/skills/moai*")` 는 `hns-*`·`harness-*`·`my-own` 을 원리적으로 매치하지 않는다. 판정력이 0 은 아니지만(AP-6 형태의 `.agents/skills/*` 는 잡는다) 얇다.

Required fix: AC-CSC-008 의 fixture 를 한 테스트 안에서 네 형태로 확장한다 — (i) 정본이 살아 있는 링크 `moai-live`, (ii) 정본이 이미 지워진 **dangling 링크** `moai-gone`, (iii) 복사 모드 산출물인 실 디렉터리 `moai-copied`, (iv) 사용자 소유 실 디렉터리 `hns-user-owned`. 단언: (i)(ii)(iii) 제거 + (iv) 내용 무변경 + 정본 `.claude/skills/moai-live/` 가 링크를 통해 지워지지 않았음. (ii) 가 N4 를 잡는 유일한 팔이다.

### N6 — AC-CSC-006 · 011(3) · 013(3) 이 전제하는 "Deploy 의 출력 버퍼"는 존재하지 않고, plan §F M2 의 재사용 전제도 사실이 아니다 (blocking)

`acceptance.md:74`(AC-006) · `:122`(AC-011 3번) · `:147`(AC-013 3번) + `plan.md:60`(M2 "출력 경로는 기존 printer 계층 재사용") ↔ `internal/template/deployer.go:60-100` — Severity: **major** — Class: **blocking**

실측: `deployer` 구조체 필드는 `fsys` / `renderer` / `forceUpdate` / `renderCache` 뿐이고, 서명은 `Deploy(ctx, projectRoot, m manifest.Manager, tmplCtx *TemplateContext) error` 다. **`io.Writer` 가 없다.** `internal/template/` 의 비-테스트 파일 전체에서 `io.Writer` 매치 0건이다. 출력 표면(`tui.ProgressLine` 등)은 호출자인 `internal/cli/` 계층에 있고 배포기까지 내려오지 않는다.

따라서 세 AC 는 **run-phase 가 새로 만들어야 하는 seam** 을 이미 있는 것처럼 전제하고, `plan.md` M2 는 "기존 printer 계층 재사용"이라고 적어 그 작업이 필요 없는 것처럼 읽힌다. REQ-CSC-005(모드 보고)는 이 seam 없이는 관측 가능한 형태로 성립하지 않는다. §D.7 은 AC-CSC-010 의 토글 seam 은 설계 제약으로 명시했는데, 같은 성격의 출력 seam 은 어디에도 적혀 있지 않다.

Required fix: (a) `plan.md` M2 에서 "기존 printer 계층 재사용"을 삭제하고 **출력 seam 신설**을 M2 의 산출물로 명시한다(예: 배포기에 `io.Writer` 를 주입하거나 모드 결과를 반환값으로 올려 호출자가 인쇄). (b) §D.7 에 AC-CSC-010 토글과 나란히 "출력 캡처 seam 이 없으면 AC-006·011(3)·013(3) 이 함께 막힌다"를 남긴다.

### N7 — REQ-CSC-016 은 `.agents/` **전체**를 무시하게 한다 — 좁은 글롭 원칙과 어긋난다 (optional)

`spec.md:166` + `acceptance.md:161` — Severity: **major** — Class: **optional**

이 SPEC 은 청소 쪽에서는 "`.agents/` 전체를 잡는 형태는 금지"(AP-6, §B.D5)라는 좁은-범위 원칙을 세운다. 그런데 `.gitignore` 쪽에서는 같은 원칙을 적용하지 않고 `.agents/` 전체를 무시하게 한다. 생성물은 `.agents/skills/moai*` 뿐이므로, 이 형태는 사용자가 스스로 만든 `.agents/skills/hns-*` 와 후속 마일스톤(M2 AGENTS.md 정본화 등)이 `.agents/` 아래 둘 수 있는 **소스 파일까지** 조용히 추적 대상에서 뺀다.

기각 여지가 있어 optional 로 분류한다 — §D 는 사용자 `.agents/` 항목을 범위 밖으로 두고 "파괴하지 않는다"만 보장하므로, 추적 제외가 그 보장을 깨지는 않는다. 다만 판단이 SPEC 에 적혀 있지 않아 run-phase 가 기본값으로 넓은 형태를 고르게 된다.

Required fix(택일): (a) REQ-CSC-016 을 `.agents/skills/moai*` 범위로 좁히거나, (b) 전체 무시를 유지하되 그 선택의 근거와 후속 마일스톤에 대한 영향을 §B 에 한 줄로 남긴다.

### N8 — AC-CSC-010 의 불변식은 REQ-CSC-007 이 말하는 "본 변경 전과 동일"을 증명하지 않는다 (optional)

`spec.md:157`(REQ-CSC-007) ↔ `acceptance.md:104-112`(AC-CSC-010) — Severity: **minor** — Class: **optional**

AC-CSC-010 이 증명하는 것은 "**미러 기능을 켜고 끈 것 사이**의 `.claude/skills/` 산출물이 같다"이다. REQ-CSC-007 이 요구하는 것은 "**본 변경 전**과 같다"이다. 미러-비활성 seam 자체가 리팩터링 과정에서 이전 동작과 달라지면, AC-CSC-010 은 통과하면서 REQ-CSC-007 은 깨진다.

이는 iter-1 D2 의 수정 지시(실행 가능한 불변식 + 일회성 수동 대조 분리)를 그대로 따른 결과이므로 결함이라기보다 **명시되지 않은 잔여**다. §D.4 의 커밋 SHA 대조가 그 구멍을 1회 덮지만 AC 가 아니다.

Required fix: REQ-CSC-007 문구를 AC-CSC-010 이 실제로 증명하는 범위에 맞춰 좁히거나("미러 기능의 활성 여부가 `.claude/skills/` 산출물을 변화시켜서는 안 된다"), 지금 문구를 유지한다면 §D.4 의 1회 대조가 REQ-CSC-007 의 **유일한** 착지 시점 근거임을 §D.2 에 명시한다.

### N9 — §G 가 §F 앞에 온다 (optional)

`spec.md:198`(§G) → `spec.md:204`(§F) — Severity: **minor** — Class: **optional**

절 순서가 알파벳·논리 양쪽으로 어긋난다. 내용에는 영향이 없다. Required fix: §F 교차 참조를 §G 앞으로 옮긴다.

---

## 기각 판단 — D12 (REQ 원자성 분리)

**기각은 정당하다.** 근거를 재확인했다.

- `spec-workflow.md:152` 은 Tier M 상한을 요구사항 16 / 판정 16 으로 두고, **초과는 tier 상향 또는 SPEC 분할 신호**라고 명시한다. 현재 16/16 이므로 REQ 두 개를 넷으로 쪼개면 18 이 되어 상한을 넘는다. §G 의 예산 논거는 사실에 부합한다.
- D12 는 iter-1 에서 **optional/minor** 로 분류됐고, M6(finding-consumption discipline)상 optional 항목은 오케스트레이터 재량이다. optional 을 근거로 FAIL 을 만들지 않는다는 규율에도 부합한다.
- 기능적 손실이 없다는 §G 의 주장도 확인된다 — REQ-CSC-006 은 AC-CSC-003 과 AC-CSC-014 로, REQ-CSC-011 은 AC-CSC-013 의 세 단언으로 절이 갈라져 실행 시점에 "어느 절이 실패했는가"가 구분된다.

다만 한 가지를 덧붙인다. §G 의 예산 논거는 **이번 blocking 수정에는 적용되지 않는다.** N1·N2·N4 의 Required fix 는 전부 기존 REQ 항목의 **절 추가·수정**이라 새 번호를 쓰지 않는다. "상한에 닿아서 못 고친다"가 다음 개정의 사유가 되어서는 안 된다.

한편 16/16 정확히 상한이라는 사실 자체는 신호다 — 위 규칙이 말하는 "tier 상향 또는 분할"의 문턱에 서 있고, 아직 규정되지 않은 항목(N4 의 제거 의미론, N6 의 출력 seam)이 남아 있다. 아래 Recommendation 의 선택지 2가 여기서 나온다.

---

## 재현한 실측 (SPEC 의 근거 검증)

| SPEC 의 주장 | 내 측정 | 판정 |
|---|---|---|
| §A.1 템플릿 배포 대상 34, 트리 링크 0 | `find … -type d \| wc -l` = **34**, `find internal/template/templates -type l \| wc -l` = **0** | 일치 |
| §A.3 core 21 / non-core 13, 전부 `optional-pack:*`, `harness_generated` 스킬 0 | `grep -c 'tier: optional-pack'` = **13**, `catalog.yaml:262` = `skills: []` | 일치(iter-1 D9 정정 반영 확인) |
| §A.4 `ManagedCleanTargets` 7개 전부 `.claude/` 하위 | `deploy.go:51-82` 항목 **7개**, `.agents/` 없음 | 일치 |
| §A.6 `Track`→`HashFile` 이 디렉터리 링크에서 실패 | `manifest.go:135` 가 `HashFile` 호출 후 error 전파, 직접 실행 → `A6 open err: <nil>` / `A6 copy err: read …: is a directory` | 일치 |
| §A.7 링크 모드 안전 / 복사 모드 전량 백업 | `backupThenRemove`(`deploy.go:371-399`) + `templateManagedPaths`(`:401-416`, 주석에 "prefix 를 template 이 안 가지면 전부 unmanaged" 명시) | 일치 |
| §A.8 템플릿 `.gitignore` 에 `.agents/` 없음 | `grep -n 'agents' internal/template/templates/.gitignore` → 매치 0 | 일치 |
| §A.9 비-`moai` 이름 0개 | `grep -cv '^moai'` = **0** | 일치 |
| §A.9 "불변식은 코드에 `moaiSkillPrefix` 로 이미 있다" | 상수는 `"moai-"`, 카탈로그에 `moai` 존재 → 코드 필터는 33개만 반환 | **불일치 → N3** |

계수는 전부 `find … -exec basename` 로 했다 — `ls | wc -l` 은 이 셸에서 실제로 long-format 으로 확장돼(감사 도중 1회 발생) SPEC 자신이 AP-7 로 규율한 오차를 재현한다.

---

## Recommendation

**FAIL.** Tier M 임계 0.80 에 0.7625(조화 0.747)로 미달했고, blocking 6건 중 N4 는 critical 이다. 이번 개정은 iter-1 blocking 8건을 **전부 닫았고** 판정 계층의 세 구멍(D1·D2·D7)은 정확히 메워졌다 — 그럼에도 총점이 내려간 것은 닫는 과정에서 요구사항 계층에 새 모순(N1·N2)이 생기고, 양쪽 iteration 이 보지 못했던 실행 순서 결함(N4·N5)이 드러났기 때문이다.

**Tier M 반복 상한 2 에 도달했으므로 iteration 3 은 자동으로 열리지 않는다.** 오케스트레이터는 사용자에게 세 선택지를 제시해야 한다.

1. **PASS-with-debt** — 권하지 않는다. N4 가 열려 있으면 이 SPEC 의 핵심 요구사항(REQ-CSC-008)이 착지 후에도 목적을 달성하지 못하고, N5 때문에 그 사실이 테스트로도 드러나지 않는다. 무음 실패가 그대로 배포된다.
2. **범위 축소(권장)** — 16/16 상한 도달 + 미규정 항목 잔존은 이 SPEC 이 한 덩어리로 크다는 신호다. 분할선이 이미 SPEC 안에 그려져 있다: **(A) 미러 생성 + 폴백**(REQ-001·003·004·005·006·011·012·013·014·015, plan M1·M2·M6)과 **(B) 청소·기록·백업·`.gitignore`**(REQ-002·007·008·009·010·016, plan M3·M4·M5). N4·N2 는 전부 (B) 안에서 닫히고, (A) 는 지금 상태로도 임계에 근접한다. 두 SPEC 이면 각각 예산 여유가 생겨 N1·N2·N4 의 절 추가도 압박 없이 들어간다.
3. **상한 연장(iteration 3)** — 선택 가능. 이 경우 수정 순서는 아래.

상한을 연장한다면 수정 순서(전부 기존 번호 안에서 처리된다):

1. **N4** — REQ-CSC-008 에 `os.Lstat` 판정 + dangling 제거 절 추가. 이 SPEC 의 목적이 걸린 항목이다.
2. **N5** — AC-CSC-008 fixture 를 4형태로 확장. N4 를 잡는 유일한 팔.
3. **N1** — REQ-CSC-001 에 예외 절(REQ-011·014) 추가 + AC-CSC-002 문구 단서.
4. **N2** — REQ-CSC-010 의 백업 금지를 "배포기가 만든 미러"로 한정하거나 REQ-CSC-009 를 실 항목까지 확장. AC-CSC-012(2) 를 그에 맞춰 좁힌다.
5. **N6** — plan M2 에서 "기존 printer 계층 재사용" 삭제, 출력 seam 신설을 산출물로 명시, §D.7 에 제약 기록.
6. **N3** — §A.9 · plan M6 · plan §H 의 `moaiSkillPrefix` / `EmbeddedMoaiSkillNames()` 인용 정정.
7. N7~N9 는 optional. 오케스트레이터 재량이며 이것들만으로 재감사를 돌릴 이유는 없다.
