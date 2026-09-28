# moai-adk-go Local Development Guide

> **Purpose**: Essential guide for local moai-adk-go development
> **Audience**: GOOS (local developer only)
> **Last Updated**: 2026-09-29

---

## 0. [HARD] 이 파일의 정본은 어느 사본인가

이 절을 나머지보다 먼저 읽는다. 2026-09-07 에 이 파일의 틀린 사본이 레인 12곳에 배포됐고, 그 사고를 되풀이하지 않기 위한 판별식이 여기에 있다.

### 0.1 [HARD] 정본은 레인이 분기하는 트리의 사본이다

**판별식은 「레인이 분기하는 트리가 지배한다」이고, 현재 그 트리는 `develop` 이다.** 카드 워크트리가 `develop` 에서 나오므로 `develop` 의 사본이 레인이 실제로 읽는 문서이며, 그것이 정본이다.

**날짜나 「나중에 전달된 쪽」을 판별식으로 쓰지 않는다.** 2026-09-07 에 나중에 전달된 텍스트가 틀린 쪽이었다 — 최신성 규칙이었다면 그 사고를 막지 못했을 것이다. 분기 트리는 push-model 이 바뀌어도 같은 방식으로 답을 낸다.

### 0.2 [HARD] 미커밋 워킹 사본은 정본으로 인용할 수 없다

**어느 브랜치에도 커밋된 적 없는 미커밋 워킹 사본은 정본이 아니며, 정본으로 인용될 수 없다.** 이것은 이 파일에 한정된 규칙이 아니라 인용 일반의 규칙이다 — 이력에 없는 텍스트는 다른 사람이 같은 것을 읽었는지 확인할 방법이 없고, 저자도 시점도 복구되지 않는다.

2026-09-07 의 가해자는 스테일한 브랜치가 아니라 **git 이 추적하지만 어느 브랜치에도 커밋된 적 없는 primary 체크아웃의 워킹 사본**이었다. 인용하기 전에 그 텍스트가 어느 커밋에 있는지 확인한다 — `git show <ref>:<path>` 로 읽히지 않으면 정본이 아니다.

### 0.3 [HARD] `main` 의 커밋본은 폐기된 제3의 모델이다

**`main`에 커밋된 옛 `CLAUDE.local.md`는 폐기된 모델이며 인용 대상이 아니다.** 그 판은 「`develop`을 원격에 올리지 않고 카드마다 `main`으로 PR을 낸다」는 체제를 서술하는데, 현행 체제(§4.1)와 정면으로 다르다. 다음 사람이 `main` 사본을 정본으로 집는 것이 2026-09-07 실패의 재현이다.

### 0.4 [HARD] primary `main` 의 구형 로컬 파일은 정본이 아니다

이 이관이 `develop`에 병합되면 그 커밋의 `AGENTS.local.md`가 정본이다. primary 체크아웃이 아직 `main`인 동안 보이는 `M CLAUDE.local.md`는 전환 이전부터 유지한 공유 워킹 사본이다. 그 상태를 새 워크트리의 수신 증거로 인용하지 않는다.

> **[HARD] 이 표식은 정리 대상이 아니다.**
>
> **primary에서 `git restore CLAUDE.local.md`를 실행하지 마라** — 공유 워킹 사본을 §0.3의 폐기 모델로 되돌리는 회귀다.
>
> 이 파일을 고칠 때는 카드 워크트리에서 고쳐 `develop`으로 병합한다(§4.1). primary의 구형 파일은 별도 전환 전까지 보존한다.

---

## 1. Quick Start

See: `.moai/docs/local-dev-guide.md` § 1 — work locations, the edit→`make build`→test→commit cycle, and the [CRITICAL] `moai` CLI vs `/moai` slash-command distinction table. 핵심 한 줄: 템플릿은 `internal/template/templates/` 에서 고치고 `make build` 로 재생성하며, 커밋은 카드 워크트리에서 한다(primary 체크아웃 금지 — AGENTS.md §2·§3).

---

## 2. File Synchronization

### Protected Directories (Never Modify During Template Sync)
```bash
# CRITICAL: These directories contain user data and must NEVER be deleted
.claude/        # Local Claude Code configuration
.moai/project/  # Project documentation (product.md, structure.md, tech.md)
.moai/specs/    # SPEC documents (active development files)
```

### Template Source (Single Source of Truth)
```bash
# All template changes MUST be made here
internal/template/templates/.claude/
internal/template/templates/.moai/
# (.agency/ 템플릿 뿌리는 제거됨 — legacy archive; [2026-08-27 감사 정정])
internal/template/templates/CLAUDE.md
```

### [HARD] Template-First Rule

When adding new files to `.claude/`, `.moai/`, or `.agency/`:

1. **Add to template FIRST**: `internal/template/templates/<path>`
2. **Run `make build`** to regenerate embedded files
3. **Then sync to local**: `moai update` or manual copy

Never add files directly to the local project directories without also adding them to the template source. This includes:
- New agents (`.claude/agents/`)
- New skills (`.claude/skills/`)
- New commands (`.claude/commands/`)
- New rules (`.claude/rules/`)
- New config files (`.moai/config/`)
- New agency files (`.agency/`)

