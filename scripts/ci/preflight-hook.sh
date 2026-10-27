#!/bin/sh
# Git hooks inherit repository-scoping variables. Tests create their own repos.
# Keep repository variables aligned with internal/gitenv.RepoScopingVars.
# GIT_CONFIG also selects the file changed by `git config` in fixtures;
# unlike GIT_CONFIG_GLOBAL it must not point at a caller-owned file here.
unset GIT_CONFIG GIT_DIR GIT_COMMON_DIR GIT_WORK_TREE GIT_INDEX_FILE \
  GIT_OBJECT_DIRECTORY GIT_ALTERNATE_OBJECT_DIRECTORIES GIT_NAMESPACE \
  GIT_PREFIX GIT_SUPER_PREFIX GIT_CEILING_DIRECTORIES GIT_QUARANTINE_PATH
exec make preflight
