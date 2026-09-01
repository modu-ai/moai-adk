# SPEC-CODEX-SKILL-NEUTRAL-001 — 진행 기록

카드: t196 · 워크트리: `.claude/worktrees/t196` · 브랜치: `WT-codex-skill-neutral`

## §E.1 Plan-phase Audit-Ready Signal

- 작성: manager-spec (t196-spec), plan-phase **iter-2 최종, v0.3.0** (iter-1 v0.1.0 → iter-2 v0.2.0 → 감사 부채 마감 v0.3.0)
- 기준 트리 HEAD: `297a21ea73b24e6605280625e576555e4316263e`
- **iter-2 판정: PASS-WITH-DEBT 0.800** (Tier M 임계 0.80). 본문 0.825 → addendum 이 D13 승격 + D16 추가로 0.800. MUST-PASS 7/7 전면 재실행 통과. 점수 회귀 없음(0.75 → 0.800)이라 STOP 절 미발동. 전문: `.moai/reports/t196/plan-audit-iter2.md` + `plan-audit-iter2-addendum.md`
- iter-2 최종에서 닫은 blocking 5건: D11(AC-CSN-012 파일 범위를 무가드 4개로 확장) · D13 승격(양성 대조 수치 제거) · D16(c)(자기참조 수치 규율을 §D.4 + AP-14 에 고정) · D12(`plan.md` §A 측정/추론 분리) · D16(b)(보고서 §4 워킹트리 글롭 제거)
- iter-2 최종 채택 optional: D14(실패 경로 매치 텍스트 기록) · D15(`:386` 기대 분류 표 + M1 `기록으로 닫음` 주석) · `:219` 보조 문구(권장, 의무 아님으로 M3 에 기재)
- **REQ/AC 개수 불변**: REQ 15 / AC 13 (Tier M 상한 16/16). 전부 기존 항목의 범위·문면 수정
- iter-1 판정: **FAIL 0.75** (Tier M 임계 0.80). MUST-PASS 7/7 통과, 결함은 판정 계층 집중. 전문: `.moai/reports/t196/plan-audit-iter1.md` + `_addendum.md`
- iter-2 에서 닫은 blocking 7건: D10(규칙 트리 4줄 범위 편입) · D1(무가드 REQ + 거짓 매핑) · D3(닫힘 조건 vs 부채 모순) · D4(`<base>` 자리표시자) · D2(§A.7 근거 오귀속 + REQ-CSN-009 오조준) · D5(실측 표제 아래 추론) · D6(반증 시 설계 재개 지시 부재)
- iter-2 채택 optional 5건: D7(≤ 한정자) · D8(`Where`→`While`, 복합 요구 해소) · D9(포섭 명시) · (1c)(passthrough 등록 안티패턴) · dogfood census pre/post 쌍
- 요구사항 14 → **15**, 판정 10 → **13** (Tier M 상한 16/16)
- 편집 대상 파일 수 12 → **14** (규칙 트리 2파일 편입), Tier M 범위(5–15) 유지
- Tier: **M** (카드 등록값 `M~L` 에서 재도출 — 근거 spec.md §E.1, 권장안 채택 시 **14파일**)
- 산출물: `spec.md` · `plan.md` · `acceptance.md` · `progress.md` (Tier M 3파일 + progress)
- 재측정 전문: `.moai/reports/t196/premise-remeasure.md`
- 카드 전제 판정: ① 반증(21→34) · ② 반증(9→14) · ③ 단위 불일치(3 vs 4스킬·9파일·46줄) · ④ 확인(11/11)
- 카드가 놓친 구조 정정: `.codex/skills/` 부재 — 스킬 축 편집은 Claude 쪽 정본 편집이다
- **[iter-2 정정] 리드의 대조 트리 `main 48239c7dc` 는 방증 자격이 없다 — 리드가 값을 철회했다.** 그 트리는 `origin/develop` 보다 **686커밋 뒤진 엄격한 조상**이다(`git rev-list --count --left-right origin/develop...48239c7dc` → `686 0`). 종전 판본은 "두 값은 각각 자기 트리에서 참"이라고 적었는데 형식상 맞지만 두 값이 대등한 것처럼 읽혀 오해를 부른다 — 한쪽은 배포선이 아니다.
  - `SPEC-CODEX-*` 수(리드 7 / 이 트리 9): **실제로 갈린 유일한 자리이며 9 가 정본이다.** ref 판독 — `git ls-tree -d --name-only origin/develop -- .moai/specs/ | grep -c CODEX` → `9`, 같은 명령을 `48239c7dc` 에 걸면 `7`. 7 은 낡은 지점의 값이므로 채택하지 않는다.
  - HARD 자리 수(리드 4 파일위치 / 본 SPEC 6 줄): **단위 차이이지 트리 차이가 아니다**(파일 위치 3 / 줄 6 / 스크립트 대상 3). 판정 단위는 줄 6 이며, 리드의 4 는 철회된 트리에서 온 값이라 대조 대상으로 세우지 않는다.
  - ①(34)·②(14)·③(3 vs 9)은 **두 ref 에서 값이 같다** — 본 SPEC 이 `git ls-tree` / `git grep <ref>` 로 각각 직접 읽어 확인했다. 이 대조들은 리드의 철회된 측정이 아니라 본 SPEC 의 ref 판독에 근거한다. 특히 ③ 에 트리 차이가 섞였을 가능성은 반증됐다(두 단위 모두 두 ref 에서 3/3, 9/9).
  - 규율: 소스 사실은 체크아웃 grep 이 아니라 **ref 에 대고** 읽는다(`git ls-tree <ref>` / `git grep <pattern> <ref> -- <paths>`) — 어느 트리를 잰 것인지가 명령에 박혀 라벨을 잃을 수 없다. 이번 철회가 라벨을 잃은 사고였다.