**Verification**: Before committing, check that every new file under `.claude/`, `.moai/`, or `.agency/` has a corresponding file in `internal/template/templates/`.

### [HARD] §2.0 에이전트 정의는 사본이 셋이고, 셋째는 손으로 고치지 않는다

에이전트 정의는 세 벌 있는데 **하나만 손편집 대상이 아니다.**

| 사본 | 경로 | 성격 |
|---|---|---|
| C1 | `.claude/agents/moai/*.md` | 로컬 도그푸드, 손편집 |
| C2 | `internal/template/templates/.claude/agents/moai/*.md` | 배포 미러, 손편집 — **중립 원본 층** |
| C3 | `internal/template/templates/.codex/agents/moai/*.toml` | **C2 로부터 기계 방출** (`internal/template/agentemit`) |

[HARD] **`internal/template/templates/.claude/agents/moai/*.md` 를 고쳤으면 `make agents-emit` 을 돌린다.** 그러지 않으면 `internal/template/templates/.codex/agents/moai/*.toml` 이 스테일한 채로 남고, 그 상태로 만든 바이너리가 옛 정의를 임베드한 채 돌아간다. C1↔C2 는 바이트 동일 관계가 **아니다**(의도된 분기) — 생성 관계는 C2 → C3 한 방향뿐이다.

[HARD] **C3 를 손으로 고치지 않는다.** 손편집은 다음 방출에서 말없이 덮인다. 고칠 것이 있으면 C2 를 고치고 재생성한다.

빠뜨렸을 때 어디서 잡히는가 — 세 지점이 각각 다른 축을 본다:

| 검사 | 무엇을 보는가 | 언제 |
|---|---|---|
| `make build` (선행 `agents-emit-check`) | 소스 층: 커밋된 `.toml` vs `.md` 방출 결과 | 로컬 빌드마다. **읽기전용 — 재생성하지 않는다** |
| `go test ./internal/template/agentemit/...` | 같은 축 | CI 매 실행 |
| `make embed-check` / `moai doctor --check "Agent Emit Embed"` | **임베드 축**: 이미 빌드된 바이너리가 실은 바이트 vs 커밋본 | 수동. `BIN=<path>` 로 설치본도 겨눌 수 있다 |

셋째가 따로 있는 이유: `go test` 는 테스트 바이너리를 매번 새로 컴파일하고 `//go:embed` 가 그 시점 커밋본을 읽으므로, 소스↔소스 비교로는 **노후한 바이너리를 원리상 볼 수 없다**. 그래서 `make embed-check` 는 `build` 를 선행으로 갖지 않는다 — 갓 빌드한 바이너리는 정의상 일치하므로 build 직후에만 도는 검사는 동어반복이다. 같은 이유로 CI 빌드 잡에도 붙이지 않는다(CI 는 자신이 검사하는 커밋에서 빌드한다).

재생성은 **`make agents-emit` 이라는 명시적 동사로만** 일어난다. build 가 조용히 덮으면 손편집이 있었다는 사실 자체가 사라진다.

**§2.1 Template Content Neutrality — Acceptable Content Range for Templates**: When editing template source files in `internal/template/templates/`, ensure content adheres to the **acceptable** kept-classes (C1/C2/C4/C5/C6/C8) and excludes the FORBIDDEN content classes (SPEC IDs, REQ tokens, Audit citations, internal dates, commit SHAs, macOS-bias paths, CLAUDE.local references), enforced by CI guard (`.github/workflows/template-neutrality-check.yaml` trigger on path change). The canonical C1-C8 acceptable-vs-forbidden content-class catalogue lives in `.moai/docs/template-internal-isolation-doctrine.md §25.1` (cross-referenced by **§25 (Template Internal-Content Isolation)** of this file, now a stub). This keeps the template neutral across the 16 supported **programming languages** (§15) and free of moai-adk internal development state. Distinct axis: user-facing locales (ko/en/ja/zh) are governed by the §8 Localization Contract in the active output style — see §15's disambiguation note.

**Pre-PR Verification (template contributor-checklist)** — before opening a PR that touches `internal/template/templates/**`, run the canonical 5-item pre-commit self-check (the CI guard `template-neutrality-check.yaml` is the safety net). See `.moai/docs/template-internal-isolation-doctrine.md` §25.3 for the full 5-item checklist and §25.1 for the forbidden/allowed content-class catalogue (C1-C8). (C3 dates + C7 commit-hashes are owned by the sibling `internal_content_leak_test.go` per §25, not this neutrality checklist.)

### Local-Only Files (Never in Templates)

전체 레지스트리(30+ 항목, 경로·사유·이력 포함): `.moai/docs/local-dev-guide.md` § 2a. 핵심만: `settings.local.json`(런타임 관리) · `.claude/rules/local/*`(dev-only 룰) · `scripts/{ci-watch,ci-autofix,jev,ac-baseline}/` · `.moai/{cache,logs,state,specs,plans,reports,docs} 내용물` · `.moai/docs/*.md` dev-only 문서 전체 — **전부 템플릿 미러가 없어야 하는 파일이며, §2.3의 관리 대상 뿌리 밖에 두거나 뿌리 안에서도 삭제되지 않는 위치에 둔다.**

