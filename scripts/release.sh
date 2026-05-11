#!/usr/bin/env bash
# scripts/release.sh — Enhanced GitHub Flow 기반 로컬 릴리스 실행
#
# 사용법:
#   scripts/release.sh v2.15.0              # Minor/Major release (release branch 머지 완료 전제)
#   scripts/release.sh v2.14.1              # Patch release (fix 직접 main)
#   scripts/release.sh v2.14.1 --hotfix     # Hotfix release
#   scripts/release.sh v2.15.0 --dry-run    # 검증만 (실제 tag/push 없이)
#
# 전제 조건 (CLAUDE.local.md §18.8):
#   - CHANGELOG.md 에 해당 버전 섹션 존재
#   - main 브랜치 checkout + origin/main 과 동기화
#   - 모든 CI 통과
#   - 작업 트리 clean
#
# 흐름:
#   1. 사전 검증 (버전 형식, CHANGELOG, git 상태, CI 상태)
#   2. 사용자 확인 (AskUserQuestion 없이 stdin prompt)
#   3. Annotated tag 생성 (CHANGELOG 섹션에서 annotation 자동 추출)
#   4. Tag push → GoReleaser 자동 실행 (release.yml workflow)
#   5. GitHub Release 상태 확인 (GoReleaser 완료까지 대기)
#
# 참고:
#   - GoReleaser가 GitHub Release 및 바이너리 배포 자동 처리
#   - 수동 `gh release create`는 GoReleaser와 충돌 가능하므로 사용 금지

set -euo pipefail

# ─── Color helpers ─────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m' # No Color

log_info()   { echo -e "${BLUE}[INFO]${NC} $*"; }
log_ok()     { echo -e "${GREEN}[OK]${NC}   $*"; }
log_warn()   { echo -e "${YELLOW}[WARN]${NC} $*"; }
log_error()  { echo -e "${RED}[FAIL]${NC} $*" >&2; }
die()        { log_error "$*"; exit 1; }

# ─── Argument parsing ──────────────────────────────────────────────────────
VERSION=""
DRY_RUN=false
HOTFIX=false
SKIP_CI_CHECK=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --dry-run)       DRY_RUN=true; shift ;;
        --hotfix)        HOTFIX=true; shift ;;
        --skip-ci-check) SKIP_CI_CHECK=true; shift ;;
        -h|--help)
            sed -n '2,20p' "$0"
            exit 0
            ;;
        -*)
            die "Unknown flag: $1 (try --help)"
            ;;
        *)
            if [[ -z "$VERSION" ]]; then
                VERSION="$1"
            else
                die "Multiple version arguments provided"
            fi
            shift
            ;;
    esac
done

[[ -n "$VERSION" ]] || die "Version argument required (e.g., v2.15.0). Try --help."

# ─── Validation 1: Version format (SemVer with v prefix) ────────────────────
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]]; then
    die "Invalid version format: $VERSION (expected: vX.Y.Z or vX.Y.Z-preN)"
fi

log_info "Release version: ${BOLD}$VERSION${NC}"
[[ "$DRY_RUN" == true ]] && log_warn "DRY RUN mode — no tag/push will occur"
[[ "$HOTFIX" == true ]] && log_info "Hotfix mode — relaxed branch check"

# ─── Validation 2: Repository root ─────────────────────────────────────────
REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || die 'Not in a git repository')"
cd "$REPO_ROOT"

# ─── Validation 3: Clean working tree ──────────────────────────────────────
if [[ -n "$(git status --porcelain)" ]]; then
    git status --short
    die "Working tree is dirty. Commit or stash changes first."
fi
log_ok "Working tree clean"

# ─── Validation 4: Current branch ──────────────────────────────────────────
CURRENT_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
if [[ "$CURRENT_BRANCH" != "main" ]]; then
    if [[ "$HOTFIX" == true ]]; then
        log_warn "On branch '$CURRENT_BRANCH' (hotfix mode — allowed)"
    else
        die "Must be on 'main' branch (current: $CURRENT_BRANCH). Use --hotfix for hotfix branches."
    fi
