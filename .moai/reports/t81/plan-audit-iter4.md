# SPEC Review Report: SPEC-CODEX-SKILLS-CANONICAL-001

Iteration: 4 (리드가 부여한 6라운드 예외 상한 중 4회차 — Tier M 기본 상한 2회를 넘어선 명시적 예외)
Verdict: **FAIL** — **STOP**
Overall Score: **0.7625** (Tier M 임계 0.80)

Reasoning context ignored per M1 Context Isolation. 이전 감사 보고서 4건은 **선행 결함 목록(닫힘 여부 판정 대상)** 으로만 읽었고, 그 안의 수치·판단은 어느 것도 근거로 인용하지 않았다. 모든 load-bearing 수치를 이 감사에서 직접 재측정했다.

## 고정 상태 확인 (감사 시작 시점)

```
$ git log --oneline -1
c1d2f415b docs(spec): revise SPEC-CODEX-SKILLS-CANONICAL-001 to v0.5.0 (card t81)
$ git status --short
(빈 출력)
$ git branch --show-current
WT-skills-canonical
$ pwd
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t81
```

감사 종료 시점 `git status --short` 도 빈 출력(프로브 디렉터리 삭제 확인). 감사 중 아티팩트 개정 없음.

---

## Must-Pass Results

- **[PASS] MP-1 REQ 번호 일관성** — `grep -o 'REQ-CSC-[0-9]\{3\}' spec.md | sort -u | wc -l` → **16**, `grep -c '^- \*\*REQ-CSC-' spec.md` → **16**. REQ-CSC-001..016 연속, 결번·중복 0, zero-padding 일관(spec.md:232-249).
- **[PASS] MP-2 GEARS 형식 준수** — **요구사항 계층(`REQ-XXX`)에 한해 판정**했다(M3 § Scope). 16개 전부가 다섯 GEARS 패턴 중 하나에 매치한다: Ubiquitous(001·015), Unwanted `shall not`(002·007), Event-driven `When`(003·005·008·011·012·013·014), Where(004), While(009), 복합 shall+shall not(006·010·016). `IF/THEN` 형태 0건. 검증 계층(`AC-XXX`)의 Given-When-Then 은 이 기준으로 감점하지 않았고 Group 4 에서 평가했다.
- **[PASS] MP-3 YAML frontmatter 유효성** — spec.md:1-15 에서 canonical 12필드 전부 확인: `id`(패턴 일치) · `title`(따옴표) · `version: "0.5.0"`(따옴표 semver) · `status: draft`(enum) · `created`/`updated: 2026-08-22`(ISO) · `author` · `priority: P2` · `phase: "v3.2.0 target"` · `module: internal/template` · `lifecycle: spec-anchored` · `tags`(콤마 구분). 거부 별칭(`created_at`/`updated_at`/`labels`/`spec_id`) 0건. `tier: M` 추가 필드 존재(Tier 해석에 사용).
- **[N/A] MP-4 §22 언어 중립성** — 본 SPEC 은 Go 단일 프로젝트(`internal/template`, `internal/cli/update/deploy`)에 스코프된다. 다국어 툴링을 다루지 않으므로 N/A, 자동 통과. (spec §E 의 "템플릿 중립성" 절은 별개 축이며 그 자체로 위반 없음.)
- **[PASS] MP-5 D7 교차 SPEC 조정** — `grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+'` 결과 3파일 전부에서 **자기 자신(SPEC-CODEX-SKILLS-CANONICAL-001)** 한 건뿐. 외부 SPEC 참조 0건 → retired/superseded/archived 참조 없음, BLOCKING 없음.
- **[PASS] MP-6 D8 크로스 플랫폼 규율** — 리터럴 `syscall` 은 spec.md 1건 · acceptance.md 1건 매치하나, **둘 다 Go `syscall` 패키지가 아니라 `os.Lstat`/`os.Stat` 을 가리키는 산문 용법**이다("판정 syscall 을 `os.Lstat` 으로 고정"). `os` 패키지는 이식 가능하며 per-OS 분기가 필요 없다. 그리고 SPEC 은 명시적 크로스 플랫폼 절을 갖는다 — §E "OS 중립성: darwin / linux / windows 에서 모두 REQ-CSC-001 이 성립", §D.4 "`GOOS=windows go vet ./internal/...`", AC-CSC-007 의 `filepath.ToSlash` [HARD] 절(Windows 백슬래시 대비). D8 이 잡으려는 실패 형태(빌드 태그 누락)는 성립하지 않는다.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' plan.md acceptance.md spec.md` → 매치 0건. `research.md` 는 부재(Tier M).