### [HARD] §2.3 moai update는 관리 대상 뿌리 안의 로컬 전용 파일을 통째로 삭제한다

**요지 — 본문 전량은 `.moai/docs/update-local-file-survival.md` 로 이관됐다(card t750).** `CleanMoaiManagedPaths`(`internal/cli/update/deploy/deploy.go:107`)가 템플릿 재배포 **전에** 관리 대상 뿌리(`.claude/settings.json` · `.claude/{commands,agents,hooks}/moai` · `.claude/skills/moai*` 글롭 · `.claude/rules/moai` · `.claude/output-styles/moai` · `.moai/config`)를 통째로 삭제하고 임베드 템플릿에 있는 것만 다시 깐다. **보호 목록 설정은 존재하지 않고**, `Updated N files` 요약에 삭제는 나타나지 않는다. **[HARD] 새 로컬 전용 파일은 위 뿌리 밖에 둔다** — 용도별 배치 표(룰·스킬·ast-grep·하네스)는 이관 문서에 있다. **[HARD] update 후엔 매번** ① 삭제 검증(`git status --porcelain | grep '^ D'` — 0이어야 정상)과 ② `git-strategy.yaml` git-flow 키 **+ `worktree_base_branch: develop`** 재적용을 실행한다(후자를 빼면 `moai worktree new` 가 카드 트리를 develop 이 아니라 main 에서 판다 — 2026-09-24 6건, card t1159). 재적용은 `--source=develop` 이다(`HEAD`=main 에는 그 키가 없다) — 명령과 실측 근거는 이관 문서에.

### [HARD] settings.local.json Separation

`settings.local.json` is **runtime-managed**. Never put it in templates.

- Modified by `moai glm`, `moai cc`, `moai cg` commands at runtime
- Modified by SessionStart hook (GLM credentials, teammateMode, CLAUDE_ENV_FILE)
- Contains per-machine values: tmux pane IDs, API tokens, absolute paths
- **Never** add effortLevel, teammateMode, or env tokens to the template

If you accidentally commit `settings.local.json`, run `git rm --cached .claude/settings.local.json`.

### [WARN] OpenTelemetry / OTEL in Tests

Do NOT use `t.Setenv` with OTEL environment variables (`OTEL_EXPORTER_*`, `OTEL_SERVICE_NAME`) in tests. Setting these in parallel tests causes data races because the OTEL SDK initializes global state from env vars on first use.

- Use a fake/no-op exporter instead of env-var configuration in tests
- If the test must set OTEL vars, make the parent test non-parallel and use `t.Setenv` only in non-parallel subtests

### Embedded Template System · Command-to-Skill Publication

본문은 `.moai/docs/local-dev-guide.md` § 2b·2c 로 이관됐다(card t1303). 요지: `internal/template/embed.go` 가 `//go:embed all:templates` + `catalog.yaml` 를 바이너리에 컴파일하므로 템플릿 편집 후 `make build` 필수; 16개 `/moai` 커맨드의 codex skill 방출은 `make commands-emit`(명시적 동사)으로만 재생성하고 방출본(`.agents/skills/moai-*/SKILL.md`)은 손으로 고치지 않는다.

---

## 3. Code Standards

Language policy는 `.claude/rules/moai/development/coding-standards.md`에 정의 (auto-loaded).

### Go-Specific (이 프로젝트 전용)

- File naming: `snake_case.go`, `snake_case_test.go`
- Error wrapping: `fmt.Errorf("operation: %w", err)` (string concatenation 금지)
- All code, comments, godoc in English

---

## 4. Git Workflow

### Before Commit
- [ ] Code in English
- [ ] 변경 대상 패키지 테스트 통과 (`go test -timeout 30m ./internal/<pkg>/...`; `-timeout 30m` = D2 derivation — 로컬 단일 패키지 최악 1118.093s 실측 대비 1.61x 여유, baseline `.moai/reports/t1253/measure-meta.txt`, SPEC-CLI-TEST-TIMEOUT-001) — **전체 스위트(`go test ./...`)를 로컬에서 돌리지 않는다**. 레인 여러 개가 동시에 돌려 load 413까지 치솟고 다른 워크스페이스를 마비시킨 사고(2026-08-15)가 있다. 전 패키지 판정은 CI 몫이며, 깨끗한 환경에서 PR head를 돌리므로 근거로도 더 강하다. 예외는 §4.1의 통합 검증 — 그때는 **직렬로 1건씩**
- [ ] Linting passing (`golangci-lint run`)
- [ ] Templates regenerated (`make build`)

### Before Push
- [ ] Branch rebased
- [ ] Commits organized
- [ ] Commit messages follow format (Conventional Commits)

### Commit Message Format

Conventional Commits — `<type>(<scope>): <description>` + 선택 본문/푸터. Types: feat, fix, docs, style, refactor, perf, test, chore, revert. 예시와 서식 전문: `.moai/docs/local-dev-guide.md` § 4a.

### §4.1 GitFlow 통합 체인 (develop)

