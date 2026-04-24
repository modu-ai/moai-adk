---
name: Release
about: Release tracking issue (Minor/Major/Hotfix)
title: "release: vX.Y.Z — <release theme>"
labels: ["type:chore", "priority:P1", "area:ci"]
assignees: []
---

## Release Plan

**Version**: `vX.Y.Z`
**Type**: [ Minor | Major | Patch | Hotfix ]
**Target date**: <YYYY-MM-DD>
**Release branch**: `release/vX.Y.Z` (Minor/Major only) or `hotfix/vX.Y.Z-*` (Hotfix)

## Scope

<!-- SPEC-IDs included in this release -->

- [ ] SPEC-XXX-001: ...
- [ ] SPEC-XXX-002: ...

## Pre-Release Checklist (CLAUDE.local.md §18.8)

### Code Quality
- [ ] 모든 SPEC 구현 완료 (`status: implemented` 로 전환)
- [ ] 51/51 (또는 전체) ACs PASS
- [ ] `go test -race ./...` PASS (ubuntu/macos/windows CI)
- [ ] `golangci-lint run ./...` 0 issues
- [ ] `manager-quality` review 통과 (TRUST 5 ≥ 4/5)

### Documentation
- [ ] `CHANGELOG.md` 에 `## [X.Y.Z] - YYYY-MM-DD` 섹션 추가 (영문 + 한국어)
- [ ] Non-breaking 선언 또는 Breaking Changes 섹션 명시
- [ ] Detection Improvements 섹션 (UX 영향 있는 감지 변경 시)
- [ ] Deferred items 섹션 (next minor/major 예약 항목)
- [ ] README.md version line 업데이트 (필요 시)

### Git State
- [ ] `release/vX.Y.Z` 브랜치 또는 main 최신
- [ ] 모든 PR merge 완료
- [ ] 작업 트리 clean (`git status --short` empty)
- [ ] origin/main과 동기화 (local/remote SHA 일치)

### Branch Protection (admin)
- [ ] `main` branch protection 활성화 (CI required, 1 review)
- [ ] Force push 차단 확인

## Release Execution

Minor/Major Release:
```bash
# 1. release 브랜치에서 main으로 PR 생성 + merge commit
gh pr merge <release-PR> --merge --delete-branch

# 2. main pull + 릴리스 스크립트
git checkout main && git pull origin main
./scripts/release.sh vX.Y.Z
# OR: make release V=vX.Y.Z
```

Hotfix Release:
```bash
# 1. main의 latest tag에서 분기
git checkout -b hotfix/vX.Y.Z-<topic> <latest-tag>

# 2. 수정 + PR + merge commit
gh pr merge <hotfix-PR> --merge --delete-branch

# 3. Release 스크립트 실행
git checkout main && git pull origin main
./scripts/release.sh vX.Y.Z --hotfix
# OR: make release-hotfix V=vX.Y.Z
```

## Post-Release Checklist

- [ ] GitHub Release 생성 확인 (GoReleaser 자동)
- [ ] Release assets 5 플랫폼 업로드 확인 (darwin amd64/arm64, linux amd64/arm64, windows amd64)
- [ ] `latest` release marker 확인
- [ ] docs-site 4개국어 reference 페이지 업데이트 (별도 PR, `docs/vX.Y.Z-reference-sync`)
- [ ] v2.X+1 backlog SPEC 분석 및 초안 작성 (next cycle kickoff)
- [ ] Release announcement (선택: Discord, Twitter, blog)
- [ ] 이 Issue close

## Release Notes

<!-- Post-release: 간단한 release 요약 붙여넣기 -->

---

**Reference**: CLAUDE.local.md §18 Enhanced GitHub Flow