- 미해결 의존: 없음 (t88 종결, `SPEC-CODEX-*` 9건 전부 `completed`)
- 관측하지 않은 것: 코덱스 런타임 거동(REQ-CSN-001 이 run-phase 의무로 건다) · 직접 실행 세션의 cwd(부채, acceptance.md §D.3(2))
- **선언된 부채 3건** — 통과로 위장하지 않는다: (1) 결속표의 실효성은 이 SPEC 이 판정하지 않는다 · (2) `moai codex` 를 거치지 않은 세션의 cwd 는 묶이지 않는다 · (3) Template-First 편집 **순서**는 최종 상태로 판정 불가. 전문 acceptance.md §D.3
- iter-2 가 철회한 부채 1건: 종전 "복사 폴백 모드 경로 해석"은 부채가 아니라 **잘못 겨눈 요구사항**이었다(루트 기준 경로는 미러에 닿지 않아 어떤 변이로도 깨지지 않는다). REQ-CSN-009 를 cwd 팔로 재조준했다

## §E.2 Run-phase Evidence

**축 B 한정 실행** — 축 A(REQ-CSN-002~005, AC-CSN-001~005)는 리드 지시로 동결이며 `AGENTS.md` 두 사본에 편집 0바이트다. 아래 판정은 축 B(M3)와 교차 관심사(M4)만 다룬다.

### 기준 이동 — AC-CSN-010 의 못박은 SHA 를 바꿨다 (조용히 교체하지 않는다)

plan-phase 기준 트리는 `297a21ea73b24e6605280625e576555e4316263e` 였다. run-phase 착수 시점 워크트리 HEAD 는 **`2c18091d127cbc723074124e1015353e077300ca`** (리드의 fast-forward)다. AC-CSN-010 의 [HARD] 절이 지시한 대로 착수 시점 HEAD 를 재측정해 판정 기준으로 삼았고, 바꿨다는 사실을 여기 남긴다. `acceptance.md` 의 AC-CSN-010 본문 SHA 는 **아직 옛 값**이며 sync-phase 가 갱신할 대상이다.

