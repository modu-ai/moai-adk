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

# ─── Validation 1: Version format (SemVer 2.0.0 with v prefix) ─────────────
# The pre-release and build-metadata groups below are the official SemVer 2.0.0
# grammar, not a loose approximation. A pre-release is a dot-separated series of
# identifiers; a numeric identifier carries no leading zero, so `rc.01` is
# rejected while `rc.1` is accepted. This matters for ordering: SemVer compares
# dot-separated numeric identifiers NUMERICALLY, so `rc.9` precedes `rc.10`,
# whereas the older undotted `rc9` / `rc10` form compares as a single
# alphanumeric identifier and sorts ASCII-lexically (`rc10` before `rc9`).
#
# The canonical pre-release form for this project is `-rc.N` (see
# CLAUDE.local.md §5 Pre-release Versioning). The undotted legacy form stays
# ACCEPTED so the historical `v3.0.0-rc12` line of tags remains valid input.
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-(0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*)(\.(0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*))*)?(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$ ]]; then
    die "Invalid version format: $VERSION (expected SemVer 2.0.0: vX.Y.Z, or vX.Y.Z-rc.N for a pre-release)"
fi

log_info "Release version: ${BOLD}$VERSION${NC}"
[[ "$DRY_RUN" == true ]] && log_warn "DRY RUN mode — no tag/push will occur"
[[ "$HOTFIX" == true ]] && log_info "Hotfix mode — relaxed branch check"

# ─── Validation 2: 하네스 경유 확인 (release provenance) ────────────────────
# 프로덕션 릴리스는 `/harness:release` 하네스를 통해서만 수행한다.
# 하네스가 MOAI_RELEASE_VIA_HARNESS=1 을 export 한 상태로 이 스크립트를 호출하며,
# 여기서 생성되는 annotated tag 에는 provenance trailer 가 기록된다
# (release.yml 의 verify-provenance job 이 이를 검증).
# --dry-run 은 tag/push 가 발생하지 않으므로 예외로 허용하되 경고를 출력한다.
if [[ "${MOAI_RELEASE_VIA_HARNESS:-}" != "1" ]]; then
    if [[ "$DRY_RUN" == true ]]; then
        log_warn "MOAI_RELEASE_VIA_HARNESS is not set — dry-run only."
        log_warn "The real release run requires the /harness:release harness (it exports MOAI_RELEASE_VIA_HARNESS=1)."
    else
        die "Release must run through the /harness:release harness (MOAI_RELEASE_VIA_HARNESS=1 not set). Run '/harness:release $VERSION' instead of invoking this script directly."
    fi
else
    log_ok "Harness provenance confirmed (MOAI_RELEASE_VIA_HARNESS=1)"
fi

# ─── Validation 3: Repository root ─────────────────────────────────────────
REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || die 'Not in a git repository')"
cd "$REPO_ROOT"

# ─── Validation 4: Clean working tree ──────────────────────────────────────
if [[ -n "$(git status --porcelain)" ]]; then
    git status --short
    die "Working tree is dirty. Commit or stash changes first."
fi
log_ok "Working tree clean"

# ─── Validation 5: Current branch ──────────────────────────────────────────
CURRENT_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
if [[ "$CURRENT_BRANCH" != "main" ]]; then
    if [[ "$HOTFIX" == true ]]; then
        log_warn "On branch '$CURRENT_BRANCH' (hotfix mode — allowed)"
    else
        die "Must be on 'main' branch (current: $CURRENT_BRANCH). Use --hotfix for hotfix branches."
    fi
fi
log_ok "On expected branch: $CURRENT_BRANCH"

# ─── Validation 6: Synced with origin ──────────────────────────────────────
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

# ─── Validation 7: Tag does not exist ──────────────────────────────────────
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    die "Tag $VERSION already exists (local)."
fi
if git ls-remote --exit-code --tags origin "$VERSION" >/dev/null 2>&1; then
    die "Tag $VERSION already exists on origin."
fi
log_ok "Tag $VERSION does not exist yet"

# ─── Validation 8: CHANGELOG.md 에 해당 버전 섹션 존재 ───────────────────────
CHANGELOG_VERSION="${VERSION#v}" # v2.14.0 → 2.14.0
CHANGELOG_HEADER="## [$CHANGELOG_VERSION]"

# CHANGELOG heading convention is split by release family: formal sections are
# bare ('## [3.1.0]'), while pre-release (rc) sections carry the v prefix
# ('## [v3.0.0-rc12]'). Accept both forms: when the bare heading is absent,
# fall back to the v-prefixed one. CHANGELOG_HEADER must point at the form
# that actually matched because the tag-annotation extraction below matches
# it literally.
if ! grep -q "^## \[$CHANGELOG_VERSION\]" CHANGELOG.md; then
    if grep -q "^## \[v$CHANGELOG_VERSION\]" CHANGELOG.md; then
        CHANGELOG_HEADER="## [v$CHANGELOG_VERSION]"
    else
        die "CHANGELOG.md missing section '## [$CHANGELOG_VERSION]' (or '## [v$CHANGELOG_VERSION]'). Add release notes first."
    fi
