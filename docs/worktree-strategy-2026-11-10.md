# 에이전트 병행 작업용 worktree 전략 — 1단계 설계 (2026-11-10)

Issue #2380 (rc.2 WT). 사용자 요청 2026-11-10T10:15Z. 작성: agent-5-t64 (claim `2380#6764771090`).

## 0. 범위와 선언

- **설계 문서만 다룬다.** 구현, 삭제, 정리 실행, host 설정·권한 모드 변경은 하지 않았다.
- **읽기 전용 측정.** 이 저장소와 `/tmp`는 읽기만 했다. `du`는 호출마다 시간 상한(20~280초)을 걸었다.
- **실험 위치.** 실험은 이 worktree의 무시되는 `target/agent5-2380/` 아래 임시 저장소와 합성 crate에서만 했다.
- **측정 시각과 기준.** 측정은 2026-11-10T10:19Z~10:50Z에 했다. 기준 tip은 `f259f15a14`이고, 실험 중 GitHub tip이 `8536b46af8`로 움직였다.
- **도구 버전.** cargo 1.94.1 (2026-03-24), git 2.54.0 (Apple Git-157), macOS arm64.
- **판정 표기.** 공식 문서를 실제로 읽은 것은 "확인", 요약본만 받은 것은 "요약(미검증)", 읽지 못한 것은 "unverified"로 적는다. 숫자는 모두 이 저장소와 실험에서 잰 값이다.
- **결정권.** 결정안은 레인의 제안이다. 사용자 결정이 필요한 항목은 §5 질문으로 따로 모았다.

## 1. 실측: 공간은 어디로 가는가

### 1.1 요약

**사용자 전제("불필요한 파일까지 커밋된 것을 써서 용량이 낭비된다")는 실측과 맞지 않는다.** 추적 파일은 worktree 하나에 약 11 MiB(그중 `.moai-cli` 5.5 MiB)이고, 낭비 전체의 1% 미만이다. 공간의 대부분은 빌드 산출물과, 모든 이력을 받는 감사 clone이 차지한다.

| 분류 | 개수 | 용량 | 비고 |
|---|---|---|---|
| `/tmp/moai-rc-target-*` (세션별 `CARGO_TARGET_DIR`) | 57 | **65.9 GiB** | worktree보다 오래 남는다 |
| `/tmp`의 감사 clone (`moai-audit-<n>*`, `moai-t6b-audit*` 등) | 86 | **39.6 GiB** | clone 하나의 `.git`이 약 470 MiB |
| `.moai-cli/worktrees/*/target/` | 39 | **27.3 GiB** | 큰 것은 대부분 cargo `debug/`(2.8~3.3 GiB), 일부에 감사 clone이 섞임 |
| `.moai-cli/worktrees` 전체(위 `target/` 포함) | 47 | 31.8 GiB | `target/`을 빼면 checkout은 약 4.5 GiB |
| worktree 하나의 checkout(`target/` 제외) | — | 8~11 MiB | 표본 8개 |
| `/tmp/moai-*` 전체 | — | 142.5 GiB | 위 분류를 포함 |
| 빈 임시 폴더 `moai-cli-home-*`·`-profiles-*`·`-gitconfig-*`·`moai-audit-receipts-*` | 903 × 4 | 0 | 11-06~11-08에 생성. 현행 `src`·`tests`·`tools`에서 이 이름을 만드는 코드는 grep으로 찾지 못함 |
| 디스크 | — | 사용률 99%, 여유 49.6 GiB | 오늘 ENOSPC로 모든 레인이 멈춘 사고가 1회 있었다 |

- 측정 명령은 `target/agent5-2380/measure{,2,3}.sh`이고, 로그는 §8 부록에 옮겨 적었다.
- 감사 clone 분류에 쓴 glob `moai-*audit*`는 빌드 폴더 `moai-rc-target-agent-2-t61-audit-2326`(1.5 GiB)도 잡는다. 그래서 감사 clone 수는 이름 접두어로 따로 셌다.
- Issue 본문의 "감사 clone 1,019개"에는 빈 폴더 903개가 섞여 있다.

### 1.2 추적 파일과 이력

| 항목 | 값 |
|---|---|
| tip 추적 파일 | 583개, 10,197,932 바이트 |
| `.moai-cli/manifests` | 18개, 3.05 MB(연구 JSON 두 개가 2.0 MB) |
| `.moai-cli/reports` | 138개, 2.33 MB(`issue-bodies/` 131개 1.19 MB, 설계 HTML 1개 1.06 MB) |
| `src` / `tests` / `docs` | 1.28 / 2.03 / 1.23 MB |
| `origin/v2026.11.0`에서 닿는 객체 | 6,564개, 디스크 5.19 MB(그중 `.moai-cli` 411개, 1.24 MB) |
| primary pack | 488.9 MiB, 객체 306,685개. Go 시대 이력과 로컬 참조가 대부분이다 |
| GitHub branch | **19개**(`git ls-remote --heads origin`) |
| 로컬 원격 추적 참조 | 1,851개. 그중 1,828개가 `refs/remotes/pr/<번호>`(설정된 `pr` remote 없음 — PR head를 받아 온 흔적), `origin/*`은 22개 |
| 로컬 branch | 224개(`origin/v2026.11.0`에 병합된 것 287개 — 원격 추적 포함 계산) |
| worktree | 58개(`.moai-cli/worktrees` 47, `.claude/worktrees` 8, 그 밖 3). **locked 0개** |