**7개 must-pass 전부 통과.** FAIL 은 must-pass 위반이 아니라 **집계 점수(0.7625) 가 Tier M 임계(0.80) 미달**이고 blocking 결함 4건이 열려 있기 때문이다.

---

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.70 | 0.50~0.75 사이 (0.75 밴드 미달) | REQ-CSC-008 의 "존재 여부를 `os.Lstat` 기준으로 판정"(spec.md:240)과 §B.D6 의 "`backupThenRemove` 의 판정을 `os.Stat` 에서 `os.Lstat` 로 바꾸는 것"(spec.md:221)은 **범위가 다르고**, 후자의 문자 그대로의 구현은 제품을 깨뜨린다(D1). 나머지 15개 요구사항은 1.0 급으로 정밀하다 |
| Completeness | 0.80 | 0.75 밴드 상단 | 필수 절 전부 존재(HISTORY spec.md:19 / §A / §B / §C / §D 범위 밖 4개 H3 + 불릿 / §E / §F / §G), frontmatter 완전. 감점 사유는 **링크 항목용 분기 설계가 spec·plan 어디에도 없다**는 것 — §A.6 이 `manifest.Track` 에서 식별한 EISDIR 위험이 `copyRegularFile` 에서 재현되는데 그 절이 없다 |
| Testability | 0.65 | 0.50 밴드 상단 | AC-CSC-008 의 1·2번이 §B.D6 이 처방한 구현에서 **통과 불가**(D1), AC-CSC-012 의 2번이 자기 Given 이 만드는 fixture 에서 **공허하게 참**(D4), AC-CSC-008 6번은 반증 경로가 거의 없다(D5). 나머지 AC 는 이진 판정 가능하고 weasel word 0건 |
| Traceability | 0.90 | 0.75~1.0 사이 | REQ 16 → AC 매핑 §D.3 이 16개 전부 커버, AC 16 → REQ 매트릭스(acceptance.md:11-27)가 16개 전부 커버, 양방향 확인. 매트릭스와 §D.3 이 서로 모순 없음. 감점은 plan §F M4 닫힘 조건의 개수 드리프트(D3) 하나뿐 |

**산술평균 = (0.70 + 0.80 + 0.65 + 0.90) / 4 = 0.7625**

---

## 선행 결함 닫힘 판정 (iter-3 목록 대비)

| ID | 내용 | 판정 | 근거 |
|---|---|---|---|
| iter-3 D2 | AC-CSC-008 마지막 단언이 올바른 구현에서 반드시 실패 | **CLOSED** (단, 대체물이 새 결함 D2 를 들여옴) | 해당 문장 제거 확인(acceptance.md:118-120 이 삭제 사실을 기록). `moai-linkprobe` 5·6번 단언으로 대체(acceptance.md:110-111) |
| iter-3 D3 | AC-CSC-012 의 2번·3번이 동시에 참일 수 없음 | **CLOSED** | acceptance.md:167 이 "템플릿이 같은 이름의 스킬을 가진 각 이름 `<n>` 에 대해"로 측정 폭을 이름 단위로 좁힘. `moai-retired` 는 템플릿 비보유이므로 2번의 대상 집합에서 제외 → 3번(보존)과 동시 성립 가능. acceptance.md:171 에 [HARD] 로 재발 방지 절 추가 |
| iter-3 D3-b | plan M3 이 AP-16 금지 형태를 처방 | **CLOSED** | plan.md:81 이 REQ-CSC-010 판별자 문구("이번 실행이 다시 만들 미러인지")로 교체, plan.md:83 에 [HARD] AP-16 금지 절 명시, plan.md:85 닫힘 조건이 3팔로 정정 |
| iter-3 D1 + D7 | 출력 seam 근거 인용 4곳이 `§B.D6` 를 오지목 | **CLOSED** | `grep -n '§B.D3'` → 출력 seam 인용 4곳(spec.md:148·237, plan.md:29·68, acceptance.md:79) 전부 `§B.D3 + §A.9b` 로 재조준. 잔존 `§B.D6` 인용 5건은 전부 dangling·순서·폭발 반경 맥락으로 **정당**. `### A.9b`(spec.md:138) 앵커 실재 |
| iter-3 D4 (optional) | `acceptance.md §D.4` 명시 | **CLOSED** | spec.md:238 REQ-CSC-007 이 "`acceptance.md` §D.4 의 1회 대조"로 파일명까지 명시 |
| iter-3 D5 (optional) | §A.11 이 코드 주석에 없는 문장을 인용 | **CLOSED** | 코드 주석 실물 확인(deploy.go:401-405 `templateManagedPaths` 주석, 426-429 `templateCarries`) — 사유 문장 없음. spec.md:180 이 인용 형태를 버리고 "주석은 동작만 적고"로 정정. 덧붙여 `templateCarries` 가 `!info.IsDir()` 분기 전용이라는 [HARD] 절 추가(spec.md:178) — **deploy.go:380 에서 검증됨** |
| iter-3 D6 (optional) | 단언 개수 표기 3곳 불일치 | **OPEN** | plan.md:101 이 여전히 `AC-CSC-008(4형태 — dangling 팔 포함)`. acceptance 는 5형태·6단언. 아래 D3 참조 |