`plan.md` §C 의 재측정 의무는 착수 시점 HEAD 에서 직접 실행했다 — 값이 전부 plan-phase 값과 일치한다(아래 사전값 표).

### 착수 전/후 census (전부 `2c18091d1` 워크트리에서 실측)

| 측정 | 명령 | 사전 | 사후 |
|---|---|---|---|
| 스킬 트리 전수 | `grep -rn 'CLAUDE_SKILL_DIR' internal/template/templates/.claude/skills \| wc -l` | **46** | **0** |
| 스킬 트리 HARD(실행 인자) | 위 + `\| grep -E ':(bash\|node) ' \| wc -l` | **6** | **0** |
| 스킬 트리 보유 파일 | `grep -rl … \| wc -l` | **9** | **0** |
| 템플릿 트리 전체 | `grep -rn … internal/template/templates/ \| wc -l` | **50** | **1** |
| 규칙 트리(템플릿) | `grep -rn … internal/template/templates/.claude/rules` | **4줄 / 2파일** | **1줄 / 1파일** |
| 로컬 dogfood | `grep -rn 'CLAUDE_SKILL_DIR' .claude/skills \| wc -l` | **46** | **46** |

**dogfood 와 템플릿이 어긋났다 — 어긋난 사실과 두 수를 그대로 보고한다.** 템플릿 0 / 로컬 46. 귀속은 사람의 몫이나, 가장 그럴듯한 설명은 로컬 `.claude/skills/` 가 **설치된 바이너리에서** 재배포되는 파생물이고 그 바이너리가 이 편집을 아직 담고 있지 않다는 것이다(`moai update` 미실행). M3 의 [HARD] 절대로 가드는 두 트리로 넓히지 **않았다** — 넓히면 바이너리 지연이 "토큰이 돌아왔다"는 거짓 적색을 낸다.

### 복구 사건 기록 — `moai/SKILL.md` 유실과 바이트 단위 복구

리드가 AC-CSN-009 의 RED 를 직접 확인하려 `moai/SKILL.md` 끝에 프로브 한 줄을 붙였다가, 지울 때 `git restore <path>` 를 써서 **미커밋 상태이던 이 파일이 HEAD 로 되돌아갔다**. 프로브 한 줄이 아니라 이 카드의 수정 19줄이 함께 사라졌다. 다른 8개 스킬 파일·규칙 트리 2파일·`catalog.yaml`·가드 테스트는 무사했다(실측: `git status --short`).

복구는 재작업이 아니라 **동일 상태 복원임이 독립 증명됐다**:

```
복구 후 sha256 : b2f9178fce43d9ca04c0d0811b943e9534300346404b505bc234175afa74f87c
유실 전 sha256 : b2f9178fce43d9ca04c0d0811b943e9534300346404b505bc234175afa74f87c
                 (m4-guard-red-planted.log 의 before=/after= 행 — 유실 이전에 기록된 값)
```

두 값이 같으므로 복구본은 "토큰이 없다"가 아니라 **유실 이전 파일과 바이트 동일**하다. `catalog.yaml` 의 `moai` 스킬 해시도 같은 값으로 재생성됐다(`make build`).

**리드 기억과 한 자리 달랐다 — 파일을 읽고 판단했다.** 19줄 중 18줄은 `Read ${'{'}CLAUDE_SKILL_DIR{'}'}/workflows/<name>.md` 형태였으나, `:273` 은 산문 문장 안의 백틱 인라인 참조였다. 형태는 같은 규칙으로 치환했다.

### AC 판정 매트릭스

