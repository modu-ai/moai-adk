# SPEC 감사 보고서: SPEC-CODEX-SKILLS-CANONICAL-001

Iteration: **2/2** (Tier M 상한 2 — 상한 도달)
감사 대상 스냅샷: `spec.md` v0.3.0 (`02:54:59`) / `acceptance.md` (`02:56:51`) / `plan.md` (`02:57:58`)
Verdict: **FAIL**
Overall Score: **0.78** (조화평균, `agent-common-protocol.md` § Skeptical Evaluation Stance) — 산술평균이면 0.80 으로 Tier M 임계와 **정확히 동률**. 어느 쪽이든 아래 critical 3건이 독립적으로 FAIL 을 강제한다.

Reasoning context ignored per M1 Context Isolation — 리드 지시문의 전제·유도 문구는 판정 근거에서 제외했다. 모든 수치는 이 worktree(`WT-skills-canonical`, `go1.26.4 darwin/arm64`)에서 직접 실행한 명령의 출력이다.

## 감사 중 개정이 발생했다 (기록)

이 감사는 v0.2.0 을 읽으며 시작했고, 측정 도중 작성자가 v0.3.0 으로 개정했다. 판정은 **현재 디스크 상태(v0.3.0)** 기준이며, v0.2.0 에 대해 수행한 측정 중 v0.3.0 에서 해소된 것은 아래 "해소 확인"으로 옮겼다. iteration 1 보고서는 `.moai/reports/t81/plan-audit-iter1.md` 로 보존했다(원래 `plan-audit.md`, 이 감사 시작 시 개명).

**개정이 아직 끝나지 않았을 가능성이 있다.** `plan.md` 가 가장 최근에 수정됐는데(02:57:58) §B 위험표만 갱신되고 §F 마일스톤은 v0.2.0 상태로 남아 있다 — 그 불일치가 아래 N10·N11 이다. 작성자가 이어서 고치는 중이라면 그 두 건은 이 보고서가 도착하기 전에 닫힐 수 있다.

---

## Must-Pass 결과

