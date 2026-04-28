## Summary

<!-- Brief description of the changes -->

## Changes

<!-- List the key changes made -->

-

## Type of Change

<!-- Select ONE primary type — see CLAUDE.local.md §18.6 for label 3축 체계 -->

- [ ] `type:feature` — New capability (non-breaking)
- [ ] `type:fix` — Bug fix (non-breaking)
- [ ] `type:docs` — Documentation only
- [ ] `type:chore` — Maintenance (dependencies, cleanup)
- [ ] `type:ci` — CI/CD / GitHub Actions
- [ ] `type:refactor` — Code restructuring (no behavior change)
- [ ] `type:security` — Security fix or hardening
- [ ] `type:test` — Test additions or improvements
- [ ] **Breaking change** — Public API change (requires major version bump)

## Merge Strategy (CLAUDE.local.md §18.3)

<!-- Reviewer: use the correct `gh pr merge` flag -->

- [ ] **`--squash`** (default for `feat/`, `fix/`, `chore/`, `docs/`, `plan/` branches) — WIP commit 정리
- [ ] **`--merge`** (`release/` or `hotfix/` branches) — 릴리스 마일스톤 + 개별 SPEC commit 보존
- [ ] **Dependabot auto-merge** (dependabot/* branches) — 자동 squash

> ⚠️ Release PR (`release/vX.Y.Z → main`)은 반드시 `--merge` 사용. Squash 시 개별 feature history 손실.

## Testing

- [ ] `go test ./...` 통과
- [ ] `go test -race ./...` (race detection) 통과
- [ ] `golangci-lint run` 통과
- [ ] 새 기능에 대한 테스트 추가됨
- [ ] 커버리지 85%+ 유지

## Quality Gates

- [ ] CI all checks green (Lint, Test ubuntu/macos/windows, Build 5 platforms, CodeQL)
- [ ] `@MX` tag 규율 준수 (fan_in ≥ 3 → ANCHOR, goroutine/complexity ≥ 15 → WARN, 필수 `@MX:REASON`)
- [ ] SPEC 변경 시 frontmatter `status` 필드 적절히 업데이트 (draft → implemented)
- [ ] SARIF / wire format 등 public contract 불변 (non-breaking 보장)

## AI Collaboration

- [ ] AI 도구 (Claude Code, Codex 등) 협업으로 생성됨
- [ ] AI 생성 코드가 사람에 의해 리뷰됨
- [ ] AI agent 이름이 commit 메시지에 명시됨 (Co-Authored-By 또는 footer)

## Checklist

- [ ] 프로젝트 코딩 표준 준수 (Go convention, 주석 한국어, 식별자 영어)
- [ ] Self-review 완료
- [ ] 이해하기 어려운 부분에 주석 추가
- [ ] 필요 시 문서 업데이트 (README, CHANGELOG, docs-site)
- [ ] Secrets/credentials 포함 없음
- [ ] Branch 명명 규칙 준수 (§18.2)

## Related Issues

<!-- Link related issues: Fixes #123, Refs #456 -->

## Post-Merge Actions (release PR 전용)

<!-- Release PR 머지 후 실행 -->

- [ ] `./scripts/release.sh vX.Y.Z` 로 tag + GoReleaser 자동 실행
- [ ] GitHub Release 생성 확인 (GoReleaser 자동)
- [ ] docs-site 4개국어 reference 페이지 업데이트 (별도 PR)