| AC | 판정 | 명령 / 관측 | 증거 |
|---|---|---|---|
| AC-CSN-001 | **미실행** (축 A 동결) | — | 리드가 M1 관측을 이미 수행 — `m1-probe-{a,a2,b}.log` |
| AC-CSN-002 | **미실행** (축 A 동결) | — | `AGENTS.md` 편집 0 |
| AC-CSN-003 | **미실행** (축 A 동결) | — | 동일 |
| AC-CSN-004 | **미실행** (축 A 동결) | — | 동일 |
| AC-CSN-005 | **미실행** (축 A 동결) | — | 동일 |
| AC-CSN-006 | **PASS** | HARD 파이프라인 계수 | 사전 **6** → 사후 **0**. 사후 초록은 AC-CSN-008 에 포섭되므로 독립 증거가 아니며, **이 AC 의 내용은 위 쌍**이다 |
| AC-CSN-007 | **PASS** | AC-CSN-008 과 함께 판정 | 산문 자리 사전 **40**(46−6) → 사후 **0**. 시끄러운 자리만 닫지 않았다 |
| AC-CSN-008 | **PASS** | `grep -rn … .claude/skills \| wc -l` | 사전 **46** → 사후 **0** |
| AC-CSN-009 | **PASS** | 독립 관측 **2회**. ① 이 세션: 심기 전 census 0 → 심은 뒤 1 → 가드 exit **1** → 되돌린 뒤 census 0 → exit **0**, 대상 파일 sha256 심기 전 == 되돌린 후. ② 리드가 별도 프로브로 재현(census 1, `moai/SKILL.md:393`) — 서로 다른 행위자가 같은 가드의 RED 를 봤다 | `m4-guard-red-initial.log`(census 46, 최초 RED) · `m4-guard-red-planted.log`(census 0→1) · `m4-guard-green-uncached.log`(`-count=1`) · `m4-guard-green-post-recovery.log`(복구 후 census 0) |
| AC-CSN-010 | **PASS** | 아래 [주의] 참조 | `git diff --stat 2c18091d1..HEAD -- <3표면>` 빈 출력 **+** `git status --short -- <3표면>` 빈 출력 |
| AC-CSN-011 | **PASS** | `grep -rn … .claude/rules` | 잔존 **1줄**, 아래 집합 표 |
| AC-CSN-012 | **PASS** | 4파일 `grep -cE` 사전/사후 쌍 + 양성 대조 | `m3-ac012-post.log` |
| AC-CSN-013 | **PASS** | §A.7 문면 3항목 + 스윕 범위 기록 | `m3-cwd-sweep.log` |

[주의] **AC-CSN-010 의 명령만으로는 공허한 초록이다.** 이 run-phase 는 커밋을 만들지 않았으므로 `HEAD == 2c18091d1` 이고, `<base>..HEAD` 는 **어떤 변경도 담을 수 없어 무조건 빈 출력**이다. 실제로 판정을 지는 것은 병기한 `git status --short` 쪽이며, 그것이 세 표면 모두 미변경임을 보인다. 커밋이 생기면 diff 팔이 비로소 힘을 갖는다 — sync-phase 가 재판정할 것.

### AC-CSN-011 — 규칙 트리 잔존 줄 집합 (개수가 아니라 내용으로 판정)

| 줄 | 분류 | 처분 |
|---|---|---|
| `skill-authoring.md:219` | **사실 기술** | **잔존** — 능력 표의 한 행. Claude Code 가 이 변수를 제공한다는 서술은 참이다. 한정자를 덧붙였다(plan 의 권장, 의무 아님): 다른 하네스에서는 참조가 빈 값으로 전개된다는 사실 |
| `skill-authoring.md:226` | 규범 문장(사전) | **제거** — 토큰이 사라졌다. 채택 설계(루트 기준 상대 경로)를 지시하는 문장으로 교체 |
| `skill-authoring.md:301` | 규범 문장(사전) | **제거** — 동일 |
| `worktree-integration.md:386` | 예시(사전) | **수정** — 토큰이 사라졌다. 표 행이 이제 `.claude/skills/<name>/<file>` 형태를 예시하며, "절대 경로 OK?" 열이 `YES` → `NO — use project-root-relative` 로 뒤집혔다 |

**사후 집합 = {`skill-authoring.md:219`} 이며 규범 문장은 0건이다.** 개수(4→1)는 근거로 쓰지 않는다.

