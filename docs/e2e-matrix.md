# E2E matrix (08-C #2014, RC1)

이 문서는 08-C E2E receipt(`schemas/e2e-receipt.schema.json`, REQ-08C-01)의 조합별 상태를 기록한다. 이 표만이 08-D #2015 지원표(`docs/support-matrix.md`)의 E2E 입력이다. 열·형식은 지원표 소유 Issue 02-C #1993이 정한 것을 따른다(host | surface | item | capability_status | unknown_class | 내용 | source); receipt 필드(os·arch·host·host_exact_version·mode·skill_digest·moai_cli_sha·result)는 각 행의 receipt JSON으로 남긴다. 실행하지 않은 조합은 숨기지 않고 그대로 적는다.

## 상태 어휘

`capability_status`는 지원표(02-C #1993)와 같은 모델을 따른다. `certified`(실제 앱 E2E)만 "지원"으로 표기할 수 있다.

| 값 | 이 표에서의 뜻 |
| --- | --- |
| `probed` | T1 결정적 fixture가 이 조합에서 실측 receipt를 남겼다 |
| `unknown` | 아직 실측하지 않았다(`unknown_class` = `시험 대기`) |

P12 결정(리더 범위 댓글 2026-11-08, 사용자 결정 적용): RC1의 조합은 **개발 Mac 단일**(aarch64-apple-darwin, macOS)이다. 다른 OS·arch·cloud/worktree mode는 전부 not_run으로 남고 OS matrix workflow는 RC1에서 만들지 않는다(08-D 소유).

## T1 — 결정적 fixture (`tests/e2e_host_fixture.rs`)

실행 명령(실제 시험 파일·시험명과 일치 — AC 명령 인용 전 확인 의무):

```sh
CARGO_HOME=/tmp/moai-rc-product-cargo CARGO_TARGET_DIR=/tmp/moai-rc-product-target \
HOME=$(mktemp -d) cargo test --offline --locked --test e2e_host_fixture -- --nocapture
```

| host | surface | item | capability_status | unknown_class | 내용 | source |
| --- | --- | --- | --- | --- | --- | --- |
| none (T1 fixture) | cli | e2e_setup_update_interrupt_restore_preserves_user_edits (AC-08C-1) | probed | | 임시 HOME·프로젝트에서 project-plan/apply(project-plan, project-apply, 01-B) + update-apply v1(04-B 양 host 설치) 후, 실제 중단(읽기 전용 projection 디렉터리로 apply가 트랜잭션 중간 실패) → recovery-plan/recovery-apply가 known-good에서 대체 파일 복구, 사용자 편집(routes.json) 바이트 보존, 성공 host(codex)는 committed v2 유지, 가짜 Go 상태 digest 불변 | cargo test 실행 기록(아래 receipt) |
| none (T1 fixture) | cli | e2e_digest_mismatch_is_fail_closed (AC-08C-2, digest 절반) | probed | | plan digest 계산 뒤 stage 바이트 변조 → update-apply 거절, HOME 트리 전체 바이트 동일 | cargo test 실행 기록 |
| none (T1 fixture) | cli | wrong_target_package (AC-08C-2, target 절반) | unknown | blocked(제품 경로 없음) | skill 패키지 update에는 target 개념이 없고, 바이너리 갱신 경로는 `src/self_update.rs` 모듈 문서상 "04-B, not built yet"이다. 이 Issue는 `src/**` 변경 금지라 시험할 대상이 없다 — 04-B 소유 | 설계 댓글 #2014 6809848052 |
| none (T1 fixture) | cli | e2e_second_update_is_diff_zero (HA-02) | probed | | 동일 stage 재요청: 모든 action이 unchanged/preserved, apply 후 HOME digest 0 diff | cargo test 실행 기록 |
| none (T1 fixture) | cli | e2e_never_writes_outside_temp_home (AC-08C-7) | probed | | **macOS 전용 측정** — sandbox-exec deny-home(실제 HOME 쓰기 전면 거절) 아래에서. 거절 경로는 `$HOME`이 아니라 passwd 항목(`getpwuid_r`)의 계정 홈이다 — 위 T1 명령은 `HOME=$(mktemp -d)`로 돌기 때문에, `$HOME`에서 만든 프로필은 버려지는 임시 디렉터리만 막았다(2026-11-11 수정 전 receipt는 이 한계를 가짐). fixture HOME이 그 계정 홈 밖에 있음도 단언한다. 경로만 읽고 실제 HOME 내용은 읽거나 쓰지 않는다. 그 아래에서 update-apply 성공 — HOME 리다이렉션이 사라져도 사용자 상태를 못 쓴다. 가짜 Go 상태 digest 불변. 실제 `~/.moai` 대량 해시 대조는 동시 변동으로 신뢰 불가하여 이 방식으로 갈음(리더 허용 방식). 이 시험은 `#[cfg(target_os = "macos")]`로 한정 — Linux에는 sandbox-exec가 없어 **컴파일되지 않으며(숨은 skip 아님) Linux CI는 AC-08C-7 not_run. Linux CI 통과는 AC-08C-7 측정이 아니다** | cargo test 실행 기록(macOS) |
| none (T1 fixture) | cli | e2e_clean_clone_carries_portable_read_plan_assets (AC-08C-8, T1 절반) | probed | | binary 없는 clean clone에 skills/moai 전체(SKILL.md frontmatter·routes.json·routes가 참조하는 workflow 전부)가 존재 — READ/PLAN 포터블 자산 완결 | cargo test 실행 기록 |
| none (T1 fixture) | cli | e2e_clean_clone_keeps_exact_attached_and_detached_revision (AC-08C-8 보조, #2104) | probed | | clean clone이 시험한 커밋(붙은 가지·detached 모두)을 그대로 고른다 | cargo test 실행 기록 |

T1 receipt 필드(위 probed 행 공통): os=Darwin, arch=aarch64, host=none, mode=local, skill_digest=SKILL.md sha256(실행 시점 기록), moai_cli_sha=빌드 바이너리 sha256, ran_at=실행 시각 실측. raw 결과는 `cargo test` 출력 그대로가 detail이며 자격증명이 없다.

최신 T1 실행(2026-11-11, #2014 agent-8): tip `2449d26eae` + 위 AC-08C-7 수정에서 위 명령(`HOME=$(mktemp -d)`)으로 6 passed / 0 failed. `skill_digest` `25081f6c…`, receipt 5줄(AC-08C-1·2·7·8, HA-02)을 `docs/e2e-receipts/receipts.jsonl`에 **추가**했고(2026-11-08 줄은 이력으로 남김), 출력은 `docs/e2e-receipts/T1-cargo-test-output-2026-11-11.txt`다. 2026-11-08 receipt의 `skill_digest` `bff1c53b…`는 현재 `skills/moai/SKILL.md`와 다르다(낡음). 지원표(`docs/support-matrix.md`, 02-C·08-D 소유)의 digest 열 갱신은 소유 Issue 몫이다.

## T2 — 실제 host 로딩 (`tests/e2e/host-claude.sh`, `tests/e2e/host-codex.sh`)

레인은 실행하지 않는다(리더 범위: 기존 인증만 쓰는 운영자 1회 실측, 체크아웃 밖). 명령과 기대 결과는 스크립트에 receipt 형식으로 준비돼 있고, 운영자 실행 전까지 전부 not_run이다. 스크립트는 자격증명 값의 존재만 확인하고 값은 어떤 파일에도 쓰지 않는다. 두 스크립트의 **모든** host 호출(`--version`, `claude plugin validate` 포함)은 임시 작업 디렉터리의 `HOME`(codex는 `CODEX_HOME`도)으로 돌아 운영자의 실제 HOME을 읽거나 쓰지 않는다(2026-11-11; 이전에는 `-p`·`exec` 실행만 격리). 가짜 host 시험 두 파일이 각 호출의 HOME을 확인한다.

**--bare 가설 판정(2026-11-09 공식 문서 재조회 — agent-1 가설 기각)**: `--bare`는 skill 자동 탐색을 생략하지만 `--add-dir`은 부분 예외다 — "A directory you name with --add-dir is a partial exception: bare mode loads skills from its .claude/skills/ folder"(headless 문서). 프롬프트 앞머리의 `/moai`는 -p 모드에서 클라이언트가 전개("Include /skill-name in the prompt string and Claude Code expands it before running"). 주의: --bare 아래 모델은 스킬 **목록**을 받지 못하므로 자동(암묵) 호출은 측정 불가 — 명시 호출만이 이 E2E의 대상이며, 운영자 `~/.claude/skills/moai`(개인)가 fixture를 덮는(personal > added directory) 것을 막기 위해 스크립트는 -p 실행에 `HOME=<임시>`를 격리한다. 측정 대상은 명시 호출 자체다.

| host | surface | item | capability_status | unknown_class | 내용 | source |
| --- | --- | --- | --- | --- | --- | --- |
| Claude Code | cli | skill_static_validation (AC-08C-5) | probed | | `claude plugin validate <tmp>/.claude/skills` — 자격증명 불필요. 실행은 AC 명령 절에서 수행하고 receipt로 기록. **receipt 낡음**: 2026-11-08 receipt의 `skill_digest` `bff1c53b…`는 현재 SKILL.md(`25081f6c…`)와 다르다. **리더 답(2026-11-12, Q2 = HOLD 유지)**: 레인은 실제 `claude` 바이너리를 실행하는 검사를 하지 않는다(실제 host 실행 HOLD). 그래서 이 receipt는 운영자 실행 때까지 낡은 채로 남는다. 레인 check set의 `tests/e2e_receipt_not_run_test.py`도 실제 `claude`를 부르므로, 레인은 이 파일을 실행하지 않고 not_run으로 적는다 | 이 문서의 AC 명령 절 |
| Claude Code | cli | explicit_invocation_setup_route (AC-08C-3) | unknown | 시험 대기 | `claude --bare -p '/moai …' --add-dir <tmp> --output-format stream-json --verbose --permission-mode dontAsk --permission-prompts none`(HOME 격리) — 기대: exit 0 + `result` 이벤트 `is_error == false` + fixture의 `.claude/skills/moai/references/routes.json`과 `workflows/setup.md`를 가리키는 `tool_use` 이벤트(skill 본문이 "references/routes.json으로 route를 고른다"고 지시하므로, skill 로드와 SETUP route 도달이 이벤트로 드러난다. dontAsk가 거부한 시도도 센다). skill 없는 일반 답변은 failed. 이벤트 구조만 단언하고 자연어는 단언하지 않는다. not_run — 운영자 실측 대기 | tests/e2e/host-claude.sh |
| Claude Code | cli | permission_denial_blocked (AC-08C-6) | unknown | 시험 대기 | 동일 옵션(`stream-json --verbose`)으로 apply 요청 → `result` 이벤트의 `permission_denials`가 비어 있지 않은 배열 + 대상 트리 shasum 전후 동일. `result` 이벤트가 없거나(예: exit≠0, 빈 출력) `[]`이면 failed. not_run — 운영자 실측 대기 | tests/e2e/host-claude.sh |
| Codex | cli | explicit_invocation_setup_route (AC-08C-4) | unknown | 시험 대기 | `HOME`·`CODEX_HOME` 임시 격리 + `codex exec --json --ephemeral --ignore-user-config --sandbox read-only --skip-git-repo-check -o last.md '$moai …' < /dev/null` — passed 조건(모두 필요, 2026-11-11 강화): exit 0, `turn.completed` 있음, `turn.failed`·`error` 없음, last message 비어 있지 않음, `command_execution` item의 `command`가 `references/routes.json`과 `workflows/setup.md`를 가리킴. 마지막 조건은 **미검증 가정**이다 — Codex가 skill 파일을 셸 `command_execution`으로 읽는다는 것과 `$moai`가 SKILL.md를 따르게 한다는 것은 공식 문서 원문이나 실제 stream으로 확인하지 않았다. 가정이 틀리면 실제 성공이 failed로 나올 수 있고(원시 stream이 receipt 옆에 남음), 실패가 passed로 바뀌지는 않는다. 판정기는 `tests/e2e_host_codex_verdict_test.py`(가짜 codex 9사례)가 고정한다. 인증은 `CODEX_API_KEY`만 쓴다. not_run — 운영자 실측 대기 | tests/e2e/host-codex.sh |
| Claude Code | cli | clean_clone_read_plan (AC-08C-8, T2 절반) | unknown | 시험 대기 · 명령 미정 | 본문 명령은 binary 없는 clean clone에서 `claude --bare -p '/moai 현재 저장소 읽기' --output-format json`이다. 그런데 저장소는 `.claude/skills`를 추적하지 않고 명령에 `--add-dir`도 없어, `--bare`가 `moai` skill을 찾을 경로가 없다. 투영 단계를 넣으면 AC가 재는 것이 바뀌므로 스크립트로 만들지 않았다. **리더 답(2026-11-12, Q1 = HOLD 유지)**: 스크립트에 skill 복사 단계를 넣지 않는다. 실제 HOME에 skill을 설치·교체하는 일은 사용자 확인 대기이며 절차는 #2483에서 구현 중이다. 운영자가 할 수동 단계로만 아래 "운영자 수동 단계(AC-08C-8 host 절반)"에 적는다. not_run(T1 절반은 위 표 probed) | 이 문서의 AC 명령 절 |

## OS·mode 조합 (P12)

| os | arch | host | mode | result | 비고 |
| --- | --- | --- | --- | --- | --- |
| macOS (Darwin) | aarch64 | none (T1) | local | T1 표 참조 — 5시험 probed | RC1 유일 실측 조합 |
| macOS (Darwin) | aarch64 | Claude Code | local | AC-08C-3·6 not_run | 운영자 실측 대기 |
| ubuntu-24.04 | x86_64 | none (T1) | local | **AC-08C-7 not_run**(sandbox 없음 — 시험 미컴파일), 나머지 T1 4시험은 CI 실행 | **Linux CI 통과 ≠ AC-08C-7 측정** |
| macOS (Darwin) | aarch64 | Codex | local | AC-08C-4 not_run | 운영자 실측 대기 |
| ubuntu-24.04 / windows-2025 / arm 변형 | — | — | local | not_run | P12 HOLD — RC1에서 workflow 만들지 않음(08-D에 not_run 반영) |
| — | — | — | worktree / cloud | not_run | RC1 범위 밖 |

## AC 명령 절 (receipt 인용용 실측 명령)

AC-08C-5(레인 실행 가능 정적 검증)는 임시 프로젝트에 skill을 복사한 뒤 로컬 검증을 돌린다 — 자격증명 불필요:

```sh
TMP=$(mktemp -d /tmp/e2e-ac5-XXXXXX) && mkdir -p "$TMP/.claude/skills" \
  && cp -R skills/moai "$TMP/.claude/skills/moai" \
  && claude plugin validate "$TMP/.claude/skills"; echo "exit=$?"; rm -rf "$TMP"
```

AC-08C-7의 Go 바이너리 해시는 읽기만 한다(실행 금지): `shasum -a 256 "$(command -v moai)"` — 실행 전후 동일.

AC-08C-9(독립 AUDIT)는 리더 통합 게이트가 진행한다.

### 운영자 수동 단계(AC-08C-8 host 절반, 레인 미실행)

리더 답(2026-11-12, #2014 Q1)에 따라 스크립트로 만들지 않고 여기에만 적는다. 레인은 실행하지 않으며, 결과가 들어오기 전까지 이 항목은 not_run이다.

본문 명령(`cd /tmp/clean && PATH=/usr/bin:/bin claude --bare -p '/moai 현재 저장소 읽기' --output-format json`)만으로는 `--bare`가 `moai` skill을 찾지 못한다. 저장소는 `.claude/skills`를 추적하지 않고, 명령에 `--add-dir`도 없기 때문이다. 그래서 운영자가 다음 수동 단계를 더해 실행할지 정한다. 이 단계를 더하면 본문 AC가 재는 것과 모양이 달라지므로, 결과에도 그렇게 적는다.

1. 체크아웃 밖의 임시 clean 사본을 만든다(예: `git clone --depth 1 <repo> "$(mktemp -d)/clean"`).
2. skill을 그 **임시 사본 안의** `.claude/skills/moai`로 복사한다. 실제 HOME(`~/.claude/skills`)에는 설치하지 않는다. 실제 HOME에 skill을 설치·교체하는 일은 사용자 확인 대기이며, 절차는 #2483에서 구현 중이다.
3. `moai-cli`가 없는 PATH와 임시 HOME으로 `claude --bare -p '/moai 현재 저장소 읽기' --add-dir <임시 사본> --output-format json`을 실행한다.
4. 받은 결과를 receipt 형식으로 남기고, 위 단계가 본문 명령과 다르다는 것을 detail에 적는다.

## 운영자 T2 실행 블록 (paste-ready — #2014 사전 점검 완료 후)

사전 조건: 이 저장소 checkout의 최신 `v2026.11.0`(본 수정안 병합 후), API 키(`ANTHROPIC_API_KEY` — `--bare`는 OAuth를 읽지 않는다. `CODEX_API_KEY` — 격리된 `CODEX_HOME`에는 저장된 로그인이 없다), `jq`, macOS.

실행은 **체크아웃 밖**에서 한다. 아래 블록은 `v2026.11.0`의 트리를 체크아웃 밖 임시 디렉터리로 꺼내 그 사본의 스크립트를 실행하고, receipts도 체크아웃 밖에 둔다. 체크아웃의 작업 트리·branch·stash는 건드리지 않는다. 스크립트는 자기 위치 기준으로 `skills/moai`를 찾으므로 어느 디렉터리에서 실행해도 된다. 키는 `read -s`로 받아 셸 이력에 남지 않게 한다. 자격증명 값은 스크립트가 어떤 파일에도 쓰지 않는다.

```sh
git -C <이 저장소의 checkout 루트> fetch origin v2026.11.0
T2="$(mktemp -d /tmp/moai-t2-XXXXXX)"
git -C <이 저장소의 checkout 루트> archive origin/v2026.11.0 tests/e2e skills/moai | tar -x -C "$T2"
RECEIPTS="$(mktemp -d /tmp/moai-t2-receipts-XXXXXX)"
read -rs -p 'ANTHROPIC_API_KEY: ' ANTHROPIC_API_KEY; echo; export ANTHROPIC_API_KEY   # AC-08C-3·6
read -rs -p 'CODEX_API_KEY (없으면 Enter): ' CODEX_API_KEY; echo; export CODEX_API_KEY  # AC-08C-4
cd "$T2"
bash tests/e2e/host-claude.sh "$RECEIPTS"
bash tests/e2e/host-codex.sh  "$RECEIPTS"
unset ANTHROPIC_API_KEY CODEX_API_KEY
cat "$RECEIPTS/receipts.jsonl"; ls "$RECEIPTS"
```

`$RECEIPTS/receipts.jsonl`의 각 줄이 `schemas/e2e-receipt.schema.json` 형식의 receipt이고, `detail`이 가리키는 원시 출력도 같은 디렉터리에 남는다. 이 디렉터리가 08-D 지원표 갱신 입력이다(AC-08C-3·4·6 행을 not_run에서 실측값으로 교체). 저장소에 옮기기 전에 원시 출력에 자격증명 값이나 개인 경로가 없는지 확인한다.
