# t217 — 보안 스캔 표면 조사 (plan-phase 선행 근거)

카드: t217 / 브랜치: `security-scan-surface` / 워크트리: `.claude/worktrees/t217`
선행 확인: PR #1625(t214) MERGED — merge commit `ea4c6736f66e8029f2347f94ef82a4e9f88fd74d` (2026-08-24T05:55:11Z)
측정 트리: `c4e90cd58` (worktree t217, origin/main 기준)

---

## Claim 1 — PreToolUse 차단 능력은 7개월간 단 한 번도 발화하지 않았다

**Evidence**

deny 사유 문자열은 `internal/hook/pre_tool.go:652` 의
`fmt.Sprintf("Security vulnerabilities detected in %s:\n%s", ...)` 하나뿐이다.
런타임 발화면 `%s` 가 치환된 형태로, 소스 인용이면 `%s` 그대로 트랜스크립트에 남는다.

```
$ grep -rlE "Security vulnerabilities detected in [^%]" /Users/goos/.claude/projects | wc -l
0
$ grep -rl  "Security vulnerabilities detected in %s"   /Users/goos/.claude/projects | wc -l
6      # 전부 pre_tool.go 소스를 읽은 기록
$ grep -rlE "Security vulnerabilities detected in [^%]" /Users/goos/.moai/claude-profiles | wc -l
1      # 이 조사 세션 자신의 트랜스크립트(자기 grep 패턴) — 자기 히트
```

**대조군(관측 가능성 검증)** — 훅 deny 사유가 트랜스크립트에 실제로 남는지 먼저 확인했다.

```
$ grep -rlE "BRANCH_GUARD_VIOLATION:" /Users/goos/.claude/projects | wc -l
151
```

BranchGuard deny 사유는 151개 파일에 남아 있다. 따라서 보안 스캔 deny 의 0건은
"기록되지 않아서 안 보이는 것"이 아니라 **실제 0회**다.

**Baseline-attribution** — 트랜스크립트 12,265개(`~/.claude/projects`) + 3,373개
(`~/.moai/claude-profiles`), 최고(最古) 파일 mtime 2026-01-28, 측정 시각 2026-08-24.
`.moai/logs/` 는 근거로 쓸 수 없다: `moai hook` 경로는 모든 slog 레코드를 `io.Discard`
로 버리므로(`internal/cli/logging.go:50-63`, MOAI_LOG_LEVEL 로도 못 엶)
`slog.Warn("security scan blocked write operation")` 은 어디에도 착지하지 않는다.

**Gaps** — 이 머신의 트랜스크립트만 훑었다. 배포 사용자 환경은 관측 범위 밖.
2026-01-28 이전 기록은 존재하지 않으므로 그 이전은 알 수 없다.

---

## Claim 2 — 기전은 살아 있다. 0건은 "고장"이 아니라 "안 걸렸다"

**Evidence** — 로컬 룰셋으로 error 심각도 findings 가 실제로 나온다.

```
$ printf 'package main\n\nconst apiKey = "sk-abcdef1234567890"\n' > /tmp/moai-t217-probe.go
$ sg scan -c .../.moai/config/astgrep-rules/sgconfig.yml --json /tmp/moai-t217-probe.go
  → ruleId "sec-hardcoded-credential", severity "error"
$ grep -rh "^severity:" .moai/config/astgrep-rules/ | sort | uniq -c
  14 severity: error
  12 severity: warning
```

`ShouldAlert` 는 error 카운트로 판정하므로(`scanner.go:166` → reporter) 경로는 도달 가능하다.

---

## Claim 3 — 그러나 워크트리 세션에서는 룰이 아예 로드되지 않는다

**Evidence** — `FindRulesConfig`(`rules.go:34`)의 탐색 경로 6개는 전부 프로젝트 루트 기준이고,
5·6번이 `.moai/config/astgrep-rules/sgconfig.{yml,yaml}` 다. 이 리포의 카드 워크트리에는
그 디렉터리가 존재하지 않는다(로컬 전용·미추적).

```
$ ls .moai/config/astgrep-rules          # 워크트리 t217
ls: .moai/config/astgrep-rules: No such file or directory
$ ls /Users/goos/MoAI/moai-adk-go/.moai/config/astgrep-rules   # primary
go  security  sgconfig.yml
```

configPath 가 빈 문자열이면 `ast_grep.go:129-134` 가 `--config` 를 빼고 `sg scan --json <file>` 를 부른다.

```
$ cd /tmp && sg scan --json /tmp/moai-t217-probe.go
Error: No ast-grep project configuration is found.
```

즉 **워크트리에서 돌아가는 모든 세션은 매 Write 마다 temp 파일 쓰기 + `sg` 프로세스 기동을
치르고 findings 0 을 받는다.** 팩토리/칸반 레인은 전부 워크트리에서 돌므로 이 리포의 실제
작업 대부분이 여기에 해당한다.

**Gaps** — 배포 사용자(비-워크트리)는 템플릿이 `.moai/config/astgrep-rules/sgconfig.yml` 을
깔아주므로(`internal/template/templates/.moai/config/astgrep-rules/`) 룰이 로드된다.
워크트리 미로딩이 배포판 결함은 아니다.

---