**닫힘 6 / 열림 1.** 이월 blocking 은 0건이며, 열린 1건은 optional 이었던 항목이 이번 개정으로 **오히려 악화**된 형태다(4형태 → 5형태로 바뀌면서 plan 만 뒤처짐).

---

## 리드가 물은 판정 — 재-기술(re-basing) 거절의 정당성

**저자의 거절은 근거가 옳고, 대안 (b)는 실제로 이 축을 판정할 수 없다.** 그러나 채택한 대체물은 같은 모호성을 완전히 제거하지 못했다.

**① 거절 사유는 참인가 — 참이다(실측).**

```
$ grep -n 'CleanMoaiManagedPaths\|\.Deploy(' internal/cli/update_template_sync.go
297:  return deploy.CleanMoaiManagedPaths(projectRoot, out, tmplFS)
323:  if deployErr := deployer.Deploy(ctx, projectRoot, mgr, tmplCtx); deployErr != nil {
```

clean(:297) → deploy(:323) 순서가 코드에 있고, `Deploy` 는 임베드 FS 에서 `.claude/skills/*` 를 재생성한다. 따라서 clean→deploy 를 모두 돌린 뒤의 최종 상태는 **청소가 링크를 따라가 정본을 파괴했든 아니든 동일**하다 — 관측하려는 손상이 관측 직전에 치유된다. 저자의 문장 그대로다.

**② 대체물이 링크 추종을 관측 가능하게 만드는가 — 부분적으로만.**

성공한 부분: `hns-linkprobe-src` 는 청소 대상 어느 것과도 매치하지 않는다(실측 — `ManagedCleanTargets` 7항목 전부 `.claude/` 하위이고 스킬 관련은 `.claude/skills/moai*` 글롭 하나뿐, deploy.go:64-68). 따라서 "정본이 자기 대상으로 지워졌다"는 교란 축이 실제로 분리된다.

실패한 부분 둘:
- 판정력이 거의 없다 — `os.RemoveAll` 은 심볼릭 링크를 **따라가지 않는다**(실측: 링크 제거 후 대상 `"keepme"` 그대로 읽힘). 6번 단언은 `filepath.EvalSymlinks` 로 해석한 뒤 지우는 이례적 구현에서만 반증된다. AC-CSC-009 는 자기 얇음을 본문에 적었는데(acceptance.md:132) 6번은 적지 않았다.
- **모호성이 자리를 옮겼다** — 탐침이 파일 링크라서, 실제 미러(디렉터리 링크)와 **다른 코드 분기를 탄다**(아래 D2). acceptance.md:113 은 *이름*이 비현실적이라는 점만 사유로 적었고 *형태*가 다르다는 점은 적지 않았다.

---

## Defects Found

### D1 — `os.Lstat` 전환의 폭발 반경이 과소 기술됐고, 처방된 형태는 `moai update` 를 실패시킨다

`spec.md:221` (§B.D6) · `plan.md:95,98` (M4) · `acceptance.md:100-104` (AC-CSC-008)
**Severity: critical — Class: blocking**

§B.D6 은 "`backupThenRemove` 의 판정을 `os.Stat` 에서 `os.Lstat` 로 바꾸는 것"을 처방하고, 폭발 반경을 **"관리 대상 뿌리 아래에 있으면서 대상이 사라진 심볼릭 링크 전부"** 로 한정해 기록했다. 이 한정은 사실과 다르다. 살아 있는 링크도 함께 바뀌며, 그쪽 결과는 제거가 아니라 **하드 실패**다.

측정한 코드 경로:

