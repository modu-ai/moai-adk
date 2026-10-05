# t642 — deny 규칙의 git 전역 옵션 앞붙임 통과: 판정

card: t642 · branch: WT-deny-git-prefix · base: 81c1d58f9 (= origin/develop, 2026-09-11 fetch 확인)
측정 세션: bypassPermissions 모드, Claude Code 2.1.268
운영자 결정(레인 세션 직접 확인, 2026-09-11): 수리 방향 = 와일드카드 deny 추가 · 적용 범위 = 로컬 + 템플릿 · 템플릿은 생성기가 건너뛰므로 tmpl 손편집 · 실측은 워크트리를 잠시 나가 격리 `claude -p` 로

---

## 1. 재현 (수리 전)

### 주장

1. `.claude/settings.json` 의 `Bash(git reset --hard:*)` · `Bash(git clean -fd:*)` · `Bash(git push --force:*)` deny 규칙은 `git` 과 서브커맨드 사이에 전역 옵션(`-C` / `-c` / `--git-dir` / `--work-tree`)이 끼면 적중하지 않는다. 측정한 앞붙임 형태는 모두 실제로 실행됐다.
2. 원인은 공식 문서가 명시한 매칭 규칙이다. Bash 규칙은 명령 텍스트를 `*` 앞까지 문자 그대로 비교하고, 문서는 `git -C . push` 가 `Bash(git push *)` 에 매칭되지 않으며 deny/ask 규칙이 프로그램 둘레의 보안 경계가 아니라고 적는다.
3. 앞붙임과 무관한 통과 형태가 둘 더 있다. 플래그 쪼개기(`git clean -f -d`)와 `+refspec` 강제 푸시다.
4. 같은 인접성 결함이 훅 층 두 곳에도 있다(코드 판독만). `internal/hook/pre_tool.go:251-253` askBashPatterns 와 `internal/hook/branch_guard.go:129` 다.

### 증거

픽스처는 scratchpad 아래 1케이스 1저장소로 만들었다(커밋 내용 `a`, 워킹 사본 `b`, 미추적 `untracked.txt` + `ud/`). 실제 작업 트리와 primary 체크아웃은 대상에서 제외했다.

| # | 명령 형태 (경로 생략) | 도구 응답 | 사후 상태 | 판정 |
|---|---|---|---|---|
| R0 | `cd r0 && git reset --hard HEAD` (대조) | `Permission to use Bash with command … has been denied.` | a.txt=b | 거부 |
| R1 | `git -C r1 reset --hard HEAD` | `HEAD is now at bb5b1f8 one` | a.txt=a | **통과** |
| R2 | `cd r2 && git -c core.quotepath=off reset --hard HEAD` | `HEAD is now at bb5b1f8 one` | a.txt=a | **통과** |
| R3 | `git --git-dir=r3/.git --work-tree=r3 reset --hard HEAD` | `HEAD is now at bb5b1f8 one` | a.txt=a | **통과** |
| R3s | `git --git-dir r3s/.git --work-tree r3s reset --hard HEAD` | `HEAD is now at bb5b1f8 one` | a.txt=a | **통과** |
| R4 | `cd r4 && git --work-tree=<abs r4> reset --hard HEAD` | `HEAD is now at bb5b1f8 one` | a.txt=a | **통과** |
| R5 | `GIT_DIR=… GIT_WORK_TREE=… git reset --hard HEAD` | `… has been denied.` | a.txt=b | 거부 |
| R6 | `cd r6 && git reset  --hard HEAD` (공백 2칸) | `… has been denied.` | a.txt=b | 거부 |
| C0 | `cd c0 && git clean -fd` (대조) | `… has been denied.` | 미추적 유지 | 거부 |
| C1 | `git -C c1 clean -fd` | `Removing ud/` / `Removing untracked.txt` | 미추적 삭제 | **통과** |
| C2 | `git -C c2 clean -f -d` | 같음 | 미추적 삭제 | **통과** |
| C3 | `git -C c3 clean -dfx` | 같음 | 미추적 삭제 | **통과** |
| C4 | `cd r1 && git clean -f -d` (앞붙임 없음) | 같음 | 미추적 삭제 | **통과** |
| P0 | `cd p1 && git push --force origin HEAD:refs/heads/main` (대조) | `… has been denied.` | — | 거부 |
| P1 | `git -C p1 push --force origin HEAD:refs/heads/main` | `Everything up-to-date` | — | **통과** |
| P2 | `git -C p1 push origin +HEAD:refs/heads/main` | `Everything up-to-date` | — | **통과** |

