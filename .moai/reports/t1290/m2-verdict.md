# t1290 M2 이관 및 수신 판정

## Claim

`t1259`의 이관 명령을 사용해 카드 작업 트리의 실제 유지자 지침을 옮겼다. 커밋 `0ab480d9d`에는 `AGENTS.local.md`만 로컬 지침 파일로 남고, 길이는 39,999자 이하이다. 네 절의 절차 본문은 `.moai/docs/`에 보존했다. MoAI Codex 런처는 새 내용을 실제로 수신했다. 직접 `codex -C`는 자동 주입을 하지 않지만, 루트 `AGENTS.md`에 첫 도구 읽기 의무를 추가한 작업 트리에서는 파일 전문을 읽었다. Claude Code 수신 조건 `AC-LIR-009`는 아직 통과하지 못했다.

## Evidence

- `git merge --no-edit develop` → `Merge made by the 'ort' strategy.` 충돌 0건. 병합 뒤 `git rev-parse HEAD:CLAUDE.local.md`와 `git rev-parse develop:CLAUDE.local.md`는 모두 `d5e9f138ee43cbc7f3220f87c99a8a7f33ea4413`이었다.
- `make build` → `go build ... -o bin/moai ./cmd/moai`, exit 0. `./bin/moai migrate --help`에 `local-instructions`가 나타났다. M1 측정용 `AGENTS.local.md`를 Git 이력에 남기고 작업 트리에서 제거한 뒤 `./bin/moai migrate local-instructions` → `Migrated CLAUDE.local.md to AGENTS.local.md (backup: .moai/backups/local-instructions/20260928T042704Z-2552219930/CLAUDE.local.md).`
- `shasum -a 256 AGENTS.local.md <backup>`의 두 SHA-256은 이관 직후 모두 `5e8439b4d38a0e312e49734ee08624d552d3bff96a2f4f49d74a549b49f52745`였다. 이후 문서 절차를 분리하고 §0·파일 참조를 갱신했다.
- `git show develop:CLAUDE.local.md | wc -m` → `45810`. `git show 0ab480d9d:AGENTS.local.md | wc -m` → `37061`. `git ls-files AGENTS.local.md CLAUDE.local.md` → `AGENTS.local.md` 한 줄. `git show 0ab480d9d:CLAUDE.local.md` → exit 128, `path 'CLAUDE.local.md' does not exist in '0ab480d9d'`.
- 백업의 §28~31과 새 `.moai/docs/{lsel-drain-operations,jev-local-operations,stale-card-premise-check,kickoff-autonomy}.md`를 기계 비교했다. 제목의 `## 번호.`를 `#`로 바꾸고 문서 끝 빈 줄만 정리한 본문은 각 절과 일치했다. 기존 gitflow 절차는 `.moai/docs/gitflow-integration-chain.md`에 있다.
- `go test ./internal/cli -run 'TestMigrateLocalInstructions|TestCodexLocalInstructions|TestLocalInstructionsAdvisory' -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 0.710s`. 템플릿·루트 문구 갱신 후 `make build` exit 0 (`/tmp/t1290-build.log`).
- 도구를 끈 Claude Code 수신 검사: `timeout 180 claude -p ... --tools '' --output-format json` → exit 1, `You've hit your weekly limit · resets Oct 1 at 12pm (Asia/Seoul)` (`m2-claude-probe.json`). GLM 환경으로 재시도한 동일 질문 → `glm-5.3-flash` 미인식 진단 뒤 `timeout` exit 124, 결과 파일 0바이트 (`m2-claude-glm-probe.json`). 응답이 없어 GLM 서버 도달 여부도 확정할 수 없다. 어느 실행도 내용 수신 PASS가 아니다.
- `codex -C <t1290 worktree> exec --sandbox read-only --json`의 도구 없는 응답은 첫 시도 `ABSENT / ABSENT`, 양성 대조 시도 `tail / ABSENT`였다 (`m2-codex-events.jsonl`, `m2-codex-control-events.jsonl` 및 각 `*-last.txt`). 이벤트에는 `item.completed`의 도구 항목이 없었다. 두 실행 모두 스킬 설명 컨텍스트 예산 초과 진단을 동반했다. `tail`은 루트 `AGENTS.md`의 Budget warning에 있고, 드레인 래퍼는 `AGENTS.local.md` §28에만 있다.
- 이후 루트·템플릿 `AGENTS.md` §8에 직접 Codex 세션의 첫 로컬 파일 읽기 의무를 적고, 같은 작업 트리에서 `codex -C <t1290 worktree> exec --sandbox read-only --json ... '프로젝트 지침이 요구하는 세션 시작 확인을 끝낸 뒤 READY라고만 답해.'`를 다시 실행했다. exit 0, 최종 `READY`. 이벤트의 두 번째 명령은 `cat AGENTS.local.md`(exit 0, 출력 37,061자)였고, 뒤이어 `sed`로 1~678행을 확인했다. 전체 명령 7개가 모두 exit 0이다. 중복 전문을 싣지 않은 추적은 `m2-codex-direct-read-trace.md`; 원본 JSONL은 작업 트리 로컬에만 둔다. 이 검사는 도구로 **읽은** 사실을 증명하며 Codex가 로컬 파일을 시작 전에 자동 주입했다는 뜻은 아니다. 검증 뒤 잘림 위험을 줄이려고 조항을 파일 앞부분으로 옮겼고, 이 최종 배치의 별도 직접 Codex 재검사는 하지 않았다.
- t1290 작업 트리에는 `.codex/hooks.json`이 없었다. primary의 훅 파일이 `moai hook ... --harness codex` 8개 명령만 담은 것을 `jq`로 읽은 뒤, 이 작업 트리의 런타임 경로로 복사했다(추적·커밋하지 않음). `./bin/moai codex status` → `wiring   wired (.codex/hooks.json, .codex/config.toml)` 및 `agents   0 TOML`. `timeout 120 ./bin/moai codex -- exec --sandbox read-only --json ...` → exit 0, 도구 없는 답변 `session_drain.sh` / `git log --all -S` (`m2-moai-codex-events.jsonl`, `m2-moai-codex-last.txt`). 호출에 `-w`나 `--worktree` 인자는 주지 않았다. 이벤트에 스킬 설명 컨텍스트 예산 초과 진단은 있으나, 답변의 두 값은 새 로컬 지침 §28·§30과 일치한다.
- 같은 GLM 인증 파일을 사용한 별도의 최소 Anthropic 호환 API 요청은 HTTP 429를 반환했다. 응답의 `rate_limit_error`는 `Usage limit reached for 5 hour. Your limit will reset at 2026-09-28 13:17:58`라고 적었다(`m2-glm-rate-limit.json`). 이 결과가 앞선 Claude Code timeout의 원인이라고 단정하지 않는다.
- [Codex 공식 지침 파일 탐색 문서](https://learn.chatgpt.com/docs/agent-configuration/agents-md)는 한 디렉터리에서 `AGENTS.override.md`, `AGENTS.md`, 설정된 fallback 중 하나만 읽는다고 명시한다. 기본 제한은 32 KiB이며, sibling `AGENTS.local.md`는 자동 합쳐지지 않는다. 위 직접 실행 결과와 부합한다.

## Baseline-attribution

이관 전 원본은 `develop`의 `9da000bf2`와 동일한 `CLAUDE.local.md` blob이다. 이관 후 구조 검사는 카드 브랜치 커밋 `0ab480d9d`를 읽었다. 직접 Codex 수신 검사는 그 뒤 `9dca1215d` 위에 루트·템플릿 지침을 수정한 작업 트리에서 측정했다. 앞부분으로 조항을 옮긴 뒤 `make build`도 exit 0이었다(`/tmp/t1290-build-early.log`). M1의 수신 근거는 `.moai/reports/t1290/m1-probes.md`의 각 leg이며, M2 새 내용 수신의 대체 근거는 아니다.

## Gaps

- 새 `AGENTS.local.md`를 실은 Claude Code의 답변은 계정 한도와 GLM 응답 시간 때문에 관측하지 못했다. `AC-LIR-009`는 열린 상태다.
- 직접 `codex -C`는 로컬 파일을 사전 주입하지 않는다. 새 첫 읽기 규칙은 도구를 사용할 수 있는 실제 세션 한 번에서 검증했으며, 도구를 꺼 둔 세션에는 적용할 수 없다. MoAI Codex 런처의 실제 수신은 위 테스트로 확인했으나, 런타임 훅 파일을 복사한 이 작업 트리에서만 잰 것이다. 기본 worktree 생성 경로가 그 파일을 심는지는 이 테스트의 범위가 아니다.
- 템플릿 문구 수정은 아직 `develop` 병합·CI 판정을 받지 않았다. primary `main` 체크아웃의 구형 공유 워킹 사본은 건드리지 않았다.

## Residual-risk

Claude 수신 재검사 없이 이관 브랜치를 `develop`에 합치면 새 카드 레인이 유지자 지침을 실제로 읽는지 확인되지 않는다. 직접 `codex -C`의 첫 도구 읽기는 런처의 사전 developer-instructions 주입과 우선순위·도구 가용성이 다르다. 현재 이관은 카드 브랜치에만 있으며, 이 판정서는 t1290 완료 또는 배포 승인으로 쓰지 않는다.

## M2 완결 보강 (worker-61, 14:45 KST — AC-LIR-009 클로즈)

재개 지시 수령 후 제 머지 HEAD `646ae8302` 에서 외부 실행분(6524d26cc 흡수 + 0ab480d9d~a8496d933, 리드 인지 상태)을 **감사-판독으로 검증**했다(재실행 아님 — 발산 보고 규율, 리드 보고에 명시):

- `git show HEAD:AGENTS.local.md | wc -m` → `37061` (≤39,999, AC-IFU-007 ✓) · `git ls-files AGENTS.local.md CLAUDE.local.md` → `AGENTS.local.md` 1행 (AC-LIR-007 ✓) · `git show HEAD:CLAUDE.local.md` → exit 128 (구파일 소멸 ✓) — 모두 이번 실행·이 트리.
- **AC-LIR-009 Claude 프로브 (GLM 게이트웨이, 429 리셋 13:17 이후, 1턴·timeout 180)**:
  ```
  $ timeout 180 claude -p "…(1) is AGENTS.md loaded…; (2) quote the exact branch name that AGENTS.local.md §0.1 names as the tree whose copy is canonical…, or NOT_PRESENT…" --model claude-haiku-4-5-20251001
  LOADED
  `develop`
  probe-exit=0
  ```
  → control present + §0.1 판별값 `develop` 수신 — **내용 교체 뒤에도 수신 생존**. 판별값은 이관된 실제 파일에서만 얻을 수 있다(TANGO9 fixture 는 소멸 상태).
- 문서 절반: 옛 부정 문장 `worktree session does not receive` → root+template AGENTS.md/CLAUDE.md 전면 grep 0 (exit 2) · 새 문면 「worktree-root AGENTS.local.md 를 다른 작업 전에 전문 읽는다 + MoAI Codex launcher 는 주입」 확인.

**M2+M3 상태: AC-LIR-001~009 전부 통과.** 잔여: develop 통합(병합 창)과 그 CI 판정. 본 보강은 원 판정서(외부 실행분)의 Gaps 중 AC-LIR-009 행을 닫는다.
