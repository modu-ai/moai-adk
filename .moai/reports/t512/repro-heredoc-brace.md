# t512 재현 원장 — heredoc 중괄호 접기 비대칭 (GH #1659)

- 측정 세션: lane-7 (t512 카드 워크트리, Claude Code worktree-isolation 세션)
- 트리/HEAD: **트리 `.claude/worktrees/t512` · HEAD `6a46c0edb`** (로컬 develop tip, 브랜치 `WT-guard-heredoc`, 카드 커밋 0개 상태)
- 날짜: 2026-09-07

## 프로브 1 — 제보 형태 (따옴표 붙은 구분자 + 본문 중괄호)

명령 (그대로 발사):

```
cat > .claude/worktrees/t512-repro-brace.txt <<'T512EOF'
{"guard": "probe", "form": "brace-in-quoted-heredoc", "issue": 1659}
T512EOF
echo "written rc=$?"
```

관측 (Claude Code가 명령을 **실행하지 않고** 반환한 거절, verbatim):

```
This session is isolated in the worktree /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t512, but this command is too complex to verify that it stays inside the worktree. Refusing to run it — a worktree-isolated session's git operations must target its own worktree. Split it into plain, separate commands and run them from /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t512.
```

미실행 증명: `ls t512-repro-brace.txt` → `No such file or directory` (rc=1) — 거절이 실제 실행을 막았다.

## 프로브 2 — 대조군 (같은 형태 + 본문 명령 치환)

명령:

```
cat > t512-repro-subst.txt <<'T512EOF'
value: $(echo substitution-inside-quoted-heredoc-body)
T512EOF
echo "subst-written rc=$?"
```

관측: **거절 없음 — 실행됨**. `subst-written rc=0`, 파일 55바이트 생성, 본문은 리터럴 그대로(`$(echo …)` 치환 미발생 — 따옴표 붙은 구분자 규격대로). 즉 가드의 분석은 명령 치환을 접어서(무해 텍스트로 보고) 통과시켰다.

## 비대칭 판정

같은 heredoc 구조에서 본문 명령 치환은 통과, 본문 중괄호는 거절 — **접기 비대칭이 이 세션에서 생으로 재현됐다**(제보자 주장과 일치). 따옴표 붙은 구분자의 본문은 bash가 어떤 확장도 하지 않는 순수 텍스트이므로 중괄호도 확장 불가능 — 거절은 오판이다.

## 소속 판별

- 거절 문면 `too complex to verify that it stays inside the worktree`는 **이 리포 소스에 0건**(`grep -rn 'too complex to verify' internal/ .claude/hooks/` → 문서 3건뿐: `worktree-integration.md:481,485,491` — HEAD 6a46c0edb).
- 리포 자체 문서(`worktree-integration.md:481`)가 명시: 이 거절은 **The Claude Code binary** 소유, "No — there is no source here that implements or configures it". `branch_guard.go`는 이 가드가 아니며 편집해도 거절에 영향 없음(:485).
- 결론: 결함의 몸은 이 저장소 밖(바이너리)에 있다. 리포가 할 수 있는 것은 문서 보강·회신·업스트림 제보 초안이다.

## 스크래치 정리

`repro-subst.txt`는 증거 인용 후 삭제 예정(워크트리 루트 무작위 파일 잔존 방지). 프로브 1 파일은 거절로 생성되지 않음.
