# [DR][08] 론처 현행 문서화(`moai-cli cc|glm|codex [-l|-a]`) + Codex `/goal` 동등 명령 확인

## 배경
종합 보고서 §10.1·§10.4: 론처는 tip에 구현돼 있으나(00-F #2028, #2053) 사용자 문서에 "현행"으로 적혀 있는지 미확인. `/goal` 자동 입력 가능 여부 중 Codex CLI 항목이 **미확인**이라 자율 모드 설계(DR-02 #2138)가 "첫 프롬프트 지침"으로 임시 처리 중이다. 두 작은 항목을 한 카드로.

## 범위
- A. `docs/user-entry-draft.md`(없으면 `docs/rust-product-path.md`의 launcher 절)에 현행 론처 표: `moai-cli cc|glm|codex`(역할 없음 = 평범한 host 세션), `-l|--leader`, `-a|--agent`, 공통 옵션(`--name --model --effort --no-advisor --permissions --dry-run`), codex argv(`-s workspace-write -a on-request [-m] [PROMPT]`), Codex는 cross-session 채널 없음. 근거는 `src/main.rs` `Cli`/`HostLaunch`/`LaunchOpts`·`src/session.rs` 머리말에서 행 인용. 각 호스트 `--dry-run` 실측 JSON 1건씩 PR 본문에 첨부(바이너리는 `cargo run --offline --locked --` 또는 `target/release/moai-cli`; 세션을 실제로 띄우지 않음)
- B. Codex CLI에 Claude Code `/goal`(턴 종료를 조건 충족까지 막는 Stop 조건)과 동등한 기능이 있는지 **공식 문서로** 확인: 결과를 `docs/host-support-matrix.md`에 `goal` 행(claude: 내장 `/goal`, 모델 입력 불가 / codex: 있음(URL, 확인일) 또는 없음/미확인)으로 추가. 문서를 못 찾으면 "미확인(검색한 URL·확인일 기록)"으로 적고 추측으로 채우지 않는다
- 코드 변경 없음

## 수용 기준
- AC-1: 론처 표의 모든 플래그가 `src/main.rs` 실물과 일치(행 인용), `--dry-run` 출력 3건(cc/glm/codex) 첨부, 실제 세션 기동 0
- AC-2: 지원표 `goal` 행에 Claude·Codex 두 칸, Codex 칸은 URL+확인일 또는 "미확인(검색 기록)"
- AC-3: 개인 절대 경로·비밀 0건, check set 실측 exit

## 사용자 결정
- 2026-11-09 00:4x: "moai-cli 빌드해서 cc/glm/codex 하도록" + `/goal` 자동 입력 여부 확인 뒤 지침 제공
