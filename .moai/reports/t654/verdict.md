# t654 — 히어독 본문의 git 문구를 실행 명령으로 판정하는 가드 오탐: 재현과 수리 선택지

- card: t654 (Class B) · worktree `.claude/worktrees/t654` · branch `WT-guard-quoted-span`
- base: 로컬 develop `00ae57ad7` (fast-forward)
- 상태: 재현 완료 → 운영자 결정 **B(절차 변경)** 적용. 가드 코드(`branch_guard.go`)는 바꾸지 않았다(§5). A1 은 후속 후보로 남긴다(§5.3).

## 1. 결론

카드의 두 사례는 **서로 다른 두 가드**다.

| 카드 사례 | 거부한 가드 | 이 저장소에서 고칠 수 있나 |
|---|---|---|
| (1) primary 체크아웃, `BRANCH_GUARD_VIOLATION: git merge` | `internal/hook/branch_guard.go` (MoAI) | 예 |
| (2) 워크트리, "워크트리 가드" | Claude Code 바이너리의 워크트리 격리 가드 — 거부 문구 `… names git in a form too complex to verify …` | 아니오 (`worktree-integration.md` § Refused Commands in a Worktree-Isolated Session) |

둘 다 실측으로 재현됐고, 둘 다 **본문을 파일에서 읽는 형태**(`moai handoff save --stdin < 파일`)에서는 통과한다. 전문: `probe-outputs.txt`.

## 2. 재현 (무해한 명령만)

모든 probe 는 `moai handoff save --help …` 이다. `--help` 는 사용법만 찍고 아무것도 저장하지 않는다. branch guard 쪽은 PreToolUse 판정 JSON 을 `moai hook pre-tool` 에 넣어 판정만 받았고, 명령은 실행되지 않았다. primary 체크아웃에서 git 명령을 실행한 일은 없다.

설치 바이너리: `v3.2.0-rc.7 … ged71054d3-dirty`. `branch_guard.go` 의 마지막 변경 `7d6f11687` 이 `ed71054d3` 의 조상임을 확인했다(`git merge-base --is-ancestor` exit 0). 두 설정(워크트리·primary) 모두 `workflow.branch_guard.enabled: true`.

### 2.1 Claude Code 워크트리 가드 (이 세션의 Bash 호출)

| probe | 명령 형태 | 결과 |
|---|---|---|
| W1 대조군 | 히어독, 본문에 git 문구 없음 | 통과, exit 0 |
| W2 | 히어독 본문에 `git merge --no-ff abc1234` | **거부** — `names git in a form too complex to verify that it stays inside the worktree` |
| W3 | 히어독 본문에 `git checkout -b feat/x` | **거부** — 같은 문구 |
| W4 | 같은 git 문구를 큰따옴표 `--body` 인자로 | 통과, exit 0 |
| W5 | 본문을 파일에서 읽음 `--stdin < body-with-git.txt` | 통과, exit 0 |

### 2.2 MoAI branch guard (`moai hook pre-tool`)

| probe | cwd | 명령 형태 | 판정 |
|---|---|---|---|
| P1 | primary | 히어독 본문에 `git merge --no-ff abc1234` | **deny** `BRANCH_GUARD_VIOLATION: git merge in primary checkout` |
| P5 | primary | 히어독 본문에 `git checkout -b feat/x` | **deny** `… git checkout <branch/-b> …` |
| P2 | primary | 같은 문구를 큰따옴표 인자로 | allow |
| P3 양성 대조군 | primary | 실제 `git merge --no-ff abc1234` | deny |
| P4 | 워크트리 | P1 과 같은 명령 | allow |

P3 이 deny 이므로 가드는 작동 중이고, P2·P4 로 인용 접기와 primary 판별이 각각 설계대로 도는 것도 확인된다. 오탐은 P1·P5 — 히어독 본문 — 에만 있다.

## 3. 문서의 스캔 범위와 대조

`main-checkout-branch-guard-detail.md` § Mechanical enforcement, **Scan scope**:

> the pattern set is matched against the command with quoted spans collapsed to a placeholder word, so a match reflects the command being invoked rather than text carried as data.