fi
log_ok "On expected branch: $CURRENT_BRANCH"

# ─── Validation 5: Synced with origin ──────────────────────────────────────
git fetch origin --tags --quiet
LOCAL_SHA="$(git rev-parse HEAD)"
REMOTE_SHA="$(git rev-parse "origin/$CURRENT_BRANCH" 2>/dev/null || echo "")"

if [[ -z "$REMOTE_SHA" ]]; then
    die "Remote branch 'origin/$CURRENT_BRANCH' not found. Push branch first."
fi

if [[ "$LOCAL_SHA" != "$REMOTE_SHA" ]]; then
    AHEAD="$(git rev-list --count "$REMOTE_SHA..$LOCAL_SHA" 2>/dev/null || echo "?")"
    BEHIND="$(git rev-list --count "$LOCAL_SHA..$REMOTE_SHA" 2>/dev/null || echo "?")"
    die "Local '$CURRENT_BRANCH' diverged from origin (ahead: $AHEAD, behind: $BEHIND). Pull/push first."
fi
log_ok "Local $CURRENT_BRANCH synced with origin"

# ─── Validation 6: Tag does not exist ──────────────────────────────────────
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    die "Tag $VERSION already exists (local)."
fi
if git ls-remote --exit-code --tags origin "$VERSION" >/dev/null 2>&1; then
    die "Tag $VERSION already exists on origin."
fi
log_ok "Tag $VERSION does not exist yet"

# ─── Validation 7: CHANGELOG.md 에 해당 버전 섹션 존재 ───────────────────────
CHANGELOG_VERSION="${VERSION#v}" # v2.14.0 → 2.14.0
CHANGELOG_HEADER="## [$CHANGELOG_VERSION]"

if ! grep -q "^## \[$CHANGELOG_VERSION\]" CHANGELOG.md; then
    die "CHANGELOG.md missing section '$CHANGELOG_HEADER'. Add release notes first."
fi
log_ok "CHANGELOG.md contains $CHANGELOG_HEADER section"

# ─── Validation 8: CI status on HEAD (optional) ────────────────────────────
if [[ "$SKIP_CI_CHECK" != true ]]; then
    if command -v gh >/dev/null 2>&1; then
        CI_STATE="$(gh pr list --head "$CURRENT_BRANCH" --state merged --limit 1 --json number --jq '.[0].number // ""' 2>/dev/null || echo "")"

        # 직접 HEAD commit의 check suite 상태 조회 (main push 이후 CI)
        HEAD_CHECKS="$(gh api "/repos/modu-ai/moai-adk/commits/$LOCAL_SHA/status" --jq '.state' 2>/dev/null || echo "unknown")"

        case "$HEAD_CHECKS" in
            success)
                log_ok "CI status on HEAD: success"
                ;;
            pending)
                log_warn "CI status on HEAD: pending (in progress)"
                echo -n "  Proceed anyway? [y/N] "
                read -r reply
                [[ "$reply" =~ ^[Yy]$ ]] || die "Aborted by user"
                ;;
            failure|error)
                die "CI status on HEAD: $HEAD_CHECKS. Fix CI before releasing."
                ;;
            *)
                log_warn "CI status on HEAD: $HEAD_CHECKS (unable to verify)"
                ;;
        esac
    else
        log_warn "gh CLI not available — skipping CI status check"
    fi
else
    log_warn "CI check skipped (--skip-ci-check)"
fi

# ─── Validation 9: SPEC status 확인 (optional, informational) ───────────────
if [[ -d .moai/specs ]]; then
    DRAFT_COUNT="$(find .moai/specs -name 'spec.md' -exec grep -l '^status: draft' {} \; 2>/dev/null | wc -l | tr -d ' ')"
    if [[ "$DRAFT_COUNT" -gt 0 ]]; then
        log_warn "$DRAFT_COUNT SPEC(s) still in 'draft' status (not blocking, review if relevant)"
    fi
fi

# ─── Extract CHANGELOG section as tag annotation ───────────────────────────
TMP_NOTES="$(mktemp)"
trap 'rm -f "$TMP_NOTES"' EXIT