- **[PASS] MP-1 REQ 번호 일관성** — `grep -c '^- \*\*REQ-CSC-'` = **16**, `REQ-CSC-001 … REQ-CSC-016` 연속, 결번·중복 0, zero-pad 일관. Tier M 상한 16 에 정확히 도달(초과 아님).
- **[PASS] MP-2 GEARS 준수 (요구사항 계층 한정)** — 16개 전부 GEARS 5패턴 대응: ubiquitous(001·007·010·015·016), unwanted `shall not`(002), event-driven `When`(003·005·008·011·012·013·014), where-gate `Where`(004), state-driven `While`(009), 파생조건+`shall not`(006). 비형식 서술 0건. `acceptance.md` 의 Given-When-Then 은 **검증 계층(AC-XXX)** 이므로 이 판정 대상이 아니며 Group 4 에서 별도 채점했다. §G 가 REQ-006·011 의 원자성 미분리를 예산 근거와 함께 **명시적으로 기각**한 것은 MP-2 위반이 아니다 — 두 항목 모두 형식 GEARS 이고, 근거도 타당하다(상한 도달 + 재번호가 세 판정 계층의 매핑을 흔든다).
- **[PASS] MP-3 frontmatter 유효성** — canonical 12필드 전부 존재·타입 적합. `version: "0.3.0"` quoted semver, `status: draft`, ISO 날짜, `priority: P2`, `lifecycle: spec-anchored`, `tags` CSV. snake_case alias 0건. 추가 필드 `tier: M` 허용.
- **[N/A] MP-4 §22 언어 중립성** — 본 SPEC 은 moai-adk 자체의 Go 배포기·템플릿 트리를 다루며 16개 지원 언어의 도구 체인을 열거·차등하지 않는다. 단일 언어 범위 → 자동 PASS.
- **[PASS] MP-5 D7 교차-SPEC 정합** — `grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md | sort -u` → 자기 자신 1건뿐. 외부 참조 0 → BLOCKING 없음.
- **[PASS] MP-6 D8 크로스 플랫폼 규율** — `grep -c syscall spec.md` = **0**. 자동 PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-CODEX-SKILLS-CANONICAL-001/` → rc=1(매치 0). Tier M 이라 `research.md` 부재는 정상.

Must-pass 는 전부 통과다. **FAIL 은 must-pass 위반이 아니라 blocking 결함 6건(critical 3 포함)과 조화평균 0.78 < 0.80 에서 나온다.**

---

## 범주별 점수 (rubric 앵커)

| 차원 | 점수 | 밴드 | 근거 |
|---|---|---|---|
| Clarity | 0.85 | 0.75–1.0 | REQ 16개가 단일 해석을 갖고, iter-1 이 지적한 모호성(대상 선점 상태·manifest 방향)이 REQ-013/014/010 으로 해소됐다. 남은 감점: REQ-CSC-005 가 "사용자에게 보고"의 **표면을 지정하지 않는데 그 표면이 실재하지 않는다**(N3). |
| Completeness | 0.75 | 0.75 | spec 쪽은 `.gitignore`(§A.8/REQ-016)·백업 비대칭(§A.7/REQ-010)·대상 선점(REQ-013/014)·접두 불변식(§A.9/REQ-015)까지 덮어 iter-1 의 공백 네 개가 닫혔다. 남은 감점 둘: REQ-CSC-008 이 **링크 형태의 미러 제거를 규정하지 않고**(N1), `plan.md` §F 가 spec 개정을 따라오지 못해 M3 이 REQ-CSC-010 과 **정면 충돌**한다(N10·N11·N12). |
| Testability | 0.65 | 0.50–0.75 | AC 계층이 크게 좋아졌다 — AC-001(양팔, `d.IsDir()` 금지 명시)·AC-010(동일 프로세스 seam 토글)·AC-011(3-상태)·AC-012(양팔)는 iter-1 지적을 정확히 닫았다. 그런데 **MUST 인 AC-008·009 가 실 디렉터리를 심어 이 SPEC 최대 결함을 검출하지 못하고**(N2), AC-006·011(3)·013(3) 세 개가 **존재하지 않는 출력 seam** 을 전제한다(N3). 주요 요구사항 위에 판정력 구멍이 남는다. |
| Traceability | 0.95 | 1.0 근접 | §D 매트릭스 16행 + §D.3 역방향 매핑이 REQ 16개를 전부 덮고 고아 AC 0. REQ-012/013/014 → AC-011 의 세 단언, REQ-009 → AC-008(생존 팔)/AC-009 의 의도적 중복도 근거와 함께 명시됐다. |

- 조화평균 = 4 / (1/0.85 + 1/0.75 + 1/0.65 + 1/0.95) = **0.784 → 0.78**
- 산술평균 = (0.85+0.75+0.65+0.95)/4 = **0.80** (임계와 동률)

점수 추이: iter-1 **0.775**(v0.2.0, 산술) → 이번 **0.80 산술 / 0.78 조화**(v0.3.0). 개정을 사이에 두고 점수가 **올랐다** — LEAN 워크플로의 STOP 에스컬레이션(점수 역행)은 해당 없다.

> 산출 방식 주의: iter-1 은 산술평균을 썼다. 이번에 조화평균을 채택한 것은 auditor 스탠스(`agent-common-protocol.md` § Skeptical Evaluation Stance — "harmonic mean of dimensions, not the average")를 따른 것이며, 두 값을 모두 적어 방식 변경만으로 FAIL 이 만들어지지 않았음을 보인다. 산술 0.80 을 채택하더라도 critical 3건이 blocking 이므로 verdict 는 동일하다.

---

## Defects Found

### N1 — 청소 글롭은 **심볼릭 링크 미러를 한 개도 지우지 못한다**. REQ-CSC-008 의 목적이 실행 시점에 무효화된다 (blocking)

`spec.md:REQ-CSC-008` + `internal/cli/update/deploy/deploy.go:371-378` + `internal/cli/update_template_sync.go:289-303` — Severity: **critical** — Class: **blocking**

세 사실이 겹친다. 전부 이 worktree 에서 측정했다.

**(1) `moai update` 는 clean → deploy 순서다.** `update_template_sync.go:297` 의 clean 스텝이 `:302` 의 Deploy 스텝보다 앞선다(스텝 배열 순서, 코드 확인).

**(2) clean 이 `.agents/skills/moai*` 를 처리하는 시점에는 정본이 이미 없다.** `CleanMoaiManagedPaths` 는 `ManagedCleanTargets` 슬라이스를 **순서대로** 돈다(`deploy.go:112`). 기존 4번째 항목이 `.claude/skills/moai*` 글롭(`deploy.go:64-68`)이고, D5/M4 가 지시하는 신규 항목은 그 뒤에 추가된다. 그러면 `.agents/skills/<name>` 링크가 처리될 때 대상 `.claude/skills/<name>` 은 **같은 실행에서 이미 삭제**돼 있다 — 은퇴 스킬만이 아니라 **모든 미러 링크가 dangling 상태**다.

**(3) dangling 링크는 무음으로 건너뛰어진다.** `backupThenRemove` 의 첫 동작이 `os.Stat`(링크를 따라감)이고 `os.IsNotExist` 면 `return 0, nil` — 제거 없이 성공으로 반환한다. 측정:

```
case2 os.Stat(dangling link): err=... no such file or directory  IsNotExist=true
  but Lstat sees it: true
case3 glob matches: 1 [.../.agents/skills/moai-gone]
```

`filepath.Glob` 은 dangling 링크를 **매치한다**. 청소는 대상을 찾아내고도 지우지 않는다.

귀결: 은퇴·개명된 스킬의 미러 링크가 사용자 프로젝트에 **영구 잔존**한다. §A.4 가 `~/.codex/skills/` 오염(직접 재측정: `moai-lang-*` **15개**, `moai-platform-*` **4개**, 2026-06-07)으로 관측한 바로 그 실패 형태이며, REQ-CSC-008 은 그것을 막으려고 존재한다. 지금 규정대로 구현하면 **막지 못한다**.

경계도 확인했다 — **링크가 살아 있을 때는** 안전하다:

```
case1 os.Stat(live link): isDir=true
  walkdir: p=moai-x isDir=false      ← WalkDir 은 루트를 Lstat, 정규 파일 0개
  RemoveAll(link) err: <nil>
  canonical survives RemoveAll(link): true
  link removed: true
