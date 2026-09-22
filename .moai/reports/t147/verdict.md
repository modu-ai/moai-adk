# t147 — 워크트리 브랜치 명명 전환 (`WT-<card-id>` → `WT-<서술 슬러그>`)

- 카드: t147 (Class B — plan 생략, run → sync 인레인)
- 브랜치: `WT-t147`
- 워크트리: `.claude/worktrees/t147`
- 베이스: `origin/release/v3.1.1` @ `b317f47c4`
- 판정: **PASS** (미푸시, 리드 통합 대기)

---

## 1. Claim (주장)

| # | 주장 |
|---|---|
| C1 | 카드 워크트리 브랜치 규약을 `WT-<card-id>`에서 `WT-<서술 슬러그>`로 전환했다 |
| C2 | 길이 상한(하이픈 3토큰·24자)과 알파벳을 규약에 명시했다 |
| C3 | 카드 ID를 브랜치명에서 제거하면서 추적성 3경로(디스패치 `card:` 필드·커밋 메시지·evidence 경로)를 [HARD]로 못박았다 |
| C4 | 독트린 4표면 + 템플릿 미러를 일관되게 갱신했고 미러 패리티가 깨지지 않았다 |
| C5 | 전수 개명은 하지 않았다 (t145 폐기 이후로 순서 고정) |
| C6 | 코드 계층에 회귀가 없다 — 브랜치명에서 카드 ID를 파싱하는 코드는 존재하지 않는다 |

## 2. Evidence (증거)

### C1 / C2 / C3 — 독트린 변경

`kanban-dispatch.md` :143 [HARD] 블록을 재작성. 새 규약 표:

| Property | Rule |
|---|---|
| Source | The card's title, not its id |
| Tokens | At most 3, hyphen-separated |
| Length | At most 24 characters (slug alone; `WT-` brings the branch to at most 27) |
| Alphabet | Lowercase `a-z`, `0-9`, and `-` |
| Card id | MUST NOT appear — not as a prefix, a suffix, or a token |

추가 [HARD] — 추적성 3경로: 디스패치 `card:` 필드 / 커밋 메시지 / evidence 경로 `.moai/reports/<card-id>/verdict.md`.
**워크트리 디렉터리는 카드 ID를 유지**한다(`.claude/worktrees/<card-id>`) — 폐기 도구와 evidence 경로가 여기에 키를 걸고 있으므로 ID는 유실되지 않는다.

`worktree-integration.md` :56 개정 + :58 신규 [HARD], 버전 `4.3.0 → 4.4.0`.

### C4 — 표면 목록과 미러 패리티

변경 파일 14개 (`git status --porcelain`):

```
 M .claude/rules/moai/workflow/kanban-dispatch.md
 M .claude/rules/moai/workflow/kanban-dispatch-detail.md
 M .claude/rules/moai/workflow/worktree-integration.md
 M README.md  README.ko.md  README.ja.md  README.zh.md
 M docs-site/content/{en,ko,ja,zh}/core-concepts/kanban-board-terms.md
 M internal/template/templates/.claude/rules/moai/workflow/{kanban-dispatch,kanban-dispatch-detail,worktree-integration}.md
```

미러 패리티 — `diff -q` 3/3 무출력(동일):

```
diff -q .claude/rules/moai/workflow/kanban-dispatch.md        internal/template/templates/.claude/.../kanban-dispatch.md         → (동일)
diff -q .claude/rules/moai/workflow/kanban-dispatch-detail.md internal/template/templates/.claude/.../kanban-dispatch-detail.md  → (동일)
diff -q .claude/rules/moai/workflow/worktree-integration.md   internal/template/templates/.claude/.../worktree-integration.md    → (동일)
```

`cp` 미러가 템플릿의 의도적 중립화를 덮어썼는지 확인 — 복사 **전** HEAD 판본이 이미 로컬과 byte-identical 이었음:

