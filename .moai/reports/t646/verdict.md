# t646 — `moai memory doctor` 프로젝트 키 유도 + MEMORY.md 한도 단위

카드: t646 · 워크트리: `.claude/worktrees/t646` · 브랜치: `WT-memory-doctor-key`
베이스: 로컬 develop `9935e4e3e` (워크트리 HEAD가 이 SHA와 동일)

적용 규칙: `.claude/rules/moai/development/verification-completeness.md` §1.1(관측된 실패),
§1.2(3요소 명세), §2(뮤턴트 프로브) · `.claude/rules/moai/core/verification-claim-integrity.md`
§2(baseline 귀속), §3(5절 보고 형식)

---

## Claim

1. **C1 — 결함 확인.** 워크트리 세션에서 `moai memory doctor`가 두 저장소 모두
   `not present`로 보고하며, 리포지터리의 실제 저장소(613 topic files, 326 orphans)에 대해
   아무것도 말하지 않는다.
2. **C2 — 판정: git common dir 기준으로 키를 *교체*하면 안 되고, primary 키를 *추가*해야 한다.**
   배차문의 물음("키 유도가 git common dir 기준이어야 하는지")에 대한 답은 **조건부 예**다.
   교체는 기존 독트린이 명시적으로 금지한다(아래 §C2 근거).
3. **C3 — 수리.** 워크트리 세션에서 후보 저장소에 primary 체크아웃 키를 추가한다.
   cwd 키는 그대로 유지된다.
4. **C4 — MEMORY.md 읽기 한도 판정 단위는 바이트가 아니라 글자(rune)다.** 그리고 이 한도는
   moai가 아니라 Claude Code 하네스가 적용하므로 moai 쪽 수리 대상이 아니다.

---

## Evidence

### C1 — 재현 (primary ↔ 워크트리, 같은 바이너리 `~/go/bin/moai`)

primary 체크아웃, `CLAUDE_PROJECT_DIR` 미설정:

```
/Users/goos/.moai/claude-profiles/moai-adk/projects/-Users-goos-moai-moai-adk-go/memory  (CLAUDE_CONFIG_DIR)
  topic files : 613 (cap 50)
  index lines : 175
  findings    :
    MEMORY_ORPHAN_NOT_INDEXED      326
    MEMORY_TOPIC_COUNT_OVER_CAP    1
```

워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t646`, `CLAUDE_PROJECT_DIR` 미설정:

```
/Users/goos/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go--claude-worktrees-t646/memory  (CLAUDE_CONFIG_DIR)
  not present

/Users/goos/.claude/projects/-Users-goos-MoAI-moai-adk-go--claude-worktrees-t646/memory  (default ~/.claude)
  not present
```

`git rev-parse --git-common-dir` → `/Users/goos/MoAI/moai-adk-go/.git` — 같은 리포다.
결함 위치: `internal/cli/memory.go` `memoryCandidateStores`가 `resolveProjectDir()`
(= `CLAUDE_PROJECT_DIR` 또는 cwd)를 `filepath.Abs`만 거쳐 유일한 슬러그로 썼다.

> 대소문자 주의: primary 측정의 슬러그가 `-Users-goos-moai-…`(소문자)인 것은 셸 cwd가 그 철자로
> 잡혔기 때문이며, macOS 대소문자 무시 파일시스템이라 `-Users-goos-MoAI-…`와 **같은 디렉터리**다
> (양쪽 4개 경로 전수 확인: 4/4 EXISTS, MEMORY.md 크기 동일). 별개 결함이 아니다.

### C2 — 왜 "교체"가 아니라 "추가"인가

**(a) 워크트리 키 밑에는 메모리가 존재하지 않는다 (전수).**

```
/Users/goos/.moai/claude-profiles/moai-adk/projects : worktree-keyed dirs=278, of which have memory/=0
/Users/goos/.claude/projects                        : worktree-keyed dirs=1,   of which have memory/=0
--- control: non-worktree keys WITH memory/ --- 2
```

대조군 2건이 검출되므로 측정 방법이 0을 찍은 것이 아니다. 워크트리 키 디렉터리가 담는 것은
세션 트랜스크립트(`<uuid>.jsonl`)뿐이다.

**(b) 그러나 per-cwd 해석은 의도된 설계다 — 교체는 독트린 위반.**

`.moai/docs/memory-dir-resolution-doctrine.md` (Gap C, Status: Active) §1·§2:

> `resolveMemoryDir`는 세션의 raw CWD를 슬러그화한다. **git-root 정규화를 수행하지 않는다.**
> main repo 세션과 git-worktree 세션은 서로 다른 메모리 디렉터리를 해석한다. 이는 결함이 아니라
> Claude Code 네이티브 메모리 모델과의 정합이다. […] git-root로 강제 정규화하면 MoAI 쓰기가
> Claude Code의 네이티브 auto-load와 desync된다. 이는 "고침"이 아니라 회귀다.

같은 취지가 `internal/hook/session_end.go:207-214`의 `@MX:NOTE`에도 박혀 있다. 독트린 §2의
변경 금지 목록은 `resolveMemoryDir` / `projectSlug` / `resolve_memory_dir_test.go`를 명시하며
`internal/cli/memory.go`는 포함하지 않는다 — 그러나 §1의 원칙("per-cwd divergence는 결함이
아니다")은 일반 진술이므로, 읽는 쪽에서 키를 교체하면 그 원칙과 정면으로 부딪힌다.

독트린의 실증 근거("`~/.claude/projects/`는 cwd별 디렉터리를 보유한다")는 이번 측정으로
**재확인**됐다(278개). 즉 독트린은 이 점에서 낡지 않았다.

**(c) 그래서 채택한 해법은 같은 파일이 이미 쓰고 있던 원칙이다.**

`memoryCandidateStores`는 프로필 저장소와 기본 저장소를 **둘 다** 돌려주며, 그 이유를 주석이
직접 밝힌다 — *"A health check that reported only one would hide half the store."* 워크트리 키와
primary 키도 똑같이 실제로 공존하므로, 같은 원칙을 적용해 **둘 다** 보고한다. 읽기 전용 감사는
어디에도 쓰지 않으므로 독트린이 경계하는 write-path desync를 유발할 수 없다.

### C3 — RED → GREEN → 뮤턴트

RED (수리 전, 최초 설계안의 테스트):

```
memory_worktree_key_test.go:85: worktree name "wt-card" leaked into the store key:
  .../projects/-var-...-001-elsewhere-wt-card/memory