```

즉 §A.7 의 "링크 모드 — 안전" 서술은 옳다. 문제는 링크가 **살아 있지 않은 순간**이고, 실제 실행 순서가 그 순간을 보장한다.

Required fix: REQ-CSC-008 에 절을 하나 더한다 — "청소는 `.agents/skills/moai*` 대상을 **`os.Lstat` 기준으로 판정**해야 하며(shall), 대상이 존재하지 않는 링크(dangling)도 제거해야 한다(shall)". 순서 제약(`.agents/` 항목을 `.claude/skills/moai*` 앞에 두기)으로 우회하는 형태는 취약하므로 권하지 않는다 — 순서는 나중에 조용히 바뀐다.

### N2 — AC-CSC-008 / AC-CSC-009 는 **실 디렉터리**를 심어 N1 을 검출할 수 없다 (blocking)

`acceptance.md:87-100` — Severity: **critical** — Class: **blocking**

리드가 지목한 "양팔 단일 테스트" 구조 자체는 확인했다 — AC-CSC-008 은 제거 단언과 생존 단언을 한 테스트에 묶고(L89-94), AC-CSC-009 는 `harness-legacy`·`my-harness-legacy`·`my-own` 으로 접두 밖 전체로 넓힌다. **구조는 옳다. fixture 의 형태가 틀렸다.**

두 AC 모두 `.agents/skills/<name>/SKILL.md` 를 **실 파일로** 심는다. 그런데 이 SPEC 이 배포하는 산출물은 D1·REQ-CSC-003 에 따라 **심볼릭 링크**다. 실 디렉터리는 `os.Stat` 이 성공하므로 청소가 정상 동작하고 두 AC 는 **통과한다** — N1 이 살아 있는 상태에서. MUST 로 분류된 AC 두 개가 이 SPEC 최대 결함의 정확히 반대편만 검사한다.

AC-CSC-008 의 [HARD] 주석은 "제거만 단언하는 테스트"를 경계하는데, 실제로 놓치는 축은 제거/생존이 아니라 **실 항목 / 링크 항목**이었다.

부가로 AC-CSC-009 의 생존 단언은 **거의 반증 불가능**하다. `filepath.Glob(".agents/skills/moai*")` 는 `hns-*`·`harness-*`·`my-own` 을 원리적으로 매치할 수 없다. 실패하려면 구현자가 글롭을 문자 그대로 다르게 써야 하고 그 경우 AC-CSC-007 이 먼저 잡는다. 판정력이 0 은 아니지만(AP-6 형태의 `.agents/skills/*` 는 잡는다) 매우 얇다.

Required fix: AC-CSC-008 의 fixture 를 **네 형태**로 확장하고 한 테스트에 둔다 — (i) `moai-gone` = 정본이 살아 있는 링크, (ii) `moai-dangling` = 대상이 이미 지워진 링크, (iii) `moai-copied` = 복사 모드 산출물인 실 디렉터리, (iv) `hns-user-owned` = 사용자 소유 실 디렉터리. 단언: (i)(ii)(iii) 제거 + (iv) 내용 무변경 + **정본 `.claude/skills/moai-gone/` 이 링크를 통해 지워지지 않았음**. (ii) 가 N1 을 잡는 유일한 팔이다.

### N10 — `plan.md` §F M3 이 `REQ-CSC-010` 과 **정면으로 모순된다** (blocking)

`plan.md:63-67` ↔ `spec.md:REQ-CSC-010` + `acceptance.md:128-137` — Severity: **critical** — Class: **blocking**

v0.3.0 개정은 REQ-CSC-010 의 방향을 **뒤집었다**: "미러 항목을 manifest 에 기록해야 한다" → "**기록해서도 안 되고**(shall not) pre-clean 백업 트리에 보존해서도 안 된다(shall not)". 근거는 §A.6(`manifest.Track` 이 디렉터리 링크에서 EISDIR)과 §A.7(복사 모드 전량 복제)이며, 두 실측 모두 독립적으로 재현했다(아래 "해소 확인" D7 참조).

그런데 `plan.md` §F M3 은 v0.2.0 문구 그대로다:

> ### M3 — manifest 기록 (Priority Medium)
> 미러 항목을 `template_managed` 로 기록한다. 링크 모드와 복사 모드에서 기록 형태가 달라야 하는지(복사본은 파일 단위, 링크는 항목 단위) 여기서 정한다.
> - 닫힘 조건: AC-CSC-012

M3 이 지시하는 작업은 REQ-CSC-010 이 금지하는 바로 그 작업이고, 그 닫힘 조건으로 걸린 AC-CSC-012 는 이제 **기록이 없음**을 단언한다. 구현자가 plan 을 따르면 REQ 를 위반하고 자기 닫힘 조건에서 실패한다. `plan.md` 의 mtime 이 셋 중 가장 최근(02:57:58)이고 §B 위험표에는 R12(`manifest.Track` EISDIR)가 추가돼 있으므로 — 같은 파일 안에서 위험표는 새 방향을, 마일스톤은 옛 방향을 말한다.

Required fix: M3 을 새 방향으로 다시 쓴다 — "미러는 manifest 에 기록하지 않고 pre-clean 백업에서도 제외한다. 기록 금지가 `Track` 호출부의 자연스러운 형태와 충돌하지 않는지 확인한다." 제목도 "manifest 기록"에서 바꾼다.

### N3 — REQ-CSC-005 / AC-CSC-006 · 011(3) · 013(3) 이 **존재하지 않는 seam** 을 전제한다 (blocking)

`spec.md:REQ-CSC-005` + `acceptance.md:73-75, 122, 147` + `plan.md:59` + `internal/template/deployer.go:100` — Severity: **major** — Class: **blocking**

AC-CSC-006 은 "`Deploy` 의 출력 writer 를 버퍼로 캡처하면"이라고 쓴다. 실제 시그니처는

```go
func (d *deployer) Deploy(ctx context.Context, projectRoot string, m manifest.Manager, tmplCtx *TemplateContext) error
```

— **`io.Writer` 가 없다.** `internal/template` 전체에 printer 계층이 없다. 출력은 호출부(`internal/cli/update_template_sync.go` 의 `out` / `tui.ProgressLine`)에만 있고 `moai init` 은 또 다른 배선을 쓴다. `plan.md` §F M2 의 "출력 경로는 기존 printer 계층 재사용"은 **사실이 아니다**.

v0.3.0 에서 이 의존이 **하나에서 셋으로 늘었다** — AC-CSC-011 의 3번 단언("건너뛰었다는 경고가 출력 버퍼에 나타난다")과 AC-CSC-013 의 3번 단언도 같은 seam 을 요구한다. 셋 중 둘이 MUST 다. 만족시키려면 `Deployer` 인터페이스(생성자 4종 + 다수 호출부)를 바꾸거나 별도 콜백 seam 을 도입해야 하는데, spec §D 의 "구현 세부는 run-phase 판단"으로 넘길 수 있는 크기가 아니고(공개 인터페이스 변경) plan 어디에도 범위로 잡혀 있지 않다.

Required fix: (a) 배포기가 모드·경고를 알리는 **경로를 REQ 수준에서 지정**한다 — 예컨대 배포 결과 구조체에 모드/경고 목록을 담아 반환하고 사용자 표시는 호출부가 한다(인터페이스 변경 최소화). (b) AC-006·011(3)·013(3)을 그 seam 에 맞춰 다시 쓴다. (c) `plan.md` §F M2 의 "기존 printer 계층 재사용"을 삭제하거나 실제 대상으로 교체한다. `acceptance.md` §D.7 이 AC-CSC-010 의 seam 토글을 "run-phase 설계 제약"으로 이미 올려놓았듯, 이 출력 seam 도 같은 자리에 올라가야 한다.

### N12 — AC-CSC-012 의 2번 팔이 요구하는 **백업 제외 로직을 어떤 마일스톤도 소유하지 않는다** (blocking)

`acceptance.md:135` + `plan.md:69-79` + `internal/cli/update/deploy/deploy.go:390-398` — Severity: **major** — Class: **blocking**

AC-CSC-012 의 2번 단언은 "`.moai-backups/**/pre-clean/.agents/` 아래 파일 수가 **0**이고, **복사 모드에서도** 0"이다. 복사 모드에서 미러는 실 디렉터리이므로 `backupThenRemove` 의 디렉터리 분기를 타고, 템플릿에 `.agents/` 가 없어 `templateManagedPaths` 가 공집합 → `backupUnmanagedTree` 가 **전량을 백업한다**(§A.7 이 정확히 이 경로를 서술한다). 즉 이 단언을 참으로 만들려면 `backupThenRemove` 또는 `CleanMoaiManagedPaths` 에 **백업 제외 분기를 추가하는 코드 변경**이 필요하다.

그런데 `plan.md` §F 에서 이 작업을 담는 마일스톤이 없다. M3 은 manifest(그나마 방향이 뒤집혔다 — N10), M4 는 "`ManagedCleanTargets` 에 글롭 추가"이고 닫힘 조건도 AC-007/008/009 뿐이다. AC-CSC-012 는 M3 의 닫힘 조건에만 걸려 있는데 M3 은 백업을 다루지 않는다.

Required fix: M4 의 범위에 "`.agents/` 대상은 pre-clean 백업에서 제외" 를 추가하고 닫힘 조건에 AC-CSC-012(2번 팔)를 건다. 또는 백업 제외 전용 마일스톤을 세운다. 어느 쪽이든 **소유자 없는 요구가 남아서는 안 된다** — spec 이 REQ 로 세운 것을 plan 이 아무에게도 배정하지 않으면 run-phase 에서 조용히 빠진다.

### N11 — `plan.md` §F M5 가 `REQ-CSC-016` 을 따라오지 못했다 (blocking)

`plan.md:81-85` ↔ `spec.md:REQ-CSC-016` + `acceptance.md:155-163` — Severity: **major** — Class: **blocking**

REQ-CSC-016 은 "배포되는 `.gitignore` 는 `.agents/` 를 무시해야 한다(shall)"로 확정됐고 AC-CSC-015 는 세 단언(항목 존재 / 임베드 FS 반영 / 중립성 게이트)으로 그것을 판정한다. §B 위험표에도 R9 로 "**높음**" 등급이 올라갔다.

그런데 M5 는 v0.2.0 문구 그대로다 — "`.gitignore` 에 `.agents/` 취급이 **필요한지 확인한다**", Priority **Low**, "기계 작업". 확정된 요구를 미확정 조사 항목으로, 높음 위험을 Low 우선순위로 적고 있다. AC-CSC-015 의 2번 단언(`make build` 반영)은 특히 M5 가 "기계 작업"으로 뭉뚱그린 것보다 구체적인 절차를 요구한다.

Required fix: M5 를 "`.gitignore` 에 `.agents/` 항목 추가 + `make build` 반영"으로 확정하고 우선순위를 올린다. §B R9 등급(높음)과 일치시킨다.

### N4 — `plan.md` §H 의 판별 기준이 **카탈로그에 실재하는 스킬을 삭제 후보로 분류한다** (blocking)

`plan.md:103` + `internal/template/skills_manifest.go:14,42` — Severity: **major** — Class: **blocking**

§H 는 `~/.codex/skills/` 정리 대상을 "현재 배포 카탈로그(`template.EmbeddedMoaiSkillNames()` **또는** `internal/template/templates/.claude/skills/*/`)에 **없는** `moai*` 항목"으로 정의한다. 두 판별자는 **같은 집합이 아니다**.

`EmbeddedMoaiSkillNames()` 는 접두 `moai-`(대시 포함)로 필터하며 주석이 의도를 명시한다: *"The trailing dash is significant: it excludes the bare `moai` unified skill directory (no trailing dash) from the core-skill set."*(`skills_manifest.go:12-14`)

측정: 템플릿에 `internal/template/templates/.claude/skills/moai/` 가 **존재**하고(`catalog.yaml:8`), 사용자 홈에도 `~/.codex/skills/moai`(2026-06-07)가 **존재**한다. 판별자로 `EmbeddedMoaiSkillNames()` 를 쓰면 `moai` 는 "카탈로그에 없음" → **삭제 후보**가 된다. 반면 청소 글롭 `moai*`(대시 없음)와 새 REQ-CSC-015("`moai` 접두", 대시 없음)는 같은 이름을 관리 대상으로 잡는다 — §A.5 가 경계한 비대칭이 §H 안에서 재현된다.

§H 가 문서화 전용이고 4단계 승인 절차를 두었다는 점이 완화하지만, 절차 1단계("대상 목록을 먼저 출력해 사람이 읽는다")가 **잘못된 목록**을 출력한다. 사람이 승인할 근거가 오염된다.

부수 위험: AC-CSC-016("임베드 템플릿 FS 의 `.claude/skills/` 1단계 디렉터리 이름 전부")은 문구상 옳게 쓰여 있으나, 구현자가 §H 를 읽고 `EmbeddedMoaiSkillNames()` 로 구현하면 bare `moai` 를 조용히 건너뛴다. 두 곳이 다른 접두를 쓰는 상태 자체가 함정이다.

Required fix: §H 의 판별자를 **하나로** 고정한다 — `internal/template/templates/.claude/skills/*/` 의 디렉터리 이름 집합(bare `moai` 포함). `EmbeddedMoaiSkillNames()` 언급은 삭제하거나, 쓰려면 "이 함수는 bare `moai` 를 제외하므로 별도 취급" 을 명시한다. AC-CSC-016 본문에도 "`EmbeddedMoaiSkillNames()` 로 구현하지 않는다"를 한 줄 넣기를 권한다.

### N5 — §D.4 의 Codex 노출 제외 사유가 t91 의 방법론과 모순된다 (optional)

`acceptance.md:§D.4` ↔ `.moai/reports/t91/README.md:83` — Severity: **minor** — Class: **optional**

v0.3.0 의 §D.4 는 상대 링크 미실측 경고(iter-1 D10)를 훌륭히 흡수했다. 그러나 제외 사유는 그대로다 — "`codex` 바이너리에서 `/skills` 에 스킬이 보이는지는 **사용자 홈 상태에 의존하므로 기계 판정 불가**".

t91 은 정확히 그 판정을 기계적으로 수행했다 — `CODEX_HOME=<scratch>/home` 으로 사용자 홈을 **격리**하고 `codex debug prompt-input`(모델 호출 **0회**)으로 노출 여부를 표로 관측했다(t91 §4, 경로 5종). 즉 "사용자 홈 의존"은 t91 이 이미 제거한 변수다.

제외 결정 자체는 옳을 수 있다(CI 러너에 `codex` 바이너리가 없고 버전 의존성이 크다). 그러나 적힌 사유가 이미 반증된 근거 위에 서 있고, 이 SPEC 은 §A 전체가 "남의 잘못된 전제를 정정한다"는 자세로 서 있다.

Required fix: 사유를 실제 이유로 교체한다 — "CI 러너에 `codex` 바이너리가 없고 버전 의존성이 커 게이트로 삼지 않는다. 운영자 수동 확인은 t91 §4 의 격리 방식(`CODEX_HOME=<scratch>`, `codex debug prompt-input`)을 따른다."

### N6 — §A.2 의 embed 서술은 조건을 명시하지 않는다. 경계를 재측정했다 (optional)

`spec.md:§A.2` — Severity: **minor** — Class: **optional**

§A.2 의 결론은 **옳다**. 독립 최소 재현(`go1.26.4`, 링크 4형태 × 패턴 2형태)으로 확인했다:

```
--- all:templates              --- plain templates
  templates/inner.txt            templates/inner.txt
  templates/innerdir             templates/innerdir
  templates/innerdir/g.txt       templates/innerdir/g.txt