```
internal/cli/update/deploy/deploy.go
 372:  info, err := os.Stat(diskPath)          ← 처방된 교체 지점
 380:  if !info.IsDir() {                      ← Lstat 로 바꾸면 모든 링크가 여기로
 381:      if templateCarries(tmplFS, relTarget) { ... }   ← .agents/ 는 템플릿 비보유 → false
 384:      if err := copyRegularFile(...); err != nil {
 385:          return 0, fmt.Errorf("back up %s: %w", relTarget, err)
 465:  data, err := os.ReadFile(src)           ← copyRegularFile 본체, 링크를 따라감
 128:  n, rmErr := backupThenRemove(match, rel, backupBase, tmplFS)
 129:  if rmErr != nil { ...
 131:      return fmt.Errorf("remove %s: %w", match, rmErr)   ← clean 단계 전체 중단
```

직접 실행한 재현(darwin/arm64, go 1.26.x):

```
live  Stat err=<nil> IsDir=true | Lstat err=<nil> IsDir=false
ReadFile(dir-symlink) err=read .../.agents/skills/moai-live: is a directory
gone  Stat IsNotExist=true | Lstat err=<nil>
```

귀결 — `os.Stat` → `os.Lstat` 을 372행에서 그대로 치환하면:

1. 살아 있는 디렉터리 링크(`moai-live`, 그리고 **기존 7개 뿌리 아래의 모든 살아 있는 디렉터리 링크**)가 `!info.IsDir()` 파일 분기로 들어가고, `templateCarries` 가 false 이므로 `copyRegularFile` → `os.ReadFile` → **EISDIR** → 에러 → deploy.go:131 에서 clean 단계 전체가 중단 → `moai update` 실패.
2. dangling 링크(`moai-gone`)도 같은 분기로 들어가 `os.ReadFile` 이 ENOENT → 같은 중단. **즉 이 SPEC 이 닫으려는 결함 자체가 "제거"가 아니라 "실패"로 바뀐다.**
3. 따라서 AC-CSC-008 의 **1번·2번 단언이 처방된 구현에서 통과 불가**하다. iter-3 D2 가 지적한 것과 정확히 같은 형태의 결함이, 판정 계층에서 요구사항·설계 계층으로 옮겨간 채 재발했다.

더 무거운 것은 이것이 **§A.6 이 이미 식별한 위험의 재현**이라는 점이다. §A.6 은 `manifest.Track → HashFile → io.Copy` 가 디렉터리 링크에서 EISDIR 로 실패한다는 사실을 실측해 설계 방향을 뒤집는 근거로 삼았다. 동일한 위험이 `copyRegularFile → os.ReadFile` 에도 있는데 SPEC 은 그 자리를 보지 않았다.

REQ-CSC-008 본문("존재 여부를 `os.Lstat` 기준으로 판정")은 범위가 더 좁아 살아남을 여지가 있지만, 그 좁은 읽기를 택하면 **심볼릭 링크 전용 분기를 새로 써야 하고 그 분기는 spec §B·§C 어디에도, plan M3·M4 어디에도 규정돼 있지 않다.** 지금 상태는 두 읽기 중 하나는 제품을 깨뜨리고 다른 하나는 미규정이다.

**Required fix**: (a) §B.D6 의 폭발 반경 서술을 "살아 있는 링크의 분기 전환 + dangling 링크 제거" 양쪽으로 정정하고, EISDIR/ENOENT 하드 실패 경로를 명시할 것. (b) REQ-CSC-008 또는 신설 절에 **심볼릭 링크 대상의 처리(백업 여부 + 제거 방식)** 를 규정할 것 — `os.Lstat` 이 `ModeSymlink` 를 보고하면 `copyRegularFile` 를 거치지 않고 `os.Remove` 한다는 취지. (c) plan M4 의 (a) 항목을 그 분기 작성 작업으로 다시 쓸 것. (d) AC-CSC-008 에 "clean 이 오류 없이 완료된다"는 단언을 추가해 이 실패를 판정 범위에 넣을 것.

### D2 — 새 탐침 `moai-linkprobe` 가 실제 미러와 다른 코드 분기를 탄다 (AP-14 형태의 재현)

`acceptance.md:97,110-113` (AC-CSC-008 fixture 5번 항목)
**Severity: major — Class: blocking**

fixture 표는 `moai-linkprobe` 를 `.claude/skills/hns-linkprobe-src/**SKILL.md**` 를 가리키는 링크로 규정한다 — **파일 링크**다. 반면 실제 미러(REQ-CSC-003)는 `.agents/skills/<name>` → `../../.claude/skills/<name>` 의 **디렉터리 링크**다. 측정하면 두 형태는 `backupThenRemove` 에서 갈린다:

```
linkprobe(파일 링크)  Stat.IsDir=false  Lstat.IsDir=false   → 항상 파일 분기
moai-live(디렉터리 링크) Stat.IsDir=true  Lstat.IsDir=false  → 분기가 뒤바뀜
ReadFile(file-symlink) data="keepme" err=<nil>              → 성공
ReadFile(dir-symlink)  err=is a directory                   → 실패
WalkDir(dir-symlink) entries=1 (type=L---------)            → 정규 파일 0개
```

즉 `moai-linkprobe` 는 D1 의 실패 모드에서 **살아남는 유일한 fixture 항목**이다. 5·6번 단언은 통과하는데 1·2번은 실패한다. 그리고 AC-CSC-008 자신이 `[HARD]` 로 세운 AP-14("fixture 를 실 디렉터리로만 심는다 — 실제 산출물은 링크이므로 결함이 통과한다", plan.md:137)는 **fixture 형태가 제품 형태와 달라선 안 된다**는 규율인데, 그 규율을 닫으려고 추가한 항목이 같은 규율을 어겼다.

acceptance.md:113 의 [HARD] 절은 *이름*이 의도적으로 비현실적이라는 사유만 적고, 형태(파일 vs 디렉터리) 차이는 언급하지 않는다. 읽는 사람은 이 차이를 알 수 없다.

부수적으로: 파일 링크에 대해 `copyRegularFile` 은 `os.ReadFile` 로 **대상의 바이트를 백업 트리에 복사한다**(실측 `data="keepme"`). §A.7 이 "복사는 비정규 항목을 건너뛴다"고 적은 것은 `WalkDir` 경로(디렉터리 링크)에 한해 참이고, 파일 링크 경로에서는 백업이 실제로 링크를 따라간다.

**Required fix**: 탐침을 **디렉터리 링크**로 바꿀 것 — `.agents/skills/moai-linkprobe` → `../../.claude/skills/hns-linkprobe-src`(디렉터리), 6번 단언은 `.claude/skills/hns-linkprobe-src/SKILL.md` 가 그대로 읽힘. 이러면 제품 형태와 같은 분기를 타면서 청소 글롭 밖 대상이라는 성질은 유지된다. 그리고 형태를 제품과 일치시켰다는 사실을 [HARD] 절에 명시할 것.

### D3 — plan M4 닫힘 조건의 fixture 개수가 개정을 따라오지 못했다

`plan.md:101` — `닫힘 조건: AC-CSC-007(양팔 — 글롭 + 순서), AC-CSC-008(4형태 — dangling 팔 포함), AC-CSC-009`
**Severity: major — Class: blocking**

acceptance 는 **5형태 / 6단언**(매트릭스 acceptance.md:18 "미러 5형태 단일 테스트", 본문 acceptance.md:91 "아래 다섯 항목을 **모두** 심어 두고", acceptance.md:105 "다음 여섯 단언"). plan 만 iter-4 판본의 `4형태` 를 유지한다.

이것은 표기 오류에 그치지 않는다 — M4 의 **닫힘 조건**이므로, plan 만 읽는 구현자는 fixture 를 4개만 심고 `moai-linkprobe` 팔을 통째로 빠뜨린 채 마일스톤을 닫았다고 판단한다. iter-5 HISTORY(spec.md:21)는 "D6 개수 표기 3건"을 반영했다고 적었는데, 반영되지 않은 한 곳이 하필 닫힘 조건이다.

같은 계열의 잔여 드리프트: `plan.md:19` R5 완화 항목이 `AC-CSC-008(양팔)` — iter-2 시절 어휘로, 지금은 6단언이다.

**Required fix**: plan.md:101 을 `AC-CSC-008(5형태 6단언 — dangling 팔 + 링크 추종 탐침 포함)` 로 정정. plan.md:19 의 `(양팔)` 도 함께 갱신.

### D4 — AC-CSC-012 의 2번 단언이 자기 Given 이 만드는 fixture 에서 공허하게 참이다

`acceptance.md:163,167-168`
**Severity: major — Class: blocking**

2번 단언은 자기 존재 이유를 본문에 적었다 — *"링크 모드뿐 아니라 **복사 모드에서도** 0이어야 한다 — 이 단언의 존재 이유가 복사 모드다"*(acceptance.md:168). 그런데 Given 절(acceptance.md:163)이 구성하는 fixture 는 **AC-CSC-002 의 배포**, 즉 링크 모드 1회뿐이다. 복사 모드를 발동시키는 절이 없다.

링크 모드에서 이 단언이 어떻게 되는지 측정했다: 디렉터리 링크에 대해 `backupThenRemove` 는 `os.Stat` → `IsDir()==true` → `templateManagedPaths` → `backupUnmanagedTree` 로 가고, `filepath.WalkDir` 은 링크 루트를 `Lstat` 하므로 항목 1개(type `L---------`)만 걷고 정규 파일을 **0개** 만난다. 즉 백업 파일 수는 구현이 무엇을 하든 0 이고, 2번은 **항상 참**이다.