- **문자 그대로는 코드와 일치한다.** 접는 대상은 "quoted spans" 이고, 코드도 `'[^']*'|"[^"]*"` 만 접는다(`branch_guard.go:168`, `substituteQuotedArguments` `:197`). 히어독 본문은 따옴표 밖의 여러 줄이라 접히지 않고, `\bgit\s+merge\s` (`:146`) 같은 패턴이 문자열 어디든 걸린다. 여는 구분자 `'EOF'` 만 `X` 로 접힌다.
- **문서가 밝힌 목적과는 어긋난다.** 목적은 "text carried as data" 를 명령으로 오인하지 않는 것인데, 히어독 본문은 전형적인 데이터 운반 형태다. 특히 구분자를 따옴표로 감싼 `<<'EOF'` 본문은 bash 가 어떤 확장도 하지 않는 순수 텍스트다(`worktree-integration.md` § Why the two heredoc delimiter forms differ). 문서가 명시한 예외(`bash -c "git …"` 과소 매칭 수용)와 비교해도, 히어독 본문 오탐은 문서가 다루지 않은 빈칸이다.

## 4. 수리 선택지 (운영자 결정)

| 안 | 내용 | 막히는 사례 | 고칠 파일 | 대가 |
|---|---|---|---|---|
| **B (권장)** 호출 절차 변경 | 인계 본문을 Write 도구로 파일에 쓴 뒤 `moai handoff save --stdin < <파일>` 로 넘기도록 절차 문구를 정한다. 히어독 인라인 금지 | (1)·(2) 둘 다 해소 (W5 실측, branch guard 는 명령줄에 git 문구가 없으니 매칭 불가) | `session-handoff.md` § Emission-Time Save Obligation, `output-styles/moai/moai.md:685` 의 저장 문구 — 로컬·템플릿 사본 각 2벌 | 가드 완화 없음. 절차를 안 지키면 재발 |
| **A1** 따옴표 구분자 히어독 본문 접기 | `<<'WORD'`·`<<"WORD"` 본문을 인용 접기처럼 자리표시어로 바꾼 뒤 매칭 | (1)만 해소. (2)는 Claude Code 가드라 그대로 | `internal/hook/branch_guard.go` + 테스트, 상세 문서 Scan scope 절 | **가드 완화.** `bash <<'EOF'` 처럼 본문을 셸이 실행하는 경우도 접혀 과소 매칭 — 문서가 이미 수용한 `bash -c "…"` 과 같은 계열이지만, 소비자가 셸일 때는 접지 않는 예외를 둘지 추가 결정 필요 |
| **A2** 모든 히어독 본문 접기 | 구분자 인용 여부와 무관하게 접기 | (1)만 | 위와 같음 | A1 보다 넓은 완화: 인용 없는 `<<EOF` 본문은 명령 치환이 실제로 실행될 수 있다 |
| **C** B + A1 | 절차로 두 가드를 모두 피하고, 가드도 데이터 형태를 인정 | (1)·(2) | 위 둘 | A1 의 대가 그대로 |

(2)는 이 저장소에서 가드를 고칠 방법이 없으므로, **B 없이는 워크트리 세션의 인계 저장이 계속 막힌다.** A 는 primary 체크아웃에서의 (1)만 추가로 푼다.

부수 기록: `worktree-integration.md` 의 "What has been observed to trip the worktree guard" 표에는 JSON 중괄호 본문과 복합 명령만 있다. **git 명령 문구가 든 히어독 본문**은 이번에 새로 1차 관측된 트리거다(W2·W3, 대조군 W1). 표에 추가할지는 B 와 함께 결정하면 된다.

## 5. 운영자 결정 B — 적용

### 5.1 수정 (로컬·템플릿 각 3파일, 수정 전 두 사본은 `cmp` 로 바이트 동일 → 로컬 수정 후 템플릿으로 복사, 복사 후 다시 `cmp` 동일)