```

심어둔 `link_to_inner.txt`(루트 **내부** 파일 링크), `link_out.txt`(**외부** 파일 링크), `linkdir_in`(내부 디렉터리 링크), `linkdir_out`(외부 디렉터리 링크) **4개 전부**가 `all:` 과 plain 양쪽에서 사라졌고 빌드 오류·경고 0건이다. SPEC 서술보다 오히려 **넓게** 성립한다.

경계가 하나 있다. 링크를 **패턴에서 직접 지목**하면 무음이 아니라 **빌드 오류**다:

```
main.go:8:12: pattern templates/link_to_inner.txt: cannot embed irregular file templates/link_to_inner.txt
```

"무음 소실"은 **디렉터리 패턴 임베드에 한정된 조건부 사실**이다. 이 프로젝트의 `//go:embed all:templates` 가 정확히 그 조건이므로 설계 결론은 바뀌지 않는다.

Required fix: §A.2 에 한 줄 — "디렉터리 패턴 임베드에서 무음 소실이며, 링크를 패턴에 직접 지목하면 `cannot embed irregular file` 빌드 오류다. 본 프로젝트는 전자에 해당한다." 회귀 가드(AC-CSC-001)가 필요한 이유가 이 조건에서 나온다.

### N8 — Codex 스킬 리스팅 예산이 34개 노출을 허용하는지 미확인 (optional)