### 치환 경로가 실제로 해석된다는 관측 (전수 0 이 잡지 못하는 뮤턴트)

`grep` 이 0 을 내는 것은 **깨진 경로로도 만족된다.** 그래서 두 층으로 확인했다.

1. **전수 존재 검사** — 편집된 9파일에서 추출한 `.claude/skills/...` 경로 **34개 전부**가 배포 트리와 템플릿 트리 **양쪽에** 실재한다. 미해결 2건은 치환 결과가 아니라 기존 산문의 자리표시자 조각(`.claude/skills/hns-<name>*/`, `.claude/skills/moai-*`)이며, 착수 전 트리에도 같은 형태로 있었다. 증거 `m3-path-resolution.log`
2. **실행 관측(HARD 자리)** — `SKILL.md:235` 의 치환된 명령을 축자 그대로 돌렸다:
   - **대조(빈 전개가 만드는 경로)**: `node /scripts/check-svg.mjs …` → 모듈 로더가 `throw err`, exit 1. 스크립트를 찾지 못했다.
   - **치환 경로**: `node .claude/skills/moai-domain-svg-infographic/scripts/check-svg.mjs …` → 스크립트가 **로드되어 실행됐고** SVG 린트 진단 5건을 출력했다. exit 1 은 린터가 내 최소 probe SVG 에 내린 판정이지 경로 해석 실패가 아니다(대조는 진단을 한 줄도 못 냈다).
   - 증거 `m3-hard-site-execution.log`

### REQ-CSN-009 — cwd 전제를 어디까지 훑었는가 (지지/미지지)

"없음"이라고 적지 않는다 — 훑은 표면을 적는다. 증거 `m3-cwd-sweep.log`.

| 표면 | 전제 지지 | 근거 |
|---|---|---|
| `moai codex` 런처 | **지지 — 단, 강등 분기 있음** | `internal/cli/codex_launcher.go:242-244`(의도 주석), `:245-249`(루트 해석), `:248-249`(루트 해석 실패 시 프로세스 cwd 로 강등), `:252`·`:271`(`Dir` 전달). 착수 시점 트리에서 줄 번호 재확인 — spec.md §A.7 의 `:245-250` 인용이 이 트리에서도 같은 코드를 가리킨다 |
| `moai codex` 를 거치지 않은 직접 코덱스 세션 | **미지지 — 부채** | 저장소 전체에서 cwd 를 강제하는 장치를 찾지 못했다(스윕 S4). 사용자 cwd 를 그대로 물려받는다. 이 SPEC 은 이 팔을 **닫지 않는다** |
| CLI 패키지의 다른 `.Dir` 대입 | **무관** | 스윕 S3 이 열거한 비테스트 3파일 중 `launcher.go:602` 는 git 헬퍼(`runGitCommand`), `worktree/guard.go` 는 워크트리 가드, `doctor_agentemit_embed.go` 는 진단이다 — 어느 것도 스킬 본문을 읽는 프로세스를 띄우지 않는다 |
| 이 세션 자신(워크트리 안 Claude) | **지지 — 실증** | cwd `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t196` 에서 `.claude/skills` 가 해석된다(스윕 S5). 워크트리 루트가 프로젝트 루트 역할을 하므로 루트 기준 형태가 성립한다 |
| Claude Code Bash 도구의 환경 변수 | **강제 없음 — [리드 보강 실측 2026-09-01]** | `printenv CLAUDE_PROJECT_DIR` → exit 1 (**미설정**). `navigator-regen.sh` 의 `ROOT="${CLAUDE_PROJECT_DIR:-$PWD}"` 폴백은 이 하네스에서 `$PWD` 로 떨어진다 — Bash 도구 cwd 는 세션 중 지속되므로 이전 호출의 `cd` 가 남아 있으면 스크립트가 잘못된 뿌리에 쓴다(읽기 실패가 아니라 mkdir -p 로 조용한 오기록). 세션 시작 cwd=프로젝트 루트(위 행)가 전제를 세우지만 호출별 강제는 없다 |