awk -v target="$CHANGELOG_HEADER " '
    $0 ~ "^"target {flag=1; print; next}
    /^## \[/ && flag {flag=0}
    flag
' CHANGELOG.md > "$TMP_NOTES"

if [[ ! -s "$TMP_NOTES" ]]; then
    die "Failed to extract CHANGELOG section for $VERSION"
fi

NOTES_LINES="$(wc -l < "$TMP_NOTES" | tr -d ' ')"
log_ok "Extracted $NOTES_LINES line(s) from CHANGELOG.md as tag annotation"

# ─── Final confirmation ────────────────────────────────────────────────────
echo
echo -e "${BOLD}=== Release Summary ===${NC}"
echo "  Version:     $VERSION"
echo "  Branch:      $CURRENT_BRANCH"
echo "  HEAD SHA:    ${LOCAL_SHA:0:12}"
echo "  Notes size:  $NOTES_LINES lines (from CHANGELOG.md $CHANGELOG_HEADER)"
echo "  Dry-run:     $DRY_RUN"
echo "  Hotfix:      $HOTFIX"
echo

if [[ "$DRY_RUN" == true ]]; then
    log_info "DRY RUN — skipping tag creation"
    echo
    echo "Tag annotation preview (first 30 lines):"
    echo "---"
    head -30 "$TMP_NOTES"
    echo "---"
    exit 0
fi

echo -n "Proceed with tag creation + push? [y/N] "
read -r confirm
[[ "$confirm" =~ ^[Yy]$ ]] || die "Aborted by user"

# ─── Create + push annotated tag ───────────────────────────────────────────
log_info "Creating annotated tag $VERSION..."
git tag -a "$VERSION" -F "$TMP_NOTES"
log_ok "Tag $VERSION created locally"

log_info "Pushing tag to origin (triggers GoReleaser)..."
git push origin "$VERSION"
log_ok "Tag pushed to origin"

# ─── Wait for GoReleaser workflow ──────────────────────────────────────────
if command -v gh >/dev/null 2>&1; then
    log_info "Monitoring GoReleaser workflow..."
    echo "  → https://github.com/modu-ai/moai-adk/actions"
    echo
    sleep 5 # Allow workflow dispatch to register

    RUN_ID="$(gh run list --workflow release.yml --limit 1 --json databaseId --jq '.[0].databaseId' 2>/dev/null || echo "")"

    if [[ -n "$RUN_ID" ]]; then
        log_info "Release workflow run: $RUN_ID (polling every 30s, ctrl-C to detach)"
        gh run watch "$RUN_ID" --interval 30 || log_warn "Watch detached (workflow may still be running)"

        WORKFLOW_STATE="$(gh run view "$RUN_ID" --json conclusion --jq '.conclusion' 2>/dev/null || echo "unknown")"
        case "$WORKFLOW_STATE" in
            success)
                log_ok "GoReleaser completed successfully"
                ;;
            failure)
                log_error "GoReleaser failed — check workflow logs"
                echo "  gh run view $RUN_ID --log-failed"
                exit 1
                ;;
            *)
                log_warn "GoReleaser state: $WORKFLOW_STATE"
                ;;
        esac
    else
        log_warn "Could not locate release workflow run — check manually"
    fi

    # Verify GitHub Release exists
    if gh release view "$VERSION" >/dev/null 2>&1; then
        RELEASE_URL="https://github.com/modu-ai/moai-adk/releases/tag/$VERSION"
        log_ok "GitHub Release available: $RELEASE_URL"
    else
        log_warn "GitHub Release for $VERSION not found yet (may take additional time)"
    fi
else
    log_warn "gh CLI not available — cannot verify GoReleaser completion"
fi

echo
log_ok "${BOLD}Release $VERSION complete${NC}"
echo
echo "Next steps:"
echo "  - Verify release assets: gh release view $VERSION"
echo "  - Update docs-site (Phase 5): docs/v$CHANGELOG_VERSION-reference/"
echo "  - Announce release if applicable"