AC-CSC-005 는 "심볼릭 링크 생성이 실패하도록 주입된 배포기"라는 seam 이 존재함을 이미 규정한다. AC-CSC-012 의 Given 은 그 seam 을 쓰지 않는다.

**Required fix**: AC-CSC-012 의 Given 에 **두 배포**를 명시할 것 — (링크 모드 1회) + (AC-CSC-005 의 주입으로 복사 모드 1회), 2번 단언을 두 결과 모두에 대해 요구. 그렇지 않으면 R8("복사 모드에서 매번 전량 백업", plan.md:22, 등급 **높음**)이 판정에서 빠진다.

### D5 — AC-CSC-008 6번 단언의 얇음이 선언되지 않았다

`acceptance.md:111`
**Severity: minor — Class: optional**

측정: `os.RemoveAll` 은 심볼릭 링크를 따라가지 않는다(링크 제거 후 대상 `"keepme"` 온전). 따라서 6번은 `filepath.EvalSymlinks` 로 해석한 뒤 지우는 이례적 구현에서만 반증된다 — 실질적으로 거의 반증 불가능한 가드다.

가드로서의 가치는 있다(D1 의 수정이 링크 분기를 새로 쓰게 되면 그때 실수 여지가 생긴다). 문제는 **선언되지 않았다**는 것이다. AC-CSC-009 는 같은 성질을 본문에 적었다 — *"올바르게 구현된 글롭에 대해 이 AC 는 반증되기 어렵다 … 얇다는 것을 알고 두는 것과 모르고 두는 것은 다르다"*(acceptance.md:132). 같은 규율을 6번에도 적용해야 일관된다.

**Required fix**: 6번 단언 뒤에 한 문장 — 반증 조건이 "링크를 해석해 지우는 구현"에 한정되며 그 좁음을 알고 둔다는 취지.

### D6 — REQ-CSC-010 이 "다시 만들지 않을 **링크**" 항목의 백업 정책을 규정하지 않는다

`spec.md:243`
**Severity: minor — Class: optional**

REQ-CSC-010 의 세 절은 (1) manifest 기록 금지, (2) 이번 실행이 다시 만들 미러는 백업 금지, (3) *"그 판별에 걸리지 않는 `.agents/skills/moai*` **실 항목**"* 은 기존 백업 규칙을 따름 — 으로 구성된다. 3절이 **실 항목**으로 한정돼 있어(이 한정 자체는 D-iter4 의 무백업 손실 경로를 닫는 신중한 문구다), 다시 만들지 않을 **링크** 항목(AC-CSC-008 의 `moai-gone` · `moai-linkprobe` 가 정확히 그 형태)은 세 절 어디에도 걸리지 않는다.

실무상 "링크는 백업하지 않는다"가 자연스러운 해석이고 그것이 옳을 가능성이 높지만, D1 이 요구하는 링크 전용 분기를 쓸 때 구현자가 참조할 규정이 없다.

**Required fix**: REQ-CSC-010 에 한 절 추가 — 심볼릭 링크 미러 항목은 정본의 파생물이므로 대상 존재 여부와 무관하게 백업하지 않고 링크만 제거한다는 취지(그리고 그것이 §A.7 의 "링크 모드 — 안전" 서술과 일관됨을 명시).

---

## SPEC · 선행 감사 · preflight 노트의 주장과 어긋나는 것

리드가 별도로 요청한 축이다. **SPEC 이 §A 에 적은 실측은 재측정 결과 전부 정확했다.** 어긋나는 항목만 아래에 적는다.

**① SPEC §B.D6 의 폭발 반경 서술 (spec.md:222)** — *"영향 범위는 '관리 대상 뿌리 아래에 있으면서 대상이 사라진 심볼릭 링크' 전부다"*. **거짓**. 살아 있는 심볼릭 링크도 함께 영향받으며 그쪽 결과가 더 무겁다(D1, 실측 근거 포함).

**② SPEC iter-5 HISTORY (spec.md:21)** — *"optional 도 전부 반영했다( … D6 개수 표기 3건)"*. **부분적으로 거짓**. plan.md:101 의 `4형태` 가 갱신되지 않았고(D3), plan.md:19 의 `양팔` 도 남았다.