`spec.md:§A.3` ↔ `.moai/reports/moai-adk-dual-harness-codex-20260817.md:46` — Severity: **minor** — Class: **optional**

재설계 문서는 Codex 스킬 리스팅 예산을 "**컨텍스트 2% 또는 8,000자**"로 기록한다. §A.3 은 재설계 문서의 성공 지표("Codex `/skills` 에 32개 노출")가 **tier 때문에** 달성 불가능하다고 정정하는데 예산 축은 다루지 않는다. 링크를 34개 만든다는 것이 34개가 노출된다는 뜻은 아니다.

REQ 는 "`.agents/skills/<name>/SKILL.md` 로 읽을 수 있는 접근 경로"까지만 요구하므로 REQ 위반은 아니다. 그러나 §A.3 이 성공 지표를 정정하는 절인 이상 한 축만 다룬 상태로 남는다.

Required fix: §A.3 에 한 줄 — "노출 개수는 tier 외에 Codex 의 리스팅 예산에도 걸리며, 그 실측은 본 SPEC 범위 밖이다."

### N9 — §F 교차 참조 1건이 이 worktree 에서 해석되지 않는다 (optional)

`spec.md:§F` — Severity: **minor** — Class: **optional**

`.moai/reports/moai-adk-dual-harness-codex-20260817.md` 는 primary 체크아웃에만 존재하고 worktree `WT-skills-canonical` 에는 없다(`ls` 확인). §A 가 정정하는 전제의 출처인데 감사자·구현자가 이 트리 안에서 열 수 없다.

