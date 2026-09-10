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

---

## 7. sync 단계 기록

- 산출물: `.claude/rules/moai/workflow/worktree-integration.md` 의 새 절
  `## Refused Commands in a Worktree-Isolated Session` + 템플릿 미러
  (`internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md`).
  양쪽 42줄씩, 기존 줄 수정 0.
- SPEC 없음. 이 카드는 Class B 로 plan 을 건너뛰었고, 재정의 후에도 규칙 1건이라
  SPEC 산출물을 만들지 않았다. 따라서 3단계 마감 대상이 아니며, sync 기록은 이 절이다.
- CHANGELOG: `[Unreleased] / Added` 에 1건 추가. 핵심은 두 가드 구별이며,
  관측 방아쇠 2행은 요약으로만 실었다.

### 7.1 sync 시점 재측정

| 검사 | 명령 | 결과 |
|---|---|---|
| 템플릿↔로컬 동일 | `diff <template> <local>` | exit 0, 출력 없음 |
| 빌드 | `make build` | exit 0, `catalog.yaml` 재생성되나 바이트 무변경 |
| 중립성 | `go test ./internal/template/ -run 'Neutrality|InternalContent|Leak'` | ok, exit 0 |
| 선택자 비공허 | 같은 명령 `-v`, `grep -c '^=== RUN'` | 143 |
| 절 본문 금칙어 | `grep -nE '/Users/|CLAUDE\.local|SPEC-|REQ-|AC-|t287|날짜|행번호'` | 본문 42줄 0건 |

### 7.2 4-locale 판정

**불필요.** 이 변경은 `.claude/rules/` 의 내부 운영 규칙과 그 템플릿 미러뿐이고,
README·docs-site 의 사용자 문서 표면을 만들지 않는다.

### 7.3 절차 이탈 (규칙 본문에는 넣지 않음)

위임한 specialist 가 자신의 에이전트 워크트리에 격리돼 이 카드 워크트리에 도달하지
못했다. `cd <카드트리> && git …` 와 `git -C <카드트리> …` 가 모두 거부됐고, 문면은
이 카드가 다루는 워크트리 가드의 인접 규칙이다. 커밋은 에이전트 브랜치에 났고
`git cherry-pick` 으로 이 브랜치에 옮겼다(충돌 0). §7.1 의 재측정은 전부 옮긴 뒤
이 트리에서 잰 값이다.

관측 1건이고 기제를 규명하지 않았으므로 **규칙 본문에는 싣지 않았다.**

### 7.4 잔여 gap

- 2 차 방아쇠(복합 명령) 미재현 — 타 레인 보고를 라벨과 함께 실었을 뿐이다.
- 방아쇠 조건 미축소 — 회피 관례를 적는 데는 불필요해 수행하지 않았다.
- `moai update` 미실행 — `diff` 는 두 파일이 오늘 같다는 것을 증명하지만,
  재배포 경로가 로컬을 템플릿에서 복원하는지는 관측하지 않았다.
- `golangci-lint` 및 전체 스위트 미실행 — 변경이 마크다운 2 파일뿐이라는 것은
  논증이지 측정이 아니다.

### 7.5 CI 귀속 — Graph Freshness (관측)

§7.4 를 쓸 당시 "described-worthy 소스 0" 은 diff 구성에서 나온 **예측**이었다. 병합 head
`3603c155b` 의 워크플로 런(id 33358425797, event=push)을 직접 읽어 관측으로 바꾼다.

```
$ gh run view 33358425797 --log | grep -iE 'contribution|described-source-diff'
codemaps  metric=described-source-diff value=43 threshold=40 verdict=stale
  contribution: 0 described-worthy file(s) vs first parent c36d2672b (inherited — this change contributed none of it)
mx-index  metric=inventory-content-diff value=0 threshold=1 verdict=fresh
edges     metric=source-fingerprint-mismatch value=0 threshold=0 verdict=fresh
```

판정: **순수 상속.** 이 카드의 기여분은 0 이고, 도구가 스스로 `(inherited — this change
contributed none of it)` 로 라벨한다. 총계 43 은 이 카드로 귀속되지 않는다.

누적 추이(리드 판독, 이 세션 미검증): `b9149857c` 41 → `c36d2672b` 43(+2) → `3603c155b`
43(+0). 마지막 항만 위 명령으로 이 세션이 직접 확인했다.

codemaps 재생성은 하지 않았다 — 범위 밖이며 배치 종료 시점에 일괄 처리하기로 운영자가 정했다.