- 런 에이전트 기록 귀속: 위 §E.2 본문은 t196-run-b 가 작성했다. 에이전트는 완료 보고 없이 두 번 유휴 전환했고 세 번째 전환은 세션 사용량 한도로 끝났다 — 디스크 산출물로 회수했고 리드가 핵심 주장(미러 등재 :57, RED 46203/46296, 증거 파일 존재, `moai/SKILL.md` 복구 sha256)을 전부 독립 재측정해 채택했다.

AC-CSN-013 의 (a)(b)(c) 세 항목은 spec.md §A.7 `:138`·`:149` 에 문면으로 존재한다(스윕 S6).

### 발견 — 범위 밖이라 고치지 않고 보고하는 잔여 1건

저장소 **로컬** `.claude/rules/moai/development/skill-authoring.md` 의 `:226`·`:301` 에 뒤집힌 규범 문장이 **그대로 남아 있다**. 이것은 템플릿 사본이 아니고(그 둘은 이미 서로 다르며 바이트 패리티 등록 대상도 아니다) REQ-CSN-015 가 말하는 "배포되는 규칙 트리"도 아니므로, spec.md §A.8 의 grep 범위와 §E.1 의 파일 수(2), AC-CSN-012 의 파일 목록 전부가 템플릿 쪽만 덮는다. 범위를 임의로 넓히지 않고 보고한다 — **이 저장소의 저자는 여전히 뒤집힌 규칙을 읽는다.** 별도 카드 또는 sync-phase 판단 대상.

`worktree-integration.md` 는 사정이 다르다: 바이트 패리티 등록 파일(`rule_template_mirror_test.go` 의 `workflowOptMirroredPaths`)이고 두 사본이 착수 시점에 동일했으므로, 미러 의무에 따라 **양쪽을 동일하게** 고쳤다. `cmp` 로 동일성 재확인.

### [BLOCKER] 로컬 규칙 미러 되돌림이 바이트 패리티 테스트를 붉게 만들었다

리드가 `.claude/rules/moai/workflow/worktree-integration.md` **로컬 사본의 수정을 되돌렸다.** 사유는 "REQ-CSN-012 상 템플릿이 원본이고 로컬은 `make build` + `moai update` 로 파생된다"였다. **그 전제가 이 파일에는 성립하지 않는다.**

이 파일은 `internal/template/rule_template_mirror_test.go` 의 `workflowOptMirroredPaths` 에 등재된 **바이트 패리티 대상**이다. 파생물이 아니라 두 트리를 같은 커밋에서 함께 고쳐야 하는 SSOT 미러이며, 테스트 자신이 실패 메시지로 `cp` 를 지시한다. 되돌림 직후 실측:

```
$ cmp .claude/rules/.../worktree-integration.md internal/template/templates/.claude/rules/.../worktree-integration.md
… differ: char 34626, line 386          (되돌리기 전에는 IDENTICAL 이었다)

$ go test -count=1 ./internal/template/ -run TestRuleTemplateMirrorDrift
--- FAIL: TestRuleTemplateMirrorDrift/worktree-integration.md
    RULE_TEMPLATE_MIRROR_DRIFT: source file … differs from its mirror at …
    (source 46203 bytes, mirror 46296 bytes); run 'cp <source> <mirror>' and stage both files
```

`internal/template` 스위트 전체에서 **실패는 이 하나뿐**이다(`m3-template-suite.log`). 지시를 임의로 뒤집지 않고 되돌린 상태 그대로 두었으므로, 지금 트리는 이 테스트가 붉은 채다.

선택지는 둘이고 한쪽만 SPEC 과 양립한다:

| 안 | 결과 | SPEC 적합성 |
|---|---|---|
| **A. 로컬 사본에 같은 한 줄을 다시 적용** | 테스트 GREEN | REQ-CSN-015·AC-CSN-011 충족. 미러 의무와도 일치 |
| B. 템플릿 쪽 `:386` 수정을 철회 | 테스트 GREEN | **AC-CSN-011 FAIL** — `[HARD]` 절이 `:386` 수정을 지시하며 원문 그대로면 FAIL 이라고 못박았다 |

**A 를 권고하며, 리드 판단 대기 중이다.** `.claude/skills/**` 로컬 사본은 지시대로 손대지 않았다 — 그쪽은 실제로 바이너리에서 재배포되는 파생물이라 리드의 전제가 참이다. 두 트리를 가르는 것은 "로컬이냐"가 아니라 **바이트 패리티 등재 여부**다.

**[해소 — 리드 판정 A안 채택, 2026-09-01]** 리드가 세 단계로 독립 재측정한 뒤 채택했다: (1) 미러 등재 확인 — `rule_template_mirror_test.go:57` 에 `.claude/rules/moai/workflow/worktree-integration.md` 존재, (2) RED 재현 — `TestRuleTemplateMirrorDrift` FAIL, source 46,203 / mirror 46,296 바이트, (3) 테스트 메시지가 제안하는 `cp` 방향(로컬→템플릿)은 **AC-CSN-011 [HARD] 가 요구한 편집을 지우는 방향**이라 따르지 않고 템플릿→로컬로 정렬했다 — 방향은 AC가 정하고 테스트는 동일성만 본다. 결과: `cmp` 바이트 동일, `TestRuleTemplateMirrorDrift` GREEN, `go test -count=1 ./internal/template/` 전체 `ok 24.159s`. 커밋 `816acd104`. 사고 귀속: 되돌림은 리드(이전 턴의 정책 획일 적용 오류) — 해소도 리드, A안 권고는 런 에이전트. 교훈: 축은 "로컬 vs 템플릿"이 아니라 **바이트 패리티 등재 여부**이며, 등재 파일은 같은 커밋에서 함께 고친다.

### 검증 실행 (범위: 변경이 닿을 수 있는 패키지)

| 명령 | 결과 | 증거 |
|---|---|---|
| `go vet ./internal/template/ ./internal/config/` | exit **0** | `m3-govet.log` |
| `go test -count=1 ./internal/template/` | **FAIL** 24.216s — 실패 1건, 전부 위 BLOCKER(미러 드리프트). 다른 서브테스트 전원 통과 | `m3-template-suite.log` |
| `go test -count=1 ./internal/config/` | **ok** 1.741s | `m3-config-suite.log` |
| `make build` | exit **0** (복구 후 재실행 — `catalog.yaml` 해시 3건 재생성) | `m3-make-build.log` · `m3-make-build-recovery.log` |

`go test ./...` 는 **돌리지 않았다** — 병렬 레인이 그것을 돌려 머신을 마비시킨 사고가 있고, 전 패키지 판정은 CI 몫이다.

`make build` 의 `gen-catalog-hashes --all` 이 `internal/template/catalog.yaml` 의 스킬 해시 3건을 재생성했다(편집한 `SKILL.md` 3종). 같은 SPEC 범위 안의 cascade 이며 별도 결정이 아니다.

### 커밋 기록 (리드가 run 종료 후 착지 — 리드의 안전망 지시에 따라)

run 에이전트는 커밋 없이 종료했고, 리드가 세 커밋으로 착지했다(커밋 직전마다 HEAD·브랜치 재판독, 명시 pathspec 스테이징 + 같은 호출 status 재판독):

| 커밋 | 내용 | 규모 |
|---|---|---|
| `25422be77` | 축 B 구현 — 템플릿 스킬 9 + 템플릿 규칙 2 + `catalog.yaml` + 신규 가드 테스트 | 13파일 +156/−53 |
| `bb14186d1` | SPEC 계획 산출물 v0.3.0 (spec/plan/acceptance/progress) | 4파일 +958 |
| `816acd104` | 로컬 미러 패리티 복원 (위 BLOCKER A안) | 1파일 +1/−1 |