카드별로 각각 검증해 머지했는데 **합쳐진 상태는 아무도 보지 않는** 구멍을 막는다. 2026-08-15에 PR 12개가 각각 초록불로 main에 들어갔고, 합류 후에야 `moai update`가 로컬 전용 파일을 지운다는 사실이 드러났다.

**[HARD] 표준 체인 — 운영자 지시 2026-08-29**

```
origin/main
   ↓ fetch
local/main
   ↓ 분기
develop  ← 통합 브랜치. 원격에 존재(origin/develop)
   ↓ 분기
worktree (카드별 WT-<slug>)
   ↓ merge --no-ff  ← 카드 완료 시 반드시 develop으로
develop
   ↓ 분기
release/vX.Y.Z
   ↓ PR
origin/main
   ↓ fetch
local/main
```

**[HARD] 규율**

1. **모든 워크트리는 develop으로 병합한다.** 카드가 끝나면 `main`이 아니라 `develop`에 합친다. 카드 브랜치는 develop에서 판다.
2. **`develop`은 원격에 있다.** push는 **리드가 일괄**로 수행하며(2026-09-02 — 레인은 push하지 않는다), 그 head의 CI가 통합 판정을 만든다.
3. **main으로는 release 브랜치의 PR만 올라간다.** 카드가 직접 main으로 PR을 내지 않는다.
4. **통합 창은 직렬이다.** `moai integration acquire --card <card-id>` → 병합 → `release` — push는 창 밖이다(리드 일괄, 2026-09-02). 락은 병합을 직렬화하는 장치이지 수리를 직렬화하는 장치가 아니므로, 수리가 남았으면 준비된 뒤에 잡는다.
5. **판정은 CI.** 로컬 통과는 조기 신호일 뿐이다 — 깨끗한 환경도, darwin/windows 매트릭스도 아니다. 병합 전 검증을 병합 후 근거로 재사용하지 않는다: 병합 트리에서 다시 재거나, 병합 커밋의 `git rev-parse <merge>^{tree}`가 재측정한 트리와 동일함을 보인다.

**[SUPERSEDED by 위 체인 — 2026-08-29]** 종전 규위(develop 원격 미푸시 · 카드별 main PR · 일회용 develop)은 폐기됐다 — 폐기 사실과 사유의 보존은 두 문서가 반대 지시를 하지 않게 하기 위함이며, 전문은 `.moai/docs/gitflow-integration-chain.md` 에 있다.

**관련 문서 포인터 (SPEC-RC-TESTBED-001)** — 절차 본문은 두지 않는다(위 규율 4의 delivery.md 위임과 마찬가지로, 두 벌이 되는 순간 갈라진다):

- 로컬 rc 빌드의 `rc.N` 번호 정책·무태그 원칙·`BUILD_ID` 빌드 식별: `.moai/docs/version-management.md` — **Local RC Numbering** 절
- 병합 후 로컬 develop 갱신(판정 기준·BranchGuard 안전 경로): `.claude/rules/local/gitflow-lane-protocol.md` — **develop 갱신** 절 (그 절이 §9 rc 런북을 교차참조한다)

**[HARD] -k / -f 모드 레인 의무**

Kanban(`moai cc -k`) / Factory(`moai cc -f N`) 모드에서 레인은 카드 작업이 끝나면 **반드시 리드에게 로컬 develop 병합을 요청한다.** 레인이 스스로 병합 창을 잡지 않는다.

- 완료 보고에 담을 것: 카드 id · 브랜치와 HEAD · 로컬 병합 SHA · 미푸시 커밋 수 · 증거 경로(primary 반출 여부) · 재측정 범위
- `moai integration status`가 `free`인 것은 **승인이 아니다.** 리드의 창 지명만이 근거다.
- 창을 받으면: `moai integration acquire --name <lane> --card <card-id>` → 본인 워크트리에서 `git merge develop` 흡수(대상은 **로컬** `develop` — 원격이 아니다. 흡수 **전에** 그 로컬 develop 이 최신인지부터 본다 — 판정식과 갱신 경로는 `.claude/rules/local/gitflow-lane-protocol.md` §11) → **병합 트리에서 재측정** → `EnterWorktree(.claude/worktrees/develop)` → `git merge --no-ff <WT-브랜치>` → `moai integration release` → `ExitWorktree keep` → 완료 보고(로컬 병합 SHA를 리드에게 보고 — push는 리드가 일괄로 한다)
- **[HARD] WT 브랜치 push·CI 직접 요청 금지 (운영자 지시 2026-09-01).** 카드가 마감되면 원격 develop 반영이 **유일한** 공개 경로다 — 리드가 창 밖에서 레인 병합 SHA를 모아 일괄로 실행하는 `git push origin develop`이며, 레인은 그 push의 주체가 아니다. 레인은 `git push origin <WT-브랜치>`를 하지 않고, `gh run rerun`/`workflow dispatch` 등 CI를 직접 요청·재요청하지도 않는다 — CI 판정은 develop push가 일으키는 실행에 맡기고, 판독은 리드 몫이다. (당일 lane-2가 `WT-version-stamp-predicate`를 origin에 push한 전례로 추가)
- **[HARD] `acquire`는 창을 기록하기 전에 호출자 트리를 먼저 단정한다.** tracked `.claude/settings.json`의 워킹 사본이 수정돼 있는지 `git --no-optional-locks status --porcelain -- .claude/settings.json`으로 재고, 적중이면 그 사본을 primary 체크아웃의 `.moai/state/settings-drift/` 아래로 보존한 뒤 같은 자리 `ledger.jsonl`에 한 줄을 남기고 보존 경로·sha256을 출력한다. **검출·보존·원장은 설정과 무관하게 매번 돈다**(9일 동안 아무도 보지 않아서 놓친 것이 문제였지 막지 않아서가 아니다). 거절만 opt-in이며(`workflow.settings_drift_gate.enabled`, 이 저장소는 켠다) 우회는 `--allow-settings-drift`다 — `--force`는 "살아 있는 보유자에게서 창을 빼앗는다"는 다른 축이라 우회로 쓰지 않는다. 창과 무관하게 손으로 확인할 때는 `moai integration preflight [경로]`. **어떤 경우에도 자동 복원하지 않는다** — 그 파일은 런타임이 쓰고 토큰·절대경로·tmux pane id를 담을 수 있어 자동 복원 자체가 데이터 파괴다. 적중 보고를 받으면 리드가 처분을 정한다.
- **워크트리는 원격 머지가 확인되기 전까지 폐기하지 않는다.** 미푸시 브랜치의 워크트리는 그 작업의 유일본이다.
- sync는 병합 **전에** 워크트리 안에서 끝낸다. run만 닫고 병합하면 SPEC이 `in-progress`로 develop에 올라가 창을 다시 받아야 한다(2026-08-29 t342 실사례).

