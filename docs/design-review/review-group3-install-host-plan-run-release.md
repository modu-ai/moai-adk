# 설계 하위 문서 7종 검토 — 설치·Host·릴리스·PLAN·RUN·소유권·GitHub 작업 구조

- 작성: 리더가 위임한 읽기 전용 검토 세션, 2026-11-09. 파일 쓰기는 이 문서 하나뿐이다.
- 대상: `references/chatgpt-space/` 아래 텍스트 사본 7건(2026-11-08T23:19:26Z 수집). 각 사본의 머리말(제목·출처·sha256)은 그대로 두었고, 인용은 사본의 행 번호(`L123`)로 한다. [83f]는 사본이 두 번 이어 붙어 있어(§1.2) §1–§6과 §7 시험 앞 3건은 첫 사본(L10–L1257), §7 나머지 시험·§8·L2453은 두 번째 사본에만 있으므로 그 행을 인용한다.
- 대조 기준: ① 상위 설계 `page_5490c4ee…md`(22절, 절 머리 행: §1 L11·§2 L122·§3 L279·§4 L355·§5 L466·§6 L604·§13 L1340·§14 L1397·§15 L1576·§17 L1763·§18 L1921·§20 L2111) ② `docs/decisions-2026-11-08.md`(이하 decisions:행) ③ RC1 소스 `origin/v2026.11.0`을 **`9bfd26d3b`에 고정해** `git show`로 읽음(검토 도중 tip이 PR #2130 병합으로 `d253f4a81`로 이동했으나 재검토하지 않음) ④ Go 기준선 `moai-adk-go@4fa2aa9` 파일 수·바이트 수. 실행·빌드·테스트는 하지 않았다.
- 선행 검토와의 관계: 상위 문서 검토(A1–A8·B1–B8)와 RC1 대조표(절별 상태표·충돌 4건·후속 후보)는 반복하지 않는다. 그 두 문서가 이미 짚은 항목은 "(선행 A2)"처럼 표시만 하고 하위 문서의 행 번호를 보탠다.
- 표기: "근거 없음"은 RC1 소스·결정 기록·Go 기준선 어디에서도 확인하지 못했다는 뜻이다. 문자열 검색 0건은 가설이며, 열거된 목록(예: `src/main.rs`의 subcommand 전체, `Cargo.toml`의 의존성 4개)과 대조한 경우에만 "없다"고 단정한다.

문서 약호: [83f] 홈 설치와 프로젝트 생성 기능 계약 · [dff] Host adapter와 hook 이식성 · [25f] 릴리스와 Desktop 상세 설계 · [e08] PLAN과 GEARS SPEC 승인 계약 · [3d9] RUN과 TDD DDD 구현 검증 흐름 · [487] 파일 소유권과 worktree 수명 · [66a] RC1 GitHub 작업 구조와 실행 계획.

---

## 1. [83f] MoAI 홈 설치와 프로젝트 생성 기능 계약 (`page_83f162ef…`, 2,621행)

### 1.1 문서 개요

사용자가 MoAI를 한 번 설치한 뒤 Claude Code·Codex 안에서 자연어로 SETUP·UPDATE를 끝내는 제품 계약이다. 홈 패키지와 host별 독립 복사(§1), 프로젝트 생성 파일 목록(§2), 기능 ID F01–F25의 입출력·경계(§3), SETUP 인터뷰 순서(§4), UPDATE와 3-way 보존(§5), host adapter·hook 계약(§6), 8단계 상태와 수용 시험 12건(§7), 상세 문서 링크(§8)로 구성된다. 설계 원칙(symlink 금지·host별 성공 분리·파일 생성 ≠ 활성화)은 RC1 구현 방향과 같다.

### 1.2 내부 결함

- **사본 중복**: L1258–L1260에서 페이지가 다시 시작해 L1262–L2575가 L10–L1257을 그대로 반복한다. §8(L2575 이하)은 두 번째 사본에만 있다. 설계 결함이 아니라 수집 결함이므로 인용은 첫 사본 기준으로 한다(§7 한계 참조).
- **깨진 링크 자리표시**: L1260·L2453·L2577–L2587의 "페이지 불러오는 중…"이 링크 제목을 대신한다. L2453은 "Hook 96개 … 네 가지 fallback은 페이지 불러오는 중…에 있다"로, 참조 대상이 사라졌다.
- **출처 없는 수치**: L1214 "Hook 96개" — Go 기준선 `.claude/hooks/` 파일 82개, `settings.json` 등록 command 82개와 맞지 않는다(96의 집계 기준 미표기). L1169 "기존 38개 tool"은 이 세션에 노출된 moai MCP 도구 38개(서버 v3.2.1-rc.6)와 일치하나 문서에 출처가 없다.
- **상태 어휘와 수용 시험의 단절**: L1223의 8단계(discovered→…→host_verified)는 §7 시험 12건(L1234–L2573) 어디에서도 통과 조건으로 쓰이지 않는다. 시험은 "0건"·"diff 0" 같은 결과만 적고 어느 상태에서 측정하는지 없다.
- **측정 불가 통과 조건**: L1256 SETUP-PROFILE "실제로 다른 최소 파일 plan" — '최소'의 기준이 없다. L2573 CROSS-HOST-PARITY "같은 계약·scope·evidence를 요구" — 비교 절차·fixture 미정.
- **예시 envelope와 본문 불일치**: L1095–L1105 `generation.plan` 예시의 `inputs.profile`·`task_backend`·`scope` 키는 §2 표의 `project.yaml` 설명(L295–L307)과 대응 관계가 없다. "설계 예시"라고 밝혔으나(L1107) 실제 schema가 어디 있는지도 없다.
- **F ID 정의 불일치(상위 문서와)**: L941 F20 "API 예산과 인증", L983 F22 "저장과 복구", L1004 F23 "릴리스와 문서", L1025 F24 "Desktop", L1046 F25 "공식 변경과 호환" — 상위 §20은 F20 판단 backend·F22 ADE·F23 Desktop·F24 유지보수/진단·F25 보관/이행(L2407–L2434). [66a] L556이 이 차이를 경고했지만 두 문서 모두 고치지 않았다.
- **"Codex Desktop"과 "Codex" 혼용**: L10은 Codex Desktop, L105·L1355는 `~/.agents/skills/moai/SKILL.md`(CLI·홈 경로). Desktop의 skill 경로가 같은지 확인한 기록이 없다.

### 1.3 상위 문서와의 불일치

- 홈 배치: 상위 §13 L1365–L1383은 `~/.moai-cli/{config/,installs/<host>/<version>/,adapters/,projects/<id>/{objects,manifests,brain,specs,workspaces},cache/}`. 하위 L37은 `<state_root>/skills/moai/<version>/`, L139 `<state_root>/installations/`, L156 `projects/<project_id>/registry.json`, L173 `tasks.md`, L190 `runs/<run_id>/`, L207 `artifacts/·indexes/`. 디렉터리 이름이 `installs`↔`installations`, `workspaces/journal.jsonl`↔`runs/`로 다르다(선행 A2의 하위 문서 쪽 근거).
- 프로젝트 배치: 상위 §13 L1349–L1356 `.moai-cli/{project.yaml,manifest.json,contracts/{spec-index.json,policy.json},specs/,teams/}`. 하위 L312 `.moai-cli/contracts.lock`, L329 `policies/`, L363 `workflows/`, L380 `manifests/`. `teams/`는 하위에 없고 `contracts.lock`·`policies/`·`workflows/`는 상위에 없다.
- 단계 이름: 하위 L58은 READ·SETUP·UPDATE·PLAN·RUN·VERIFY·AUDIT·SYNC 8개(RC1 `routes.json`과 동일). 상위 §2는 UPDATE·READ가 없다(선행 A7). 이 하위 문서 쪽이 맞다.
- 상위 §1 L100 "commit·push·Issue 공개·PR·merge·release를 같은 승인으로 합치지 않음"을 하위 F17 L894·F18 L915가 반복하나, RC1의 게이트 수준 merge 자율화(decisions:39)는 어느 쪽에도 없다(선행 충돌 #1).

### 1.4 RC1 결정·실물과의 불일치

- 홈 패키지 루트 `<state_root>/skills/moai/<version>/`(L37): RC1에는 해당 상수가 없다. `package-resolve`가 `--repo_package`/`--user_package` 경로를 인자로 받는다(`src/main.rs:112-118`). 홈 상태 디렉터리는 `installations/`만 있다(`src/install.rs:4`, `src/update.rs:18`).
- `registry.json`·`tasks.md`·`runs/`·`artifacts/`·`indexes/`(L156–L219): `src/` 문자열 검색 0건(가설). `tasks.md`는 호출자가 경로를 넘기는 입력 파일이며 홈에 두지 않는다(`src/forge/promote.rs:4,50-51`; decisions:69).
- `.moai-cli/contracts.lock`·`policies/`(L312–L341): `src/` 검색 0건. RC1 승인 snapshot은 05-A `Snapshot{reference,digest}`와 05-B `Approval`이며 파일로 저장하지 않는다(`src/plan.rs:6` "nothing is stored").
- UPDATE 3-way·known-good·per-file receipt(L1086–L1090): 구현됨. `src/update.rs:5-15`(baseline=직전 host receipt의 ownership manifest), `:332-344`(known-good 사본), `src/recovery.rs:9-21`(matches_after/before/conflict/missing). 다만 L1090 "부분 실패 시 성공 host 유지"는 RC1이 "pending host는 새 UPDATE 거절" 조건을 더한다(`skills/moai/workflows/update.md` 절차 6) — 문서에 없다.
- symlink 금지(L134·L238): RC1 일치. symlink projection은 `blocked`(`src/update.rs:22`), 설치 디렉터리 symlink 거절(`src/install.rs:309-310`).
- Claude `AGENTS.md` 로딩 조건 doctor(L273): 구현됨(`src/claude.rs:5-14`, v2.1.277, 억제 파일 3종).
- `.claude/settings.json` structured merge(L473): 02-B가 `outputStyle`·`statusLine` 두 키만 관리(`src/claude.rs:269`). 문서의 "hook·style·statusline"(L469) 중 hook은 RC1 0개(decisions:67).
- `.codex/config.toml`·`hooks.json`(L482–L494): RC1은 TOML을 읽지도 쓰지도 않고(`src/hosts/codex.rs:16-18,31-33`; decisions:66–67) hook 0개. 문서는 "probe를 통과한 키만"이라 모순은 아니나 RC1 상태(not_evaluated)가 없다.
- `.moai-cli/project.yaml`(L295–L307): 구현됨. 다만 L1101 예시 키 `profile`·`task_backend`는 RC1 관리 키 10개(`src/profile.rs:31-42`)에 없다. 실물 파일은 JSON 본문이다(`.moai-cli/project.yaml` 1–17행).
- 생성 파일 `.claude/agents/<role>.md`·`.codex/agents/<role>.toml`·`.moai-cli/workflows/<id>/`(L431–L460, L363): 구현됨(`src/generation.rs:11-14`). Codex 역할 파일은 `name`·`description`·`developer_instructions`만(`:29-30`).
- 외부 skill "공식 폴더에 설치한 뒤 전체 읽기"(L1072): 구현됨(`skills/moai/workflows/setup.md` 절차 3: `skill-source-plan/apply`, `skill-read-receipt`, `manifests/sources.json`).
- F23 "M0 spike → … → release candidate 또는 blocked"(L1012): RC1 verdict는 `release blocked`(`docs/release-checklist.md:3`, `docs/rc-release-review.md:17`) — 문서 상태 어휘와 일치.
- MCP "기존 38개 tool … 선택 소비자 계약"(L1169): RC1은 CLI만, MCP는 다음 버전(decisions:51; AC-01C-2/3/4 not_applicable). 문서에 RC1 범위 표시가 없다.
- 로컬 TODO 선택 승격(L2615–L2621): 구현됨(07-B, `todo-promote-preview/apply`, `src/main.rs:164-169`).
- Go 전환 계획(L2611 "테스트 완료 후 기존 Go 설치를 제거하고 … moai 명령으로 전환"): decisions:91 ②로 04-D는 RC1 밖, Go 상태 무변경 유지. "따른다"는 서술이 RC1에서는 적용되지 않는다.

### 1.5 낡은 내용

- L1217 "기존 exit 0 no-op은 MoAI 추가 검사 미실행"·L1214 Hook 96개: Go 세대 hook 서술이며, RC1은 Claude hook 0개·Codex hook 0개로 결정했다(decisions:66–67). 설치 문서의 §6 hook 절반은 RC1에 적용 대상이 없다.
- L18 `~/moai/projects·~/moai/worktrees … 바뀌지 않음`: Go 기준선에는 `~/.moai/worktrees/`만 있고(`go@4fa2aa9 internal/cli/cc.go:304`) `~/moai/`는 없다. RC1 worktree 위치는 `<primary>/.moai-cli/worktrees/<issue>-<key>`(`src/worktree.rs:12`; decisions:34). 존재하지 않는 기본값을 "유지"한다고 적었다.
- L1223 8단계 상태: RC1 receipt 상태는 `pending|committed|recovered`(`src/update.rs:251-252`), 파일 판정은 `create/merge/unchanged/conflict`. 8단계는 어디에도 없다.

### 1.6 개선 제안(영향 순)

1. §1 표의 홈 경로를 RC1 실물(`installations/<install_id>.json`, `installations/<id>/hosts/<host>.json`, `…/known-good/<hex>`)로 바꾸고, 홈 패키지 루트·registry·runs는 "후속(P09)"으로 표시한다.
2. §6 hook 절에 "RC1: Claude hook 0·Codex hook 0·TOML 미처리" 한 줄을 넣어 §7 HOOK-EVENT-PROBE가 RC1 시험이 아님을 밝힌다.
3. F20·F22–F25 이름을 상위 §20과 한쪽으로 맞춘다.
4. §7 시험마다 8단계 중 어느 상태의 결과인지와 RC1 대응 시험(`tests/e2e_host_fixture.rs`, `tests/apply_recover_partial.rs`, `tests/claude_update_test.py`)을 적는다.
5. L1101 예시의 키를 `src/profile.rs` 관리 키로 바꾸거나 "schema 미정" 표시를 붙인다.
6. 수집 사본의 중복 구간을 제거하고 §8 링크 제목을 복원한다.

### 1.7 미확인·한계

- 두 사본 구간의 바이트 동일성은 눈으로만 확인했다(diff 미실행).
- "Hook 96개"의 집계 기준(Go 세대 hook 스크립트+내부 handler?)은 확인하지 못했다. 82는 단순 파일 수다.
- §8 링크 6건의 대상 페이지는 제목이 지워져 어떤 하위 문서인지 알 수 없다.

---

## 2. [dff] MoAI Host adapter와 hook 이식성 (`page_dff088ea…`, 942행)

### 2.1 문서 개요

공통 MoAI가 의미를 소유하고 host adapter는 발견·호출·native 기능 차이만 다룬다는 경계(§1), capability 상태 모델 unknown→documented→probed→certified(§2), Claude/Codex의 AGENTS·skill 발견 규칙(§3), rules·subagent·hook·style·statusline의 native 대응(§4), local/cloud hook 지원표(§5), no-binary fail-safe(§6), MCP 선택 1급 adapter(§7), 기존 구현 차이(§8), HA-01~10 수용 시험(§9), 뒤에 공통 UX 결정표와 공식 문서 확인표가 붙는다. 외부 공식 문서 확인일을 적고 "실제 앱 시험은 아직"이라고 분명히 밝힌 점이 좋다.

### 2.2 내부 결함

- L942 "20개 후속 시험은 모두 미실행": 본문의 시험 ID는 HA-01~10(10건)과 SOURCE-DEFAULT 등 8건으로 18건이다. 20의 목록이 없다.
- L483 "13 harness 소비자": 집계 기준·목록 없음. 같은 문장에서 "M0 inventory에서 재확인"이라 했으나 M0 결과가 어디 있는지 없다.
- L573 "Codex t1909 제거 이력 주장": `t1909`가 무엇(commit? ticket?)인지 정의 없음.
- L232 Claude 동명 skill 우선순위 "enterprise → personal → project": 출처 링크 없음(공식 문서 행 미인용).
- L438 Codex hook 세부(prompt/agent handler parsed-but-skipped, SessionEnd async 무시): 외부 사실인데 근거 절(L721)은 제목만 있고 인용 위치가 없다.
- L719 비교 지표(p50/p95 latency 등)에 baseline·측정 환경이 없어 §9 시험과 연결되지 않는다.
- 두 상태 모델 혼재: L86 capability는 4단계(+stale), L440 hook은 "binary 설치·hook 파일 발견·event 호출·verdict 적용" 4검사. 둘을 잇는 표가 없다.
- L742 "Codex hook 12개 event adapter"는 검증된 수치다(`go@4fa2aa9 .codex/hooks.json` 이벤트 12종). 이렇게 맞는 수치와 L483·L942의 근거 없는 수치가 같은 어조로 섞여 있다.

### 2.3 상위 문서와의 불일치

- 상위 §4 L462 "기존 소유권 문서에 남은 선택 symlink 표현은 계승하지 않음" — 이 문서 L248 "외부 symlink가 있어야 읽히는 생성물은 portable 실패"와 일치. 반면 [487] L359가 조건부 링크를 남겨 두어 상위가 가리킨 "기존 소유권 문서"가 바로 [487]이다(§8 교차 모순 참조).
- 상위 §1 L120 "G 코드는 특정 제품에 매핑하지 않음" ↔ 이 문서 L835 "ZCode는 사용자가 지정한 Z.ai 제품으로 식별" — 상위가 고쳐지지 않은 채 하위만 앞서 갔다(선행 보조 관찰).
- 상위 §2는 Model/Storage Adapter 계층을 두지만 이 문서 §7 MCP·L16 "얇은 adapter"는 CLI/MCP transport만 다룬다. 계층 이름이 서로 참조되지 않는다.
- HA-01~10(L592–L717)은 상위 §17 시험군(HOST SETUP UPDATE 등)과 ID 체계가 다르고 대응표가 없다.

### 2.4 RC1 결정·실물과의 불일치

- Claude `AGENTS.md` v2.1.277+·억제 파일 3종·`~/.claude/CLAUDE.md` 예외(L130, L242): 구현 일치(`src/claude.rs:6-9,29,32,148`).
- Codex override→AGENTS→fallback, 32 KiB(L244, L880): 구현 일치(`src/hosts/codex.rs:8-11,29`).
- L884 "약65.6KB MoAI style·약22.5KB AGENTS": Go 기준선 바이트 65,627·22,505와 일치(`go@4fa2aa9 .claude/output-styles/moai/moai.md`, `AGENTS.md`).
- L854 "기본값은 MoAI-Easy": RC1 02-B 시험이 `outputStyle: "MoAI-Easy"`를 기본 예로 사용(`src/claude.rs:731-736`). 일치.
- Statusline(L327, L906): RC1은 `statusLine` 키 소유권만 다루고 renderer 없음(`src/claude.rs:269`); Codex `tui.status_line`은 식별자 미열거로 쓰지 않음(`src/hosts/codex.rs:33`). 문서의 "Codex 기본14항목"과 Go의 "29 canonical Codex TUI" allowlist(`go@4fa2aa9 internal/codexwiring/configtoml.go:32`)가 다르다 — 14의 출처 불명.
- Hooks "양사 현재 지원 … 최소 집합 opt-in"(L301, L371): RC1 결정은 Claude hook 0·Codex hook 0(decisions:66–67). HA-05·HA-07·HA-08은 RC1 시험 대상이 없다.
- MCP 선택 1급·HA-09 golden fixture(L477–L563, L700): RC1은 MCP 다음 버전(decisions:51). HA-09는 RC1에서 not_applicable인데 §9에 그 표시가 없다.
- Codex Desktop 발견·SETUP·UPDATE 실패 시 "RC release blocked"(L573): RC1 실측은 Codex CLI(`tests/e2e/host-codex.sh`, `CODEX_API_KEY`)이며 Desktop·CLI 모두 not_run(decisions:66; `docs/support-matrix.md:28`). Desktop 자체를 시험하는 스크립트가 없다.
- HA-02 "재실행 diff 0"(L613): passed receipt 있음(`docs/support-matrix.md:17`). HA-01·03·04·06·10은 receipt 없음 — 이 문서가 RC acceptance라 부른 10건 중 RC1 receipt는 1건이다.
- "Codex 재지원 spike를 독립 gate"(L573): RC1에서 02-C(#1993)로 수행, 통합됨(decisions:74). 문서는 미착수 어조.
- L569 "현재 공개 main의 CLAUDE.md는 AGENTS import 포함": Go 기준선 `go@4fa2aa9`에는 루트 `CLAUDE.md`가 없다(`git ls-tree` 부재). 어느 main·어느 경로를 말하는지 불명.

### 2.5 낡은 내용

- §5 지원표(L363–L436)와 §8(L569–L571) handle-pre-tool.sh·session-start 서술은 Go 세대 hook 전제다. RC1이 hook을 하나도 싣지 않기로 한 뒤에는 "RC 처리" 열 전체가 미래 버전 이야기다.
- L573 "Codex 재지원 spike" 미완 어조 — 02-C·02-D 완료(decisions:63, 74).
- L16 "service-free는 MCP까지 제거한다는 뜻이 아니다"는 맞지만, RC1에서 MCP가 범위 밖이라는 사실(decisions:51)이 없다.

### 2.6 개선 제안(영향 순)

1. §9 각 HA 행에 RC1 상태(passed receipt / not_run / not_applicable)를 붙이고 `docs/support-matrix.md`를 가리킨다.
2. hook·MCP 절 머리에 "RC1 범위 밖(hook 0개, MCP 다음 버전)" 배너를 둔다.
3. L942 "20개"를 실제 ID 목록으로 바꾸거나 수를 고친다.
4. "Codex Desktop"과 "Codex CLI"를 구분해 쓰고, Desktop 실측 스크립트 부재를 명시한다.
5. 검증된 수치(12 events, 65.6KB, 22.5KB, 32 KiB, v2.1.277)와 미검증 수치(13 harness, 14항목, t1909)를 표기로 구분한다.

### 2.7 미확인·한계

- Claude/Codex 공식 문서(skill 우선순위, Codex hook handler 종류, personality deprecated)는 재조회하지 않았다.
- "13 harness 소비자"·"Codex 기본14항목"은 Go 소스에서 찾지 못했다(검색 0건, 가설).

---

## 3. [25f] MoAI 릴리스와 Desktop 상세 설계 (`page_25fa9b23…`, 751행)

### 3.1 문서 개요

제품명·경계·출시 원칙(§1), Rust domain/process 인터페이스(§2), crate 8개 소스 구조(§3), 최초 추가·UPDATE 상태전이(§4), 배포 version과 Go 3.x 경계(§5), legacy 이행(§6), M0–M5 단계 gate(§7), 문서 개편(§8), 후속 Desktop(§9), RL-01~10·DT-01~02 수용 시험과 미결(§10), 끝에 격리 시험·설치 전환 계약이 붙는다. Go updater의 feed 조회 방식(§5 L203)을 소스 근거로 적은 점은 RC1 04-A 결정과 정확히 맞물린다.

### 3.2 내부 결함

- L745 "P17에 따라": 결정 기록에는 P04–P15만 있고(decisions:94–96) P17 정의가 어디에도 없다.
- L114 vs L157–L160([e08])·L267–L272([3d9]): 같은 "내부 예시 명령"이 문서마다 다르다(`setup --dry-run`, `maintenance update`, `project init --apply-plan` ↔ `spec plan`, `task submit`). 한 곳에서 정한 명령 체계가 없다.
- L126–L150 crate 8개 구조와 L152 "모든 crate를 publish한다는 뜻 아님": 경계라고 하면서 §7 gate나 §10 시험 어디에도 crate 분리 여부를 묻지 않는다.
- L273 artifact manifest 필수 필드 11개 — §4 L189 apply plan 필드 9개와 겹치는 항목(target/protocol/digest/rollback)의 관계가 없다.
- RL-02 "channel 오염·이중 updater 0건"(L580)·RL-07 "target별 restore 검증"(L645): target 목록(P12 HOLD)이 없어 측정 범위가 비어 있다.
- L197 "v2026.11.0-rc.1 tag는 후보"와 L712 "실행 이름 moai-cli … 확정" — 같은 문서 안에서 tag는 미정, 이름은 확정인데 §10 "정확한 repository/tag/feed 식별자는 구현·검증으로 구체화"(L712)와 겹쳐 세 번 다른 확정 수준으로 말한다.
- L8 "Codex Desktop", L16 "Codex Desktop에서 native $ skill 선택" — [dff]와 같은 혼용.

### 3.3 상위 문서와의 불일치

- 상위 §18 L1951 "실제 RC repository·배포 artifact·signing·feed·platform 지원"을 미결로 둔 것과 같은 상태(L154, L712). 둘 다 decisions:7–10의 결정을 반영하지 않았다(선행 A1).
- 상위 §13 L1367 `installs/<host>/<version>/`와 이 문서 L743 "별도 product/channel·manifest·journal·lock" — 설치 상태의 파일 위치를 이 문서는 전혀 적지 않아 상위와 대조할 수 없다.
- 상위 §14 Pi/Desktop(L1397)과 §9(L518–L540)는 Electron·Pi 선택이 같다. 다만 상위 §1 L109 "RC1 선행조건 아님"에 대응하는 배너가 §9 머리에 없고 M5(L494)에서만 보인다(선행 B5와 같은 성격).
- 상위 §5의 SYNC drift 7분류(L586–L600)와 이 문서 M2 "한 SPEC 계획→…→Issue 완료"(L446)는 SYNC 완료 조건을 다르게 둔다.

### 3.4 RC1 결정·실물과의 불일치

- crate 8개(L126–L150): RC1은 단일 crate `moai-cli`(`Cargo.toml:1-5`; decisions:54 "단일 crate `src/spec.rs`"). `crates/moai-evidence/`는 [e08] L129에도 남아 있다.
- 내부 명령(L114): `setup`·`project init`·`maintenance update`·`doctor`·`capabilities`는 `src/main.rs` subcommand 목록(ProjectPlan/ProjectApply/UpdatePlan/UpdateApply/RecoveryPlan/RecoveryApply/ClaudeDoctor/CodexDoctor/InstallInspect 등 50여 개)에 없다. 열거 대조이므로 "명칭 불일치"로 단정한다.
- tag "v2026.11.0-rc.1 후보"(L197): 결정은 `moai-cli-v<version>`(decisions:9; `docs/release-checklist.md:8`). Go stable `/releases/latest`와 RC `v`/`go-v` 필터 모두 피한다는 근거는 이 문서 L203과 같다.
- cargo-dist "packaging 후보"(L271): 결정 A(decisions:7)인데 RC1 체크리스트는 "cargo-dist 미도입 — 수동 sha256.sum"(`docs/release-checklist.md:29`). 결정됨·미구현 상태를 문서가 모른다.
- Windows 실행 중 교체 "실제 artifact로 검증"(L271): 결정은 self-replace crate 1.5.1(decisions:12)이나 `Cargo.toml` 의존성 4개(clap·serde·serde_json·sha2)에 없다. 결정됨·미구현.
- 서명 "signature/digest/provenance"(L187, L273, L645): 결정은 attestation 단독(decisions:8), 생성 실측 없음(`docs/release-checklist.md:34`). L273 "checksum이 서명을 대체하지 않음"은 REQ-08D-08과 일치(`:35`).
- 설치 receipt(L273 필드 11개): RC1 `InstallReceipt`는 schema_version·product·generation·channel·version·binary_path·binary_sha256·installer_owner·protocol_version·created_at·updated_at(`src/install.rs:54-68`). target triple·minimum OS·signature·rollback compatibility 없음. `InstallerOwner`는 direct/cargo/homebrew/msi/unknown(`:72-78`) — L187 "package-manager가 update owner" 개념은 enum으로만 존재.
- Go updater 조회 경로(L203): RC1 특성화 시험 `go_updater_feed_contamination_table`(`tests/install_cli.rs`, `docs/release-checklist.md:44`)과 fixture `tests/fixtures/go-releases-sample.json`. 일치.
- `.moai/user-assets.json`(L208): Go 기준선에 존재(`go@4fa2aa9 internal/cli/bundle.go:40`). 일치.
- 이행 순서 "opt-in migration → default 전환"(L395)·"명령 전환"(L749)·RL-COMMAND-CUTOVER(L751): 04-D는 RC1 밖(decisions:91 ②). RC1에는 cutover 시험 없음.
- M1 "양쪽 모두 필수. Claude-only 대체 불가"(L437)·L20: RC1 verdict `release blocked`의 gap #5가 정확히 이 조건(`docs/rc-release-review.md:27`). 일치.
- M2 service-free 완주(L446): 제품 경로 GitHub 쓰기는 not_run(선행 §5). 일치하나 문서에 상태 없음.
- 독립 감사 필수(L504): RC1 05-D `audit::reduce`가 자기 보고 receipt를 증거로 세지 않음(`src/audit.rs:5-9`). 일치.
- RL-01 "installer shell 암기 요구 없음"(L567)과 decisions:10 "installer가 PATH 안내": PATH 설정을 사용자가 해야 하면 "암기 없음"의 기준이 애매하다.
- SBOM·license review(L156): HOLD(`docs/release-checklist.md:56`). 일치.

### 3.5 낡은 내용

- L197·L154 tag/repo 미결 → decisions:9(같은 repo, `moai-cli-v*`)로 결정.
- L199 npm dist-tag 서술: Rust 바이너리+cargo-dist 결정 뒤에는 적용 대상이 없다.
- L271 cargo-dist "후보" → 결정 A.
- L8 "Codex Desktop … SETUP·UPDATE" 필수: 실측 대상이 CLI 스크립트로 바뀌었다(`tests/e2e/host-codex.sh`).

### 3.6 개선 제안(영향 순)

1. §5·§10의 tag·서명·packaging을 "결정됨(decisions:7–10)/구현 상태(`release-checklist.md` (a)(d)(e)(g))" 두 열로 다시 쓴다.
2. L114 내부 명령 예시를 `src/main.rs` 실제 subcommand로 바꾸거나 삭제한다.
3. §3 crate 구조를 "단일 crate(RC1) → 분리 후보"로 고친다.
4. "P17" 정의를 넣거나 삭제한다.
5. §9 머리에 "RC1 비범위" 배너를 둔다.
6. RL-GENERATION-ISOLATION 등 4개 시험(L751)에 RC1 대응 시험(`tests/install_cli.rs`)과 not_run 여부를 적는다.

### 3.7 미확인·한계

- Electron/Pi/Tauri 외부 링크(L729–L731)와 SemVer·Go version query 문서는 재조회하지 않았다.
- `tests/install_cli.rs` 본문은 읽지 않았고 체크리스트 인용으로만 확인했다.

---

## 4. [e08] PLAN과 GEARS SPEC 승인 계약 (`page_e0874134…`, 407행)

### 4.1 문서 개요

PLAN의 책임 경계(§1), Issue 초안과 불변 승인 snapshot의 분리(§2), GEARS 문법과 acceptance 연결(§3), PLAN 상태 전이(§4), 7단계 로직(§5), SPEC·plan schema 예시(§6), 내부 CLI(§7), 변경·예외·waiver(§8), evidence 연결(§9), RC1 acceptance PLAN-01~08(§10). 7개 하위 문서 가운데 RC1 05-A·05-B 구현과 가장 가깝다.

### 4.2 내부 결함

- L26 "snapshot은 기존 Git commit 또는 artifact 저장소에 보존" ↔ L28 "cache는 원본이 아니며 원본을 찾지 못하면 승인 완료를 주장하지 않음": 어디에 어떤 형식으로 저장하는지(경로·파일명) 없음. PLAN-08 "별도 서비스 없이 재개"(L399)가 무엇을 다시 읽는지 정해지지 않았다.
- L58–L79 상태 10개(Draft…Cancelled)와 L139 validation 4축, L135 `status: "ready"`가 서로 다른 어휘다. 상태 도식의 Ready와 schema의 `ready`만 일치.
- L48 "필요한 시간·용량·오차 한도는 정량화한다" — 정량화 대상을 고르는 기준이 없다(모든 REQ인지 일부인지).
- L113 `"ref": "owner/repo#42"` 예시가 [66a] L1166 "#42만으로 키를 만들지 않는다"와 어긋나 보인다(owner/repo가 붙어 있으므로 위반은 아니나 설명이 없다).
- L20 Jev 서술: "제품을 미식별이라고 확정하지 않음"과 "특정 엔진에 종속시키지 않음"이 한 문단에 있어 결론이 없다.
- L52 "GEARS 저자 원문 · EARS 출판 기록" 링크 제목만 있고 판 번호가 없다(RC1은 `gears@2026-01-14`, `ears@2009`로 고정 — `src/spec.rs:19-21`).

### 4.3 상위 문서와의 불일치

- 상위 §5 L466 이하는 PLAN·RUN·VERIFY·AUDIT·SYNC를 한 절에 두고 "승인 snapshot·task DAG admission"을 말한다. 하위 §4의 10개 상태는 상위에 없다.
- 상위 §6 L756의 4필드(assignee/execution_owner/provider_session/integration_owner)가 하위 L140 `approval.reference`·L99 "현재 owner"와 연결되지 않는다.
- 상위 §1 L100 승인 분리 원칙은 하위 L32 "가역적 작업을 반복 질문하며 막지 않음"과 긴장 관계이나 어느 쪽도 다른 쪽을 인용하지 않는다.

### 4.4 RC1 결정·실물과의 불일치

- 4축 검사(L95, L139 syntax/structure/semantic/execution): 구현 일치. `Semantic{Unknown,Reviewed}`, `Execution{NotRun}`에 passed 값이 없다(`src/spec.rs:3-6,420-429`).
- GEARS 절 분해(L41–L44): 구현 일치. RC1은 `where`·`while`을 배열로 허용하고 순서(Where→While→When)를 강제한다(`src/spec.rs:33-37,248`). 문서는 각 1개처럼 쓴다.
- EARS import "원문·ID·source range·분류 보존, IF→WHEN 기계 치환 금지"(L200–L217): 구현 일치(`src/spec.rs:343-376`, `syntax: "EARS-import"`).
- 승인 snapshot digest(L130): RC1은 `SortedKeys` 정렬 직렬화 임시, JCS 후속(decisions:54; `src/spec.rs:8-10`). 문서에 canonicalization 언급 없음.
- snapshot 저장(L26): RC1 PLAN은 "nothing is stored"(`src/plan.rs:6`; decisions:62 "plan 결과 제품 저장 없음"). PLAN-08 재개 fixture가 RC1에 없다.
- expected revision(L138): RC1은 `HEAD^{commit}`(`src/plan.rs:58,463-470`), dirty overlay는 gap(decisions:62).
- admission 어휘(L135 `ready`): RC1 `ready|stale|blocked`(`src/plan.rs:4-5,107-108`), stale 사유 `snapshot-digest-mismatch`·`approval-reference-mismatch`·`expected-revision-mismatch`, blocked 사유 `not-approved`·`approval-revoked`·`approval-scope`(`:381-400`). 문서 §8 표(L183–L269)와 방향 일치하나 어휘 대응표 없음.
- task DAG(L93): Kahn 순서 + Tarjan SCC(`src/plan.rs:89-92`; decisions:62). 문서는 "dependency"만 말한다.
- 내부 명령(L157–L160): RC1 `plan-admit`(op `plan.admit`), `snapshot-describe`(`src/main.rs:402,411`). `spec plan`·`task status`·`skill show`·`capabilities` 없음(열거 대조).
- Jev(L20): RC1 07-D는 `native` 판단만, 별도 backend는 P10 HOLD(`src/decision.rs:1-6`). 문서 어조와 일치.
- 승인 "approval 레코드 … host의 native 정책과 실제 grant를 충족해야"(L30): RC1 `Approval` doc "never grants host permissions"(`src/plan.rs:37`). 일치.
- PLAN-01 "양 host 같은 schema"(L308): host별 fixture 없음(근거 없음). PLAN-03 revision mismatch(L338): 판정 경로는 있음(`src/plan.rs:393-394` `expected-revision-mismatch`); `tests/fixtures/plan/`의 fixture는 dag_cycle·dag_linear·dag_tie·dangling_dependency·scope_overlap·verifier_missing으로 revision 전용은 없고, `tests/plan_admission.rs`가 다루는지는 본문 미확인. PLAN-02 legacy corpus(L325): `tests/spec_contract.rs` 존재(본문 미확인).
- L285 "modality lint는 얕은 검사": RC1은 lint가 아니라 파서(`parse_gears`)로 구조를 확정한다 — 문서의 Go lint 서술은 RC1과 무관.

### 4.5 낡은 내용

- L285–L287 기존 lint·SPEC workflow 링크는 Go 세대. RC1 05-A가 대체했다.
- L162 "현행 moai plan report 유틸·/moai plan skill"은 Go 명령. RC1 route PLAN은 `requires_binary: false`(`skills/moai/references/routes.json`).

### 4.6 개선 제안(영향 순)

1. §2에 "RC1: snapshot은 저장하지 않고 요청마다 재계산(`plan.rs:6`), 재개는 Issue 본문+`HEAD^{commit}` 재계산"을 적고 PLAN-08을 그에 맞게 고친다.
2. §4 상태 10개를 RC1 `ready|stale|blocked` + 사유 코드와 대응표로 잇는다.
3. §6 digest에 "정렬 직렬화 임시, JCS 후속(P06)"을 넣는다.
4. §7 명령을 `plan-admit`·`snapshot-describe`로 바꾼다.
5. GEARS 예시에 where/while 복수 허용과 절 순서 규칙을 적는다.

### 4.7 미확인·한계

- `tests/spec_contract.rs`·`tests/plan_admission.rs`와 `src/spec.rs` 단위 시험 본문은 읽지 않았다.
- GEARS 참조 페이지(sublang.ai 2026-01-14)와 Spec Kit·OpenSpec 링크는 재조회하지 않았다.

---

## 5. [3d9] RUN과 TDD DDD 구현 검증 흐름 (`page_3d917728…`, 535행)

### 5.1 문서 개요

RUN의 입출력(§1), 변경 유형별 전략표(§2), 상태 전이와 단일 실행 소유자(§3), 7단계 로직(§4), 테스트 무결성(§5), DDD와 동작 보존 병행(§6), 요청·evidence schema(§7), 내부 CLI(§8), 복구·재개 표(§9), RUN-01~09(§10). 7개 문서 가운데 RC1 코드와 가장 정확히 겹친다 — 전략 6종·evidence 이름·method 필드·producer 구분이 `src/run.rs`와 1:1이다.

### 5.2 내부 결함

- L165 task 상태 Ready→Running→Review→Verified(+Blocked/Ready/Cancelled)와 L146–L163 도식 노드 이름이 다르고, [66a] L1256의 claim 상태나 [e08] L58의 PLAN 상태와도 연결되지 않는다.
- L167 "로컬 병렬 worker는 작은 durable journal과 OS lock"과 L169 "전역 fencing이라는 주장이 아님" 사이에 cross-host 조정 방법이 비어 있다(RC1은 Issue 댓글 id로 메웠다 — §5.4).
- L218 "max_attempts: 3은 예시 … 실측 최적값 아님"은 좋으나, L185 "retry budget과 deadline을 적용"에 budget이 없을 때의 동작이 없다.
- L199 "Kiro correctness", L201 "TDFlow 논문", L278 "NIST SSDF·OWASP Agentic Top 10", L398 "Anthropic eval 가이드" — 링크 제목만 있고 어떤 주장을 뒷받침하는지 문장별 연결이 없다.
- L121 "성공률 분포, 권한 위반, 비용·지연" eval 지표의 측정 방식·반복 횟수 기준 없음(RUN-04가 "예시의 전략과 evidence"만 요구).

### 5.3 상위 문서와의 불일치

- 상위 §5 L466 이하 "change_strategy·domain_modeling 독립"은 이 문서 L233·L404와 일치(선행 대조표 §5행).
- 상위 §6 L758 "단일 조정 owner의 작은 journal"과 이 문서 L167 "한 task를 시작한 native session/host의 단일 dispatcher": 조정 주체가 리더(상위)인지 세션(하위)인지 다르다.
- 상위 §15 L1576 이하의 "PASS-WITH-DEBT 비평탄화"가 이 문서 §5 테스트 무결성에 없다(waiver는 [e08] §8에만).

### 5.4 RC1 결정·실물과의 불일치

- 전략표(L41–L139) ↔ `src/run.rs:23-30` STRATEGIES: tdd(red/green/regression)·bugfix(repro-fail/fixed/regression)·preserve(baseline/after/intent-diff)·contract(contract-tests/migration/rollback)·eval(trials)·docs(schema/links/smoke). 6행 1:1 일치.
- `method.change_strategy`/`domain_modeling`(L233) ↔ `src/run.rs:35-40,165-167`; legacy `tdd|ddd` 매핑(L404) ↔ `legacy_mode`("ddd"→preserve+required). 일치.
- evidence 필드 spec_digest·base_revision·tree_digest·policy_digest(L243–L246) ↔ `src/run.rs:130-134`. producer runner-receipt/self-report(L248, L261) ↔ `src/run.rs:7`, `src/audit.rs:5-7`. 일치.
- stale 판정(L261, L340) ↔ `run::stale_reasons`(tree/spec/policy digest 비교, `src/run.rs:253-267`). 일치.
- RED 환경 오류 분리(L179, L310) ↔ `verify::FailureClass{EnvironmentFailure,TestFailure,Unclassified}`(`src/verify.rs:18-25`), RED 확인은 tdd/bugfix만(`src/run.rs:217-222`). 일치.
- budget(L236): RC1 `Budget.max_attempts` 필수·기본값 없음(P11, `src/run.rs:291-293`); 시도 이력은 호출자 보관, receipt 미저장(P09, `:298`). 문서 L185의 deadline은 RC1에 없다(`Budget` 필드 1개).
- 명령 출처 "repo에 정의된 검증 명령"(L177): RC1은 승인 snapshot의 `verification_commands`만, argv 분리·shell 없음·env 접두 거부(decisions:72; `src/verify.rs:3`). 문서보다 RC1이 좁다.
- 단일 소유자(L167–L169): RC1은 ① 같은 clone 내 journal+advisory lock(`src/claim.rs:5-13`) + ⑤ Issue 댓글 id fencing(최소 live 댓글 id 승, `:15-23`; decisions:21). 문서는 댓글 claim을 모른다. "assignee·label을 lock으로 쓰지 않음"(L167)은 RC1과 일치.
- handoff "이전 실행 중지·상태 조회·명시 handoff"(L167): 06-B `claim-handoff`는 리더 명시 요청이 승인이며 timeout으로 사망 판정하지 않음(`src/handoff.rs:4-7`). 일치.
- 내부 명령(L267–L272): RC1 `run-verify`(op `run.verify`, `src/main.rs:407`). `task submit/status/cancel`·`spec verify`·`artifacts list` 없음(열거 대조). `--execution cloud`(L268) 대응 없음 — remote/CI 위임은 RC1 미구현.
- 상태 Ready→Running→Review→Verified(L165): RC1 `RunOutcome` 상태는 `ready|stale|blocked`(`src/run.rs:149`), claim 상태는 `assigned…in_review`(`src/claim.rs:50-58`). Review/Verified 상태 없음.
- RUN-02 RED 로그(L442): `tests/run_verify.rs`·`tests/audit_admission.rs` 존재하나 RED 전용 fixture 여부는 본문 미확인. RUN-03 diff guard(L455): 05-D `reduce`가 "Passing while the oracle itself changed needs a reader"(`src/run.rs:338`)로 처리 — 일치. RUN-07 remote 미설치(L507): `requires_binary: false`(RUN route). 일치.

### 5.5 낡은 내용

- L140 "기존 ANALYZE–PRESERVE–IMPROVE 유지"·L141 "기존 MoAI 모듈" 링크는 Go 세대 skill(`moai-workflow-ddd`). RC1 이름은 `preserve`.
- L274 "/moai 구현해 줘 또는 $moai": `$moai` Codex 호출은 not_run(decisions:66).

### 5.6 개선 제안(영향 순)

1. §3에 RC1 소유권 실물(① journal+lock, ⑤ 댓글 id fencing, 만료 자동 takeover 금지)을 한 문단으로 넣는다.
2. L165 상태를 `RunOutcome`(ready/stale/blocked)과 claim 상태로 분리해 적는다.
3. §8 명령을 `run-verify`로 바꾸고 remote/CI 위임은 "RC1 미구현"으로 표시한다.
4. budget에 deadline이 없다는 RC1 차이를 L236 예시에 반영한다.
5. 외부 링크 7건을 문장별 각주로 바꾼다.

### 5.7 미확인·한계

- `tests/run_verify.rs`·`tests/audit_admission.rs`·`tests/claim_record_cli.rs` 본문은 읽지 않았다.
- TDFlow·Kiro·SSDF 등 외부 문서는 재조회하지 않았다.

---

## 6. [487] MoAI 파일 소유권과 worktree 수명 (`page_487fa473…`, 794행)

### 6.1 문서 개요

Git 작은 계약·프로젝트 공용 불변 데이터·worktree별 변경 상태의 3분리와 불변조건 5개(§1), 데이터 authority 표(§2), 홈·프로젝트 전체 배치(§3), identity·cache key·secret 규칙(§4), resolver/publish/GC 인터페이스와 object 상태(§5), 권한·GC·offline·backup(§6), Go 기존 구현 차이(§7), ST-01~09(§8), 뒤에 선택적 TODO 파일과 홈 hook 소유권이 붙는다. identity를 `git rev-parse` 출력으로만 정하는 원칙(§4)은 RC1 01-A와 같다.

### 6.2 내부 결함

- L156 "기본 배치는 ~/moai/projects/<project>·~/moai/worktrees/<project>/<task>"와 같은 문단 "~/moai/projects·~/moai/worktrees … 바뀌지 않음": 존재를 전제한 경로인데 출처가 없다(§6.4·§6.5).
- L359 "Windows·archive·sandbox에서 검증한 경우에만 링크를 선택" ↔ L292 "외부 home symlink에 의존하지 않음" ↔ L571 "hardlink나 .moai 전체 공유 금지": 같은 문서 안에서 링크 허용 조건이 세 가지로 갈린다.
- L540–L561 object 상태 9개(Staged…Deleted)와 L573 worktree 상태 5개(registered…cleanup-eligible)가 §8 시험의 통과 조건에 쓰이지 않는다.
- ST-01 "1/4/8 worktree와 동일 artifact"(L614): artifact 크기·측정 지표(L760 logical/allocated bytes)와 통과 임계값이 없다.
- L164 `config.toml`(개인 profile·workspace_root·quota)과 L219 `project.yaml`(profile)의 역할 분담이 없다.
- L444 "Remote object key … projects/<project-id>/artifacts/sha256/…" — remote(어느 저장소?)가 §2 표의 "선택 기존 object store"뿐이어서 key 형식의 적용 대상이 없다.

### 6.3 상위 문서와의 불일치

- 홈 배치: 상위 §13 L1365–L1383 ↔ 하위 L163–L174. `installs/`↔(없음), `projects/<id>/{objects,brain,specs,workspaces}`↔`projects/<id>/{registry.json,contracts/<sha256>/,artifacts/,indexes/,runs/,workspace-state/,retention/}`. 이름이 거의 전부 다르다(선행 A2).
- 상위 §4 L462 "기존 소유권 문서에 남은 선택 symlink 표현은 계승하지 않음"은 이 문서 L359를 가리킨다. 상위가 하위를 부정했는데 하위는 고쳐지지 않았다.
- 상위 §12 L1318 "자동 archive 정책·기간 미결"과 하위 L575 GC(mark→tombstone→grace→sweep)·L173 `retention/`: 하위가 정책 없이 메커니즘을 확정했다(선행 B4와 같은 성격).
- 상위 §13 L1385 "spec-index.json·profile·snapshot 중 무엇을 Git에 둘지 프로젝트별" ↔ 하위 L232 `contracts.lock` "tracked" 단정.

### 6.4 RC1 결정·실물과의 불일치

- worktree 위치 `~/moai/worktrees/<project>/<task>`(L156, L161): RC1 결정은 `<primary>/.moai-cli/worktrees/<issue>-<key>`(decisions:34; `src/worktree.rs:11-12`). Go 기준선은 `~/.moai/worktrees/`(`go@4fa2aa9 internal/cli/cc.go:304`). `~/moai/`(점 없음)는 Go `internal/` 검색 0건(가설) — 적극 근거는 RC1과 Go가 모두 다른 경로를 이름 붙였다는 것이다.
- worktree 수명 5단계(L573): RC1은 `git worktree add --lock --reason "moai claim …"`만 수행하고 삭제·prune 없음(`src/worktree.rs:7-9`); 레인은 ExitWorktree(keep)·삭제 금지(decisions:90). retire/cleanup 단계는 RC1에 없다.
- identity `git rev-parse --path-format=absolute --git-common-dir`(L367): 구현 일치(`src/identity.rs:188-189`). `--absolute-git-dir`·`worktree list --porcelain -z`는 01-A가 쓰지 않고 06-C가 `worktree list`를 쓴다(미확인 — 가설).
- journal 위치 `runs/<run-id>/`(L171): RC1 claim journal은 `<common dir>/.moai-cli/claims/`(`src/claim.rs:5-9`; decisions:86 ③). receipt는 `<project>/.moai-cli/receipts/`(`src/apply.rs:30`; decisions:78). outbox `<project>/.moai-cli/forge/outbox/`(선행 충돌 #2).
- `config.toml`·`registry.json`·`contracts/<sha256>/`·`artifacts/`·`indexes/`·`workspace-state/`·`retention/`·`cache/`(L164–L174): `src/` 검색 0건(가설). 홈에는 `installations/`만 있다.
- `contracts.lock` tracked(L232): 없음(§4.4).
- `.moai-cli/manifests/` "공통 manifest tracked, local apply receipt ignored"(L279): 일치 — `manifests/generation.json`·`sources.json`·`profile-baseline.json`·`claude-settings.json`·`github.json`(`src/generation.rs:50-51`, `src/profile.rs:8`, `src/claude.rs:320`, `src/github_templates.rs`), receipt는 `.gitignore` 관리 구간으로 제외(`src/apply.rs:22-25`).
- hardlink 금지(L571) ↔ RC1 `publish_exclusive`가 hard link로 배타 생성(`src/apply.rs:6-7`; decisions:78 "hard_link는 사용자 승인 아님, P12 HOLD"). 문서의 금지 대상(writable code/SPEC/state 공유)과 RC1 용도(같은 디렉터리 안 원자 생성)가 다르지만 문서는 구분하지 않는다.
- 3-way 보존·expected hash 중단(L587): 구현 일치(`src/apply.rs:1-9`, `src/profile.rs:5-12`).
- ST-02 bare·separate-git-dir(L631): 01-A가 common dir/git dir로 구분(`src/identity.rs:3-5`), GIT_* 환경변수 제거(`:25-37`, #2033). 일치. ST-03 publish crash(L648): `tests/apply_recover_partial.rs` 존재. ST-06 `git clean -x` 보존(L699): receipt가 worktree 삭제와 함께 사라지는 것은 RC1이 스스로 기록한 gap(`src/apply.rs:25-28` `@MX:CEILING`)이라 ST-06 조건을 충족할 수 없다(시험은 미실행).
- Go 기존 구현(L585 paths.go·worktreeroot.go·state_guard.go): 세 파일 모두 `go@4fa2aa9`에 존재. 일치.
- 선택 TODO `projects/{project_id}/tasks.md`(L780): RC1은 경로 인자(§1.4). "StorageResolver가 identity로 찾는다"는 미구현.

### 6.5 낡은 내용

- L156·L159–L161 `~/moai/` 배치 전체.
- L585 Go `.moai/state·.moai/reports/worktree-guard` 서술은 기준선 관찰이며 RC1은 Go 경로에 쓰지 않는다(decisions:89).
- L359 조건부 링크 허용 — 상위 §4 L462와 RC1 `update.rs:22`가 폐기.

### 6.6 개선 제안(영향 순)

1. L156–L174를 "RC1 실물(installations/, .moai-cli/{manifests,receipts,forge,worktrees,claims@common-dir}) / 후속 목표(objects·runs·retention…)"로 나눈다.
2. L359 조건부 링크 문장을 삭제하고 상위 §4 L462를 인용한다.
3. worktree 5단계를 "RC1: add --lock, keep, 삭제 금지 / 후속: retire·cleanup"으로 고친다.
4. ST-06에 "RC1 알려진 gap(receipt 소실)"을 적는다.
5. `~/moai/`를 `~/.moai/`(Go)로 정정하거나 "새 제안 기본값"으로 표시한다.
6. `config.toml`과 `project.yaml`의 역할을 한 표로 정리한다.

### 6.7 미확인·한계

- 06-C의 `git worktree list` 사용 여부는 소스 본문을 읽지 않아 가설이다.
- XDG·SQLite·S3 conditional write 링크는 재조회하지 않았다.

---

## 7. [66a] MoAI RC1 GitHub 작업 구조와 실행 계획 (`page_66acadc2…`, 1,693행)

### 7.1 문서 개요

RC1을 milestone → 영역 상위 Issue(EP-01~09) → 한 세션 실행 Issue(00-A~08-D, 32건)로 나누는 발행 계획이다. 기존 라벨·양식 기준(§2), GitHub 객체별 책임(§3), 구조(§4), 32개 Issue와 선행(§5), 공통 포맷(§6), 3종 intake 템플릿(§7), 한 Issue 한 세션 claim(§8), `.github` 개발본→Rust 템플릿(§9), A–G 단계(§10), GH-* 수용 기준(§11), 공식 출처(§12). RC1은 실제로 이 32개 키로 Issue를 발행·완료했으므로 7개 문서 가운데 가장 많이 "실행된" 문서다.

### 7.2 내부 결함

- L11·L107–L111 "아직 실행하지 않음"·L1353 "먼저 전체 계획을 사용자 채팅으로 보고하고 … Issue를 만들거나 배차하지 않는다": 문서 상태 배너가 발행 전 시점에 고정돼 있어 현재 독자는 어디까지 실행됐는지 알 수 없다.
- L227 "라벨 수가 25라고 적힌 이전 요약과 실제 26개": 어느 요약인지 미표기. 26은 검증됨(§7.4).
- L408·L604 "190개 Claude 비교": 집계 기준 없음. Go 기준선 `.claude/agents/` 22파일, `.claude/` 전체 600파일 — 190과 맞는 단순 집계가 없다.
- L1256 상태 11개와 L1272 claim record 필드, L1146 YAML 메타데이터(claim.state/session_id/epoch)가 세 벌이며 서로 어떤 것이 projection인지 L1166에서만 설명한다.
- L375 "sub-issue 부모당 100개, 깊이 8", L377 "관계 종류별 50개": 출처 절(L1677–L1683)에 제목만 있어 문장 단위 근거가 없다.
- L1337 "Issue forms는 공식 문서상 public preview": 확인일(L1669)에 대한 공식 문서 인용 없음.
- L556 "F20·F22–F25 이름이 다르다"고 경고하면서 어느 문서의 어느 이름인지 적지 않았다(실제 차이는 §1.2 참조).

### 7.3 상위 문서와의 불일치

- 상위 §18 L1921 이하 7단계 구현 순서(Core Brain MVP가 Issues 협업 앞) ↔ 이 문서 §5 순서(00→08, Core Brain 없음). 이 문서가 RC1 실물과 일치하고 상위가 어긋난다(선행 A3).
- 상위 §6 L758 "단일 조정 owner의 작은 journal" ↔ 이 문서 L1270 같은 입장. 둘 다 RC1 ⑤ 댓글 claim을 모른다(선행 A4).
- 상위 §5 통합 행(선행 A5) ↔ 이 문서 L1264 "통합 owner … PR 개설이 merge 허가를 만들지 않음": 메커니즘(ruleset·`moai/audit`) 둘 다 없음.
- 상위 §13 L1362 `.github/ GitHub 선택 프로젝트만` ↔ 이 문서 L1347 같은 원칙. 일치.

### 7.4 RC1 결정·실물과의 불일치

- 라벨 26개(L156): `go@4fa2aa9 .github/labels.yml` name 항목 26개. 일치. 다만 RC1 00-B가 읽은 원격 라벨은 94개(`docs/github-inventory-rc1.md:30`) — 문서 L227의 "원격 현황을 읽기 전 단정 않음"이 옳았고, 결과가 문서에 없다.
- 32개 Issue(L577): RC1은 00-A~08-D 32건에 00-E(#2020)·00-F(#2028)·04-D(#2016, RC1 밖으로 이동)·결함 카드(#2033·#2043·#2048·#2054·#2082·#2104) 등을 더했다. 문서는 추가분을 모른다.
- 양식 파일명 feature.yml·bug.yml·research-design.yml·config.yml(L1174, L1311–L1314): RC1 실물은 `moai-feature.yml`·`moai-bugfix.yml`·`moai-investigation.yml`, `config.yml`은 의도적으로 템플릿 아님(`src/github_templates.rs:13`; `.github/ISSUE_TEMPLATE/` 3파일). 라벨은 `.github/moai/labels.yml`, workflow는 `rc-label-sync.yml`·`rc-ci.yml`, `validate-issue-config.yml` 없음(`git ls-tree` 열거).
- `<Rust-source>/templates/github/{ISSUE_TEMPLATE,PULL_REQUEST_TEMPLATE.md,labels.yml,manifest.json}`(L1321–L1325): RC1 `templates/github/` 존재, `include_str!`로 바이너리에 포함(`src/github_templates.rs:3-10,40-59`). PR 템플릿 파일명은 `pull_request_template.md`(소문자). 일치(이름만 다름).
- fixture `clean-project/existing-user-templates/conflicting-labels/interrupted-update`(L1327–L1331): RC1 `tests/fixtures/github/`는 `labels.json`·`issues-*.json`·`sub-issues.json`·`dependencies.json` 등 수집 fixture. 네 디렉터리 없음(열거 대조).
- claim 상태(L1256): RC1 `ACTIVE_STATES` = assigned·running·verifying·pr_ready·in_review·blocked·handoff_pending·uncertain(`src/claim.rs:50-58`), 종결 outcome claimed/unclaimed/released/expired/cancelled(`:751-988`). 문서 어휘와 거의 같다. `integrated`만 RC1 상태가 아니다(통합은 리더 게이트 밖).
- "journal이 authority, GitHub 댓글은 projection"(L1270–L1272) ↔ RC1 "⑤ 댓글 claim이 primary, ① journal은 같은 host 직렬화"(`src/claim.rs:1-3`; decisions:21). L1270이 "provider 조정 기능이 원자성을 실제 보장하면 adapter로 쓸 수 있다"고 열어 둔 대체 경로가 RC1에서 primary가 됐는데, 문서는 journal을 authority로 둔 채 갱신되지 않았다. **다음 구현자가 이 문장대로 journal을 primary로 만들면 RC1 06-A와 반대가 된다.**
- "heartbeat는 조정 record에, Issue에는 고빈도로 쓰지 않음"(L1274): RC1 lease 수치 없음·heartbeat 없음(decisions:86 ③; `src/handoff.rs:4-7`). 더 보수적.
- "무응답 자동 takeover를 새로 승인하지 않음"(L1292): RC1 일치(`src/claim.rs:22-23` REQ-06A-16).
- branch에 logical key·Issue URL·SPEC digest·SHA 연결(L1298): RC1은 `Moai-Claim: <issue>#<comment-id>` 토큰을 branch suffix `c<issue>-<id>`·commit trailer·PR 본문에 전파(`src/claim.rs:19-21`). 문서에 claim 토큰 개념 없음. branch prefix는 HOLD(`src/worktree.rs:13-14`).
- `Refs #` 기본, `Closes #` 조건부(L1300): `AGENTS.md` 워크플로 6단계와 일치.
- "동일 세션에 두 Issue 활성 claim 금지"(L1300): 일치. 단 RC1 lane loop는 PR 뒤 병합을 기다리지 않고 다음 카드로 이동(decisions:77) — 문서 L1300 "리더가 재배정"과 양립.
- 통합 "최신 base·CI·review·권한 확인 후"(L1264, L1441): RC1은 ruleset `RC1 integration gate`(PR 필수, `moai/audit` status, strict up-to-date) + verdict 첫 줄 형식 + `gh pr merge --auto`(decisions:39, 43). 문서에 없음(선행 A5).
- 메타데이터 YAML(L1146–L1164): RC1 `github-issues.json`은 issue당 `id·node_id·number·title·url·created_by·reused`와 상위 `dependencies·links·milestone`(파일 키 열거). `spec_ref`·`claim`·`permission_scope`·`evidence_status`는 Issue 본문 "사용자 결정" 절과 decisions에만 있다. L1166 "Issue object ID·number/URL 기록"은 일치.
- 리더 행동 L1258 "실행 가능한 Issue를 배정" ↔ RC1 리더 레인 상시 점검 루틴(decisions:87)·레인 보고 중계(decisions:33). 문서가 더 단순.
- Dependencies·sub-issue API(L375–L379): RC1 00-B fixture `tests/fixtures/github/{sub-issues,dependencies}.json` 존재. 한도 수치는 미검증.
- 조사설계 양식 라벨 "type:research 임의 확정 안 함"(L1247): RC1 실물 `moai-investigation.yml`의 기본 라벨은 확인하지 않았다(미확인).

### 7.5 낡은 내용

- L11·L107–L111·L1353·L1663 "아직 실행하지 않음/미결" 전부 — 2026-11-08 15:5x~2026-11-09 00:2x 사이에 발행·배차·통합·결정이 진행됨(decisions 전체).
- L1663 "signing·release 운영 미결" → decisions:8(attestation 단독). tag ruleset·attestation workflow만 HOLD(`docs/release-checklist.md:62-63`).
- L1292 "P08 자동 인계 세부 미결" → decisions:21 방식 결정(⑤+①), 수치만 HOLD.
- L1139–L1168 공통 Issue 포맷: RC1 본문은 `.moai-cli/reports/issue-bodies/<n>-<key>.md` 사본 + "사용자 결정" 절로 운영(decisions:3, 56). 문서의 YAML 블록 형식은 채택되지 않았다.

### 7.6 개선 제안(영향 순)

1. §8 원자성 경계(L1270–L1274)를 RC1 ⑤ 댓글 id fencing·`Moai-Claim` 토큰·① 보조 journal로 다시 쓴다 — 권위 방향이 반대인 채로 두면 다음 구현자가 journal을 primary로 만든다.
2. 문서 머리에 "발행 완료(2026-11-08), 상태는 Issue 원장 참조" 배너와 decisions 링크를 둔다.
3. §9 파일명·workflow 이름을 실물(`moai-*.yml`, `.github/moai/labels.yml`, `rc-label-sync.yml`)로 바꾸고 fixture 네 디렉터리를 삭제한다.
4. L1256 상태표에 `integrated`를 "RC1 상태 아님(리더 게이트)"으로 표시하고 ruleset·`moai/audit`을 통합 행에 적는다.
5. L408 "190개"의 집계 기준을 적거나 삭제한다.
6. L1663 미결 목록을 decisions:94–96의 HOLD 목록으로 교체한다.

### 7.7 미확인·한계

- 원격 GitHub 상태(라벨 94개 이후 변동, Issue 수)는 재조회하지 않았다. `github-inventory-rc1.md`는 2026-11-08T07:56Z 기준이다.
- GitHub 공식 한도(100/8/50)와 Issue forms 상태는 확인하지 않았다.
- `.github/ISSUE_TEMPLATE/moai-*.yml` 본문(기본 라벨)은 읽지 않았다.

---

## 8. 일곱 문서 사이의 모순

| # | 주제 | 문서별 서술(행) | RC1 실물 | 판정 |
|---|---|---|---|---|
| X1 | 홈 설치 경로 | [83f] L37 `<state_root>/skills/moai/<version>/`·L139 `installations/`; [487] L165 `skills/moai/<version>/`·L166 `projects/<id>/`; [25f] L743 "별도 manifest·journal·lock"(경로 없음); 상위 §13 L1367 `installs/<host>/<version>/` | `~/.moai-cli/installations/<install_id>.json`, `…/<id>/hosts/<host>.json`, `…/known-good/<hex>`(`src/install.rs:4`, `src/update.rs:18,335`); 홈 패키지 루트 상수 없음 | 세 문서 + 상위가 네 가지 이름. [83f]의 `installations/`만 RC1과 같다 |
| X2 | symlink 대 복사 | [83f] L134·L238 "symlink 공유하지 않음"; [dff] L248 "외부 symlink 의존은 portable 실패"; [487] L359 "검증한 경우에만 링크 선택"·L571 hardlink 금지; 상위 §4 L462 조건부 symlink 표현 폐기 | symlink projection = blocked(`src/update.rs:22`); 배타 생성에 hard link 사용(`src/apply.rs:6-7`, 사용자 승인 아님 decisions:78) | [487]만 링크를 열어 둠. hardlink는 [487]의 금지 범위와 RC1 용도가 달라 설명이 필요 |
| X3 | worktree 위치·수명 | [487] L156 `~/moai/worktrees/<project>/<task>`, L573 5단계(…retired→cleanup-eligible); [83f] L18 `~/moai/worktrees` 유지; [66a] L1298 "다른 worktree" (위치 없음); [3d9] L175 "reset·clean 금지" | `<primary>/.moai-cli/worktrees/<issue>-<key>`(decisions:34; `src/worktree.rs:12`), `add --lock`만, ExitWorktree(keep)·삭제 금지(decisions:90); Go는 `~/.moai/worktrees/` | `~/moai/`는 어느 실물에도 없음. 수명 단계는 [487]만 정의, RC1은 retire/cleanup 없음 |
| X4 | 릴리스 tag·서명·packaging | [25f] L197 tag 후보 `v2026.11.0-rc.1`·L271 cargo-dist 후보·서명 미정; [66a] L1663 signing 미결; [83f] L1008 F23 "서명·버전"(방식 없음) | tag `moai-cli-v2026.11.0-rc.1`(decisions:9), attestation 단독(decisions:8), cargo-dist 결정 A(decisions:7)이나 미도입(`release-checklist.md:29`), self-replace 결정(decisions:12)이나 `Cargo.toml` 의존성에 없음 | 세 문서 모두 결정 이전 상태. 결정됨·미구현 항목 3건(cargo-dist·self-replace·attestation)을 어느 문서도 구분하지 않음 |
| X5 | 상태 어휘 | [83f] L1223 8단계 discovered…host_verified; [25f] L161–L170 Inspect…Blocked 10개; [e08] L58–L67 Draft…Cancelled 10개; [3d9] L165 Ready→Running→Review→Verified; [66a] L1256 unclaimed…integrated 11개 | receipt `pending|committed|recovered`(`src/update.rs:251`); admission `ready|stale|blocked`(`src/plan.rs:4`); claim `ACTIVE_STATES` 8 + outcome 5(`src/claim.rs:50-58`) | 다섯 벌의 어휘 중 [66a]만 RC1 claim 상태와 대응. 나머지는 서로도 RC1과도 대응표가 없다 |
| X6 | journal·receipt·outbox 위치 | [487] L166–L174 홈 `projects/<id>/runs/`; [83f] L190 홈 `runs/<run_id>/`; [25f] L116 "journal"(위치 없음); 상위 §13 L1378–L1382 홈 `workspaces/<id>/{journal.jsonl,outbox}` | claim journal `<common dir>/.moai-cli/claims/`(`src/claim.rs:5-9`), receipt `<project>/.moai-cli/receipts/`(`src/apply.rs:30`), outbox `<project>/.moai-cli/forge/outbox/` | 문서는 모두 홈, RC1은 모두 프로젝트/clone 쪽(선행 충돌 #2). `src/apply.rs:29`가 교체 지점을 `@MX:UPGRADE`로 표시 |
| X7 | 실행 소유권의 권위 | [66a] L1270–L1272 journal authority·댓글 projection(단, "provider primitive를 adapter로" 허용); [3d9] L167 세션 단일 dispatcher·journal+OS lock; [e08] L99 "현재 owner 확인"(메커니즘 없음); 상위 §6 L758 리더 journal | ⑤ Issue 댓글 claim primary(최소 live 댓글 id 승) + ① 같은 host journal(`src/claim.rs:1-23`; decisions:21) | [66a]가 열어 둔 대체 경로가 RC1 primary가 됐으나 [66a]·상위 본문은 journal authority 그대로; [3d9]는 cross-host 수단이 비어 있음 |
| X8 | MCP 범위 | [dff] L477 "선택적 1급", HA-09; [25f] L120 "optional first-class", RL-08; [83f] L1165 "기존 38개 tool"; [e08] L164·[3d9] L276 "기존 MCP 소비자 compatibility" | RC1은 CLI만, MCP 다음 버전(decisions:51; AC-01C-2/3/4 not_applicable) | 다섯 문서가 MCP를 RC1 시험 대상처럼 서술. "범위 밖" 표시는 어디에도 없다 |
| X9 | 내부 CLI 명령 체계 | [25f] L114 `setup/project init/maintenance update/doctor/capabilities`; [e08] L157–L160 `skill show/spec plan/task status`; [3d9] L267–L272 `task submit/cancel, spec verify, artifacts list` | `project-plan/apply`, `update-plan/apply`, `recovery-plan/apply`, `generation-plan/apply`, `plan-admit`, `run-verify`, `claude-doctor`, `codex-doctor`, `install-inspect` …(`src/main.rs:99-526`) | 세 문서가 세 가지 체계, RC1은 네 번째. 공통 명령은 `capabilities`뿐인데 RC1에 없다 |
| X10 | Codex 대상 표면 | [83f] L10·[dff] L18·[25f] L8 "Codex Desktop"; [dff] L185·[83f] L105 `~/.agents/skills`(CLI/홈); [66a] L753 "Codex Desktop 발견 호출" | 실측 스크립트는 CLI(`tests/e2e/host-codex.sh`), Desktop·CLI 모두 not_run(decisions:66; `support-matrix.md:28`) | Desktop을 RC gate로 쓰면서 Desktop 전용 시험이 없다 |
| X11 | 세 문서에 붙여 넣은 같은 문단 | [83f] L18 = [25f] L737 = [487] L156(홈·프로젝트 root 분리); [83f] L2597 = [dff] L827 = [487] L792(홈 hook 소유권); [83f] L2611 = [25f] L739(Go 전환) | — | 한 곳을 고치면 세 곳을 같이 고쳐야 한다. 이미 [487] L156만 `~/moai/`·`config.toml` 문장을 추가해 어긋났다 |
| X12 | Go 전환(cutover) | [83f] L2611 "테스트 완료 후 Go 제거·moai 전환 계획을 따른다"; [25f] L739·L749·L751 RL-COMMAND-CUTOVER; [66a] L1552 "old Go Issue·tag 일괄 처리 금지" | 04-D RC1 밖, Go 상태 무변경(decisions:91 ②, 89) | 두 문서가 RC1 안의 계획처럼 서술, [66a]만 보존 쪽 |
| X13 | F 기능 ID 이름 | [83f] L941–L1046 F20 API 예산·F22 저장·F23 릴리스·F24 Desktop·F25 호환; 상위 §20 L2407–L2434 F20 판단 backend·F22 ADE·F23 Desktop·F24 유지보수·F25 보관; [66a] L556 "이름이 다르다" 경고만 | — | 같은 번호가 다른 기능을 가리킨다. 번호로 누락 검사를 하면 틀린다 |
| X14 | 수용 시험 ID 체계 | [83f] HOME-/SETUP-/GENERATION-; [dff] HA-; [25f] RL-/DT-; [e08] PLAN-; [3d9] RUN-; [487] ST-; [66a] GH-; 상위 §17 15개 시험군 | RC1 receipt는 AC-08C-n·HA-02(`docs/e2e-receipts/receipts.jsonl`, `support-matrix.md:15-20`) | HA-02 하나만 문서 ID가 RC1 receipt에 쓰였다. 나머지 60여 개 ID는 대응 시험이 없거나 이름이 다르다 |

## 9. 종합 — 우선 고칠 것

1. **권위 방향 미갱신(X7)**: [66a] L1270의 "journal authority, 댓글 projection"은 RC1(⑤ 댓글 primary)과 반대로 읽힌다. 문서가 허용한 대체 경로가 실물이 됐는데 본문이 그대로라, 이 한 줄이 가장 위험하다.
2. **결정됨·미구현 3건(X4)**: cargo-dist, self-replace, attestation은 결정 기록에는 있고 `Cargo.toml`·체크리스트에는 없다. 어느 문서도 이 상태를 적지 않는다.
3. **홈 경로 네 가지 이름(X1)**과 **`~/moai/`(X3)**: 실물과 같은 것은 [83f]의 `installations/`뿐이다.
4. **MCP·hook을 RC1 시험처럼 서술(X8, [dff] §5·§9)**: RC1 결정(hook 0, MCP 다음 버전)이 반영되지 않아 HA-05·07·08·09가 시험 불가인데 그 표시가 없다.
5. **상태 어휘 다섯 벌(X5)**: 대응표 하나 없이 문서마다 다르다.
6. 잘된 점: [3d9]의 전략표·evidence·method 필드는 `src/run.rs`와 1:1이고, [e08]의 4축 검사·EARS import, [dff]의 Claude/Codex 로딩 조건·바이트 수치, [66a]의 claim 상태 어휘·라벨 26개는 실물로 확인된다. 이 문서들은 고칠 양이 적고 "RC1 상태 한 줄"만 보태면 된다.

## 10. 이 검토의 한계

- 기준 SHA `9bfd26d3b`에서 읽었고, 세션 중 `origin/v2026.11.0`이 `d253f4a81`(PR #2130)로 이동했다. `git diff --stat` 결과 그 사이 변경은 `src/claim.rs`(+45)·`src/handoff.rs`(+206/−47)·`src/worktree.rs`(+3)·`docs/rust-product-path.md`(1행)뿐이다. 이 문서가 인용한 `src/claim.rs`·`src/handoff.rs`·`src/worktree.rs`의 행 번호는 `9bfd26d3b` 기준이며 tip에서는 밀렸을 수 있다. 다른 인용 파일은 두 SHA에서 동일하다.
- `src/` 파일은 모듈 머리말·상수·grep 결과로 확인했고 함수 본문 전체·`tests/*.rs` 본문·`cargo test`는 보지 않았다.
- 외부 공식 문서(Claude·Codex·GitHub·Electron·GEARS)는 하나도 재조회하지 않았다. 문서가 외부 사실로 적은 항목은 "미확인"으로만 다뤘다.
- Go 기준선 수치(hook 82, agents 22, `.claude/` 600, labels 26, 바이트 65,627·22,505, Codex hook 이벤트 12)는 파일·바이트 단순 집계이며 문서가 쓴 집계 기준과 같다는 보장은 없다.
- [83f] 사본의 중복 구간은 눈으로 대조했고 diff 명령은 돌리지 않았다.