fi
log_ok "CHANGELOG.md contains $CHANGELOG_HEADER section"

# ─── Validation 9: CI status on HEAD (optional) ────────────────────────────
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

# ─── Validation 10: SPEC status 확인 (optional, informational) ───────────────
if [[ -d .moai/specs ]]; then
    DRAFT_COUNT="$(find .moai/specs -name 'spec.md' -exec grep -l '^status: draft' {} \; 2>/dev/null | wc -l | tr -d ' ')"
    if [[ "$DRAFT_COUNT" -gt 0 ]]; then
        log_warn "$DRAFT_COUNT SPEC(s) still in 'draft' status (not blocking, review if relevant)"
    fi
fi

# ─── Extract CHANGELOG section as tag annotation ───────────────────────────
TMP_NOTES="$(mktemp)"
trap 'rm -f "$TMP_NOTES"' EXIT

# NOTE: literal prefix match via index() — NOT a regex match. CHANGELOG_HEADER
# contains "[" / "]" which are regex metacharacters (character class); a regex
# match (`$0 ~ "^"target`) treats "[3.0.0]" as the class [3.0.0] and never
# matches "## [3.0.0] - date", so extraction silently yielded nothing. The
# trailing space in target disambiguates "3.0.0" from "3.0.0-rc1". The `started`
# guard stops after the first section so a duplicate header (e.g. a localized
# "## [3.0.0]" section) does not re-open extraction.
awk -v target="$CHANGELOG_HEADER " '
    !started && index($0, target) == 1 {flag=1; started=1; print; next}
    /^## \[/ && flag {flag=0}
    flag
' CHANGELOG.md > "$TMP_NOTES"

if [[ ! -s "$TMP_NOTES" ]]; then
    die "Failed to extract CHANGELOG section for $VERSION"
fi

# NOTE: the tag annotation is English-only. A per-version Korean notes file used
# to be appended here; release notes are now English-only and composed by the
# release harness (harness:release Phase 7) via `gh release edit --notes-file`,
# which is the load-bearing delivery path for the GitHub release body.

# ─── Provenance trailer 추가 ────────────────────────────────────────────────
# annotation 본문 끝에 provenance trailer 3줄을 append 한다.
# release.yml 의 verify-provenance job 이 이 3개 키를 literal 로 파싱하므로
# 키 철자(Released-via / Release-version / Release-commit)를 변경하면 안 된다.
TAG_COMMIT="$(git rev-parse HEAD^{commit})"
{
    echo
    echo "Released-via: harness:release"
    echo "Release-version: $VERSION"
    echo "Release-commit: $TAG_COMMIT"
} >> "$TMP_NOTES"

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

    # Bind the lookup to THIS tag's commit. Taking the newest release.yml run
    # instead would watch whatever ran most recently — a re-run of an earlier
    # tag, or a concurrent release — and report its verdict as this one's.
    TAG_SHA="$(git rev-parse "${VERSION}^{commit}")"

    # The run does not exist the instant the tag lands, so poll for it rather
    # than reading once. Bounded: a run that never appears is a failure to
    # report, not a reason to wait forever.
    RUN_ID=""
    LOOKUP_DEADLINE=$(( $(date +%s) + 180 ))
    while [[ -z "$RUN_ID" ]] && (( $(date +%s) < LOOKUP_DEADLINE )); do
        RUN_ID="$(gh run list --workflow release.yml --commit "$TAG_SHA" --limit 1 --json databaseId --jq '.[0].databaseId // empty' 2>/dev/null || echo "")"
        [[ -n "$RUN_ID" ]] && break
        sleep 5
    done

    if [[ -n "$RUN_ID" ]]; then
        log_info "Release workflow run: $RUN_ID (commit ${TAG_SHA:0:9}, polling every 30s, ctrl-C to detach)"
        # Polled rather than `gh run watch`: the watch blocks until the run
        # ends with no deadline of its own, and a hung workflow would hold the
        # release script open indefinitely. 45 minutes is well past the
        # observed build time and short enough to surface a stall.
        WATCH_DEADLINE=$(( $(date +%s) + 45 * 60 ))
        while (( $(date +%s) < WATCH_DEADLINE )); do
            RUN_STATUS="$(gh run view "$RUN_ID" --json status --jq '.status' 2>/dev/null || echo "")"
            [[ "$RUN_STATUS" == "completed" ]] && break
            sleep 30
        done
        if [[ "$RUN_STATUS" != "completed" ]]; then
            log_warn "GoReleaser still running after 45m — detaching"
            echo "  gh run watch $RUN_ID"
        fi

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