Required fix: 참조에 "primary 체크아웃 기준"을 명시하거나 해당 문서를 worktree 로 가져온다.

---

## 회귀 점검 — iteration 1 결함 11건

`plan-audit-iter1.md` 의 D1~D11 을 v0.3.0 에 대해 재확인했다. **spec·acceptance 쪽은 9건 해소, plan 쪽은 2건이 스펙만 고쳐지고 계획이 따라오지 못했다**(N10·N11 로 재개봉).

| # | iter-1 결함 | 상태 | 근거 |
|---|---|---|---|
| D1 | AC-CSC-001 이 `d.IsDir()` 수집이라 링크가 양쪽 집합에서 동시에 빠져 통과 | **RESOLVED** | AC-CSC-001 이 양팔로 재작성 — 1번 트리 전체 `d.Type()&fs.ModeSymlink` 카운트 0, 2번 `d.IsDir()` 사용 금지 명시([HARD] 주석 포함). REQ-002 의 트리 전체 범위도 반영 |
| D2 | AC-CSC-010/013 이 Go 테스트로 만들 수 없는 커밋 기준선을 참조 | **RESOLVED** | AC-CSC-010 이 "동일 프로세스 seam 토글 불변식"으로 재작성되고 [HARD] 로 커밋 기준선 형태를 금지. 일회성 수동 대조는 §D.4 로 이동. AC-013 2번도 그 불변식을 참조 |
| D3 | 미러 집합 ⊅ 청소 집합 비대칭이 단언되지 않음 | **RESOLVED** | REQ-CSC-015 신설 + AC-CSC-016 + §A.9. 불변식 자체도 재측정 확인 — 템플릿 34개 중 `moai` 로 시작하지 않는 이름 **0개** |
| D4 | 복사 모드에서 매 update 마다 미러 전량이 pre-clean 백업으로 복제 | **RESOLVED(spec)** / 소유자 없음(plan) | REQ-CSC-010 + AC-CSC-012(2번 팔) + §A.7 + R8. 다만 이를 구현할 마일스톤이 없다 → **N12** |
| D5 | `.gitignore` 의 `.agents/` 취급이 REQ·AC 없이 plan Low | **RESOLVED(spec)** / plan 미반영 | REQ-CSC-016 + AC-CSC-015 + §A.8 + R9. M5 는 여전히 "필요한지 확인" Low → **N11** |
| D6 | REQ-CSC-012 가 잘못된 링크·사용자 실 디렉터리를 규정하지 않음 | **RESOLVED** | REQ-013(교체) + REQ-014(건너뛰고 경고, 제거·덮어쓰기 금지) + AC-CSC-011 3-상태 단일 테스트. [HARD] 로 `EEXIST` → "지우고 재생성" 형태를 데이터 손실로 못박음 |
| D7 | REQ-CSC-010/AC-012 가 manifest seam 으로 구현 불가 | **RESOLVED(spec)** / plan 모순 | REQ-CSC-010 방향 반전 + §A.6. 독립 재측정으로 §A.6 확정: `HashFile(symlink-to-dir) err: read ...: is a directory` / `HashFile(real dir) err: ... is a directory` / `HashFile(through link, file): 0deeb8fa1dbb <nil>` — 링크·실 디렉터리 **양쪽 다** 기록 불가. 단 M3 이 옛 방향 그대로 → **N10** |
| D8 | AC-CSC-007 슬래시 리터럴이 Windows 에서 실패 | **RESOLVED** | `filepath.ToSlash(t.DisplayPath) == ".agents/skills/moai*"` 로 교체 + [HARD] 사유 |
| D9 | §A.3 tier 분해 오류 | **RESOLVED** | "non-core 13 전부 `optional-pack:*`" 로 정정 + AP-7 자기 적용 문장 추가. 재측정 일치: `core 21` / `optional-pack:backend 3` · `design 1` · `devops 5` · `frontend 4` = 13, `harness-generated` tier 는 에이전트 `builder-harness` 단 하나(`catalog.yaml:261-268`) |
| D10 | Codex 의 **상대** 링크 해석 미실측 | **RESOLVED** | §D.4 에 명시 편입 — t91 이 상대/절대·파일/디렉터리를 기록하지 않았음을 적고 운영자 확인 항목으로 승격 |
| D11 | §B.D5 의 doctrine 인용이 기계적 보호 범위를 과장 | **RESOLVED** | §B.D5 에 [HARD] 단락 추가 — `IsUserOwnedNamespace` 가 `.claude/` 접두로만 판정해 `.agents/` 는 시야 밖이며 "글롭이 좁다는 사실 하나"가 유일한 방어임을 명시 |