**운영 절차 이하는 `.moai/docs/gitflow-integration-chain.md` 로 이관됐다(card t750) — 창 집행 bash(통합 워크트리 진입·흡수·재측정·병합), 리드 develop 일괄 push 절차, 로컬 CI 기각 기록, BranchGuard 조회 과다 매칭 마찰, docs-site Vercel 바인딩 주의.** 위 [HARD] 규율과 레인 의무가 변하지 않는 한 이 요지로 충분하고, 절차를 실행할 때 이관 문서를 연다.

---

## 5. Version Management

See: `.moai/docs/version-management.md` — SemVer 2.0.0 pre-release form (`-rc.N`), build version injection via ldflags, files requiring version sync, release process under the PR-mandatory regime.

---

## 6. Testing Guidelines

### ⚠️ IMPORTANT: Prevent Accidental File Modifications

When running tests, **always check if they modify project files**.

### Test Isolation

**[HARD] All test temp directories MUST be created under `/tmp` and cleaned up automatically** — `t.TempDir()` for every temporary directory (auto-cleanup registered).

**[HARD] 함정 — `filepath.Join(cwd, absPath)`는 두 번째 인수의 선행 `/`를 벗기지 않는다**: `filepath.Join("/a/b", "/var/folders/x") = "/a/b/var/folders/x"` (오동작). 사용자 제공 경로 해석엔 `filepath.Abs()` 를 쓴다. 실측 예시와 코드 전문: `.moai/docs/local-dev-guide.md` § 6a.

### Coverage Targets

본문은 `.moai/docs/local-dev-guide.md` § 6a 로 이관됐다 — 패키지 85%(`quality.yaml` `test_coverage_target`), critical(cli·template·hook) 90%+는 목표치이며 기계적 근거는 strict 프로필의 전역 게이트뿐이다.

### Go Test Execution Rules

- [HARD] After fixing ANY test, run the AFFECTED packages only and let CI give the full-suite verdict — the command, its `-timeout` derivation, and the no-local-full-suite rule live in §4 Before Commit (single source, not restated here)
- Do not declare success after fixing only the initially failing tests
- Run `go test -count=1 ./internal/<pkg>/...` to disable test caching when debugging flaky tests
- Run `go test -race ./internal/<pkg>/...` for concurrency safety on any code touching goroutines or channels
- Run `go vet ./internal/<pkg>/...` on the changed packages before committing

---

## 7. Hook Development Guidelines

See: `.moai/docs/hook-development.md` — shell-script-only hook pattern, hook wrapper template, settings.json format, `$CLAUDE_PROJECT_DIR` quoting rules, platform differences, timeout policy.

---

## 8. Template Variable Strategy

본문은 `.moai/docs/local-dev-guide.md` § 8 로 이관됐다(card t1303). 요지: 템플릿(`*.tmpl`)은 `{{.GoBinPath}}`류 Go 템플릿 변수(`moai init` 시 렌더링), 로컬 설정은 `$CLAUDE_PROJECT_DIR`(Claude Code가 실행 시 해석) — 경로 변수 전략이 둘 다 필요한 이유와 `TemplateContext`·`Deploy` 서명은 이관 문서에. **[HARD] `.sh.tmpl` 폴백엔 `.HomeDir` 금지 — `$HOME` 사용(§14와 동일 축).**

---

## 9. Configuration System