- Issue의 "원격 branch 1,851개"도 대부분 로컬에만 있는 `pr/*` 참조다. GitHub 쪽 정리 대상은 거의 없다.
- worktree가 모두 `locked=0`이라는 것은 레인이 `moai-cli worktree-apply`(`git worktree add --lock`)가 아니라 손으로 `git worktree add`를 쓴다는 뜻이다.

### 1.3 기존 기록

- `src/worktree.rs`: `worktree-plan`/`worktree-apply`가 claim 하나에 worktree 하나를 잠금과 함께 만든다. 기본 위치 `.moai-cli/worktrees/<issue>-<key>`는 사용자 결정 16:5x→17:0x(11-08)이고, branch 접두어는 현재 구현값 `issue/<number>-<key>`로 확정됐다(리더 확정 2026-11-12, `docs/decisions-2026-11-08.md` 90행 절차, #2380 댓글 6816161048).
- `src/lane_handoff.rs`: worktree를 떠나기 전의 읽기 전용 점검이다(clean, head=remote, PR, trailer, 보고, 불확실 효과 없음).
- **정리·용량을 다루는 기능은 없다.**
- 결정 22:0x(11-08): EnterWorktree로 들어가고, ExitWorktree(keep)로 나오며, **삭제 금지**.
- 결정 06:2x(11-09) ②: worktree별 `CARGO_TARGET_DIR` + 실행 tree SHA.
- `AGENTS.local.md`: "세션마다 자기 `CARGO_TARGET_DIR=/tmp/moai-rc-target-<label>`".
- `.moai-cli/reports/cache-cleanup.sh`(추적 안 됨, SAFETY-CACHE-1501): 실행 시 다시 확인하는 조건(`CACHEDIR.TAG`, 120분 내 변경, `lsof`·`ps`)과 dry-run 기본값을 이미 갖추고 있다.

## 2. 조사 기록 (공식 문서, 확인일 2026-11-10)

| # | 출처 · 버전 | 핵심 내용(원문 요지) | 판정 | 실측 |
|---|---|---|---|---|
| G1 | https://git-scm.com/docs/git-worktree (last updated 2.56.0) | 링크된 worktree는 HEAD·index 같은 worktree별 파일만 따로 두고 나머지를 공유한다. `refs/`는 공유된다. `remove`는 clean한 worktree만 지우며, 아니면 `--force`가 필요하다. 잠긴 worktree는 `--force`를 두 번 줘야 지운다. 손으로 지운 worktree는 `prune`이나 `gc.worktreePruneExpire`로 정리된다 | 확인, 채택 | 잠금 성공. 삭제 거부 시험은 not_run(§7) |
| G2 | https://git-scm.com/docs/git-sparse-checkout (2.55.0) | cone 모드가 기본이고 non-cone은 deprecated다. `set`은 worktree별 설정으로 올리며(`extensions.worktreeConfig`), 다른 worktree에 영향을 주지 않는다. 명령 자체가 "EXPERIMENTAL"이다. `worktree add --no-checkout` 뒤에 설정할 수 있다 | 확인, **보류**(§3.1) | `.moai-cli`를 뺀 worktree 5.6 MiB, 전체 11.1 MiB. 다른 worktree에는 `.moai-cli`가 그대로 있다. `worktreeConfig=true` |
| G3 | https://git-scm.com/docs/partial-clone (2.56.0) | 빠진 객체는 promisor remote에서 필요할 때 받는다. 그래서 온라인이어야 하고, 하나씩 받아 느리다 | 확인, 감사 clone에는 기각(오프라인 검사·이력 조회가 막힐 수 있음) | GitHub `--filter=blob:none --single-branch --no-checkout` `.git` 1.2 MiB |
| G4 | https://git-scm.com/docs/git-clone (2.54.0) | `--reference`는 alternates를 만든다. `--shared`는 "possibly dangerous": 원본에서 객체가 지워지면 clone이 깨진다. `--dissociate`는 복사해서 의존을 끊는다. `--single-branch` | 확인. `--reference`/`--shared`는 기각, `--single-branch`는 채택 | GitHub `--single-branch` `.git` 5.0 MiB, `--reference` 1.0 MiB(alternates), `--dissociate` 5.2 MiB(alternates 없음). 지금 감사 clone은 약 470 MiB |
| G5 | https://git-scm.com/docs/gitrepository-layout (2.49.0) | `worktrees/<id>/locked`가 있으면 prune을 막는다. `objects/info/alternates`. 대부분의 `$GIT_DIR` 파일은 worktree별이다 | 확인 | — |
| C1 | https://doc.rust-lang.org/cargo/reference/build-cache.html (판 미표시) | 산출물은 "target"과 "build" 디렉터리에 있다. 여러 workspace가 의존성 빌드를 공유하려면 sccache를 쓸 수 있다. build-dir 구조는 내부용이라 바뀔 수 있다 | 확인 | — |
| C2 | https://doc.rust-lang.org/cargo/reference/config.html | `build.build-dir`(`CARGO_BUILD_BUILD_DIR`, 기본값은 target-dir)과 템플릿 `{workspace-path-hash}`, `{cargo-cache-home}`. `cache.auto-clean-frequency`는 전역 캐시에만 적용되고 build 산출물 추적은 미구현(cargo#13136) | 확인. "1.91에서 안정화"는 요약(미검증) | cargo 1.94.1에서 `-Z` 없이 동작했다(§3.3 실험) |
| C3 | 공유 target-dir의 안전성 | 문서에 언급 없음 | unverified → 실측으로 대체 | **같은 이름 패키지를 가진 두 tree가 target-dir나 build-dir를 공유하면 B를 빌드해도 0개를 컴파일하고 A의 바이너리를 남긴다**(§3.3) |
| C4 | https://github.com/mozilla/sccache (v0.19.0, 2026-10-24) | 링커를 부르는 crate(bin·dylib·cdylib·proc-macro)는 캐시하지 못한다. incremental crate도 캐시하지 못한다. 기본 설정에서는 절대 경로가 같아야 hit가 난다(`SCCACHE_BASEDIRS`로 완화). 로컬 캐시 기본 10 GB이고, 로컬 저장소는 서버 하나만 쓴다 | 확인, **나중**(설치 금지라 실측 not_run) | `sccache` 없음 |
| H1 | https://code.claude.com/docs/en/worktrees | 기본 위치는 `.claude/worktrees/<name>/`, branch는 `worktree-<name>`, `worktree.baseRef`는 `fresh`(원격 기본 branch) 또는 `head`다. 원문: "When Claude enters a path outside the repository's `.claude/worktrees/` directory, Claude Code asks for your approval first." "...only `bypassPermissions` mode skips it." `cleanupPeriodDays` sweep은 subagent worktree 중 변경 없는 것만 지운다. worktree는 새 checkout이므로 의존성을 설치해야 한다 | 확인(원문) | 이 세션은 bypass 모드라 확인 창이 뜨지 않음(증거 아님) |
| H2 | https://code.claude.com/docs/en/hooks — WorktreeCreate/Remove | hook이 기본 git 생성을 대체하고, 마지막 stdout 줄로 경로를 돌려준다. `.worktreeinclude`는 처리하지 않는다 | 요약(미검증) | not_run |
| H3 | https://code.claude.com/docs/en/permissions, cli-reference | 기본 접근 범위는 실행한 폴더다. `--add-dir`, `permissions.additionalDirectories`가 있다. `blockWritesOutsideWorkingDirectories`(v2.1.314+) | 일부 요약(미검증) | — |
| X1 | https://learn.chatgpt.com/docs/agent-approvals-security, /sandboxing, /config-file/config-reference (developers.openai.com/codex에서 308 리디렉트) | workspace-write는 현재 폴더와 `/tmp`에 쓸 수 있고, `writable_roots`·`--add-dir`로 넓힌다. `<writable_root>/.git`은 읽기 전용이다(pointer 파일이면 실제 git dir도). `exclude_slash_tmp` | 요약(미검증) | not_run(codex 실행 없음) |
| X2 | https://learn.chatgpt.com/docs/environments/git-worktrees | Codex 앱 worktree는 `$CODEX_HOME/worktrees`에 만들고, 기본으로 최근 15개를 유지한다. 고정되었거나 진행 중인 chat의 worktree는 보호한다. 지우기 전에 snapshot을 남긴다. branch는 한 worktree에서만 checkout한다 | 요약(미검증) | — |

참고로 Codex `workspace-write`에서 `.git`이 읽기 전용이면 링크된 worktree에서 커밋할 수 없을 가능성이 있다. 공통 `.git/objects`·`.git/worktrees/<id>`에 써야 하기 때문이다. 다만 이것은 추론이고 not_run이다. 오늘부터 역할 세션은 `-s danger-full-access`로 열린다(#2359).

## 3. 실험 (임시 저장소·합성 crate)

### 3.1 worktree와 sparse-checkout (GitHub tip clone 기준)

로컬 primary에서 받은 첫 clone은 primary의 낡은 로컬 branch(`b37cdf2dd7`, 파일 182개)였다. 그래서 그 숫자는 버리고 GitHub에서 다시 받았다.

```
git clone -q --single-branch --branch v2026.11.0 --no-checkout https://github.com/modu-ai/moai-adk.git exp/single-gh   # .git 5168 KiB, HEAD f259f15a14
git -C exp/single-gh worktree add -q -b exp-wt ../gh-wt-full HEAD                                               # 11332 KiB (.moai-cli 5604 KiB)
git -C exp/single-gh worktree add -q --no-checkout -b exp-sparse ../gh-wt-sparse HEAD
git -C exp/gh-wt-sparse sparse-checkout set --cone src tests docs skills schemas tools templates adapters scripts experiments .github
git -C exp/gh-wt-sparse checkout -q exp-sparse                                                                   # 5728 KiB, .moai-cli 없음, Cargo.toml 있음
git -C exp/single-gh config --get extensions.worktreeConfig                                                      # true
```

### 3.2 감사 clone 형태 (GitHub)

| 명령 | `.git` |
|---|---|
| 지금 방식(전체 clone, 표본 7개) | 469~485 MiB |
| `git clone --single-branch --branch v2026.11.0 --no-checkout <url>` | 5,168 KiB |
| 위 + `--filter=blob:none` | 1,184 KiB |
| 위 + `--reference exp/single-gh` | 1,012 KiB(alternates가 `single-gh`를 가리킴) |
| 위 + `--reference … --dissociate` | 5,348 KiB(alternates 없음) |

### 3.3 cargo: 공유 target-dir / build-dir

같은 패키지 이름 `probe`, 의존성 `serde =1.0.228`을 가진 crate `a`(`tree-A` 출력)와 `b`(`tree-B` 출력)를 두었다. `CARGO_HOME=/tmp/moai-rc-product-cargo`, `--offline`으로 빌드했다. 스크립트는 `target/agent5-2380/cargo/run.sh`, `run2.sh`.

| 실험 | 설정 | 결과 |
|---|---|---|
| C1 | A와 B가 `CARGO_TARGET_DIR` 하나를 공유 | A는 8개 컴파일. **B는 exit 0인데 0개 컴파일하고, `shared/debug/probe`가 `tree-A`를 출력**한다 |
| C1b | 이어서 `touch b/src/main.rs` 뒤 B 빌드 | 1개를 컴파일하고 `tree-B`를 출력한다. **이제 A의 스크립트가 실행하는 같은 경로가 `tree-B`를 낸다** |
| C2 | `CARGO_BUILD_BUILD_DIR` 공유 + target-dir은 tree별 | A는 8개 컴파일, B는 0개. **`tb/debug/probe`가 `tree-A`를 출력**한다 |
| C4 | `CARGO_BUILD_BUILD_DIR=<공통>/{workspace-path-hash}` + tree별 target-dir | A와 B 모두 8개를 컴파일하고 각자 `tree-A`/`tree-B`를 출력한다. 하위 폴더 2개, 59 MB — **공유 효과 없음** |
| C3 | tree별 기본 `target/` | 각 8개 컴파일, 30 MB씩 |

- 해석(추론): cargo는 경로 패키지의 신선도를 소스 mtime과 dep-info 시각으로 판정한다. 그래서 같은 이름의 다른 tree를 "이미 최신"으로 본다. 이 원인 설명은 문서로 확인하지 못했다.
- 결과 자체는 실측이다. 오늘 겪은 "다른 tree의 바이너리 실행" 사고는 `CARGO_TARGET_DIR` 하나를 여러 worktree가 쓰는 한 다시 생긴다. `build-dir`만 공유해도 똑같다.

## 4. 결정안 (Issue 9개 항목)

각 항목에 결정안 1개, 버린 대안, 근거를 적는다.

### D1. checkout에 무엇을 넣는가

- **결정안**: 1단계에서는 sparse-checkout을 쓰지 않는다. 추적 원장(`.moai-cli/reports/issue-bodies/**`, `.moai-cli/manifests/issue-design-research*.json`, `.moai-cli/reports/rc1-issue-design-2026-11-08.html`, `rc2_issue_gen.py`(2026-11-13 #2682 에서 삭제 — 마지막 판 `5f9109829a`), 합계 약 4.2 MB)을 추적에서 뺄지는 사용자에게 묻는다(Q2). 근거:
  - 이득이 작다. worktree 하나에 5.5 MiB, 58개면 약 310 MiB로, 낭비 전체의 0.2%다.
  - `src/agents.rs`, `src/apply.rs`, `src/generation.rs`, `src/scope.rs`, `src/sources.rs`, `src/github_templates.rs`와 시험 10여 개가 `.moai-cli/manifests` 경로를 읽는다(그중 저장소 자체 파일을 읽는지, fixture 경로만 쓰는지는 카드에서 파일별로 확인해야 한다). 그래서 sparse로 빼면 검사 묶음이 깨질 위험이 있다.
  - sparse-checkout은 공식 문서가 "EXPERIMENTAL"로 표시한다.
- **버린 대안**:
  - (a) 모든 worktree에 cone sparse 적용: 이득 0.2%에 비해 검사 실패 위험과 실험적 명령에 기대는 부담이 크다.
  - (b) 지금 바로 추적 해제: 사용자 결정(11-08 "07-A/07-B forge adapter 전까지 그대로")과 충돌한다.

### D2. 위치와 이름

- **결정안**: 위치는 사용자 결정대로 `<primary>/.moai-cli/worktrees/<issue>-<key>`를 유지하고, branch는 `issue/<issue>-<key>`다.
  - 감사용 임시 checkout은 `moai-audit-<pr>-<head8>-<lane>`(§D4)다.
  - 이 위치는 공식 문서상 Claude Code가 들어갈 때마다 승인을 묻는다(bypass 모드 제외). 이것이 오늘 "permission-root" 확인 창으로 레인이 멈춘 원인이다.
  - #2359부터 `-l`/`-a` 역할 세션은 `--dangerously-skip-permissions`로 열린다. 그래서 역할 레인에서는 이 확인 창이 사라진다. 역할 없는 세션과 `--permissions` 세션에서는 남는다(Q3).
- **버린 대안**:
  - (a) `.claude/worktrees/`: 확인 창은 없지만, `.claude/`는 보존 대상인 Go 상태이고 00-A 측정을 오염시킨다(사용자가 철회한 결정 16:5x).
  - (b) WorktreeCreate hook으로 위치 바꾸기: hook 계약이 요약(미검증)이고, hook이 만든 경로가 승인 대상에서 빠지는지 문서에 없다(unverified). 그래서 2단계 시험 카드로 둔다.

### D3. 빌드 산출물

- **결정안**: **build 폴더는 worktree마다 하나이고, worktree 안의 `target/`에 둔다**(`CARGO_TARGET_DIR`은 지정하지 않는다).
  - worktree가 정리될 때 함께 사라져 수명이 같다.
  - 같은 세션이 여러 tree를 오가도 섞이지 않는다(§3.3 C1·C2).
  - `AGENTS.local.md`의 "세션별 `/tmp/moai-rc-target-<label>`" 규칙은 이 결정으로 대체한다. 그 폴더들(57개, 65.9 GiB)이 worktree보다 오래 남는 것이 가장 큰 낭비원이다.
  - 감사 clone도 자기 안의 `target/`을 쓴다.
- **줄이는 방법**: 공유가 아니라 수명 주기로 줄인다.
  - PR을 넘긴 뒤에는 `target/`을 먼저 지운다(재빌드 가능, §D5 trim).
  - 동시에 살아 있는 `target/` 수에 상한을 둔다(§D7).
  - dev profile의 debuginfo 축소(`line-tables-only` 등)는 효과를 아직 재지 않았다. 측정 카드로 둔다.
- **버린 대안**:
  - (a) `CARGO_TARGET_DIR` 하나를 공유: C1에서 다른 tree의 바이너리가 나왔다.
  - (b) `build.build-dir` 공유: C2에서 다른 tree의 바이너리가 나왔다.
  - (c) `build-dir`을 `{workspace-path-hash}`로 나누기: 안전하지만 공유 효과가 0이다(C4).
  - (d) sccache: 의존성 crate 캐시 효과는 있을 수 있다. 하지만 설치가 필요하고, 기본 설정에서는 절대 경로가 같아야 hit가 나며, bin·proc-macro는 캐시하지 않는다. 실측 not_run이라 2단계 선택 항목으로 둔다(Q9).

### D4. 감사 clone

- **결정안**: GitHub에서 `git clone --single-branch --branch <PR head branch> --no-checkout <url> <dir>`로 받는다. 이어서 `git fetch origin <base branch>`(merge-base 계산용), 정확한 SHA를 detached로 checkout하고 HEAD를 단언한다.
  - `.git`이 약 470 MiB에서 약 5 MiB로 준다(약 94분의 1).
  - 독립성은 그대로다. 작성자의 객체 저장소나 환경을 쓰지 않고 GitHub에서 새로 받는다.
  - 빌드는 clone 안의 `target/`에서 한다.
  - 판정을 게시하고 다시 읽어 확인한 뒤 clone을 지운다. 그 전에 감사 증거(로그 몇 KB)를 clone 밖의 작은 보존 위치로 옮긴다. 후보는 `~/.moai-cli/projects/<id>/audits/<pr>-<head8>/`이고, 위치는 Q6에서 정한다.
- **버린 대안**:
  - (a) `--reference`/`--shared`로 primary 공유: 공식 문서가 원본 정리 시 clone이 깨진다고 경고한다. 작성자의 객체 저장소에 기대게 되어 독립성도 약해진다.
  - (b) `--filter=blob:none`: 더 작지만(1.2 MiB) 이력 조회 때 온라인이 필요하다. 오프라인 검사(`--offline`) 원칙과 맞지 않는다.
  - (c) 감사용 linked worktree: 작성자 clone의 `.git`을 공유하므로 독립 감사 원칙과 충돌한다.

### D5. 수명 주기

상태와 전이:

```
planned ─worktree-apply(lock: "moai claim <key> e<epoch>")→ active
active ─PR 생성 + lane-handoff-check PASS→ handed_over
handed_over ─trim(target/ 삭제, 재빌드 가능)→ handed_over(trimmed)
handed_over ─PR merged/closed + claim 해제→ done
done ─remove(사용자 확인)→ (없음)
```

- **"사용 중" 판정**: 아래 하나라도 해당하면 지우지 않는다. 판정은 2026-11-10 SAFETY-CACHE 스크립트의 조건을 제품으로 옮긴다.
  - claim lease가 살아 있음
  - git worktree lock이 있음
  - 프로세스가 그 경로를 작업 폴더로 쓰거나 파일을 열고 있음(`lsof`·`ps`)
  - 최근 120분 안에 변경이 있음
  - 미커밋 변경이 있음
  - HEAD가 원격(`ls-remote`)과 다름
- **trim**: `target/`만 지운다. `CACHEDIR.TAG`가 있을 때만 대상이 된다. 소스와 커밋은 남는다.
- **remove**: `git worktree unlock` 뒤 `git worktree remove`를 쓴다(**`--force` 금지**). 병합된 로컬 branch는 `git branch -d`(병합 안 된 branch는 git이 거부)로 지운다.
- **기본은 dry-run**이다. 실제 실행은 목록 digest를 확인한 다음에만 한다(Q4).
- **결정 22:0x("worktree 삭제 금지")를 바꾸는 제안**이다. 사용자 결정이 필요하다(Q4).

### D6. branch 정리

- **결정안**:
  - GitHub branch는 19개라 지금 정리할 대상이 거의 없다. 앞으로 PR 병합 시 head branch를 자동으로 지우는 저장소 설정을 쓸지는 운영자가 결정한다(Q7).
  - 로컬: `refs/remotes/pr/*` 1,828개는 로컬에만 있는 PR head 흔적이다. 목록 제시와 확인 뒤 `git update-ref -d`로 지운다.
  - 병합된 로컬 branch는 `git branch -d`로만 지운다.
  - 보존 ref(`main`, `2026.11.0-rc.1`, `archive/*`, Go 시대 ref)는 제외한다.
- **버린 대안**: `git fetch --prune` 일괄 실행. `pr/*`는 설정된 remote가 없어 prune 대상이 아니고, 보존 ref를 실수로 건드릴 위험이 있다.

### D7. 용량 상한과 관측

- **결정안**:
  - `moai-cli worktree-report`(읽기 전용)가 분류별 개수와 용량을 보여 준다: worktree checkout, worktree `target/`, 세션별 빌드 폴더(`/tmp/moai-target-*`와 #2563 전 이름 `/tmp/moai-rc-target-*`, 출력 필드 `tmp_target`), 감사 clone, 빈 임시 폴더, 로컬 ref. 디스크 여유와 trim·remove 후보도 함께 보여 준다. `du`에는 시간 상한을 걸고, 넘치면 `partial`로 표시한다.
  - `worktree-apply`는 디스크 여유가 하한 미만이면 새 worktree를 거부하고, 프로젝트 합계가 경고선을 넘으면 경고한다. 값은 사용자 결정(Q5)이다. 참고로 오늘 ENOSPC 직전 여유가 49.6 GiB였고, worktree 하나가 빌드 후 1~5 GiB다.
- **버린 대안**: 상한 없이 주기적으로 정리하기. 오늘 ENOSPC 사고처럼 정리보다 생성이 빠르다.

### D8. `moai-cli` 표면

| operation | 종류 | 내용 | 기존과의 관계 |
|---|---|---|---|
| `worktree.plan` / `worktree.apply` | 기존 | 디스크 하한 검사를 더한다(D7). lock reason에 claim 토큰을 넣는다 | 레인이 실제로 이 경로를 쓰도록 절차를 바꾼다(locked 0 → 전원) |
| `lane.handoff.check` | 기존 | 확인 항목에 "trim 후보 기록"을 더한다(차단 아님) | 그대로 |
| `worktree.report` | **신규**, 읽기 전용 | D7 표. 스키마 `schemas/worktree-report.schema.json` | — |
| `worktree.gc.plan` | **신규**, 읽기 전용 | 후보마다 `{path, class, action: trim\|remove\|skip, reasons[], bytes}`와 목록 digest | SAFETY-CACHE 조건을 제품으로 옮긴다 |
| `worktree.gc.apply` | **신규**, 쓰기 | `--digest`가 같을 때만 실행하고, 실행 직전에 모든 조건을 다시 검사한다. `--force` 없음 | 결정 22:0x 개정이 필요하다(Q4) |
| `audit.clone.plan` | **신규**, 읽기 전용 | D4 명령 열을 만든다(경로·branch·head·base) | 감사 절차 문서 |

### D9. 이전 계획 (삭제는 모두 목록 제시 → 사용자 확인 → 실행)

1. `worktree-report`를 구현하고 첫 보고를 만든다(쓰기 0).
2. 규칙을 바꾼다: 새 작업부터 worktree 안 `target/`과 single-branch 감사 clone을 쓴다(AGENTS.local.md, lane-loop 문서).
3. `/tmp/moai-rc-target-*` 57개(65.9 GiB): 사용 중이 아닌 것만 지운다(SAFETY-CACHE 조건).
4. 병합된 PR의 감사 clone 86개(39.6 GiB): 판정 댓글이 게시된 것만 지운다.
5. 넘겨졌거나 병합된 worktree의 `target/` trim(27.3 GiB 중 사용 중 아닌 것).
6. 빈 임시 폴더 3,612개(용량 0, 항목 수만 정리).
7. 병합·종료된 worktree remove(clean·push 확인 뒤).
8. 로컬 `refs/remotes/pr/*`와 병합된 로컬 branch.
- `.claude/worktrees/` 8개는 보존 대상인 `.claude/` 안에 있으므로 이전 대상에서 빼고 사용자에게 묻는다(Q8).

예상 회수량은 3~5단계에서 최대 약 132 GiB다. 사용 중인 것을 빼면 더 줄어든다.

## 5. 사용자 결정 질문 (리더가 묻는다)

1. **Q1 우선순위**: 실측상 낭비는 커밋된 파일이 아니라 빌드 폴더(93 GiB)와 전체 이력 감사 clone(40 GiB)이다. 1차 목표를 "빌드·감사 clone 수명 관리"로 두고, 추적 파일 정리는 2차로 미뤄도 되는가?
2. **Q2 추적 원장**: `.moai-cli/reports/issue-bodies/**`, 연구 JSON·설계 HTML(약 4.2 MB)을 어떻게 할까? (a) 지금처럼 추적(11-08 결정 유지) (b) forge(Issue)에만 두고 추적 해제 (c) home layout(`~/.moai-cli/projects/<id>/`)으로 이동.
3. **Q3 worktree 위치**: `.moai-cli/worktrees/`를 유지하면 역할 없는 세션과 `--permissions` 세션은 들어갈 때마다 Claude Code 승인 창이 뜬다. (a) 유지(역할 레인은 전체 권한이라 영향 없음) (b) WorktreeCreate hook 검증 카드를 먼저 진행 (c) 다른 위치.
4. **Q4 정리 권한**: 결정 22:0x("worktree 삭제 금지")를 바꿔, 병합·종료된 작업의 `target/` trim과 worktree remove를 허용할까? 실행 주체는 (a) 사용자만 (b) 리더가 목록 digest 확인 후 (c) 레인이 자기 worktree만.
5. **Q5 상한 값**: 새 worktree를 거부할 디스크 여유 하한(예: 30 GiB)과 프로젝트 합계 경고선(예: 40 GiB). 값을 정해 주세요.
6. **Q6 감사 증거 보존 위치**: clone을 지운 뒤 로그를 어디에 남길까? (a) 판정 댓글만 (b) `~/.moai-cli/projects/<id>/audits/` (c) `/tmp`(재부팅·정리 시 사라짐).
7. **Q7 GitHub 설정**: PR 병합 시 head branch 자동 삭제를 켤까? (저장소 설정 변경)
8. **Q8 `.claude/worktrees/` 8개**: 보존 대상으로 둘까, 이전 대상에 넣을까?
9. **Q9 sccache**: 바이너리 설치가 필요하다. 2단계에서 측정 카드를 만들까?

## 6. 구현 카드 분할안 (2단계)

| 카드 | 내용 | 파일 범위 | 선행 |
|---|---|---|---|
| WT-1 | `worktree-report` 읽기 전용 + 스키마 + 시험(합성 저장소, du 상한) | 새 `src/worktree_report.rs`, `schemas/worktree-report.schema.json`, `tests/worktree_report.rs`. `src/main.rs`·`src/lib.rs` 등록은 리더 | 없음 |
| WT-2 | 규칙 문서 개정: worktree 안 `target/`, single-branch 감사 clone, 세션별 `/tmp` target 폐지 | `AGENTS.local.md`, `docs/lane-loop-contract.md`(리더 소유) | Q1 |
| WT-3 | `audit.clone.plan` + 감사 절차 문서 | 새 `src/audit_clone.rs`, `docs/` 감사 절, 시험 | WT-1(이름 규칙), Q6 |
| WT-4 | `worktree.gc.plan`/`apply`(trim·remove 단계, lock·claim·lsof·mtime·clean·push 조건, digest) | 새 `src/worktree_gc.rs`, 스키마, 시험(조건마다 N/N+1·변이) | WT-1, Q4 |
| WT-5 | `worktree-apply` 디스크 하한·경고 | `src/worktree.rs`, 시험 | WT-1, Q5 |
| WT-6 | 레인이 `worktree-apply`(lock)를 쓰도록 절차를 바꾸고 branch 접두어 HOLD를 해소 | `docs/lane-loop-contract.md`, `src/worktree.rs` 기본값 | Q3, 접두어 HOLD |
| WT-7 | 추적 원장 이동 | `.moai-cli/**`, `.gitignore`, 관련 시험 | Q2, 07-A/07-B |
| WT-8 | 빈 임시 폴더를 만든 출처 찾기(11-06~11-08)와 시험 정리 보장 | `tests/**` 해당 파일 | 없음 |
| WT-9 | 이전 실행(운영자) | 실행만, 코드 0 | WT-4, 사용자 확인 |
| WT-10 (선택) | dev profile debuginfo 축소 효과 측정, sccache 측정 | `Cargo.toml` profile, 측정 문서 | Q9 |

## 7. not_run · unverified · 도구 기록

- **not_run**
  - `git worktree remove`(잠김·미커밋 거부) 실험: worktree 격리 가드가 명령을 거부해서 실행하지 않았다. 잠금(`worktree lock`)은 성공했다. 근거는 G1 문서다.
  - alternates 원본 정리 시 손상 실험: 원본 pack 이동이 필요해 하지 않았다. 근거는 G4 문서의 경고다.
  - 로컬 `file://` partial clone: `--upload-pack` 옵션을 가드가 거부했다. GitHub에서 대신 측정했다.
  - sccache, Codex 실제 세션, WorktreeCreate hook, 역할 없는 Claude 세션의 `.moai-cli/worktrees` 진입 확인 창.
- **unverified(요약만)**: cargo `build.build-dir` 1.91 안정화 버전 번호(동작은 1.94.1에서 실측), Codex `.git` 읽기 전용 문구, WorktreeCreate hook 계약, cargo mtime 원인 설명(추론).
- **도구 기록**
  - worktree 격리 가드가 `git`이 들어간 복합 명령과 스크립트 파일 이름(`exp-git.sh`)을 거부했다. 거부된 명령은 실행되지 않았음을 확인했다(파일과 폴더 그대로).
  - 한 번은 "거부" 알림과 달리 `measure.sh`가 실행되어 로그를 남겼다. 실행 결과(exit 0, 로그)를 확인하고 그대로 썼다.
  - 이 세션은 bypass 권한 모드이므로, 확인 창이 없었다는 것은 D2의 증거가 아니다.

## 8. 부록: 측정 원문 발췌

```
date 2026-11-10T10:19Z tip f259f15a14c9fc5a396514e8e0995b891f7f4f3f
cargo 1.94.1 (29ea6fb6a 2026-03-24) / git 2.54.0 (Apple Git-157)
tracked TOTAL 10197932 bytes 583 files; .moai-cli/manifests 3048386/18; .moai-cli/reports 2326985/138
count-objects: in-pack 306685, size-pack 488.88 MiB
worktrees: total=58 moai-cli=47 claude=8 other=3 locked=0 prunable=0
worktree checkout (excluding target/) KiB: 8104 9464 9416 11024 9428 9248 9592 9584
targets=39 total_KiB=28619520; largest: 2337 4796380, 2298 3429416, 2348 3413456, 2264 3311584
  2337 split: debug 2884336 KiB, nested .git 2944 KiB; 2298: debug 1863636, nested .git 1025020
.moai-cli/worktrees total KiB 33344252
/tmp/moai-rc-target-* 57 dirs, 69101880 KiB; audit clones 86 dirs, 41573288 KiB; /tmp/moai-* 149398832 KiB
audit clone samples .git KiB: 484680 478428 469428 469428 470636 469416 469420
empty temp dirs: moai-cli-profiles 903, moai-cli-home 903, moai-cli-gitconfig 903, moai-audit-receipts 903 (0 KiB)
reachable from origin/v2026.11.0: objects=6564 disk_bytes=5187826; under .moai-cli objects=411 disk_bytes=1244586
git ls-remote --heads origin: 19; local refs/remotes/pr/*: 1828; local branches 224
df: 99% used, 52041808 KiB available
```

## 9. 1단계 뒤 실측 보강 (2026-11-10T17:38Z, WT-1 `worktree-report`)

1단계 설계(§1, 10:19Z 측정) 뒤 같은 날 레인 운영에서 생긴 사실이다. 결정안(§4)은 바꾸지 않고, D3·D4·D7의 근거로만 덧붙인다.

- **측정 방법**: WT-1 `moai-cli worktree-report <primary> --budget-ms 180000`(읽기 전용, 할당 블록·inode 한 번, git `--no-optional-locks`). 17:38:36Z 시작, 17:41:37Z 끝. `/tmp` 세 분류는 예산 안에 다 돌지 못해 `partial`이며 그 값은 하한이다. 공개 기록에는 경로 없이 집계만 적는다.

| 분류 | §1 (10:19Z) | 17:38Z | 비고 |
|---|---|---|---|
| git worktree | 58 | 108 | detached 25(이름에 `audit`이 든 것 21), dirty 12, locked 0 |
| worktree `target/` | 39개, 27.3 GiB | 88개, 75.5 GiB | 늘어난 대부분이 감사·사전 감사용 worktree |
| worktree checkout(`target/` 제외) | 표본 8~11 MiB | 108개 합계 5.8 GiB | |
| `/tmp/moai-rc-target-*` | 57개, 65.9 GiB | 60개, 57.4 GiB 이상(partial) | |
| `/tmp` 감사 clone | 86개, 39.6 GiB | 53개, 21.4 GiB 이상(partial) | WT-1 분류는 이름에 `audit` + `.git` 있음 |
| `/tmp` 빈 임시 폴더 | 903 × 4 | 1,716 | 0 B, WT-8 대상 |
| 로컬 branch / `refs/remotes/pr/*` | 224 / 1,828 | 254 / 1,832 | |
| 디스크 여유 | 49.6 GiB | 22.4 GiB | |

- **D3 근거 보강 — 한 빌드 폴더 = 한 worktree**: 같은 날 14:3xZ, 두 worktree가 같은 `CARGO_TARGET_DIR`을 써서 `Compiling … (자기 tree)`를 찍고도 다른 tree의 바이너리로 시험이 돈 사고가 레인에서 실제로 났다(#2401 댓글 6770623896, 다른 레인의 같은 사례 댓글 6770673087·6770674526·6770677078·6770783481·6770850433). 리더는 '새 빌드 폴더 금지, 기존 것 재사용' 지시를 '한 빌드 폴더 = 한 worktree'로 정정했다(결정 기록 14:58Z(11-10) 행). §3.3 C1·C2 실험 결과가 운영에서 재현된 것이고, D3 결정안(worktree 안 `target/`, `CARGO_TARGET_DIR` 미지정)과 같은 방향이다.
- **D4 관찰 — 감사가 clone이 아니라 linked worktree로 옮겨 갔다**: 이날 레인 감사·사전 감사의 대부분은 primary 저장소의 detached linked worktree(`.moai-cli/worktrees/audit-*`)에서 돌았다(위 detached 25). D4가 버린 대안 (c)(작성자 clone의 `.git` 공유)에 해당한다. 같은 primary를 쓰더라도 감사 대상은 GitHub에서 fetch한 정확한 SHA라 내용 독립성은 지켰지만, 객체 저장소 공유라는 점은 그대로다. 디스크 압박에서 나온 운영 선택이므로 D4를 바꾸지 않고 Q6·Q1 질문과 함께 리더에게 올린다.
- **D7 근거 보강 — 정리보다 생성이 빠르다**: 10:19Z 뒤 7시간 반 동안 worktree가 50개, worktree `target/`이 49개(약 48 GiB) 늘었다. 이 사이 리더가 병합 끝난 worktree의 `target/` 몇 곳을 지운 정리가 있었다(리더 보고, 회수량 약 20 GiB로 전달받았으나 공개 기록으로는 확인하지 못함 — unverified). 정리 뒤에도 여유는 49.6 → 22.4 GiB로 줄었다. 상한(D7, Q5) 없이 수동 정리에 기대는 방식의 한계를 다시 보여 준다.
