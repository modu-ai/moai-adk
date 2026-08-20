# 미검증 항목 (gaps)

phase 1 감사에서 **확인하지 못했거나** 착지 단계에서 별도 확정이 필요한 항목.

## G1. 착지 시점에만 확정 가능한 것

- **hugo.toml `params.version`** — 현재 `v3.1-rc.2` / `releaseDate 2026-08-13`.
  v3.1.1 스탬프는 릴리즈 프로세스 소관(버전 SSOT 규칙). 드래프트는
  버전 문자열을 하드코딩하지 않았다.
- **README Release 배지 (v3.1.0)** — 태그 시점 갱신, 릴리즈 소관.
- **"전체 36개 커맨드" (README.ko L672)** — v3.1.1에 graph/tokens/memory
  최상위 커맨드가 추가돼 숫자가 변했을 것. `moai --help` 전체 목록
  세기는 릴리즈 바이너리로 해야 정확 (worktree bin/moai는 있으나
  릴리즈 확정판이 아님 — 세지 않았다).

## G2. 확인했으나 드래프트에 반영하지 않은 것 (의도적)

- **kanban-mode.md L142 "보드 상태 저장소에 호출자가 없다"** —
  `LoadBoard` 프로덕션 호출자 grep 0건으로 여전히 사실 확인.
  웹 콘솔·todo·factory가 `internal/kanban`을 import하지만 board-state
  저장소가 아니라 record/backlog 타입을 쓴다. 서술 유지.
- **README "v3.1 새 기능" 절의 v3.1.0 톤("광복절에 맞춰")** — 결정
  지점 D1로 운영자에게 이관. 드래프트는 팩토리 H3 추가만.
- **docs-site 산문 이모지** — 원본 ko 22건 전부가 "이모지가 주제인
  문서화"(statusline 예시·faq 버전 표시) 관행. 드래프트의 글리프
  설명(📡 📫 🏷️ 등)도 같은 부류로 유지. verify §5 스캔 결과와 판정
  근거 함께 기록.

## G3. 읽지 못한 페이지 (미감사 — 다음 라운드 후보)

- `cli-reference/ast-grep.md` — 배포 기본 룰 6종 승격(77922a36b)과
  gate.yaml SSOT(af5c7160d) 반영 여부 미확인. F16 커버리지 △/미확인.
- `advanced/hooks-reference.md` — branch-guard 조회 허용(2df70172f),
  거부 사유 정정(ee2a3dfb2), kanban 알림 백로그 요약(194eed3a0),
  hook wrapper 쌍 동기화(b06cec203) 반영 여부 미확인.
- `worktree/faq.md`·`worktree/guide.md` — worktree done 잠금 안내
  (901e5244f)·앵커 가드(af749fafe) 반영 여부 미확인.
- `claude-code/agentic/worktrees.md` — WorktreeCreate active-creator
  (fa667b067, F18) 반영 여부 미확인. grep 'WorktreeCreate'가
  hooks-reference 등 5파일에 걸렸으나 본문 수준 검토는 안 함.
- `multi-llm/model-policy.md` — GLM per-agent→session inherit
  (62c2bf939, F22) 반영 여부 미확인.
- `utility-commands/moai-loop.md` — `/loop` 포어맨 문단 유무 재확인
  안 함(grep foreman 0건 확인은 했음 — 착지 필요).

## G4. 드래프트 수치의 근거와 한계

- **statusline 예시 값**(`cc v2.1.212`, `v3.1.0`, `TODO: 1 / 3`,
  `🔀 2 / 1`) — 렌더 코드 포맷 문자열(renderer.go)을 따랐으나 실제
  렌더 스크린샷 대조는 안 함. 착지 전 `moai statusline` 샘플 파이프로
  한 번 확인 권장 (statusline.md 트러블슈팅 절의 방법).
- **팩토리 "기본 8 워커"** — cc.go 도움말("takes the default of 8
  workers") 근거. 런타임 상수 확인 안 함.
- **`moai tokens record` 플래그표** — tokens.go cobra 정의 근거
  (`--transcript/--session/--card/--role/--json`).
- **설정 탭 11개** — `consoleTabs()` 소스 근거. 웹콘솔 실화면 스크린
  대조 안 함.
- **Agent Teams "v3.1.1부터 기본 탑재"** — settings.json.tmpl L386에
  env=1 포함 확인. 단 정확히 어느 버전부터 들어갔는지는 커밋 시점
  조사 안 함(94bc55370가 배치 안이므로 v3.1.1로 서술했음 — 착지 전
  커밋 순서 재확인 권장).

## G5. 드래프트 미생성 (Stage 2 소형 변경)

authoring-plan.md Stage 2 표의 5개 파일(doctor·installation·
multi-model-audit·moai-review·model-policy)은 문안 방향만 제시,
전체 파일 드래프트는 미생성 — phase 2 또는 소형 착지 카드로.

## G6. 4-로케일 파생 검증

드래프트는 ko 정본만. en/ja/zh 파생물의 정합성(번역 누락·용어
통일 t59 규칙 준수)은 locale-translator 산출물에 대한 별도 검증
대상 — 본 phase 범위 밖.

## G7. launchers.md 기존 플래그 정확성

`-c, --continue`·`--permission-mode` 행 등 원본 표 항목의 소스 대조는
하지 않았다(원본 유지 항목). 새로 추가한 `-k` 4형태·`--name`만
cc.go 도움말과 대조 완료.

## G8. 검증 게이트 실행 기록 (phase 1)

- 드래프트 URL 블랙리스트 grep: 매치 0 — PASS
- 드래프트 mermaid LR/RL grep: 매치 0 — PASS
- 드래프트 본문 이모지: 원본 관행(이모지가 주제인 문서화)과 동일
  부류만 — PASS(판정 기록 G2)
- README 드래프트 H2 12개 = 원본 12개 — PASS (H3은 59→60, 팩토리 H3
  추가로 정상 — 파생 시 4-로케일 동일 적용 전제)
- 원본 트리 hugo --minify --gc: exit 0·무경고·sitemap OK — PASS
  (빌드 산출물 public/ 제거로 worktree 오염 0 확인)