## Claim 4 — 커버리지 비대칭: 15개 확장자가 스캔을 트리거하는데 룰은 4개 언어뿐

**Evidence**

```
$ grep -rh "^language:" internal/template/templates/.moai/config/astgrep-rules/ | sort | uniq -c
  20 language: go
   2 language: javascript
   2 language: python
   2 language: typescript
```

`supportedLanguages`(`ast_grep.go:358-374`)는 python/javascript/typescript/go/rust/java/
kotlin/c/cpp/ruby/php/swift/csharp/elixir/scala **15개**. `IsSupportedExtension` 이 통과시키는
확장자면 무조건 temp 파일 + `sg` 기동이 일어난다. 11개 언어는 룰이 0개인 채로 전액을 낸다.

---

## Claim 5 — PreToolUse 와 PostToolUse 는 "같은 스캔"이 아니다

**Evidence** — 엔진도 룰도 다르다.

| | PreToolUse `scanWriteContent` | PostToolUse `security-scan` |
|---|---|---|
| 엔진 | ast-grep(`sg`) 셸아웃, temp 파일 | in-process 정규식(`ScanBuffer`) |
| 룰 | `.moai/config/astgrep-rules` 26개(error 14) | `guardianClasses` 10개 취약점 클래스(`patterns.go:77-171`) |
| 언어 | 룰 4개 언어 / 트리거 15개 확장자 | 언어 무관(regex) |
| 판정 | error 있으면 **deny** | 항상 advisory(`additionalContext`), 차단 없음 |
| 대상 | 쓰기 **전** payload 내용 | 쓰기 **후** 버퍼 |

겹치는 것은 `hardcoded-secret`(regex, Critical) ↔ `sec-hardcoded-api-key`/
`sec-hardcoded-credential`(ast-grep, error) 정도다. **어느 쪽도 다른 쪽의 상위집합이 아니다.**
감사 F-4 의 "같은 내용을 두 번 스캔한다"는 바이트가 같다는 뜻이지 판정이 같다는 뜻이 아니다 —
PreToolUse 를 그냥 지우면 ast-grep 룰 26개가 쓰기 경로에서 통째로 사라진다.

---

## Claim 6 — 항목 B(프로세스 병합)는 스캔을 하나도 없애지 않는다

**Evidence** — PostToolUse `Write|Edit|MultiEdit` matcher 에는 3개 엔트리가 걸려 있다
(`.claude/settings.json`):

| 엔트리 | async | timeout | 하는 일 |
|---|---|---|---|
| `handle-post-tool.sh` | true | 10 | LSP 진단 + MX 태그 검증 + ast-grep(`post_tool.go`) |
| `status-transition-ownership.sh` | **없음(동기)** | 5 | `status:` grep, 감사 로그 1줄 |
| `handle-security-scan.sh` | true | 5 | regex `ScanBuffer` → advisory |

B 는 3번을 1번 안으로 접는 것이다. **regex 스캔 로직 자체는 살아남는다** — 사라지는 것은
`moai` 프로세스 기동 1회(감사 실측 회당 260~340ms, load 8~10 환경이라 비율로 읽을 것)뿐이다.
남는 것: LSP + MX + ast-grep + regex 전부. 사라지는 것: 중복 DI 그래프 1개.

**주의(설계 제약)** — 두 핸들러 모두 stdout 에 `hookSpecificOutput` JSON 을 쓴다. 병합하면
`additionalContext` 를 **합쳐서 1개 JSON 으로** 내보내야 한다. 지금은 각자 다른 프로세스라
Claude Code 가 둘 다 받지만, 병합 후 한쪽 필드를 덮어쓰면 advisory 가 조용히 사라진다.
이것이 B 의 유일한 실질 회귀 위험이다.

---

## 부수 관측 — 조사 중 PostToolUse 레이어가 실제로 발화했다

이 보고서 파일을 쓰는 순간 PostToolUse regex 레이어가 울렸다:

```
[MoAI Security Guardian] 1 finding(s): hardcoded-secret (critical) line 52:
Hardcoded credential or private key committed to source
```

52행은 Claim 2 의 `sk-` 프로브 문자열이다. 세 가지가 동시에 확인된다:
(1) PostToolUse regex 레이어는 살아서 advisory 를 낸다, (2) `.md` 는 `IsSupportedExtension`
대상이 아니므로 **PreToolUse 는 같은 내용을 보고도 아무 말도 하지 않았다** — Claim 5 의
"상위집합이 아니다"에 대한 직접 증거, (3) advisory 는 실제로 세션에 도달한다(항목 B 병합 시
이 채널이 보존돼야 하는 이유).

---

## Residual-risk

- 트랜스크립트 grep 은 **이 머신** 한정. "한 번도 안 막았다"는 이 개발 환경에 대한 참이지
  배포 사용자 전체에 대한 참이 아니다. 배포 사용자는 (a) 룰이 실제로 로드되고 (b) Go 프로젝트가
  다수이므로 발화 확률이 이 리포보다 높을 수 있다.
- 절대 시간 재측정은 하지 않았다 — 측정 시점 load average 38.65/28.81/21.31 이라 절대값이
  무의미하다. 감사의 164ms / 260~340ms 를 **비율**로만 인용한다.
