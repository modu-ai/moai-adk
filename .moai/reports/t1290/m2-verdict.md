# t1290 M2 이관 및 수신 판정

## Claim

`t1259`의 이관 명령을 사용해 카드 작업 트리의 실제 유지자 지침을 옮겼다. 커밋 `0ab480d9d`에는 `AGENTS.local.md`만 로컬 지침 파일로 남고, 길이는 39,999자 이하이다. 네 절의 절차 본문은 `.moai/docs/`에 보존했다. 이 결과는 파일 이관의 구조적 조건을 뒷받침하지만, 새 내용에 대한 Claude Code 수신 조건 `AC-LIR-009`는 아직 통과하지 못했다. 직접 `codex -C` 진입에서도 이 로컬 파일은 수신되지 않았다.

## Evidence

- `git merge --no-edit develop` → `Merge made by the 'ort' strategy.` 충돌 0건. 병합 뒤 `git rev-parse HEAD:CLAUDE.local.md`와 `git rev-parse develop:CLAUDE.local.md`는 모두 `d5e9f138ee43cbc7f3220f87c99a8a7f33ea4413`이었다.
- `make build` → `go build ... -o bin/moai ./cmd/moai`, exit 0. `./bin/moai migrate --help`에 `local-instructions`가 나타났다. M1 측정용 `AGENTS.local.md`를 Git 이력에 남기고 작업 트리에서 제거한 뒤 `./bin/moai migrate local-instructions` → `Migrated CLAUDE.local.md to AGENTS.local.md (backup: .moai/backups/local-instructions/20260928T042704Z-2552219930/CLAUDE.local.md).`
- `shasum -a 256 AGENTS.local.md <backup>`의 두 SHA-256은 이관 직후 모두 `5e8439b4d38a0e312e49734ee08624d552d3bff96a2f4f49d74a549b49f52745`였다. 이후 문서 절차를 분리하고 §0·파일 참조를 갱신했다.
- `git show develop:CLAUDE.local.md | wc -m` → `45810`. `git show 0ab480d9d:AGENTS.local.md | wc -m` → `37061`. `git ls-files AGENTS.local.md CLAUDE.local.md` → `AGENTS.local.md` 한 줄. `git show 0ab480d9d:CLAUDE.local.md` → exit 128, `path 'CLAUDE.local.md' does not exist in '0ab480d9d'`.
- 백업의 §28~31과 새 `.moai/docs/{lsel-drain-operations,jev-local-operations,stale-card-premise-check,kickoff-autonomy}.md`를 기계 비교했다. 제목의 `## 번호.`를 `#`로 바꾸고 문서 끝 빈 줄만 정리한 본문은 각 절과 일치했다. 기존 gitflow 절차는 `.moai/docs/gitflow-integration-chain.md`에 있다.
- `go test ./internal/cli -run 'TestMigrateLocalInstructions|TestCodexLocalInstructions|TestLocalInstructionsAdvisory' -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 0.710s`. 템플릿·루트 문구 갱신 후 `make build` exit 0 (`/tmp/t1290-build.log`).
- 도구를 끈 Claude Code 수신 검사: `timeout 180 claude -p ... --tools '' --output-format json` → exit 1, `You've hit your weekly limit · resets Oct 1 at 12pm (Asia/Seoul)` (`m2-claude-probe.json`). GLM 환경으로 재시도한 동일 질문 → `glm-5.3-flash` 미인식 진단 뒤 `timeout` exit 124, 결과 파일 0바이트 (`m2-claude-glm-probe.json`). 응답이 없어 GLM 서버 도달 여부도 확정할 수 없다. 어느 실행도 내용 수신 PASS가 아니다.
- `codex -C <t1290 worktree> exec --sandbox read-only --json`의 도구 없는 응답은 첫 시도 `ABSENT / ABSENT`, 양성 대조 시도 `tail / ABSENT`였다 (`m2-codex-events.jsonl`, `m2-codex-control-events.jsonl` 및 각 `*-last.txt`). 이벤트에는 `item.completed`의 도구 항목이 없었다. 두 실행 모두 스킬 설명 컨텍스트 예산 초과 진단을 동반했다. `tail`은 루트 `AGENTS.md`의 Budget warning에 있고, 드레인 래퍼는 `AGENTS.local.md` §28에만 있다.
- [Codex 공식 지침 파일 탐색 문서](https://learn.chatgpt.com/docs/agent-configuration/agents-md)는 한 디렉터리에서 `AGENTS.override.md`, `AGENTS.md`, 설정된 fallback 중 하나만 읽는다고 명시한다. 기본 제한은 32 KiB이며, sibling `AGENTS.local.md`는 자동 합쳐지지 않는다. 위 직접 실행 결과와 부합한다.

## Baseline-attribution

이관 전 원본은 현재 `develop`의 `9da000bf2`와 동일한 `CLAUDE.local.md` blob이다. 이관 후 구조 검사는 카드 브랜치 커밋 `0ab480d9d`를 읽었다. 템플릿 문구 수정과 빌드는 그 뒤의 미커밋 작업 트리에서 측정했으므로, `0ab480d9d`의 내용으로 주장하지 않는다. M1의 수신 근거는 `.moai/reports/t1290/m1-probes.md`의 각 leg이며, M2 새 내용 수신의 대체 근거는 아니다.

## Gaps

- 새 `AGENTS.local.md`를 실은 Claude Code의 답변은 계정 한도와 GLM 응답 시간 때문에 관측하지 못했다. `AC-LIR-009`는 열린 상태다.
- 직접 `codex -C`에서 로컬 지침 수신이 관측되지 않았다. 현재 작업 트리의 `moai codex status`는 `.codex/hooks.json` 누락으로 `wiring partial`, `agents 0 TOML`을 반환하므로 MoAI 런처 실세션의 새 내용 수신도 확인하지 못했다.
- 템플릿 문구 수정은 아직 커밋·병합·CI 판정을 받지 않았다. primary `main` 체크아웃의 구형 공유 워킹 사본은 건드리지 않았다.

## Residual-risk

Claude 수신 재검사 없이 이관 브랜치를 `develop`에 합치면 새 카드 레인이 유지자 지침을 실제로 읽는지 확인되지 않는다. 직접 `codex -C`를 `moai codex`와 동등한 로컬 지침 경로로 안내하면 유지자 지침이 빠진 세션을 만들 수 있다. 현재 이관은 카드 브랜치에만 있으며, 이 판정서는 t1290 완료 또는 배포 승인으로 쓰지 않는다.
