# t287 판정 — 카드 전제 2건 반증, 착수 보류

- 카드: t287 (GitHub 이슈 #1659), Class B, Tier S
- 측정 트리: `.claude/worktrees/t287`, 브랜치 `WT-guard-substitution`, HEAD 1e5199b88 (origin/develop 조상)
- Claude Code: 2.1.251
- 상태: **구현 착수하지 않음.** 카드가 지시한 조치 (2)(3)이 이 저장소에서 성립하지 않는다.

## 1. Claim

카드의 전제 두 가지가 모두 사실과 다르다.

1. **"명령 치환은 우회된다"는 거짓이다.** 명령 치환은 heredoc 본문에서 올바르게 접힌다.
   결함은 우회(false negative)가 아니라 **과다 거부(false positive)** 이며, 대상은 중괄호다.
2. **이 가드는 moai 코드가 아니다.** Claude Code 바이너리 안에 있어, 토큰화 개선을
   이 저장소에서 착지시킬 수 없다.

## 2. Evidence — 두 프로브, 같은 세션·같은 형태

두 명령은 heredoc 본문 한 줄만 다르고 나머지는 동일하다. 대상 경로도 같다.

### 프로브 A — 본문에 중괄호 (JSON 한 줄)

    cat > /tmp/t287-probe-a.txt <<'EOF'
    (JSON 한 줄: 중괄호 안에 큰따옴표로 감싼 키와 값)
    EOF

결과: **거부**

    This session is isolated in the worktree /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t287,
    but this command is too complex to verify that it stays inside the worktree.

### 프로브 B — 본문에 명령 치환

    cat > /tmp/t287-probe-b.txt <<'EOF'
    (달러+괄호로 감싼 echo 한 줄)
    EOF

결과: **통과**, rc=0

두 프로브가 이슈 #1659의 대조표를 그대로 재현한다. 결정적 대조군은 B다 — 가드는 따옴표 붙은
heredoc 본문이 확장되지 않는다는 것을 **이미 알고 명령 치환에 적용하고 있으면서**, 중괄호
판정에만 그 지식을 쓰지 않는다.

### 부수 관측

이 보고서 자체를 heredoc으로 쓰려다 같은 거부를 만났다. 본문에 JSON 예시가 들어 있었기
때문이며, python 경유로 우회해 작성했다. 이슈의 "실무 영향" 절이 서술한 상황과 동일하다.

## 3. Evidence — 가드의 소재

`grep -rl "too complex to verify" .` → 히트 8건 전부 `.moai/reports/**` 와 `.moai/specs/**`
(보고서·SPEC 산문). **소스 코드 히트 0건.**

선행 조사가 같은 결론을 이미 기록해 두었다 — `.moai/reports/backlog/t25-cwdguard-investigation.md`
L9 "The guard is not MoAI code. It lives in the Claude Code binary, not in this [repo]",
L75 "The message string is in the Claude Code binary", L226 "Governing MoAI rule: none for this
guard", L251 "opaque upstream dependency".

`internal/hook/branch_guard.go`는 **다른 가드**다 — primary 체크아웃의 브랜치 상태 변경을
막는 moai 소유 가드이고, 이 워크트리 가드를 구현하지도 설정하지도 않는다.

## 4. 카드 조치 3단계의 성립 여부

| 카드 조치 | 성립 | 근거 |
|---|---|---|
| (1) 우회 재현 | **불가 — 우회가 존재하지 않음** | 프로브 B가 통과하는 것이 정상 동작. 접어야 할 것을 접고 있다 |
| (2) 치환 위치 토큰화 개선 | **불가 — 이 저장소에 대상 코드 없음** | §3. 고칠 소스가 Claude Code 바이너리 안에 있다 |
| (3) 회귀 쌍 | **불가 — (2)에 종속** | 착지시킬 변경이 없으면 회귀도 없다 |

## 5. Gaps

- **업스트림 보고 여부 미확인.** 이 오탐이 Anthropic에 이미 보고됐는지 검색하지 않았다.
  t25 조사도 같은 gap을 남겼다(L242).
- **버전 경계 미측정.** 2.1.251에서만 쟀다. 어느 판번호에서 도입·수정됐는지 모른다.
- **중괄호 판정의 정확한 조건 미탐색.** 이슈 표에 따르면 중괄호 하나, 쉼표 두 항목은
  통과하고 따옴표 낀 키/값에서 거부된다. 어떤 부분 문자열이 방아쇠인지 좁히지 않았다 —
  업스트림 보고서에는 그 축소가 필요할 수 있다.

## 6. Residual-risk

이 판정이 맞더라도 실무 마찰은 남는다. 워크트리 세션에서 JSON을 담은 문서를 heredoc으로
쓸 수 없고, 우회(python 경유, 파일 분할)는 매번 사람이 기억해야 한다. **가드를 고치는
카드가 아니라, 이 저장소 쪽 회피 관례를 문서화하는 카드**라면 성립한다 — 다만 그것은
카드 t287이 적은 것과 다른 작업이므로 운영자 판정 사항이다.