| 파일 | 변경 |
|---|---|
| `.claude/rules/moai/workflow/session-handoff.md` § Emission-Time Save Obligation (`:27`) | "pipe the block to `moai handoff save --stdin …`" → Write 도구로 파일에 쓰고 `moai handoff save --stdin … < <file>` 로 넘긴다. 인라인 히어독 금지와 그 이유(본문의 git 명령 문구를 가드가 git 실행으로 읽어 branch guard 는 primary 에서 거부, Claude Code 워크트리 가드는 워크트리에서 거부 → fail-open 규칙으로 저장이 조용히 건너뛰어짐) |
| `.claude/output-styles/moai/moai.md` 저장 의무 요약(`:685`) | 같은 절차로 맞춤 — `session-handoff.md` 의 SSOT↔렌더 표면 동기화 규칙(Drift-mitigation sentinel)을 따름 |
| `.claude/rules/moai/workflow/worktree-integration.md` 관측 트리거 표 | "git 하위 명령을 이름으로 담은 히어독 본문" 1차 관측 1행 추가(W1~W5 요약), 표 앞뒤 문장의 "two observations / both rows" 를 행 수에 맞게 수정 |

- 템플릿 금지 표지 점검: 추가된 줄에 카드 번호·SPEC ID·날짜·개인 경로 없음(`git diff -U0 | grep -E "^\+[^+]" | grep -i -E "t[0-9]{3}|SPEC-|2026|/Users/|goos|lane"` → 적중 없음).
- 크기(`git cat-file -s HEAD:<f>` 대 `wc -c <f>`): `session-handoff.md` 21197→21566 (+369B, 항상 로드), `moai.md` 66611→66773 (+162B, 항상 로드), `worktree-integration.md` 48555→49048 (+493B, `paths:` 범위). 항상 로드되는 두 파일 모두 1,000B 미만이라 `rule-authoring.md` §statement duty (b) 대상 아님.
- `internal/template/catalog.yaml` 에는 세 파일 항목이 없어 재생성할 해시 없음.

### 5.2 검증

```
go test ./internal/template/... -count=1        exit=0   (template-tests.txt)
ok  	github.com/modu-ai/moai-adk/internal/template	55.376s
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	1.502s
ok  	github.com/modu-ai/moai-adk/internal/template/commandemit	0.931s
```

`-run` 선택자 없이 패키지 전체를 돌렸으므로 `rule_template_mirror_test.go`, `template_neutrality_audit_test.go`, `internal_content_leak_test.go`, `catalog_slim_audit_test.go` 가 모두 실행 범위에 든다(리드 확인: 이 패키지는 `internal/cli` 를 컴파일하지 않아 슬롯 불필요).

### 5.3 후속 후보 — A1

branch guard 가 따옴표 구분자 히어독(`<<'WORD'`) 본문도 인용 접기처럼 다루게 하는 안(A1)은 이번 결정에서 제외했다. B 는 절차를 따르는 세션에서만 오탐을 피하므로, 절차를 벗어난 히어독은 여전히 P1·P5 처럼 거부된다. A1 을 진행한다면 본문을 셸이 실행하는 경우(`bash <<'EOF'` 등)를 접지 않을 예외를 함께 설계해야 한다.

## Gaps

- 카드의 원 사례 두 건(lane-1, lane-5)의 실제 거부 문구·명령 원문은 보지 못했다. lane-5 의 "워크트리 가드"가 Claude Code 가드라는 판단은 같은 형태를 이 세션에서 재현한 결과와 저장소에 해당 문구의 구현이 없다는 사실(`worktree-integration.md` 표)에 근거한다.
- 워크트리 가드의 판정 경계는 재지 않았다. "본문에 `git` + 하위 명령" 이 트리거라는 것까지는 W1 대 W2·W3 로 보였지만, `git` 단어만 있는 경우나 `git status` 같은 읽기 명령 문구는 시험하지 않았다.
- branch guard 판정은 설치 바이너리(rc.7, `ed71054d3` 기반)로 쟀다. 소스(`00ae57ad7`)의 `branch_guard.go` 는 `7d6f11687` 이후 변경이 없어 같은 코드다.
- `moai hook pre-tool` 실행이 다른 PreToolUse 처리기를 통해 로그를 남겼는지는 확인하지 않았다(`.moai/logs/` 는 git 추적 밖).

## Residual-risk

- B 는 절차 문구라서, 세션이 히어독을 쓰면 오탐이 그대로 재발한다. 강제 장치는 아니다.
