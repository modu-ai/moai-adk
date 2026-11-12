# contract-83f162ef-2389 개정안 — MoAI 홈 설치와 프로젝트 생성 기능 계약 (2026-11-10 개정본)

- 대상: Space 페이지 `page_83f162ef14d08191917913738184f167`의 2026-11-10 개정본. 읽은 사본은 사용자가 제공한 Markdown 내보내기 1,659행, sha256 `c1ae18d8adb6bb44d09fd9a55109b783c1dd6abcbb3e2f0a83ce357516561cbc`, 2026-11-12T00:12Z 읽음(필요한 행만). 사본과 Space 페이지가 같은 판인지는 리더 분석 노트와 마찬가지로 unverified다. 아래 행 번호는 이 사본 기준이다.
- 기존 `group3-83f162ef.md`는 11-09 이전 판(2,621행) 대상이라 이 문서와 행 번호가 다르다. 두 문서는 대체 관계가 아니다.
- 추적 Issue: #2389. 대조표: Issue #2389 댓글 6810415885(25건, tip `cc2a864392`). 이 문서는 그 표에서 "개정안 작성"으로 분류한 14건과, 선행 Issue가 끝나 덧붙인 B12(명령 경로 부분, #2391 R3)를 다룬다. B12 행의 근거는 tip `50166e3acc` 기준이다.
- 고정 근거 tip: **`cc2a864392`**. 결정 기록 행 번호는 그 tip의 `docs/decisions-2026-11-08.md` 기준이다.
- Space 페이지에 직접 적용하는 것은 외부 게시라서 이 문서가 하지 않는다. 적용은 사용자가 확인한 뒤에 한다.

## 수정 항목