**③ SPEC §A.7 (spec.md:99)** — *"복사는 비정규 항목을 건너뛰며"*. **조건부로만 참**. `WalkDir` 경로(디렉터리 링크)에서는 참이지만, `!info.IsDir()` 파일 분기(파일 링크, 그리고 `Lstat` 전환 후에는 모든 링크)에서는 `copyRegularFile` 이 `os.ReadFile` 로 **대상 바이트를 복사한다**(실측 `data="keepme"`). 새 fixture 항목이 정확히 이 경로를 탄다.

**④ preflight 노트 `.moai/reports/t81/m1-preflight-measurements.md`** — SPEC §A.1 이 이미 정정한 "스킬 36개"는 재측정에서도 **34개**로 확인됐다(`find … -mindepth 1 -maxdepth 1 -type d | wc -l` → 34, `SKILL.md` 보유 34, 로컬 44 = 34 + `hns-*` 10). SPEC 의 정정이 옳고 preflight 노트의 36 이 틀렸다. AP-7(`ls | wc -l` +2) 설명도 산술적으로 일관된다.

**⑤ 선행 감사와의 관계** — iter-1~iter-3 어느 보고서도 D1(Lstat 분기 전환)을 지적하지 않았다. iter-4 개정이 `os.Lstat` 을 [HARD] 로 못박은 것은 dangling 팔의 **공허한 참**을 닫으려는 옳은 조치였으나, 같은 syscall 교체가 `info.IsDir()` 의미까지 바꾼다는 사실은 네 차례 감사와 다섯 차례 개정을 모두 통과했다. 이번 개정이 그 [HARD] 를 acceptance 판정 계층까지 확장하면서 노출 면이 넓어졌다.

**⑥ 재측정으로 확인된 §A 주장 (어긋남 없음)** — 참고로 열거한다. §A.1(34/44/10) · §A.3(catalog 스킬 34 = core 21 + optional-pack 13, `harness_generated.skills: []` → 0) · §A.4(`ManagedCleanTargets` 7항목 전부 `.claude/` 하위, 4번째가 `.claude/skills/moai*` 글롭) · §A.8(템플릿 `.gitignore` 에 `agents` 매치 0건) · §A.9(`moaiSkillPrefix = "moai-"`, `grep -cv '^moai'` → 0, `grep -cv '^moai-'` → 1 = `moai`, 함수 주석이 제외를 의도로 명시) · §A.9b(`io.Writer` 비테스트 매치 0건, `Deploy` 시그니처 및 `deployer` 필드 4개 문자 그대로 일치) · §A.10(clean :297 < deploy :323, `Stat` IsNotExist / `Lstat` nil / `Glob` 이 dangling 을 매치 — 전부 재현) · §A.11(`templateCarries` 가 deploy.go:380 `!info.IsDir()` 아래에만 존재).

---

## Regression Check (iteration 2+)

이월 blocking 결함 **0건**. iter-3 이 남긴 blocking 3건(D2 / D3 / D3-b)과 인용 결함(D1+D7)은 모두 닫혔고, optional 4건 중 3건이 닫히고 1건(D6 개수 표기)이 열려 있다.

그러나 **세 차례 연속으로 같은 패턴이 반복된다**: 개정이 이전 blocking 을 전부 닫으면서 새 blocking 을 들여온다(iter-3 판본 → 3건, iter-4 판본 → 3건, iter-5 판본 → 4건). 이것은 개별 결함의 미해결이 아니라 **개정 방식 자체의 문제**를 시사한다 — 판정 계층을 정밀화할 때마다 요구사항·설계 계층과의 상호작용이 새로 생기는데, 그 상호작용을 실행 가능성 기준으로 재검토하는 절차가 개정 루프에 없다. D1 이 그 전형이다: iter-4 가 판정 정밀화를 위해 도입한 `os.Lstat` 이 iter-5 에서 [HARD] 로 굳으면서, 공유 코드의 분기 의미 변화라는 미검토 축을 그대로 통과시켰다.

정체(stagnation) 판정은 해당 없음 — 세 판본에 걸쳐 변하지 않은 결함은 없다.

---

## STOP 신호 — 점수 회귀

| 판본 | 감사 | 집계 |
|---|---|---|
| v0.2.0 | iter1 | 0.775 |
| v0.3.0 | plan-audit | 0.78 |
| v0.3.0 | iter2 (독립 재판독) | 0.7625 |
| v0.4.0 | iter3 | 0.7875 |
| **v0.5.0** | **iter4 (본 감사)** | **0.7625** |

직전 비교 대상(v0.4.0, 0.7875) 대비 **-0.025 회귀**. LEAN Workflow 의 STOP escalation 조건에 해당하므로 `STOP` 을 발신하고, 오케스트레이터는 무조건 반복 대신 사용자에게 아래 세 갈래를 제시할 것을 권고한다.