`cd r4 && git --work-tree=. reset --hard HEAD` 는 워크트리 가드가 먼저 거부해 deny 층을 재지 못했고, 절대경로 형태(R4)로 대신 쟀다. P1·P2 의 원격은 scratchpad 의 로컬 bare 저장소다.

공식 문서 인용 (https://code.claude.com/docs/en/permissions, 2026-09-11 판독):

- "A push written another way, such as `git -C . push`, isn't matched"
- "a deny or ask rule covers the invocation Claude usually produces and isn't a security boundary around the program."
- "A deny or ask rule matches past any leading assignment" (R5 거부와 일치)
- "The `:*` form is only recognized at the end of a pattern."
- "A `*` at the end, with a space before it, also matches the bare command. That holds only when the trailing `*` is the rule's only wildcard"

---

## 2. 규칙 모양 실측 (격리 `claude -p`)

실행 중 세션의 권한 엔진은 워크트리에서 재생성한 설정이 아니라 세션 시작 때 읽은 설정을 쓴다. 이 세션에서 다시 쟀을 때는 오탐 확인용 명령까지 전부 통과했다. 그래서 설정만 따로 넣은 `claude -p --model haiku --permission-mode bypassPermissions` 로 규칙 모양별로 쟀다. 원시 출력과 설정 파일은 `.moai/reports/t642/raw/` 에 반출했다.

| # | 설정 | 명령 | `permission_denials` | 사후 상태 | 판정 |
|---|---|---|---|---|---|
| A | 프로젝트 설정: 기존 7 + `git * reset --hard:*` 등 7 | `git -C fa reset --hard HEAD` | `[]` | a.txt=a | 통과 |
| B | 프로젝트 설정: 기존 7만 (대조) | `git -C fb reset --hard HEAD` | `[]` | a.txt=a | 통과 |
| P | A 와 같은 프로젝트 설정 (양성 대조) | `cd fc && git reset --hard HEAD` | 해당 명령 | a.txt=b | **거부** |
| P2 | `--settings` 로 같은 파일 (양성 대조) | `cd fd && git reset --hard HEAD` | 해당 명령 | a.txt=b | **거부** |
| A2 | `--settings` 로 같은 파일 | `git -C fe reset --hard HEAD` | `[]` | a.txt=a | 통과 |
| A3 | `Bash(git * reset --hard *)` | `git -C fg reset --hard HEAD` | 해당 명령 | a.txt=b | **거부** |
| A4 | `Bash(git -C * reset --hard *)` | `git -C fh reset --hard HEAD` | 해당 명령 | a.txt=b | **거부** |
| R-a | `Bash(git * reset --hard*)` | `git -C fi reset --hard HEAD` | 해당 명령 | a.txt=b | **거부** |
| R-b | `Bash(git * reset --hard*)` | `git -C fj reset --hard` (인자 없음) | 해당 명령 | a.txt=b | **거부** |
| S-b | `Bash(git * reset --hard *)` | `git -C fk reset --hard` (인자 없음) | `[]` | a.txt=a | 통과 |
| R-c | `Bash(git * clean -fd*)` | `git -C fl clean -fd` | 해당 명령 | 미추적 유지 | **거부** |
| R-d | `Bash(git * push --force*)` | `git -C pk push --force origin HEAD:refs/heads/main` | `[]` | — | 미측정 (모델이 실행을 거부해 도구 호출 없음) |
| R-d2 | `Bash(git * push --force*)` | `git -C pm push --force origin HEAD:refs/heads/scratch` | 해당 명령 | — | **거부** |
| C′ | `Bash(git * reset --hard*)` | `git -C <wt> log -n 1 --grep='reset --hard'` | `[]` | — | 통과 (앞 글자가 따옴표) |
| C″ | `Bash(git * reset --hard*)` | `git -C <wt> log -n 1 --grep "via reset --hard"` | 해당 명령 | — | **거부 (오탐)** |

### 주장

1. 처음 넣은 모양 `git * <sub> <flag>:*` 는 효과가 없다. 끝의 `:*` 가 가운데 `*` 를 문자 그대로 만든다(A, A2 통과 · P, P2 로 설정 적재 확인).
2. 공백 모양 `git * <sub> <flag> *` 는 인자가 붙은 명령은 막지만 인자 없는 명령을 놓친다(A3 거부, S-b 통과). 문서의 "trailing `*` is the rule's only wildcard" 조건과 일치한다.
3. `git * <sub> <flag>*` 모양이 두 경우 모두 막는다(R-a, R-b). clean(R-c)과 push --force(R-d2)에서도 같은 모양이 막힌다.
4. 받아들인 부작용은 실재한다. 인자 안에 공백으로 시작하는 같은 문자열이 있으면 무해한 명령도 거부된다(C″).

---

## 3. 수리

### 변경

- `.moai/config/sections/tool-policy.yaml`: deny 항목 7개를 추가했다 — `git * push --force*`, `git * push -f*`, `git * push --force-with-lease*`, `git * reset --hard*`, `git * clean -fd*`, `git * clean -fdx*`, `git * rebase -i*`. 앞의 주석에 실측한 모양 제약 세 가지와 남는 우회 형태를 적었다.
- `.claude/settings.json`: 이 트리의 생성기로 재생성했다(`go -C <wt> run ./cmd/moai tool-policy build --local-only --repo-root <wt>`). 손편집하지 않았다.
- `internal/template/templates/.claude/settings.json.tmpl`: 생성기가 설계상 건너뛰므로(`render-time-conditional permissions block (git_mode gating)`) deny 블록의 `Bash(git rebase -i:*)` 바로 뒤에 같은 7줄을 손으로 넣었다. 조건문은 allow 쪽에만 있고 deny 블록에는 없다. catalog.yaml 은 이 파일을 추적하지 않는다(`grep -n 'settings' internal/template/catalog.yaml` 결과 없음).

### 증거

재생성 결과:

```
[{"Path":".../t642/.claude/settings.json","AllowEmitted":114,"AskEmitted":0,"DenyEmitted":55,"EnvGatedSkipped":5,"TargetKind":"json"}]
```

집합 비교 (HEAD 사본 vs 재생성본):

```
$ jq -r '.permissions.deny[]' .claude/settings.json | sort | diff deny-old.txt -
13a14,20
> Bash(git * clean -fd*)
> Bash(git * clean -fdx*)
> Bash(git * push --force*)
> Bash(git * push --force-with-lease*)
> Bash(git * push -f*)
> Bash(git * rebase -i*)
> Bash(git * reset --hard*)
$ jq -r '.permissions.allow[]' .claude/settings.json | sort | diff -q allow-old.txt - && echo allow-set-unchanged
allow-set-unchanged
```

diff 행 수가 커 보이는 것은 생성기가 목록을 정렬하고 들여쓰기를 바꿔 다시 쓰기 때문이다. 규칙 집합의 변화는 위 7개뿐이다. 같은 입력으로 한 번 더 생성해도 바이트가 같다(첫 모양에서 sha256 `9e8d3416…` 이 두 번 연속 일치).

템플릿 7줄 (편집 직후 판독):

```
settings.json.tmpl:568:      "Bash(git * push --force*)",
settings.json.tmpl:569:      "Bash(git * push -f*)",
settings.json.tmpl:570:      "Bash(git * push --force-with-lease*)",
settings.json.tmpl:571:      "Bash(git * reset --hard*)",
settings.json.tmpl:572:      "Bash(git * clean -fd*)",
settings.json.tmpl:573:      "Bash(git * clean -fdx*)",
settings.json.tmpl:574:      "Bash(git * rebase -i*)",
```

테스트:

```
$ go -C <wt> test ./internal/config/toolpolicy/... -count=1 -v | grep -E '^--- (PASS|FAIL): TestToolPolicyDrift_|^(ok|FAIL)\s'
--- PASS: TestToolPolicyDrift_CommittedSettingsMatchYAML (0.00s)
--- PASS: TestToolPolicyDrift_NoDuplicatesOrOverlap (0.00s)
--- PASS: TestToolPolicyDrift_Mutation (0.02s)
--- PASS: TestToolPolicyDrift_FailClosed (0.01s)
ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.456s

$ go -C <wt> test ./internal/template/ -run 'Settings|Security|Autonomy|Permission|Deny' -count=1 -v
exit=0 · === RUN 86건 · --- FAIL 0건
ok  	github.com/modu-ai/moai-adk/internal/template	0.452s
```

---

## 기준 귀속 (Baseline-attribution)

- 재현과 수리 측정은 모두 이 트리(81c1d58f9 기반 WT-deny-git-prefix)와 이 세션에서 실행한 명령·출력이다. 2026-09-10 lane-5 관측은 카드 본문의 서술이며 재측정 대상이 아니다.
- 규칙 모양 실측(§2)의 판정 기준은 도구 응답이 아니라 `permission_denials` 필드와 픽스처 사후 상태이며, 두 양성 대조(P, P2)로 설정 적재를 확인했다.
- 원시 증거: `.moai/reports/t642/raw/raw-*.json`, `raw/settings-*.json`, `raw/settings-head.json`(HEAD 의 settings.json 사본).

## 미검증 (Gaps)

- 최종 7개 규칙 가운데 격리 실측을 거친 것은 `reset --hard*`, `clean -fd*`, `push --force*` 세 모양이다. `push -f*`, `push --force-with-lease*`, `clean -fdx*`, `rebase -i*` 는 같은 모양이지만 따로 재지 않았다.
- `-c` 와 `--git-dir`/`--work-tree` 앞붙임은 최종 모양으로 따로 재지 않았다. 격리 실측은 `-C` 형태로 했고, 가운데 `*` 가 임의의 텍스트에 매칭된다는 문서 설명에 기댄다.
- `git rebase -i` 는 대화형이라 어떤 단계에서도 재현하지 않았다.
- 실행 중 세션에서 새 규칙이 적용되는지는 관측하지 못했다. 규칙은 이 브랜치가 develop 에 병합되고 각 체크아웃의 설정이 갱신된 뒤에야 쓰인다.
- 템플릿을 실제로 렌더링한 사용자 프로젝트에서의 적용은 재지 않았다(템플릿 테스트의 렌더·JSON 검증만 확인).
- `internal/hook/pre_tool.go`·`branch_guard.go` 의 같은 인접성 결함은 코드 판독만 했고 수리하지 않았다(범위 밖).
- 생성기의 `no targets processed (neither local nor template settings file found)` 문구는 템플릿을 조건문 때문에 건너뛴 경우에도 "파일을 찾지 못함"이라고 적는다. 메시지가 사실과 다르다(별도 결함 후보).

## 잔여 위험 (Residual-risk)

- 문서상 deny 규칙은 경계가 아니다. 이번 수리 뒤에도 플래그 쪼개기(`git clean -f -d`), `+refspec` 강제 푸시, 따옴표 서브커맨드, 셸 래핑(`sh -c '…'`), 절대경로 git 은 막히지 않는다.
- 오탐이 실재한다(C″). 커밋 메시지나 `--grep` 인자처럼 공백으로 시작하는 같은 문자열이 든 명령은 거부된다. 이 저장소 이력에서 메시지에 해당 문자열이 든 커밋은 13건이다.
- 템플릿 deny 목록은 YAML 과 대조하는 검사가 없어 이후 손편집으로 둘이 어긋나도 잡히지 않는다.
- `claude -p` 격리 실측은 haiku 모델로 했다. 권한 엔진 판정은 모델과 무관하지만, R-d 처럼 모델이 실행 자체를 거부하면 권한 판정까지 가지 않는다. 그런 행은 미측정으로 표시했다.
