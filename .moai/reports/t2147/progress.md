# t2147 progress - bump version to v3.2.1

Template: previous bump a85be12449 (v3.2.0; also eba919e44b for v3.1.3). They touched version.go, system.yaml, 4 READMEs, docs-site/hugo.toml (a85be12449 also edited CHANGELOG + tests; CHANGELOG NOT edited here, per card).

## Files changed (commit 801eba9d92)
- pkg/version/version.go: v3.2.0 -> v3.2.1
- .moai/config/sections/system.yaml: template_version and version v3.2.0 -> v3.2.1
- README.md, README.ko.md, README.ja.md, README.zh.md line 24: Release-v3.1.2 -> Release-v3.2.1
- docs-site/hugo.toml: version v3.2.0 -> v3.2.1; releaseDate 2026-10-23 -> 2026-11-07 (not in lane pre-read; previous bump changed it)
- system.yaml.tmpl uses {{.Version}}; no literal, unchanged. No make build needed (no template edit).

## Verification
- go build ./... : no output (empty = success; exit status masked by `| tail`)
- go test -count=1 ./pkg/version/ : ok  github.com/modu-ai/moai-adk/pkg/version 0.256s
- go test -count=1 -run TestVersionStamp ./internal/cli/ : ok  github.com/modu-ai/moai-adk/internal/cli 8.722s

## Gaps
- releaseDate 2026-11-07 is an assumption (today); confirm against the actual release date.
- TestReleasePreflight* and other tests not run; no full suite.

## 리더 판정 반영 (lane-28, 2026-11-07)

- Codex 호출 1회가 실행 전 `stale_server`로 거절(기록 `codex-refusal-1.md`, 커밋 aaaebd5e03). 리더 판정: 리뷰 전 거절은 305c 횟수에 넣지 않음(t2118 ede0와 같은 계산). /mcp 재연결은 리더가 운영자에게 요청함. 재연결 뒤 같은 HEAD에서 Codex 1회 재호출.
- D1 `docs-site/hugo.toml:56` `releaseDate = "2026-11-07"`: 태그 날 정정하는 쪽으로 기록(리더 지시).
- 범위 밖 메모: CHANGELOG `[3.2.1]` 절은 t2148(lane-26) 담당. `.moai/docs/version-management.md:146-150`의 `make plugin-emit` 단계는 대상·`plugins/` 디렉터리가 없어 낡은 문구 — 메모로만 남김.