```
git show HEAD:.claude/.../kanban-dispatch.md      vs  git show HEAD:internal/template/.../kanban-dispatch.md      → 동일
git show HEAD:.claude/.../worktree-integration.md vs  git show HEAD:internal/template/.../worktree-integration.md → 동일
```

빌드·검증:

```
make build                                                    → exit 0, catalog.yaml (12899 bytes), bin/moai 재생성
MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -run 'Leak|Neutral' -count=1
                                                              → ok  github.com/modu-ai/moai-adk/internal/template  2.173s
go test ./internal/template/ -count=1                         → ok  github.com/modu-ai/moai-adk/internal/template  23.457s
go vet ./internal/template/                                   → rc=0
hugo --quiet --logLevel warn (docs-site)                       → rc=0, 경고 0
```

4로케일 정합성:

```
grep -c 'WT-\|worktrees/t' docs-site/content/{en,ko,ja,zh}/core-concepts/kanban-board-terms.md  → 8 / 8 / 8 / 8
grep -n 'flowchart\|^graph ' (동일 4파일)                                                        → flowchart TD ×8, LR/RL 0
grep -c '^| ' README.md README.ko.md README.ja.md README.zh.md                                  → 145 / 145 / 145 / 145
```

잔존 `WT-t0` 3건은 의도적 반례(`WT-t0`이 아니라 `WT-todo-queue`)이며 docs-site 4로케일 중 ko/ja/zh 본문에만 존재한다.

### C5 / C6 — 순서 제약과 코드 영향

전수 개명 미수행 — `git status`에 브랜치 개명 흔적 없음. 기존 52개 `WT-*` 브랜치 무변경.

`.moai/reports/t145/dispose-merged.sh` 실측:

```
grep -c '^dispose ' .moai/reports/t145/dispose-merged.sh  → 52
```

코드 계층 — 브랜치명에서 카드 ID를 파싱하는 지점 없음:

```
grep -rn 'SessionWorktreeBranchPrefix' --include='*.go' internal/
  internal/cli/session_worktree.go:40         const SessionWorktreeBranchPrefix = "WT-"
  internal/cli/session_worktree_prmerge.go:147  if !strings.HasPrefix(e.branch, SessionWorktreeBranchPrefix)
```

`HasPrefix` 접두사 검사만 존재하고 나머지 문자열은 읽지 않는다. 슬러그 전환은 코드 무영향.

## 3. Baseline-attribution (baseline 귀속)

모든 측정은 워크트리 `.claude/worktrees/t147`, 브랜치 `WT-t147`, 베이스 `origin/release/v3.1.1` @ `b317f47c4` 기준. 이번 실행에서 직접 실행해 관측한 출력이며 이전 카드에서 옮겨온 수치는 없다.

```
git rev-parse --show-toplevel  → /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t147
git branch --show-current      → WT-t147
git rev-parse --short HEAD     → b317f47c4 (편집 전)
```

**베이스 정정 기록**: 배차의 `EnterWorktree(t147)`은 기본값 `origin/main`(`4100d8767`)으로 트리를 깔았으나, 카드가 지목한 `kanban-dispatch.md:79`·`:143`이 그 판본과 불일치했다(:79 내용 상이, :143 부재). 두 브랜치 diff는 해당 2파일에서 86+14줄. `origin/release/v3.1.1`에서 :79·:143이 정확히 일치함을 확인하고 `git merge --ff-only origin/release/v3.1.1`로 베이스를 이동했다(main은 release의 조상이라 fast-forward).

## 4. Gaps (미검증)

- **전체 테스트 스위트 미실행.** 레인 로컬 검증만 수행(`internal/template`). 전 패키지 판정은 CI 몫 — 배차의 "전체 스위트 금지"에 따름.
- **docs-site 렌더 결과 육안 미확인.** `hugo` 빌드가 경고 0으로 통과했을 뿐, mermaid 서브그래프 라벨이 실제 브라우저에서 어떻게 줄바꿈되는지는 보지 않았다.
- **README ASCII 다이어그램 정렬을 터미널에서 미확인.** `WT-t0`(5자) → `t0`(2자) 축소분을 공백 3칸으로 보정했으나, 실제 렌더 폭(CJK 이중폭 포함)은 재지 않았다.
- **`golangci-lint` 미실행.** 마크다운만 변경했고 Go 소스 무변경이라 생략.
- **슬러그 규약의 기계적 강제 없음.** 24자·3토큰 상한을 검사하는 린터나 훅은 만들지 않았다 — 이번 카드는 독트린 표면만 범위.

