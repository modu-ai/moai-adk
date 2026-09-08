# t452 M3 — 매니페스트 행 전환과 재생성

카드: t452 · SPEC-CODEX-SKILL-LOADER-001 · BASELINE_SHA `c529b2e4aaf5148aee7e6c67649bf392837bbb06`
codex 버전(이 회차 관측): `codex-cli 0.152.1`

## 전환 — `deferred-m1` → `documented-drop`

M2 가 세 성질을 모두 **확인되지 않음**으로 판정했으므로(`m2-verdict.md`), REQ-CSL-005 가
방출을 금지하고 REQ-CSL-007 경로가 성립한다. `internal/template/agentemit/agents-codex.yaml`
의 `skill-loader` 행을 `documented-drop` 으로 전환하고 rationale 에 관측 명령·출력 요지·
codex 버전을 담았다.

```
$ grep -A4 'class: skill-loader' internal/template/agentemit/agents-codex.yaml
  - class: skill-loader
    disposition: documented-drop
    rationale: >-
      Measured on codex-cli 0.152.1: an agent-role skills key is not
      observable at all on the non-interactive path, so the ship-omitted rule
```

`deferred-m1` 은 어느 행에도 남지 않는다:

```
$ grep -c 'deferred-m1' internal/template/agentemit/agents-codex.yaml
0
```

(`deferred-m1` 문자열은 `manifest.go`/`agentemit_test.go` 의 허용 disposition 집합에
상수로 남아 있다. 그것은 파서가 받아들이는 값의 목록이지 어떤 행의 처분이 아니므로,
D4("유예 행은 어느 분기에서도 `deferred-m1` 로 남지 않는다")는 충족된다.)

## `codex_measured_version` 은 **의도적으로 유지**했다 — `0.147.0`

```
$ grep 'codex_measured_version' internal/template/agentemit/agents-codex.yaml
codex_measured_version: "0.147.0"
```

AC-CSL-009 의 [HARD] 조항: 이 필드는 "아래 필드 의미론을 잰 버전"을 뜻한다. 이 회차가
재측정한 것은 `skill-loader` 한 축뿐이고, 선행 7 개 필드(`args`, `command`,
`description`, `developer_instructions`, `model_reasoning_effort`, `name`,
`sandbox_mode`)는 재지 않았다. 0.152.1 로 올리면 매니페스트가 **갖지 않은 커버리지를
주장**하게 되므로 유지했다. 이 회차의 버전은 `skill-loader` 행의 rationale 첫 줄에
`codex-cli 0.152.1` 로 박혀 있다 — 잰 대상 옆이다.

**이것은 조용한 미갱신이 아니라 기록된 보류다.** 그 구분을 남기는 것이 AC-CSL-009 가
요구하는 바다.

## 네 명령 — 순서대로, 전부 exit 0

```
$ make agents-emit                                   → exit 0   (m3-runs/m3-agents-emit.txt)
AGENTEMIT_UPDATE=1 go test ./internal/template/agentemit/... -run TestGoldenCommittedArtifactsMatchEmission
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	1.327s

$ make build                                         → exit 0   (m3-runs/m3-build.txt, 말미)
catalog.yaml updated successfully (12899 bytes)
go build -ldflags "..." -o bin/moai ./cmd/moai

$ make agents-emit-check                             → exit 0   (m3-runs/m3-emit-check.txt)
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.492s

$ go test ./internal/template/agentemit/... -count=1 → exit 0   (m3-runs/m3-gotest.txt)
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.331s
```

마지막 명령의 출력에 `ok` 가 있다 → AC-CSL-008 PASS.

## [HARD] 물려받은 드리프트 2 건 — 이 회차의 변경이 **아니다**

`make agents-emit` 과 `make build` 가 이 SPEC 과 무관한 스테일 상태 두 건을 수리했다.
이 회차의 성과로 주장하지 않고, 귀속을 여기 적는다.

### (a) `internal/template/templates/.codex/agents/moai/sync-auditor.toml`

중립 `.md` 원본에 있던 "Export mandate" 문단이 커밋된 `.toml` 에는 없었다 — 즉
**BASELINE_SHA 시점에 이미 소스 축 드리프트가 있었고 `make agents-emit-check` 는 그때
붉었다.** baseline 블롭으로 확인:

```
$ git show c529b2e4aaf5148aee7e6c67649bf392837bbb06:.claude/agents/moai/sync-auditor.md | grep -c "Export mandate"
1
$ git show c529b2e4aaf5148aee7e6c67649bf392837bbb06:internal/template/templates/.codex/agents/moai/sync-auditor.toml | grep -c "Export mandate"
0
$ git show c529b2e4aaf5148aee7e6c67649bf392837bbb06:internal/template/templates/.claude/agents/moai/sync-auditor.md | grep -c "Export mandate"
1
```

수리를 커밋에 담는 이유: 담지 않으면 `make agents-emit-check` 가 커밋된 상태에서 다시
붉어진다(= AC-CSL-008 이 요구한 초록이 워킹 트리에만 존재하게 된다). 되돌리는 쪽이
일부러 붉음을 복원하는 것이므로 더 나쁘다. REQ-CSL-008 이 명령을 의무화했고 이 파일은
그 명령의 출력이므로, scope envelope 안의 cascade 로 본다.

### (b) `internal/template/catalog.yaml` — 해시 2 개

`make build` 의 선행 `gen-catalog-hashes.go --all` 이 스테일 해시 두 개를 갱신했다:
`sync-auditor`(위 (a) 의 `.md` 미러)와 `moai` 스킬. **두 원본 파일 모두 이 회차에
편집하지 않았다** — `git status --porcelain` 에 나타나지 않는다. 즉 baseline 에서 이미
스테일했다. (a) 와 같은 이유로 커밋에 담고 귀속을 여기 적는다.

## [HARD] 귀속 불가 변경 1 건 — `.claude/settings.json` (커밋하지 않음)

세션 도중 `.claude/settings.json` 이 21 행에서 215 행으로 바뀌었다. **이 SPEC 의 작업이
아니며 무엇이 썼는지 확정하지 못했다.**

- 세션 첫 `git status --porcelain` 에서는 깨끗했다(수정된 추적 파일은 `acceptance.md` 뿐).
- mtime `2026-09-03 18:11:48` — `make agents-emit`(18:15:11 산출물)·`make build`(18:15:28
  catalog, 18:15:47 바이너리)보다 **앞선다**. 따라서 M3 의 세 명령이 쓴 것이 아니다.
- Makefile `build` 타깃은 `agents-emit-check` · `templ-generate` · `gen-catalog-hashes.go`
  · `go build` 넷뿐이고 `.claude/` 를 대상으로 하지 않는다.

**조치: 스테이징하지 않는다.** 추적 파일이므로 git 이 안전망이고, 워킹 트리에 남겨 리드가
판단하게 한다. 조용히 담거나 조용히 되돌리는 쪽이 둘 다 관측 사실을 지운다.