D12(REQ 원자성)는 §G 에서 예산 근거와 함께 **명시적으로 기각**됐다. 기각 사유(상한 16 도달 + 재번호가 세 판정 계층 매핑을 흔든다 + 각 절이 대응 AC 에서 갈라져 실행 시점에 구분된다)는 타당하다. 조용히 넘기지 않고 기각을 기록한 형태가 옳다.

---

## 내 측정이 SPEC · 선행 실측 · t91 과 어긋난 지점

리드가 별도로 요구한 항목이다. **v0.3.0 기준으로 결론이 뒤집힌 전제는 없다.** 어긋난 것은 부분 수치와 사유 서술이며, 개정이 대부분을 이미 흡수했다.

| 출처 | 적힌 것 | 내 측정 | 처리 |
|---|---|---|---|
| `spec.md:§A.2` | `//go:embed` 는 링크를 무음으로 버린다 (무조건 서술) | 조건부로 참 — **디렉터리 패턴**에서 무음, **직접 지목**하면 `cannot embed irregular file` 빌드 오류. 무음 범위는 오히려 더 넓다(링크 4형태 × `all:`/plain) | 설계 결론 무영향, 조건 명시 필요 (N6) |
| `acceptance.md:§D.4` | Codex 노출은 "사용자 홈 상태에 의존하므로 기계 판정 불가" | t91 이 `CODEX_HOME=<scratch>` + `codex debug prompt-input`(모델 호출 0)으로 정확히 기계 판정했다 | 제외 결정은 유지 가능, **사유가 틀렸다** (N5) |
| `spec.md:§A.3` (v0.2.0) | non-core 13 = `optional-pack:*` 12 + `harness-generated` 1 | optional-pack **13**, harness-generated **0** | **v0.3.0 에서 정정 완료** (D9 RESOLVED) |
| `m1-preflight-measurements.md:§D` | `ManagedCleanTargets` 관리 대상에 `.moai/config` 포함(8개 나열) | 그 함수의 반환은 **7개, 전부 `.claude/` 하위**. `.moai/config` 는 같은 함수가 아니라 `CleanMoaiManagedPaths` 의 별도 코드 경로(`deploy.go:165`)에서 지워진다 | **선행 실측의 오류.** spec §A.4 는 "7개, 전부 `.claude/` 하위"로 이미 정정 — SPEC 쪽이 옳다 |
| `m1-preflight-measurements.md:§D/§E` | `~/.codex/skills/` 에 구 moai 스킬 **46개** | 디렉터리 **45개**(= `moai*` 44 + `.system` 1). `moai-lang-*` 15, `moai-platform-*` 4 는 확인 | 46 은 재현되지 않는다. §A.1 이 지적한 `ls` 별칭 편향과 같은 계열. spec 은 "다수"로 표현해 채택하지 않았다 — SPEC 쪽이 옳다 |
| `t91/README.md:93` | `.agents/skills/t91-link-src -> .claude/skills/t91-link-src` | 액면대로 읽으면 `.agents/skills/` 기준 상대 경로로 **해석되지 않는** 링크다. 보고서는 상대/절대·파일/디렉터리를 기록하지 않았다 | "Codex 가 링크를 따라간다"는 성립. REQ-CSC-003 의 `../../` 형태는 미실측 — **v0.3.0 §D.4 가 이미 흡수** |
| `plan.md:§H` | 판별자로 `EmbeddedMoaiSkillNames()` 또는 템플릿 디렉터리 목록 | 두 집합이 다르다 — 전자는 접두 `moai-`(대시)라 bare `moai` 를 제외하며, 그 스킬은 템플릿에도 `~/.codex/skills/` 에도 실재한다 | **미해소** (N4) |

역으로, 내 측정이 **확증한** 전제:

- **§A.1 스킬 인벤토리** — 템플릿 34 / 로컬 44 / `hns-*` 10 / `SKILL.md` 34, 템플릿 트리 링크 **0**, `.agents/` 양쪽 부재. 추가로 두 이름 집합의 **동일성**까지 확인했다(`diff <(템플릿 34) <(로컬 non-hns 34)` → 차이 0). "44 = 34 + 10" 은 합뿐 아니라 집합으로도 참이다.
- **§A.2** — 위 조건 단서를 붙여 성립.
- **§A.3** — slim FS 는 `internal/cli/init.go:658` 에서만 쓰이고 update 경로는 전량 FS 를 쓴다. `shouldDistributeAll`(`init.go:362-367`)은 `--all` 또는 `MOAI_DISTRIBUTE_ALL ∈ {1, true}` 에서만 참. "21 또는 34" 는 정확하다.
- **§A.4** — `ManagedCleanTargets` 7개 전부 `.claude/` 하위, `.agents/` 없음. 사용자 홈 오염 실재(`moai-lang-*` 15, `moai-platform-*` 4).
- **§A.6** — `manifest.Track` → `HashFile` EISDIR, 링크·실 디렉터리 양쪽. 재현 출력 위 표에 인용.
- **§A.7 링크 모드 안전** — `WalkDir` 이 링크 루트를 `Lstat` 해 정규 파일 0개, `RemoveAll` 은 링크만 제거, 정본 생존. 측정 출력 N1 에 인용.
- **§A.9 접두 불변식** — 템플릿 34개 중 `moai` 로 시작하지 않는 이름 0개.
- **AC-CSC-003 은 검사 가능하다** — `template.NewSlimDeployerWithRenderer` 가 실재하므로 slim 배포기를 테스트에서 구성할 수 있고, v0.3.0 이 추가한 3번 단언(슬림 < 전량)이 리드가 물은 "동어반복" 문제를 정확히 닫는다. 배포 후 대상 FS 를 다시 읽는 구현에서도 3번은 반증력을 갖는다.
- **AC-CSC-011 은 2회차 Deploy 를 실제로 태운다** — 그리고 v0.3.0 이 3-상태로 확장해 `moai init` 재실행 경로를 명시적으로 범위에 넣었다(126행).

---

## Recommendation

FAIL. Tier M 상한 2회에 도달했으므로 Retry Loop Contract 상 **오케스트레이터는 사용자에게 에스컬레이션**해야 한다 — (1) PASS-with-debt(잔여 결함 기록 후 run 진입), (2) 범위 축소, (3) 명시적 상한 연장 중 선택. 다만 아래 1·2 는 **한 번의 개정으로 닫히는 크기**이고 그중 절반은 이미 작성된 문장을 옮기는 수준이므로, 상한 연장(3)이 합리적이라고 본다.

우선순위:

1. **N1 + N2 를 함께 고친다 (critical).** 이 SPEC 의 존재 이유(REQ-CSC-008: 은퇴 스킬이 영구 잔존하지 않게 한다)가 현재 규정대로 구현하면 성립하지 않고, MUST AC 두 개가 그 사실을 검출하지 못한다. REQ-008 에 `Lstat` 기준 판정 + dangling 제거를 명시하고, AC-008 의 fixture 를 링크 3형태 + 사용자 실 디렉터리로 확장한다.
2. **N10 을 고친다 (critical, 한 문단).** `plan.md` §F M3 이 REQ-CSC-010 과 정반대를 지시한다. 위험표(R12)는 이미 새 방향이므로 마일스톤 본문만 맞추면 된다.
3. **N3 · N12 · N11 · N4 (major 4건).** 출력 seam 의 실재를 REQ 수준에서 확정하고(N3), 백업 제외 로직에 소유 마일스톤을 준다(N12). M5 를 REQ-016 에 맞추고(N11), §H 판별자를 하나로 고정한다(N4).
4. **optional 4건(N5 · N6 · N8 · N9)** 은 오케스트레이터 재량이다. 넷 다 한 줄 수정이고, N5 · N6 은 §A·§D.4 가 "남의 틀린 전제를 정정한다"는 자세로 서 있는 문서라 자기 정확도가 곧 권위다 — 함께 처리하기를 권한다.

**v0.3.0 개정 자체는 잘 됐다.** iter-1 의 blocking 8건 중 spec·acceptance 계층 것은 전부 닫혔고, 특히 REQ-CSC-010 의 방향 반전(기록한다 → 기록하지 않는다)은 실측이 설계를 뒤집은 옳은 사례다. §G 에 기각 사유를 남긴 형태도 다음 감사의 중복 지적을 막는다. 이번 FAIL 의 무게중심은 **spec 이 앞서가고 plan 이 따라오지 못한 간격**(N10·N11·N12)과, **개정이 손대지 않은 청소 경로의 링크 형태**(N1·N2)다.

설계 방향(D1 정본 `.claude/skills/` · D2 역방향 기각 · D3 복사 폴백 · D4 파생 미러 집합 · D5 접두 한정)에는 이의가 없다. 다섯 결정 모두 근거가 실측 위에 서 있다.

---

## 감사 방법 기록

- 재현 코드는 worktree 안 임시 모듈(`.embedprobe/`, `.linkprobe/`)에서 실행하고 **전부 삭제**했다. 감사 종료 시 `git status --porcelain` → `?? .moai/reports/t81/`, `?? .moai/specs/SPEC-CODEX-SKILLS-CANONICAL-001/`, `?? tmp-audit-probe/`. 앞의 둘은 감사 이전과 동일하고, **`tmp-audit-probe/` 는 내가 만든 것이 아니다** — 같은 worktree 에서 병행 중인 SPEC 작성자 세션의 재현 디렉터리로 보이며 건드리지 않았다(정리는 그 세션 소관).
- SPEC 아티팩트는 **읽기만** 했다. 수정 0건. (감사 중 발생한 v0.2.0 → v0.3.0 변경은 작성자 측 개정이며 내 쓰기가 아니다.)
- `go test ./...`(전량)은 실행하지 않았다. 최소 재현 모듈과 소스 대조로 판정했다.
- `~/.codex/` 는 읽기만 했다(`find` · `ls`). 변경 0건.