브랜치 `WT-codex-skill-neutral`, base `2c18091d1`. `draft → in-progress` 전이와 AC-CSN-010 의 sync 재판정은 sync 쪽 소관 그대로다. `.moai/reports/t196/` 증거 디렉터는 progress 갱신 커밋에 함께 스테이징한다.

### 셀렉터 매치 수 (리드 관측 — AC 측 명기는 manager-spec 배치로 이월)

`go test -run TestSkillDirToken` 은 **0매치** — `no tests to run` 문구와 함께 PASS 를 돌려준다(0매치 셀렉터의 초록은 판정이 아니다). 실제 테스트 이름은 `TestSkillTreeHasNoClaudeSkillDirToken` 이며 **1매치**다. 위 `go test -count=1 ./internal/template/` 전체 스위트(ok 24.159s)에는 이 테스트 1건이 실제로 포함돼 돈다. AC 측 기록(셀렉터가 실제로 몇 개를 잡는지 명기)은 REQ-CSN-003 문면 정정 배치와 함께 manager-spec 에 이월한다.

### 런 에이전트 종료 상태

완료 보고 없이 유휴 전환 2회, 세 번째 전환은 세션 사용량 한도로 종결. 위 §E.2 의 모든 판정은 에이전트의 디스크 산출물을 리드가 독립 재측정해 채택한 것으로, "에이전트가 했다고 함" 수준의 인용은 없다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-09-01"
run_commit_sha: "25422be77 (축 B 본체, 리드 착지) · 816acd104 (미러 복원) — base 2c18091d1"
run_status: partial-axis-b-only
axis_a_status: unfrozen-lead-verdict-2026-09-01   # 어휘 축 판정(11번째 클래스 신설) + 예산 축 판정(1번 채택)으로 동결 해제됨; 구현은 manager-spec 재위임(REQ-CSN-003 문면 정정) 후 예정
ac_pass_count: 8      # AC-CSN-006..013
ac_fail_count: 0
ac_not_run_count: 5   # AC-CSN-001..005 (축 A — 동결 해제됐으나 미실행, manager-spec 문면 정정 선행)
preserve_list_post_run_count: 1   # 로컬 .claude/rules/.../skill-authoring.md:226,:301 — 범위 밖, 보고만
l44_pre_commit_fetch: n/a          # 커밋은 리드가 착지 — 커밋 전 HEAD 재판독 수행
l44_post_push_fetch: n/a           # 푸시 없음 (develop 통합은 창 소관)
new_warnings_or_lints_introduced: 0   # go vet exit 0
blocking_test_failure: 0   # TestRuleTemplateMirrorDrift — 816acd104 로 해소, 스위트 ok 24.159s
cross_platform_build:
  darwin: pass        # make build exit 0 (host darwin/arm64)
  windows: not-observed
  linux: not-observed
total_run_phase_files: 14   # 편집 13 (템플릿 스킬 9 + 템플릿 규칙 2 + 로컬 미러 1 + catalog.yaml 1) + 신규 가드 테스트 1
m1_to_mN_commit_strategy: "리드가 run 종료 후 3커밋 착지 (25422be77 / bb14186d1 / 816acd104)"
evidence_dir: .moai/reports/t196/
```

**감사가 먼저 볼 것 4가지**: (1) ~~미해결 BLOCKER~~ → **해소됨** — `TestRuleTemplateMirrorDrift` 는 `816acd104` 로 GREEN 복귀, 전체 스위트 ok 24.159s(위 해소 절 참조). (2) AC-CSN-010 의 diff 팔이 커밋 부재로 공허하다는 위 [주의] — 커밋이 이제 존재하므로 sync-phase 재판정 시 diff 팔이 실질 판정으로 전환된다. (3) dogfood census 46 vs 템플릿 0 의 어긋남 — 바이너리 지연 가설이며 확정 아님. (4) 로컬 `skill-authoring.md` 규범 문장 잔존 — 의도적 범위 밖.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
