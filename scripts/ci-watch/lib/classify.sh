#!/bin/sh
# scripts/ci-watch/lib/classify.sh — Required vs auxiliary check discrimination
# Reads .github/required-checks.yml via yq (fallback: grep-based heuristic).
# Usage: . classify.sh; is_required "Lint" "main"
# Returns: exit 0 if required, exit 1 if auxiliary/unknown.

# Determine classify.sh's own directory (works both for direct execute and sourcing).
# When sourced, $0 is the parent script, so we use a marker approach.
_classify_self="$0"
if [ -f "$(dirname "$_classify_self")/classify.sh" ]; then
    SCRIPT_DIR_CLASSIFY="$(cd "$(dirname "$_classify_self")" 2>/dev/null && pwd)"
else
    # Sourced from another script — guess relative to the sourcing file's lib/ dir.
    # Callers that source classify.sh must have already sourced _common.sh.
    SCRIPT_DIR_CLASSIFY="$(cd "$(dirname "$0")/../lib" 2>/dev/null && pwd || echo ".")"
fi
# Only source _common.sh if log_step is not already defined.
if ! command -v log_step >/dev/null 2>&1; then
    # shellcheck source=_common.sh
    . "$SCRIPT_DIR_CLASSIFY/_common.sh"
fi

REPO_ROOT="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
REQUIRED_CHECKS_FILE="$REPO_ROOT/.github/required-checks.yml"

# _yq_available returns 0 if yq is installed, 1 otherwise.
_yq_available() {
    command -v yq >/dev/null 2>&1
}

# _load_required_checks_yq reads contexts for a branch via yq.
# $1 = branch pattern (e.g. "main", "release/*")
# Prints one context name per line.
_load_required_checks_yq() {
    branch_pattern="$1"
    yq e ".branches[\"$branch_pattern\"].contexts[]" "$REQUIRED_CHECKS_FILE" 2>/dev/null
}

# _is_in_auxiliary_yq returns 0 if check is in auxiliary list.
_is_in_auxiliary_yq() {
    check_name="$1"
    yq e ".auxiliary[]" "$REQUIRED_CHECKS_FILE" 2>/dev/null | grep -qxF "$check_name"
}

# _load_required_checks_grep uses grep+sed heuristic when yq is absent.
# This is a best-effort fallback for environments without yq.
# Only looks at literal branch names (no glob expansion).
_load_required_checks_grep() {
    branch_pattern="$1"
    # Extract the contexts block for the branch by scanning between branch header and next key.
    # Very rough — only handles simple indentation.
    awk "
        /^  ${branch_pattern}:/ { found=1; next }
        found && /^    contexts:/ { in_ctx=1; next }
        in_ctx && /^      - / { gsub(/^      - \"?|\"?$/, \"\"); print }
        in_ctx && /^    [^-]/ { in_ctx=0 }
        /^  [^ ]/ && found && !in_ctx { exit }
    " "$REQUIRED_CHECKS_FILE"
}

# _is_in_auxiliary_grep returns 0 if check is in auxiliary via grep.
_is_in_auxiliary_grep() {
    check_name="$1"
    # Look for the check under the 'auxiliary:' section.
    awk "
        /^auxiliary:/ { found=1; next }
        found && /^  - / { name=\$0; gsub(/^  - \"?|\"?$/, \"\", name); if (name == \"$check_name\") exit 0 }
        found && /^[^ ]/ { exit 1 }
    " "$REQUIRED_CHECKS_FILE"
    return $?
}

# is_auxiliary returns 0 if check_name is in the auxiliary list.
# $1 = check_name
is_auxiliary() {
    check_name="$1"
    if [ ! -f "$REQUIRED_CHECKS_FILE" ]; then
        return 1
    fi
    if _yq_available; then
        _is_in_auxiliary_yq "$check_name"
    else
        _is_in_auxiliary_grep "$check_name"
    fi
}

# _branch_matches_pattern returns 0 if branch matches pattern.
# Supports exact match and simple "prefix/*" glob.
_branch_matches_pattern() {
    branch="$1"
    pattern="$2"
    if [ "$branch" = "$pattern" ]; then
        return 0
    fi
    # Handle "prefix/*" style glob.
    case "$pattern" in
        *'/*')
            prefix="${pattern%/*}"
            case "$branch" in
                "$prefix"/*) return 0 ;;
            esac
            ;;
    esac
    return 1
}

# is_required returns 0 if check_name is required for the given branch.
# Returns 1 if auxiliary, unknown, or if the SSoT is missing.
# $1 = check_name
# $2 = branch_name
is_required() {
    check_name="$1"
    branch_name="$2"

    if [ ! -f "$REQUIRED_CHECKS_FILE" ]; then
        log_step "WARNING: $REQUIRED_CHECKS_FILE not found — treating '$check_name' as not required"
        return 1
    fi

    # Auxiliary checks are never required.
    if is_auxiliary "$check_name"; then
        return 1
    fi

    # Extract all branch pattern keys and test each one.
    if _yq_available; then
        branch_patterns="$(yq e '.branches | keys | .[]' "$REQUIRED_CHECKS_FILE" 2>/dev/null)"
    else
        branch_patterns="$(awk '/^branches:/{found=1;next} found && /^  [^ ]/{gsub(/:$/,"",$1);print $1} found && /^[^ ]/{exit}' "$REQUIRED_CHECKS_FILE")"
    fi

    for pat in $branch_patterns; do
        if _branch_matches_pattern "$branch_name" "$pat"; then
            if _yq_available; then
                if _load_required_checks_yq "$pat" | grep -qxF "$check_name"; then
                    return 0
                fi
            else
                if _load_required_checks_grep "$pat" | grep -qxF "$check_name"; then
                    return 0
                fi
            fi
        fi
    done

    return 1
}