## 5. Residual-risk (잔여 위험)

- **순서 의존이 문서화되지 않은 채 남아 있다.** 새 규약은 오늘부터 만드는 브랜치에만 적용되지만, 기존 52개가 `WT-<card-id>`로 남아 혼재 기간이 생긴다. 규약 파일에는 이 과도기 서술이 없다.
- **`dispose-merged.sh`의 고장 형태가 카드 서술과 다르다** (아래 §6 정정 참조). 순서 제약 자체는 유효하지만, 리드가 "건너뛴다"로 알고 있으면 사후 확인 지점을 잘못 잡을 수 있다.
- **슬러그 중복 가능성.** 서로 다른 카드가 비슷한 제목이면 같은 슬러그를 만들 수 있고, `git branch -m`이 기존 브랜치와 충돌하면 실패한다. 규약에 충돌 해소 절차를 넣지 않았다.
- **범위 이탈 1건** (아래 §6).

## 6. 리드에게 보고할 정정 2건

### (1) `dispose-merged.sh` 고장 형태 — 카드 서술 정정

카드 [HARD]는 *"먼저 개명하면 dispose-merged.sh 가 브랜치명 하드코딩이라 52개를 전부 건너뛴다"* 라고 적었다. 스크립트 실측 결과 **순서 제약은 유효하나 고장 형태가 다르다.**

`dispose()`의 스킵 조건은 4개뿐이고 전부 브랜치명과 무관하다:

```
1. [ ! -d "$path" ]                                  디렉터리 부재
2. ! git merge-base --is-ancestor "$sha" origin/main  미병합
3. git -C "$path" status --porcelain -uall 비어있지 않음  미커밋/미추적
4. live_cwds 에 매칭                                   라이브 세션 앵커
```

브랜치명(`$2`)은 **마지막 줄에서만** 쓰인다:

```
git worktree remove "$path" && git update-ref -d "refs/heads/$branch"
```

따라서 먼저 개명하면 → **건너뛰지 않고 트리는 정상 제거되며, `update-ref -d`만 52번 실패해 고아 브랜치 ref 52개가 남는다.** 결과적으로 "개명은 t145 폐기 이후" 라는 순서는 그대로 지켜야 하지만, 실패 징후는 "SKIP 로그"가 아니라 "폐기 후 남은 `WT-*` ref"다.

### (2) 범위 이탈 1건 — `kanban-dispatch-detail.md`

배차 범위는 `kanban-dispatch.md(:79, :143)`였으나, 그 지연 companion인 `kanban-dispatch-detail.md:27`이 용어집 정의로 `Branch named WT-<card-id>`를 들고 있었다. 그대로 두면 새 [HARD]와 정면 모순이고, docs-site 4로케일 용어집이 이 파일을 미러하므로 모순이 사용자 문서까지 전파된다. **범위를 1파일 넓혀 고쳤다** (3줄). 되돌릴지는 리드 판단.

---

## 7. 다음 카드로 넘길 후보

1. **전수 개명 카드** — t145 폐기(APPLY) 완료 후. 기존 `WT-<card-id>` 52개를 슬러그로 개명하거나, 과도기를 인정하고 그대로 소멸시킬지 결정 필요.
2. **슬러그 검증 훅** — 24자·3토큰·카드 ID 부재를 `git branch -m` 시점에 기계적으로 검사. 지금은 사람이 지키는 규약.
3. **슬러그 충돌 해소 절차** — 동일 슬러그 발생 시 규약.
