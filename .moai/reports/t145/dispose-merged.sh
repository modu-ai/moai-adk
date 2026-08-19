#!/usr/bin/env bash
# t145 — origin/main 병합 확인된 L1 워크트리 폐기 스크립트 (감사 기준일 2026-08-20)
#
# [HARD] 실행 주체는 리드. 실행 전에 살아 있는 세션 목록과 대조할 것:
#   pgrep -x claude | while read p; do lsof -a -p $p -d cwd -Fn | sed -n 's/^n//p'; done
# 감사 시점 라이브 앵커: t143 t144 t146 (+ t145 = 이 카드 트리). 아래 목록에서 이미 제외돼 있고,
# 스크립트가 실행 시점에 한 번 더 cwd 를 확인해 물려 있으면 건너뛴다.
#
# 재검증 3종을 모두 통과한 트리만 지운다: (1) origin/main 조상 (2) 추적 미커밋 0 (3) 미추적 0
# 기본은 DRY RUN. 실제 삭제는:  APPLY=1 ./dispose-merged.sh
#
# L1(.claude/worktrees/) 이라 폐기 동사는 git worktree remove 다.
# moai worktree done 은 L2(~/.moai/worktrees/) 전용이므로 여기서는 범주 오류.
# 브랜치 ref 는 update-ref 로 지운다 (primary 체크아웃의 branch guard 회피가 아니라, 동등하고 더 좁은 동사).
set -uo pipefail

ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT" || exit 1
APPLY="${APPLY:-0}"
git fetch origin main >/dev/null 2>&1

live_cwds="$(pgrep -x claude 2>/dev/null | while read -r p; do lsof -a -p "$p" -d cwd -Fn 2>/dev/null | sed -n 's/^n//p'; done)"

removed=0; skipped=0
dispose() {
  name="$1"; branch="$2"; path=".claude/worktrees/$1"
  if [ ! -d "$path" ]; then echo "SKIP  $name — 디렉터리 없음 (prune 대상)"; skipped=$((skipped+1)); return; fi
  sha="$(git -C "$path" rev-parse HEAD 2>/dev/null)"
  if ! git merge-base --is-ancestor "$sha" origin/main 2>/dev/null; then
    echo "SKIP  $name — origin/main 미병합 ($sha)"; skipped=$((skipped+1)); return; fi
  if [ -n "$(git -C "$path" status --porcelain -uall 2>/dev/null)" ]; then
    echo "SKIP  $name — 미커밋/미추적 잔여"; skipped=$((skipped+1)); return; fi
  if printf '%s\n' "$live_cwds" | grep -qxF "$ROOT/$path"; then
    echo "SKIP  $name — 라이브 세션 앵커"; skipped=$((skipped+1)); return; fi
  if [ "$APPLY" = "1" ]; then
    git worktree remove "$path" && git update-ref -d "refs/heads/$branch"
    echo "REMOVED  $name ($branch)"
  else
    echo "WOULD REMOVE  $name ($branch)"
  fi
  removed=$((removed+1))
}

dispose agent-a05f5722891df4994 WT-t50
dispose agent-a099178d284386869 WT-t95
dispose agent-a0a46903178bf9897 WT-t99
dispose agent-a3b7729a864e4caef WT-t51
dispose agent-a55dd6a3e0ddbb399 WT-t54
dispose agent-a66650c94c57df3f8 WT-t56
dispose agent-a6ad15916b3c82119 WT-t49
dispose agent-a828e36a6a5eddb09 WT-t48
dispose agent-a88a53c67fbaa94fe WT-t60
dispose agent-a8e645d1c0d614569 worktree-agent-a8e645d1c0d614569
dispose agent-ad93b0b3d713a9364 WT-t96
dispose agent-adbea08ea93f5f41c WT-t55
dispose agent-add5f3f520fd7123d WT-t52
dispose agent-ae3785cfe611541f0 WT-t41
dispose agent-af71696d26ebce0bc WT-t63
dispose card-mcpcmd worktree-card-mcpcmd
dispose cc234-align WT-cc234-align
dispose factory-mode worktree-factory-mode
dispose moai-home-paths worktree-moai-home-paths
dispose release-v310 release/v3.1.0
dispose t103 WT-t103
dispose t104 WT-t104
dispose t106 WT-t106
dispose t107 WT-t107
dispose t110 WT-t110
dispose t111 WT-t111
dispose t113 WT-t113
dispose t115 WT-t115
dispose t116 WT-t116
dispose t32 WT-t32
dispose t36 WT-t36
dispose t39 worktree-t39
dispose t40 worktree-t40
dispose t45-i18n-keys worktree-t45-i18n-keys
dispose t47 WT-t47
dispose t53 WT-t53
dispose t58 worktree-t58
dispose t59 WT-t59
dispose t61-availability-docs worktree-t61-availability-docs
dispose t69 WT-t69
dispose t72-worktree-docs worktree-t72-worktree-docs
dispose t75 WT-t75
dispose t76 WT-t76
dispose t84 WT-t84
dispose t85 WT-t85
dispose t86 WT-t86
dispose t92 WT-t92
dispose t97 WT-t97
dispose wscfg-graph WT-wscfg-graph
dispose wscfg-paths WT-wscfg-paths
dispose wscfg-timing WT-wscfg-timing
dispose wscfg-worktree WT-wscfg-worktree

echo "---"
echo "대상 $removed / 건너뜀 $skipped"
[ "$APPLY" = "1" ] && git worktree prune && echo "prune 완료"