--- FAIL: TestMemoryCandidateStores_WorktreeResolvesToPrimaryCheckout (0.42s)
```

RED가 **맞는 이유로** 빨갛다 — 워크트리 디렉터리 이름이 키에 들어갔다.

GREEN (개정된 계약, 3개 테스트):

```
--- PASS: TestMemoryCandidateStores_WorktreeAlsoAuditsPrimaryCheckout (3.03s)
--- PASS: TestMemoryCandidateStores_PrimaryCheckoutIsUnchanged (0.29s)
--- PASS: TestMemoryCandidateStores_IdenticalRootsCollapse
```

뮤턴트 행렬 — 각 뮤턴트는 **적용됐는지 diff 줄 수로 먼저 확인**한 뒤 판정을 읽었다
(미적용 뮤턴트와 진짜 생존자는 같은 `ok`를 찍는다):

| 뮤턴트 | 적용 | 판정 |
|---|---|---|
| M1 primary 루트 추가 제거 (결함 복원) | 4줄 | KILLED |
| M2 추가 대신 교체 (독트린 위반형) | 4줄 | KILLED |
| M4 양변 정규화 불일치 비교로 되돌리기 | 6줄 | KILLED |
| M5 중복제거 제거 | 8줄 | KILLED |

M3(중복 short-circuit 제거)은 최초 실행에서 생존했고, 조사 결과 **뒤 검사가 흡수하는 중복
코드**임이 드러나 해당 검사를 삭제했다 — 테스트를 늘린 것이 아니라 코드를 줄였다.
M5는 최초 실행에서 생존했는데, 리팩터 이전 코드가 가지고 있던 중복제거 동작에 테스트가 없었기
때문이다. `IdenticalRootsCollapse`를 추가해 막았다.

**중간에 테스트가 잡아낸 실제 결함 1건**: 최초 구현은 `primary`(git이 심볼릭 링크를 푼 경로)를
raw `abs`와 비교해, 워크트리가 아닌 primary 체크아웃에서도 두 번째 저장소를 만들어냈다
(`worktree resolved 4 stores, primary checkout 4 — expected strictly more`). 양변을 같은
정규화로 비교하도록 수정했다.

### C3 — 수리 후 실제 바이너리 동작 (워크트리에서)

```
.../projects/-Users-goos-MoAI-moai-adk-go--claude-worktrees-t646/memory  (CLAUDE_CONFIG_DIR)
  not present
.../projects/-Users-goos-MoAI-moai-adk-go--claude-worktrees-t646/memory  (default ~/.claude)
  not present
.../projects/-Users-goos-MoAI-moai-adk-go/memory  (CLAUDE_CONFIG_DIR — primary checkout)
  topic files : 615 (cap 50)
  findings    : MEMORY_ORPHAN_NOT_INDEXED 326 / MEMORY_TOPIC_COUNT_OVER_CAP 1
.../projects/-Users-goos-MoAI-moai-adk-go/memory  (default ~/.claude — primary checkout)
  topic files : 1097 (cap 50)
  findings    : MEMORY_ORPHAN_NOT_INDEXED 808 / … / MEMORY_INDEX_OVERFLOW 1