| # | 위치 | 현재(요지) | 교체 | 근거 |
| --- | --- | --- | --- | --- |
| A1 | L1317–1323 절 「로컬 TODO의 선택 승격」 | 연결 뒤 로컬 TODO를 골라 Issue로 승격 | 절을 지우고 한 단락으로 바꾼다: "작업 원장은 연결된 GitHub/GitLab Issue뿐이다. forge가 연결되지 않으면 원장 명령은 로컬 TODO를 만들거나 읽거나 옮기지 않고, 공용 forge 오류 분류를 그대로 돌려준다. 로컬 대체 저장소가 없으므로 나중에 연결해도 승격할 로컬 목록이 없다." | 결정 L179②(로컬 TODO 승격 취소), `docs/p07-decision.md:8`, #2384 종료 6781175232, 표 행 37 B(L102) |
| A1 | L1140 `tasks.md` 행 | 미연결 개인 프로젝트의 선택 지속 체크리스트 | 행 삭제 | 같음 |
| A1 | L1180 F07 「주요 경계」 | "연결 시 Issues, 미연결 세션 TODO, 지속 파일 선택" | "원장은 연결된 Issues만. 미연결이면 원장 작업은 forge 오류로 끝난다. host 세션의 TODO는 원장이 아니다" | 같음 |
| A1 | L1233 예시 envelope | `"task_backend": "host-session"` | 키 삭제. `"profile"`도 관리 키가 아니므로 `group3-83f162ef.md` 5번 규칙(관리 키 `src/profile.rs` PROFILE_KEYS로 교체 또는 예시 표시)을 따른다 | 같음, `src/profile.rs` |
| A1 | L1274 TASK-MINIMAL | 미연결 단일 세션에서 DB·서버·지속 파일 생성 0 | ID를 `LEDGER-UNCONNECTED`로 바꾸고 "forge 미연결에서 원장 연산이 로컬 TODO·DB·지속 파일을 만들지 않고 forge 오류 분류로 답한다" | 같음 |
| A2 | L1301 Statusline 단락 | "기존 Go renderer·화면 구성을 구현 근거로 계승하지 않는다" | "상태줄은 Go MoAI의 화면 구성을 계승한다. 전환할 때 사용자가 차이를 느끼지 않는 것이 목표이며, Rust `moai-cli statusline`이 Go와 같은 출력을 내는지 정답 출력 대조(`--check`)로 확인한다. 렌더 중 git 조회는 허용한다. 새 3줄 배치는 #2376에서 따로 다룬다. PATH의 Go `moai` 제거와 상태줄 명령 전환은 #2383이 정하며, 실제 사용자 설정을 바꾸는 단계는 실행 직전에 사용자 확인을 받는다." 뒤의 Claude/Codex 차이·사용자 UI 보존·ZCode 분리 문장은 그대로 둔다 | 결정 L179①·L186·L97, #2350 종료 6782681329(PR #2374·#2424) |
| A3 | L254 | `/moai --mode agent`, `/moai --mode leader` | `/moai --mode-agent`, `/moai --mode-leader` | 결정 L179③, `skills/moai/SKILL.md`, `skills/moai/workflows/mode-agent.md`·`mode-leader.md`, `docs/design-review/autonomy-modes-spec.md:3` |
| A3 | L631 | "`/moai agent`와 `/moai leader`는 top-level 운영 모드" | "`/moai --mode-agent`와 `/moai --mode-leader`는 top-level 운영 모드". 끝에 한 문장 추가: "런처의 `-a`·`-l`은 세션을 띄우는 런처 옵션이며 skill 모드 표기가 아니다" | 같음, `src/main.rs` `role_of`(`-l` → leader, `-a` → agent) |
| A6 | L1187 F14 제목 | "RUN과 TDD/DDD" | "RUN과 구현 전략" | 정본 L207(6전략, DDD/TDD 이분법 복구 안 함), `src/run.rs:24-29` |
| A6 | L529 `references/`·`resources/` 행 | "GEARS SPEC, TDD·DDD, …" | "GEARS SPEC, RUN 6전략(tdd·bugfix·preserve·contract·eval·docs)과 domain_modeling 축, …" | 같음. `src/run.rs:178`은 옛 입력 `ddd`를 `preserve`로 읽는다 |
| A7 | L1206 SETUP 3단계 | 외부 skill 후보 조사 → 공식 폴더에 설치 → 전체 내용 읽기 → 맞춤 구성 | "3. 초기 인터뷰와 저장소 사실을 바탕으로 프로젝트 맞춤 agent·workflow·설정을 MoAI 공통 skill `moai` 안에서 구성한다. 외부 skill을 조사·설치하거나 읽기 영수증으로 들이는 단계는 없다(외부 skill 설치 기능을 쓰지 않는다). 기존 파일은 보존한다. script 실행·보안 권한·secret·비용의 필수 승인은 별도다." | 결정 L178②·L100, 표 행 35 A, `docs/rust-product-path.md:192`(`skill-source-plan`·`skill-source-apply`·`skill-read-receipt` 제거, PR #2474) |
| A7 | L1175 F02 | 실행 로직 "선택한 외부 skill 출처·범위·본문 검토", 경계 "외부 skill 무차별 실행·비밀 공유 금지" | 실행 로직에서 해당 단계를 빼고, 경계를 "외부 skill 설치 없음·비밀 공유 금지"로 | 같음 |
| A8 | L906 | "protocol operation 73개가 연결돼 있으며" | "protocol operation 목록의 정본은 `docs/design-review/appendix/interfaces.md` § "RC1 OPERATIONS 전수"이며, `src/main.rs`의 시험이 코드와의 일치를 강제한다(tip `cc2a864392`에서 95개). 수는 병합마다 바뀌므로 본문에 고정하지 않는다. launcher와 `statusline`은 operation 밖이다." | `docs/design-review/appendix/interfaces.md`(목록 95행), `src/main.rs:2404-2498`(동기화 시험, `statusline` 제외) |
| A8 | L11 | "기존 OMP 코드·Issue를 이 문서 개정으로 삭제·변경하지 않는다" | "OMP 지원 코드는 사용자 결정에 따라 #2378로 제거됐다. 이 문서는 남은 OMP 관련 Issue를 바꾸지 않는다." | #2389 본문(OMP 제외, #2378 병합), tip `src/`에 `omp` 단어 0건(`grep -rniw omp src`) |
| A9 | L357 | "사용자는 … hook 48개 중 42개가 `moai`를 호출한다고 전달했다. 이 수치는 사용자 보고이며 …" | "#2383 관련 'hook 스크립트 42개가 PATH의 `moai`에 묶여 있다'는 수치는 리더가 이 기계에서 센 측정이다(결정 기록 11:50Z 행 '리더 실측'). 셈 방법과 재현 기록이 없어 가설로 다루며, 48이라는 분모는 tip 기록에서 출처를 찾지 못했다. 과거 source의 다른 event·wrapper 개수와 합산하지 않는다." 굵은 글씨 문장은 그대로 둔다 | 결정 L186. tip의 `docs`·`src`·`skills`·Issue 본문 사본에서 `48개 중` 0건 |
| B1 | L455 다음, 제거 표에 행 추가 | (없음) | "\| Go MoAI의 Claude plugin(`moai@moai-adk`): `moai:*` skill 이름 공간, plugin MCP 서버, plugin hook \| host의 `enabledPlugins` 항목(user·project scope)과 plugin cache \| 파일을 지우지 않고 host 명령으로 plugin을 해제한 뒤(`claude plugin uninstall <name> --scope <scope>`), `enabledPlugins`에서 빠졌는지 다시 읽어 확인한다. project scope 항목과 cache 폴더는 따로 확인할 대상으로 남긴다. 실행은 사용자 확인 직후에만 \|" | 결정 L178③(2026-11-10 13:15Z 이 기계에서 실행, user 설정에서 사라짐, project scope는 disabled 표시, cache 남음), `docs/2383-statusline-switch-design.md:27`. 명령의 공식 문서 재확인은 이 카드에서 not_run |
| B5 | L254 단락 뒤에 추가 | (claim 경쟁 승자 규칙 없음) | "claim 경쟁: 같은 Issue에 살아 있는 claim 댓글이 여럿이면 댓글 id가 가장 낮은 쪽이 이긴다. 게시한 쪽은 다시 읽어 승자를 확인하고, 진 쪽은 물러나 다른 Issue를 고른다. 승자가 결정적으로 정해지므로 양쪽이 서로 물러나는 교착이 없다. lease는 4시간이며 만료만으로 인수하지 않는다(인계 줄 또는 운영자 결정). 감사 선점도 같은 규칙이다: PR head마다 `moai-audit-claim` 줄 가운데 가장 낮은 id가 이긴다." | `src/claim_verify.rs:5`, `src/audit_claim.rs:6-18`, 결정 L99(감사도 claim 대상), `AGENTS.md` Workflow 1 |
| B7 | 홈 결과물·GLM 서술(L11 "GLM은 Claude host의 provider profile") 뒤에 추가 | GLM 키 위치 없음 | "GLM API 키의 정본 위치는 MoAI 키 보관함이다: `~/.moai-cli/vault/<folder>/<provider>.key`에 암호화 파일로 두고(세 OS 같은 형식, OS 보관함은 쓰지 않음), 사용할 때 복호화한다. `moai-cli key-status`·`key-set`·`key-remove`로 다룬다. 옛 `~/.moai/.env.glm`·`~/.moai-cli/.env.glm`은 가져오기를 위해 읽기만 하며 이름 변경·삭제는 Go 도구 제거(#2383) 뒤로 미룬다. 실제 사용자 키 이전은 실행 직전 사용자 확인." | 결정 L103(파일 암호화, 6777274025)·L96·L177, `src/key_vault.rs:1-25`·`:773`(LEGACY_GLM_FILES), #2393 종료 6794425742 |
| B10 | L1172 F 표 머리 | 열: ID와 기능·입력·실행 로직·출력과 다음 단계·주요 경계 | 마지막에 열 「출시 단계·Issue」를 더한다. 값은 `rc.1`·`rc.2`·`후속(HOLD)` 가운데 하나와 담당 Issue 번호다. 근거 없는 값은 쓰지 않고 `미정`으로 둔다. 이 카드는 열 규칙만 정하고 값은 채우지 않는다(리더 노트의 "F04·F13·F20·F21·F24 = EP-F 후순위/HOLD"도 tip 근거를 다시 대조한 뒤 채움) | 리더 분석 노트 B10, 대조표 6810415885 |
| B12 | L365 「새 hook 선검증」 | "sandbox home에서 Rust 절대 binary 경로·protocol·…을 확인한다" | "Rust 절대 binary 경로"를 "생성 hook의 명령 경로(현재 PATH의 bare 이름 `moai-cli`)"로 바꾼다. 그 단락 끝에 한 문장 추가: "버전 경로·고정 링크·bare 이름 세 안의 비교와 현재 선택의 이유·대가(같은 이름 아래 실행체가 바뀌어도 host 재검토가 없고, hook 프로세스의 PATH에 따라 다른 `moai-cli`가 실행될 수 있음)는 `docs/hooks-generation.md` §3a에 둔다. Codex hook hash가 정의만 덮는지 실행 파일도 덮는지는 미확인이다." L930("현재 command는 bare moai-cli를 PATH에서 찾는다")은 이미 tip과 같아 고치지 않는다. trust 미승인·hash 변경 시험(L791·L1000)은 이 행에서 다루지 않는다(아래 #2539) | `docs/hooks-generation.md:64-80`(§3a, #2391 R3, PR #2536), `src/hooks.rs:71-79`(`hook_command`·`HOOK_BINARY = "moai-cli"`), #2391 닫기 6824006920 |
| B13 | L1016 단락 뒤에 추가 | Stop 반복만 시험 목록에 있음 | "Stop hook 공존: host 입력의 `stop_hook_active`가 참이면(다른 Stop hook 때문에 이미 이어지는 턴) MoAI turn-validation은 막지 않고 언제나 허용한다. Go의 /goal 연계 기능은 옮기지 않는다. hook 자체 실패·시간 초과는 지금 진행(fail-open)이며, Go hook을 지우기 전에 '명령 직전 검사만 차단'으로 바꿀지 다시 묻는다." | 결정 L101, 표 행 23·24, `src/hook_engine.rs:696-726` |
| B15 | L1016 단락 뒤에 추가(B13 다음) | 시험의 시간 의존·부하 규칙 없음 | "시험은 시간·경합에 의존하지 않는다. 한 기계에서 전체 빌드·전체 시험은 한 번에 하나만 돌리고, 이미 돌고 있으면 정해진 간격으로 기다리며 그 대기를 기록한다. 빌드 폴더는 worktree마다 하나다. 이름 필터 시험은 target마다 따로 돌리며 `0 passed`는 not_run이다. 부하·디스크 수치 기준은 아직 HOLD다." | `skills/moai/workflows/verify.md:40`, 결정 L172, #2385(닫기 준비 6784745839) |
| B16 | L1002·L1016 끝에 각각 한 문장 | "Windows/WSL … 별도 시험한다" | "다만 rc 인증 범위는 macOS(arm64·x86_64)와 Linux CI이고, Windows는 빌드와 self-replace 구현까지만이다. Windows/WSL 항목은 지원표에서 not_run으로 적고 PASS로 세지 않는다." | 결정 L131(OS matrix), `docs/e2e-matrix.md:57` |

## 이 문서가 다루지 않는 것

- 다른 레인의 열린 Issue가 내용을 정하는 8건: A5·B8·B14(#2237), B3(#2388), B4(#2390), B9(#2380), B11과 B12의 trust 부분(#2539). 각 Issue가 끝나면 이 문서에 덧붙인다.
  - #2391은 '엔진 + R1·R3·R8'로 닫혔다(리더 6824006920). B12의 명령 경로 부분은 R3으로 정해져 위 표에 넣었다. B11(호출당 지연 목표, R5)과 B12의 trust 미승인·hash 변경 시험(R2)은 #2539(needs:decision, HOLD)로 옮겨졌다.
- HOLD: B2(Go 소유권 digest 목록의 출처 미정, 실제 HOME 이전은 사용자 gate), B6(런처가 띄운 세션의 `/moai` 진입 경로 — tip `src/`에 전달 경로가 없고 대응 카드도 없음). PATH의 Go `moai` 제거 실행과 Go skill 폴더 교체 실행도 실행 직전 사용자 확인 대상이다.
- A4는 정본 문장이 이미 `moai`라서 바꿀 곳이 없다.

## 비고

- 가설만 근거로 한 교체 0건(A9는 측정의 한계를 그대로 적는 교체다). 개인 절대 경로·비밀 0건.
