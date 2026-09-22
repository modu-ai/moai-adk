# t1078 구현 전 재검증

## Claim

기존 설계의 중복 옵션 탐지 범위와 spawn 크기 보호 전제를 보완해야 한다. 이 기록은 구현 또는 수용 완료 판정이 아니다.

## Evidence

`codex exec --help`의 실제 출력:

```text
  -c, --config <key=value>
          Override a configuration value that would otherwise be loaded from `~/.codex/config.toml`.
```

REQ-LMD-005의 `-c`만이 해당 값을 전달할 수 있다는 전제와 달리, 설치된 CLI는 `--config`를 명시적으로 제공한다. 단축형만 검사하는 구현은 충분하지 않다.

작은따옴표 40,000개를 JSON 문자열로 인코딩하고 기존 shellQuote 방식으로 인용한 Node 측정 결과:

```json
{"encoded_token_bytes":40025,"declared_ceiling":126976,"spawn_quoted_token_bytes":160027,"linux_limit_named_by_spec":131072}
```

측정 절차는 `developer_instructions=`에 JSON 문자열을 붙이고, 각 작은따옴표를 shellQuote의 닫기·이스케이프·다시 열기 표현으로 바꾼 후 양끝 인용부호를 추가한다. 이 측정은 문자열 길이 반례이며 Linux에서 실제 E2BIG를 재현한 결과는 아니다.

## Baseline-attribution

- 작업 트리: `.claude/worktrees/t1078`
- 브랜치: `WT-codex-local-md`
- HEAD: `1e00e35f8`
- 이번 실행에서 CLI 도움말과 문자열 변환 결과를 직접 측정했다.

## 보완 방향

- 충돌 검사는 CLI가 받는 단축형·긴 옵션형을 함께 처리한다. 붙여 쓰기와 등호형은 파서 검증 후 포함한다.
- 기존 합성 토큰 크기 검사에 더해 spawn에 실제 전달되는 최종 명령 문자열 크기를 검사한다. 고정 여유분만으로 인용 팽창을 보장한다고 기술하지 않는다.
- LIVE 수용의 원본 로컬 파일 수정 절차는 카드의 입력 파일 수정 금지와 충돌한다. 별도 검증 프로젝트의 로컬 파일에 nonce를 넣어 런처부터 실제 Codex 응답까지 측정하는 절차로 조정해야 한다.

## Gaps

- 위 보완은 SPEC 본문·plan·acceptance에 반영했다. 구현은 아직 변경하지 않았다.
- 실제 Linux exec 한도는 미측정이다. Codex 옵션의 붙여 쓰기·등호형은 아래 추가 측정에서 config 파서 도달을 확인했다.
- 기존 R1~R4 결정 기록과 구현·독립 감사·로컬 develop 병합이 남아 있다.

## Residual-risk

인자 하나의 크기 검사만으로 전체 argv와 환경의 합산 제한을 보장할 수 없다. 런처는 운영체제 실행 오류도 정확히 보고해야 한다.

## 구현 전 기존 테스트 재측정

2026-09-22, `WT-codex-local-md @ 1e00e35f8`에서 다음 명령을 실행했다.

```text
$ go test ./internal/cli -run '^TestCodexLocalInstructions_' -count=1 -timeout=60s
ok  	github.com/modu-ai/moai-adk/internal/cli	0.790s
exit=0
```

이는 기존 AGENTS.local.md 주입 테스트의 기준선이다. CLAUDE.local.md 병합, 옵션 충돌 보호, 안전한 파일 열기, 크기 보호, 실제 Codex LIVE 수용이 구현됐다는 증거는 아니다. 이번 기준선에서 해당 함수는 단일 파일을 Lstat 후 ReadFile하며 별도의 크기 검사를 수행하지 않는다.

## Codex config 파서 경계 추가 측정

설치된 `codex-cli 0.155.1`에서 아래 명령 5개를 각각 실행했다. `--help`나 `--version` 단축 경로를 사용하지 않고, 의도적으로 잘못된 정수 값을 넣어 config 타입 검사까지 도달하는지 확인했다.

```text
codex exec --config 'developer_instructions=1' 'Do not call tools; reply OK'
codex exec '--config=developer_instructions=1' 'Do not call tools; reply OK'
codex exec -c 'developer_instructions=1' 'Do not call tools; reply OK'
codex exec '-c=developer_instructions=1' 'Do not call tools; reply OK'
codex exec '-cdeveloper_instructions=1' 'Do not call tools; reply OK'
```

각 명령의 출력은 동일하며 모두 exit=1이었다.

```text
WARNING: proceeding, even though we could not create PATH aliases: Operation not permitted (os error 1)
Error loading config.toml: invalid type: integer `1`, expected a string
in `developer_instructions`
```

따라서 충돌 검사는 위 5가지 표기를 모두 포함해야 한다. 이 결과는 옵션의 config 파서 도달 증거이며, 모델의 지침 수신이나 LIVE 수용 통과 증거가 아니다.

## 변경 전 scoped coverage 기준선

동일한 `WT-codex-local-md @ 1e00e35f88f8029b978ff8b00f2f5be8f2d9c873`, Go `go1.26.8 darwin/arm64`에서 실행했다. 코드 변경은 없고 SPEC 문서만 수정된 상태다.

```text
$ go test ./internal/cli -run '^(TestCodexLocalInstructions_|TestCodexLocalSeparation$)' -count=1 -timeout=60s -coverprofile=/private/tmp/t1078-localmd-baseline.cover
ok  	github.com/modu-ai/moai-adk/internal/cli	0.849s	coverage: 6.5% of statements
exit=0

$ go tool cover -func=/private/tmp/t1078-localmd-baseline.cover | rg 'codexLocalDeveloperInstructionArgs|buildCodexSpawnCommand|total:'
github.com/modu-ai/moai-adk/internal/cli/codex_launcher.go:118: codexLocalDeveloperInstructionArgs 76.5%
github.com/modu-ai/moai-adk/internal/cli/codex_launcher.go:193: buildCodexSpawnCommand 0.0%
total: (statements) 6.5%
exit=0
```

위 함수별 출력은 정렬용 공백만 정규화했다. 6.5%는 선택한 테스트를 실행한 패키지 전체 분모의 수치이며 전체 테스트 실행의 커버리지가 아니다. 후속 구현은 동일 테스트 선택 기준선과 신규 테스트의 실행 범위를 구분하고, 신규·변경 함수의 85% 이상 커버리지를 따로 입증해야 한다.