4 stores resolved. …
```

`archive`의 기본 대상은 `stores[0]`이고 `stores[0]`은 여전히 cwd 키다 — **파일을 옮기는 경로의
동작은 불변**이다.

범위 검증:

```
gofmt -l internal/cli/                 : 0 files
go vet ./internal/cli/                 : exit 0
go build ./internal/...                : exit 0
GOOS=windows GOARCH=amd64 go build ./internal/... : exit 0
GOOS=linux   GOARCH=amd64 go build ./internal/... : exit 0
go test ./internal/cli/ -run 'TestMemory|TestArchive|TestDoctor' : 91 PASS, ok 283.673s
```

이 91에는 이 카드가 추가한 3건이 포함된다. 중간 측정에서 90이 나온 시점에는 최초 설계안의
테스트 2건만 트리에 있었다(88 + 2 → 90, 이후 +1 → 91). 두 수치는 서로 다른 트리를 잰 것이며,
90은 기준선이 아니다.

### C4 — MEMORY.md 한도 단위

대상: `~/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md`

```
bytes=38105   runes=29358   lines=175
bytes/1024=37.2      runes/1024=28.7
```

이 세션이 받은 하네스 경고문은 `MEMORY.md is 28.7KB (limit: 24.4KB)`였다.
**28.7 = runes/1024**이고 bytes/1024(37.2)가 아니다 → 판정 단위는 **글자(rune)**, KB는 1024 단위.
한도 24.4KB ≈ 24,986자.

한도의 주인:

```
grep "index entries are too long" internal/ pkg/ cmd/ .claude/ : 0 hits
control: grep "MEMORY.md" internal/                            : 139 hits
```

대조군 139건이 나오므로 grep이 트리를 읽었다 — 즉 이 절단은 **moai가 하지 않는다**.
Claude Code 하네스 동작이며 moai 쪽 수리 대상이 아니다.

---

## Baseline-attribution

모든 측정은 이 실행에서, 이 트리에 대해 수행했다.

- 트리: 워크트리 `.claude/worktrees/t646`, HEAD `9935e4e3e` (배차가 지정한 로컬 develop 베이스)
- `git merge-base --is-ancestor 9935e4e3e HEAD` → 성립
- 수정: `internal/cli/memory.go` (M), `internal/cli/memory_worktree_key_test.go` (신규),
  `.moai/reports/t646/verdict.md` (신규)
- go toolchain: go1.26.8 darwin-arm64
- C1의 primary 측정치는 워크트리 진입 **전** 같은 세션에서, 같은 커밋·같은 설치 바이너리로 수집
- C3의 실제 바이너리 출력은 이 트리에서 `go build -o /tmp/moai-t646 ./cmd/moai` 한 결과물

---

## Gaps (관측하지 않은 것)

- **`internal/cli` 전체 패키지 테스트는 로컬에서 판정하지 못했다.** 기본 10분 테스트 타임아웃에서
  중단(`FAIL … 600.808s`), 실패 테스트 0건(`--- FAIL` 0줄), 중단 지점은 `t.Parallel()` 대기.
  당시 load average 33.6/43.7 — 다른 레인 동시 실행 중. 전수 판정은 CI 몫이다.
- **Claude Code가 키를 어떤 문자열로 만드는지** 직접 관측하지 않았다. 워크트리 키 디렉터리
  278개의 존재로부터 cwd 기준이라고 **추론**했을 뿐이다.
- **독트린 §2의 변경 금지 대상은 건드리지 않았다** — `resolveMemoryDir` / `projectSlug` /
  `resolve_memory_dir_test.go` 모두 무수정. 다만 §1 원칙과의 정합은 해석이며, 리드 판단 필요.
- **MEMORY.md 색인 정리는 하지 않았다** — 배차문이 리드 소관으로 지정했다.

---

## Residual-risk

- **비대칭 정규화.** 리포 안에서는 git이 경로를 풀어 돌려주고, 리포 밖에서는 풀지 않는다.
  의도한 동작이고 M4가 지키지만, 동일 리포를 비-리포 경로로 지목하는 호출자는 다른 키를 얻는다.
- **출력이 길어진다.** 워크트리 세션에서 저장소 2개 → 4개. 대부분 `not present` 두 줄이라
  비용은 작지만, 스크립트로 `--json`을 파싱하는 소비자가 있다면 배열 길이가 달라진다.
  (현재 트리에 그런 소비자는 확인하지 못했다 — 미관측.)
- **독트린 해석.** 읽기 전용 감사는 write-path desync를 유발할 수 없다는 것이 이 수리의 전제다.
  독트린 §2가 CLI를 명시적으로 제외하지 않았다면 이 해석은 성립하지 않는다.

---

## 후속 (이 카드에서 하지 않음)

- **F1** — `internal/cli/preference/cmd.go:159` `memorySlug`도 같은 cwd 기준 유도를 쓰는지 확인,
  그렇다면 동일 취급(추가, 교체 아님).
- **F2** — 독트린 문서(`memory-dir-resolution-doctrine.md`)에 "읽기 경로는 양쪽을 모두 감사한다"는
  이번 결정을 §2 옆에 명문화할지 리드 판단. 지금은 코드 주석에만 있다.
- **F3** — MEMORY.md 한도 단위(글자)를 문서화할지. 현재 어느 moai 문서도 단위를 명시하지 않아
  다음 사람이 바이트로 읽는다.