1. **범위 축소 (권장)** — 아래 § 권고 참조.
2. **PASS-with-debt** — 현 판본을 부채와 함께 수용. **권장하지 않는다**: D1 은 문서 정합성이 아니라 *제품이 깨지는* 처방이며, 부채로 이월하면 run-phase 가 `moai update` 를 실패시키는 코드를 쓰거나 미규정 분기를 즉흥으로 발명한다.
3. **사용자 명시 override** — 5회차 속행.

---

## Recommendation

**FAIL. blocking 4건(D1~D4)을 닫기 전에는 run-phase 진입 불가.** 수정은 아래 순서로 — 앞의 것이 뒤의 것의 전제다.

1. **D1 을 먼저 닫는다 (critical).** 이것만이 설계 계층 결함이고 나머지 셋은 판정·문서 계층이다. §B.D6 의 폭발 반경을 "살아 있는 링크의 분기 전환 + dangling 제거" 양쪽으로 정정하고, **심볼릭 링크 대상의 처리 절**을 REQ-CSC-008(또는 신설 절)에 규정한다 — `Lstat` 이 `ModeSymlink` 를 보고하면 `copyRegularFile` 를 거치지 않고 링크만 제거. plan M4 의 (a) 를 그 분기 작성 작업으로 다시 쓴다. AC-CSC-008 에 "clean 이 오류 없이 완료된다"를 단언으로 추가한다.
2. **D2 를 닫는다.** `moai-linkprobe` 를 디렉터리 링크(`../../.claude/skills/hns-linkprobe-src`)로 바꾸고, 제품과 형태를 일치시켰다는 사실을 [HARD] 절에 적는다. 1번의 새 분기가 이 fixture 로 실제로 검증되게 하는 것이 목적이다.
3. **D4 를 닫는다.** AC-CSC-012 의 Given 에 복사 모드 배포(AC-CSC-005 의 주입 seam 재사용)를 추가해 2번 단언을 공허하지 않게 만든다.
4. **D3 을 닫는다.** plan.md:101 → `AC-CSC-008(5형태 6단언 …)`, plan.md:19 의 `(양팔)` 갱신.
5. optional D5·D6 은 오케스트레이터 재량. 다만 D6 은 1번 작업 중에 자연히 답이 나오는 자리이므로 함께 처리하는 편이 싸다.

**범위 축소 권고 (STOP 대응 옵션 1).** 다섯 판본을 거치며 요구사항이 12 → 16(Tier M 상한)에 닿았고, 점수는 0.775~0.7875 대에서 진동하며 임계를 넘지 못한다. 관찰되는 구조는 이렇다 — **하나의 SPEC 이 두 개의 독립된 변경을 함께 지고 있다.**

- **(가) 미러 배포**: REQ-CSC-001~007, 011~016. `internal/template` 에 국한되고, 새 코드만 추가하며, 기존 동작을 바꾸지 않는다. 이쪽 요구사항·AC 에서는 이번 감사가 blocking 을 하나도 찾지 못했다.
- **(나) 청소 경로 + `os.Lstat` 전환**: REQ-CSC-008~010. `internal/cli/update/deploy` 의 **공유 코드**를 바꾸고 기존 7개 청소 뿌리 전부에 영향을 준다. 다섯 판본의 blocking 결함이 **전부 이쪽에서** 나왔다(iter-3 D2·D3·D3-b, 본 감사 D1·D2·D3·D4 — 하나도 예외 없이).

(나)를 별도 SPEC 으로 떼면 (가)는 즉시 Tier S~M 으로 통과 가능한 형태가 되고, (나)는 공유 코드 변경에 걸맞은 자체 §A 실측(살아 있는 링크의 분기 전환, 기존 7개 뿌리에 실제로 존재하는 링크 인벤토리, EISDIR/ENOENT 경로)을 처음부터 갖춘 채 작성될 수 있다. 지금은 그 실측이 미러 배포 SPEC 의 부속으로 끼어들어 매 개정마다 상호작용을 새로 만들고 있다.

분리에 따르는 비용은 정직하게 적는다 — §A.4·§A.5 가 세운 명제("청소 규칙과 배포 규칙이 같은 집합을 가리키지 않으면 조용히 어긋난다")는 두 SPEC 사이의 순서 제약이 되고, (가)만 먼저 착지하면 은퇴 스킬 잔존이 일시적으로 열린다. 그 창을 감수할지는 **사용자 결정**이며 감사관이 정할 사항이 아니다.