본문은 `.moai/docs/local-dev-guide.md` § 9 로 이관됐다(card t1303). 요지: 단일 `config.yaml`은 없다 — 설정은 `.moai/config/sections/*.yaml` 32개 파일로만 존재한다. 우선순위: env var(`MOAI_DEVELOPMENT_MODE`·`MOAI_LOG_LEVEL`·`MOAI_LOG_FORMAT`·`MOAI_NO_COLOR` — 상수는 `internal/config/envkeys.go`) > sections/*.yaml > 템플릿 기본값. `MOAI_USER_NAME`/`MOAI_CONVERSATION_LANG`은 미구현이다.

---

## 10. Build and Development Commands

빌드/테스트/린트 타깃은 `Makefile`에 `##` 도움말과 함께 정의돼 있다 — `make help` 로 조회. 템플릿 편집 → `make build` → 테스트 → 커밋 순서는 §2 [HARD] Template-First Rule 참조.

---

## 11. Frequent Issues and Solutions

자주 발생 4건 — 한 줄 요지와 상세는 `.moai/docs/local-dev-guide.md` § 11a: 템플릿 미갱신(`make build` 필수) · 테스트가 실설정 오염(`t.TempDir()` 격리) · 훅 타임아웃(`{"timeout": 60}`) · **exit 137 재설치 잔재**(반드시 `rm -f ~/go/bin/moai && cp bin/moai ~/go/bin/moai` clean 재설치 + `sh scripts/verify-local-install.sh` 검증; 맨손 `go install` 금지 — `LDFLAGS` 누락이 provenance 검증을 무력화한다).

### [HARD] 사용 중 버그·개선 발견 → 즉시 `/moai:feedback`

MoAI-ADK를 사용하다 버그나 개선이 필요한 부분을 발견하는 족족 `/moai:feedback`으로 피드백을 제출한다 — 세션을 마친 뒤 몰아서 남기지 않는다. 대상: `moai` CLI 동작, 훅, 템플릿, 스킬, 에이전트, 팩토리·칸반 운영 결함 전반. 재현 명령과 관측된 출력을 함께 남긴다. 구분: 유지자에게 보고할 사안은 `/moai:feedback`, 작업으로 예정할 사안은 `/moai todo add`.

---

## 12. YAML Frontmatter 빠른 참조

범용 형식 규칙은 `.claude/rules/moai/development/` 내 `skill-authoring.md`, `agent-authoring.md`에 정의. 로컬 개발 체크리스트(CSV 금지·`skills:` 배열 예외·`metadata.*` 인용·`make build`·로컬 동기화)와 탐지 스크립트 위치는 `.moai/docs/local-dev-guide.md` § 12 로 이관됐다(card t1303).

---

## 13. GLM Integration Testing

### [HARD] Dev 프로젝트에서 GLM 통합 테스트 실행 금지

`moai cc`/`moai glm` 커맨드 플로우는 실제 settings 파일을 수정하므로 dev project에서 절대 실행 금지.

- Unit tests: dev project, 변경 패키지만(`go test ./internal/<pkg>/...` — 로컬 전체 스위트 금지, §4), `t.TempDir()` 내 파일만
- Integration tests: `/tmp/test-project`에서 `claude -p`로 실행
- Auth token: `loadGLMKey()` (reads `~/.moai/.env.glm`), 없으면 `t.Skip()`
- 금지: `t.Setenv("HOME", tmpDir)` (병렬 테스트 오염), 하드코딩 fake key

---

## 14. 하드코딩 방지

### [HARD] Go 코드 (internal/, pkg/) 하드코딩 금지

- URL, 모델명, 조직명, API 헤더 → `const`로 추출
- 환경변수명 → `internal/config/envkeys.go`에 상수 정의 후 참조
- 임계값 → `config/defaults.go`에 단일 원천 정의, 중복 금지
- 크로스 플랫폼 → `$HOME`, `HOMEBREW_PREFIX` 등 환경변수 우선

### [HARD] .sh.tmpl 폴백 경로에 `.HomeDir` 금지

`.HomeDir`/`.GoBinPath`는 `moai init` 시점의 절대 경로로 굳어짐. 폴백에는 `$HOME` 사용:
- Primary: `{{posixPath .GoBinPath}}/moai` (OK, init-time)
- Fallback: `$HOME/go/bin/moai` (MUST use `$HOME`)
- `renderer.go`: `$HOME`은 `claudeCodePassthroughTokens`에 이미 등록

### 하드코딩 허용 영역

`AGENTS.local.md`, `settings.local.json`, `_test.go` (t.TempDir() 내).

---

## 15. 템플릿 프로그래밍-언어 중립성

> **[HARD] 용어 — "언어"는 두 축을 가리킨다. 수식어 없이 쓰지 않는다.**
>
> | 축 | 값 | 요구되는 것 | 지배 규칙 |
> |---|---|---|---|
> | **프로그래밍 언어** (16) | go, python, typescript, … swift | **중립** — 어느 하나를 PRIMARY로 두지 않음 | 이 §15 |
> | **사용자 대화 로케일** (4) | ko, en, ja, zh | **번역** — 로케일마다 자연스러운 원어 | 활성 output style §8 Localization Contract |
>
> 이 둘은 서로 독립이다. 템플릿은 16개 프로그래밍 언어에 중립이면서, 동시에 4개 로케일로 번역된다.
> "16개 언어"를 배포 로케일 수로 읽는 것이 반복 관측된 오독이므로, 문서에서는 항상
> **"프로그래밍 언어"** 또는 **"로케일"** 로 수식해 쓴다.

### [HARD] `internal/template/templates/` 하위는 16개 프로그래밍 언어를 동등 취급

도구의 구현 언어(Go)와 사용자 프로젝트의 프로그래밍 언어는 별개. 템플릿은 모든 사용자를 위한 것.

- 프로그래밍-언어 편향 허용: `AGENTS.local.md`, `settings.local.json`, 로컬 `.moai/config/`
- 프로그래밍-언어 편향 금지: `internal/template/templates/**` 전체

### 16개 지원 프로그래밍 언어 (모두 동등)

```
go, python, typescript, javascript, rust, java, kotlin, csharp,
ruby, php, elixir, cpp, scala, r, flutter, swift
```

Dart/Flutter 캐논 이름: **"flutter"** (not "dart").

### 체크리스트 (템플릿 수정 시)

- [ ] 특정 언어를 "PRIMARY"로 배치하지 않았는가?
- [ ] 16개 프로그래밍 언어가 동등 수준으로 나열되어 있는가?
- [ ] 특정 언어만 "enabled", 나머지 "planned"로 격하하지 않았는가?
- [ ] project_markers 기반 자동 감지 로직이 포함되어 있는가?
- [ ] 로컬 config와 템플릿이 달라도 정상 (같으면 오히려 의심)

상세 교훈: 전역 프로젝트 메모리의 `lessons.md` #5 — 읽는 법: 그 저장소(`~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/`)는 **휴면**이라 세션이 로드하지 않는다. 직접 열어 읽고, 두 저장소 확인은 `moai memory doctor`.

---

## 16. 오케스트레이터 자가 점검

### [HARD] 자가 점검 4 질문 (복잡 작업 시작 전 필수)

1. 이 작업은 전문 에이전트의 고유 도메인인가?
2. 해당 전문 에이전트가 카탈로그에 존재하는가? (CLAUDE.md Section 4)
3. 직접 수행보다 위임이 품질/독립성/편향 방지에 유리한가?
4. 이 작업의 일부를 read-only sub-agent 병렬 spawn으로 분해할 수 있는가? (Anthropic "Exploration-First Pattern": WebFetch + Explore subagent로 분석 → main agent로 종합)

**3개 이상 YES → 직접 수행 금지**. 4번째 질문도 YES인 경우 Exploration-First Pattern (read-only sub-agent 병렬 spawn) 우선 적용. AskUserQuestion으로 위임 방식 확인 후 실행.

### 수량 기반 트리거

- 같은 종류 파일 **5+** 생성 → `manager-develop` 또는 `builder-harness` 위임 강제
- Go 코드 **500+ LOC** 신규 → `manager-develop` (cycle_type=tdd, domain context: backend) 강제 — 과거 `expert-backend`는 archived per SPEC-V3R6-AGENT-TEAM-REBUILD-001
- 에이전트/스킬 **3+** 생성 → `builder-harness` 강제

### 허용되는 직접 수행

Typo/포맷 수정, 설정 1개 편집, 사용자 명시 요청, 위임 대상 부재, 오케스트레이션 자체, git 작업, `/tmp` 작업.

### 순서: Rule 5 → §16 → Rule 1

Rule 5(WHAT) → §16(WHO) → Rule 1(HOW) → 실행

상세 교훈 및 5 Whys: 전역 프로젝트 메모리의 `lessons.md` #4 — **휴면 저장소**라 세션이 로드하지 않는다(§15의 읽는 법과 동일: 직접 열어 읽고, `moai memory doctor` 로 두 저장소 확인).

---

## 17. docs-site 4개국어 문서 동기화 규칙

docs-site는 `adk.mo.ai.kr` 공식 사용자 문서. URL 표준, 4-locale 동기화 의무, Mermaid TD-only, Vercel 프로젝트 바인딩, 빌드/배포 체크리스트 등 전체 doctrine은 외부 파일 참조.

See: `.moai/docs/docs-site-i18n-rules.md`

### §17.1 디자인 컴포넌트 + 아이콘 규약 → See `.moai/docs/docs-site-design-components.md`

docs-site Claude Warm Editorial [HARD] 세부(아이콘 shortcode·코드블록·Mermaid·푸터·CSS 캐시 버스팅·라이트 단일 테마)는 외부 파일로 이관. 핵심만: 본문 장식 이모지 금지→`{{</* icon */>}}` shortcode; 사이드바 `icon:` 값은 `menu.html` SVG case 필수; CSS 수정 후 dev 반영 안 되면 hugo 재시작.

---

## References

Sections §18-27 were consolidated into external `.moai/docs/` files to reduce launch-time context. Each entry below is the authoritative location for its domain.

- **§5 Version Management** (SemVer pre-release, ldflags injection, release process): `.moai/docs/version-management.md`
- **§7 Hook Development** (shell-script-only pattern, settings.json format, quoting rules): `.moai/docs/hook-development.md`
- **§18 Git Workflow** (Enhanced GitHub Flow 본문 + [2026-08-27] git-flow 전환 상위모델 노트 — 정본은 §4.1, branch protection `enforce_admins: true`, Hybrid Trunk RETIRED): `.moai/docs/git-workflow-doctrine.md`
- **§19 AskUserQuestion Enforcement + §19.1 Implementation Kickoff Approval Mandatory Restoration** (REQ-ATR-015): canonical SSOT at `.claude/rules/moai/core/askuser-protocol.md` + `.claude/rules/moai/workflow/orchestration-mode-selection.md` §E (the gate is mandatory and score-independent; plan-auditor PASS never auto-bypasses it)
- **§20 Vercel Build Cost Guard** [HARD]: all Vercel projects MUST use Elastic build machine ($0.0035/CPU min vs Turbo $0.126/min); check Build Machine setting first on cost anomalies
- **§21 Dev-Only Commands Isolation** (split harnesses, `SPLIT_HARNESS_NAMESPACE_LEAK` sentinel): `.moai/docs/dev-only-commands-isolation.md`
- **§22 Dev Settings Intent** (settings.json key semantics): `.moai/docs/local-dev-settings-intent.md`
- **§23 Local Git Workflows** (PR-mandatory 1-person OSS, all tiers via PR — 릴리스 PR 경로; [2026-08-27] 카드 작업은 git-flow `develop` 병합으로 전환, 상위모델 노트 참조): `.moai/docs/git-local-workflow-doctrine.md`
- **§24 Harness Namespace** (template-managed vs user-owned separation): `.moai/docs/harness-namespace-doctrine.md`
- **§25 Template Internal-Content Isolation** (neutrality catalogue, CI guard): `.moai/docs/template-internal-isolation-doctrine.md`
- **§26 Linear 연동** (local-only): `.moai/docs/local-linear-integration.md`
- **§27 Agent-Skill Architecture** [HARD]: every agent gets ≥1 skill set (4 elements: workflow skill + knowhow reference + scripts + trigger); all skill bodies in English; `/moai:<sub>` slash-wrapping maintained; `/moai:harness` meta-harness (v4 Builder); Analyze-First execution plan before any work

---

## 28. LSEL 드레인 운영 (지역 자가진화 루프)

[HARD] `.claude/settings.local.json`의 SessionStart 배선이 `session_drain.sh`와 `backlog_check.sh`를 실행한다. `drain.sh`를 직접 호출하지 말고 보존 래퍼 `session_drain.sh --inbox .moai/lessons-inbox.jsonl --state-dir .moai/state/lsel`만 쓴다. 후보 제안은 휘발성 live 파일이 아닌 `clusters-history/` 보존본에서 읽는다. 인박스는 도구·테스트 실패 스텁에 한정된다. 배선·검증·학습 경계는 `.moai/docs/lsel-drain-operations.md`와 `.moai/docs/learning-channel-scope.md`에 있다.

---

## 29. Jev (TypeSafe System One) — 로컬 전용

[HARD] `scripts/jev/`는 이 저장소의 로컬 도구이며 제품에 배선되지 않았다. `-k`/`-f` 리드는 묵은 카드 배차 전에 `scripts/jev/triage.sh <id>`, 레인 질문으로 멈췄을 때 `scripts/jev/route.sh < 질문`을 자율 실행한다. 출력은 판단 자료일 뿐이다. 완료·병합·큐 변경·운영자 게이트에는 Jev를 판정 근거로 쓰지 않는다. 키는 `~/.moai/.env.typesafe`에만 두고, 외부 전송 전에 카드의 비밀·고객 데이터를 확인한다. 키나 네트워크가 없으면 독트린과 직접 읽은 증거로 판단한다. 0.50 신뢰도 문턱은 잠정값이다. 등급·명령·측정 한계는 `.moai/docs/jev-local-operations.md`에 있다.

## 30. 배차 전 전제 판정 (며칠 지난 카드)

[HARD] 묵은 카드의 전제를 배차 전에 `git log --all -S` 심볼 교차, `git merge-base --is-ancestor` 귀속, 수리 이전 커밋의 양성 대조로 순서대로 잰다. 양성 대조 없는 무출력은 부재가 아니라 미측정이다. 모델 신뢰도 점수로 카드를 자동 취소하거나 전제 소멸을 확정하지 않는다. 소멸 근거는 판정서에 남기고 큐 변경은 운영자 판정을 따른다. 측정법과 기각 실험은 `.moai/docs/stale-card-premise-check.md`에 있다.

## 31. 킥오프 자율·의사결정 위임 (운영자 정책)

[HARD] 킥오프는 카드가 운영자 게이트를 명시하지 않는 한 자율로 진행한다. 레인 창에서 운영자가 이미 고른 모드는 덮지 않는다. 레인은 선택을 운영자에게 되묻지 않고 리드에게 올린다. 리드는 Jev를 쓰더라도 완료·병합·큐 변경·운영자 게이트를 모델 답으로 판정하지 않는다. 근거와 적용 범위는 `.moai/docs/kickoff-autonomy.md`에 있다.
